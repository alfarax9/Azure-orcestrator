# ADR 0007: Onboarding, pemasangan server, dan checklist pertama kali

| Hal | Isi |
| --- | --- |
| Status | **Diusulkan** (2026-10-29); berlaku sampai ditolak pemilik produk |
| Pemutus | Maulana Alfara Salim |
| Pengusul | Claude Code |
| Milestone | M1, potongan 6 |
| Terkait | PRD F1-02 sampai F1-07, §8, N-12, N-13; ADR 0005, 0006; `docs/monorepo.md` |

## Konteks

Daemon sudah bisa dijalankan (potongan 4) tetapi pemasangannya masih manual: `setup-isolation.sh`, salin binary, tulis unit systemd. PRD §8.3 meminta app desktop melakukannya lewat SSH dalam satu klik, dengan pemeriksaan kesiapan, persetujuan sebelum perubahan, pairing lewat token yang tidak pernah tampil, dan checklist pertama kali di UI. App desktop (Wails) belum ada, dan logika ini tidak boleh menunggunya.

## Keputusan

### 1. Logika pemasangan di sisi client, satu paket, dua cara menjalankan

`internal/server` (capability) berisi pemeriksaan kesiapan, rencana pemasangan, dan pairing. Ia hanya bicara lewat interface `Runner` (jalankan perintah, kirim stdin, baca stdout dan kode keluar). Ada dua Runner: **SSH** (`internal/remote`, untuk app desktop dan CLI di laptop) dan **lokal** (untuk `astra install` yang dijalankan langsung di VPS, F1-06). Kedua jalur menjalankan langkah yang sama persis, jadi yang diuji dan yang dipakai pengguna tidak berbeda.

Pemasangan dipakai lebih dulu lewat CLI `astra` (`cmd/astra`). App desktop nanti memakai paket yang sama; aturan `desktop-is-a-client` di `.golangci.yml` dan `docs/monorepo.md` diubah supaya `internal/server` dan `internal/remote` boleh diimpor desktop, karena merekalah yang membuat daemon ada sebelum `internal/api/client` bisa dipakai.

### 2. SSH

`golang.org/x/crypto/ssh` (sudah disetujui di `docs/go-rules.md`) dan `golang.org/x/term` (satu-satunya tambahan, hanya untuk membaca passphrase kunci tanpa menampilkannya; dependensi langsung sekarang 5 dari maksimal 25). Kunci ed25519; kunci baru dibuat oleh `astra key new` yang mencetak public key untuk ditambahkan ke VPS (F1-02). Kunci yang dilindungi passphrase didukung lewat prompt.

- **Host key: tanya dulu, ingat, jangan percaya diam-diam.** Koneksi pertama ke host baru berhenti dengan sidik jari (SHA-256) dan baru dilanjutkan setelah pengguna mengonfirmasi; sidik jari itu disimpan di `known_hosts` milik Astra. Host yang sidik jarinya **berubah** ditolak tanpa pilihan untuk melanjutkan; pengguna harus menghapus entrinya sendiri setelah memastikan penyebabnya.
- Tidak memakai `ssh-agent` dan tidak membaca `~/.ssh/config` untuk saat ini; keduanya bisa ditambah tanpa mengubah antarmuka.

### 3. Berkas client

`~/.config/astra/` (atau `$XDG_CONFIG_HOME/astra`): `servers.json` (nama, host, port, user, path kunci, peran, token perangkat), `known_hosts`, `keys/`. Semua berkas mode `0600`, folder `0700`. Token perangkat memang tersimpan di sini; keychain OS menggantikannya di app desktop.

### 4. Pemeriksaan kesiapan (F1-03)

Setiap pemeriksaan menghasilkan `pass`, `warn`, atau `fail`, dengan penyebab dan cara memperbaiki. Memeriksa: OS (Ubuntu 22.04/24.04, Debian 12), arsitektur (amd64, arm64), systemd dan cgroup v2 (batas resource sesi, ADR 0005), akses root atau `sudo` tanpa password, RAM, CPU, disk, port 80 dan 443 bebas, dan `curl`. Kekurangan resource dari minimum PRD (2 vCPU, 4 GB, 40 GB) adalah `warn`, kecuali sangat kecil (RAM di bawah 1,75 GiB atau disk di bawah 10 GB), yang `fail`. Pemasangan menolak berjalan bila ada `fail`.

### 5. Pemasangan (F1-04)

Rencana ditampilkan lebih dulu dan hanya dijalankan setelah pengguna menyetujui (`--yes` untuk skrip). Sumber binary: `--binary-dir` (file `astrad` dan `astra-priv` untuk arsitektur server, dikirim lewat SSH) atau `--release-url` (server mengunduh `astra-<arch>.tar.gz` dan `SHA256SUMS` dan memverifikasinya). Dalam kedua kasus SHA-256 diverifikasi **di server** sebelum apa pun dipasang. Tanda tangan rilis (N-13) belum diverifikasi; itu bagian pipeline rilis.

Langkah (idempoten, boleh diulang untuk update): pasang binary (`/usr/local/bin/astrad`, `/usr/libexec/astra/astra-priv`), buat grup dan user `astrad` (tanpa shell, home `/var/lib/astrad` mode `0700`), folder `/var/lib/astra/agents`, aturan sudo `astrad` untuk `astra-priv` (dicek `visudo`), unit `astrad.service` (`User=astrad`, `Restart=on-failure`), `systemctl enable --now`, lalu tunggu `GET /api/v1/version` menjawab. Update mengganti binary dan me-restart service; sesi agent berjalan sebagai unit sendiri dan tidak terputus (ADR 0005). Docker dipasang hanya dengan `--docker` (persetujuan eksplisit), lewat paket distro.

### 6. Pairing (F1-05)

Melalui SSH: `astrad pair` dijalankan sebagai user `astrad` dan token dibaca dari stdout oleh proses, tidak pernah dicetak. Client membuka tunnel SSH ke `127.0.0.1:7437` di server (N-12), menukar token ke token perangkat lewat `POST /api/v1/pairing`, dan menyimpannya di `servers.json`. Token pairing kedaluwarsa setelah dipakai (ADR 0006). Untuk F1-06, `astra install` di VPS mencetak token pairing satu kali untuk web UI; itu satu-satunya tempat token tampil, karena tidak ada pembaca lain.

### 7. Daemon melayani web UI dan checklist

- `ASTRA_WEB_DIR` menunjuk folder hasil `npm run build`; daemon melayaninya dari origin yang sama dengan API (dengan `index.html` sebagai fallback), jadi tunnel cukup satu port. Meng-embed folder itu ke binary menunggu pipeline rilis.
- `GET /api/v1/checklist` mengembalikan langkah pertama kali (F1-07): server (selalu selesai bila UI terhubung), login agent, project pertama, dan domain. **Domain belum bisa diselesaikan sebelum fitur preview ada**, jadi ia tampil sebagai "menyusul" dan belum dihitung dalam persentase maupun dalam syarat checklist hilang; persentase menghitung langkah yang bisa diselesaikan sekarang.
- **Login agent** diketahui dari keberadaan berkas kredensial agent di home user agent (`~/.claude/.credentials.json`, `~/.codex/auth.json`), bukan isinya. Daemon tidak boleh membukanya (F3-03), jadi pengecekan itu verb keempat `astra-priv`, `agent-status <slug>`, yang hanya mengembalikan dua boolean. Di mode lokal (pengembangan) pengecekan yang sama dilakukan pada home pengembang. Letak berkas itu adalah asumsi tentang versi agent saat ini; bila salah, hasilnya "belum login", yang aman.

## Konsekuensi

- Satu binary `astra` baru ditambahkan ke rilis; ukuran `astrad` tidak berubah, karena kode SSH hanya ada di `cmd/astra`.
- Tidak ada layar tambah server di web UI: menambah server butuh SSH, yang tidak bisa dilakukan peramban. Layar itu ada di app desktop memakai paket yang sama.
- Pembaruan daemon otomatis, uninstall (§8.6), dan verifikasi tanda tangan rilis belum termasuk.

## Hasil (2026-10-29)

Dibangun: `internal/remote` (SSH, host key, kunci, tunnel), `internal/server` (kesiapan, rencana pemasangan, pairing, daftar server), `internal/checklist`, verb `agent-status` di `astra-priv`, `ASTRA_WEB_DIR` dan `GET /api/v1/checklist` di daemon, CLI `astra`, dan di web: checklist "Get started" serta dialog Add project.

| Uji | Hasil |
| --- | --- |
| `go test` | `internal/remote` melawan server SSH dalam proses (host key tidak dikenal berhenti dengan sidik jari, yang berubah ditolak, kunci tak dikenal, kunci dengan passphrase, tunnel yang menutup koneksi terbukanya); `internal/server` dengan Runner palsu (setiap kegagalan kesiapan punya cara memperbaiki, rencana ditolak bila ada yang gagal atau checksum tidak cocok, urutan perintah, folder sementara selalu dihapus termasuk saat dibatalkan, token tidak pernah masuk galat) |
| `scripts/vm/test-install.sh` | 58 pemeriksaan dari Mac ke VM lewat SSH sungguhan, tanpa instalasi sebelumnya: kunci baru, tanya sebelum memercayai host, kesiapan, instalasi (service aktif dan enabled, daemon bukan root, aturan sudo hanya untuk `astra-priv`), pairing (tidak ada token di layar, di berkas selain `servers.json` mode 600, atau di journal), tunnel, web UI dengan CSP, checklist 33% lalu 66% dari fakta, update dengan `.previous`, unduh rilis (http ditolak tanpa opt-in, rilis yang dirusak ditolak dan tidak ada yang terpasang), host key berubah ditolak dengan peringatan, dan daemon hidup lagi dengan datanya setelah VM di-reboot |
| Browser | Checklist 33%, dialog Add project (klien baru, nama folder dan user terlihat), 66% tanpa reload, dengan daemon lokal |

Satu hal yang diperbaiki dari uji: `HEAD /` ke web UI dijawab 404 karena hanya GET yang dilayani.

**Belum ada:** F1-01 (installer bertanda tangan), web UI yang di-embed ke binary, verifikasi tanda tangan rilis (N-13), layar tambah server di app desktop, uninstall dan pembaruan otomatis (§8.5, §8.6), `ssh-agent` dan `~/.ssh/config`, serta uji di VPS sungguhan dan Debian 12. Letak berkas login agent (`~/.claude/.credentials.json`, `~/.codex/auth.json`) belum dicocokkan dengan agent sungguhan di Linux; di Mac mode lokal membaca home pengembang, tempat Claude Code memakai keychain, sehingga hasilnya "belum login".
