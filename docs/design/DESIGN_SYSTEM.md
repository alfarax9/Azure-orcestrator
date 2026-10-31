# Design System — Astra

Sumber kebenaran untuk token, tipografi, komponen, dan pola layar aplikasi.
Aturan perilaku dan larangan ada di `DESIGN_RULES.md`. Kalau keduanya bertentangan, `DESIGN_RULES.md` menang.

---

## 1. Arah visual

| Hal | Keputusan |
| --- | --- |
| Gaya utama (tema gelap, default) | `data-terminal` dari taste-skill, seed `1861568764`, kind `devtools` |
| Gaya tema terang | Palet dari `framed-sheet` (taste-skill), dipetakan ke peran yang sama |
| Alasan | Dipakai dalam sesi panjang (berjam-jam di depan terminal agent), dekat dengan monitoring deploy, dan pengguna adalah developer |
| Wajah produk | **Loop strip**: jalur kerja worker dari prompt sampai production (§8.9). Elemen ini hanya dimiliki produk ini |
| Tipografi | Font UI bisa dipilih user: Plus Jakarta Sans (default), SF Pro Display (macOS), atau sans-serif sistem; monospace sistem untuk terminal |
| Satu warna data | Biru (`--accent`) untuk elemen data utama dan seleksi, di kedua tema |

**Yang ditolak dari ui-ux-pro-max** (dicatat agar tidak dipakai lagi):

- Palet slate/navy dengan aksen hijau. Dark theme navy atau slate dilarang oleh taste-skill, dan "near-black dengan satu aksen hijau terang" adalah ciri desain generik.
- Efek glow. Dilarang di semua layar.
- Pola halaman FAQ, karena tidak relevan untuk layar aplikasi.

**Yang diambil dari ui-ux-pro-max:** kepadatan dashboard (skala spasi rapat), dan seluruh checklist aksesibilitas, interaksi, form, serta chart (lihat `DESIGN_RULES.md`).

---

## 2. Warna

### 2.1 Peran

| Token | Peran | Gelap | Terang |
| --- | --- | --- | --- |
| `--ground` | Latar jendela di belakang panel, bilah atas, dan bilah status | `#060606` | `#edeef0` |
| `--surface` | Panel shell (sidebar, area utama, panel kanan), dialog, terminal | `#181818` | `#f5f6f7` |
| `--raised` | Kartu, composer, hover baris, isi kontrol sekunder | `#202020` | `#ffffff` |
| `--active` | Baris atau item yang dipilih | `#2b2b2b` | `#e3e5e8` |
| `--border` | Tepi panel dan kartu | `#2a2a2a` | `#dcdee2` |
| `--rule` | Garis antarbaris tabel dan list | `#222222` | `#e8eaed` |
| `--control` | Isi tombol sekunder | `#303030` | `#d3d5d9` |
| `--control-strong` | Border input dan checkbox (≥ 3:1 di surface dan raised) | `#6e6e6e` | `#8a8c91` |
| `--text` | Semua teks utama | `#ffffff` | `#16161a` |
| `--muted` | Label, meta, teks sekunder (≥ 4.5:1) | `#9a9a9a` | `#5b5d62` |
| `--accent` | Seleksi (garis 1–2px), link, elemen data utama, focus ring | `#4a8ad6` | `#1d5fbf` |
| `--positive` | Live, sehat, gate lulus | `#44a282` | `#1e7b4f` |
| `--negative` | Gagal, error, rollback paksa | `#e5566f` | `#c4242a` |
| `--warning` | Butuh kamu, mendekati limit | `#d9a441` | `#92400e` |
| `--primary-bg` / `--primary-ink` | Tombol utama | `#f2f2f3` / `#171719` | `#16161a` / `#ffffff` |
| `--scrim` | Latar di belakang dialog | `rgba(0,0,0,.6)` | `rgba(19,20,22,.4)` |

**Pill status** (latar pucat + tinta dengan hue yang sama):

| Token | Gelap (bg / ink) | Terang (bg / ink) |
| --- | --- | --- |
| `--pill-positive` | `#1b2828` / `#5fc39c` | `#dcfae5` / `#04250e` |
| `--pill-negative` | `#2c1a20` / `#e5566f` | `#ffedea` / `#831c1e` |
| `--pill-warning` | `#2a2416` / `#d9a441` | `#fffbeb` / `#92400e` |
| `--pill-neutral` | `#202020` / `#bdbdc0` | `#e3e5e8` / `#353736` |

### 2.2 Kontras yang sudah diverifikasi

Dihitung ulang 2026-10-31 untuk ADR 0009 (rasio WCAG).

| Pasangan | Gelap | Terang |
| --- | --- | --- |
| `--text` di `--surface` | 17.8:1 | 16.7:1 |
| `--text` di `--raised` | 16.3:1 | 18.0:1 |
| `--muted` di `--surface` | 6.3:1 | 6.1:1 |
| `--muted` di `--raised` | 5.8:1 | 6.6:1 |
| `--muted` di `--active` | 5.0:1 | 5.2:1 |
| `--accent` di `--surface` | 5.0:1 | 5.6:1 |
| `--accent` di `--raised` | 4.6:1 | 6.1:1 |
| `--positive` di `--surface` | 5.7:1 | 4.9:1 |
| `--negative` di `--surface` | 5.0:1 | 5.3:1 |
| `--warning` di `--surface` | 7.9:1 | 6.6:1 |
| `--control-strong` di `--surface` | 3.5:1 | 3.1:1 |
| `--control-strong` di `--raised` | 3.2:1 | 3.4:1 |
| Pill negatif (ink di bg) | 4.6:1 | 8.6:1 |

**Batas pada `--active`:** di atas `--active`, `--accent` dan `--negative` sebagai teks hanya 4.0:1 (gelap), dan `--control-strong` 2.8:1 (gelap) dan 2.7:1 (terang). Di atas `--active` hanya `--text` dan `--muted` yang dipakai sebagai teks; status lewat pill, dan border input atau checkbox tidak diletakkan langsung di atasnya.

Setiap token baru wajib dicek: teks ≥ 4.5:1, komponen non-teks (border input, ikon fungsional) ≥ 3:1.

### 2.3 Palet terminal (ANSI)

Output program boleh memakai semua hue ANSI, termasuk magenta. Larangan violet hanya berlaku untuk UI aplikasi, bukan output terminal.

| Warna | Gelap | Terang |
| --- | --- | --- |
| black / bright black | `#3a3a3e` / `#8a8c92` | `#131416` / `#5d5f60` |
| red / bright | `#e5566f` / `#f07a8e` | `#b4232a` |
| green / bright | `#5fc39c` / `#7fd6b2` | `#1e7b4f` |
| yellow / bright | `#d9a441` / `#e8bd66` | `#8a5a00` |
| blue / bright | `#5b9be0` / `#82b4ec` | `#1d5fbf` |
| magenta / bright | `#c58ad8` / `#d6a6e4` | `#8b3fa8` |
| cyan / bright | `#4fb6c4` / `#76cbd6` | `#0f6f7d` |
| white / bright | `#bdbdc0` / `#ffffff` | `#5d5f60` / `#131416` |

Latar terminal: `--surface`. Kursor: `--accent`. Seleksi teks: `--accent` dengan opacity 35%.

---

## 3. Tipografi

### 3.1 Font yang bisa dipilih user

Font UI adalah **preferensi user**. Pilihannya berlaku untuk seluruh tampilan aplikasi (desktop app, web UI, dan PWA di HP), disimpan per akun di daemon, dan ikut di semua perangkat user tersebut.

| Opsi di Settings | Nilai `data-font` | Font yang dipakai | Ketersediaan |
| --- | --- | --- | --- |
| **Plus Jakarta Sans** (default) | `jakarta` | Plus Jakarta Sans, di-bundle di app | Semua OS |
| **SF Pro Display** | `sf` | Font sistem Apple lewat `-apple-system` | Hanya macOS dan iOS |
| **Sans-serif** | `sans` | Sans-serif generik bawaan OS | Semua OS |

**Kenapa Plus Jakarta Sans jadi default:** lisensinya SIL Open Font License sehingga boleh di-bundle; tampil sama di macOS, Windows, dan Linux; punya tabular figures (`tnum`); x-height-nya tinggi (536/1000) sehingga tetap terbaca di 12–13px; dan dibuat oleh Tokotype, studio huruf asal Indonesia.

**Aturan opsi SF Pro Display:**

- Lisensi Apple hanya mengizinkan SF untuk produk di platform Apple, jadi file SF **tidak pernah di-bundle**. App memanggilnya lewat `-apple-system, BlinkMacSystemFont`, yang memang menghasilkan SF di macOS dan iOS. Jangan menulis `"SF Pro Display"` sebagai nama font, karena di WebView nama itu hanya terbaca bila user meng-install SF secara manual.
- Di Windows dan Linux, opsi ini tetap terlihat tapi nonaktif dengan keterangan "Available on macOS and iOS", tidak disembunyikan diam-diam.
- Bila user memilih SF di Mac lalu membuka app di Windows, app memakai Plus Jakarta Sans di perangkat itu tanpa mengubah preferensinya, dan Settings menampilkan keterangan kecil: "SF Pro Display isn't available on Windows. Showing Plus Jakarta Sans."

**Aturan opsi Sans-serif:** memakai `sans-serif` generik, sehingga hasilnya berbeda per OS (Helvetica di macOS, Arial di Windows, DejaVu Sans atau Liberation Sans di Linux). Ini untuk user yang ingin tampilan paling netral.

**Yang tidak ikut berubah:**

- **Terminal, kode, diff, SHA, branch, path, dan ID** selalu memakai monospace sistem (`--font-mono`), apa pun pilihan user.
- **Overlay feedback klien** selalu memakai Plus Jakarta Sans, karena klien tidak memilih font dan tampilannya harus sama di setiap perangkat klien.

**Bundling Plus Jakarta Sans:** variable font woff2 dari [repo resmi Tokotype](https://github.com/tokotype/PlusJakartaSans) disimpan di `web/assets/fonts/` bersama `OFL.txt`, lalu dimuat dengan `@font-face` agar desktop app bisa jalan offline:

```css
@font-face {
  font-family: "Plus Jakarta Sans";
  src: url("/assets/fonts/PlusJakartaSans[wght].woff2") format("woff2-variations");
  font-weight: 200 800;
  font-style: normal;
  font-display: swap;
}
```

**Monospace sistem** (tidak ada font mono yang di-bundle): SF Mono di macOS, Cascadia Mono atau Consolas di Windows, DejaVu Sans Mono di Linux.

| Token | Nilai |
| --- | --- |
| `--font-sans` | Ditentukan oleh `data-font` di `<html>` (lihat §11) |
| `--font-mono` | `ui-monospace, "SF Mono", Menlo, "Cascadia Mono", Consolas, "DejaVu Sans Mono", monospace` |

### 3.1.1 Konsekuensi untuk layout

Ketiga font punya lebar dan x-height berbeda, jadi UI harus tahan terhadap pergantian font:

- Tidak ada lebar tetap untuk elemen berisi teks (tombol, pill, tab, header kolom). Lebar mengikuti isi, dengan `min-width` bila perlu.
- Line-height ditulis dalam px (lihat §3.2) agar tinggi baris dan pitch tabel tidak berubah saat font diganti.
- Teks yang bisa terpotong memakai ellipsis + tooltip, dan diuji dengan font yang paling lebar.
- Tracking default (0) untuk semua font. Judul 24px boleh `-0.01em` hanya untuk Plus Jakarta Sans.
- Stylistic set Plus Jakarta Sans (`ss01`–`ss03`) tidak dipakai.
- Inspektur taste-skill dijalankan untuk ketiga opsi font, bukan hanya default.

### 3.1.2 Ganti font tanpa kedip

- Preferensi disimpan di daemon dan disalin ke cache lokal (config desktop app, atau `localStorage` di web).
- Atribut `data-font` disetel di `<html>` **sebelum render pertama** dari cache lokal, lalu diselaraskan dengan daemon setelah tersambung.
- Mengganti font di Settings langsung berlaku tanpa reload dan tanpa animasi.

### 3.2 Skala

| Token | Ukuran / line-height / weight | Dipakai untuk |
| --- | --- | --- |
| `--fs-title` | 24 / 32 / 600 | Judul layar (`h1`), satu per layar |
| `--fs-section` | 16 / 24 / 600 | Judul panel dan section |
| `--fs-kpi` | 20 / 28 / 500 | Angka KPI |
| `--fs-body` | 14 / 20 / 400 | Teks utama, isi tabel, input |
| `--fs-dense` | 13 / 18 / 400 | Tabel padat, isi terminal |
| `--fs-small` | 12 / 16 / 400 | Label, meta, axis, timestamp |

- Tidak ada teks di bawah 12px.
- Di phone, input minimal 16px agar iOS tidak auto-zoom.
- `font-variant-numeric: tabular-nums` di semua tabel, KPI, timer, dan nilai yang berubah.
- Sentence case di semua label. Tidak ada huruf kapital dengan letter-spacing.
- Weight hanya 400, 500, dan 600.

---

## 4. Spasi, layout, dan breakpoint

### 4.1 Skala spasi (basis 4px)

| Token | Nilai | Dipakai untuk |
| --- | --- | --- |
| `--space-1` | 4px | Ikon ke label |
| `--space-2` | 8px | Label ke field, antar pill |
| `--space-3` | 12px | Padding sel tabel horizontal |
| `--space-4` | 16px | Padding panel padat, antar field |
| `--space-6` | 24px | `--gap` utama: antar panel, horizontal dan vertikal |
| `--space-8` | 32px | Padding layar desktop |
| `--space-12` | 48px | Jeda antar kelompok besar di settings |

Satu nilai gap (`24px`) untuk horizontal dan vertikal di level layar. Pengecualian: **shell** memakai `--shell-inset` (12px, tepi luar) dan `--shell-gap` (8px, antarpanel) (ADR 0009).

### 4.2 Dimensi shell dan kontrol

| Token | Nilai |
| --- | --- |
| `--shell-inset` | 12px (tepi luar jendela ke panel) |
| `--shell-gap` | 8px (antarpanel) |
| `--sidebar-width` | 300px (dapat di-resize 240–360px) |
| `--right-width` | 320px (panel kanan; preview, inspector, tugas) |
| `--pane-min` | 320px |
| `--topbar-height` | 32px (kontrol di atas latar, tanpa panel sendiri) |
| `--tabbar-height` | 36px |
| `--statusbar-height` | 28px (tanpa panel sendiri) |
| `--control-height` | 32px desktop, 44px phone |
| `--row-height` | 42px tabel padat, 48px list |

### 4.3 Breakpoint

| Rentang | Perilaku |
| --- | --- |
| ≤ 640px (phone) | Navigasi menjadi bottom bar berlabel; satu panel per layar; tabel menjadi baris dua tingkat |
| 641–1100px (tablet) | Daftar worker menjadi drawer; navigasi menjadi bottom bar; panel kanan menjadi sheet |
| ≥ 1101px (desktop) | Shell penuh: sidebar, area utama bertab, panel kanan opsional |

Lebar yang wajib diuji: 390, 768, 1440.

---

## 5. Radius, border, elevasi, z-index

| Token | Nilai | Dipakai untuk |
| --- | --- | --- |
| `--radius-s` | 4px | Bar di chart, `kbd`, segmen loop strip |
| `--radius-m` | 8px | Input, tombol persegi, kartu, tabel |
| `--radius-l` | 12px | Panel shell, dialog, composer, popover |
| `--radius-full` | 999px | Pill status, toggle, segmented, tombol kecil di bilah, avatar, titik status |
| `--border-width` | 1px | Semua border |
| `--shadow-float` | `0 12px 28px rgba(0,0,0,.2)` gelap; `0 8px 24px rgba(19,20,22,.12)` terang | Hanya untuk yang melayang: menu, popover, tooltip, dialog, command palette |

Dua radius (`8px` dan `12px`) ditambah pil, dipilih sekali (ADR 0009). Kartu dan panel dipisahkan oleh beda tone atau border 1px, tidak pernah dengan shadow.

| Layer | z-index |
| --- | --- |
| Konten | 0 |
| Header sticky | 10 |
| Dropdown, popover | 20 |
| Scrim | 40 |
| Dialog, sheet | 50 |
| Toast | 60 |
| Tooltip | 70 |

---

## 6. Motion

| Token | Nilai | Dipakai untuk |
| --- | --- | --- |
| `--dur-fast` | 120ms | Hover, press, focus |
| `--dur-base` | 180ms | Buka/tutup panel, ganti tab, expand baris |
| `--dur-slow` | 240ms | Dialog, sheet, command palette masuk |
| `--ease-out` | `cubic-bezier(0.2, 0, 0, 1)` | Elemen yang masuk |
| `--ease-in` | `cubic-bezier(0.3, 0, 1, 1)` | Elemen yang keluar (durasi 70% dari masuk) |
| `--pulse` | 1.6s opacity 1 → 0.4 → 1 | Hanya titik status "agent sedang bekerja" |

Hanya `transform` dan `opacity` yang dianimasikan. Di bawah `prefers-reduced-motion: reduce`, semua durasi jadi 0 dan pulse berhenti (titik tetap terlihat penuh).

---

## 7. Ikon

- Set: **Lucide** (lisensi ISC), outline, stroke 1.5.
- Ukuran: 16px di dalam UI, 18px di sidebar.
- Warna: `--muted`; `--text` saat aktif atau hover.
- Ikon tanpa label wajib punya `aria-label` dan tooltip.
- Tidak ada ikon di dalam kotak atau lingkaran berwarna.

---

## 8. Komponen

### 8.1 Button

| Varian | Tampilan | Kapan dipakai |
| --- | --- | --- |
| Primary | Isi `--primary-bg`, teks `--primary-ink`, 500 | Satu per layar: aksi utama layar itu |
| Secondary | Isi `--control` (gelap) atau `--raised` (terang), teks `--text` | Aksi pendukung |
| Ghost | Tanpa isi, teks `--text`; hover `--raised` | Aksi di baris tabel, toolbar |
| Danger | Isi `--pill-negative` bg, teks negative ink | Rollback, hapus, stop agent; selalu dengan konfirmasi |

- Tinggi 32px (desktop), 44px (phone). Padding horizontal 12px. Radius `--radius-m`.
- State: hover (tone naik satu langkah), pressed (`scale(0.98)`), focus (ring), disabled (opacity 0.45, `aria-disabled`), loading (spinner 14px menggantikan ikon, label tetap, tombol dinonaktifkan).
- Label adalah kata kerja yang menyebut hasilnya: "Ship to production", bukan "Submit".

### 8.2 Input, select, textarea

- Tinggi 32px, isi `--surface`, border 1px `--control-strong`, radius `--radius-m`.
- Label 12/500 `--muted` di atas field, jarak 8px. Tidak pernah placeholder sebagai label.
- Helper text 12px di bawah field, diikat dengan `aria-describedby`.
- Error: border `--negative`, pesan 12px di bawah field berisi penyebab dan cara memperbaiki.
- Field secret (token, API key): tombol tampil/sembunyikan, dan nilai tidak pernah ditampilkan ulang setelah disimpan (hanya 4 karakter terakhir).

### 8.3 Toggle dan checkbox

- Toggle 28×16 untuk on/off yang langsung berlaku. Checkbox untuk pilihan yang disimpan dengan tombol.
- Track off: `--control`; on: `--accent`. Thumb `#ffffff`.

### 8.4 Tabs (gerakan khas gaya ini)

- Setiap seleksi (tab, toggle segmen, rentang waktu, item nav) ditandai **garis aksen 2px selebar labelnya** di tepi terdekat.
- Label tab 14/500 `--muted`; tab aktif `--text`.
- Tidak ada pill berisi warna atau segmented control berwarna.

### 8.5 Pill status

- Tinggi 20px, padding 0 8px, 12/500, radius `--radius-full`.
- Selalu teks, boleh diawali titik 6px. Warna tidak pernah satu-satunya pembawa arti.
- Hanya di dalam baris atau di samping nilai, tidak pernah di atas judul.

### 8.6 Tabel dan list

- Header 12/500 `--muted`, sentence case, tanpa isi dan tanpa band.
- Baris: pitch 42px (padat) atau 48px; garis 1px `--rule`; hover `--raised`. Tanpa zebra.
- Kolom identitas mendapat lebar terbanyak. Angka rata kanan dengan tabular figures, unit di `--muted`.
- Satu tanda status per baris, satu menu kebab di paling kanan.
- Overview: maksimal 6 baris + "View all" + total. List penuh: total di toolbar, "Load more" di akhir.
- Lebih dari 50 baris: virtualisasi.

### 8.7 Panel

- Isi `--surface` di atas `--ground`, border 1px `--border`, radius `--radius-m`, padding 16–24px.
- Judul panel 16/600 di kiri; aksi panel (maksimal dua) di kanan sejajar judul.

### 8.8 Kartu worker (Kanban)

```
┌──────────────────────────────────────────┐
│ Fix checkout total rounding      ● Working│  ← judul task 14/600 · status
│ client-acme / web · fix/checkout-total    │  ← project · branch (mono)
│ ▬▬▬▬▬▬▬▬▬▬▬▬━━━━──────────────────────    │  ← loop strip mini (4px)
│ Claude Code · 12m · +184 −22              │  ← agent · umur · diff
└──────────────────────────────────────────┘
```

- Lebar kolom 280–320px, padding 12px, radius `--radius-m`, isi `--surface`.
- Satu-satunya warna di kartu: titik status dan segmen loop strip yang sedang berjalan.
- Kartu yang butuh tindakan kamu menampilkan aksinya langsung ("Reply", "View failed check").

### 8.9 Loop strip (elemen khas produk)

Jalur tetap tujuh tahap untuk setiap worker:

`Prompt → Worktree → Preview → Client → Gate → Live → Watch`

| Kondisi tahap | Tampilan |
| --- | --- |
| Selesai | Garis 2px `--text` dengan opacity 60% |
| Sedang berjalan | Garis 2px `--accent`, label 12/600 `--text` |
| Belum | Garis 2px `--rule` |
| Gagal | Garis 2px `--negative`, label berisi penyebab singkat ("Gate failed: 2 tests") |
| Dilewati | Garis putus-putus `--rule` (misalnya Client dilewati untuk project internal) |

- Ukuran penuh (detail worker): tinggi 32px dengan label di bawah setiap segmen. Ukuran mini (kartu): 4px tanpa label.
- Markup `<ol>` dengan `aria-current="step"` pada tahap berjalan. Nomor tahap tidak ditampilkan.
- Hover atau focus pada segmen menampilkan tooltip waktu mulai dan durasi.

### 8.10 Terminal

- Isi `--surface`, font `--font-mono` 13/18, padding 12px, palet ANSI §2.3.
- Header terminal 32px: nama agent, status, model, dan aksi (Interrupt, Detach).
- Indikator "skip permissions" selalu terlihat di header dengan pill warning, tidak bisa disembunyikan.
- Lebih dari 16 terminal terbuka: render WebGL (seperti Orchestrator.inc) agar tetap responsif.

### 8.11 Panel preview

- Toolbar 40px: URL preview (mono, bisa disalin), Refresh, Open, Share with client.
- Iframe preview di atas `--surface`. Saat dev server belum siap: skeleton dengan teks "Waiting for the dev server on :3000".

### 8.12 Toast

- Kanan bawah, lebar 360px, isi `--raised`, `--shadow-float`.
- Tutup otomatis 4 detik, kecuali toast dengan aksi (misalnya "Rolled back · Undo") yang bertahan 8 detik.
- `aria-live="polite"`, tidak mencuri focus.

### 8.13 Dialog konfirmasi

- Lebar 440px, isi `--surface`, `--shadow-float`, scrim `--scrim`.
- Judul menyebut aksi dan objeknya: "Ship fix/checkout-total to acme.com?"
- Isi: apa yang berubah, apa yang terjadi kalau gagal (auto-rollback), dan siapa yang terdampak.
- Tombol: aksi di kanan dengan label yang sama dengan judul ("Ship to production"), "Cancel" di kiri. `Esc` menutup.

### 8.14 Command palette

- `⌘K` (macOS) atau `Ctrl+K` (Windows, Linux). Lebar 640px, posisi 20% dari atas.
- Kelompok: Workers, Projects, Servers, Actions. Hasil terakhir dipakai muncul pertama.

### 8.15 State kosong, memuat, error

| State | Tampilan |
| --- | --- |
| Kosong | Frame dan header tetap; satu kalimat + aksi pertama ("No workers yet. Start a task") |
| Nol | Ditulis normal ("0 deploys"); tidak pernah disembunyikan |
| Tidak tersedia | Dash abu-abu `—` |
| Memuat | Layout final dengan bar abu-abu datar seukuran nilai aslinya; tanpa shimmer; tanpa layout shift |
| Error | Di dalam blok yang gagal: penyebab + kapan data terakhir valid + "Retry" ("Couldn't reach acme-prod-1. Last updated 09:14.") |
| Basi | Stempel waktu + tombol refresh |

---

## 9. Pola layar

### 9.1 Shell

```
 [project ▾] [◧]       [ Board | Workers | Overview ]        [kapsul] [◨]   ← bilah atas 32px, tanpa panel
┌────────────┐ ┌──────────────────────────────────┐ ┌──────────────┐
│ Sidebar    │ │ Area utama (tab, bisa di-split)  │ │ Panel kanan  │
│ 300px      │ │  Title row · Loop strip          │ │ 320px        │
│ ● Fix total│ │  Terminal | Chat | Diff | Gate   │ │ Preview /    │
│ ○ Hero copy│ │                                  │ │ inspector    │
└────────────┘ └──────────────────────────────────┘ └──────────────┘
 koneksi · 3 worker berjalan · pemakaian hari ini · server           ← bilah status 28px
```

- Tiga panel `--surface` mengambang di atas `--ground`, tepi luar `--shell-inset` (12px) dan celah `--shell-gap` (8px), radius `--radius-l`.
- Bilah atas: pemilih project, tombol sidebar, navigasi layar, kapsul aksi (command palette, notifikasi, akun), tombol panel kanan. Tombol kecil di bilah berbentuk pil.
- Sidebar: Board, daftar worker per project, server; Settings di bawah. Tooltip di setiap ikon.
- Area utama memakai tab (`--tabbar-height`); satu layar menampilkan satu pertanyaan.
- Panel kanan opsional; panel bawah opsional untuk terminal dan aktivitas.
- Bilah status memuat item yang bisa dipilih: koneksi daemon, worker berjalan, pemakaian hari ini, server. Tiap item membuka halaman terkait.
- Di bawah 1100px daftar worker menjadi drawer (tombol di bilah atas), navigasi menjadi bottom bar berlabel, dan panel kanan menjadi sheet yang dibuka dari bilah atas. Berbeda dari referensinya, shell wajib punya tata letak ringkas.

### 9.2 Board

- Pertanyaan layar: **"Apa yang butuh saya sekarang?"**
- Kolom: Working, Needs you, Review (preview live atau menunggu klien), Ready to ship, Live.
- "Needs you" selalu kolom kedua dan menampilkan jumlahnya di judul layar.
- Aksi utama: "New task".

### 9.3 Detail worker

- Pertanyaan layar: **"Sampai mana task ini, dan apa langkah berikutnya?"**
- Title row: judul task (`h1`), status pill, project · branch, aksi utama sesuai tahap (Share preview, Ship, View failed check).
- Loop strip ukuran penuh di bawah title row.
- Tab: Terminal, Chat, Diff, Gate, Feedback. Panel kanan: preview.

### 9.4 Overview project (per klien)

- Pertanyaan layar: **"Apakah production klien ini sehat, dan versi mana yang live?"**
- Pola dari sheet `data-terminal` untuk devtools: kartu "Live now" (pesan commit 16/600, branch, SHA mono, worker asal, umur; di kanan: domain, live sejak, waktu build), strip kesehatan (3–4 KPI dalam satu kontainer terbagi), chart 24 jam dengan penanda deploy (2/3) dan penjelasan angka terburuk (1/3), tabel deploy maksimal 6 baris.
- Aksi utama: "Ship" bila ada worker siap, atau "View logs" bila deploy terakhir gagal.

### 9.5 Servers

- Pertanyaan layar: **"Server mana yang penuh atau tidak tersambung?"**
- Tabel: nama, peran (agent, production, all-in-one), status koneksi, CPU, RAM, disk sebagai bar rasio dengan kedua ujung berlabel ("2.9 of 4.0 GB"), jumlah worker, jumlah app.

### 9.6 Settings

- Sidebar Settings bergrup (Workspace, Agents, Application) dengan kolom cari; isi di area utama. **Toggle dan select berlaku langsung** dengan toast; tombol Save hanya untuk form multi-kolom (ADR 0009). Kolom tengah ±1152px. Setiap section satu baris: kolom label ±400px (judul 16/500, deskripsi 14px `--muted`) di samping panel ±700px.
- Grup: Account, Agents (status login per agent), Domains, Servers, Notifications, Appearance.
- **Appearance → Font:** radio group tiga opsi. Setiap opsi menampilkan namanya dan satu baris contoh yang dirender dengan font itu sendiri ("Fix checkout total · 2.9 of 4.0 GB · 12m"). Opsi yang tidak tersedia di OS ini nonaktif dengan alasannya. Perubahan langsung berlaku dan diumumkan lewat `aria-live` ("Font changed to SF Pro Display").

### 9.7 Overlay feedback klien (halaman publik)

- Dilihat klien, bukan developer: selalu tema terang kecuali situs klien gelap, teks bahasa sederhana, tanpa istilah teknis.
- Bar bawah 48px: nama project, "Leave feedback", jumlah komentar. Mode komentar: klik elemen → kotak komentar 320px.
- Tidak boleh menutupi konten situs klien di luar bar dan kotak komentar yang sedang aktif.

### 9.8 Phone

- Navigasi menjadi bottom bar: Board, Projects, Servers, Settings, masing-masing dengan label.
- Board menjadi satu kolom per tab status. Detail worker: loop strip, lalu tab; preview dibuka penuh layar.
- Aksi persetujuan (Approve ship, Reply) selalu terjangkau dengan jempol di bagian bawah.

### 9.9 Pola lintas layar (ADR 0009)

- **Daftar entitas** (agent, rule, memori, server): kartu baris `--raised` dengan judul, lencana, cuplikan, tombol hapus, Edit, toggle aktif. Hapus lewat dialog yang menyebut nama objek, apa yang ikut hilang, dan apa yang tidak.
- **Keadaan kosong:** teks yang menyebut tombol yang harus ditekan; tanpa ilustrasi atau ikon di wadah berwarna.
- **Operasi mahal** (build, spec, Wiki): estimasi biaya sebelum mulai, progres per fase, tombol Batalkan yang terlihat.
- **Validasi form:** kolom wajib bergaris `--negative` dengan pesan sebaris; tombol utama redup dengan petunjuk "Perbaiki kolom bertanda".
- **Command palette** (`⌘K`): grup berjudul, deskripsi, keycap; item merusak diredupkan; satu sumber perintah dan shortcut.
- **Label bagian** 11px kapital berjarak (`--fs-label`) hanya sebagai judul grup di Settings dan sidebar.

---

## 10. Kosakata status

Satu nama untuk satu kondisi, dipakai sama di kartu, judul, toast, dan notifikasi.

| Kondisi (fakta) | Label (EN) | Label (ID) | Tanda |
| --- | --- | --- | --- |
| Agent sedang menulis atau menjalankan perintah | Working | Sedang bekerja | Titik `--accent` dengan pulse |
| Agent menunggu input atau approval | Needs you | Butuh kamu | Pill warning |
| Preview hidup | Preview live | Preview aktif | Pill neutral + link |
| Link sudah dibagikan, belum ada respons | Waiting on client | Menunggu klien | Pill neutral |
| Klien memberi komentar baru | Client feedback | Masukan klien | Pill warning + jumlah |
| Gate gagal | Gate failed | Gate gagal | Pill negative + penyebab |
| Gate lulus, menunggu approval | Ready to ship | Siap ship | Pill positive |
| Sedang deploy | Shipping | Sedang ship | Titik `--accent` dengan pulse |
| Live dan sehat | Live | Live | Titik `--positive` |
| Di-rollback otomatis | Rolled back | Di-rollback | Pill warning + alasan |

---

## 11. Token CSS

```css
:root,
:root[data-theme="dark"] {
  color-scheme: dark;
  --ground: #060606;
  --surface: #181818;
  --raised: #202020;
  --active: #2b2b2b;
  --border: #2a2a2a;
  --rule: #222222;
  --control: #303030;
  --control-strong: #6e6e6e;
  --text: #ffffff;
  --muted: #9a9a9a;
  --accent: #4a8ad6;
  --positive: #44a282;
  --negative: #e5566f;
  --warning: #d9a441;
  --primary-bg: #f2f2f3;
  --primary-ink: #171719;
  --scrim: rgba(0, 0, 0, .6);
  --pill-positive-bg: #1b2828; --pill-positive-ink: #5fc39c;
  --pill-negative-bg: #2c1a20; --pill-negative-ink: #e5566f;
  --pill-warning-bg: #2a2416;  --pill-warning-ink: #d9a441;
  --pill-neutral-bg: #202020;  --pill-neutral-ink: #bdbdc0;
  --shadow-float: 0 12px 28px rgba(0, 0, 0, .2);
}

:root[data-theme="light"] {
  color-scheme: light;
  --ground: #edeef0;
  --surface: #f5f6f7;
  --raised: #ffffff;
  --active: #e3e5e8;
  --border: #dcdee2;
  --rule: #e8eaed;
  --control: #d3d5d9;
  --control-strong: #8a8c91;
  --text: #16161a;
  --muted: #5b5d62;
  --accent: #1d5fbf;
  --positive: #1e7b4f;
  --negative: #c4242a;
  --warning: #92400e;
  --primary-bg: #16161a;
  --primary-ink: #ffffff;
  --scrim: rgba(19, 20, 22, .4);
  --pill-positive-bg: #dcfae5; --pill-positive-ink: #04250e;
  --pill-negative-bg: #ffedea; --pill-negative-ink: #831c1e;
  --pill-warning-bg: #fffbeb;  --pill-warning-ink: #92400e;
  --pill-neutral-bg: #e3e5e8;  --pill-neutral-ink: #353736;
  --shadow-float: 0 8px 24px rgba(19, 20, 22, .12);
}

/* Tema dipilih lewat atribut data-theme di <html>.
   Saat start, app membaca preferensi OS (matchMedia "prefers-color-scheme")
   lalu menyetel data-theme, kecuali user sudah memilih di Settings → Appearance.
   Dengan begitu nilai token tidak perlu ditulis dua kali. */

:root {
  --font-jakarta: "Plus Jakarta Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --font-sf: -apple-system, BlinkMacSystemFont, "Plus Jakarta Sans", sans-serif;
  --font-generic: sans-serif;
  --font-sans: var(--font-jakarta);
  --font-mono: ui-monospace, "SF Mono", Menlo, "Cascadia Mono", Consolas, "DejaVu Sans Mono", monospace;

  --fs-title: 24px;   --lh-title: 32px;
  --fs-section: 16px; --lh-section: 24px;
  --fs-kpi: 20px;     --lh-kpi: 28px;
  --fs-body: 14px;    --lh-body: 20px;
  --fs-dense: 13px;   --lh-dense: 18px;
  --fs-small: 12px;   --lh-small: 16px;
  --fs-label: 11px;   --lh-label: 16px;

  --space-1: 4px;  --space-2: 8px;  --space-3: 12px; --space-4: 16px;
  --space-6: 24px; --space-8: 32px; --space-12: 48px;
  --gap: var(--space-6);
  --shell-inset: var(--space-3);
  --shell-gap: var(--space-2);

  --sidebar-width: 300px;
  --right-width: 320px;
  --pane-min: 320px;
  --topbar-height: 32px;
  --tabbar-height: 36px;
  --statusbar-height: 28px;
  --control-height: 32px;
  --row-height: 42px;

  --radius-s: 4px;
  --radius-m: 8px;
  --radius-l: 12px;
  --radius-full: 999px;

  --dur-fast: 120ms;
  --dur-base: 180ms;
  --dur-slow: 240ms;
  --ease-out: cubic-bezier(0.2, 0, 0, 1);
  --ease-in: cubic-bezier(0.3, 0, 1, 1);

  --z-sticky: 10; --z-dropdown: 20; --z-scrim: 40;
  --z-dialog: 50; --z-toast: 60;    --z-tooltip: 70;
}

@media (max-width: 640px) {
  :root { --control-height: 44px; --row-height: 48px; }
}

@media (prefers-reduced-motion: reduce) {
  :root { --dur-fast: 0ms; --dur-base: 0ms; --dur-slow: 0ms; }
}

/* Font pilihan user (Settings → Appearance → Font).
   App hanya menyetel data-font="sf" di macOS/iOS;
   di OS lain preferensi "sf" dirender sebagai "jakarta". */
:root[data-font="jakarta"] { --font-sans: var(--font-jakarta); }
:root[data-font="sf"]      { --font-sans: var(--font-sf); }
:root[data-font="sans"]    { --font-sans: var(--font-generic); }

body { font-family: var(--font-sans); }

/* Overlay klien tidak mengikuti pilihan developer. */
.client-overlay { --font-sans: var(--font-jakarta); }
```

Komponen hanya boleh memakai token di atas. Nilai hex langsung di komponen dianggap error oleh lint (lihat `DESIGN_RULES.md` §9).
