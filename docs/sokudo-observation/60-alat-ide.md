# A7. Alat di dalam IDE (catatan ringkas, belum lengkap)

Pengamatan Sokudo 0.19.6, 2026-10-31, hanya di jendela Sokudo. Tanda: **[terlihat langsung]**, **[inferensi]**. Bukti: berkas `a7-*.jpg` di `docs/sokudo-observation/screens/` (tidak di-commit).

## Dibuat dan dibersihkan
- Proyek Canvas "Untitled" (Design): dibuat lalu dihapus lewat menu klik-kanan (Rename, Delete).
- Sumber data `ATM-TEST-db` (SQLite in-memory): dibuat di Database Manager, lalu dihapus lewat klik kanan → Delete Source. Dialog konfirmasi menyebut yang dihapus hanya data milik Sokudo di perangkat (koneksi tersimpan, kredensial, skema ter-cache, riwayat query), bukan database-nya, dan "This cannot be undone". Keadaan akhir: "No sources yet". Menu klik-kanan sumber: Connect, Refresh, Open Console, Edit Source, Duplicate, Add to Favorites, Color, Move to Folder, Export Source, Back Up, Restore, Delete Source.
- Tab browser `example.com` dan sesi terminal kedua: ditutup. Di terminal hanya dijalankan `echo ATM-TEST && pwd`.
- Tidak ada perintah AI dijalankan, tidak ada koneksi ke server nyata, Docker tidak diubah.

## Canvas
- **[terlihat langsung]** Proyek baru menawarkan dua jenis: **Design** (layar UI/UX dengan AI) dan **Presentation** (slide dengan AI).
- Layar proyek: tiga kolom. Kiri "Design Context" (tombol membuka editor markdown design system dengan placeholder Palette dan Typography, penghitung karakter, Cancel/Save) dan daftar "DESIGN NODES". Tengah kanvas bertitik dengan zoom (− 100% +) dan toggle terang/gelap. Kanan "Design chat" dengan riwayat dan chat baru. Bilah atas: judul, **Share**, **Publish**.
- **Share**: undang lewat email dengan peran (Can view), daftar "People with access" (You · Owner), tautan ke workspace lain. Teks: penerima mendapat email dan notifikasi.
- Hapus proyek: "Deletes the project, its design nodes, revision history, and thumbnails. Its chat conversations are archived, not deleted."
- Publish (ke Cloudflare Pages, A4 §5) tidak diuji.

## In-App Browser
- **[terlihat langsung]** Dibuka dari panel kanan (`⇧⌘O`). Toolbar: back, forward, reload, tombol pilih elemen, kolom alamat dengan ikon pengaturan situs dan buka-keluar, ekstensi, menu ⋮.
- Menu ⋮: Find in page, Print, Save as PDF, Zoom, Device toolbar, Take a screenshot, Import cookies and passwords, Passwords and autofill, Downloads, History, Clear browsing data, Browser settings, **Network logs**.
- Halaman `example.com` dimuat dan tampil di panel, tab berganti nama sesuai judul halaman. Butuh beberapa detik sebelum konten tampil.

## Terminal
- **[terlihat langsung]** Panel bawah (`⌘J`), shell zsh di folder workspace. Toolbar: tambah (dengan dropdown), tombol server (berisi "Manage Remote SSH"), mention terminal di chat (@), split, hapus, tutup. Daftar sesi di kanan; sesi punya × saat disorot.

## Docker
- **[terlihat langsung]** Sidebar: Compose (tombol utama), Containers, Volumes, Images, Networks. Semua halaman menampilkan kartu "Switch to Sokudo engine" selama engine aktif adalah Docker Desktop; tidak ada isi yang bisa dilihat tanpa berpindah engine (tidak dilakukan).

## Database Manager
- **[terlihat langsung]** Tiga kolom: sumber (Add New Source, filter, favorit), konsol, dan chat database dengan pemilih sumber aktif. Konsol: tab "Console N", tombol jalankan, jalankan terpilih, jalankan semua, pemilih transaksi, toggle "Continue after errors", simpan, riwayat, status Disconnected dengan tombol Connect; tab hasil Result, Output, Plan.
- **Add Data Source**: pemilih engine dikelompokkan: Relational (PostgreSQL, CockroachDB, MySQL, MariaDB, SQL Server, Oracle, Firebird), Analytics (ClickHouse), Lightweight (SQLite, rqlite, DuckDB, H2), Docs/Cache/Search (MongoDB, Redis, Elasticsearch, Apache Solr), Graph/Vector (Qdrant), Time series (QuestDB), Message queues (MQTT, Apache Pulsar, dan lainnya). Lencana kemampuan: Operate, Browse, Broker; lencana unduhan untuk Java runtime atau DuckDB worker; baris status verifikasi per engine.
- Form SQLite: tab General, Options, Schemas & Databases, Advanced. Validasi: kolom wajib diberi garis merah dan teks "This field is required.", dan footer "Fix the marked fields to save." dengan tombol Test Connection, Cancel, Apply, OK.

## Arti untuk Astra (bahan Fase B, belum keputusan)

| Temuan Sokudo | **[usulan]** untuk Astra |
| --- | --- |
| Canvas: desain UI dan slide dibuat bersama AI, dengan konteks design system, berbagi lewat email, dan Publish | Di luar lingkup M1. Bila kelak ada, "Design Context" sebanding dengan `DESIGN_SYSTEM.md` per project yang disuntikkan daemon ke agent. Berbagi memakai URL preview bertoken, bukan undangan email |
| Browser dalam aplikasi dengan Network logs, device toolbar, dan screenshot | Tidak ditiru (A4 §7). Fitur kecil yang berguna: log jaringan dan tangkapan preview untuk umpan balik klien (`feedback`) |
| Terminal di folder workspace, banyak sesi, tombol "mention di chat" | Terminal worker sudah ada (xterm.js). Tambahkan "kirim potongan keluaran ke chat/tugas" sebagai aksi |
| Docker hanya berfungsi dengan engine terkelola; selain itu kartu ajakan berpindah | Astra memakai Docker di VPS lewat daemon; tampilkan status dan daftar container sebagai informasi, dan semua aksi lewat approval |
| Database Manager: puluhan engine, konsol SQL, riwayat, chat database, sumber bisa dihapus tanpa menyentuh database asli | P2 atau di luar M1. Pola yang layak ditiru: konfirmasi hapus yang menyebut apa persisnya yang dihapus dan apa yang tidak |
| Validasi form: kolom wajib bergaris merah dengan teks, tombol simpan memberi petunjuk "Fix the marked fields" | Sesuai `DESIGN_RULES.md`; pakai pola pesan sebaris yang sama |

## Belum teramati
Remote SSH Manager (butuh profil), menjalankan query di konsol, riwayat query, Docker dengan engine Sokudo, Publish dari Canvas, desain yang dihasilkan AI, dan isi Network logs.
