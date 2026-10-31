# A3. Settings, bagian Agents

Pengamatan Sokudo 0.19.6 (macOS), 2026-10-31, lewat computer use, hanya di jendela Sokudo. Halaman: Agents, Providers, MCP, Token Usage. Tema gelap.

Tanda: **[terlihat langsung]** tampak di layar; **[inferensi]** dugaan; **[usulan]** bahan untuk Astra, belum keputusan.

Bukti: nama berkas di `docs/sokudo-observation/screens/` (tidak di-commit). Ukuran dalam piksel tangkapan 1383 lebar (skala ≈0,915 dari pt), kasar (±3). Ukuran presisi dikerjakan di A8.

Yang **tidak** dilakukan:
- tidak ada agent yang dibuat, diubah, atau dihapus. Agent "Pm" milik pemilik produk tidak dibuka di Edit;
- tidak ada CLI yang diinstal (semua tombol "Install" tidak diklik) dan tidak ada toggle yang digeser;
- kolom API key di Providers tidak diisi;
- dialog Add Agent dibuka dan ditutup dengan Cancel; memilih provider di form tidak menyimpan apa pun;
- satu baris "Add script" dibuat sebagai draf lalu dihapus dengan ikon sampah di baris itu, tanpa menekan Save;
- tombol ↗ di Recent Activity tidak diklik (membuka chat milik pengguna).

Catatan keadaan: saat A3 workspace sudah berisi pekerjaan pemilik produk (satu agent "Pm", CLI Claude Code dan OpenCode tersambung, ada usage). Isi percakapan tidak dicatat; hanya tampilan.

## 1. Agents

Subjudul: "Configure and manage AI agents, their permissions, and execution environments." Halaman bergulir; enam bagian. Bukti: `a3-agents-1-top.jpg`, `a3-agents-2-scripts-git.jpg`.

### 1.1 AGENTS (daftar agent)

- **[terlihat langsung]** Header bagian dengan menu ⋮ (tidak dibuka) dan tombol "+ Add Agent" (garis luar).
- **[terlihat langsung]** Satu agent tampil sebagai baris kartu ±75 px tinggi:
  - kiri: nama tebal ("Pm"), lencana biru berisi role ("Project Manager"), lalu baris abu "Provider: OpenCode CLI · Model: `<id model>`" (model monospace);
  - kanan: tombol sampah merah (kotak merah penuh berikon), tombol "Edit" (garis luar), dan toggle aktif (hijau).
- Keadaan kosong (A1): "No agents yet. Click Add Agent to create one."

### 1.2 SPEC-DRIVEN DEVELOPMENT

Tiga baris dengan stepper `− 3 +` (kotak ±82×28 px, tombol minus dan plus di sisi, angka di tengah). Default semuanya **3**.

| Baris | Teks bantuan |
| --- | --- |
| Requirements | "Max critic ↔ reviser rounds while drafting the requirements doc. It stops earlier once the critic has no objections — higher means more polish but slower." |
| Design | "Max review rounds while drafting the design doc." |
| Tasks | "Max review rounds while drafting the task breakdown." |

**[inferensi]** Alur spec punya tiga tahap berurutan (Requirements → Design → Tasks), tiap tahap berupa putaran penulis dan peninjau dengan batas putaran. Berhenti lebih awal bila peninjau tidak keberatan.

### 1.3 AGENT ORCHESTRATION

- Satu baris "Max concurrent workers", stepper default **4**. "How many swarm workers may run at the same time. Extra workers wait in the queue and start as slots free up — higher is faster but heavier on CPU and provider rate limits."
- **[inferensi]** "Swarm worker" adalah agent yang berjalan di worktree sendiri; ada antrean bila melebihi batas.

### 1.4 PROJECT SCRIPTS

- Header dengan tombol "+ Add script" dan "Save" (keduanya garis luar). Penjelasan: "Commands this project needs before an agent can work in it. The one marked **Run on worktree create** runs in every new swarm worktree before the agent starts — if it fails, the worker fails and the worktree is removed."
- Keadaan kosong: "No scripts yet. Add one to give swarm workers a provisioned worktree." (miring, abu)
- Baris script (draf, dilihat lalu dihapus). Bukti: `a3-agents-project-script-draft-row.jpg`:
  - kolom nama, placeholder "Install dependencies";
  - kolom perintah monospace, placeholder `bun install --frozen-lockfile`;
  - ikon sampah di ujung kanan baris nama;
  - toggle "Run on worktree create" (mati): "Runs in the new worktree before the agent's first turn, with `SOKUDO_PROJECT_ROOT` and `SOKUDO_WORKTREE_PATH` set. 10-minute limit."
- **[terlihat langsung]** Berbeda dari halaman lain, di sini ada tombol Save eksplisit (daftar script disimpan sebagai satu kesatuan).

### 1.5 GIT

Dua baris dengan dropdown bernilai "Auto":
- **Commit messages**: "Model behind "Generate with AI" in the File Changes commit box. Auto follows the model picked in the chat composer."
- **Merge conflict resolution**: "Model behind "Resolve with AI" on a conflicted file. Auto follows the model picked in the chat composer."

**[terlihat langsung]** Bagian ini tidak terlihat di A1 karena ada di bawah lipatan.

### 1.6 Dialog Add Agent

Modal ±457 px lebar, bergulir di dalam. Bukti: `a3-add-agent-*.jpg`. Isi, dari atas:

| Field | Bentuk | Detail |
| --- | --- | --- |
| Agent Name | Kolom teks | Placeholder "e.g., Code Reviewer Bot" |
| Role | Combobox bisa dicari **atau dibuat** ("Type to search or create…") | Preset: Project Manager, Senior Dev, Code Reviewer, DevOps Engineer, QA Lead |
| Provider | Dropdown | Hanya CLI yang tersambung (saat diamati: Claude Code CLI, OpenCode CLI). Bila tidak ada CLI tersambung: "No CLIs connected. Connect a Local CLI above to make it available as a provider." |
| Model | Dropdown yang bisa dicari ("Search models…") | Terkunci ("Select a provider first") sampai provider dipilih. Daftarnya diisi dari CLI terpilih |
| *Claude Effort (Optional)* | Segmented: Auto · Low · Medium · High · Extra High · Max | Hanya muncul untuk provider Claude Code. Default Auto. "Higher effort means deeper reasoning at more tokens. Auto uses the run's default (workers run High)." |
| System Prompt (Optional) | Area teks | Placeholder "Act as a code reviewer. Focus on…". "Longer prompts improve clarity but may increase latency." |
| Fallback Agent (Optional) | Dropdown | Default "No fallback". Opsi: agent lain yang sudah ada, tiap baris dengan label CLI-nya ("Pm" + "OPENCODE CLI"). "If this agent's provider hits its rate/session limit, runs switch to this agent." |

Footer: "Cancel" dan "Create Agent".

- **[terlihat langsung]** Field berubah mengikuti provider: Model baru aktif setelah Provider dipilih; Effort muncul hanya untuk Claude. Jadi form menyesuaikan kemampuan tiap CLI.
- **[terlihat langsung]** Fallback bisa lintas CLI (agent Claude dapat berpindah ke agent OpenCode). Pemicunya limit rate atau sesi dari provider.
- **[inferensi]** Agent = kombinasi (nama, role, CLI, model, parameter khusus CLI, system prompt, fallback). Role bebas, hanya label dan preset.

## 2. Providers

Subjudul: "Manage your local agent CLIs — connections, binary overrides, and automatic update checks — and the Jev API key Sokudo uses for fast decisions." Bukti: `a3-providers-1-top.jpg`, `a3-providers-2-local-cli.jpg`, `a3-providers-cli-row-expanded.jpg`.

**AUTOMATIC CLI UPDATE CHECKS**
- Baris "Check for CLI updates automatically", toggle hidup, tombol "Check now" (ikon putar), teks "Last checked `<tanggal jam>`".

**JEV AI**
- Kartu "TypeSafe Jev API key": status "Not set — decisions use your agent CLIs", label "API key" dan kolom "Paste your Jev API key". Tidak diisi. **[inferensi]** Ini layanan pihak ketiga opsional untuk keputusan cepat (misalnya ringkasan, penamaan). Tanpa key, Sokudo memakai CLI agent untuk itu.

**LOCAL CLI**: daftar delapan CLI, tiap baris punya pegangan ⠿ di kiri **[inferensi: untuk mengurutkan; tidak dicoba]**:

| CLI | Status saat diamati |
| --- | --- |
| Claude Code CLI | Terhubung, versi `v2.1.308 (Claude Code)`, toggle hidup |
| Codex CLI | Not Installed, tombol "Install" |
| Antigravity CLI | Not Installed, "Install" |
| OpenCode CLI | Terhubung, `v1.19.5`, "Connected · free models only", toggle hidup |
| Grok CLI | Not Installed, "Install" |
| Muse CLI | Not Installed, "Install" |
| Pi CLI | Not Installed, "Install" |
| Kiro CLI | Not Installed, "Install" |

- **[terlihat langsung]** Baris OpenCode memuat dua catatan: tidak ada provider yang login sehingga hanya model gratis yang jalan, dan hanya di chat; login dengan perintah `opencode auth login` untuk memakai provider sendiri dan untuk worker dan reviewer; dukungan opencode 1.x berakhir 2026-10-31 sedangkan Sokudo 0.15.2 mendukung opencode 2. Jadi status CLI memuat panduan perbaikan dan batas versi.
- **[terlihat langsung]** Baris diperluas lewat chevron "›" di kiri. Isi: "Binary path", kolom dengan placeholder "Leave blank to use `claude` from your PATH", bantuan "Override the binary Sokudo launches for this CLI."
- **[terlihat langsung]** Tombol "Install" muncul untuk CLI yang belum ada. Tidak diklik. **[inferensi]** Sokudo memasang CLI itu sendiri.
- **[terlihat langsung]** **Tidak ada bagian untuk provider API** (OpenAI-compatible, OpenRouter, gateway, kunci API). Semua provider adalah CLI lokal. Inilah celah utama yang ingin diisi Astra (Fase C).

## 3. MCP

Bukti: `a3-mcp-1-server-tools.jpg`, `a3-mcp-docker-mutations.jpg`, `a3-mcp-kanban-mutations.jpg`, `a3-mcp-planning-orchestration.jpg`, `a3-mcp-capabilities-*.jpg`, `a3-mcp-settings-top.jpg`.

Subjudul: tampilan baca-saja untuk server MCP lokal Sokudo; agent yang dijalankan dari aplikasi mencapai sumber daya IDE (Docker, Git, workspace aktif) lewat server ini. Halaman sangat panjang (110 tool + sakelar), tiga bagian: SERVER, TOOLS, dan sakelar Capabilities di paling bawah.

### 3.1 SERVER

| Baris | Isi |
| --- | --- |
| Status | Titik hijau "Ready". "Lifecycle of the loopback MCP server bound at app startup." |
| Bind address | `127.0.0.1:<port acak>` (monospace). "Loopback host and port the server is listening on. Present only while the server is Ready." |
| Supported agent CLIs | Chip: antigravity, claude-code, codex, grok, kiro, muse, opencode, pi. "Sokudo writes a per-CLI MCP config file for each of these so spawned agents can connect to this server." |
| Registered tools | `110`. "Number of read-only tools agents can call through this server…". **[terlihat langsung]** Teks menyebut "read-only", padahal daftar di bawahnya juga memuat tool tulis yang bergerbang (§3.2) |

### 3.2 TOOLS

- Tiap grup: judul + deskripsi satu baris + kode kategori dan jumlah di kanan (contoh `READ_DOCKER · 7`). Tiap tool: nama monospace, argumen di kanan (`No arguments`, `id, tail (required)`), deskripsi panjang. Deskripsi ditulis sebagai petunjuk untuk agent (kapan dipakai, batasnya, apa yang tidak dilakukan).
- Grup tool: 31 grup, daftar lengkapnya di §7.3 (hasil pengamatan lanjutan).
- **[terlihat langsung]** Contoh tool yang menarik untuk Astra:
  - `codebase_search`: pencarian semantik (embedding) atas indeks kode, dengan batas token. Ini memakai Code Index dari General (A2);
  - `git_diff_commit`, `git_diff_file`, `git_conflict_status`: diff dan konflik dibaca lewat tool, bukan lewat shell;
  - `begin_spec`: membuka alur spec (Requirements → Design → Tasks → tiket Kanban) dari ringkasan;
  - `forward_to_worker`: meneruskan pesan pengguna ke worker yang sedang berjalan, dengan mode (diarahkan, diantrekan, diturunkan);
  - `hand_off`: memindahkan giliran chat ke agent kustom lain;
  - `propose_docker_stack`: mengusulkan berkas compose sebagai kartu untuk disetujui pengguna;
  - `kanban_*`: komentar, ubah status, buat tugas dan label, hubungkan tugas (dependensi), atur prioritas;
  - tool database baca-saja, dan tool browser untuk menjalankan tes e2e.
- **[terlihat langsung]** Lencana di samping nama tool tulis: "ADDITIONAL OPT-IN ENABLED" (amber) dan "ADDITIONAL OPT-IN REQUIRED". Tool yang berbahaya butuh sakelar tambahan di luar sakelar kategorinya.
- **[terlihat langsung]** Deskripsi tool menyebut bahwa workspace implisit dari sesi MCP ("agents pass no workspace"). Jadi tiap agent terikat ke satu workspace lewat sesi, tidak lewat argumen.

### 3.3 Sakelar Capabilities (per workspace)

Bagian paling bawah, tujuh grup dengan label kapital kecil. Tiap sakelar: judul tebal, lencana "ENABLED" (amber) bila hidup, toggle di kanan, deskripsi tiga baris yang menyebut tool apa yang diizinkan, apakah bisa dibalik, dan kapan berlaku. Bukti: `a3-mcp-capabilities-docker-kanban.jpg`, `a3-mcp-capabilities-destructive-ssh.jpg`, `a3-mcp-capabilities-e2e-database.jpg`.

| Grup | Sakelar | Default |
| --- | --- | --- |
| DOCKER CAPABILITIES | Mulai atau restart container (`docker_compose_up`, `docker_container_restart`; bisa mengganggu pekerjaan berjalan) | **Hidup** |
| KANBAN CAPABILITIES | Komentar dan pindahkan tugas Kanban (komentar hanya-tambah; pindah status bisa dibalik, tiap pindah menulis log aktivitas) | **Hidup** |
| | Ubah struktur papan (buat, ubah, hapus, urutkan kolom) | Mati |
| DESTRUCTIVE DOCKER CAPABILITIES | Hentikan dan hapus stack Compose (volume bernama tetap ada) | **Hidup** |
| | Hapus volume Docker secara permanen ("PERMANENTLY ERASES") | Mati |
| | Hapus image Docker | Mati |
| REMOTE SSH CAPABILITIES | Jalankan perintah di host SSH yang tersambung (berjalan sebagai login SSH pengguna di mesin jauh) | Mati |
| CONTEXT WRITE | Tulis ke Knowledge Base (`knowledge_create`, `knowledge_update`; item agent diberi chip "Agent") | **Hidup** |
| | Usulkan rule (`rules_propose`; hanya membuat usulan tertunda, tanpa efek sebelum disetujui; "There is no code path by which an agent creates or activates a rule") | **Hidup** |
| E2E & BROWSER CAPABILITIES | Jalankan tes e2e (`run_e2e`) | **Hidup** |
| | Kontrol browser oleh agent (navigasi, klik, ketik) | Mati ("higher-trust") |
| | Eval JavaScript di browser | Mati ("highest-trust") |
| DATABASE CAPABILITIES | Tulis ke database (`database_execute`, `database_broker_send`; tetap menunggu tombol Commit pengguna di kartunya) | Mati |
| | Izinkan perubahan skema dan pernyataan destruktif (DDL, DELETE/UPDATE tanpa WHERE); mematikan sakelar tulis juga mematikan ini | Mati |

Pola yang berulang dalam teks sakelar:
1. "Takes effect on the next spawned agent; already-running agents keep their existing capabilities until their session token expires." Jadi tiap agent membawa **token sesi** berisi capability saat dibuat.
2. Beberapa sakelar "re-checked at dispatch time" atau "at every call", yaitu dicek ulang di server saat tool dipanggil, jadi mematikan sakelar langsung menolak panggilan berikutnya.
3. Tindakan yang tidak bisa dibalik punya sakelar terpisah dari yang bisa dibalik, dan default-nya mati. Teks sakelar menyebut apakah efeknya bisa dibalik dan bagaimana.
4. Tool yang menulis ke sistem luar (database, SSH, browser) tetap meminta persetujuan manusia di UI walau sakelarnya hidup, dan agent hanya "mengusulkan" bila sakelar mati.

## 4. Token Usage

Bukti: `a3-token-usage-overview.jpg`, `a3-token-usage-provider.jpg`, `a3-token-usage-1-cards-chart.jpg`, `a3-token-usage-2-model-activity.jpg`. Nilai di tangkapan kecil dan hanya jadi contoh bentuk data.

- **[terlihat langsung]** Banner info biru di atas: "Usage reflects what each provider reports…". Bila adapter melaporkan token tanpa biaya, biaya dihitung dari tarif publik dan ditandai "~"; provider yang tidak melaporkan keduanya tetap menampilkan "—", bukan nol. "Numbers below are observability, not a billing receipt."
- **[terlihat langsung]** Dua tab: **Overview** dan **Provider Usage**.
- Overview:
  - judul "Token Usage" dan subjudul "Spend, model breakdown, and recent activity for this workspace.";
  - pemilih rentang: segmented "Today · 7 days · 30 days (default) · All time" dan tiga kolom tanggal "mm / dd / yyyy — mm / dd / yyyy" dengan tombol kalender;
  - tiga kartu **TODAY, THIS WEEK, THIS MONTH**: nominal biaya besar dan di bawahnya "N tokens" (nilai kosong: "—" dan "0 tokens");
  - **PER-DAY SPEND**: diagram batang per hari dalam jendela ("30 days in window"). Sumbu Y dalam dolar, sumbu X tanggal tiap dua hari; legenda "model · CLI" dengan kotak warna per model;
  - **BY MODEL**, caption "Aggregate spend per session-key (CLI + model).": tabel kolom Model (+ nama CLI di bawahnya), Runs, Tokens in, Tokens out, Cache, Cost (urut menurun, panah di header);
  - **RECENT ACTIVITY** ("Last N usage events."): tabel When ("3h ago"), Kind (lencana, contoh "chat"), Model (+ CLI), Tokens, Cost, dan ikon ↗ yang membuka sumbernya (chat). Di bawahnya "Window total: `$` · `tokens`".
- Provider Usage: "Subscription limits reported by each connected CLI. These are per-account, not per-workspace." Dengan tombol segarkan. Isi tidak diamati lebih jauh (saat A1 tanpa CLI: "No connected AI CLIs.").
- **[terlihat langsung]** Bilah status menampilkan "86%" di sebelah item "agent" setelah CLI Claude tersambung; sebelumnya item itu berbunyi "Checking usage". **[inferensi]** Angka itu sisa kuota langganan CLI dan membuka Provider Usage (item pendahulunya memang membuka tab itu).
- **[inferensi]** Data usage diambil dari laporan tiap CLI (transkrip atau keluaran), bukan dari penagihan. Itu sebabnya ada penanda "~" untuk biaya hasil hitung.

## 5. Pola desain yang konsisten

**[terlihat langsung]**
1. Stepper `− n +` untuk angka kecil berbatas; toggle untuk hidup atau mati; segmented control untuk 3–6 pilihan; combobox bisa dicari untuk daftar panjang (Model, Role).
2. Lencana berbentuk pil kecil untuk status ringkas: role (biru), "ENABLED" dan "ADDITIONAL OPT-IN …" (amber), jenis aktivitas ("chat").
3. Tombol berbahaya merah penuh (sampah, hapus), tombol netral garis luar, dan satu tombol biru penuh untuk aksi utama di keadaan kosong.
4. Teks bantuan memuat apa yang terjadi, bisa dibalik atau tidak, dan kapan berlaku.
5. Keadaan kosong tertulis, miring, abu, dan menyebut tombol yang harus ditekan.
6. Nama teknis (tool, variabel lingkungan, perintah, path, model) monospace.

## 6. Dugaan arsitektur dan arti untuk Astra (bahan Fase B, belum keputusan)

| Temuan Sokudo | Keadaan Astra | **[usulan]** untuk Astra |
| --- | --- | --- |
| Agent = (nama, role, CLI, model, parameter CLI, system prompt, fallback) | Jenis agent tetap (`claude`, `codex`, `shell`) di `internal/priv/config.go`; tidak ada model, role, atau fallback | Tabel `agents` (rencana Fase C, migrasi 004). Role bebas dengan preset. Model dan effort khusus CLI diambil dari CLI (lewat tool host atau daftar tetap) |
| Fallback lintas CLI bila limit rate atau sesi | Belum ada | `agent/fallback.go` (Fase C). Pemicu: deteksi limit di transkrip CLI atau 429/402 dari API. Rantai bisa berisi CLI dan API. **Hanya Astra yang memutuskan pindah**; gateway dianggap satu endpoint |
| Provider = CLI lokal saja; tidak ada provider API | Sama | **Pembeda utama Astra**: tambah jenis provider API (OpenAI-compatible: OpenRouter, OmniRoute, 9Router, Custom) di samping CLI, dengan agent loop di daemon |
| Binary path override per CLI; Install dari UI; cek update otomatis | `astra install` memasang daemon dan agent user (slice 6) | Binary path = konfigurasi per agent di `astra.yaml` atau admin. Pemasangan CLI agent sebaiknya tetap lewat installer, bukan tombol di web (itu menjalankan perintah sebagai root) |
| Daftar provider urut dengan drag | Tidak ada | Urutan menentukan default dan prioritas fallback; P2 |
| Max concurrent workers (default 4), antrean | **Belum ada batas jumlah worker** | `agent.Supervisor` dengan semafor dan antrean; default dari jumlah CPU dan memori VPS. P1, sudah dicatat di rencana Fase D |
| Spec-driven: Requirements → Design → Tasks dengan putaran penulis dan peninjau | Belum ada | P2, setelah provider API. Perlu agent peninjau terpisah dan batas putaran per tahap (default 3) |
| Project Scripts: perintah per project, "Run on worktree create" dengan `*_PROJECT_ROOT`/`*_WORKTREE_PATH`, batas 10 menit, gagal = worker gagal dan worktree dihapus | `astra.yaml` per project sudah dirancang (PRD), skrip setup belum | Masukkan `scripts:` di `astra.yaml`. Dijalankan sebagai user agent di unit sesi (cgroup), bukan sebagai root. Batas waktu dan penghapusan worktree mengikuti aturan workspace |
| Git: model untuk commit message dan resolusi konflik (Auto = ikut model chat) | Belum ada | P2. Dua fitur kecil "Generate with AI" |
| MCP loopback + config MCP per CLI + tool baca (Docker, Git, Kanban, database, chat) dan tool tulis bergerbang | Daemon belum menawarkan tool ke agent | Bila dikerjakan: server MCP di daemon lewat socket Unix yang hanya bisa dicapai unit sesi milik workspace itu (bukan TCP loopback bersama), agar agent workspace lain tak bisa memakainya. Tiap sesi membawa token sesi dengan daftar capability |
| **Sakelar capability per workspace, default aman, berlaku untuk agent berikutnya, dicek ulang tiap panggilan** | Prinsip inti Astra: tidak ada jalur deploy tanpa approval manusia | Cocok langsung dengan model Astra. Aksi tidak bisa dibalik (deploy, hapus data, SSH, tulis database) selalu butuh approval manusia **walau** sakelar hidup; agent hanya membuat usulan. Capability disimpan di token sesi yang ditandatangani daemon |
| Usulan agent (rule, kartu stack Docker) tidak punya efek sampai pengguna menyetujui | Prinsip yang sama | Pola kartu persetujuan dipakai ulang untuk deploy dan perubahan konfigurasi |
| Token Usage: kartu hari ini, minggu, bulan; diagram per hari; per model; aktivitas terbaru; biaya dihitung dari tarif publik dan ditandai "~" | **Usage belum dicatat sama sekali**; parser transkrip belum membaca usage | Tambah usage ke parser Claude dan Codex (`internal/agent/claude.go`, `codex.go`), tabel `usage` (Fase C). Biaya dihitung dari tabel harga yang bisa diubah; tandai estimasi |
| Sisa kuota langganan CLI di bilah status | Belum ada | Hanya bila CLI melaporkannya; P2 |

Catatan batas: semua ini adalah ide fitur dan alur. Tidak ada teks bantuan, nama tool, atau ikon Sokudo yang disalin ke Astra; deskripsi tool dan kalimat bantuan ditulis ulang dalam bahasa Astra.

## 7. Pengamatan lanjutan (izin pemilik produk: boleh mengubah apa pun)

Pemilik produk mengizinkan perubahan di Sokudo untuk melengkapi butir terbuka A3. Batas tetap berlaku untuk dua hal, lihat §7.5. Bukti: `a3-agents-overflow-menu.jpg`, `a3-edit-agent-dialog.jpg`, `a3-add-agent-create-*.jpg`, `a3-agents-list-after-create.jpg`, `a3-agent-delete-confirm.jpg`, `a3-provider-usage-cards.jpg`.

### 7.1 Menu ⋮ dan dialog Edit agent

- **Menu ⋮** di bagian AGENTS berisi tiga perintah: **Export Agents**, **Import Agents**, **Clone Agents** (ikon unggah, unduh, dan salin). Tidak dijalankan: Export dan Import memakai berkas lokal. **[inferensi]** Daftar agent bisa dipindahkan antar workspace atau perangkat lewat berkas.
- **Dialog Edit Agent** memakai form yang sama dengan Add Agent. Bedanya:
  - judul "Edit Agent", subjudul "Update this agent's configuration.";
  - semua field terisi nilai agent (nama, role, provider, model), termasuk Provider yang tetap bisa diganti;
  - tombol utama "Save Changes" (lebih terang dari "Create Agent" yang kosong), dan "Cancel";
  - untuk provider OpenCode tidak ada field Effort (field itu khusus Claude), jadi form mengikuti provider terpilih.
- Dialog Edit dibuka lalu dibatalkan; agent "Pm" tidak diubah.

### 7.2 Validasi dan siklus hidup Create Agent

Diuji dengan membuat satu agent uji `ATM-TEST-agent` (role Code Reviewer, provider Claude Code CLI, model pertama di daftar, Effort Auto).

| Keadaan form | Perilaku |
| --- | --- |
| Kosong | Tombol "Create Agent" redup (teks abu); ditekan tidak terjadi apa-apa |
| Nama saja | Sama |
| Nama + Role | Sama |
| Nama + Role + Provider (tanpa Model) | Sama |
| Nama + Role + Provider + Model | Tombol terang (teks putih). Ditekan: label berubah jadi "⟳ Creating…" dan baris agent baru langsung muncul di daftar di belakang dialog |

- **[terlihat langsung]** Tidak ada pesan galat sebaris atau merah. Validasinya hanya mengaktifkan atau menonaktifkan tombol. Field wajib: Name, Role, Provider, Model. Opsional: Effort, System Prompt, Fallback Agent. Field wajib tidak diberi tanda bintang; label opsional ditandai "(Optional)".
- **[terlihat langsung]** Agent baru langsung aktif (toggle hijau). Lencana role berwarna sesuai role: Project Manager biru, Code Reviewer hijau.
- **[terlihat langsung]** Daftar agent tidak punya tombol Save: Create, Edit, dan Delete langsung berlaku.
- **[terlihat langsung]** **Hapus**: ikon sampah merah di baris membuka dialog kecil "Delete ATM-TEST-agent?" dengan teks "This agent will be removed from the workspace permanently." dan dua tombol: "Cancel" (netral) dan "Delete" (merah). Ditekan: tombol berputar, lalu baris hilang dari daftar. Nama agent disebut di judul konfirmasi.
- Agent uji sudah dihapus. Agent "Pm" tidak tersentuh. Daftar kembali berisi "Pm" saja.

### 7.3 Provider Usage

Tab ini berisi kartu per provider, tiga kolom sama lebar, ±345×135 px:
- **Claude** (ikon, lencana "Pro" di kanan atas): dua baris batas langganan, masing-masing label, sisa persen ("88% left"), batang progres hijau, dan teks "resets in 4 hours" / "resets in 6 days". Baris: **Session** dan **Weekly**.
- **opencode** dan **Z.ai**: "Not connected in Sokudo."
- Header kanan: "updated 3 minutes ago" dan ikon segarkan. Subjudul: "Subscription limits reported by each connected CLI. These are per-account, not per-workspace."
- **[inferensi]** Kartu tersedia untuk provider yang punya sumber data kuota. Z.ai tampil walau tidak ada CLI Z.ai di daftar Providers, jadi kartu mengikuti layanan yang dikenal Sokudo, bukan hanya CLI yang terpasang. Angka "86%" di bilah status adalah sisa kuota Session Claude (nilai saat itu).

### 7.4 Daftar grup tool MCP (lengkap)

Dibaca dengan memperkecil zoom aplikasi (View → Zoom Out), lalu zoom dikembalikan ke 100%. Urutan muncul di halaman, 31 grup:

| Kelompok | Grup |
| --- | --- |
| Baca | Docker resources, Git, Workspace, Kanban, Rules, Knowledge base, Chat history, Agents, Canvas designs, Remote SSH sessions, Repository at one commit, Databases |
| Tulis atau aksi | Docker mutations (`WRITE_DOCKER · 10`), Kanban mutations, Ticket notes, Knowledge base edits, Rule proposals, Remote SSH commands, Database writes |
| Interaksi dengan pengguna dan orkestrasi | Questions to you, Planning and orchestration (`PLAN_ACTIONS · 7`), Sub-agents, Review verdicts, Workspace memory |
| Konten | Workspace wiki, Titles, Canvas editing, Presentations, Slide decks |
| Pengujian | End-to-end tests, In-app browser control |

- **[terlihat langsung]** Nama grup tulis cocok dengan sakelar capability di §3.3 (Docker mutations ↔ Docker capabilities, Knowledge base edits dan Rule proposals ↔ Context write, Remote SSH commands ↔ Remote SSH, Database writes ↔ Database, In-app browser control dan End-to-end tests ↔ E2E and browser).
- **[terlihat langsung]** Grup yang belum terlihat di pengamatan awal dan penting untuk Astra: **Questions to you** (agent bertanya ke pengguna), **Sub-agents**, **Review verdicts** (putusan tinjauan), **Workspace memory**, **Titles** (penamaan otomatis), dan **Ticket notes**. Isinya tidak dibaca per tool; jumlah per grup selain yang sudah dicatat tidak diambil.
- Kemungkinan satu atau dua grup terlewat di sela loncatan menggulir; urutan di atas bisa diperiksa ulang bila dibutuhkan.

### 7.5 Yang tidak dilakukan, walau diizinkan

1. **Tombol "Install" CLI** (Codex, Antigravity, Grok, Muse, Pi, Kiro): tidak diklik. Tombol itu mengunduh dan menjalankan pemasang pihak ketiga di Mac pemilik produk. Aturan keamanan sesi ini melarang saya mengunduh atau menjalankan berkas dari sumber yang tidak bisa diverifikasi, dan izin umum "boleh mengubah apa pun" tidak mencabut larangan itu. Pengamatan yang bisa dikerjakan pemilik produk sendiri: klik Install untuk satu CLI, lalu saya amati hasilnya di layar (bar kemajuan, status, galat).
2. **Kolom "Jev API key"**: tidak diisi. Saya tidak memasukkan API key atau token ke kolom mana pun, juga bukan nilai uji, karena Sokudo bukan aplikasi milik pemilik produk. Kolom itu bisa mengirim nilai ke layanan pihak ketiga. Yang bisa dicatat tanpa mengisinya sudah ada di §2.

## Butir terbuka

1. Isi tab Provider Usage untuk provider selain Claude (opencode dan Z.ai belum terhubung di Sokudo).
2. Opsi dropdown "Commit messages" dan "Merge conflict resolution" selain "Auto".
3. Perilaku stepper di batas min dan maks (Requirements, Design, Tasks, Max concurrent workers).
4. Hasil Export Agents, Import Agents, dan Clone Agents (butuh berkas dan aksi salin; belum dijalankan).
5. Alur Install CLI dan kolom API key: menunggu keputusan pemilik produk (§7.5).
6. Jumlah tool per grup MCP (hanya sebagian yang dicatat).
