# ADR 0006: API daemon, pairing, dan WebSocket

| Hal | Isi |
| --- | --- |
| Status | **Diusulkan** (2026-10-29), sudah dibangun; berlaku sampai ditolak pemilik produk |
| Pemutus | Maulana Alfara Salim |
| Pengusul | Claude Code |
| Milestone | M1, potongan 4 |
| Terkait | PRD F1-05, F3-02, N-01, N-02, N-11, N-12, §7 (API v1); ADR 0003, 0005; `docs/go-rules.md` |

## Konteks

Potongan 4 membuat daemon `astrad` menjadi proses yang melayani client: aliran perubahan fakta (`GET /api/v1/events`), terminal agent dua arah (`GET /api/v1/sessions/{id}/pty`), objek dasar, dan pairing lewat token sekali pakai. Beberapa hal belum diputuskan oleh PRD: siapa yang mengenkripsi lalu lintas, seperti apa token, bentuk pesan WebSocket, dan dari mana pustaka WebSocket berasal (pustaka standar Go tidak punya server WebSocket).

## Keputusan

### 1. Daemon hanya mendengarkan di loopback; enkripsi ada di luar daemon

PRD N-12 meminta komunikasi "terenkripsi (SSH tunnel atau tsnet), terautentikasi dengan token pairing". Daemon mendengarkan di `127.0.0.1:7437` (bisa diubah, tetapi alamat non-loopback ditolak kecuali `ASTRA_ALLOW_REMOTE=1`) dan **tidak membuat TLS sendiri**. Desktop app membuka SSH tunnel ke port itu; akses dari browser atau HP (F1-06) lewat proxy TLS atau tsnet di M3, yang meneruskan ke port yang sama. Dengan begitu tidak ada port API yang terbuka ke internet secara tidak sengaja.

### 2. Pairing dan token perangkat

- `astrad pair` dijalankan lewat SSH oleh desktop app. Ia membuat **token pairing** (32 byte acak, base64url), menyimpan hanya **SHA-256**-nya dengan batas waktu 10 menit, dan mencetak token itu ke stdout, yang dibaca app dan tidak pernah ditampilkan (F1-05).
- `POST /api/v1/pairing {token, name}` menukar token itu dengan **token perangkat** (32 byte acak). Penukaran bersifat atomik: token pairing dihapus pada pemakaian pertama, jadi kedaluwarsa setelah dipakai. Token perangkat disimpan sebagai SHA-256 di tabel `devices`; token mentahnya hanya ada di response itu.
- Semua endpoint selain `GET /api/v1/version` dan `POST /api/v1/pairing` meminta `Authorization: Bearer <token perangkat>`. Perbandingan memakai hash dan `crypto/subtle`. Browser tidak bisa memasang header pada `new WebSocket`, jadi WebSocket juga menerima subprotocol `bearer.<token>`; server menjawab `astra.v1`, tidak pernah menjawab token.
- `astrad devices` menampilkan perangkat, dan `astrad devices revoke <id>` mencabutnya. Token tidak pernah ditulis ke log, dan bertipe `domain.Secret` selama di memori.
- Karena autentikasi memakai header, bukan cookie, CSRF tidak berlaku. CORS hanya mengizinkan origin yang terdaftar di `ASTRA_ORIGINS` (untuk dev server web); WebSocket memeriksa `Origin` terhadap daftar yang sama.

### 3. Aliran perubahan fakta

`GET /api/v1/events` (WebSocket) mengirim pesan JSON teks. Pesan pertama adalah `{"type":"snapshot","data":{"workspaces":[…],"projects":[…],"workers":[…]}}`; setelah itu setiap perubahan fakta menjadi satu pesan `{"type":"workspace"|"project"|"worker","data":{…}}` berisi objek lengkap. **Pesan menyimpan fakta, bukan status**; client menurunkan status dengan aturan yang sama (`contract/status-cases.json`).

- Server mendaftar sebelum membaca snapshot, jadi tidak ada perubahan yang hilang; client menggantikan objek dengan `id` yang sama, jadi pesan ganda tidak berbahaya.
- Perubahan datang dari `store`: setiap penulisan yang berhasil memberi tahu pendengar, sehingga tidak ada layanan yang bisa lupa mengirim.
- Client yang lambat (antrean 256 pesan penuh) diputus, dan client menyambung ulang dengan snapshot baru. Tidak ada antrean tak terbatas.

### 4. Terminal agent

`GET /api/v1/sessions/{id}/pty` (`id` = id worker). Frame biner dari client adalah input terminal, frame biner dari server adalah output agent (scrollback dikirim lebih dulu saat tersambung). Frame teks dari client: `{"type":"resize","rows":N,"cols":M}`. Frame teks dari server: `{"type":"exit","code":N}`. Input lewat jalur ini memberi tahu layanan agent, sehingga status "Needs you" hilang saat pengguna menjawab dialog dari terminal.

### 5. Pustaka WebSocket: `github.com/coder/websocket`

Dependensi baru (satu dari maksimal 25). Alasan: tanpa dependensi lain, mendukung `context`, kompresi opsional, batas ukuran pesan, pemeriksaan `Origin`, dan subprotocol; dirawat aktif. Menulis RFC 6455 sendiri (masking, fragmentasi, close handshake, ping) berarti merawat kode keamanan yang tidak membedakan produk ini. Versi dikunci di `go.mod` dengan checksum.

### 6. Bentuk

Semua endpoint di bawah `/api/v1`. Galat berbentuk `{"error":{"code":"not_found","message":"…"}}` dengan status 400, 401, 404, 409, atau 500; pesan galat internal tidak dikirim ke client, hanya dicatat. Body dibatasi 1 MiB dan field yang tidak dikenal ditolak. `GET /api/v1/version` mengembalikan versi daemon dan versi API (N-16).

## Konsekuensi

- Tidak ada akses jarak jauh tanpa tunnel atau proxy; M3 menambahkan proxy dan tsnet.
- Token perangkat tidak kedaluwarsa sendiri; pencabutan manual lewat `astrad devices revoke`. Kedaluwarsa dan UI-nya masuk F11.
- Token pairing 256 bit tidak bisa ditebak, jadi tidak ada pembatasan percobaan. Kode pairing pendek untuk web UI (F1-06) membutuhkan pembatasan dan diputuskan di ADR sendiri.

## Hasil (2026-10-29)

Dibangun: `internal/auth` (pairing dan token perangkat), `internal/api` (REST, `events`, terminal), notifikasi perubahan di `internal/store`, `agent.Supervisor`, dan `astrad serve|pair|devices`. Dependensi baru hanya `github.com/coder/websocket` v1.8.15 (dependensi langsung sekarang 3 dari maksimal 25).

Rute: `GET /api/v1/version`, `POST /api/v1/pairing` (tanpa token); dengan token perangkat: `GET|POST /workspaces`, `GET|POST /projects`, `PUT /projects/{id}/commands`, `GET /workers`, `GET /workers/{id}`, `GET /usage?days=30` (token per hari, agent, dan model; F14-04), `POST /projects/{id}/workers`, `POST /workers/{id}/start|stop|interrupt|messages`, `GET /events` (WebSocket), `GET /sessions/{id}/pty` (WebSocket). `POST /workers/{id}/start` menjawab **202** (bukan 200) dengan worker berstatus Queued bila server sudah menjalankan sebanyak `ASTRA_MAX_WORKERS` worker (F10-05, ADR 0008); worker itu dimulai sendiri saat satu slot kosong. Cara memakai dari nol: `astrad pair` di server (lewat SSH) mencetak token, lalu `POST /pairing {"token":"…","name":"Laptop"}` mengembalikan token perangkat.

| Uji | Hasil |
| --- | --- |
| `go test` (`internal/api`, 87,6%) | Setiap rute tertutup tanpa token; token pairing dan perangkat tidak masuk log; semua jenis galat punya status dan kode sendiri, galat internal hanya ke log; body dibatasi dan field asing ditolak; CORS dan origin WebSocket hanya untuk daftar yang diizinkan; token lewat subprotocol hanya diterima di rute WebSocket dan tidak pernah dipantulkan; terminal meneruskan byte dua arah, resize, dan exit |
| `go test` (`cmd/astrad`) | Satu daemon sungguhan di loopback dengan agen tiruan: pairing sekali pakai, objek dibuat lewat API, worker muncul di aliran `events`, "Needs you" terlihat, dialog dijawab lewat WebSocket terminal, tugas dikirim lewat REST, giliran selesai dengan `+2` muncul di aliran |
| `scripts/vm/test-serve.sh` | `astrad serve` sebagai layanan systemd milik `astrad`: hanya mendengarkan 127.0.0.1, alamat non-loopback ditolak, pairing, token tidak ada di journal, worker dan agent lewat API, daemon di-restart dan sesi agent tidak tersentuh, daemon baru melanjutkan pengawasan, pencabutan perangkat langsung efektif, SIGTERM menghentikan daemon |

Satu galat nyata ditemukan uji VM: `session-start` kembali begitu unit dimulai, padahal host baru membuka socket beberapa milidetik kemudian, sehingga watcher pertama melihat worker "tanpa sesi". Sekarang daemon menunggu socket menjawab sebelum `Start` selesai (`pty.WaitReady`).

Belum diuji: klien browser sungguhan (xterm.js dan `new WebSocket`, potongan 5), client yang lambat sampai antreannya penuh (logikanya ada, belum ada uji yang memenuhi antrean), dan akses dari luar lewat tunnel SSH (M3).

## Klien web (potongan 5, 2026-10-29)

- **Alamat.** Halaman dan API satu origin: di pengembangan Vite meneruskan `/api` (termasuk WebSocket) ke `127.0.0.1:7437`, dan daemon harus punya `ASTRA_ORIGINS=http://localhost:5173` karena proxy meneruskan `Origin` browser; saat produksi daemon melayani halamannya sendiri. `scripts/dev/local-daemon.sh` menyiapkan semuanya (data sementara, agen tiruan atau `REAL_AGENT=1`, satu workspace dan project, token pairing).
- **Token.** Token perangkat disimpan di `localStorage` supaya reload tidak meminta pairing lagi. Risikonya adalah XSS membaca token; karena itu halaman tidak memuat skrip pihak ketiga, dan app desktop nanti menyimpannya di keychain OS. Peramban yang memblokir storage tetap bekerja selama tab hidup. Token pairing dimasukkan lewat kolom tersembunyi dan dikosongkan setelah dipakai.
- **Koneksi.** `GET /events` disambung ulang dengan jeda 1, 2, 4, sampai 15 detik. Peramban tidak bisa membaca alasan handshake gagal, jadi setelah putus klien bertanya ke `GET /workspaces`: 401 berarti token dicabut dan layar kembali ke pairing dengan penjelasan; selain itu daemon dianggap tidak terjangkau. Saat tidak tersambung, data terakhir tetap tampil dengan jamnya di status bar dan New task dinonaktifkan dengan alasannya.
- **Terminal.** `/sessions/{id}/pty` dibuka dengan xterm.js (dimuat malas, 86 KB gzip, hanya di layar worker). Koneksi yang gagal sebelum pernah terbuka berarti tidak ada sesi: terminal berhenti dan menawarkan Start agent dan Retry. Koneksi yang putus setelah terbuka disambung ulang, dan layar dilukis ulang dari scrollback host. Ctrl+J memfokuskan terminal, Shift+Esc keluar darinya (tanpa jebakan keyboard), dan ANSI memakai palet `DESIGN_SYSTEM.md` §2.3.
- **Mode contoh.** `/?sample` menampilkan Board dengan data contoh tanpa daemon (dipakai inspektur taste-skill); `/?sim` menambah simulator fakta.

Hasil: uji Vitest 122 lulus (termasuk daemon palsu lewat WebSocket palsu di `App.live.test.tsx`), `lint.mjs --app` 0 FAIL, `inspect.mjs --app` 0 FAIL di `/?sample` dan layar pairing. Di browser dengan daemon sungguhan: pairing, New task, dialog trust dijawab di terminal, tugas dikirim, `+2 −0` muncul di kartu dan layar worker tanpa reload, daemon dimatikan dan dihidupkan (data tetap, status bar berubah, New task menunggu, tersambung sendiri), reload tidak meminta pairing lagi, dan 390 px (kolom jadi tab, bottom bar).
