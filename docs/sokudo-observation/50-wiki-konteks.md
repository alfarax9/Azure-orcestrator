# A6. Wiki Engine dan Project Context

Pengamatan Sokudo 0.19.6 (macOS), 2026-10-31, lewat computer use, hanya di jendela Sokudo. Tema gelap. Cakupan: Settings → Context (tab Wiki, Knowledge Base, Rules, Memory; Skills sudah di A2 §3.5) dan cara konteks dibagikan ke agent.

Tanda: **[terlihat langsung]** tampak di layar; **[inferensi]** dugaan; **[usulan]** bahan untuk Astra, belum keputusan.

Bukti: berkas `a6-*.jpg` di `docs/sokudo-observation/screens/` (tidak di-commit). Isi ringkasan Wiki memuat deskripsi kode milik pemilik produk (`agency.py`); di sini hanya struktur tampilan dan jenis isinya yang dicatat, bukan isi kodenya.

## 0. Yang dibuat, diubah, dan dibersihkan

| Item | Tindakan | Keadaan akhir |
| --- | --- | --- |
| Wiki workspace `test soduko` | **Dibangun** (Build Wiki, model `claude-haiku-5-5`, Claude Effort Low, tanpa Scope, Guidance, dan Ignore rules diubah) | **Masih ada** (3 berkas dianalisis). Tombol hapus Wiki tidak ditemukan; hanya "Rebuild". Disimpan oleh Sokudo di perangkat |
| Item uji `ATM-TEST-knowledge` (Process), `ATM-TEST-rule` (Quality, "Include in every session" hidup), `ATM-TEST-memory` (Gotcha, Pin hidup) | Dibuat satu per tab, dibuka di Edit, lalu dihapus lewat dialog konfirmasi | Dihapus; ketiga tab kembali kosong |
| Toggle "Show functions & classes" di graf | Dihidupkan lalu dimatikan | Mati (semula) |

Biaya: Build Wiki memakai kuota Claude pemilik produk, sekitar 5 poin sesi (29% → 24%). Uji "apakah agent melihat memori yang di-pin" **tidak dijalankan** untuk menghemat kuota (lihat Butir terbuka).

Yang **tidak** dilakukan: Export, Import, dan Clone di menu ⋮ Knowledge dan Rules, dan toggle Auto-update wiki.

Catatan keadaan: workspace berisi pekerjaan pemilik produk. Wiki ikut menganalisis dua berkas `requirements.md` di `.sokudo/specs/atm-test/` dan `.sokudo/specs/hello-cli/` (sisa percobaan A5), jadi keberadaan keduanya terkonfirmasi lewat graf, tanpa membuka folder itu secara langsung.

## 1. Tab Wiki

### 1.1 Keadaan awal dan dialog Build Wiki

Bukti: `a6-wiki-tab-before-build.jpg`, `a6-wiki-build-dialog*.jpg`.

- **[terlihat langsung]** Satu kartu: "Build a living map of this codebase", teks "Sokudo reads every analyzable file, works out how the modules connect, and keeps the result here for you and for agents to consult. It runs in the background — nothing is built until you ask.", dan tombol "Build Wiki" (ikon graf).
- **Dialog Build Wiki** (modal ±475 px, bergulir): subjudul "Map this codebase once, in the background. You can keep working while it runs."
  - **Agent or model**: dropdown agent atau model; nilai awalnya `claude-haiku-5-5`. **[inferensi]** nilai awal mengikuti agent atau model yang terakhir dipakai (provider agent "Pm" baru saja diubah ke model itu). "Every analysis pass runs on this selection. A custom agent contributes its own provider and model."
  - **Claude Effort (Optional)**: Auto (bawaan), Low, Medium, High, Extra High, Max. Muncul hanya untuk model Claude.
  - **Estimasi**: kotak dengan ikon, "3 analyzable files across 0 languages → ~6 analysis passes (2 batches)". Jadi biaya diperkirakan sebelum mulai, dalam satuan "analysis pass".
  - **Scope to subdirectory (Optional)**: placeholder `packages/api`.
  - **Guidance (Optional)**, bagian lipat: area teks, placeholder "Anything the pages should emphasise — a subsystem to go deep on, vocabulary to use, an audience to write for." Catatan: "Steers the written pages only. The layer and tour analysis stay machine-derived, so guidance can never bend the map away from the code."
  - **Ignore rules**, bagian lipat: berkas pola bergaya `.understandignore` ("patterns for files/dirs to exclude from analysis") yang bisa disunting langsung.
  - Footer: "Cancel" dan "Build Wiki".

### 1.2 Proses build

Bukti: `a6-wiki-build-starting.jpg`, `a6-wiki-build-phase-*.jpg`.

- Tombol berganti jadi "⟳ Starting…", dialog menutup, dan kartu berubah menjadi kartu progres: ikon putar, "**Mapping this codebase**", baris fase ("Scan · phase 1 of 11"), bilah progres dengan hitungan ("Scan 0/1 files", "Analyze 0/1 batches", "Pages 0/5 pages") dan persentase, serta tombol merah **Cancel**.
- **[terlihat langsung]** Sebelas fase. Yang terlihat: fase 1 **Scan**, fase 4 **Analyze**, fase 10 **Pages**; fase lain tidak sempat diamati. Total waktu sekitar dua menit untuk tiga berkas.
- Build berjalan di latar; antarmuka lain tetap bisa dipakai.

### 1.3 Setelah selesai

Bukti: `a6-wiki-built-autoupdate-reader.jpg`.

- Kartu menjadi "**Wiki built** · 3 files analyzed. Rebuild whenever the codebase has moved on." dengan tombol **Rebuild**.
- Kartu baru **Auto-update wiki** dengan toggle (mati secara default): "Re-read the parts of the codebase that changed, in the background, without being asked. Checked when you open this workspace, when indexing finishes, and every 15 minutes while it is open — and only started when the change is worth it. Off by default: an update spends analysis passes."
- Bagian **READER** dengan segmented **Pages · Graph**.

### 1.3.1 Pembaca: Pages

Bukti: `a6-wiki-reader-pages.jpg`.

- Dua kolom. Kiri: **pohon halaman** yang bisa dilipat. Urutannya: halaman ringkasan proyek (ikon buku) → halaman lapisan/arsitektur (ikon kubus bertingkat, "Agent Orchestration") → halaman modul (ikon kubus) → halaman terpisah per dokumen ("ATM Test Specification", "Hello CLI Specification"). Nama dipotong dengan "…".
- Kanan: isi halaman berformat teks panjang bernarasi dengan judul, paragraf pembuka, dan sub-judul berupa baris pendek ("How the project is layered", "How the parts connect", "Where to start"). Istilah kunci dicetak tebal, nama berkas dan simbol berfont monospace.
- **[terlihat langsung]** Isi ditulis seperti panduan orientasi untuk pembaca baru: apa proyeknya, lapisannya, bagaimana bagian-bagian terhubung, dan harus mulai membaca dari mana. Disebutkan juga bahan yang **tidak** terlihat di potongan analisis ("The slice does not show what, if anything, calls …"). Jadi halaman jujur soal batas analisisnya.
- **[terlihat langsung]** Nama berkas yang disebut di halaman mengikuti kode sebenarnya; ringkasan menyebut spec di `.sokudo/specs/` sebagai "reference material about intended behaviour, not as inputs that the code reads".

### 1.3.2 Pembaca: Graph

Bukti: `a6-wiki-graph-top.jpg`, `a6-wiki-graph-canvas-tour.jpg`, `a6-wiki-graph-with-functions.jpg`, `a6-wiki-graph-node-inspector.jpg`.

- Toolbar: kolom "Find a node by name or path…", toggle **Show functions & classes** (mati), toggle **Verified edges only** (mati), dan hitungan di kanan ("3 nodes · 0 edges · 8 hidden"; dengan fungsi dan kelas: "11 nodes · 16 edges").
- Baris chip legenda per **lapisan** dengan titik warna dan jumlah: "Agent Orchestration 1", "ATM Test Specification 1", "Hello CLI Specification 1", dan (dengan fungsi) "Unassigned 8".
- Kanvas graf berupa node berbingkai (satu per berkas, garis kiri berwarna sesuai lapisan, judul berkas besar dan path di bawahnya) dengan titik sambung di atas dan bawah. Edge digambar sebagai garis putus-putus. Kanvas bisa di-zoom dengan gulir.
- **Panel kanan**, dua bagian:
  - **Inspektor** ("Select a node to inspect it."): setelah node dipilih memuat nama, lencana jenis ("file") dan lapisan, path, ringkasan satu paragraf, tag (misalnya "service", "multi-agent"), tombol **Open in editor**, lalu daftar relasi **OUTGOING** dikelompokkan per jenis ("contains 8", "exports 8") dan tiap relasi punya lencana asal **"inferred"**.
  - **TOUR**: langkah bernomor ("1. Project Entry Point", "2. Core Orchestration Building Blocks", "3. Planning and Execution Flow", …), masing-masing berjudul dan satu paragraf. Ini jalur baca yang disarankan.
- **[terlihat langsung]** Relasi punya **asal data**: "inferred" (disimpulkan) dibedakan dari "verified" (terbukti) lewat toggle "Verified edges only". Pada percobaan ini hanya relasi inferred yang muncul.

### 1.4 Arsitektur yang tampak

**[inferensi]**
- Wiki dibangun bertahap (scan → analisis per kelompok berkas → penentuan lapisan → penulisan halaman → tur), dengan estimasi "analysis pass" sebelum mulai dan hitungan kemajuan per fase.
- Analisis berjalan lewat agent atau model pilihan pengguna (bukan model khusus), jadi biayanya masuk kuota agent itu.
- Hasil berupa dua bentuk: dokumen teks (Pages) dan graf berjenis (node, edge dengan asal data, lapisan, tur). Graf dan teks diturunkan dari analisis yang sama ("layer and tour analysis stay machine-derived").
- Agent memakainya lewat grup tool MCP **Workspace wiki** (A3 §7.4) dan "Repository at one commit"; pengguna memakainya lewat Pembaca.

## 2. Knowledge Base

Bukti: `a6-knowledge-*.jpg`.

- Header "KNOWLEDGE ITEMS", filter **All · Mine · Agent** (segmented), menu ⋮ dan "+ Add Item". Keadaan kosong: "No knowledge items yet. Click Add Item to create one."
- **Menu ⋮**: **Export Knowledge** (redup bila kosong), **Import Knowledge**, **Clone Knowledge**. Tidak dijalankan.
- **Dialog Create Knowledge Item** (sama dengan A2 §3.3): judul, kategori, isi (opsional), toggle "Include in every session" (bawaan mati).
- **Kategori** (combobox: bisa dicari atau dibuat, "Type to search or create…"): Technical, API, Process, Best Practices, Troubleshooting.
- **Kartu item** setelah dibuat: judul tebal, lencana kategori biru di kanannya, cuplikan isi satu baris di bawahnya, lalu di kanan tombol hapus merah (ikon sampah), "Edit", dan toggle aktif (hijau, hidup secara default). Item baru tidak diberi penanda "Mine".
- **Hapus**: dialog "Delete `<judul>`?" dengan teks "This knowledge item will be removed from the workspace permanently." dan tombol "Cancel" dan "Delete" (merah).
- **[terlihat langsung]** Knowledge tetap kosong walau "Self-improvement" hidup dan agent sudah dijalankan beberapa kali di workspace ini (A5, termasuk dua alur spec). Jadi agent tidak otomatis menulis ke Knowledge pada setiap giliran; ada ambang "substansial" atau waktu tunda yang tidak terlihat. **[inferensi]**

## 3. Rules

Bukti: `a6-rule-*.jpg`.

- Pola sama dengan Knowledge: header "RULES", filter All · Mine · Agent, ⋮, "+ Add Rule". Kosong: "No rules yet. Click Add Rule to create one."
- **Kategori**: Performance, Security, Quality, Documentation, Naming (bisa dicari atau dibuat).
- **Dialog Create New Rule**: Rule Name, Category, Rule Definition (opsional), "Include in every session".
- **Kartu**: nama, lencana kategori, cuplikan definisi, hapus, Edit, toggle aktif.
- **[terlihat langsung]** Kartu rule **tidak menunjukkan** apakah "Include in every session" hidup; tidak ada ikon atau lencana pembeda dibandingkan rule yang tidak selalu disuntikkan.
- Hapus: "This rule will be removed from the workspace permanently."
- Aturan "Rules never take effect until you approve them" berlaku untuk usulan agent (A3 §3.3); rule buatan pengguna langsung aktif.

## 4. Memory

Bukti: `a6-memory-*.jpg`.

- Header "MEMORIES", filter **All · Decision · Gotcha · Structure · Preference · Stale · Archived**, dan "+ Add Memory". Kosong: "No memories yet. Click Add Memory to record one, or let agents distill them for you."
- **Empat jenis**, masing-masing lencana berwarna dan satu kalimat arti (di dalam dropdown Kind):
  - **Decision** (biru): "A choice that was made and should not be re-litigated."
  - **Gotcha** (amber): "A trap that already cost time once."
  - **Structure** (ungu): "Where something lives, or how the pieces fit together."
  - **Preference** (hijau): "How you want things done in this workspace."
- **Dialog**: Kind, Title, Body, toggle "**Pin to every turn**" ("Pinned memories are always injected, ahead of the similarity-ranked ones."). Catatan di bawah Body: "Keep it compact — memories are injected into agent turns."
- **Kartu**: lencana jenis **di depan** judul (berbeda dengan Knowledge dan Rules yang menaruhnya sesudah judul), cuplikan isi, hapus, Edit, toggle aktif. Edit memakai dialog yang sama berjudul "Edit Memory" ("Update what agents remember about this workspace.") dengan tombol "Save Changes". Pin yang dihidupkan tercermin di dialog Edit.
- **Filter "Stale" dan "Archived"**: keberadaan filter menunjukkan memori punya siklus hidup (menjadi usang, diarsipkan) yang tidak diamati karena tidak ada memori yang berumur.
- **Hapus**: "This memory will be removed from the workspace permanently and will stop being injected into agent turns." Hanya memori yang konfirmasinya menyebut efek ke agent.
- **[terlihat langsung]** Memori kosong walau "Automatic memory extraction" hidup dan beberapa giliran agent sudah berjalan. Teksnya mengatakan hanya "substantial" turn yang diekstrak.

## 5. Cara konteks sampai ke agent

Disimpulkan dari teks antarmuka; penyuntikan sebenarnya **tidak diuji**.

| Sumber | Cara sampai ke agent (menurut teks UI) |
| --- | --- |
| Knowledge item atau Rule dengan "Include in every session" | Disuntikkan ke system prompt pada tiap giliran |
| Knowledge atau Rule tanpa tanda itu | **[inferensi]** Tersedia lewat tool MCP baca (grup "Knowledge base" dan "Rules") atau dipilih per giliran |
| Memory dengan "Pin to every turn" | Selalu disuntikkan, lebih dulu daripada yang diperingkat |
| Memory lain | Disuntikkan sesuai peringkat kemiripan |
| Wiki | Lewat tool MCP "Workspace wiki" dan "Repository at one commit"; pengguna lewat Pembaca |
| History compaction | Giliran lama diganti ringkasan saat percakapan panjang; pesan pengguna dan pertukaran terbaru tetap utuh |
| Skills (A2 §3.5) | Berkas `SKILL.md`; tidak diulas di sini |

## 6. Pola desain

**[terlihat langsung]**
1. Semua daftar konteks (Knowledge, Rules, Memory, Agents) memakai kartu baris yang sama: judul, lencana, cuplikan, hapus merah, Edit, toggle aktif. Konsisten, kecuali posisi lencana pada Memory.
2. Dialog tambah dan ubah memakai bentuk yang sama; tombol utama berubah label (Create, Save Changes).
3. Aksi hapus selalu lewat dialog konfirmasi yang menyebut nama item dan sifat permanen; untuk memori disebut juga efeknya ke agent.
4. Operasi berat (Build Wiki) memberi estimasi biaya dalam "analysis pass" sebelum mulai, progres per fase, dan tombol Cancel.
5. Teks bantuan menjelaskan biaya ("an update spends analysis passes") dan alasan nilai bawaan ("Off by default").
6. Data hasil analisis menyertakan asal data (inferred atau verified) dan penyaring untuk hanya yang terbukti.

## 7. Arti untuk Astra (bahan Fase B, belum keputusan)

| Temuan Sokudo | Keadaan Astra | **[usulan]** untuk Astra |
| --- | --- | --- |
| Wiki: peta kode berbentuk halaman bernarasi, graf berlapis, dan tur baca, dibangun sekali atas permintaan, berjalan di latar | Belum ada | P2. Daemon membangun Wiki per project di VPS memakai agent pilihan; hasil disimpan di database dan berkas di worktree utama. Penting: estimasi biaya, progres per fase, tombol batalkan, dan batas token total |
| Estimasi "analysis pass" sebelum build | | Tampilkan jumlah berkas, estimasi token, dan perkiraan biaya sebelum menjalankan; minta persetujuan bila melewati ambang |
| Auto-update Wiki mati secara default, dicek tiap 15 menit hanya bila perubahan layak | | Pemicu otomatis hanya setelah merge ke `main` dan hanya bila diaktifkan admin |
| Graf dengan asal relasi (inferred atau verified) | | Tandai relasi yang berasal dari penguraian kode (verified) dan yang disimpulkan LLM (inferred); tampilkan hanya yang terbukti secara bawaan |
| Tur baca bernomor | | Berguna untuk onboarding agent dan klien baru; P2 |
| Knowledge dan Rules dengan "Include in every session", Memory dengan Pin | Belum ada | Daemon menulis bagian terkelola di berkas konteks worktree (`CLAUDE.md`/`AGENTS.md`) dari item yang aktif dan di-pin; item lain lewat tool. Isi berkas ini dibatasi ukuran. Disuntikkan oleh daemon, bukan oleh agent, sehingga aman dari agent |
| Usulan agent (rule, memori) tidak berlaku sebelum disetujui | Prinsip "tanpa approval manusia" | Sama persis: agent hanya mengusulkan, manusia menyetujui; tampilkan di antrean persetujuan |
| Empat jenis memori (Decision, Gotcha, Structure, Preference) dan status Stale atau Archived | Belum ada | Pakai empat jenis yang sama sebagai ide; siklus hidup (usang setelah N hari tanpa dipakai, arsip) sebagai P2 |
| Self-improvement dan ekstraksi memori otomatis hanya untuk giliran "substansial" | | Wajib ada ambang dan batas biaya; pengamatan ini menunjukkan tidak ada keluaran dari beberapa giliran, jadi pemicunya harus terlihat (tampilkan "N giliran sejak ekstraksi terakhir") |
| Export, Import, Clone untuk Knowledge dan Rules | | Ekspor ke berkas Markdown di repo (`docs/`) lebih cocok bagi Astra karena versi lewat git |

Catatan batas: ini ide fitur dan alur. Tidak ada teks, ikon, atau nama Sokudo yang disalin; kalimat bantuan Astra ditulis ulang.

## Butir terbuka

1. **Penyuntikan ke agent sebenarnya** (apakah memori yang di-pin terlihat oleh agent): tidak diuji karena kuota Claude pemilik produk sudah turun ke sekitar 24%. Uji kecil: chat Haiku "sebutkan judul memori yang di-pin" dengan satu memori uji.
2. Menghapus Wiki: tidak ditemukan tombol selain Rebuild. Tidak dicoba Rebuild.
3. Fase 2, 3, 5–9, dan 11 dari build Wiki, dan perilaku Cancel.
4. Auto-update wiki (toggle, tidak dihidupkan), Scope ke subdirektori, Guidance, dan perubahan Ignore rules.
5. Isi Knowledge dan Memory yang dibuat agent (Self-improvement dan ekstraksi memori), dan siklus Stale atau Archived.
6. Export, Import, dan Clone (tidak dijalankan).
7. Halaman Wiki di dalam IDE (di luar Settings): tidak ditemukan; Wiki hanya ada di Settings → Context. **[inferensi]**
8. Graf pada proyek dengan relasi "verified" dan lebih banyak berkas (pada percobaan ini 3 berkas, 0 relasi antarberkas).
