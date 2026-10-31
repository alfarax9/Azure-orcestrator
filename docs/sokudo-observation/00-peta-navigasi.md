# A1. Peta navigasi dan shell Sokudo

Pengamatan Sokudo 0.19.6 (macOS) pada 2026-10-30, lewat computer use, hanya di jendela Sokudo. Workspace yang terbuka: `test soduko` (milik pemilik produk). Saat A1 workspace itu masih kosong (tanpa repo git, agent, atau CLI tersambung). Pada 2026-10-31 pemilik produk mulai memakainya (chat dan satu agent berjalan), jadi pengamatan A1 menggambarkan keadaan kosong. Tidak ada item yang dibuat, diubah, atau dihapus oleh pengamatan ini.

Tanda:
- **[terlihat langsung]**: tampak di layar atau di pohon aksesibilitas.
- **[inferensi]**: dugaan dari yang terlihat, belum dibuktikan.

Bukti berupa nama berkas di `docs/sokudo-observation/screens/`. Folder itu tidak di-commit (`.gitignore`). Ukuran dalam titik (pt) diambil dari tangkapan 1:1 berukuran 1512×949 (`a1-shell-fullsize-1512.jpg`). Ukuran ini nilai pengamatan kasar (±2 pt), bukan hasil inspeksi kode. Ukuran yang lebih rinci, warna, dan tipografi dikerjakan di A8.

## 1. Bentuk jendela

- **[terlihat langsung]** Satu jendela utama berjudul "Sokudo", 1512×949 pt, mengisi layar di bawah menu bar. Tidak ada title bar atau tombol lampu lalu lintas yang terlihat; bilah atas aplikasi langsung menempel di bawah menu bar macOS. Bukti: `a1-shell-fullsize-1512.jpg`.
- **[terlihat langsung]** Seluruh isi jendela adalah satu web view (AXWebArea). Ini sesuai dengan stack Tauri.
- **[terlihat langsung]** Tema yang aktif gelap: latar jendela hampir hitam, panel abu-abu sangat gelap dengan sudut membulat.
- **[terlihat langsung]** Ukuran minimum jendela: **tidak ada**. Saya sendiri tidak bisa mengubah ukuran (seret di tepi layar memilih teks; memindahkan jendela butuh akses Finder), tapi pemilik produk mengecilkannya dan ukuran terbaca 105×280 pt. Lihat Butir terbuka 1.
- **[terlihat langsung]** Window → Zoom tidak mengubah ukuran (jendela sudah penuh).

## 2. Menu bar macOS

**[terlihat langsung]**, dibaca lewat menu bar tanpa menekan apa pun kecuali View → Command Palette dan Window → Zoom.

| Menu | Isi |
| --- | --- |
| Sokudo | About Sokudo, Hide Sokudo, Hide Others, Show All, Quit Sokudo |
| File | New Workspace, New Workspace in New Window, New Chat, Switch Workspace…, Close |
| Edit | Undo, Redo, Cut, Copy, Paste, Select All, AutoFill, Start Dictation…, Emoji & Symbols |
| View | Command Palette, Zoom In, Zoom Out, Reset Zoom, Go to Chat, Go to File Changes, Go to Kanban, Go to Containers, Go to Docker Compose, Focus Editor Group 1/2/3, Next Tab, Previous Tab, Reload Window |
| Window | Minimize, Zoom, Toggle Full Screen |

Tidak ada menu Help dan tidak ada Settings di menu Sokudo. Settings hanya dibuka dari ikon di bilah atas atau dari Command Palette (`⌘,`).

## 3. Kerangka layar IDE

Susunan dari atas ke bawah. Bukti: `a1-shell-chat-empty.jpg`, `a1-shell-fullsize-1512.jpg`, `a1-zoom-top-bar.jpg`, `a1-zoom-status-bar.jpg`.

```
┌ bilah atas (±47 pt) ─────────────────────────────────────────────────────────┐
│ [workspace ▾] [◧]          [ Chat | Canvas | File Changes ]   [⛁ ⬡ ▦ ⚙ | 🔔 ◉] [◨] │
├──────────────┬──────────────────────────────────────────┬────────────────────┤
│ sidebar kiri │ area editor: tab + konten                 │ sidebar kanan      │
│ (±302 pt)    │                                           │ (±321 pt, opsional)│
│              ├──────────────────────────────────────────┤                    │
│              │ panel bawah (opsional)                    │                    │
├──────────────┴──────────────────────────────────────────┴────────────────────┤
│ bilah status (±25 pt)                                                         │
└──────────────────────────────────────────────────────────────────────────────┘
```

**[terlihat langsung]** Setiap area adalah panel sendiri dengan sudut membulat, dipisah celah ±10 pt dari latar jendela yang lebih gelap.

### 3.1 Bilah atas

Kiri ke kanan, label dari pohon aksesibilitas:

| Kontrol | Peran | Perilaku |
| --- | --- | --- |
| `test soduko ▾` | Pemilih workspace | Membuka menu: daftar workspace (titik warna + nama + centang untuk yang aktif), pemisah, "Archived Workspaces", "Add Workspace". Bukti: `a1-workspace-switcher.jpg` |
| ◧ "Toggle sidebar" | Buka/tutup sidebar kiri | Area editor melebar ke kiri. Bukti: `a1-left-sidebar-hidden.jpg` |
| Segmented control `Chat · Canvas · File Changes` | Mode area utama | Di tengah bilah. Segmen aktif berlatar abu terang dengan sudut membulat |
| ⛁ "Database" | Layar Database | Lihat §4.4 |
| ⬡ "Containers" | Layar Docker | Lihat §4.5 |
| ▦ "Projects" | Layar Kanban Board | Lihat §4.6 |
| ⚙ "Settings" | Layar Settings | Lihat §4.7 |
| 🔔 "Notifications" | Panel notifikasi | Titik kuning di ikon bila ada notifikasi. Lihat §6 |
| Avatar "User profile" | Menu akun | Isi: "Profile", "Sign Out". Tidak dipilih apa pun |
| ◨ "Toggle right sidebar" | Buka/tutup sidebar kanan | Hanya ada di layar IDE, tidak ada di Docker, Kanban, dan Settings |

**[terlihat langsung]** Empat ikon layar dan dua ikon akun dikelompokkan dalam satu kapsul berbingkai, dengan garis pemisah tipis di antara kedua kelompok.

**[terlihat langsung]** Di Docker, Kanban, dan Settings, bilah atas berganti: tombol "← Back to IDE" + judul layar ("Docker", "Kanban Board", "Settings") menggantikan segmented control, dan kapsul ikon bergeser ke kanan karena tombol sidebar kanan hilang. Di Database, tombol sidebar menjadi "Toggle explorer" dan sidebar kanan menjadi "Toggle database chat".

### 3.2 Sidebar kiri (mode Chat)

- **[terlihat langsung]** Tombol "+ New Chat" selebar sidebar (tombol garis luar berbentuk pil), label grup "FOLDER" dengan tombol "New folder" (muncul saat hover), lalu daftar chat. Keadaan kosong: "No chats yet." (miring, abu). Saat memuat: "Loading chats...". Bukti: `a1-shell-chat-empty.jpg`, `a1-right-panel-current-tasks.jpg`.
- **[terlihat langsung]** Di mode File Changes, isi sidebar berganti menjadi tiga bagian yang bisa dilipat: "REPOSITORIES (0)", "CHANGES (0)", dan "CONFLICTS (0)" dengan tombol Refresh, menempel di bawah. Keadaan kosong: "No git repositories found in this workspace." dan "No merge in progress." Bukti: `a1-file-changes-empty.jpg`.
- **[terlihat langsung]** Di mode Canvas, sidebar kiri berisi pesan kosong "Open a canvas project to see its design nodes". Bukti: `a1-canvas-empty.jpg`.

### 3.3 Area editor dan tab

- **[terlihat langsung]** Tab chat di bagian atas panel: ikon + "New Chat"; tab aktif sedikit lebih terang dan menyatu dengan isi panel. Tab punya tombol tutup ("Close tab") saat hover. Ada "+" (New chat) setelah tab terakhir.
- **[terlihat langsung]** Di kanan bar tab ada "Split editor right" dan "Split editor down". Split kanan membuat grup editor kedua dengan bar tab dan tombol × sendiri. Bukti: `a1-editor-split-right.jpg`. Command Palette menyebut sampai tiga grup editor (`⌘1`–`⌘3`).
- **[terlihat langsung]** Keadaan kosong chat: judul besar di tengah "What you want to build today?" dan composer di bawahnya.
- **[terlihat langsung]** Composer (±680×143 pt, sudut membulat, berbingkai):
  - placeholder "Ask anything · @ to mention anything · / to use a skill";
  - baris bawah: "+" (More actions), pemilih agent ("No agents or CLIs available ▾"), pemilih akses (ikon perisai oranye + "Full access ▾"), dan tombol kirim bulat (panah atas, nonaktif saat kosong);
  - "+" membuka: "Add photos & files" dan "Plan mode" (toggle, mati). Bukti: `a1-composer-more-actions.jpg`;
  - pemilih akses: "Full access" (dicentang) dan "Default permissions". Bukti: `a1-composer-access-menu.jpg`;
  - **[terlihat langsung]** Saat area editor sempit (split), label "Full access" menyusut menjadi ikon perisai saja dan placeholder terpotong dengan "…". Bukti: `a1-editor-split-right.jpg`.

### 3.4 Sidebar kanan

- **[terlihat langsung]** Berisi tab panel berbentuk pil: "Current Tasks" dan "Agents Status", plus tombol "+" untuk menambah panel. Menu "+": Current Tasks, Agents Status, Files, Browser. Setiap tab punya tombol tutup ("Close …"). Bukti: `a1-right-panel-current-tasks.jpg`, `a1-right-panel-agents-status.jpg`, `a1-right-panel-add-menu.jpg`.
- Keadaan kosong: "No tasks yet." dan "No custom agents yet — add one in Settings → Agents."
- **[terlihat langsung]** Di mode Canvas, sidebar kanan berisi pesan kosong "Open a canvas project to start designing with AI".

### 3.5 Panel bawah

- **[terlihat langsung]** Dibuka dari bilah status (ikon masalah) atau `⌘J` (Open Terminal). Tab: "Agent Mode", "Problems", "Terminal", dan tombol tutup panel. Ada pegangan geser (garis pendek) di antara editor dan panel bawah. Bukti: `a1-bottom-panel-*.jpg`.
- Agent Mode, kosong: "No worker activity yet." (monospace).
- Problems: judul "PROBLEMS", tombol filter, collapse, dan clear. Kosong: "No problems in the files changed or opened recently. Language servers analyze a file after it changes or when you open it."
- Terminal:
  - toolbar: "New terminal" (+ dengan dropdown), pilihan tata letak, "Mention this terminal in chat" (@), "Split terminal", "Kill terminal";
  - daftar sesi di kanan ("zsh");
  - prompt shell nyata di folder workspace. Tidak ada perintah yang diketik.

### 3.6 Bilah status

Kiri ke kanan. Bukti: `a1-zoom-status-bar.jpg`, `a1-status-*.jpg`.

| Item | Isi saat diamati | Klik membuka |
| --- | --- | --- |
| ⊗ 0 ⚠ 0 | Jumlah error dan peringatan | Panel bawah, tab Problems |
| `</>` Indexed | Status indeks kode | Settings → General (bagian Code Index Status) |
| 🔗 0 | (angka) | Settings → Token Usage, tab Overview **[inferensi: penghitung token atau biaya]** |
| 🤖 0 agents | Jumlah agent | Settings → Agents |
| ◔ Checking usage | Label AX: "AI usage limits — no data yet" | Settings → Token Usage, tab Provider Usage |
| ● Docker Running | Status engine (titik hijau) | Settings → Docker |
| ● MCP | Status server MCP lokal | Settings → MCP |
| ● LSP: Idle | Language server | Popover "LANGUAGE SERVERS: No language server is running for this workspace." |
| ● Online | Konektivitas | Tidak ada efek yang terlihat |
| ✻ Ready | Label AX: "Background scans — none running" | Popover "BACKGROUND SCANS: No background scans." |

**[terlihat langsung]** Pola: setiap item status adalah pintu ke halaman Settings yang relevan atau ke popover kecil. Popover tidak tertutup dengan Escape. Popover LSP tetap menempel sampai item itu diklik lagi.

## 4. Layar tingkat atas

| # | Layar | Masuk lewat | Keluar lewat | Bukti |
| --- | --- | --- | --- | --- |
| 4.1 | IDE, mode Chat | Default; segmen Chat; View → Go to Chat | | `a1-shell-chat-empty.jpg` |
| 4.2 | IDE, mode Canvas | Segmen Canvas; palette "Go to Canvas" | | `a1-canvas-empty.jpg` |
| 4.3 | IDE, mode File Changes | Segmen; `⇧⌘G` | | `a1-file-changes-empty.jpg` |
| 4.4 | Database | Ikon ⛁; palette "Go to Database Manager" | "← Back to IDE" | `a1-database-empty.jpg` |
| 4.5 | Docker | Ikon ⬡; View → Go to Containers / Docker Compose | "← Back to IDE" | `a1-docker-engine-gate.jpg` |
| 4.6 | Kanban Board | Ikon ▦ (label "Projects"); `⇧⌘K` | "← Back to IDE" | `a1-kanban-board-empty.jpg`, `a1-kanban-list-empty.jpg` |
| 4.7 | Settings | Ikon ⚙; `⌘,`; item bilah status | "← Back to IDE" | `a3-agents-settings-empty.jpg`, `a2-general-settings-top.jpg` |

### 4.2 Canvas

**[terlihat langsung]** Tiga kolom. Tengah: judul "Canvas", subjudul "Design projects in this workspace. Open one to design with AI.", kartu bergaris putus-putus "+ New project", dan teks kosong "No projects yet — create one to start designing with AI." Detail di A7.

### 4.4 Database

**[terlihat langsung]** Tiga kolom:
- kiri: tombol utama biru "+ Add New Source", kolom "Filter sources by name", toggle "Favorites only" (bintang), kosong "No sources yet";
- tengah: "No console open" dengan tombol "+ New console" dan "Query history";
- kanan: chat database, dengan pemilih "No active source ▾", tombol riwayat dan chat baru, teks "No messages yet — send your first prompt below.", dan composer kecil.

**[inferensi]** Ini satu-satunya tombol berwarna (biru) di layar kosong yang diamati; tombol lain netral.

### 4.5 Docker

**[terlihat langsung]**
- Sidebar: tombol pil "Compose", lalu navigasi Containers (aktif), Volumes, Images, Networks.
- Isi: kartu tengah dengan ikon petir kuning, judul "Switch to Sokudo engine", teks "Sokudo's Docker tools are available only with the managed Sokudo engine.", "Active engine: dockerDesktop" (monospace), dan tombol biru "Switch to Sokudo".
- Tombol itu **tidak diklik**, karena mengubah engine.
- Docker tidak punya bilah status.

### 4.6 Kanban Board

**[terlihat langsung]**
- Kolom cari "Search tasks by id, title, assignee, or label...", tombol Filter, dan toggle "Board | List".
- Board: empat kolom dengan titik warna: Backlog (abu), In Progress (biru), Under Review (kuning), Done (hijau), masing-masing dengan "(0 items)". Tiap kolom punya tombol "Create" yang hanya muncul di pohon AX, kemungkinan saat hover. Done tidak punya tombol Create.
- List: "(0 items)", "Group by: None ▾", kosong "No tasks yet."

Detail di A5.

### 4.7 Settings

**[terlihat langsung]**
- Sidebar kiri: kolom "Search settings...", tombol "Collapse sidebar", dan grup:
  - **WORKSPACE**: General, Context, Docker, Environment Variables, Remote SSH;
  - **AGENTS**: Agents, Providers, MCP, Token Usage;
  - **APPLICATION**: Appearance, Keyboard, Browser, Databases, Sites.
- Isi kanan: judul halaman, subjudul, label bagian berhuruf kapital kecil berjarak ("AGENTS", "SPEC-DRIVEN DEVELOPMENT"), dan kartu berisi baris pengaturan (judul tebal, deskripsi abu, kontrol di kanan).

Detail per halaman di A2, A3, dan A4. Temuan yang sudah terlihat sambil lalu dicatat di §7.

## 5. Command Palette (`⌘P`)

**[terlihat langsung]**
- Modal di tengah atas dengan kolom "Type a command or search..." dan tombol ×. Item dikelompokkan dengan judul grup; tiap item punya ikon, judul, deskripsi abu opsional, dan shortcut sebagai keycap kecil.
- Item yang merusak (Delete Current Workspace, Kill Active Terminal, Clear Code Index, Sign Out) tampil redup.
- Item di bawah kursor ikut tersorot saat menggulir.

Bukti: `a1-command-palette-1.jpg` sampai `-5.jpg`. Tidak ada perintah yang dijalankan.

| Grup | Perintah (shortcut) |
| --- | --- |
| Workspace | New Workspace (`⌘N`), New Workspace in New Window (`⌘⌥N`), Switch Workspace… (`⌘K`), Archived Workspaces, Archive Current Workspace ("Hide from the switcher; keep files"), Delete Current Workspace ("Remove from your account and trash the folder"), daftar workspace (nama + path) |
| Navigation | Go to Chat, New Chat (`⇧⌘N`), Close Tab (`⌘W`, "terminal, sidebar or chat"), Save File (`⌘S`), Next/Previous Chat Tab, Focus Editor Group 1–3 (`⌘1`–`⌘3`), Go to File Changes (`⇧⌘G`), Filter Changed Files (`⌘F`), Open Terminal (`⌘J`), Open Current Tasks Panel, Open Agents Status Panel, Open Files Panel (`⇧⌘F`), Open Browser Panel (`⇧⌘O`), Go to Kanban (`⇧⌘K`), Go to Containers, Go to Docker Compose, Go to Volumes, Go to Images, Go to Networks, Go to Database Manager, Go to Canvas, View: Open Problems |
| View | Zoom In (`⌘=`), Zoom Out (`⌘-`), Reset Zoom (`⌘0`), Toggle Left Sidebar, Toggle Right Sidebar |
| Chat | Toggle Plan Mode, Change Model…, Attach File… |
| Database | Open Query History, New Query Console, New Connection… |
| Terminal | Create New Terminal, Kill Active Terminal |
| Browser | New Tab, Open URL…, History, Downloads, Clear Browsing Data, Import Cookies and Passwords |
| Engine | Open Logs ("The engine and app log folder, in the file manager") |
| Settings | Open Settings (`⌘,`), Go to Knowledge Base / Rules / Skills (semuanya di Context), lalu satu perintah per halaman Settings (General sampai Sites) |
| Preferences | Color Theme: Light / Dark / System |
| Create | Add Agent…, Add Knowledge Item…, Add Rule…, Install Skill… |
| Indexing | Re-Index Workspace, Clear Code Index… |
| Kanban | New Task (`⌘T`), Open Kanban Filter (`⌘F`) |
| Palette | Open Command Palette (`⌘P`) |
| Account | Sign Out |

**[inferensi]** `⌘F` dipakai dua kali (Filter Changed Files dan Open Kanban Filter), jadi shortcut bergantung konteks layar. Daftar lengkap dan cara mengubahnya dicek di A4 (Settings → Keyboard).

## 6. Notifikasi

- **[terlihat langsung]** Panel in-app dari ikon lonceng: judul "Notifications", item dikelompokkan per sumber. Contoh grup "PROVIDERS": "OpenCode CLI update available, v1.18.18 → v1.19.5". Titik kuning di lonceng menandai notifikasi belum dibaca. Bukti: `a1-notifications.jpg`.
- **[terlihat langsung]** Notifikasi desktop diatur di Settings → General:
  - "Notifications: Get desktop alerts for approvals, agent questions, failures, and completed work." (toggle, mati);
  - "Notify when a turn finishes: A desktop alert when a chat's turn finishes or fails. It names the chat and the agent, never the answer." (pilihan, "Never");
  - peringatan kuning "Notifications are off for Sokudo in System Settings." dengan tautan "Open system settings".

  Bukti: `a2-general-settings-top.jpg`.
- **[inferensi]** Empat jenis kejadian yang memicu notifikasi: approval, pertanyaan agent, kegagalan, dan pekerjaan selesai. Ini cocok dengan kosakata status Astra (`DESIGN_SYSTEM.md` §10), dan Astra sudah mendeteksi "needs input".

## 7. Temuan sambil lalu untuk kelompok berikutnya

Catatan: halaman Settings bagian Workspace (General, Context, Docker, Environment Variables, Remote SSH) sudah dikerjakan lebih rinci di `10-settings-workspace.md`. Butir A2 di bawah dipertahankan hanya sebagai jejak. Halaman Agents, Providers, MCP, dan Token Usage (A3) ada di `20-settings-agents.md`; butir A3 di bawah juga hanya jejak.

Dicatat di sini supaya tidak hilang; diperdalam di kelompok masing-masing.

- **A3, Agents:**
  - "AGENTS": kosong, dengan tombol "+ Add Agent" dan menu ⋮.
  - "SPEC-DRIVEN DEVELOPMENT": Requirements (max critic ↔ reviser rounds), Design, dan Tasks (max review rounds), masing-masing stepper − 3 +.
  - "AGENT ORCHESTRATION": Max concurrent workers, stepper − 4 +. Teksnya: "How many swarm workers may run at the same time. Extra workers wait in the queue and start as slots free up…".
  - "PROJECT SCRIPTS": "+ Add script" dan "Save". Penjelasannya: script yang ditandai "Run on worktree create" berjalan di setiap worktree swarm baru sebelum agent mulai. Bila gagal, worker gagal dan worktree dihapus.

  Bukti: `a3-agents-settings-empty.jpg`.
- **A3, dialog Add Agent** (sudah terbuka saat pengamatan dimulai, ditutup dengan Cancel):
  - Agent Name (placeholder "e.g., Code Reviewer Bot");
  - Role ("Select a role...");
  - Provider (kosong: "No CLIs connected. Connect a Local CLI above to make it available as a provider.");
  - System Prompt (Optional): "Longer prompts improve clarity but may increase latency.";
  - Fallback Agent (Optional): "Select a provider first". Petunjuknya: "If this agent's provider hits its rate/session limit, runs switch to this agent.";
  - tombol Cancel dan Create Agent.

  Bukti: `a3-agents-add-agent-provider-open.jpg`.
- **A3, MCP:**
  - Halaman read-only untuk server MCP lokal: Status "Ready", Bind address `127.0.0.1:<port>` (loopback).
  - Supported agent CLIs: antigravity, claude-code, codex, grok, kiro, muse, opencode, pi.
  - Registered tools: 110, dikelompokkan per kategori (contoh: "Docker resources", `READ_DOCKER · 7`), dengan nama tool monospace dan argumennya.

  Bukti: `a3-mcp-settings-top.jpg`.
- **A3, Token Usage:**
  - Banner info biru: "Usage reflects what each provider reports…". Bila adapter hanya melaporkan token, biaya dihitung dari tarif publik dan ditandai "~".
  - Tab Overview dan Provider Usage.
  - Rentang Today / 7 days / 30 days / All time dan pemilih tanggal.
  - Kartu TODAY, THIS WEEK, THIS MONTH; PER-DAY SPEND; BY MODEL ("Aggregate spend per session-key (CLI + model)").
  - Provider Usage: "Subscription limits reported by each connected CLI. These are per-account, not per-workspace."

  Bukti: `a3-token-usage-overview.jpg`, `a3-token-usage-provider.jpg`.
- **A2, General:**
  - Workspace Name dan Workspace Path (Open, Change…).
  - Notifikasi (§6).
  - Code Index Status (Re-Index, Clear Index merah).
  - Language servers (diagnostics), toggle hidup.
  - Auto-fix diagnostics after a turn: "When a turn leaves errors in the files it changed, Sokudo asks the agent to fix them (up to 2 rounds)", toggle hidup.

  Bukti: `a2-general-settings-top.jpg`.
- **A2, Docker:**
  - Engine: Source (Docker Desktop), Managed engine version, Data directory size, Data directory.
  - Switch to Sokudo engine.
  - System resources dan Maintenance (Reinstall engine).

  Bukti: `a2-docker-settings-top.jpg`.

## 8. Arti untuk Astra (bahan Fase B, belum keputusan)

| Pola Sokudo | Keadaan Astra sekarang | Catatan |
| --- | --- | --- |
| Bilah atas: pemilih workspace, segmented control mode, kapsul ikon layar, akun | Board tanpa bilah atas seperti ini | Kandidat kerangka web UI baru |
| Tiga panel (sidebar kiri, editor bertab, sidebar kanan berpanel) + panel bawah + bilah status | Board dan detail worker | Pola utama yang ditiru |
| Bilah status sebagai pintu ke Settings | Belum ada | Cocok untuk status daemon, jumlah worker, usage |
| Command Palette dengan grup, deskripsi, keycap, item merusak diredupkan | Belum ada | Kandidat P1 |
| Settings dengan sidebar bergrup dan kolom cari | Belum ada | Dibutuhkan untuk provider API (Fase C) |
| Kanban 4 kolom dengan Board dan List | Board Astra memakai kolom status §10 | Bandingkan kosakata status di A5 |
| Max concurrent workers | Belum ada batas | Sudah tercatat di rencana Fase D |
| Fallback agent per agent | Belum ada | Sudah tercatat di rencana Fase C |
| Notifikasi desktop untuk approval, pertanyaan, gagal, selesai | Status "needs input" ada, notifikasi belum | Web UI memakai Notification API browser |
| Docker, Database, Canvas, Browser in-app | Di luar lingkup M1 | Diputuskan di Fase B |

## Butir terbuka

1. **Ukuran minimum jendela: tidak ada** (terselesaikan 2026-10-31). Pemilik produk mengecilkan jendela sendiri dan melaporkan tidak ada batas minimum. Saya memverifikasi lewat daftar jendela: ukuran terbaca **105×280 pt**, dan UI tetap dirender (bilah atas, sidebar, dan bilah status terpotong, tanpa tata letak ringkas atau menu hamburger). Bukti: `a1-window-tiny-105x280.jpg`. **[inferensi]** Sokudo tidak menetapkan ukuran minimum jendela, dan tata letaknya tidak berubah menjadi ringkas pada lebar yang kecil. Untuk Astra, web UI harus tetap diuji di 390 pt (`DESIGN_RULES.md`) dan sebaiknya punya tata letak ringkas sendiri.
2. **Tema terang.** Belum dilihat (tema gelap aktif). Diamati di A4 (Appearance) dan A8.
3. **Keadaan berisi.** Catatan A1 hanya memuat keadaan kosong. Workspace kini berisi chat dan agent milik pemilik produk, tapi isi percakapannya sengaja tidak dicatat. Tampilan berisi (daftar chat, tab, kartu perubahan file, status agent) diamati di A5 dan A7 dengan item uji `ATM-TEST-` yang perlu persetujuan.
4. **Popover LSP** masih menempel di pojok kanan bawah Sokudo setelah pengamatan, karena hanya tertutup dengan mengklik item "LSP: Idle" lagi. Tidak mengubah data.
