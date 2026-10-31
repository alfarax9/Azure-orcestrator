# A4. Settings, bagian Application

Pengamatan Sokudo 0.19.6 (macOS), 2026-10-31, lewat computer use, hanya di jendela Sokudo. Halaman: Appearance, Keyboard, Browser, Databases, Sites. Tema gelap (kecuali pengujian mode terang di §1.5).

Tanda: **[terlihat langsung]** tampak di layar; **[inferensi]** dugaan; **[usulan]** bahan untuk Astra, belum keputusan.

Bukti: nama berkas `a4-*.jpg` di `docs/sokudo-observation/screens/` (tidak di-commit). Ukuran dalam piksel tangkapan 1383 lebar (skala ≈0,915 dari pt), kasar (±3). Ukuran presisi dikerjakan di A8.

Yang diubah selama pengamatan (dengan izin pemilik produk) dan sudah dikembalikan:
- Color Mode diganti ke Light lalu kembali ke Dark;
- zoom aplikasi diperkecil lewat View → Zoom Out dan dikembalikan ke 100% lewat Reset Zoom;
- mode penangkapan shortcut dibuka dan dibatalkan dengan Escape (tidak ada shortcut yang diubah);
- pemilih Loader Icon, Text Animation, dan tema kode dibuka lalu ditutup tanpa memilih.

Yang **tidak** dilakukan:
- tombol impor di Browser → Import (cookie dan password dari Chrome dan Safari) dan "Open System Settings";
- "Connect Cloudflare" di Sites. Keduanya melibatkan kredensial atau otorisasi akun, dan yang pertama mengubah pengaturan sistem.

Dua catatan keadaan: workspace sudah berisi pekerjaan pemilik produk (isi percakapan tidak dicatat), dan semua halaman bagian ini berbunyi "per-device" (per perangkat), tidak per workspace, kecuali Databases (§4).

## 1. Appearance

Subjudul: "How the IDE looks while it's working — per-device preferences." Empat bagian. Bukti: `a4-appearance-1-top.jpg`, `a4-appearance-2-code-appearance.jpg`.

### 1.1 THINKING LOADING

Dua baris dengan nilai saat ini di kiri tombol dan tombol "Change" (ikon di kiri label) di kanan.
- **Loader Icon**: "The animated icon shown next to the Thinking... label while an agent is working." Nilai: "Ring" dengan pratinjau ikon.
- **Text Animation**: "How the Thinking... label itself animates during agent work." Nilai: "Blink" dengan contoh "Thinking...".

**Pemilih Loader Icon** (modal ±660 px lebar, bergulir). Bukti: `a4-appearance-loader-icon-picker.jpg`.
- Judul "Choose Loader Icon", subjudul "Pick the animation rendered next to the Thinking... label while an agent is working."; kolom cari "Search animations... (e.g. ring, dots)".
- Grid 4 kolom. Tiap kartu menampilkan ikon beranimasi, label "Thinking..." dan nama huruf kapital kecil. Kartu terpilih berbingkai terang dengan tanda centang di pojok.
- Nama yang terlihat (24 dari lebih banyak): Ring (terpilih), Spokes, Classic, Dots Ring, Spiral, Swirling, Arc, Dual Arc, Concentric Ring, Orbit Ring, Satellite Ring, Clock Ring, Quarter Ring, Dash Ring, Comet, Pulse, Ripple, Pulse Dot, Dots, Bobbing Dots, Bouncing Dots, Typing, Twin Orbit, Pulsating Dots. Daftar bisa digulir.

**Pemilih Text Animation**: judul "Choose Text Animation", kolom cari, dan empat kartu dalam dua kolom: **Blink** (terpilih), **Dots**, **Shimmer**, **Shimmer Wave**. Bukti: `a4-appearance-text-animation-picker.jpg`.

**[inferensi]** Pilihan ini hanya kosmetik, tapi penting bagi rasa produk: indikator "agent sedang bekerja" dibuat bisa diganti dan dipratinjau.

### 1.2 APPEARANCE MODE

- **Color Mode**: segmented **System · Light · Dark** (aktif: Dark). Teks bantuan: "Light/dark for the app. Today this switches the code viewer + diff theme (and Nuxt UI controls); the rest of the chrome stays dark for now."
- **[terlihat langsung]** Teks itu **sudah usang**. Memilih Light mengubah seluruh antarmuka jadi terang (§1.5), bukan hanya penampil kode. **[inferensi]** Frontend memakai pustaka komponen yang disebut "Nuxt UI" (sesuai teks itu).

### 1.3 ZOOM

- **Zoom Level**: stepper `− 100% +` dan tombol "Reset to 100%" (ikon putar). "Scales the entire interface. Applies to the whole app — not just this workspace — and is remembered after you quit. Shortcuts: `⌘=` to zoom in, `⌘-` to zoom out, `⌘0` to reset."
- **[terlihat langsung]** Memperkecil zoom lewat menu View → Zoom Out menampilkan lebih banyak konten per layar dan semua tata letak tetap utuh (tiga kali Zoom Out; persen pastinya tidak dibaca). Ini berguna untuk membaca daftar panjang (dipakai untuk daftar tool MCP di A3).

### 1.4 CODE APPEARANCE

- Teks: "Themes for the code viewer and diffs. The light theme is used in light mode, the dark theme in dark mode."
- Dua dropdown berdampingan: **Light theme** (default "GitHub Light") dan **Dark theme** (default "Atom One Dark"). Di bawah tiap dropdown ada pratinjau empat baris kode (`function greet(name: string) {…}`) dengan satu baris dihapus (merah) dan satu ditambah (hijau).
- Opsi Dark theme (daftar bisa dicari): Atom One Dark (terpilih), Dracula, Monokai, Monokai Dimmed, GitHub Dark. Bukti: `a4-appearance-dark-theme-options.jpg`.
- **[terlihat langsung]** Pemilihan tema kode terpisah dari mode warna dan hanya memengaruhi penampil kode dan diff.

### 1.5 Pengujian mode terang

Color Mode diganti ke Light, diamati, lalu dikembalikan ke Dark. Bukti: `a4-appearance-light-mode.jpg`, `a4-ide-light-mode.jpg`, `a4-settings-general-light-mode.jpg`.

- **[terlihat langsung]** Seluruh antarmuka berubah: latar jendela putih keabu-abuan sangat terang, kartu dan panel putih atau abu sangat muda, teks hitam, pegangan garis tepi abu muda. Struktur, jarak, dan ukuran sama persis dengan mode gelap.
- **[terlihat langsung]** Aksen tetap konsisten: toggle hidup hijau, tombol hapus merah, lencana "error" merah muda, peringatan amber. Pratinjau tema kode Dark tetap gelap di tengah halaman terang (karena tema kode dipilih terpisah).
- **[terlihat langsung]** Di IDE mode terang, panel kanan kini memuat tab "Review" di samping "Current Tasks" dan "Agents" (tab Review muncul karena ada perubahan file; dikerjakan di A5).
- Nilai warna diukur di A8.

## 2. Keyboard

Bukti: `a4-keyboard-*.jpg`.

### 2.1 Halaman dan interaksi

- Subjudul: "Per-device hotkey bindings for the Command Palette. Click Edit on a row and press the combo you want, or press Backspace to unbind it."
- Kanan atas: tombol "Import" (ikon unggah), "Export" (ikon unduh), dan tautan teks "Reset all". Tidak dijalankan.
- Daftar dikelompokkan dengan label bagian kapital kecil dan kartu per bagian. Satu baris: ikon, judul, deskripsi abu opsional, di kanan keycap (⌘ ⇧ ⌥ + huruf, satu kotak per tombol) atau teks miring abu "No hotkey", dan tombol "Edit" (garis luar).
- **Mode penangkapan** (klik Edit): bagian kanan baris berubah menjadi teks amber miring "Press a combo... (Esc to cancel · ⌫ to clear)" dengan tombol "Cancel". Escape membatalkan, dan baris kembali seperti semula. Bukti: `a4-keyboard-capture-state.jpg`.
- **[terlihat langsung]** Daftar ini **sama dengan isi Command Palette** (A1 §5): setiap perintah palet bisa diberi atau dilepas shortcut. Tidak ada daftar shortcut terpisah dari palet.

### 2.2 Perintah dan shortcut bawaan

Hanya perintah yang punya shortcut bawaan disebut tombolnya; sisanya "No hotkey".

| Bagian | Perintah | Shortcut |
| --- | --- | --- |
| Workspace | New Workspace | `⌘N` |
| | New Workspace in New Window | `⌘⌥N` |
| | Switch Workspace… | `⌘K` |
| | Archived Workspaces, Archive Current Workspace, Delete Current Workspace | tanpa |
| Navigation | New Chat | `⇧⌘N` |
| | Close Tab ("The tab in focus: terminal, sidebar or chat") | `⌘W` |
| | Save File | `⌘S` |
| | Focus Editor Group 1, 2, 3 | `⌘1`, `⌘2`, `⌘3` |
| | Go to File Changes | `⇧⌘G` |
| | Filter Changed Files ("File Changes or the Review tab") | `⌘F` |
| | **Find in Diff** ("The file open on File Changes") | `⌘F` |
| | Open Terminal ("Reveal the terminal panel (spawns a shell if none)") | `⌘J` |
| | Open Files Panel | `⇧⌘F` |
| | Open Browser Panel | `⇧⌘O` |
| | Go to Kanban | `⇧⌘K` |
| | Go to Chat, Next/Previous Chat Tab, Open Current Tasks Panel, Open Agents Status Panel, Go to Containers / Docker Compose / Volumes / Images / Networks / Database Manager / Canvas, View: Open Problems | tanpa |
| View | Zoom In, Zoom Out, Reset Zoom ("Back to 100%") | `⌘=`, `⌘-`, `⌘0` |
| | View: Toggle Left Sidebar, View: Toggle Right Sidebar | tanpa |
| Chat | Chat: Stop Generating, Toggle Plan Mode, Change Model…, Attach File…, **Rename Current Chat…** | tanpa |
| Database | Open Query History, New Query Console, New Connection…, **Cancel Query** | tanpa |
| Git | **Commit, Push, Pull, Fetch, Checkout Branch…, Create Branch…, Open Source Control** | tanpa |
| Terminal | Create New Terminal, Kill Active Terminal | tanpa |
| Browser | New Tab, **Reload**, **Hard Reload** ("Reload the page, bypassing the cache"), Open URL…, History, Downloads, Clear Browsing Data, Import Cookies and Passwords | tanpa |
| Engine | **Start, Stop, Restart**, Open Logs ("The engine and app log folder, in the file manager") | tanpa |
| Settings | Open Settings | `⌘,` |
| | Go to Knowledge Base / Rules / Skills (Context → …), serta satu perintah untuk tiap halaman Settings (General sampai Sites) | tanpa |
| Preferences | Color Theme: Light / Dark / System | tanpa |
| Create | Add Agent…, Add Knowledge Item…, Add Rule…, Install Skill… | tanpa |
| Indexing | Re-Index Workspace, Clear Code Index… | tanpa |
| Kanban | New Task ("Create a new Kanban ticket") | `⌘T` |
| | Open Kanban Filter | `⌘F` |
| Palette | Open Command Palette | `⌘P` |
| Account | Sign Out | tanpa |

- **[terlihat langsung]** Yang tidak muncul di palet saat A1 tetapi ada di daftar Keyboard (dicetak tebal di atas): Find in Diff, Chat: Rename Current Chat, Database: Cancel Query, seluruh bagian Git, Browser: Reload dan Hard Reload, dan Engine: Start, Stop, Restart. Jadi palet menyembunyikan perintah yang tidak relevan dengan konteks layar saat itu, sedangkan halaman Keyboard menampilkan semuanya. **[inferensi]**
- **[terlihat langsung]** `⌘F` dipakai tiga kali (Filter Changed Files, Find in Diff, Open Kanban Filter). Artinya shortcut bergantung konteks layar yang aktif. Halaman ini tidak menandai bentrokan; apakah penetapan ganda ditolak tidak diuji.

## 3. Browser

Subjudul: "How the in-app browser searches, saves downloads and keeps history, what agents may do in it, and which engine runs it. Per-device preferences." Empat tab: **Settings · History · Downloads · Import** (segmented, tab aktif berlatar abu terang). Bukti: `a4-browser-*.jpg`.

### 3.1 Tab Settings (tujuh bagian)

| Bagian | Baris | Default | Teks bantuan, ringkas |
| --- | --- | --- | --- |
| GENERAL | Search engine | Google (dropdown) | "Used when what you type in the address field is not a web address." |
| | Web URL and link open destination | Default browser (dropdown) | "Where web links from chats, terminals and cards open." |
| | Local URL open destination | Sokudo (dropdown) | "Where local development sites open by default." |
| | Show full URL | Mati | "Include the path, query and fragment in the address bar." |
| | Annotation screenshots | Always include (dropdown) | "Screenshots help the agent understand and address your comments, but add image tokens to each message." |
| DOWNLOADS | Downloads folder | `/Users/mac/Downloads` + tombol "Change" | |
| | Ask where to save each file | Mati | "Show the save dialog for every download instead of saving to the folder above." |
| HISTORY | Record history | **Hidup** | "Keep a list of the pages you open, on this device only, for 90 days. Turning it off stops new entries and keeps the existing ones." |
| | Browsing history | tombol "Open history" | "Search, open or remove pages you visited." |
| | Clear browsing data | tombol "Clear browsing data" | "History, the download list, cookies and site data, and cached files, for a time range you pick." |
| AUTOFILL AND PASSWORDS | Offer to save passwords | **Hidup** | "Ask to save a password when you sign in to a site. Passwords stay in the browser engine's own store." |
| | Password manager | tombol "Manage" | "Add, delete, and edit saved passwords." |
| | Save and fill contact info | **Hidup** | "Addresses, phone numbers and email addresses are offered on your tabs." |
| | Contact info | tombol "Manage" | |
| EXTENSIONS | Manage extensions | tombol "Manage extensions" | "Turn extensions on or off, or remove them." |
| PERMISSIONS | Approval | Always ask (dropdown) | "Choose if the app asks for approval before opening websites." |
| | Site permissions | tombol "+ Add" | "Override the defaults above for specific sites. Only sites with custom permissions appear here." |
| DEVELOPER MODE | Enable full CDP access | **Mati** | Lencana amber berikon perisai "ELEVATED RISK". "Allow the app to use full Chrome DevTools Protocol (CDP) access in connected Browser Use sessions. Full CDP access lets the app inspect and control sensitive browser internals that may put your data at risk." |
| ENGINE | Engine | Teks: "Chromium (CEF, Chrome style), version `<angka>`" | "The browser engine that runs pages on this device." |

- **[terlihat langsung]** Browser ini adalah engine Chromium sungguhan di dalam aplikasi, lengkap dengan password manager, ekstensi, izin per situs, dan CDP.
- **[terlihat langsung]** Pola keamanan: kemampuan tingkat tinggi (CDP penuh) diberi lencana risiko dan mati secara default. Ini sejalan dengan sakelar browser agent di A3 (Agent browser control dan JS eval, keduanya mati).

### 3.2 Tab History

Subjudul: "Pages opened in the in-app browser on this device, from every workspace. Agent visits are kept too and shown on request." Kolom cari "Search history", dua chip penyaring "Include agent visits" (ikon robot) dan "This workspace" (ikon folder), tombol "Clear browsing data" di kanan, hitungan "No visits" di kanan baris chip. Keadaan kosong: "No browsing history yet." Bukti: `a4-browser-tab-history.jpg`.

### 3.3 Tab Downloads

Subjudul: "Files downloaded in the in-app browser. Removing an item or clearing the list never deletes the file on disk." Kolom cari "Search downloads by file name", tombol "Clear list", hitungan "No downloads", kosong "No downloads yet." Bukti: `a4-browser-tab-downloads.jpg`.

### 3.4 Tab Import

Hanya dibaca; tidak ada yang diimpor. Bukti: `a4-browser-tab-import.jpg`.
- Subjudul: bawa cookie dari browser lain di perangkat ini; password lewat berkas yang diekspor; "Nothing is read until you press Import." Catatan: "Sokudo reads only what you import, only when you press Import."
- Kartu peringatan berikon perisai: "Allow Full Disk Access to import from Safari" dengan teks bahwa Safari menyimpan cookie di folder terlindungi sehingga Sokudo perlu Full Disk Access di System Settings. Tombol "Not now" dan tombol biru "Open System Settings" (tidak diklik).
- Kartu **Google Chrome · Default** dan **Safari**: baris Cookies dan baris Passwords. Cookie: tombol "Import cookies" (Chrome: peringatan amber "Close Google Chrome to import."). Password: tombol "Import passwords" dan tiga langkah bernomor (ekspor CSV dari browser asal, tekan Import passwords lalu pilih berkas di tab Passwords, hapus CSV setelah selesai karena berisi password polos). "Sokudo never reads, copies or deletes this file. Only the Passwords tab opens it, only when you pick it there."
- Di bawah: tautan "Export passwords" untuk mengeluarkan password dari browser Sokudo ke CSV.

## 4. Databases

Subjudul: "Runtimes and drivers Sokudo downloaded for engines that need them, such as DuckDB and the JDBC engines; every workspace shares them. For this workspace, how long an agent's change is held before it is rolled back, and what the query history keeps." Bukti: `a4-databases.jpg`.

| Bagian | Isi |
| --- | --- |
| MANAGED RUNTIMES | Keadaan kosong: "No runtime installed. An engine that needs one offers to download it when you set up its data source." |
| AGENT CHANGES | **Hold an agent's change before rolling it back**, dropdown default "5 minutes". "A change an agent runs waits in a transaction for your Commit. If you do not decide within this time (1–30 minutes), it is rolled back. A new length applies to changes that start after it; this workspace only." |
| QUERY HISTORY | Teks saja: "Open the Database Manager for this workspace in this window to change its query history settings." |

- **[terlihat langsung]** Ini satu-satunya halaman bagian ini yang berlaku **per workspace** (selain runtime yang dibagi).
- **[terlihat langsung]** Pola persetujuan: perubahan yang dijalankan agent ke database **menunggu di dalam transaksi sampai pengguna menekan Commit**; bila tidak diputuskan dalam 1–30 menit, otomatis dibatalkan (rollback). Selaras dengan sakelar "Allow agents to write to databases" di A3 (mati secara default).

## 5. Sites

Subjudul: "Publish canvas projects to your own Cloudflare Pages account and manage your live sites." Bukti: `a4-sites-checking.jpg`, `a4-sites-connect-cloudflare.jpg`.

- **CONNECTION**: awalnya satu baris dengan putaran dan teks "Checking connection..."; lalu menjadi kartu "Connect Cloudflare": "Link your Cloudflare account to publish canvas projects as live Cloudflare Pages sites. Sokudo never hosts them — they deploy to and live in your own account." dan tombol biru "Connect Cloudflare" (ikon awan). **Tidak diklik** (alur otorisasi akun).
- **PUBLISHED SITES**: "Nothing published yet."
- **[terlihat langsung]** Publikasi hanya untuk proyek Canvas, ke akun Cloudflare milik pengguna. Sokudo tidak menyimpan hasil publikasi sendiri.

### 5.1 Sites setelah Cloudflare terhubung (2026-10-31)

Pemilik produk menghubungkan akun Cloudflare sendiri; saya hanya membaca tampilannya. Alamat email akun tidak dicatat di sini. Tidak ada yang dipublikasikan dan tombol "Disconnect" tidak ditekan. Bukti: `a4-sites-connected.jpg`.

- **CONNECTION** kini berupa kartu dengan ikon awan, nama akun ("`<email akun>`'s Account") dan lencana hijau "Connected". Teks: "Your canvas projects deploy to this Cloudflare account." dan "Disconnecting doesn't take your published sites down — they live in your Cloudflare account." Tombol merah "Disconnect" (ikon sambungan) di kanan.
- **PUBLISHED SITES** tetap "Nothing published yet." Menerbitkan situs butuh proyek Canvas; tidak diuji.
- **[terlihat langsung]** Sokudo memakai identitas akun pengguna (tampil sebagai nama akun) dan menyebut dengan jelas bahwa memutus koneksi tidak menurunkan situs yang sudah terbit.

## 6. Pola desain yang konsisten

**[terlihat langsung]**
1. Pilihan visual (ikon loader, animasi teks, tema kode) memakai pemilih berkartu dengan pratinjau langsung, bukan daftar teks.
2. Perintah dan shortcut punya satu sumber kebenaran: daftar perintah palet.
3. Pengaturan yang berisiko tinggi diberi lencana atau label "ELEVATED RISK" dan mati secara default.
4. Setiap aksi yang menyentuh data pengguna di luar aplikasi (impor, ekspor, hapus) menjelaskan apa yang dibaca, kapan, dan apa yang dihapus setelahnya.
5. Teks bantuan sesekali tertinggal dari perilaku (Color Mode). Ini pelajaran bagi Astra: tulis teks bantuan sebagai bagian dari definisi selesai komponen.

## 7. Dugaan arsitektur dan arti untuk Astra (bahan Fase B, belum keputusan)

| Temuan Sokudo | Keadaan Astra | **[usulan]** untuk Astra |
| --- | --- | --- |
| Appearance per perangkat: indikator "Thinking…" bisa dipilih (24+ ikon, 4 animasi teks) | `DESIGN_RULES.md` melarang animasi berlebih; belum ada setelan tampilan | Tiga opsi font sudah ada (`DESIGN_SYSTEM.md`). Tambah mode warna System/Light/Dark dan satu indikator "bekerja" bawaan. Banyaknya pilihan animasi **tidak** ditiru (ikon hanya Lucide; hormati `prefers-reduced-motion`) |
| Color Mode System/Light/Dark; mode terang menggeser seluruh UI | Token `web/src/styles/tokens.css` baru satu tema | Terapkan mode terang lewat token (`DESIGN_SYSTEM.md` §11) dan `prefers-color-scheme`. Diukur di A8 |
| Zoom seluruh UI dengan `⌘=`/`⌘-`/`⌘0`, diingat setelah keluar | Browser sudah punya zoom | Tidak perlu fitur sendiri; pastikan tata letak tahan zoom browser sampai 200% (sudah dicek di 390, 768, 1440) |
| Tema kode terpisah dari mode warna | Terminal xterm.js memakai palet tetap | P2: tema terminal dan diff mengikuti token, tanpa daftar tema ke-3 |
| Keyboard: semua perintah palet bisa dirangkai ulang, per perangkat, dengan Import/Export dan Reset all | Belum ada palet perintah | Palet perintah (P1) sebagai pintu semua aksi. Rangkai ulang shortcut disimpan di `localStorage` per perangkat (aman karena hanya preferensi) |
| `⌘F` berarti tiga hal tergantung layar | Belum ada | Tentukan aturan konteks di awal dan tampilkan bentrokan |
| Browser dalam aplikasi (Chromium) dengan riwayat, password manager, ekstensi, izin, CDP | Preview worktree dibuka di tab browser pengguna | **Tidak ditiru.** Web UI Astra memakai browser pengguna sendiri; preview dibuka sebagai tautan. Kontrol browser oleh agent (CDP) tidak diberikan |
| "Annotation screenshots" (komentar di halaman disertai tangkapan) | Rencana F-umpan balik klien (`feedback`) sudah memuat komentar di preview | Nilai konfigurasi "sertakan tangkapan" (selalu, saat diminta, tidak pernah), karena tangkapan menambah token |
| Pengaturan "Approval" sebelum membuka situs; izin per situs | Preview publik dengan token per URL | Tidak perlu |
| Databases: perubahan agent ke database ditahan dalam transaksi sampai Commit, rollback otomatis setelah 1–30 menit | Belum ada fitur database | Bila nanti ada, pakai pola yang sama; ini selaras dengan prinsip "tanpa approval manusia" |
| Sites: publikasi ke Cloudflare Pages milik pengguna lewat tautan akun | Deploy ke production dengan gerbang dan rollback (PRD); token Cloudflare tidak boleh tersedia untuk agent | Astra menyimpan token Cloudflare di daemon (terenkripsi, root:astrad) dan hanya daemon yang menjalankan deploy setelah approval. Bedanya: Sokudo mengotorisasi lewat akun pengguna; Astra memakai token yang dibuat admin dengan izin sekecil mungkin |
| Teks bantuan usang (Color Mode) | | Sertakan pemeriksaan teks bantuan dalam §13 `DESIGN_RULES.md` |

Catatan batas: ini ide fitur dan alur. Tidak ada teks bantuan, ikon, atau nama Sokudo yang disalin; semua kalimat bantuan Astra ditulis ulang.

## Butir terbuka

1. Opsi dropdown Browser: Search engine, Web URL open destination, Local URL open destination, Annotation screenshots, Approval.
2. Isi dialog "Password manager", "Contact info", "Manage extensions", dan "Site permissions → Add". Tidak dibuka karena menyangkut data pribadi.
3. Perilaku Import, Export, dan Reset all di Keyboard; penetapan ganda satu kombinasi (misalnya `⌘F`).
4. Daftar situs terbit dan alur publikasi dari Canvas (Cloudflare sudah terhubung, tapi belum ada proyek Canvas): A7.
5. Query history di Database Manager (dikerjakan di A7 bersama Database).
6. Nilai warna, jarak, dan tipografi mode gelap dan terang: A8.
