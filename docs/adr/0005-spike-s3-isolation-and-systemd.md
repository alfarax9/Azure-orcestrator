# ADR 0005: Spike S3 dan systemd, isolasi agent dan umur sesi

| Hal | Isi |
| --- | --- |
| Status | **Sebagian** (2026-10-26): terbukti di VM Ubuntu 24.04.5 arm64 (QEMU/HVF di macOS); belum di VPS fisik, Debian 12, atau dengan Claude Code dan Codex sungguhan di dalam VM |
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

1. **Satu user agent untuk semua worker.** Agent untuk klien A dan klien B berjalan sebagai user yang sama, sehingga bisa membaca worktree satu sama lain dan memakai login agent yang sama. Ini bertentangan dengan semangat F2-01 (project klien A tidak terlihat dari klien B). **Perlu keputusan sebelum worker M1:** (a) satu user per workspace klien (`astra-agent-<workspace>`, tetapi login agent harus diulang per user), (b) sandbox per worker (bubblewrap/user namespace) di atas satu user, atau (c) container rootless per worker. Uji ini tidak menyentuh pertanyaan tersebut.
2. **Daemon tidak boleh membaca home agent.** Transcript untuk S2 ada di `~/.claude/projects` dan `~/.codex/sessions`, satu folder dengan kredensial login agent (F3-03: daemon tidak boleh membaca atau menyalin kredensial). Claude Code membuat file transcript dengan mode `0600`, jadi izin grup atau ACL tidak membantu. **Usul:** pembaca transcript S2 (`internal/agent`) berjalan di dalam session host, yang sudah berjalan sebagai user agent, dan mengirim event ke daemon lewat frame baru di socket yang sama. Daemon tidak pernah membuka file di home agent.
3. **Umum dan harus dijaga:** (i) `/proc` dipasang tanpa `hidepid`, sehingga agent **melihat argumen** proses user lain (diuji: argumen `--token=…` milik `astrad` terlihat lewat `ps`). Secret tidak boleh pernah dikirim lewat argv, dan pada VPS `/proc` sebaiknya dipasang dengan `hidepid=2`. (ii) Agent punya akses internet penuh (wajar) dan bisa menjangkau layanan di localhost, termasuk API daemon nanti; token pairing tidak boleh ada di environment agent, dan aturan nftables per-UID (`meta skuid`) perlu dipertimbangkan untuk M3. (iii) Alamat metadata cloud 169.254.169.254 tidak terjangkau di VM ini; di VPS tertentu bisa terjangkau dan harus diblokir untuk UID agent.
4. **Hak untuk membuat unit.** Varian D diuji dengan `astrad ALL=(root) NOPASSWD: /usr/bin/systemd-run`, yang terlalu lebar untuk produksi (siapa pun yang menjadi `astrad` bisa menjalankan apa saja sebagai root). Opsi yang harus dipilih di M1 dan diuji di VM yang sama: (a) aturan polkit untuk `org.freedesktop.systemd1.manage-units` yang hanya menerima unit `astra-s-*`, (b) unit template `astra-session@.service` yang dimulai lewat `sudo systemctl start` dengan sudoers sempit dan parameter lewat `EnvironmentFile`, atau (c) helper kecil milik root yang memvalidasi permintaan. Rekomendasi awal: (a).
5. **Red-team dengan agent sungguhan belum dijalankan.** Perintah yang sama dengan prompt PRD ("baca .env production", "jalankan docker ps") sudah diuji di OS yang sama, user yang sama, dan hasilnya tidak bergantung pada agent. Uji dengan Claude Code dan Codex di dalam VM (butuh login di VM dan kuota) menjadi uji regresi di setiap rilis (F12-01, metrik PRD §13), dengan prompt sebagai teks dan hasil periksa lewat transcript S2.
6. **Binary `astrad` harus bisa dieksekusi oleh user agent** (`/usr/local/bin`, mode 0755), dan folder socket harus dibuat oleh installer (F1-04), bukan oleh daemon, karena daemon tidak boleh memiliki folder yang ditulis agent.

## Belum terbukti

1. VPS fisik dan Debian 12 (N-15): mekanismenya sama, tetapi belum dijalankan.
2. Uji stabil 2 jam dengan 5 sesi (F3-02) dan angka RAM/CPU: butuh VPS 4 GB sungguhan.
3. Penegakan batas memori dan CPU (S6) dan dampaknya ke latensi aplikasi production.
4. Pilihan di temuan 1 dan 4.

## Keputusan

Isolasi user plus sesi sebagai unit transient dipakai untuk M1, dengan temuan terbuka 1 dan 4 diselesaikan sebelum fitur worker dan Ship dibangun di atasnya.
