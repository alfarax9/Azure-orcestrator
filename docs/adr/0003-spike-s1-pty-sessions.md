# ADR 0003: Spike S1, sesi agent di PTY yang selamat saat daemon restart

| Hal | Isi |
| --- | --- |
| Status | **Sebagian** (2026-10-26): desain terbukti di macOS; uji di VPS Linux dan dengan CLI agent sungguhan belum |
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

### Angka

| Ukuran | Nilai | Catatan |
| --- | --- | --- |
| Binary `astrad` linux/amd64 (`-trimpath -ldflags="-s -w"`) | 3,891,360 byte | Belum dijadikan `budgets/binary-size.txt`; riset meminta baseline setelah S1 selesai, termasuk bagian VPS |
| RSS satu session host | ±6.5 MB di macOS (5 host = 32.3 MB) | Termasuk ring buffer 1 MiB; ukur ulang di Linux untuk N-09 |

## Temuan selama spike

1. **Input hilang bila daemon langsung memutus koneksi.** Host menutup koneksi saat gagal mengirim scrollback, sehingga input yang sudah ada di buffer socket tidak terbaca. Diperbaiki: host hanya menutup sisi tulis, membaca input sampai habis, dan mencatat semua koneksi untuk ditutup saat berhenti.
2. **Probe sesi memutus terminal yang sedang dipakai.** Versi awal memperlakukan setiap koneksi sebagai daemon baru. Diperbaiki dengan frame pembuka `attach` dan `ping`.
3. **Koneksi lama menendang daemon yang baru reconnect.** Dengan aturan "attach terbaru menggantikan yang lama", goroutine koneksi lama yang terlambat diproses sempat memutus koneksi baru (test gagal 3 dari 6 run). Diperbaiki: semua koneksi yang attach menerima output dan tidak saling memutus. Sisa batasan: input dari dua koneksi yang aktif bersamaan bisa tiba dalam urutan apa pun; dengan satu daemon per VPS, ini hanya terjadi sesaat ketika daemon reconnect.
4. **Frame tanpa isi tidak terkirim** (perintah Stop). Diperbaiki: payload kosong tetap dikirim sebagai satu frame.
5. **Host yatim** bila folder socket dihapus. Diperbaiki dengan pemeriksaan file socket berkala.
6. **Repaint menambah baris prompt** di shell biasa (readline menulis ulang prompt saat SIGWINCH). Kosmetik; untuk TUI justru itu yang dibutuhkan.
7. **Exit code -1** bila agent dihentikan dengan sinyal (perilaku `exec.ExitError`). Cukup, tetapi API daemon nanti sebaiknya membedakan "dihentikan" dari "gagal".

## Belum terbukti (syarat menutup S1)

1. **CLI agent sungguhan.** `claude` dan `codex` belum terpasang di Mac pengembang. Perlu dipasang, lalu pemilik produk login dengan akunnya sendiri (F3-03); binary tidak dimodifikasi.
2. **VPS Linux.** Kriteria F3-02 adalah 5 sesi paralel stabil selama 2 jam, dan N-07 menuntut 100% sesi tersambung kembali setelah restart atau update daemon. Keduanya diuji di Ubuntu 24.04.
3. **systemd.** Dengan `KillMode=control-group` (default), `systemctl stop/restart astrad` ikut membunuh semua proses di cgroup daemon, termasuk host. Unit `astrad.service` perlu `KillMode=process`, atau host dijalankan di scope sendiri (`systemd-run --scope`). Keputusan diambil saat uji di VPS.
4. **Isolasi user (S3).** Agent harus berjalan sebagai user Linux terpisah (F12-01). Spawner perlu `SysProcAttr.Credential` atau `sudo -u`; itu bagian spike S3.
5. **Versi protokol.** Host yang berjalan bisa lebih tua dari daemon setelah update. Frame `attach` sebaiknya membawa versi protokol sebelum M1 (N-16).
6. **Ketikan < 100 ms (N-01)** diukur bersama terminal web di M1.

## Keputusan

Arsitektur session host terpisah per sesi dipakai untuk M1, menggantikan rencana fallback tmux, karena di macOS semua perilaku yang dibutuhkan sudah terbukti tanpa dependensi selain `creack/pty`. Keputusan ini direvisi bila uji VPS (poin 2 dan 3 di atas) gagal.
