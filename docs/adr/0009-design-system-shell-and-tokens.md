# ADR 0009: Design system: shell panel mengambang dan token baru

| Hal | Isi |
| --- | --- |
| Status | **Disetujui** pemilik produk (2026-10-31) |
| Pemutus | Maulana Alfara Salim |
| Pengusul | Claude Code |
| Milestone | Fase 0 (ADR 0008 §6); mengubah `DESIGN_SYSTEM.md` §2, §4, §5, §9, §11 dan `DESIGN_RULES.md` §3, §4 |
| Terkait | ADR 0001, 0008; `docs/sokudo-observation/90-visual-tokens.md`; `web/src/styles/tokens.css` |

## Konteks

Pemilik produk meminta tampilan Astra mendekati Sokudo, dengan aturan desain Astra disesuaikan. A8 mengukur Sokudo (1512×949, gelap dan terang): tiga kolom panel mengambang di atas latar yang lebih gelap, margin dan celah kecil (±11 dan ±10 px), sidebar ±300 dan panel kanan ±320, radius panel ±12, tombol kecil berbentuk pil, permukaan netral murni, satu aksen biru, bilah status dan bar atas tanpa panel sendiri.

Astra saat ini: rail 56 px + daftar 280 + main + panel kanan; **gap 24**; **satu radius 6 px**; permukaan `--surface` **lebih gelap** daripada `--ground` (`#131217` di atas `#171719`) dengan rona sedikit ungu; status bar 24 px; Settings dengan "Save per section".

ADR ini memilih apa yang disesuaikan, apa yang dipertahankan, dan nilai tokennya. Nilai Sokudo adalah **bahan pembanding**, bukan nilai yang disalin (ADR 0008 §1).

## Keputusan yang diusulkan

### 1. Yang dipertahankan (tidak berubah)

- Loop strip sebagai wajah produk (PRD, `DESIGN_RULES.md` §2.3).
- Tiga opsi font (Plus Jakarta Sans, SF Pro Display, sans-serif); mono sistem hanya untuk terminal, kode, ID.
- Lucide sebagai satu-satunya pustaka ikon; ikon tanpa wadah berwarna.
- Satu aksen biru; warna hanya untuk status dan data (<3% piksel); satu tombol utama per layar.
- Larangan: gradient, glow, backdrop blur, emoji, shadow pada elemen diam.
- Aksesibilitas §6: teks ≥ 4,5:1, komponen non-teks ≥ 3:1, status tidak hanya lewat warna.
- Dua tema (gelap default, terang) dan opsi mengikuti sistem (F11-01).

### 2. Shell: tiga panel mengambang

Menggantikan rail 56 + daftar 280 + main:

```
┌ bilah atas (tanpa panel, 32 px) ─────────────────────────────────────────┐
│ [project ▾] [◧]          [ Board | Workers | Overview ]        [kapsul ⌃] [◨] │
├──────────┬──────────────────────────────────────┬───────────┤
│ sidebar  │ area utama (bertab, bisa di-split)    │ panel     │
│ 300      │                                       │ kanan 320 │
├──────────┴──────────────────────────────────────┴───────────┤
│ bilah status (tanpa panel, 28 px)                                         │
└──────────────────────────────────────────────────────────────────────────┘
```

- Tiga panel permukaan (`--surface`) mengambang di atas `--ground`, dipisah celah 8 px dan margin luar 12 px (kelipatan 4; Sokudo memakai 10/11, dibulatkan ke skala basis 4 px Astra).
- Sidebar: Board, daftar worker per project, server; lebar bawaan 300 px, dapat diubah 240–360 px. Panel kanan: preview, inspector, panel tugas; lebar bawaan 320 px, `--pane-min` 320 px.
- Area utama memakai tab (Terminal, Chat, Diff, Gate, Feedback per worker; F4-03), bar tab 36 px.
- Panel bawah opsional untuk terminal dan aktivitas.
- Bilah status 28 px (naik dari 24): koneksi daemon, worker berjalan, usage hari ini, server; tiap item membuka halaman terkait.
- Breakpoint (§4.3): di bawah 1100 px daftar worker menjadi drawer, navigasi menjadi bottom bar, dan panel kanan menjadi sheet (diterapkan begitu pada implementasi Fase 0; rencana awal memisahkan 768 px). **Astra memiliki tata letak ringkas**, berbeda dari Sokudo yang tidak punya ukuran minimum.
- Perubahan `--gap` dari 24 ke 8 hanya berlaku untuk shell; jarak di dalam layar tetap memakai skala spasi yang ada.

### 3. Hierarki permukaan dan warna

**Urutan baru**: `--ground` (jendela) < `--surface` (panel) < `--raised` (kartu, composer) di tema gelap; kebalikan kecerahan di tema terang, tetapi hierarkinya sama. Rona dibuat netral murni (gelap) dan sedikit dingin (terang).

Nilai usulan (kontras dihitung 2026-10-31; wajib dihitung ulang saat `tokens.css` diubah):

| Token | Gelap | Terang |
| --- | --- | --- |
| `--ground` | `#060606` | `#edeef0` |
| `--surface` | `#181818` | `#f5f6f7` |
| `--raised` | `#202020` | `#ffffff` |
| `--active` | `#2b2b2b` | `#e3e5e8` |
| `--border` | `#2a2a2a` | `#dcdee2` |
| `--rule` | `#222222` | `#e8eaed` |
| `--control` | `#303030` | `#d3d5d9` |
| `--control-strong` | `#6e6e6e` (≈3,5:1) | `#8a8c91` (≈3,1:1) |
| `--text` | `#ffffff` | `#16161a` |
| `--muted` | `#9a9a9a` (≈6,3:1 di surface) | `#5b5d62` (≈6,0:1) |
| `--accent`, `--positive`, `--negative`, `--warning` | tidak berubah (`#4a8ad6`, `#44a282`, `#e5566f`, `#d9a441`) | tidak berubah |

Kontras terhitung untuk nilai di atas (rasio WCAG):

| Pasangan | Gelap | Terang |
| --- | --- | --- |
| `--text` di `--surface` / `--raised` | 17,8 / 16,3 | 16,7 / 18,0 |
| `--muted` di `--surface` / `--raised` / `--active` | 6,3 / 5,8 / 5,0 | 6,1 / 6,6 / 5,2 |
| `--accent` di `--surface` / `--raised` | 5,0 / 4,6 | 5,6 / 6,1 |
| `--negative` di `--surface` / `--raised` | 5,0 / 4,6 | 5,3 / 5,8 |
| `--positive` di `--surface` / `--raised` | 5,7 / 5,2 | 4,9 / 5,3 |
| `--warning` di `--surface` / `--raised` | 7,9 / 7,2 | 6,6 / 7,1 |
| `--control-strong` di `--surface` / `--raised` | 3,5 / 3,2 | 3,1 / 3,4 |

**Batas yang berlaku karena hasil ini:** di atas `--active` (baris atau item terpilih), `--accent` dan `--negative` sebagai teks hanya 4,0 (tema gelap), dan `--control-strong` 2,8 (gelap) dan 2,7 (terang). Jadi pada `--active` hanya `--text` dan `--muted` yang dipakai sebagai teks; status ditampilkan lewat pill (yang punya bg sendiri), dan border input atau checkbox tidak ditempatkan langsung di atas `--active`.

Dasar perbandingan A8: latar jendela `#000`–`#040404`, panel `#181818`, composer `#1d1d1d`–`#1f1f1f`, tepi `#393939`–`#3c3c3c`; terang: jendela `#edeef2`, panel `#f1f2f6`, kartu `#f7f7f9`. Nilai Astra sengaja berbeda tipis (tidak disalin).

**Tidak ditiru**: teks redup Sokudo `#777` di atas `#181818` hanya ≈4,0:1 (gagal AA kecil); `--muted` Astra tetap ≥ 4,5:1 di semua permukaan (diuji juga di `--raised`).

Pill status, `--primary-*`, dan `--scrim` mengikuti perubahan permukaan (dihitung ulang, bukan diubah bentuknya).

### 4. Bentuk dan kontrol

- **Dua radius dan pil**, menggantikan "satu radius 6 px":
  - `--radius-m` **8 px**: input, tombol persegi, kartu, tabel;
  - `--radius-l` **12 px**: panel shell, dialog, composer, popover;
  - `--radius-full` **999 px**: pill status, toggle, segmented, tombol kecil di bar, avatar.
- `--radius-s` 4 px tetap untuk `kbd`, bar chart, dan segmen loop strip.
- Border 1 px tetap; kartu dan panel tidak memakai shadow (aturan lama tetap).
- Tinggi kontrol 32 px (desktop), 44 px (phone), bar tab 36 px, baris tabel 42/48 px: tidak berubah.
- Composer pesan memakai `--raised`, border `--control`, radius `--radius-l`, tinggi minimum ±144 px pada keadaan kosong.
- Dialog: lebar 400 (konfirmasi), 460 (form), 640–960 (form lebar); latar diredupkan `--scrim`.

### 5. Pola layar yang disesuaikan

- **Settings**: sidebar bergrup (Workspace, Agents, Application) dengan kolom cari; isi berupa kartu baris (judul, deskripsi, kontrol di kanan). **Toggle dan select berlaku langsung** dengan umpan balik toast; tombol Save hanya untuk form multi-kolom (misalnya skrip project). Ini mengubah `DESIGN_RULES.md` §4 ("Save per section").
- **Daftar entitas** (agent, rule, memori, server): kartu baris dengan judul, lencana, cuplikan, tombol hapus, Edit, toggle aktif; hapus lewat dialog yang menyebut nama objek, apa yang ikut hilang, dan apa yang tidak.
- **Keadaan kosong**: teks bertulis yang menyebut tombol yang harus ditekan, tanpa ilustrasi atau ikon di wadah berwarna (larangan lama tetap).
- **Operasi mahal** (build, spec, Wiki): estimasi biaya sebelum mulai, progres per fase, dan tombol Batalkan yang terlihat.
- **Validasi form**: kolom wajib bergaris `--negative` dengan pesan sebaris; tombol utama redup dengan petunjuk "Perbaiki kolom bertanda".
- **Command palette** (`⌘K`): grup berjudul, deskripsi, keycap; item merusak diredupkan; satu sumber perintah dan shortcut.
- **Label bagian** 11 px kapital berjarak diizinkan **hanya** sebagai judul grup di Settings dan sidebar. Larangan "eyebrow di atas judul layar" tetap.

### 6. Tipografi

Tangga ketat tetap dengan satu penyesuaian: tambah **11 px** untuk label bagian dan bilah status (sudah ada skala 12; 11 dipakai hanya untuk kapital berjarak). Tidak ada perubahan keluarga font. Skala yang diukur di Sokudo (11/12/14/21/30) dipakai sebagai acuan, bukan sebagai spesifikasi.

### 7. Motion

Aturan §8 tetap (≤ 200 ms, `prefers-reduced-motion` dihormati). Indikator "agent bekerja" satu bentuk (busur berputar atau titik berdenyut), tidak bisa dipilih pengguna.

## Perubahan dokumen dan kode (dikerjakan setelah ADR disetujui)

1. `DESIGN_SYSTEM.md`: §2.1 (tabel warna), §4.2 (dimensi shell), §5 (radius), §9.1 (shell), §11 (token CSS); satu tabel baru §9 untuk Settings dan daftar entitas.
2. `DESIGN_RULES.md`: §3 (satu pengecualian label bagian), §4 (Settings tanpa Save per section; tabel layar), §13 (butir pemeriksaan "teks bantuan sesuai perilaku" dan "ada tombol berhenti untuk pekerjaan latar").
3. `web/src/styles/tokens.css` diperbarui dari §11; `web/src/shell/` dibangun ulang; komponen hanya memakai token.
4. Pemeriksaan: `lint.mjs` dan `inspect.mjs --app` 0 FAIL, uji di 1440/768/390 dengan tiga font, dan kontras tiap token baru dihitung dan dicatat di `DESIGN_SYSTEM.md` §2.2.

## Konsekuensi

**Positif:** tampilan lebih dekat dengan Sokudo tanpa menyalin aset; hierarki permukaan lebih jelas; tata letak ringkas mengatasi kelemahan Sokudo di lebar kecil; aturan kontras tetap lebih ketat.

**Negatif dan risiko:**
- Semua layar yang sudah lulus §13 (Board, detail worker, pairing, onboarding) perlu diperiksa ulang.
- Radius dan permukaan berubah, jadi uji visual di tiga font dan dua tema diulang.
- Tiga panel mengambang butuh ruang: di bawah 1100 px perilaku drawer harus diuji sungguhan.
- Memperlonggar "Save per section" menuntut umpan balik yang jelas agar pengguna tahu perubahan tersimpan.

## Alternatif yang ditolak

- **Mempertahankan rail dan satu radius.** Tidak memenuhi permintaan tampilan mirip.
- **Menyalin nilai Sokudo apa adanya.** Melanggar batas ADR 0008 §1 dan membawa teks redup yang gagal AA.
- **Menambah tema ketiga atau pemilih ikon/animasi.** Bertambah kerja tanpa nilai bagi alur agent.

## Pertanyaan terbuka

1. Persetujuan hierarki permukaan baru (`--ground` lebih gelap dari `--surface` di tema gelap).
2. Apakah label bagian kapital diterima sebagai satu pengecualian terbatas.
3. Lebar bawaan sidebar 300 dan panel kanan 320 (Astra sebelumnya 280 dan 360).
4. Nama project di bilah atas (pemilih) atau tetap di sidebar.
