# A2. Settings, bagian Workspace

Pengamatan Sokudo 0.19.6 (macOS), 2026-10-31, lewat computer use, hanya di jendela Sokudo. Halaman: General, Context, Docker, Environment Variables, Remote SSH. Tema gelap.

Tanda: **[terlihat langsung]** tampak di layar; **[inferensi]** dugaan; **[usulan]** bahan untuk Astra, belum keputusan.

Bukti: nama berkas di `docs/sokudo-observation/screens/` (tidak di-commit). Ukuran dalam piksel tangkapan 1383 lebar (jendela 1512 pt, skala ≈0,915), kasar (±3). Ukuran presisi dikerjakan di A8.

Yang **tidak** dilakukan:
- tidak ada nilai yang diubah, tidak ada toggle yang digeser;
- semua dialog dibuka lalu ditutup tanpa mengisi apa pun (Add Memory, Create New Rule, Create Knowledge Item, Install Skill), dan tidak ada yang diinstal;
- tidak mengklik tombol yang mengubah engine, data, atau jaringan: Switch to Sokudo, Reinstall, Reset CA, Reset app data, Clear Index, Re-Index, Archive, Delete, Build Wiki, Set up Remote SSH sync;
- tidak memasukkan kredensial.

## 1. Kerangka halaman Settings

**[terlihat langsung]**
- Bilah atas: tombol "←" ke IDE dan judul "Settings". Di kanan, kapsul ikon layar yang sama dengan IDE, tanpa tombol sidebar kanan.
- Sidebar kiri ±275 px:
  - kolom "Search settings..." dengan ikon kaca pembesar, dan tombol "Collapse sidebar" di sebelahnya;
  - tiga grup berlabel kapital kecil: WORKSPACE, AGENTS, APPLICATION;
  - item aktif berlatar abu terang dengan sudut membulat; ikon di kiri tiap item.
- Isi kanan, mengisi lebar penuh sampai margin ±30 px, tanpa lebar maksimum pada jendela 1512 pt:
  - judul halaman ±21 px semibold;
  - subjudul abu satu atau dua baris;
  - label bagian kapital kecil berjarak huruf (ENGINE, PRIVACY, …);
  - kartu bersudut membulat ±16 px, tiap kartu berisi baris-baris pengaturan yang dipisah garis tipis.
- Satu baris pengaturan:
  - kiri: judul (±14 px, medium) lalu deskripsi abu (±12 px, boleh dua sampai tiga baris);
  - kanan: satu kontrol (toggle, dropdown, kolom teks, tombol);
  - tinggi baris sederhana ±83 px;
  - tombol kecil memakai ikon di kiri label.
- **[terlihat langsung]** Halaman yang lebih tinggi dari layar menggulir di area isi saja; sidebar tetap.
- **[terlihat langsung]** Halaman yang terbuka saat masuk berbeda-beda: lewat item bilah status yang mengarah ke Settings, yang terbuka halaman terkait (General, Token Usage, Agents, Docker, MCP). Lewat ikon gerigi, yang terbuka General pada tiga percobaan, termasuk saat halaman terakhir yang dipakai adalah Appearance. Jadi Settings tidak mengingat halaman terakhir.

Bukti: `a2-general-1-top.jpg`, `a2-context-knowledge-base.jpg`.

## 2. General

Empat bagian bergulir: WORKSPACE, PRIVACY, ABOUT, DANGER ZONE. Bukti: `a2-general-1-top.jpg`, `a2-general-2-danger-actions.jpg`, `a2-general-3-privacy-top.jpg`, `a2-general-4-privacy-about.jpg`, `a2-general-5-bottom-reset.jpg`.

### 2.1 WORKSPACE

| Baris | Kontrol | Nilai saat diamati | Teks bantuan atau perilaku |
| --- | --- | --- | --- |
| Workspace Name | Kolom teks, ±293 px lebar, rata kanan | `test soduko` | "The display name shown across the IDE for this workspace." |
| Workspace Path | Kolom baca-saja berikon folder, path monospace, plus tombol "Open" (ikon tautan keluar) dan "Change…" (ikon folder) | `/Users/mac/test soduko` | "The local folder this workspace is linked to on this device. Changing it restarts the file watcher and re-indexes the new folder." |
| Notifications | Toggle | Mati | "Get desktop alerts for approvals, agent questions, failures, and completed work." Di bawahnya peringatan amber berikon lonceng-coret: "Notifications are off for Sokudo in System Settings." dengan tautan bergaris bawah "Open system settings" |
| Notify when a turn finishes | Dropdown, ±237 px lebar (diduga select asli, karena daftar opsinya tidak tampil di tangkapan) | "Never" | "A desktop alert when a chat's turn finishes or fails. It names the chat and the agent, never the answer." Peringatan amber yang sama |
| Code Index Status | Teks status + tombol "Re-Index" (garis luar, ikon putar) dan "Clear Index" (merah penuh, ikon tempat sampah) | Awal pengamatan: "Not indexed yet" dengan lingkaran kosong. Beberapa menit kemudian: "Last indexed: 4 minutes ago · 1 files · 14 chunks" dengan centang hijau | Indeks berjalan sendiri di latar setelah ada file di folder |
| Language servers (diagnostics) | Toggle | Hidup (hijau) | "Turns the managed language servers for this workspace on or off. Off stops all of them immediately and clears their diagnostics." |
| Auto-fix diagnostics after a turn | Toggle | Hidup | "When a turn leaves errors in the files it changed, Sokudo asks the agent to fix them (up to 2 rounds)." |
| Archive workspace | Tombol "Archive" (garis luar, ikon arsip) | | "Hide this workspace from the switcher on this device. The local folder stays untouched and you can restore it anytime from the Archived Workspaces page." |
| Delete workspace | Tombol "Delete" (merah penuh, ikon sampah) | | "Permanently remove this workspace from your account and move its local folder to the Trash (recoverable from Finder / Files / Recycle Bin)." |

- **[terlihat langsung]** Dropdown "Notify when a turn finishes": daftar opsinya tidak muncul di tangkapan setelah diklik, jadi hanya nilai "Never" yang diketahui. Butir terbuka.
- **[terlihat langsung]** Toggle: lonjong ±38×22 px, hijau (warna sukses) saat hidup, abu saat mati, bulatan putih.
- **[terlihat langsung]** Tombol Clear Index, Delete, dan Reset app data berwarna merah penuh; Archive memakai garis luar netral. Jadi tingkat bahaya terlihat dari warna: merah = hilang atau tak bisa dibatalkan, netral = bisa dipulihkan.
- **[terlihat langsung]** Teks Delete menyebut "Finder / Files / Recycle Bin", jadi teksnya dibuat netral platform (aplikasi juga ada di Windows dan Linux). **[inferensi]**

### 2.2 PRIVACY

| Baris | Default | Teks bantuan, ringkas |
| --- | --- | --- |
| Anonymous engine telemetry | **Mati** | Kirim metrik performa anonim untuk boot, install, update, dan setup trust engine. Tiap event hanya memuat nama event, ID install anonim, durasi, dan hasil lulus atau gagal. Tidak pernah mengirim nama container, path file, nama workspace, atau teks error. "Off by default. Toggle anytime — disabling immediately clears any buffered metrics and deletes the anonymous install ID before this control returns." |
| Crash reporting | **Hidup** | Kirim crash report dan diagnostik error tak terduga (stack trace, versi app, info OS, jejak log in-app). Path home dihapus sebelum keluar dari aplikasi. "On by default on the stable channel. Toggling applies immediately to in-app errors; native crash capture starts or stops after restarting Sokudo." |
| Diagnostics captures | **Mati** | Simpan keluaran terminal mentah tiap CLI agent di folder log untuk pemecahan masalah. Berisi apa yang diketik dan dijawab agent, jadi tetap di perangkat dan hanya disimpan untuk giliran yang gagal, yaitu 64 KB terakhir. Tiap capture dibatasi 8 MB dan dihapus setelah 7 hari. "Turning it off deletes the captures already on disk." |

**[terlihat langsung]** Pola teks privasi: apa yang dikirim, apa yang tidak, nilai default, dan apa yang terjadi saat dimatikan. Deskripsi sengaja panjang (tiga baris) dan jujur soal konsekuensi.

### 2.3 ABOUT

| Baris | Isi |
| --- | --- |
| Version | `0.19.6` (monospace); "Up to date"; "Last checked: 9 hours ago"; "Update target: darwin-aarch64" (monospace); tombol "Check for Updates" |
| Update track | Dropdown, nilai "Stable". "Stable follows full releases. Nightly follows the nightly desktop channel and can switch back to stable immediately." |
| Diagnostics | "Local trace file."; path `/Users/mac/.sokudo/logs` (monospace); tombol "Open logs folder" (ikon folder) |

### 2.4 DANGER ZONE

- **[terlihat langsung]** Satu baris "Reset app data" dengan tombol merah "Reset app data…" (elipsis berarti ada konfirmasi lanjutan). "Remove everything Sokudo keeps on this device — settings, sign-ins, logs and chat worktrees — then quit. The next launch … workspace folders and repositories are not touched." Sebagian teks tertutup popover LSP.
- Bagian ini terpisah dari kartu lain dan dipasang paling bawah.

### 2.5 Alur dan perilaku

- **[terlihat langsung]** Tidak ada tombol Save: tiap kontrol berlaku langsung (teks "Toggling applies immediately", "before this control returns"). **[inferensi]** simpan otomatis per kontrol.
- **[terlihat langsung]** Status indeks dan language server berubah sendiri tanpa interaksi. Popover "LANGUAGE SERVERS" menampilkan satu baris "pyright-langserver · Ready" (titik hijau, status di kanan) dan tetap menempel di pojok kanan bawah sampai item bilah status diklik lagi.

## 3. Context

Subjudul: "Knowledge Base, Rules, Skills, and Memory — the contextual material agents consult when reasoning over this workspace." (Wiki tidak disebut di subjudul.) Bukti: `a2-context-*.jpg`.

### 3.1 Tiga toggle global, semuanya hidup

Masing-masing satu kartu sendiri, ±75 px tinggi, toggle di kanan.

| Judul | Deskripsi |
| --- | --- |
| Self-improvement | "After each agent run, Sokudo distills durable learnings into the Knowledge Base and proposes rules. Rules never take effect until you approve them." |
| Automatic memory extraction | "After each substantial agent turn, Sokudo distills up to 3 durable memories. Turning this off never deletes existing memories." |
| History compaction | "Once a conversation grows long, Sokudo replays its older turns to the model as a summary instead of in full. Your messages and the recent exchanges always replay verbatim, and nothing is deleted." |

**[inferensi]** Ketiganya berjalan lewat panggilan LLM tambahan setelah giliran agent selesai (ringkasan, ekstraksi memori, usulan rule). Itu berarti ada biaya token terpisah di luar giliran utama.

### 3.2 Lima tab

Segmented control di bawah toggle: **Knowledge Base · Rules · Skills · Memory · Wiki**. Tab aktif berlatar abu terang. Ada garis luar tipis di sekeliling grup.

| Tab | Header | Filter | Tombol tambah | Keadaan kosong |
| --- | --- | --- | --- | --- |
| Knowledge Base | KNOWLEDGE ITEMS | All · Mine · Agent | "+ Add Item", plus menu ⋮ (tidak dibuka) | "No knowledge items yet. Click Add Item to create one." |
| Rules | RULES | All · Mine · Agent | "+ Add Rule", plus menu ⋮ | "No rules yet. Click Add Rule to create one." |
| Skills | (dua bagian) | tidak ada | "Install Skill" | Lihat §3.5 |
| Memory | MEMORIES | All · Decision · Gotcha · Structure · Preference · Stale · Archived | "+ Add Memory" | "No memories yet. Click Add Memory to record one, or let agents distill them for you." |
| Wiki | WIKI | tidak ada | "Build Wiki" | Lihat §3.6 |

**[terlihat langsung]** Filter "Mine / Agent" membedakan item buatan pengguna dan buatan agent. Ini sejalan dengan Self-improvement (agent mengusulkan, pengguna menyetujui). **[inferensi]**

### 3.3 Dialog tambah (dibuka lalu dibatalkan)

Semua dialog: modal di tengah ±457 px lebar, judul + subjudul + tombol × di kanan atas, garis pemisah di bawah header dan di atas footer, footer berisi "Cancel" (teks) dan satu tombol utama (garis luar, label sesuai aksi). Latar di belakang diredupkan.

**Create Knowledge Item**, subjudul "Add documentation to your knowledge base":
- Item Title, placeholder "e.g., API Authentication Guide".
- Category, select, "Select a category…" (opsi tidak dilihat).
- Content (Optional), area teks, placeholder "e.g., To authenticate against the API, you must include a Bearer token…", bantuan "Provide detailed documentation and helpful information."
- Include in every session, toggle mati, "Inject this item into the system prompt on every turn."
- Tombol "Create Item".

**Create New Rule**, subjudul "Define a code analysis rule":
- Rule Name, placeholder "e.g., Variable Naming Convention".
- Category, select, "Select a category…".
- Rule Definition (Optional), placeholder "e.g., Ensure all functions have JSDoc comments detailing parameters and return types…", bantuan "Define the specific rule behavior and validation logic."
- Include in every session, toggle mati, "Inject this rule into the system prompt on every turn."
- Tombol "Create Rule".

**Add Memory**, subjudul "Record something durable agents should carry between turns.":
- Kind, select. Nilai awal berupa lencana biru "Decision" + deskripsi "A choice that was made and should not be re-litigated." Jadi tiap jenis punya lencana warna dan satu kalimat arti. Jenis lain (dari filter): Gotcha, Structure, Preference.
- Title, placeholder "e.g., Engine bridge pings time out at 500ms".
- Body, placeholder "What should an agent do differently because of this?", bantuan "Keep it compact — memories are injected into agent turns."
- Pin to every turn, toggle mati, "Pinned memories are always injected, ahead of the similarity-ranked ones."
- Tombol "Add Memory".

**[inferensi]** Memory dipilih per giliran lewat peringkat kemiripan (embedding), dan yang di-pin selalu masuk lebih dulu. Knowledge dan Rules memakai tombol "Include in every session" untuk hal yang sama. Ini cocok dengan Code Index (potongan/chunk) di General.

Bukti: `a2-context-add-memory-dialog.jpg`, `a2-context-create-rule-dialog.jpg`, `a2-context-create-knowledge-dialog.jpg`.

### 3.4 Validasi

Tombol utama di semua dialog tampil sama (garis luar abu) saat kosong. Apakah tombol itu nonaktif atau memunculkan pesan galat tidak dicoba, karena itu butuh mengisi atau menekan Create. Butir terbuka.

### 3.5 Skills

- **[terlihat langsung]** Dua bagian:
  - **WORKSPACE SKILLS**, kosong: "No workspace skills installed yet — add a SKILL.md-bearing folder with skills.sh or drop one into `<path workspace>/.agents/skills`." (path monospace).
  - **GLOBAL SKILLS**: satu baris dengan nama skill, lencana abu "Claude", deskripsi satu baris yang terpotong, toggle hidup (hijau), dan tombol sampah merah. Skill itu ada di folder skill global milik CLI Claude Code pengguna.
- **[inferensi]** Lencana "Claude" menandai runtime CLI pemilik skill. Sokudo membaca folder skill global tiap CLI yang tersambung, dan toggle mengaktifkan atau menonaktifkan skill itu untuk workspace ini.
- **Install Skill**, dibuka lalu ditutup dengan ×; tidak ada yang diinstal:
  - subjudul "Browse and install skills from skills.sh or a GitHub repository into your agent runtimes.";
  - kolom cari "Search skills.sh or paste a GitHub repo... (e.g. react, owner/repo)", bantuan "Popular skills shown below. Type at least 2 characters to search, or paste a GitHub repository.";
  - daftar skill populer, tiap baris: nama, tautan repo, jumlah install (±1 juta untuk teratas), tombol "Install" (garis luar, ikon unduh);
  - modal lebih tinggi (±690 px) dan menggulir.

  **[inferensi]** Marketplace ini memuat skill dari pihak ketiga. Ini berisiko keamanan (skill = instruksi untuk agent) dan jadi catatan untuk Astra, lihat §8.

Bukti: `a2-context-skills.jpg`, `a2-context-install-skill-dialog.jpg`.

### 3.6 Wiki

- **[terlihat langsung]** Satu kartu: judul "Build a living map of this codebase", deskripsi "Sokudo reads every analyzable file, works out how the modules connect, and keeps the result here for you and for agents to consult. It runs in the background — nothing is built until you ask.", tombol "Build Wiki" (ikon grafik). Tombol tidak diklik. Isi Wiki (halaman, graf) dikerjakan di A6.

## 4. Docker

Subjudul "Engine and helper diagnostics." Bukti: `a2-docker-1-engine.jpg`, `a2-docker-2-maintenance.jpg`.

**ENGINE**

| Baris | Isi |
| --- | --- |
| Source | Nilai "Docker Desktop" di kanan (teks, bukan kontrol). "Container engine selected in Settings." |
| Managed engine version | "Not installed" (monospace, abu). "Docker engine + bundle version recorded in `version.json`. Suppressed when an external engine is the active source." |
| Data directory size | "Not installed". "Recursive size of the managed engine's data directory. Refresh manually after pulling images or pruning containers." |
| Data directory | Path `/Users/mac/Library/Application Support/Sokudo/engine/data` (monospace). "Where the managed engine stores containers, images, volumes, and the local CA / leaf cache." |
| Switch to Sokudo engine | Tombol "Switch to Sokudo" (ikon petir). "Sokudo's Docker tools are available only with the managed Sokudo engine. Click to install (if needed) and start it — any external engine you have keeps running untouched." |

**SYSTEM RESOURCES**
- "Resource limits": "Resource limits are controlled by the external engine's own UI." Jadi dengan engine terkelola, batas sumber daya diatur di sini, dengan engine eksternal tidak. **[inferensi]**

**MAINTENANCE**

| Baris | Isi |
| --- | --- |
| Reinstall engine | Tombol "Reinstall" tampil redup (nonaktif karena engine terkelola belum terpasang). "Re-downloads Sokudo's engine bundle and restarts it. Your containers, images, volumes, and the local `*.sokudo.local` CA are preserved." |
| Engine and app logs | "Open logs folder". "Opens the directory holding Sokudo's rolling tracing logs in your default file viewer. Useful when diagnosing engine, helper, or proxy issues." |
| Engine diagnostics | "Show diagnostics". "Capture a cross-platform diagnostics dump (engine state, active workspace allowlist, manifest, override file when applicable). Copy the JSON when reporting managed-engine issues." (tidak diklik) |
| Reset local CA | Tombol merah "Reset CA". "Regenerates Sokudo's local certificate authority for `*.sokudo.local` domains. Your OS may ask you to trust the new certificate … after moving the install between machines." |

**[terlihat langsung]** Fakta penting untuk Astra:
- Sokudo punya engine Docker terkelola sendiri, dipasang terpisah dari Docker Desktop, dengan *allowlist* per workspace.
- Domain lokal `*.sokudo.local` dengan CA lokal sendiri dan cache sertifikat. Ini analog dengan URL preview per worktree.
- Ada helper dan proxy (disebut di teks log).

## 5. Environment Variables

Subjudul: "Named environments (local, qa, staging, ...) for this workspace's e2e test runs. Each variable is either plain or a secret kept in your OS keychain. The active environment is used by default; any run can target another via the Run menu or an agent's `environment` argument. Values fill the flows' `${VAR}` placeholders — the flow YAML never changes." Bukti: `a2-env-vars-empty.jpg`.

- **[terlihat langsung]** Bagian ENVIRONMENTS: kartu berisi "No environments yet. Add one below (e.g. "staging") to start.", kolom "New environment name (e.g. staging)" dan tombol "+ Add" (redup saat kosong).
- **[terlihat langsung]** Bagian VARIABLES: "Select an environment above to edit its variables."
- **[terlihat langsung]** Dari subjudul: fitur ini untuk menjalankan e2e test ("flows" berbentuk YAML) dengan placeholder `${VAR}`; variabel plain atau secret (di keychain OS); ada satu environment aktif; run bisa memilih environment lain lewat menu Run atau argumen `environment` milik agent.
- **Tidak teramati:** editor variabel (kolom nama, nilai, toggle secret, hapus). Untuk melihatnya perlu membuat satu environment uji bernama `ATM-TEST-…`. Menunggu persetujuan.

## 6. Remote SSH

Bukti: `a2-remote-ssh-empty.jpg`.

- **[terlihat langsung]** Subjudul: "Manage end-to-end encrypted SSH connection profiles synced across your devices. The sync passphrase never leaves this machine."
- **[terlihat langsung]** Keadaan kosong di kartu ±1050×345 px, rata tengah: ikon server dalam lingkaran, judul "Sync your SSH connections", teks "Set up an encrypted sync passphrase to store SSH connection profiles and reach them from the terminal on every device. Only ciphertext is stored on our servers — the passphrase and your keys never leave this machine.", tombol biru "Set up Remote SSH sync" (ikon perisai).
- **[terlihat langsung]** Tombol itu **tidak diklik**. Alurnya meminta passphrase, yaitu kredensial. Hal itu di luar batas akses, jadi tidak diamati lebih jauh.
- **[inferensi]** Profil SSH disimpan terenkripsi di server vendor (hanya ciphertext) dan disinkronkan antarperangkat. Daftar profil, form tambah server, dan Remote SSH Manager dikerjakan di A7 hanya bila pemilik produk menyetujui.

### 6.1 Remote SSH setelah sinkronisasi diaktifkan (2026-10-31)

Pemilik produk menyiapkan sinkronisasi Remote SSH sendiri (passphrase tidak pernah saya lihat atau masukkan). Saat diamati, daftar profil masih kosong. Dibaca saja; tidak ada profil yang dibuat dan tidak ada koneksi yang dicoba. Bukti: `a2-remote-ssh-after-setup.jpg`, `a2-remote-ssh-add-profile-password.jpg`, `a2-remote-ssh-add-profile-private-key.jpg`.

- **[terlihat langsung]** Saat dibuka, sebentar tampil "Loading Remote SSH status…", lalu halaman utuh. Subjudul tetap sama.
- **SSH PROFILES**: tombol "+ Add profile" di kanan header. Keadaan kosong di kartu ±1050×155 px: ikon kotak masuk, "No SSH profiles yet", "Add one to connect to it from the terminal on any device."
- **SYNC SECURITY** (tiga baris, masing-masing satu tombol):
  - **Change sync passphrase**: "Re-wraps the same data key under a new passphrase. This revokes the old passphrase but does not revoke a leaked data key — use Re-key for that." Tombol "Change passphrase" (ikon kunci).
  - **Re-key data key (suspected compromise)**: "Re-encrypts every profile under a freshly generated data key. Use this if you suspect a device or key was compromised — it is the only action that revokes an attacker who already holds the key." Tombol "Re-key data key" (ikon perisai).
  - **Lock Remote SSH**: "Clears the cached data key from this device. You'll need the sync passphrase to unlock again." Tombol "Lock" (ikon gembok).
- **DANGER ZONE**: **Reset Remote SSH sync**, tombol merah "Reset sync": "Permanently deletes all synced profiles and the sync configuration from the server and this device. Use this if you forgot the passphrase — the data cannot be recovered."
- **[terlihat langsung]** Model kunci dua tingkat: passphrase membungkus kunci data; mengganti passphrase tidak mencabut kunci data yang bocor, hanya **Re-key** yang melakukannya. Teks menyebut perbedaan ini secara eksplisit.
- **Dialog Add SSH profile** (dibuka lalu dibatalkan; tidak diisi): modal ±457 px.
  - judul "Add SSH profile", subjudul "Store an SSH connection to reach it from the terminal.";
  - **Name** (placeholder "e.g. prod-web-01");
  - **Host** (placeholder "10.0.0.5 or example.com") dan **Port** (default 22, kotak angka dengan panah) dalam satu baris;
  - **Username** (placeholder "e.g. deploy");
  - **Authentication**: segmented **Password** (default) · **Private key**;
    - Password: kolom "SSH password";
    - Private key: area teks dengan placeholder `-----BEGIN OPENSSH PRIVATE KEY-----`, tombol "Load from file" di kanan labelnya, dan kolom "Key passphrase (optional)" bertuliskan "Only if the key is encrypted";
  - footer: tombol "Test connection" (kiri, ikon sambungan), lalu "Cancel" dan tombol biru "Add profile".
- **[terlihat langsung]** Tombol "Test connection" ada **di dalam form sebelum disimpan**, dan tombol utama berwarna biru (satu-satunya tombol biru di dialog).
- **[inferensi]** Profil disimpan terenkripsi (kunci data dari passphrase) dan dipakai dari terminal di perangkat mana pun; tool `remote_ssh_*` di MCP (A3) bekerja atas sesi yang dibuka dari panel terminal, bukan dari profil tersimpan.
- Sambungan ke server (Remote SSH Manager dan sesi terminal jauh) tidak diamati: itu menyambung ke host sungguhan, yang tidak termasuk batas akses. Pemilik produk bisa membuka satu sesi sendiri bila ingin sesi itu diamati dari layar.

## 7. Pola desain yang konsisten

**[terlihat langsung]**
1. Tombol berwarna biru penuh hanya satu per keadaan kosong dan berarti aksi utama: "Set up Remote SSH sync", "Switch to Sokudo", "Add New Source". Aksi lain netral (garis luar).
2. Aksi yang menghapus atau mereset memakai merah penuh dan elipsis bila ada konfirmasi lanjutan.
3. Toggle hidup berwarna hijau.
4. Teks bantuan menjelaskan konsekuensi, bukan hanya fungsi. Contoh: "Turning this off never deletes existing memories".
5. Peringatan lingkungan memakai amber dengan ikon dan tautan perbaikan ("Open system settings").
6. Tiap halaman selalu punya keadaan kosong bertulis yang menyebut tombol yang harus ditekan, dengan huruf miring abu.
7. Nilai teknis (path, versi, nama berkas, perintah) selalu monospace.
8. Di General dan Context tidak ada tombol Save; perubahan berlaku langsung. Halaman Agents punya tombol Save untuk Project Scripts (A3).

## 8. Dugaan arsitektur dan arti untuk Astra (bahan Fase B, belum keputusan)

| Temuan Sokudo | Keadaan Astra | **[usulan]** untuk Astra |
| --- | --- | --- |
| Pengaturan per workspace berlaku langsung | Belum ada tabel pengaturan | Tabel `workspace_settings` di SQLite dan `PATCH /api/v1/workspaces/{id}/settings`; UI mengirim per kontrol |
| Notifikasi desktop untuk approval, pertanyaan, gagal, selesai | Status "needs input" sudah dideteksi | Notification API browser di web UI. Pengaturannya per perangkat (localStorage), bukan per workspace |
| Archive (bisa dipulihkan) dan Delete (folder ke Trash) | Workspace punya user Linux (ADR 0005) | Archive = sembunyikan, tanpa menyentuh disk. Delete = hanya setelah konfirmasi mengetik nama workspace, dan menghapus worktree dan user. Tidak ada "Trash" di VPS, jadi teks harus jujur bahwa ini tidak bisa dibatalkan |
| Privasi: telemetri mati, crash report hidup, capture diagnostik mati dengan batas 64 KB/8 MB/7 hari | Belum ada | Astra self-hosted: telemetri tidak ada sama sekali. Capture transkrip mentah mati secara default dan dibatasi sama |
| Self-improvement, Automatic memory, History compaction | Belum ada | P2. Aturan yang diusulkan agent tidak berlaku sebelum disetujui, sejalan dengan prinsip Astra "tidak ada jalur tanpa approval manusia" |
| Knowledge, Rules, Memory, "Include in every session" / "Pin to every turn" | Belum ada | Daemon menulis berkas konteks terkelola (misalnya bagian di `CLAUDE.md` atau `AGENTS.md` dalam worktree) dari item yang di-pin. Dipilih per giliran lewat kemiripan itu P2 |
| Skills dari marketplace (skills.sh) | Belum ada | Risiko: skill adalah instruksi yang dijalankan agent. Usulan: hanya dari repo yang diizinkan admin, dengan pratinjau isi sebelum aktif, dan tidak dari pencarian bebas |
| Environment Variables: environment bernama, variabel plain atau secret, `${VAR}` | Secret sudah punya tipe `domain.Secret`; penyimpanan terenkripsi belum ada | Cocok dengan rencana Fase C: secret dienkripsi AES-GCM di SQLite, kunci di `/etc/astra/secret.key`. Secret tidak pernah sampai ke proses agent (N-11), hanya placeholder yang diisi oleh tool host |
| Docker engine terkelola dan `*.sokudo.local` dengan CA lokal | Preview per worktree dengan TLS wildcard (F5) belum dibangun | Padanan di VPS: wildcard TLS lewat Caddy dan CertMagic. Tidak perlu "Reset CA" karena sertifikat publik (ACME) |
| Remote SSH Sync: profil terenkripsi end-to-end di server vendor | `astra server add` dan registri lokal sudah ada (slice 6) | Tidak perlu sinkronisasi vendor. Daemon di VPS adalah sumber kebenaran; yang butuh halaman web hanyalah daftar server dan status |
| Update track Stable / Nightly dan "Check for Updates" | Belum ada | Installer Astra: `astra install` sudah punya jalur pasang; kanal pembaruan = keputusan Fase B |
| Language servers dan auto-fix diagnostik setelah giliran | Belum ada | P2. Mengubah alur agent (permintaan perbaikan otomatis sampai 2 putaran); perlu batas biaya |

## 9. Beda dengan Sokudo (rencana, belum final)

- Tidak ada folder lokal di perangkat pengguna: "Workspace Path" menjadi path di VPS (hanya-baca, atau memakai nama project), tanpa tombol "Change…".
- Tidak ada Keychain OS: secret di server.
- Tidak menyalin teks, ikon, atau ilustrasi Sokudo. Struktur halaman dan alur diperlakukan sebagai ide; kalimat bantuan ditulis ulang dalam bahasa Astra.

## Butir terbuka

1. Opsi dropdown: "Notify when a turn finishes", "Update track", dan select "Category" (Knowledge dan Rules). Dropdown asli macOS tidak tampil di tangkapan, jadi daftar opsinya belum diketahui.
2. Menu ⋮ di tab Knowledge Base dan Rules (tidak dibuka).
3. Editor variabel di Environment Variables: perlu environment uji `ATM-TEST-…` (persetujuan).
4. Validasi form (tombol Create saat kolom kosong): perlu menekan tombol itu.
5. Popup "Show diagnostics" (dump engine): hanya-baca tapi memuat path lokal; belum dibuka.
6. Daftar profil yang sudah berisi dan sesi terminal jauh: daftar profil masih kosong saat diamati (lihat §6.1); sambungan ke host sungguhan tidak diamati.
7. Isi Wiki dan item Knowledge, Rules, Memory yang berisi: A6, dengan item uji.
8. Tema terang dan ukuran minimum jendela: A4 (Appearance) dan A8.
