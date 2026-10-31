# A8. Nilai visual terukur (ukuran, warna, tipografi)

Pengamatan Sokudo 0.19.6 (macOS), 2026-10-31. Semua angka adalah **nilai pengamatan dari piksel tangkapan layar**, bukan aset atau kode yang disalin, dan bukan hasil inspeksi frontend. Ini bahan untuk Fase B (ADR 0009, revisi `DESIGN_SYSTEM.md` §11), bukan keputusan.

Tanda: **[terukur]** dibaca dari piksel; **[estimasi]** disimpulkan dari bentuk atau proporsi; **[inferensi]** dugaan.

## 1. Metode dan keterbatasan

- Sumber: tangkapan 1:1 jendela 1512×949 pt (`a1-shell-fullsize-1512.jpg`, tema gelap) dan tangkapan berskala (≈0,915 atau lebih kecil) untuk tema terang dan halaman lain. Warna diambil dengan program Go kecil yang membaca piksel (garis pindai, wilayah, dan warna paling sering); program itu tidak disimpan di repo.
- Galat: JPEG memberi derau ±2 level per kanal, tepi teks tampak lebih gelap atau terang dari warna sebenarnya, dan posisi tepi ±1–2 px. Warna teks tipis (antialias) hanya bisa dibaca sebagai batas bawah; yang dipakai adalah piksel paling terang di area huruf tebal.
- **Tidak terukur** pada pengamatan ini: keadaan hover, fokus, dan aktif; durasi animasi; ukuran minimum jendela (tidak ada, lihat A1); nama font (hanya kesan visual); ketebalan garis ikon; nilai tema terang selain permukaan.

## 2. Tata letak (jendela 1512×949, tema gelap) [terukur]

| Elemen | Nilai |
| --- | --- |
| Margin luar kiri dan kanan | 11 px |
| Celah antarpanel | ±10 px (celah kanan ±11) |
| Bilah atas (kontrol) | y 8–39, tinggi ±32 px, di atas latar jendela, tanpa panel sendiri |
| Panel mulai dari | y = 51 |
| Sidebar kiri | x 11–312, lebar **302 px**, tinggi sampai y ≈ 911 |
| Area editor | x 323–1168, lebar 846 px (fleksibel), tinggi sampai y ≈ 895 |
| Sidebar kanan | x 1180–1500, lebar **321 px** |
| Bar tab di dalam editor | y 52–87, tinggi **36 px** |
| Bilah status | y 912–948, tinggi ±37 px, di atas latar jendela, tanpa panel |
| Tombol "New Chat" sidebar | x 28–296 × y 64–96: lebar 268, tinggi **33 px** |
| Kontrol segmented atas | x 632–879, tinggi ±32 px |
| Composer (kosong) | x 405–1086 × y 435–581: **682×146 px**, terpusat di area editor, 2/3 lebar editor |
| Tombol kirim | lingkaran ±32 px, di sudut kanan bawah composer |
| Popover bilah status | x 1062–1358 (±297 px), tinggi ±62 px, menempel di atas bilah status |

Irama jarak: panel-ke-panel 10 px, tepi luar 11 px, padding isi sidebar ±17 px (tombol mulai 17 px dari tepi panel). **[estimasi]** sistem jarak berbasis kelipatan 4 px dengan satu pengecualian 10–11 px untuk celah panel.

## 3. Warna [terukur]

### 3.1 Tema gelap (netral, tanpa rona)

| Peran | Nilai |
| --- | --- |
| Latar jendela (di luar panel) | `#000000`–`#040404` |
| Permukaan panel (sidebar, editor, popover) | `#181818` (variasi `#161616`–`#191919`) |
| Bar tab (strip di atas editor) | `#101010` |
| Tab aktif (menyatu dengan panel) | `#181818` |
| Sidebar kanan | `#171717` |
| Isi composer | `#1d1d1d`–`#1f1f1f` |
| Garis tepi composer | `#393939`–`#3c3c3c`, 1 px |
| Kapsul ikon dan tombol kecil di bilah atas | `#0d0d0d` dengan garis tepi tipis |
| Tombol sekunder ("New Chat" sidebar) dan segmen aktif | `#303030` |
| Tab panel kanan aktif | `#3a3a3a` |
| Tombol kirim (nonaktif) | `#404040` |
| Teks utama | `#ffffff` (judul), `#f3f3f3`–`#fbfbfb` (label aktif) |
| Teks sekunder | `#c0c0c0`–`#c6c6c6` |
| Teks redup (label bagian, placeholder) | `#777777`–`#878787` |

Kontras yang dihitung dari nilai di atas:
- placeholder `#878787` di atas `#1f1f1f` ≈ **4,6:1** (lolos AA untuk teks kecil, tipis);
- teks redup `#777777` di atas `#181818` ≈ **4,0:1** (**tidak lolos AA 4,5:1** untuk teks kecil; Astra tidak boleh meniru nilai ini, `DESIGN_RULES.md` §6).

### 3.2 Aksen dan status (dibaca dari area penuh, ±3 level)

| Peran | Nilai | Dipakai untuk |
| --- | --- | --- |
| Primer (biru) | `#4190ed` / `#458eed` | satu tombol utama per keadaan (Add New Source, Switch to Sokudo, Save) |
| Bahaya (merah) | `#b5313f` / `#b43141` | tombol hapus dan reset (isi merah penuh, teks putih) |
| Sukses (hijau) | `#58c060`–`#60c050` | toggle hidup, titik status, lencana sukses |
| Peringatan (amber) | `#f0b848`–`#f4c311` | titik kolom Under Review, lencana "ENABLED", jenis memori Gotcha |
| Galat (merah muda) | `#e86858`–`#f06058` | ikon dan teks galat sebaris |
| Info (biru lembut) | `#5068a0`–`#5081dd` | lencana kategori (Process, Quality, Decision), titik In Progress |
| Aksen perisai izin | `#bb8a61` (oranye hangat) | ikon "Full access" |
| Memori: Structure, Preference | ungu `#6a5fb5`-an dan hijau `#58c848` **[estimasi]** | lencana jenis memori |

**[terlihat langsung]** Hanya satu rona selain netral yang mendominasi tema gelap: biru primer. Semua status memakai warna jenuh kecil pada latar netral, dan status selalu disertai label atau ikon (bukan warna saja).

### 3.3 Tema terang (dari tangkapan berskala)

| Peran | Nilai |
| --- | --- |
| Latar jendela | `#edeef2` (sedikit kebiruan: B lebih tinggi dari R dan G) |
| Permukaan panel | `#f1f2f6` |
| Kartu halaman Settings | `#f7f7f9` |
| Permukaan sorotan di area editor | `#fbfbfd` |
| Teks utama | `#2c2b2f` **[estimasi, dari piksel tergelap pada area teks]** |
| Teks redup | `#6c6d72`–`#808186` **[estimasi]** |

**[inferensi]** Mode terang adalah pembalikan hierarki, bukan sekadar inversi: latar jendela sedikit lebih gelap daripada panel (panel "terangkat"), sedangkan di mode gelap latar jendela justru lebih gelap dan panel lebih terang. Jadi hierarki sama: jendela < panel < kartu.

## 4. Bentuk

| Elemen | Nilai |
| --- | --- |
| Radius panel | **≈12 px** **[terukur dari lengkung sudut]** |
| Radius composer | **≈14–16 px** **[terukur]** |
| Radius tombol segmented dan tombol | pil penuh (setengah tinggi) pada "New Chat", segmented, tombol kirim; kartu dan input ±8–10 px **[estimasi]** |
| Garis tepi | 1 px, abu `#303030`–`#3c3c3c`; kartu tanpa bayangan |
| Bayangan | tidak terlihat pada panel; popover dan dialog hanya mengandalkan latar yang diredupkan |
| Dialog modal | latar belakang diredupkan, dialog `#1c1c1c` **[estimasi]**, lebar 435–457 px (konfirmasi ±400, form lebar ±660–1000) |

## 5. Tipografi [estimasi]

- **Keluarga**: sans-serif grotesk bergaya UI modern (huruf `a` dan `g` satu tingkat, kerning rapat), mirip Inter. Nama font tidak dapat dipastikan dari tangkapan; kode menggunakan font monospace berbeda (bergaya JetBrains Mono atau SF Mono). Tidak ada ketentuan yang diambil dari sini: Astra tetap memakai tiga opsi font `DESIGN_SYSTEM.md`.
- **Ukuran** (dari tinggi huruf):
  - judul layar kosong ("What you want to build today?"): ±30 px, semibold, putih;
  - judul halaman Settings: ±21 px semibold;
  - judul baris pengaturan: ±14 px medium; deskripsi: ±12 px, abu;
  - label bagian: ±11 px kapital, huruf berjarak, abu redup;
  - label tombol, tab, dan sidebar: ±14 px;
  - bilah status dan popover: ±11–12 px;
  - kode dan nilai teknis: ±12–13 px monospace.
- Bobot: normal untuk isi, medium untuk judul baris, semibold untuk judul layar. Tidak ada bobot di bawah 400.
- Rasio ukuran antartingkat sempit (11, 12, 14, 21, 30): tangga ketat dengan sedikit lompatan.

## 6. Ikon dan elemen kecil [terlihat langsung]

- Ikon garis (outline), ±14–16 px di bilah atas dan status, ±18 px di sidebar dan kartu, bergaya garis tipis seragam dengan sudut membulat. Gayanya tidak dapat diidentifikasi sebagai pustaka tertentu; Astra tetap memakai Lucide.
- Titik status ±8 px dengan warna status; spinner berupa busur yang berputar.
- Toggle: ±38×22 px, bulatan putih, hijau saat hidup, abu `#3a3a3a`-an saat mati.
- Segmented control: kontainer berbingkai tipis, segmen aktif terisi `#303030` dengan teks putih.
- Lencana: pil kecil ±20 px tinggi, teks 11–12 px, isi warna status berintensitas rendah dengan teks berwarna terang.
- Avatar profil: lingkaran ±20 px.

## 7. Arti untuk Astra (bahan ADR 0009, belum keputusan)

| Pengamatan | **[usulan]** |
| --- | --- |
| Tiga lapis permukaan (jendela < panel < kartu) dengan selisih kecerahan kecil (`#000` → `#181818` → `#1f1f1f`) dan garis tepi 1 px, tanpa bayangan | Pola yang cocok dengan token Astra: tiga token permukaan + satu token garis. Nilai Astra diturunkan dari `DESIGN_SYSTEM.md` §11, bukan disalin dari angka di atas |
| Satu aksen (biru) untuk tindakan utama, merah penuh untuk aksi merusak, hijau/amber/merah untuk status | Sejalan dengan "satu tombol utama per layar" (`CLAUDE.md`). Periksa kontras tiap status terhadap AA, termasuk teks redup |
| Teks redup `#777` pada `#181818` tidak lolos AA | Jangan ditiru. Token teks sekunder Astra harus ≥ 4,5:1 di kedua tema |
| Tata letak tiga kolom: sidebar 302, panel fleksibel, panel kanan 321, tab 36, kontrol 32, composer 682 lebar | Rujukan awal untuk web UI: sidebar ±300, panel kanan ±320, tinggi kontrol 32–36, tinggi bar 36. Perlu diuji di 1440/768/390 (`DESIGN_RULES.md`) |
| Radius 12 untuk panel, 14–16 untuk komposer, pil untuk tombol kecil | Dua radius (12 dan pil) cukup untuk token Astra |
| Hierarki tipografi sempit (11/12/14/21/30) | Pakai tangga sempit serupa dengan satu ukuran isi (14) dan satu label kapital (11) |
| Mode terang dengan hierarki permukaan yang tetap dan sedikit rona dingin | Rona dingin kecil (B > R,G ±5) pada permukaan terang memberi kesan lebih lembut; bisa dipakai sebagai satu opsi di tema terang |
| Tidak ada ukuran minimum jendela, tata letak terpotong pada lebar kecil (A1) | Web UI Astra wajib punya tata letak ringkas di bawah ±768 px |

Batas: tidak ada warna, bentuk, atau ikon yang disalin sebagai aset; angka di atas adalah pengamatan untuk perbandingan.

## Butir terbuka

1. Keadaan hover, fokus, aktif, dan nonaktif (warna dan garis fokus).
2. Durasi dan kurva animasi (transisi panel, spinner, pemilih ikon "Thinking…").
3. Nama dan ukuran pasti font (butuh inspeksi yang tidak dilakukan, karena pembedahan antarmuka berada di luar batas akses).
4. Nilai tema terang yang lengkap (teks, garis tepi, aksen), karena hanya permukaan yang diukur.
5. Ukuran pada pengaturan zoom selain 100% (zoom mengubah semua ukuran secara proporsional, A4 §1.3).
