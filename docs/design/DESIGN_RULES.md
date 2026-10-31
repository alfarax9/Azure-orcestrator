# Design Rules — Astra

Aturan yang wajib diikuti setiap kali membuat atau mengubah UI aplikasi, baik oleh manusia maupun oleh Claude Code.
Nilai konkret (warna, ukuran, komponen) ada di `DESIGN_SYSTEM.md`.

**Urutan prioritas saat bertentangan:** permintaan eksplisit pemilik produk → file ini → `DESIGN_SYSTEM.md` → default dari skill.

---

## 1. Skill yang dipakai

| Skill | Peran | Install |
| --- | --- | --- |
| [taste-skill](https://github.com/Leonxlnx/taste-skill) (`design-taste-frontend`, MIT) | Arah visual, larangan anti-generik, alur kerja layar aplikasi, dan inspektur | `npx skills add https://github.com/Leonxlnx/taste-skill --skill "design-taste-frontend"` |
| [ui-ux-pro-max](https://github.com/nextlevelbuilder/ui-ux-pro-max-skill) (MIT) | Checklist UX: aksesibilitas, interaksi, form, navigasi, chart | `npm install -g ui-ux-pro-max-cli` lalu `uipro init --ai claude` |

Kebutuhan: Node 22+ dan Chrome/Edge/Chromium untuk inspektur taste-skill; Python 3 untuk pencarian ui-ux-pro-max.

**Pembagian tugas:** taste-skill memutuskan *seperti apa tampilannya*; ui-ux-pro-max memastikan *apakah bisa dipakai semua orang*. Kalau rekomendasi visual ui-ux-pro-max bertentangan dengan taste-skill (misalnya glow, palet slate, aksen hijau neon), taste-skill yang menang.

### Alur kerja untuk setiap layar baru

1. **Tulis pertanyaan layar** dalam satu kalimat dan **aksi utamanya** (lihat §4).
2. **Tulis data contoh** di file JSON: nama worker, branch, durasi, jumlah yang bisa dijumlahkan. Total dihitung dari baris, tidak diketik manual.
3. **Bangun dari token** di `DESIGN_SYSTEM.md`. Tidak boleh ada nilai hex atau ukuran langsung di komponen.
4. **Cek UX** untuk masalah spesifik, satu topik per pencarian:
   `python3 <ui-ux-pro-max>/scripts/search.py "error summary validation" --domain ux`
5. **Jalankan gerbang taste-skill** dalam mode aplikasi:
   `node <taste-skill>/scripts/lint.mjs web/src`
   `node <taste-skill>/scripts/inspect.mjs http://localhost:5173 --app`
6. **Lihat screenshot** di 1440, 768, dan 390px. Perbaiki tiga masalah terbesar, ulangi maksimal tiga putaran.
7. **Laporkan** di ringkasan sesi: layar, pertanyaan layar, hasil inspektur (0 FAIL), dan pengecualian yang diambil beserta alasannya.

Halaman marketing atau landing page **bukan** layar aplikasi. Untuk itu pakai alur halaman taste-skill (`draw.mjs` tanpa `--app`) dengan palet dan font yang sama.

---

## 2. Prinsip inti

1. **Warna untuk data, bukan untuk chrome.** Hampir seluruh layar netral. Warna hanya muncul pada status, elemen data utama, seleksi, dan satu tombol utama. Target: di bawah 3% piksel berwarna.
2. **Setiap layar menjawab satu pertanyaan.** Blok terbesar adalah blok yang menjawabnya. Blok tanpa pertanyaan dihapus.
3. **Produk punya wajah sendiri.** Loop strip (`Prompt → Worktree → Preview → Client → Gate → Live → Watch`) muncul di setiap worker. Nama dan tanda produk selalu terlihat di shell.
4. **Data jujur.** Nol ditulis nol, kosong ditulis `—`, angka kecil tetap kecil, dan setiap angka relatif menyebut pembandingnya ("+2 dari kemarin").
5. **Chrome tenang, status jelas.** Navigasi, border, dan label mundur ke abu-abu agar status worker dan production langsung terbaca.
6. **Keyboard dulu.** Semua aksi bisa dilakukan tanpa mouse, dan aksi yang sering punya shortcut.
7. **Kepercayaan sebelum kecepatan.** Aksi yang menyentuh production atau server selalu menampilkan apa yang akan terjadi, dan sebisa mungkin bisa dibatalkan.

---

## 3. Larangan

| Dilarang | Ganti dengan |
| --- | --- |
| Gradient di latar, tombol, kartu, teks, atau KPI | Warna datar |
| Glow, glassmorphism, backdrop blur | Permukaan solid satu langkah tone dari latar |
| Shadow pada elemen diam | Beda tone atau border 1px; shadow hanya untuk yang melayang |
| Dark theme navy atau slate; aksen violet atau indigo; krem dengan terracotta | Netral gelap murni dan biru `--accent` |
| Ikon di dalam kotak atau lingkaran berwarna | Ikon abu-abu 16px tanpa wadah, atau tanpa ikon |
| Sparkline di KPI; KPI di 36–48px berwarna brand | Angka 20px dengan perubahan dalam kata-kata; tren di chart sendiri |
| Chart pelangi (lebih dari 4 hue, satu hue per bar) | Satu hue per chart, sisanya netral |
| Emoji di UI, nav, label, atau state kosong | Kata-kata |
| Header tabel kapital dengan letter-spacing, zebra, band header | Header 12px abu-abu sentence case |
| Angka rata kiri atau proporsional di kolom | Rata kanan, tabular figures |
| Label kecil kapital di atas judul (eyebrow), pill di atas judul (kecuali label bagian, lihat di bawah) | Informasinya masuk ke judul atau dihapus |
| Monospace untuk angka, label, atau meta | Plus Jakarta Sans dengan tabular figures; mono sistem hanya untuk terminal, kode, SHA, branch, path, port, ID |
| Font di luar tiga opsi user (Plus Jakarta Sans, SF Pro Display, sans-serif) dan monospace sistem | Tidak ada pengecualian tanpa ADR |
| Komponen yang memaksa font tertentu (selain mono) | Selalu `var(--font-sans)`, agar pilihan user berlaku di seluruh tampilan |
| Nomor urut dekoratif (01 / 02 / 03) | Urutan lewat posisi; `<ol>` bila memang berurutan |
| "Welcome back" dengan emoji, banner ilustrasi | Judul kata benda 24px |
| Lorem ipsum, Acme, John Doe, angka bulat palsu | Data contoh yang masuk akal dan bisa dijumlahkan |
| Aksen di setiap tombol, tab, dan link | Satu tombol utama per layar; seleksi berupa garis aksen |
| Warna status pada nilai nol atau perubahan yang tidak berarti | Abu-abu, atau perubahan dalam kata-kata |
| Pagination di bawah list overview pendek | Maksimal 6 baris + "View all" + total |
| Warna sebagai satu-satunya pembawa arti | Warna + teks (+ ikon bila perlu) |

### Diizinkan dengan batas (khusus layar aplikasi)

| Boleh | Batas |
| --- | --- |
| Border 1px pada panel, tabel, kontrol | Border atau shadow, tidak pernah keduanya |
| Garis antarbaris | 1px `--rule` |
| Garis putus-putus | Hanya di grid chart, slot kosong, dan tahap yang dilewati di loop strip |
| Label kecil di atas nilai | 12px, sentence case, `--muted` |
| Label bagian kapital berjarak | 11px (`--fs-label`), `--muted`, hanya sebagai judul grup di Settings dan sidebar (ADR 0009); bukan eyebrow di atas judul layar |
| Pill | Hanya di dalam baris atau di samping nilai, tinggi 20px |
| Monospace | Sesuai daftar di atas |
| Shadow | Menu, popover, tooltip, dialog, command palette, toast |

---

## 4. Aturan per layar

Setiap layar punya satu pertanyaan dan satu aksi utama. Aksi yang paling sering dilakukan harus bisa dijangkau dalam satu klik dari layar pertama.

| Layar | Pertanyaan | Aksi utama |
| --- | --- | --- |
| Board | Apa yang butuh saya sekarang? | New task |
| Detail worker | Sampai mana task ini, dan apa langkah berikutnya? | Sesuai tahap: Share preview, Ship, View failed check, Reply |
| Overview project | Apakah production klien ini sehat, dan versi mana yang live? | Ship (bila ada yang siap) atau View logs (bila gagal) |
| Servers | Server mana yang penuh atau tidak tersambung? | Add server |
| Settings | Apa yang perlu saya atur atau perbaiki? | Toggle dan select berlaku langsung dengan toast; Save hanya untuk form multi-kolom |
| Overlay klien | Bagaimana saya memberi masukan? | Leave feedback |

Tambahan:

- **Temuan utama** layar (misalnya "Gate failed on 2 tests after the last change") ditulis di bawah judul, di samping aksi yang menyelesaikannya.
- **Bukti di samping jawaban.** Status "Rolled back" selalu disertai alasannya ("error rate 4.1% within 3 min of deploy").
- **Semua blok yang disebut di brief layar terlihat di 1440×900** tanpa scroll, minimal header dan baris pertamanya.

---

## 5. Bahasa dan copy

- **Bahasa UI default Inggris**, dengan lokal Indonesia. Kedua bahasa mengikuti aturan yang sama.
- **Kata kerja yang menyebut hasil.** "Ship to production", "Share preview", "Roll back to a3f9c21". Bukan "Submit", "OK", atau "Proceed".
- **Satu aksi, satu nama di seluruh alur.** Tombol "Ship" menghasilkan toast "Shipped", status "Shipping", dan riwayat "Shipped by Maulana".
- **Istilah dari dunia produk**, bukan istilah generik: "Waiting on client", "Gate failed", "Preview live". Bukan "Items", "Records", "Status".
- **Overlay klien memakai bahasa awam.** Klien melihat "Leave feedback" dan "Your comment was sent", tidak pernah "worktree", "branch", atau "agent".
- **Error tidak minta maaf dan tidak samar.** Sebutkan apa yang gagal, kapan data terakhir valid, dan cara memperbaikinya: "Couldn't reach acme-prod-1 over SSH. Check that the server is running, then Retry."
- **State kosong adalah ajakan bertindak.** "No servers yet. Add your first server to start running agents."
- **Sentence case** di semua judul, label, dan tombol. Tanpa tanda seru.

---

## 6. Aksesibilitas (wajib, dari ui-ux-pro-max)

- Kontras teks ≥ 4.5:1; komponen non-teks (border input, ikon fungsional, focus ring) ≥ 3:1. Token baru wajib dicek.
- Focus ring terlihat di semua elemen interaktif: 2px `--accent` dengan offset 2px. Tidak pernah dihapus.
- Urutan tab mengikuti urutan visual. Ada skip link ke konten utama.
- Tombol ikon wajib punya `aria-label` dan tooltip; ikon dekoratif di samping teks diberi `aria-hidden="true"`.
- Hierarki heading berurutan: satu `h1` per layar, tanpa melompati level.
- Perubahan status (worker berubah state, deploy selesai) diumumkan lewat satu region `aria-live="polite"` dengan kalimat lengkap ("Fix checkout total is ready to ship"), tanpa memindahkan focus.
- Target klik minimal 24×24px di desktop dan 44×44px di phone.
- Semua drag (pindah kartu Kanban, resize panel) punya alternatif keyboard dan klik tunggal.
- Sticky header, toast, dan overlay tidak boleh menutupi elemen yang sedang di-focus.
- Login dan input secret mendukung paste dan password manager.
- `prefers-reduced-motion` dihormati di semua animasi.

---

## 7. Interaksi dan keyboard

### 7.1 Shortcut

| Aksi | macOS | Windows / Linux |
| --- | --- | --- |
| Command palette | `⌘K` | `Ctrl+K` |
| Task baru | `⌘N` | `Ctrl+N` |
| Pindah worker berikutnya / sebelumnya | `⌘]` / `⌘[` | `Ctrl+]` / `Ctrl+[` |
| Fokus ke terminal | `⌘J` | `Ctrl+J` |
| Buka preview | `⌘P` | `Ctrl+P` |
| Interrupt agent | `Ctrl+C` di terminal | `Ctrl+C` di terminal |

Shortcut tidak boleh menimpa shortcut sistem atau shortcut di dalam terminal agent. Saat focus berada di terminal, hanya `⌘K`/`Ctrl+K` yang ditangkap app.

### 7.2 Aksi berisiko

| Aksi | Perlindungan |
| --- | --- |
| Ship ke production | Dialog konfirmasi dengan ringkasan perubahan dan domain tujuan |
| Rollback manual | Dialog konfirmasi yang menyebut versi tujuan |
| Stop atau hapus worker | Dialog konfirmasi bila ada perubahan yang belum di-commit |
| Hapus server, project, atau domain | Ketik nama objek untuk konfirmasi |
| Arsipkan worker, tutup komentar klien | Langsung, dengan toast "Undo" selama 8 detik |

Tombol berbahaya selalu terpisah secara visual dan posisi dari aksi normal, dan tidak pernah menjadi tombol default saat menekan `Enter`.

### 7.3 Umpan balik

- Respons visual dalam 100ms untuk setiap klik.
- Tombol async: nonaktif + spinner selama proses, lalu state sukses atau error.
- Operasi lebih dari 1 detik: tampilkan progres nyata (langkah build, persentase upload), bukan spinner tanpa akhir.
- Timeout: pesan jelas + "Retry".

---

## 8. Motion

- Motion hanya untuk **menjawab aksi** (buka, tutup, pindah, konfirmasi) atau **menunjukkan perubahan state** (agent mulai bekerja, deploy selesai).
- Tidak ada animasi masuk untuk layar atau section, dan tidak ada animasi saat scroll.
- Satu-satunya gerakan berulang: pulse pada titik status "Working" dan "Shipping".
- Durasi dari token (`--dur-fast`, `--dur-base`, `--dur-slow`). Exit lebih cepat dari enter.
- Hanya `transform` dan `opacity`. Animasi harus bisa diinterupsi dan tidak pernah memblokir input.
- Kartu yang pindah kolom di Board bergerak ke posisi barunya (shared-element), bukan menghilang lalu muncul.

---

## 9. Aturan kode UI (pelengkap aturan clean code di bagian 10 dokumen riset)

- Komponen hanya memakai token CSS. Nilai hex, `px` lepas untuk warna atau spasi, dan nama font langsung di komponen dianggap error oleh lint.
- Satu set ikon (Lucide). Ikon dari set lain tidak boleh di-import.
- Satu file token (`web/src/styles/tokens.css`) yang menjadi turunan langsung dari `DESIGN_SYSTEM.md` §11.
- Setiap komponen baru mendokumentasikan state-nya: default, hover, focus, pressed, disabled, loading, error, kosong.
- Tidak ada komponen UI dari library pihak ketiga yang membawa gaya visualnya sendiri. Komponen headless (tanpa gaya) boleh, dengan ADR.
- Gerbang `make check` untuk frontend menambahkan: `lint.mjs` taste-skill, `inspect.mjs --app` tanpa FAIL di 1440/768/390, dan anggaran ukuran bundle.

---

## 10. State

- **Kosong:** frame dan header tetap, satu kalimat, aksi pertama.
- **Nol:** ditulis apa adanya.
- **Tidak tersedia:** `—` abu-abu.
- **Memuat:** layout final dengan bar abu-abu datar seukuran nilai, tanpa shimmer dan tanpa layout shift. Refresh data yang sudah tampil tidak mengosongkan layar; cukup perbarui stempel waktu.
- **Error:** di dalam blok yang gagal, bukan banner layar penuh.
- **Offline / server terputus:** status bar berubah, data terakhir tetap terlihat dengan stempel waktu, aksi yang butuh server dinonaktifkan dengan alasan di tooltip.

---

## 11. Data dan chart

- Satu chart menjawab satu pertanyaan. Kalau dua angka sudah menjawab, jangan buat chart.
- Garis 2px lurus tanpa titik, tanpa area fill. 3–5 gridline samar, tanpa garis sumbu Y.
- Satu seri tanpa legend. Lebih dari satu seri: legend di atas chart, dekat datanya.
- Penanda deploy di chart berupa tick 1px dengan SHA pendek; batas SLO berupa garis putus-putus.
- Warna untuk baik atau buruk, bukan naik atau turun. Latensi naik berarti merah; deploy bertambah berarti netral.
- Setiap angka relatif menyebut jendela waktunya ("+12% vs last 24 h").
- Tooltip bisa dijangkau keyboard. Setiap chart punya ringkasan teks untuk screen reader.
- Rasio pemakaian (RAM, disk, limit agent) ditampilkan sebagai bar dengan kedua ujung berlabel ("2.9 of 4.0 GB").

---

## 12. Desktop app (macOS, Windows, Linux)

- **Title bar:** macOS memakai traffic lights bawaan dengan judul di tengah; Windows dan Linux memakai tombol jendela bawaan OS di kanan. Jangan meniru kontrol jendela OS lain.
- **Shortcut:** `⌘` di macOS, `Ctrl` di Windows dan Linux, dari satu definisi yang dipetakan otomatis.
- **Menu:** macOS memakai menu bar sistem (App, File, Edit, View, Window, Help). Windows dan Linux memakai menu di title bar atau di command palette.
- **Font:** user memilih Plus Jakarta Sans (default, di-bundle lokal), SF Pro Display (hanya macOS/iOS, dipanggil sebagai font sistem dan tidak pernah di-bundle karena lisensi Apple), atau sans-serif sistem. Detail di `DESIGN_SYSTEM.md` §3.1.
- **Tema:** mengikuti preferensi OS saat pertama kali dibuka, lalu mengikuti pilihan user.
- **Notifikasi native** hanya untuk hal yang butuh kamu: worker "Needs you", gate gagal, rollback otomatis, komentar klien baru. Bukan untuk setiap perubahan state.
- Uji di ketiga OS setiap rilis (spike S10 di dokumen riset).

---

## 13. Checklist review sebelum merge

Dipakai saat QA manual. Semua harus "ya".

- [ ] Layar punya satu pertanyaan dan satu aksi utama yang terlihat di layar pertama.
- [ ] Nama dan tanda produk terlihat di shell.
- [ ] Hanya token dari `DESIGN_SYSTEM.md` yang dipakai.
- [ ] Tidak ada satu pun item dari tabel larangan §3.
- [ ] Data contoh masuk akal dan totalnya cocok dengan baris.
- [ ] Semua state (kosong, memuat, error, offline) sudah dibuat.
- [ ] Bisa dipakai penuh dengan keyboard; focus ring selalu terlihat.
- [ ] Kontras token baru sudah dicek.
- [ ] `lint.mjs` bersih dan `inspect.mjs --app` 0 FAIL di 1440, 768, dan 390.
- [ ] Layar sudah dicek dengan ketiga opsi font: tidak ada teks terpotong, tombol melebar aneh, atau baris tabel bergeser.
- [ ] Screenshot ketiga lebar sudah dilihat sendiri, bukan hanya hasil inspektur.
- [ ] Aksi berisiko punya konfirmasi atau undo sesuai §7.2.
- [ ] Copy memakai kata kerja yang menyebut hasil, dan nama aksi konsisten di seluruh alur.
- [ ] Teks bantuan sesuai perilaku sebenarnya (tidak usang setelah fitur berubah).
- [ ] Setiap pekerjaan latar (build, spec, Wiki) punya estimasi biaya, progres, dan tombol Batalkan yang terlihat.
- [ ] Teks di atas `--active` hanya `--text` atau `--muted`; status lewat pill (DESIGN_SYSTEM.md §2.2).
- [ ] Shell tetap bisa dipakai di 390px (sidebar menjadi drawer, navigasi menjadi bottom bar).
