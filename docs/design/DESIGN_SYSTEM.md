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
| `--ground` | Latar jendela, rail, dan frame | `#171719` | `#f9fbfc` |
| `--surface` | Panel, kartu, terminal, tabel | `#131217` | `#ffffff` |
| `--raised` | Hover baris, isi kontrol sekunder | `#212023` | `#f4f6f8` |
| `--active` | Baris atau item yang dipilih | `#2a292d` | `#ebedee` |
| `--border` | Tepi panel dan kartu | `#222224` | `#e8eaec` |
| `--rule` | Garis antarbaris tabel dan list | `#1b1a1f` | `#f0f1f2` |
| `--control` | Isi tombol sekunder | `#3a3a3e` | `#cdced1` |
| `--control-strong` | Border input dan checkbox (≥ 3:1) | `#6b6d73` | `#8f9196` |
| `--text` | Semua teks utama | `#ffffff` | `#131416` |
| `--muted` | Label, meta, teks sekunder (≥ 4.5:1) | `#8a8c92` | `#5d5f60` |
| `--accent` | Seleksi (garis 1–2px), link, elemen data utama, focus ring | `#4a8ad6` | `#1d5fbf` |
| `--positive` | Live, sehat, gate lulus | `#44a282` | `#1e7b4f` |
| `--negative` | Gagal, error, rollback paksa | `#e5566f` | `#c4242a` |
| `--warning` | Butuh kamu, mendekati limit | `#d9a441` | `#92400e` |
| `--primary-bg` / `--primary-ink` | Tombol utama | `#f2f2f3` / `#171719` | `#131416` / `#ffffff` |
| `--scrim` | Latar di belakang dialog | `rgba(0,0,0,.6)` | `rgba(19,20,22,.4)` |

**Pill status** (latar pucat + tinta dengan hue yang sama):

| Token | Gelap (bg / ink) | Terang (bg / ink) |
| --- | --- | --- |
| `--pill-positive` | `#1b2828` / `#5fc39c` | `#dcfae5` / `#04250e` |
| `--pill-negative` | `#2c1a20` / `#e5566f` | `#ffedea` / `#831c1e` |
| `--pill-warning` | `#2a2416` / `#d9a441` | `#fffbeb` / `#92400e` |
| `--pill-neutral` | `#212023` / `#bdbdc0` | `#ebedee` / `#353736` |

### 2.2 Kontras yang sudah diverifikasi

| Pasangan | Gelap | Terang |
| --- | --- | --- |
| `--text` di `--surface` | 18.6:1 | 18.4:1 |
| `--muted` di `--surface` | 5.6:1 | 6.4:1 |
| `--muted` di `--raised` | 4.8:1 | — |
| `--accent` di `--ground` | 5.0:1 | 5.9:1 |
| `--positive` di `--surface` | 6.0:1 | 5.3:1 |
| `--negative` di `--surface` | 5.2:1 | 5.8:1 |
| `--warning` di `--surface` | 8.3:1 | 7.1:1 |
| `--control-strong` di `--surface` | 3.6:1 | 3.2:1 |
| Pill negatif (ink di bg) | 4.6:1 | 8.6:1 |

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

Satu nilai gap (`24px`) untuk horizontal dan vertikal di level layar.

### 4.2 Dimensi shell dan kontrol

| Token | Nilai |
| --- | --- |
| `--rail-width` | 56px |
| `--list-width` | 280px (kolom daftar worker, bisa di-resize 240–360px) |
| `--pane-min` | 360px (panel preview atau inspector) |
| `--titlebar-height` | 40px |
| `--control-height` | 32px desktop, 44px phone |
| `--row-height` | 42px tabel padat, 48px list |

### 4.3 Breakpoint

| Rentang | Perilaku |
| --- | --- |
| ≤ 640px (phone) | Rail menjadi bottom bar berlabel; satu panel per layar; tabel menjadi baris dua tingkat |
| 641–1100px (tablet) | Rail tetap; kolom daftar worker dilipat jadi drawer; panel preview pindah ke tab |
| ≥ 1101px (desktop) | Shell penuh: rail, daftar worker, main, panel kanan opsional |

Lebar yang wajib diuji: 390, 768, 1440.

---

## 5. Radius, border, elevasi, z-index

| Token | Nilai | Dipakai untuk |
| --- | --- | --- |
| `--radius-s` | 4px | Bar di chart, `kbd`, segmen loop strip |
| `--radius-m` | 6px | Semua tombol, input, panel, kartu, dialog |
| `--radius-full` | 999px | Pill status, avatar, titik status |
| `--border-width` | 1px | Semua border |
| `--shadow-float` | `0 12px 28px rgba(0,0,0,.2)` gelap; `0 8px 24px rgba(19,20,22,.12)` terang | Hanya untuk yang melayang: menu, popover, tooltip, dialog, command palette |

Satu radius (`6px`) untuk semua permukaan, dipilih sekali. Kartu dan panel dipisahkan oleh beda tone atau border 1px, tidak pernah dengan shadow.

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
- Ukuran: 16px di dalam UI, 18px di rail.
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
┌────┬──────────────┬──────────────────────────────┬────────────────┐
│Rail│ Workers      │ Main                         │ Preview /      │
│56px│ 280px        │                              │ inspector      │
│    │ (per project)│                              │ ≥ 360px        │
│ ◆  │ ● Fix total  │  Title row                   │                │
│ ▤  │ ● Add login  │  Loop strip                  │                │
│ ⧉  │ ○ Hero copy  │  Terminal | Chat | Diff      │                │
│ ⚙  │              │                              │                │
└────┴──────────────┴──────────────────────────────┴────────────────┘
```

- Rail: tanda produk di atas, lalu Board, Projects, Servers; Settings di bawah. Tooltip di setiap ikon.
- Title bar desktop app 40px: nama project aktif, server, dan pencarian (`⌘K`).
- Status bar 24px di bawah: server tersambung, jumlah worker bekerja, pemakaian limit agent.

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

- Kolom tengah ±1152px. Setiap section satu baris: kolom label ±400px (judul 16/500, deskripsi 14px `--muted`) di samping panel ±700px.
- Grup: Account, Agents (status login per agent), Domains, Servers, Notifications, Appearance.
- **Appearance → Font:** radio group tiga opsi. Setiap opsi menampilkan namanya dan satu baris contoh yang dirender dengan font itu sendiri ("Fix checkout total · 2.9 of 4.0 GB · 12m"). Opsi yang tidak tersedia di OS ini nonaktif dengan alasannya. Perubahan langsung berlaku dan diumumkan lewat `aria-live` ("Font changed to SF Pro Display").

### 9.7 Overlay feedback klien (halaman publik)

- Dilihat klien, bukan developer: selalu tema terang kecuali situs klien gelap, teks bahasa sederhana, tanpa istilah teknis.
- Bar bawah 48px: nama project, "Leave feedback", jumlah komentar. Mode komentar: klik elemen → kotak komentar 320px.
- Tidak boleh menutupi konten situs klien di luar bar dan kotak komentar yang sedang aktif.

### 9.8 Phone

- Rail menjadi bottom bar: Board, Projects, Servers, Settings, masing-masing dengan label.
- Board menjadi satu kolom per tab status. Detail worker: loop strip, lalu tab; preview dibuka penuh layar.
- Aksi persetujuan (Approve ship, Reply) selalu terjangkau dengan jempol di bagian bawah.

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
  --ground: #171719;
  --surface: #131217;
  --raised: #212023;
  --active: #2a292d;
  --border: #222224;
  --rule: #1b1a1f;
  --control: #3a3a3e;
  --control-strong: #6b6d73;
  --text: #ffffff;
  --muted: #8a8c92;
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
  --pill-neutral-bg: #212023;  --pill-neutral-ink: #bdbdc0;
  --shadow-float: 0 12px 28px rgba(0, 0, 0, .2);
}

:root[data-theme="light"] {
  color-scheme: light;
  --ground: #f9fbfc;
  --surface: #ffffff;
  --raised: #f4f6f8;
  --active: #ebedee;
  --border: #e8eaec;
  --rule: #f0f1f2;
  --control: #cdced1;
  --control-strong: #8f9196;
  --text: #131416;
  --muted: #5d5f60;
  --accent: #1d5fbf;
  --positive: #1e7b4f;
  --negative: #c4242a;
  --warning: #92400e;
  --primary-bg: #131416;
  --primary-ink: #ffffff;
  --scrim: rgba(19, 20, 22, .4);
  --pill-positive-bg: #dcfae5; --pill-positive-ink: #04250e;
  --pill-negative-bg: #ffedea; --pill-negative-ink: #831c1e;
  --pill-warning-bg: #fffbeb;  --pill-warning-ink: #92400e;
  --pill-neutral-bg: #ebedee;  --pill-neutral-ink: #353736;
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

  --space-1: 4px;  --space-2: 8px;  --space-3: 12px; --space-4: 16px;
  --space-6: 24px; --space-8: 32px; --space-12: 48px;
  --gap: var(--space-6);

  --rail-width: 56px;
  --list-width: 280px;
  --pane-min: 360px;
  --titlebar-height: 40px;
  --control-height: 32px;
  --row-height: 42px;

  --radius-s: 4px;
  --radius-m: 6px;
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
