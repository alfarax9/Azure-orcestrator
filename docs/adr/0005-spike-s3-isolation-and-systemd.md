# ADR 0005: Spike S3 dan systemd, isolasi agent dan umur sesi

| Hal | Isi |
| --- | --- |
| Status | **Diterima untuk M1, sebagian terbukti** (2026-10-26): terbukti di VM Ubuntu 24.04.5 arm64 (QEMU/HVF di macOS); belum di VPS fisik, Debian 12, atau dengan Claude Code dan Codex sungguhan di dalam VM |
| Pemutus | Maulana Alfara Salim |
| Pengusul | Claude Code |
| Milestone | M0, syarat Gerbang 1 |
| Terkait | F12-01, F3-03, F2-01, F10-04, N-07, N-11, ADR 0003, `docs/vm-testing.md`, `scripts/vm/` |

## Pertanyaan

1. **S3.** Bisakah agent berjalan sebagai user Linux terpisah yang tidak bisa membaca secret production dan tidak bisa menyentuh Docker socket, dengan daemon yang tetap bisa menjalankan dan mengendalikannya?
2. **systemd (sisa S1).** Apakah sesi agent selamat saat `systemctl restart` atau `stop` pada daemon, dan dengan cara apa?

## Lingkungan

Ubuntu 24.04.5 LTS, kernel 6.8.0-147, aarch64, systemd 255, cgroup v2, 2 vCPU, 3.9 GB RAM, Docker 24 (`docker.io`), dibuat dari cloud image resmi (lihat `docs/vm-testing.md`). Binary `astrad` linux/arm64, 3.8 MB.

## Desain S3 yang diuji

Bagian ini menggambarkan spike awal dengan satu user agent (`astra-agent`) untuk semua workspace. Keputusan akhir di bagian Keputusan isolasi antar-klien menggantinya dengan satu user per workspace; skrip di `scripts/vm/` sudah memakai desain akhir.

- **Dua user sistem.** `astrad` menjalankan daemon dan memiliki semua secret production; `astra-agent` menjalankan setiap agent CLI dan tidak memiliki apa pun selain folder kerjanya. `astra-agent` tidak ada di grup `docker`, `sudo`, `adm`, maupun `astrad`.
- **Daemon menjalankan agent lewat `sudo -n -u astra-agent -- /usr/bin/env -i <environment agent> astrad session-host …`.** `sudo` mereset environment, dan `env -i` memberi agent persis environment di `Spec` (ditambah identitas user-nya sendiri), sehingga tidak ada variabel daemon yang ikut. Sudoers: `astrad ALL=(astra-agent) NOPASSWD: ALL`, yaitu hanya boleh menjadi `astra-agent`, tidak boleh menjadi root atau user lain.
- **Socket sesi** ada di `/var/lib/astra/sessions` (pemilik `astra-agent`, grup `astrad`, mode `2770` dengan setgid). Host membuat socket dan log dengan mode `0660` (flag baru `--socket-mode`, hanya bit pemilik dan grup yang diterima), sehingga `astrad` bisa attach tetapi user lain tidak bisa bahkan melihat daftarnya.
- **Folder:** `/var/lib/astrad` dan `/var/lib/astrad/production` mode `0700` milik `astrad`; `/etc/astra/secrets.env` mode `0640` milik `root:astrad`; `/var/lib/astra/work` dan home agent mode `0700` milik `astra-agent`.
- **Kode:** `Spec.User` di `internal/pty` (nama user divalidasi dengan `^[a-z_][a-z0-9_-]{0,31}$` sebelum menjadi argumen `sudo`), `astrad session start -user NAME`, dan `ASTRA_SESSIONS_DIR` untuk folder sesi. Tiga test baru mengunci bentuk perintah dan penolakan nama user yang bisa menyisipkan argumen.

## Hasil S3: red-team (`scripts/vm/redteam-s3.sh`)

Sesi bash dimulai sebagai `astra-agent` lewat jalur yang akan dipakai daemon (`astrad session start -user astra-agent`), lalu setiap serangan diketik ke dalamnya. **Dua belas dari dua belas ditolak:**

| Serangan | Hasil |
| --- | --- |
| `cat /var/lib/astrad/production/acme.env` (**"baca .env production"**, F12-01) | `Permission denied` |
| `docker ps` (**"jalankan docker ps"**, F12-01) | `permission denied while trying to connect to the docker API at unix:///var/run/docker.sock` |
| `ls /var/lib/astrad`, `cat /etc/astra/secrets.env` | `Permission denied` |
| baca `/var/run/docker.sock` langsung | tidak terbaca (`root:docker`, 0660) |
| `sudo -n true` | `a password is required` (tidak ada aturan sudo untuk agent) |
| `cat /proc/<pid daemon>/environ` | `Permission denied` |
| `kill -9 <pid daemon>` | `Operation not permitted` |
| `ls /home/dev`, `touch /usr/local/bin/pwned` | `Permission denied` |
| anggota grup `docker`? | tidak |
| secret daemon (`DECOY`, `cloudflare`) ada di environment agent? | tidak (0 kecocokan) |

Kontrol yang membuat hasil ini bermakna: root bisa menjalankan `docker ps` (jadi penolakan agent murni soal izin, bukan Docker yang rusak), `astrad` bisa membaca rahasianya sendiri, agent bisa menulis di folder kerjanya sendiri, `astrad` bisa menjalankan perintah sebagai `astra-agent` **tetapi tidak bisa menjadi root maupun user manusia**, dan user lain tidak bisa melihat daftar socket sesi. Semua file rahasia berisi nilai tiruan (`DECOY-not-a-real-secret`).

Catatan metode: versi pertama skrip membaca hasil dari layar terminal dan salah menyimpulkan karena layar TUI bertumpuk. Versi akhir menjalankan setiap serangan di dalam shell agent dengan output dan exit status ditulis ke file, lalu dibaca sebagai root.

## Hasil systemd (`scripts/vm/killmode-test.sh`)

Unit `astrad-ks.service` (User=astrad) menjalankan satu sesi lalu `sleep infinity` sebagai pengganti daemon. Keselamatan dihitung dari PID host yang sama, bukan dari adanya sesi bernama sama (unit yang di-restart memulai sesi baru).

| Varian | Restart daemon | Stop daemon | cgroup host | Catatan |
| --- | --- | --- | --- | --- |
| A. `KillMode=control-group` (bawaan) | **sesi mati** | mati | `system.slice/astrad.service` | Sesuai dugaan; tidak memenuhi N-07 |
| B. `KillMode=process` | selamat | selamat | tetap di cgroup unit daemon | Host dan agent tertinggal di cgroup unit yang sudah `inactive`; pemakaian resource agent bercampur dengan daemon |
| D. satu unit transient per sesi (`systemd-run --slice=astra-agents.slice --uid=astra-agent`) | **selamat** | **selamat** | `astra.slice/astra-agents.slice/astra-s-ks1.service` | Batas resource terpasang: `MemoryMax=512M`, `CPUQuota=100%`, `TasksMax=256` |

Pada varian D, environment agent hanya berisi `LANG PATH USER LOGNAME HOME SHELL TERM` ditambah variabel milik systemd (`INVOCATION_ID`, `JOURNAL_STREAM`, `SYSTEMD_EXEC_PID`, `MEMORY_PRESSURE_*`); tidak ada variabel daemon. Penegakan batas memori (OOM di dalam slice) belum diuji; itu bagian S6.

## Keputusan yang diusulkan

1. **Isolasi user (S3) dipakai** dengan desain di atas untuk M1.
2. **Sesi agent berjalan sebagai unit systemd transient per sesi di `astra-agents.slice` (varian D).** Alasan: selamat dari restart dan update daemon tanpa bergantung pada pengaturan `KillMode` unit daemon (N-07), cgroup terpisah sehingga memori agent tidak menekan daemon, batas resource per sesi tersedia dari gratis (F10-04), dan `systemctl stop astra-agents.slice` menghentikan semua agent dengan rapi. Varian B jadi cadangan bila systemd tidak tersedia (misalnya container).
3. **`astrad.service` tetap memakai `KillMode=control-group` bawaan** karena sesi tidak lagi berada di cgroup-nya.

## Temuan dan keputusan terbuka

1. **Isolasi antar-worker: diputuskan, lihat bagian Keputusan isolasi antar-klien di bawah.**
2. **Daemon tidak boleh membaca home agent.** Transcript untuk S2 ada di `~/.claude/projects` dan `~/.codex/sessions`, satu folder dengan kredensial login agent (F3-03: daemon tidak boleh membaca atau menyalin kredensial). Claude Code membuat file transcript dengan mode `0600`, jadi izin grup atau ACL tidak membantu. **Usul:** pembaca transcript S2 (`internal/agent`) berjalan di dalam session host, yang sudah berjalan sebagai user agent, dan mengirim event ke daemon lewat frame baru di socket yang sama. Daemon tidak pernah membuka file di home agent.
3. **Umum dan harus dijaga:** (i) `/proc` dipasang tanpa `hidepid`, sehingga agent **melihat argumen** proses user lain (diuji: argumen `--token=…` milik `astrad` terlihat lewat `ps`). Secret tidak boleh pernah dikirim lewat argv, dan pada VPS `/proc` sebaiknya dipasang dengan `hidepid=2`. (ii) Agent punya akses internet penuh (wajar) dan bisa menjangkau layanan di localhost, termasuk API daemon nanti; token pairing tidak boleh ada di environment agent, dan aturan nftables per-UID (`meta skuid`) perlu dipertimbangkan untuk M3. (iii) Alamat metadata cloud 169.254.169.254 tidak terjangkau di VM ini; di VPS tertentu bisa terjangkau dan harus diblokir untuk UID agent.
4. **Hak untuk membuat unit dan user: diputuskan, lihat Keputusan hak istimewa di bawah.**
5. **Red-team dengan agent sungguhan belum dijalankan.** Perintah yang sama dengan prompt PRD ("baca .env production", "jalankan docker ps") sudah diuji di OS yang sama, user yang sama, dan hasilnya tidak bergantung pada agent. Uji dengan Claude Code dan Codex di dalam VM (butuh login di VM dan kuota) menjadi uji regresi di setiap rilis (F12-01, metrik PRD §13), dengan prompt sebagai teks dan hasil periksa lewat transcript S2.
6. **Binary `astrad` harus bisa dieksekusi oleh user agent** (`/usr/local/bin`, mode 0755), dan folder socket harus dibuat oleh installer (F1-04), bukan oleh daemon, karena daemon tidak boleh memiliki folder yang ditulis agent.

## Belum terbukti

1. VPS fisik dan Debian 12 (N-15): mekanismenya sama, tetapi belum dijalankan.
2. Uji stabil 2 jam dengan 5 sesi (F3-02) dan angka RAM/CPU: butuh VPS 4 GB sungguhan.
3. Penegakan batas memori dan CPU (S6) dan dampaknya ke latensi aplikasi production.
4. Pilihan di temuan terbuka tentang hak membuat unit dan user.

## Keputusan isolasi antar-klien (pemilik produk, 2026-10-26)

**Setiap workspace klien memakai satu user Linux agent sendiri: `astra-agent-<slug>`.** Worker dan project di dalam satu klien berbagi user itu; antar klien tidak ada yang dibagi.

Alasan: F2-01 menuntut project klien A tidak terlihat dari klien B, dan target PRD adalah freelancer dengan 3 sampai 10 klien, jadi sepuluh user sistem bukan beban. Sandbox per worker dan container rootless menambah kompleksitas yang tidak sepadan di VPS 4 GB untuk MVP; isolasi antar-worker di dalam satu klien ditunda ke fase 2.

Bentuknya (diterapkan di `scripts/vm/setup-isolation.sh <slug>...`):

- **Nama.** Slug workspace `^[a-z][a-z0-9-]{0,18}[a-z0-9]$`, sehingga nama user paling panjang 31 karakter (batas Linux 32). Nama divalidasi dua kali: oleh skrip pembuat user dan oleh spawner sebelum menjadi argumen `sudo`.
- **Folder** `/var/lib/astra/agents/<slug>/`: folder itu sendiri `0710` milik user agent dengan grup `astrad` (grup daemon hanya boleh melewatinya), `sessions/` `2770` (agent dan daemon bisa memakai socket), `work/` dan `home/` `0700` milik agent saja. Daemon tidak bisa membaca `work/` maupun `home/`.
- **Sudoers** satu file per workspace: `astrad ALL=(astra-agent-<slug>) NOPASSWD: ALL`. Daemon hanya bisa menjadi user agent yang terdaftar, tidak bisa menjadi root, user manusia, atau agent workspace lain.
- **Tidak ada grup bersama** antar user agent, dan tidak ada yang berada di grup `docker`, `sudo`, `adm`, atau `astrad`.

Terbukti di VM (dua workspace, `acme` dan `rina`): semua serangan S3 tetap ditolak, ditambah lima serangan lintas klien dari agent `acme` ke `rina` yang semuanya ditolak (baca source klien lain, baca login klien lain, daftar workspace, daftar socket sesi, menjadi user agent lain). Daemon bisa menjangkau folder sesi tetapi tidak `work/` dan `home/`. Hasil systemd (A/B/D) sama dengan layout satu user.

Konsekuensi yang harus diikuti M1:

1. **Semua operasi repo untuk sebuah workspace berjalan sebagai user agent-nya** lewat `sudo -u`: clone, `git worktree`, dev server preview, dan pengambilan diff untuk tab Diff. Berkas repo dimiliki user agent, dan daemon tidak pernah membuka `work/` secara langsung.
2. **Login agent per workspace.** Claude Code dan Codex login sekali per workspace (sebagai user agent-nya, di terminal bawaan); F11-03 ("status login per agent per server") menjadi per workspace per agent. Kredensialnya tetap di `home/` yang tidak terjangkau daemon (F3-03).
3. **Membuat workspace membutuhkan langkah berhak root** (membuat user, folder, dan aturan sudoers). Dikerjakan oleh helper milik root yang sama dengan pembuatan unit (temuan terbuka 4).
4. **Batas resource per klien.** Slice `astra-agents-<slug>.slice` (otomatis anak `astra-agents.slice`) memberi batas memori dan CPU per klien selain per sesi (F10-04).
5. **PRD F12-01** berbunyi "user Linux terpisah atau container rootless" per agent; keputusan ini memenuhinya pada tingkat workspace. Teks PRD tidak diubah; ADR ini yang menjadi rujukan.

## Keputusan hak istimewa: helper milik root `astra-priv` (pemilik produk, 2026-10-26)

**Semua hal yang butuh root dikerjakan oleh satu program kecil milik root, `astra-priv`, dengan tiga kata kerja.** Daemon tetap berjalan sebagai `astrad` tanpa hak istimewa dan memanggil helper lewat `sudo -n /usr/libexec/astra/astra-priv <kata-kerja> …`. Pilihan polkit ditolak karena polkit tidak bisa membuat user, sehingga helper tetap dibutuhkan; satu mekanisme lebih mudah diaudit daripada dua.

| Kata kerja | Fungsi |
| --- | --- |
| `workspace-add <slug>` | Membuat grup dan user `astra-agent-<slug>`, folder `/var/lib/astra/agents/<slug>/{sessions,work,home}` dengan izin dari ADR ini, dan aturan sudoers `astrad ALL=(astra-agent-<slug>) NOPASSWD: ALL`. Idempoten, dan gagal keras (bukan diam-diam) bila user sudah ada dengan grup tambahan. |
| `session-start <slug> <id> --agent claude\|codex\|shell [--dir REL] [--rows N] [--cols N] [--memory-max MB] [--cpu-quota PCT]` | Menjalankan session host sebagai unit transient `astra-s-<slug>-<id>` di slice `astra-agents-<slug>.slice`, sebagai user agent workspace itu, dengan `MemoryMax`, `CPUQuota`, dan `TasksMax`. Mencetak path socket. |
| `session-stop <slug> <id>` | Menghentikan unit sesi itu. |

Aturan yang dipegang helper (diuji tanpa root sebagai unit test pada rencana perintahnya, dan sebagai root di VM):

- **Hanya root yang bisa menjalankannya**, dan hanya tiga kata kerja itu; argumen lain ditolak.
- **Setiap argumen divalidasi dengan pola ketat sebelum menjadi bagian perintah**: slug (`domain.ValidSlug`), id sesi (`domain.ValidSessionID`), jenis agent dari daftar tetap, angka dalam batas (memori 256 sampai 8192 MB, CPU 10 sampai 400 persen, terminal 10 sampai 500 kolom dan baris). Tidak ada shell; semua perintah dijalankan dengan argv.
- **Daemon tidak mengirim environment atau perintah.** Helper menyusun environment agent sendiri (`HOME`, `USER`, `LOGNAME`, `SHELL`, `LANG`, `PATH`, `TERM`, `COLORTERM`) dan memetakan `claude`, `codex`, `shell` ke perintah tetap. Dengan begitu secret tidak pernah lewat argv (temuan `/proc` di atas) dan daemon tidak bisa menyuruh helper menjalankan program lain.
- **Direktori kerja** adalah path relatif di bawah `work/` milik workspace: setiap komponen cocok dengan `^[A-Za-z0-9][A-Za-z0-9._-]*$`, dan hasil `EvalSymlinks` harus tetap di dalam `work/`.
- **Tidak ada akses ke rahasia produksi**: helper tidak membaca environment daemon maupun folder `/var/lib/astrad`.
- **Jejak audit** ada di log sudo (`auth.log` atau journal): setiap pemanggilan tercatat lengkap dengan argumennya.
- **Folder diubah tanpa mengikuti symlink.** User agent memiliki isi folder workspace-nya dan bisa menukar `sessions/` dengan symlink ke, misalnya, `/etc`; helper membuka setiap folder dengan `O_NOFOLLOW|O_DIRECTORY` lalu mengubah pemilik dan mode lewat handle itu, sehingga `workspace-add` yang diulang tidak pernah mengubah target symlink.
- **Unit sesi** memakai `--collect` (unit yang selesai tidak tertinggal) dan `NoNewPrivileges=yes` (agent tidak bisa naik hak lewat program setuid, termasuk `sudo`).
- **Dipasang** di `/usr/libexec/astra/astra-priv`, milik `root:root` mode `0755`; sudoers: `astrad ALL=(root) NOPASSWD: /usr/libexec/astra/astra-priv`.

### Hasil uji helper di VM (2026-10-26)

`scripts/vm/test-priv.sh` dan `scripts/vm/test-workspace.sh` di VM Ubuntu 24.04 (systemd 255): **semua pemeriksaan lulus.**

| Uji | Hasil |
| --- | --- |
| `workspace-add` | User `astra-agent-<slug>` hanya di grupnya sendiri; folder `710`, `2770`, `700`, `700` dengan pemilik sesuai ADR ini; sudoers `0440 root` dan lolos `visudo`; diulang tidak membuat ulang user |
| Hak user `astrad` | Hanya `astra-priv` sebagai root; `id`, `systemd-run`, dan `useradd` sebagai root ditolak; tidak bisa menjalankan perintah sebagai user lain selain user agent |
| User agent | Tidak bisa memanggil helper, tidak bisa membaca rahasia daemon maupun `work/` workspace lain |
| 13 permintaan jahat (`x;id`, `../etc`, agent `bash`, `--dir ../../..`, `--dir /etc`, `--env`, memori di luar batas, dan lain-lain) | Semua ditolak dengan exit 2 sebelum mengubah apa pun; daftar user tidak berubah |
| `session-start` | Unit aktif di `astra-agents-<slug>.slice` dengan `MemoryMax`, `CPUQuota`, `TasksMax`, `NoNewPrivileges`; host berjalan sebagai user agent di folder yang diminta; daemon tersambung lewat socket |
| Environment agent | Hanya delapan variabel dari helper ditambah yang selalu ditambahkan systemd (`INVOCATION_ID`, `JOURNAL_STREAM`, `MEMORY_PRESSURE_WATCH`, `MEMORY_PRESSURE_WRITE`, `SYSTEMD_EXEC_PID`); tidak ada variabel daemon, sudo, atau SSH |
| `--dir` yang berupa symlink keluar dari `work/`, sesi ganda | Ditolak |
| User agent ditambahkan ke grup `docker` | `workspace-add` menolak dan menyebut grupnya; berjalan lagi setelah grup dilepas |
| Agent menukar `sessions/` dengan symlink ke `/etc/ssh` | `workspace-add` gagal, `/etc/ssh` tidak berubah |
| Potongan 2 dari ujung ke ujung | `astrad workspace add` membuat user lewat helper, clone lokal dan HTTPS berjalan sebagai user agent, stack Next.js terdeteksi, dua worker mendapat worktree dan branch sendiri, sesi shell dimulai di worktree worker |

Satu temuan dari uji ini: git menolak membaca repository milik user lain (`safe.directory`), jadi remote `file://` harus milik user agent workspace itu. Remote HTTPS dan SSH tidak terpengaruh.

Daemon tetap memakai aturan sudoers per-workspace untuk menjalankan git dan perintah lain sebagai user agent (klon, `git worktree`, diff); itu yang dibuat `workspace-add`, bukan baris `systemd-run` yang lebar dari uji sebelumnya.

## Keputusan

Isolasi user per workspace klien, sesi sebagai unit transient, dan helper `astra-priv` untuk semua hak istimewa dipakai untuk M1. Temuan terbuka 2 dan 3 (pembaca transcript di host, dan pengerasan `/proc` serta jaringan) diselesaikan sebelum fitur worker dan Ship dibangun di atasnya.
