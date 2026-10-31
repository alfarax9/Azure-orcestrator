# ADR 0010: Login akun dengan Google dan GitHub, email lewat Resend

| Hal | Isi |
| --- | --- |
| Status | **Diusulkan** (2026-10-31); belum ada yang dibangun |
| Pemutus | Maulana Alfara Salim |
| Pengusul | Claude Code |
| Milestone | Setelah M3 (butuh akses web lewat proxy, F1-06); mengubah PRD §2 (fitur tim), F1-05, F1-06 |
| Terkait | PRD F1-05, F1-06, F5-03, F6-02, F12-02, N-11, N-12, N-18, §10, §15; ADR 0006 (pairing dan token perangkat), ADR 0008 §4 (penyimpanan key); `docs/go-rules.md` |

## Konteks

Pemilik produk meminta login akun user memakai **akun Google dan GitHub**, dengan **Resend** untuk email (login dan reset password), dan menyebut project ini sudah terhubung ke **Supabase**.

Keadaan sekarang, dari kode dan PRD:

- Satu-satunya autentikasi adalah **token perangkat** hasil pairing lewat SSH (ADR 0006). Tidak ada akun, tidak ada password, tidak ada email.
- Tidak ditemukan konfigurasi atau kode Supabase, Resend, Google, atau GitHub di repo (berkas `.env` di-`.gitignore`). Project Supabase yang dimaksud belum terlihat dari sini.
- Fitur tim (role, undangan, audit per anggota) ada di fase 2 (PRD §2). Model bisnis, termasuk fitur berbayar untuk tim dan relay terkelola, belum diputuskan (PRD §15, sebelum M5).
- F12-02 dan N-12: dashboard daemon tidak publik; yang publik hanya preview, production, dan `POST /public/v1/feedback/{share}`. Klien preview tidak membuat akun (F6-02), dan tautan preview punya password sendiri yang opsional (F5-03).
- Produk ini **dipasang sendiri di VPS milik tiap user**. Tiap instalasi punya domain sendiri, dan tidak ada layanan pusat milik Astra (PRD §10, N-18).

Dua fakta itu menentukan bentuk keputusan ini: login publik bertabrakan dengan F12-02, dan satu aplikasi OAuth tidak bisa mendaftarkan domain semua instalasi.

## Keputusan yang diusulkan

### 1. Tujuan dan batas

- Akun memberi **identitas** pada orang yang memakai Astra (siapa yang menyetujui deploy, siapa yang mencabut perangkat), bukan akses ke akun Claude, Codex, atau repo.
- Login Google atau GitHub **tidak pernah dipakai** untuk memanggil Anthropic, OpenAI, atau GitHub atas nama user, dan token penyedia **tidak disimpan**: dipakai sekali untuk membaca identitas, lalu dibuang (PRD §10, N-11).
- Klien preview tetap tanpa akun (F6-02). Akun tidak mengubah tautan berbagi.
- Tidak ada jalur deploy production tanpa approval manusia (`CLAUDE.md`); dengan akun, approval itu tercatat atas nama anggota.

### 2. Tanpa Supabase untuk login (usulan; pertanyaan 1)

Daemon tetap sumber kebenaran: akun disimpan di SQLite daemon, di server milik user. Supabase tidak dipakai untuk login karena:

- Satu project Supabase milik pemilik produk untuk semua instalasi berarti **layanan pusat** yang memegang identitas semua user, satu titik gagal, dan bertentangan dengan N-18 serta dengan produk yang dipasang sendiri.
- Satu project Supabase per user menambah langkah pasang yang berat bagi persona freelancer.
- Daemon sudah punya token perangkat, tabel `devices`, dan kunci enkripsi (ADR 0008 §4); akun hanya menambah tabel.

Supabase tetap cocok untuk hal di luar daemon (situs produk, penagihan, daftar tunggu). Bila pemilik produk memutuskan penawaran **berbayar terkelola** (PRD §15), akun terpusat di Supabase menjadi ADR tersendiri; keputusan di sini tidak menutup jalan itu.

### 3. Alur OAuth: device flow dari daemon (usulan; pertanyaan 2)

Daemon yang meminta login, bukan browser yang memanggil daemon:

1. UI menekan "Masuk dengan Google" atau "Masuk dengan GitHub"; daemon meminta **kode perangkat** ke penyedia (RFC 8628) dan UI menampilkan kode serta tautan verifikasi penyedia.
2. User membuka tautan itu di browser mana pun, memasukkan kode, dan menyetujui.
3. Daemon menanyai penyedia sampai disetujui, membaca identitas (Google: ID token OIDC; GitHub: `GET /user` dan `GET /user/emails`, hanya email utama yang terverifikasi), lalu membuang token penyedia.
4. Daemon membuat **token perangkat** (ADR 0006) yang terikat ke akun, dan UI mengambilnya lewat penukaran sekali pakai yang sama dengan pairing.

Akibatnya:

- Daemon hanya **menelepon keluar**; tidak ada rute auth publik baru, jadi F12-02 dan N-12 tetap utuh.
- Tidak ada redirect URI, jadi tidak perlu mendaftarkan domain tiap instalasi, dan berfungsi di VPS tanpa domain.
- Token tetap di header `Authorization`, jadi API dan WebSocket tidak berubah, dan CSRF tetap tidak berlaku.

Biaya: user mengetik kode di tab kedua. Itu wajar untuk developer dan agency, tetapi lebih kasar daripada redirect. Pilihan lain ada di pertanyaan 2. Dukungan device flow untuk scope yang dibutuhkan harus **diverifikasi di spike S11** sebelum ADR ini menjadi final.

Scope minimal: Google `openid email profile`; GitHub `read:user user:email`. **Tanpa** scope repo.

### 4. Siapa yang boleh masuk (usulan; pertanyaan 3)

Berhasil login bukan berarti diizinkan.

- Pemilik pertama dibuat lewat jalur yang sudah ada: `astrad pair` lewat SSH membuktikan akses ke server, lalu perangkat itu menautkan akun pertama sebagai `owner`.
- Akun lain masuk lewat **undangan** (email, peran, kedaluwarsa); tanpa undangan, login ditolak dengan pesan netral. Peran dan audit lengkap adalah fase 2; di sini hanya `owner` dan `member`.
- Kunci akun adalah **penyedia + id subjek**, bukan email, karena email bisa berubah. Email hanya untuk undangan dan tampilan.
- `astrad accounts` dan `astrad devices revoke` tetap bisa dipakai lewat SSH untuk memulihkan akses bila penyedia atau email bermasalah.

### 5. Tanpa password dan tanpa reset (usulan; pertanyaan 4)

Dengan Google dan GitHub tidak ada password untuk disimpan atau di-reset. Login email-password menambah penyimpanan hash, alur reset, dan permukaan serangan, tanpa kebutuhan yang jelas untuk persona ini. Bila pemilik produk tetap menginginkannya, hash memakai Argon2id (`golang.org/x/crypto`, modul yang sudah disetujui untuk SSH), dan itu perlu persetujuan eksplisit di ADR ini.

### 6. Resend (usulan; pertanyaan 5)

Dipakai untuk email transaksional dari daemon lewat `net/http` (tanpa SDK, tanpa dependensi baru):

- undangan anggota (token undangan sekali pakai, hanya hash-nya disimpan, kedaluwarsa 7 hari);
- peringatan "perangkat baru masuk ke akun Anda".

Tidak untuk email pemasaran, tidak untuk reset password (tidak ada password).

Penanganan kunci (berlaku sekarang): kunci Resend **tidak pernah** ada di repo, di `web/`, di log, atau di lingkungan agent. Daemon membacanya dari `/etc/astra/resend.key` (`root:astrad`, `0640`, seperti `secret.key` di ADR 0008 §4), bertipe `domain.Secret`. Kunci memakai izin **sending access** yang dibatasi ke satu domain pengirim.

> **Tindakan segera.** Kunci Resend yang sempat ditempel di chat sesi pengembangan harus dianggap bocor: buat ulang di dashboard Resend dan hapus yang lama. Hal yang sama berlaku untuk rahasia lain yang ditempel di chat.

### 7. Sesi perangkat berakun

- `devices` mendapat kolom `account_id` (kosong untuk perangkat hasil pairing lama, yang tetap berlaku).
- Token perangkat berakun **kedaluwarsa** setelah 30 hari tanpa dipakai (usulan; pertanyaan 7) dan bisa dicabut dari UI dan dari `astrad devices revoke`.
- Percobaan login dibatasi per perangkat dan per IP; undangan dan penukaran kode dibatasi seperti pairing web di F1-06.
- Peristiwa login, cabut, dan undangan dicatat sebagai fakta (siapa, kapan, dari perangkat apa). Tidak ada token, kode, atau email lengkap di log.

### 8. Bentuk di kode (usulan, belum dibangun)

- `internal/auth` (capability) bertambah akun, undangan, dan penukaran kode. Ia mendefinisikan dua interface kecil: `Identity` (menjalankan device flow dan mengembalikan identitas terverifikasi) dan `Mailer` (mengirim satu email).
- Paket infrastruktur daun baru `oauth` (Google, GitHub) dan `mail` (Resend); `cmd/astrad` yang menyambungkan, sesuai `docs/go-rules.md`.
- Migrasi: tabel `accounts`, `invitations`, dan kolom `devices.account_id`.
- API: `POST /api/v1/auth/{provider}/start`, `POST /api/v1/auth/{provider}/poll`, `GET /api/v1/me`, `POST /api/v1/invitations`, `GET /api/v1/accounts`. Rute `start` dan `poll` boleh dipanggil tanpa token dan dibatasi rate; sisanya meminta token perangkat.

### 9. Perubahan PRD yang menyertai (setelah disetujui)

- Requirement baru (usulan ID): **F16-01** login Google dan GitHub dengan device flow (P1), **F16-02** undangan dan daftar akun lewat Resend (P1), **F16-03** sesi perangkat berakun: daftar, kedaluwarsa, cabut (P1), **F16-04** catatan audit login dan undangan (P2); peran dan audit lengkap tetap fase 2.
- Spike baru **S11**: device flow Google dan GitHub dengan scope minimal, verifikasi ID token, dan batas laju.
- F1-05 dan F1-06 dirujuk: pairing tetap jalur pemulihan; kode pairing web (F1-06) tetap, login akun menjadi cara kedua.
- §15: pertanyaan terbuka diperbarui sesuai bagian di bawah.

## Konsekuensi

**Positif**
- Identitas per anggota tanpa layanan pusat, tanpa password, dan tanpa membuka dashboard ke internet.
- Tidak ada dependensi Go baru (`net/http`, `crypto/*`).
- Pairing yang sudah ada tetap berfungsi dan menjadi jalur pemulihan.

**Negatif dan risiko**
- Login membutuhkan **akses web ke daemon** (proxy M3, F1-06), sehingga fitur ini tidak bisa dibangun sebelum itu. Untuk pemilik tunggal yang memakai desktop app lewat SSH, pairing sudah cukup; manfaat akun terutama ada pada akses dari HP atau browser dan pada tim.
- Pemilik produk harus membuat **aplikasi OAuth** di Google Cloud dan GitHub dengan nama "Astra" (tanpa kata "Claude"). Client ID ikut dalam binary; ia identitas aplikasi, bukan kredensial user.
- Device flow lebih canggung daripada redirect, dan dukungan Google untuk scope identitas harus dibuktikan (S11). Bila gagal, jalur cadangannya adalah redirect lewat relay (pertanyaan 2), dan itu butuh layanan pusat.
- Tabel akun menyimpan email dan nama; itu data pribadi di server user. Penghapusan akun menghapus barisnya, dan N-18 (tanpa telemetri) tidak berubah.

## Alternatif yang ditolak

- **Supabase Auth pusat untuk semua instalasi.** Layanan pusat dengan identitas semua user; bertentangan dengan produk yang dipasang sendiri (lihat §2).
- **Supabase Auth per instalasi.** Setiap user harus membuat dan menjaga project sendiri.
- **Redirect OAuth langsung ke daemon.** Butuh rute publik (mengubah F12-02) dan pendaftaran redirect URI per domain instalasi.
- **Login email-password dengan reset.** Menambah permukaan serangan tanpa kebutuhan jelas (lihat §5).
- **Memakai token GitHub untuk clone repo atau akses lain.** Mencampur identitas dengan akses; akses repo adalah keputusan terpisah.

## Pertanyaan terbuka

1. **Peran Supabase.** Usulan: tidak dipakai untuk login (hanya situs, penagihan, atau penawaran terkelola di masa depan). **Butuh dari pemilik produk:** project Supabase mana yang terhubung, dan untuk apa project itu dibuat.
2. **Alur OAuth.** Usulan: device flow dari daemon. Alternatif: redirect lewat relay milik Astra (lebih mulus, tetapi layanan pusat dan keputusan model bisnis).
3. **Siapa boleh masuk.** Usulan: pemilik pertama lewat `astrad pair`, anggota lain hanya lewat undangan; login Google atau GitHub saja tidak memberi akses.
4. **Password dan reset.** Usulan: tidak ada. Bila dibutuhkan, Argon2id di daemon.
5. **Resend.** Usulan: hanya undangan dan peringatan perangkat baru. **Butuh dari pemilik produk:** domain pengirim (SPF, DKIM, DMARC) dan siapa yang membuat kunci baru yang dibatasi ke domain itu.
6. **GitHub.** Usulan: identitas saja, tanpa scope repo; akses clone repo privat dibahas terpisah (PRD F2).
7. **Kedaluwarsa sesi.** Usulan: 30 hari tanpa dipakai, bisa dicabut kapan saja.
8. **Aplikasi OAuth.** Siapa pemilik akun Google Cloud dan organisasi GitHub-nya, dan apakah layar persetujuan Google perlu diverifikasi untuk pengguna publik.
9. **Model bisnis.** Apakah akun menjadi bagian fitur berbayar (tim, relay terkelola; PRD §15) atau tetap inti gratis. Usulan: inti gratis, dan akun pusat ditunda sampai keputusan bisnis.
