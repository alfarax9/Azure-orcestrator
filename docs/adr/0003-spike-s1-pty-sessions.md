# ADR 0003: Spike S1, sesi agent di PTY yang selamat saat daemon restart

| Hal | Isi |
| --- | --- |
| Status | **Sebagian** (2026-10-26): terbukti di macOS, termasuk Claude Code dan Codex yang mengerjakan task nyata; uji di VPS Linux belum |
| Pemutus | Maulana Alfara Salim |
| Pengusul | Claude Code |
| Milestone | M0, syarat Gerbang 1 |
| Terkait | F3-02, F3-03, F3-04, N-01, N-07, N-09, PRD §7 (Stack), `internal/pty` |

## Pertanyaan spike

Bisakah daemon menjalankan CLI agent resmi (Claude Code, Codex) secara interaktif di PTY tanpa memodifikasi binary-nya, dan apakah sesi itu tetap hidup, dengan scrollback utuh, saat daemon ditutup atau di-restart?

## Desain yang diuji

- **Satu proses session host per sesi.** Host memegang PTY dan proses agent, menyimpan output terakhir di ring buffer (default 1 MiB), dan melayani lewat Unix socket `<dir>/<id>.sock` (mode 0600). Host adalah binary `astrad` yang sama dengan subcommand `session-host`, jadi tidak ada binary kedua.
- **Host bukan anak daemon.** Spawner menjalankan `astrad session-host --detach`, yang menjalankan ulang dirinya dengan `setsid` lalu keluar. Host diadopsi init/launchd, tidak pernah menjadi zombie daemon, dan tidak ikut mati saat daemon berhenti.
- **Protokol socket sederhana:** frame `jenis (1 byte) + panjang (uint32) + isi`. Koneksi dibuka dengan `attach` (menerima scrollback lalu output live, dan boleh mengetik) atau `ping` (sekadar memeriksa host hidup). Beberapa koneksi boleh attach bersamaan dan menerima output yang sama; koneksi baru tidak pernah memutus yang lama.
- **Repaint saat attach:** scrollback dikirim ulang, lalu ukuran terminal digeser satu kolom dan dikembalikan agar TUI menggambar ulang layarnya (SIGWINCH).
- **Environment agent hanya dari daftar izin** (`HOME`, `USER`, `LOGNAME`, `PATH`, `SHELL`, `LANG`, `LC_ALL`, `TMPDIR`, ditambah `TERM`, `COLORTERM`). Spike menemukan bahwa environment proses di mesin pengembang berisi puluhan variabel `CLAUDE_*` dan `ANTHROPIC_*`, termasuk token sesi; tanpa daftar izin semuanya akan ikut ke agent.
- **Penghentian:** SIGHUP ke process group agent, SIGKILL setelah 3 detik bila masih hidup; timer dibatalkan begitu agent keluar agar PID yang dipakai ulang tidak pernah disinyal.
- **Host berhenti sendiri** bila file socket-nya hilang (diperiksa tiap 10 detik), agar tidak ada sesi yang tidak bisa dijangkau tetapi tetap berjalan.

## Hasil di macOS (Apple silicon, Go 1.27.3)

| Uji | Hasil |
| --- | --- |
| Sesi bertahan saat daemon pergi dan kembali | Lulus. Koneksi ditutup, proses baru menemukan sesi lewat `Sessions`, menerima scrollback lengkap, dan bisa mengetik lagi (`TestSessionSurvivesDaemonRestart`, juga lewat `astrad session` dari proses terpisah) |
| Host lepas dari proses yang menjalankannya | Lulus. Parent host adalah PID 1 |
| TUI layar penuh (`vi`) | Lulus. Teks yang diketik tetap terlihat setelah attach ulang, layar digambar ulang, `:q!` keluar dengan kode 0 |
| Input dari daemon yang langsung memutus koneksi | Lulus setelah perbaikan (lihat Temuan) |
| Exit code sampai ke daemon yang terlambat attach | Lulus (`TestExitCodeReachesALateDaemon`) |
| Environment daemon tidak bocor ke agent | Lulus (`TestAgentGetsOnlyItsSpecEnvironment`) |
| Resize sampai ke agent | Lulus (`stty size` membaca ukuran baru) |
| Goroutine bocor setelah host berhenti | Nol, diukur dengan profil `goroutineleak` |
| `make check` | Lulus; coverage `internal/pty` 72.8% |
| Stabilitas test | 15 kali run berturut-turut dengan `-race` tanpa gagal, tanpa host tertinggal |

### Dengan CLI agent sungguhan (Claude Code 2.1.310, Codex 0.164.0, binary tidak dimodifikasi)

Pemilik produk login sendiri lewat Terminal macOS (F3-03). Uji tidak mengirim prompt, sehingga tidak memakai kuota.

| Uji | Claude Code | Codex |
| --- | --- | --- |
| Mulai di PTY lewat `astrad session start` | Lulus; host siap dalam 54 ms, layar pertama muncul | Lulus; layar pertama muncul |
| Daemon pergi, proses baru menemukan sesi dan attach ulang | Lulus; layar yang sama dirender ulang | Lulus; layar yang sama dirender ulang |
| Input sampai ke agent | Lulus; Enter pada "No, exit" membuat agent keluar | Ctrl-C diabaikan di dialog trust (perilaku Codex) |
| Exit code sampai ke daemon | Lulus; kode 1 setelah menolak trust | Lulus; kode 0 setelah Stop |
| Stop menghentikan seluruh pohon proses | — | Lulus; wrapper Node dan binary native Codex mati bersamaan |
| Login terbaca dari dalam session host dengan environment terbatas | Lulus; `claude auth status`: `loggedIn: true`, `claude.ai`, `pro` | Lulus; `codex login status`: "Logged in using ChatGPT" |

Login kedua agent (macOS Keychain dan `~/.codex`) tetap terbaca dari host yang dilepas dengan `setsid`, hanya dengan `HOME` dan variabel lain di daftar izin. Daemon tidak membaca kredensial itu.

### Uji kerja dengan kuota (dijalankan pemilik produk, 2026-10-26)

Pemilik produk menjalankan `astrad session start/send/tail` dari Terminal macOS di folder `~/astra-s1/work`. Setiap `send` dan `tail` adalah proses terpisah yang attach lalu pergi, jadi agent bekerja tanpa daemon yang terpasang di antara perintah.

| Agent | Hasil |
| --- | --- |
| Claude Code | **Lulus.** Trust diterima lewat input (`\x1b[B\r`), prompt dikirim dan dijalankan. Menurut transcript, pukul 11:04:09 Claude membuat `notes.md` dan `count.sh` lewat Bash, lalu memeriksa dan menjalankannya pukul 11:04:19. Kedua file benar dan `count.sh` executable. Prompt yang terkirim dua kali ditangani dengan benar ("Both files are already there"). `/exit` keluar dengan kode 0. |
| Codex, run pertama (pemilik produk) | **Tidak selesai.** Trust diterima (folder tercatat di `~/.codex/config.toml`). Layar terakhir menampilkan "tab to queue message", yang berarti Codex sedang mengerjakan prompt; `astrad session stop` dijalankan beberapa detik kemudian dan menghentikannya di tengah jalan, sehingga `README.md` tidak dibuat dan tidak ada thread yang tersimpan. |
| Codex, run ulang (Claude Code, kuota minimal) | **Lulus.** Prompt dikirim sebagai bracketed paste (`\x1b[200~...\x1b[201~`), jeda 1 detik, lalu Enter. Tanpa daemon terpasang, Codex membaca `count.sh`, membuat `README.md` (+3 −0) yang benar, tidak mengubah file lain, dan selesai dalam 8 detik. Database Codex mencatat tepat satu thread. Stop setelahnya keluar dengan kode 0. |

Temuan dari uji ini:

- **Prompt untuk Codex harus dikirim sebagai bracketed paste.** Teks yang dikirim sekaligus tanpa penanda paste ditahan oleh deteksi "paste burst" Codex: tampil terlambat satu langkah (baru muncul saat input berikutnya datang), dan Enter sesudahnya tidak mengirim prompt. Bracketed paste lalu Enter terpisah langsung berhasil. Ini penting untuk M1, karena daemon harus memasukkan task ke agent tanpa manusia mengetik (komentar klien F6-03, error gate F7-02). Codex menyalakan mode bracketed paste (`ESC[?2004h`) sendiri, jadi cara ini memang didukung.
- **TUI bertanya ke terminal dan tidak ada yang menjawab.** Saat mulai, Codex mengirim query posisi kursor (`ESC[6n`), device attributes (`ESC[c`), flag kitty keyboard protocol (`ESC[?u`, setelah menyalakan `ESC[>7u`), dan warna latar (OSC 10/11). Tanpa emulator terminal, query itu tidak terjawab; Codex tetap jalan, tetapi untuk M1 perlu diputuskan apakah host menjawab query dasar ini sendiri ketika tidak ada terminal web yang terpasang.
- **Database Codex ditulis setelah satu giliran selesai**, bukan saat prompt dikirim. Polling tabel `threads` tidak bisa dipakai untuk mengetahui apakah prompt sudah terkirim; layar (atau adapter S2) yang harus dipakai.
- **Layar TUI baru bisa dibaca dengan emulator terminal sungguhan.** Uji ulang memakai `pyte` di virtualenv sementara untuk merender scrollback; hasilnya persis seperti layar Codex, sedangkan `tail -strip` menumpuk layar lama dan baru.
- **Sesi yang sudah selesai masih muncul di `astrad session ls`** selama host menunggu exit code diambil (sampai 10 menit). Sesuai desain, tetapi daftar sesi perlu membedakan "berjalan" dan "selesai, exit code belum diambil".
- **`tail -strip` tidak layak untuk melihat TUI.** Scrollback berisi seluruh riwayat redraw, sehingga layar lama dan baru bertumpuk. Untuk M1, tampilan terminal harus memakai emulator sungguhan (xterm.js di web UI), bukan teks yang dibersihkan.
- **Untuk S2:** transcript Claude Code berupa JSONL di `~/.claude/projects/<path>/<uuid>.jsonl`, dengan entri `tool_use` (nama tool dan input) yang cukup untuk tampilan terstruktur. Codex 0.164 tidak lagi menulis `~/.codex/sessions/*.jsonl`; state-nya ada di SQLite (`~/.codex/state_5.sqlite`, tabel `threads` dengan kolom `rollout_path`). Format itu tidak didokumentasikan, sehingga adapter Codex di S2 lebih berisiko (PRD §14).

### Angka

| Ukuran | Nilai | Catatan |
| --- | --- | --- |
| Binary `astrad` linux/amd64 (`-trimpath -ldflags="-s -w"`) | 3,891,360 byte | Belum dijadikan `budgets/binary-size.txt`; riset meminta baseline setelah S1 selesai, termasuk bagian VPS |
| RSS satu session host | ±6.5–6.8 MB di macOS (5 host = 32.3 MB) | Termasuk ring buffer 1 MiB; ukur ulang di Linux untuk N-09 |
| RSS Claude Code saat diam | ±207 MB | Lima sesi ±1 GB; besar untuk VPS 4 GB (F10-04) |
| RSS Codex saat diam | ±149 MB (wrapper Node 47.5 MB + binary native 101 MB) | Lima sesi ±0.75 GB |

## Temuan selama spike

1. **Input hilang bila daemon langsung memutus koneksi.** Host menutup koneksi saat gagal mengirim scrollback, sehingga input yang sudah ada di buffer socket tidak terbaca. Diperbaiki: host hanya menutup sisi tulis, membaca input sampai habis, dan mencatat semua koneksi untuk ditutup saat berhenti.
2. **Probe sesi memutus terminal yang sedang dipakai.** Versi awal memperlakukan setiap koneksi sebagai daemon baru. Diperbaiki dengan frame pembuka `attach` dan `ping`.
3. **Koneksi lama menendang daemon yang baru reconnect.** Dengan aturan "attach terbaru menggantikan yang lama", goroutine koneksi lama yang terlambat diproses sempat memutus koneksi baru (test gagal 3 dari 6 run). Diperbaiki: semua koneksi yang attach menerima output dan tidak saling memutus. Sisa batasan: input dari dua koneksi yang aktif bersamaan bisa tiba dalam urutan apa pun; dengan satu daemon per VPS, ini hanya terjadi sesaat ketika daemon reconnect.
4. **Frame tanpa isi tidak terkirim** (perintah Stop). Diperbaiki: payload kosong tetap dikirim sebagai satu frame.
5. **Host yatim** bila folder socket dihapus. Diperbaiki dengan pemeriksaan file socket berkala.
6. **Repaint menambah baris prompt** di shell biasa (readline menulis ulang prompt saat SIGWINCH). Kosmetik; untuk TUI justru itu yang dibutuhkan.
7. **Exit code -1** bila agent dihentikan dengan sinyal (perilaku `exec.ExitError`). Cukup, tetapi API daemon nanti sebaiknya membedakan "dihentikan" dari "gagal".
8. **Kedua agent meminta "trust folder" di setiap path baru.** Karena setiap worker punya worktree sendiri, setiap worker baru akan dimulai dengan dialog ini dan berstatus "Needs you" sebelum bisa bekerja. Pilihan untuk M1, perlu keputusan pemilik produk: (a) developer menyetujui trust per worker dari UI, atau (b) daemon mencatat trust untuk worktree yang dibuatnya sendiri di konfigurasi agent. Opsi (b) berarti daemon menulis ke `~/.claude.json` dan konfigurasi Codex milik user; perlu ADR dan dicek terhadap aturan PRD §10.
9. **Memori agent jauh lebih besar dari memori Astra.** Claude Code ±207 MB dan Codex ±149 MB per sesi, dibanding ±7 MB per host. Batas cgroups (F10-04, spike S6) dan jumlah sesi paralel default harus dihitung dari angka agent, bukan dari daemon.

## Belum terbukti (syarat menutup S1)

1. **VPS Linux.** Kriteria F3-02 adalah 5 sesi paralel stabil selama 2 jam, dan N-07 menuntut 100% sesi tersambung kembali setelah restart atau update daemon. Keduanya diuji di Ubuntu 24.04.
2. **systemd.** Dengan `KillMode=control-group` (default), `systemctl stop/restart astrad` ikut membunuh semua proses di cgroup daemon, termasuk host. Unit `astrad.service` perlu `KillMode=process`, atau host dijalankan di scope sendiri (`systemd-run --scope`). Keputusan diambil saat uji di VPS.
3. **Isolasi user (S3).** Agent harus berjalan sebagai user Linux terpisah (F12-01). Spawner perlu `SysProcAttr.Credential` atau `sudo -u`; itu bagian spike S3.
4. **Versi protokol.** Host yang berjalan bisa lebih tua dari daemon setelah update. Frame `attach` sebaiknya membawa versi protokol sebelum M1 (N-16).
5. **Ketikan < 100 ms (N-01)** diukur bersama terminal web di M1.

## Keputusan

Arsitektur session host terpisah per sesi dipakai untuk M1, menggantikan rencana fallback tmux, karena di macOS semua perilaku yang dibutuhkan sudah terbukti, termasuk dengan Claude Code dan Codex sungguhan, tanpa dependensi selain `creack/pty`. Keputusan ini direvisi bila uji VPS (poin 1 dan 2 di bagian Belum terbukti) gagal.
