# A5. Alur kerja: Kanban, Review, composer, dan spec-driven

Pengamatan Sokudo 0.19.6 (macOS), 2026-10-31, lewat computer use, hanya di jendela Sokudo. Tema gelap. Cakupan: Kanban (papan dan tugas), panel Review, composer chat (agent, Plan mode, izin), dan alur spec-driven (Requirements → Design → Tasks).

Tanda: **[terlihat langsung]** tampak di layar; **[inferensi]** dugaan; **[usulan]** bahan untuk Astra, belum keputusan.

Bukti: berkas `a5-*.jpg` di `docs/sokudo-observation/screens/` (tidak di-commit). Ukuran dalam piksel tangkapan, kasar.

## 0. Yang dibuat, diubah, dan dibersihkan

| Item | Tindakan | Keadaan akhir |
| --- | --- | --- |
| Tugas Kanban `ATM-TEST-task` (TASK-001) | Dibuat, dipindah, dibuka, lalu dihapus lewat dialog konfirmasi | Dihapus; papan kosong |
| Chat uji "Hello CLI Tool" (OpenCode, model gratis `big-pickle`, Plan mode, izin Default permissions) | Dibuat dengan dua pesan, lalu dihapus lewat menu chat | Dihapus. Worker spec yang masih berjalan berhenti (indikator berubah jadi centang) |
| Chat uji kedua `ATM-TEST: tiny Python…` (Claude Code CLI, model `claude-haiku-5-5`, Plan mode, Default permissions, slug spec `atm-test`) | Dibuat, tahap Requirements disetujui (Approve), lalu dihapus lewat menu chat saat tahap Design masih berjalan | Dihapus. Worker spec berhenti dengan galat ("Spec Reviser failed · Review of design failed"), sekitar 2 menit setelah dihapus; putaran di header kartu Agent Mode tetap tampil sebagai indikator basi. Dipakai kuota Claude pemilik produk: sesi turun dari 45% ke sekitar 30% |
| Folder `.sokudo/specs/hello-cli` (dan kemungkinan `atm-test`) di folder workspace `test soduko` | Dibuat sendiri oleh Sokudo saat spec dimulai | **Masih ada.** Isinya tidak diperiksa, karena pengaman sesi menolak akses ke folder `.sokudo` dan saya tidak mencari jalan memutar. Pemilik produk bisa menghapusnya bila tidak dibutuhkan |
| Ukuran jendela | Pemilik produk mengecilkannya; saya memperbesarnya lewat Window → Zoom | Memenuhi layar (1512×949), bukan mode layar penuh |
| Draf composer "ATM-TE" | Hilang saat dialog sistem ditutup | Tidak ada sisa |

Yang **tidak** dilakukan: tombol Undo dan Review pada pekerjaan milik pemilik produk, Export, Import, dan Clone agent. Isi percakapan dan kode milik pemilik produk (chat "AI Agency Agent Logic", diff `agency.py`) tidak dicatat; hanya struktur tampilannya.

Efek samping yang perlu diketahui: kontrol layar penuh dipakai untuk mengetik (lihat §6.3), sehingga aplikasi Discord, Safari, dan VS Code tersembunyi sementara. Tidak ada yang ditutup.

## 1. Kanban

Bukti: `a5-kanban-*.jpg`.

### 1.1 Papan

- **[terlihat langsung]** Layar penuh dengan bilah atas "←  Kanban Board". Di bawahnya kolom cari "Search tasks by id, title, assignee, or label…" (ikon kaca pembesar), tombol Filter (ikon corong), dan segmented **Board · List**. Detail Board dan List sama dengan A1 §4.6.
- Empat kolom sama lebar dengan titik warna dan hitungan: **Backlog** (abu), **In Progress** (biru), **Under Review** (kuning), **Done** (hijau), masing-masing "(N items)".
- **[terlihat langsung]** Tombol "+ Create" muncul di bawah judul kolom yang sedang disorot kursor, bukan permanen. Pintasan membuat tugas: `⌘T`.
- Kartu tugas: baris kecil ID dengan panah ("TASK-001 ⌄") dan judul tebal di bawahnya. Kartu berbingkai tipis, sudut membulat, selebar kolom, tinggi ±60 px. Pada tugas uji tanpa assignee, label, atau deskripsi kartunya hanya memuat dua baris itu.
- **[terlihat langsung]** Format ID: `TASK-NNN`, nomor urut tiga digit per workspace.

### 1.2 Dialog Create Task

Modal lebar ±1000 px dengan dua kolom. Bukti: `a5-kanban-create-task-dialog.jpg`, `a5-kanban-create-task-priority-open.jpg`, `a5-kanban-create-task-label-input.jpg`.

- Header "Create Task" dan tombol ×.
- **Kolom kiri**: judul besar sebagai kolom teks ("What needs to be done?"), lalu **Description** dengan editor teks kaya. Toolbar editor: tebal, miring, coret, kode sebaris, H1, H2, daftar bullet, daftar bernomor, kutipan, blok kode. Placeholder "Add a more detailed description… (type @ to mention)". Jadi `@` menyebut sesuatu (berkas, tugas, atau agent) di dalam deskripsi.
- **Kolom kanan "Properties"**:
  - **Status**: dropdown dengan titik warna, bawaan sesuai kolom tempat Create ditekan (Backlog);
  - **Assignee**: dropdown "Unassigned";
  - **Priority**: dropdown dengan ikon panah per tingkat; bawaan **Low**; opsi yang terlihat Highest, High (dan seterusnya, daftar terpotong di dalam modal);
  - **Labels**: tautan "+ Add label" yang membuka kolom "Search or create…" (combobox bisa dicari atau dibuat).
- Footer: "Cancel" dan "Create" (keduanya garis luar).
- **[terlihat langsung]** Modal terpotong pada jendela yang pendek: dropdown di dalamnya terpotong di tepi modal, dan Escape menutup **seluruh** dialog, bukan hanya dropdown.
- **[terlihat langsung]** Pembuatan tugas muncul di papan dengan jeda sekitar satu detik.

### 1.3 Dialog Edit Task

Bukti: `a5-kanban-edit-task-dialog.jpg`, `a5-kanban-task-history.jpg`, `a5-kanban-status-options.jpg`.

- Header "Edit Task" dengan lencana ID `TASK-001`. Judul menjadi teks besar yang bisa disunting.
- **Kolom kiri**, bagian yang bisa dilipat (chevron ⌄, tombol + di kanan):
  - **Description**;
  - **Attachments**: "Add attachment or drop files here";
  - **Sub-tasks**: "Add sub-task";
  - **Linked work items**: "Add linked work item";
  - **Web links**: "Add web link".
  Saat dibuka, keempat bagian sebentar menampilkan "Loading attachments…", "Loading sub-tasks…", "Loading links…".
- Di bawahnya tab **Comments · History**:
  - **Comments**: editor yang sama dengan deskripsi (toolbar lengkap, tombol klip untuk lampiran), dan keadaan kosong "No comments yet.";
  - **History**: log aktivitas berurutan terbaru di atas, satu baris per kejadian dengan pelaku dan waktu relatif: "`<pengguna>` created the task · just now", "`<pengguna>` changed the status · just now".
- **Kolom kanan "Properties"**: Status, Assignee, Priority, Labels, dan **Reporter** (nama pengguna, hanya-baca).
- Footer: kiri "Duplicate" dan "Delete" (teks merah, ikon sampah); kanan "Cancel" dan "Save changes".
- **Status** adalah dropdown berisi empat opsi dengan titik warna: Backlog, In Progress, Under Review, Done. Opsi aktif bertanda centang.
- **Hapus**: dialog kecil "Delete task?" dengan teks `Delete "<judul>"? This removes the task and all its comments, links, and history.` dan tombol "Cancel" dan "Delete" (merah). Bukti: `a5-kanban-delete-task-confirm.jpg`.

### 1.4 Perilaku

- **Drag-and-drop** antar kolom berfungsi (kartu berpindah dan hitungan kolom ikut berubah), tapi tidak konsisten lewat input otomatis: satu dari tiga percobaan drag berhasil.
- Memilih status lewat dropdown di dialog Edit tidak teruji: klik saya pada opsi tidak terdaftar, sehingga perubahan status yang tercatat di History berasal dari drag.
- **[terlihat langsung]** Setiap perubahan status menulis satu baris History ("changed the status"). Teks baris itu tidak menyebut dari dan ke status mana.
- **[inferensi]** Agent bisa membuat dan memindahkan tugas lewat tool Kanban MCP (A3 §3), dan log ditulis dengan pola yang sama.

## 2. Panel Review dan File Changes

Bukti: `a5-review-panel-inline-diff.jpg`, `a5-review-panel-file-tree.jpg`, `a5-file-changes-no-git.jpg`.

- **[terlihat langsung]** Tab **Review** muncul di sidebar kanan (di samping Current Tasks dan Agents Status) ketika ada perubahan file dari agent. Header "Review" dengan jumlah baris ditambah ("+191" hijau), menu ⋯, dan tiga ikon: **Jump to file**, **Switch to split diff**, **Toggle file tree**. Kolom "Filter files…". Setelah satu baris file (lencana "A" = ditambahkan, nama berkas, jumlah baris), diff inline dengan nomor baris dan warna latar hijau (ditambah) atau merah (dihapus).
- **Toggle file tree** menampilkan pohon berkas di kanan panel.
- **[terlihat langsung]** Di chat, setiap giliran yang mengubah berkas menutup dengan **kartu perubahan**: ikon pensil, "1 file changed +N −M", daftar berkas (lencana huruf M atau A, jumlah baris), tombol **Review** dan **Undo** (hanya untuk perubahan terbaru), serta baris "Checked by language servers · 1 fresh". Kartu perubahan yang lebih lama berlabel "Stale", dengan tooltip bahwa Undo dan Review hanya tersedia untuk perubahan terakhir. Bila diagnostik mendeteksi galat, kartu memuat lencana merah "1 error" dan teks bahwa Sokudo meminta agent memperbaikinya (auto-fix round 1/2, A2 §2.1).
- Di bawah balasan agent ada empat ikon: salin balasan, jempol naik, jempol turun, dan ulangi.
- **File Changes** (mode git) kosong untuk workspace ini: "No git repositories found in this workspace.", karena folder bukan repo git. Commit, Push, dan alur "Review before shipping" versi git karena itu **tidak teramati**. Pilihan: pemilik produk menjalankan `git init` di workspace uji, atau membuat workspace baru berisi repo uji.

## 3. Composer dan pemilih agent

Bukti: `a5-composer-model-picker-claude.jpg`, `a5-composer-model-picker-opencode-zen.jpg`, `a5-composer-plan-mode-on.jpg`.

- **Pemilih agent dan model** (tombol di composer, berisi ikon CLI + nama model) membuka popover dengan **rel ikon vertikal** di kiri: favorit (bintang), agent kustom (robot), lalu satu ikon per CLI yang tersambung. Di kanan rel: judul CLI, kolom "Search models…", dan daftar model yang bisa digulir. Model terpilih bertanda centang.
- **[terlihat langsung]** OpenCode mengelompokkan model per penyedia: "OPENCODE ZEN (11 models)" dengan nama berakhiran `-free` untuk model gratis.
- **[terlihat langsung]** Popover menutup sendiri saat ada tangkapan layar atau perubahan fokus; klik harus dikirim berurutan tanpa jeda.
- **Menu "+"**: "Add photos & files" dan "Plan mode" (toggle). Bila Plan mode hidup, muncul chip biru **Plan** di composer.
- **Izin**: dropdown dengan dua opsi, **Full access** (ikon perisai oranye, bawaan) dan **Default permissions** (ikon tangan). **[inferensi]** Default permissions meminta persetujuan untuk tindakan agent; tidak diuji.
- Tombol kirim bulat putih (panah atas) aktif saat ada teks. Saat agent bekerja, tombol berganti menjadi **tombol stop** merah (kotak).

## 4. Chat dan giliran agent

Bukti: `a5-chat-agent-running.jpg`, `a5-chat-agent-working-stream.jpg`, `a5-chat-plan-result.jpg`, `a5-chat-work-details-expanded.jpg`.

- **[terlihat langsung]** Tepat setelah kirim: pesan pengguna tampil sebagai gelembung abu rata kanan. Di bawahnya baris agent dengan ikon CLI, nama CLI ("OpenCode"), dan status "Elucidating… 0:05" (kata kerja acak ditambah penghitung waktu). Tab chat dan item sidebar menampilkan putaran selama bekerja.
- Status berubah menjadi "Working for 0:04" dengan teks pemikiran model mengalir di bawahnya (terpotong dengan "Show more"), dan selesai menjadi "Worked for 0m 26s" dengan chevron.
- **Rincian giliran** (klik baris "Worked for"): daftar langkah yang bisa dibuka, misalnya "Thought for 5s", "Naming the chat · `<judul>`" (panggilan tool), dan "Thought".
- **[terlihat langsung]** Judul chat dibuat otomatis oleh agent lewat tool penamaan (grup "Titles" di MCP, A3 §7.4). Chat baru semula "New Chat" lalu berganti jadi "Hello CLI Tool" tanpa campur tangan pengguna.
- **Plan mode**: agent menjawab dengan rencana berformat teks, dua bagian: daftar bernomor berjudul tebal ("Plan (3 requirements)") dan satu kalimat asumsi ("Say the word if you'd prefer Node/Bash instead."). Tidak ada kartu atau tombol Approve pada tahap ini.
- **Menu klik-kanan pada item chat** di sidebar: Pin Chat, Rename Chat, Copy Chat ID, Delete Chat (merah). Hapus meminta konfirmasi: `Delete chat? This permanently removes "<judul>" and all its messages. This can't be undone.` Bukti: `a5-chat-context-menu.jpg`, `a5-chat-delete-confirm.jpg`.
- **Tutup tab chat** (`⌘W`) tidak menghentikan pekerjaan yang masih berjalan di latar.

## 5. Alur spec-driven

Bukti: `a5-chat-spec-flow-opening.jpg`, `a5-chat-spec-requirements-card.jpg`, `a5-agent-mode-panel-spec-log.jpg`, `a5-agent-mode-critic-round.jpg`, `a5-agent-mode-after-delete.jpg`.

### 5.1 Cara memicu

- **[terlihat langsung]** Permintaan yang hanya berbunyi "rencanakan" (§4) **tidak** memulai alur spec; agent cukup menjawab dengan rencana teks.
- Setelah pengguna menulis pesan lanjutan yang menyatakan brief sudah jelas dan meminta alur spec dimulai, agent memanggil tool `begin_spec` (A3 §3.2). Terlihat di chat sebagai baris "Opening the spec flow · hello-cli" dan di pemikiran agent ("brief is clear … I should call begin_spec").
- **[terlihat langsung]** Spec punya **slug** (`hello-cli`) yang dipakai sebagai nama di kartu dan nama folder `.sokudo/specs/hello-cli` di workspace.

### 5.2 Kartu spec di chat

- Kartu satu baris di bawah giliran agent: ikon dokumen, judul tahap (**Requirements**), lencana status (**Reviewing**), slug di kanan, dan baris kedua dengan putaran: "Reviewing requirements… You can see the progress in Agent Mode."
- Kartu **tidak bisa diklik** selama tahap berjalan. Tombol stop atau batalkan untuk tahap ini tidak terlihat. Dengan model gratis kartu ini tidak pernah selesai (§5.4); dengan Haiku 5.5 kartu berubah jadi "Proposed" dan punya aksi persetujuan (§5.5).

### 5.3 Panel Agent Mode (bawah)

- Dibuka dengan `⌘J` (membuka panel bawah di tab Terminal), lalu tab **Agent Mode**. Isi: kartu per pekerjaan dengan judul "SPEC *Requirements · hello-cli*" dan nama sub-agent ("Spec Generator"), ikon aktivitas dan penanda status di kanan (putaran saat berjalan, centang saat selesai atau berhenti).
- Log berupa baris monospace yang menggulir otomatis dan bisa memanjang: "Drafting requirements…", "Consulting Spec Drafter…", lalu pemikiran sub-agent.
- **[terlihat langsung]** Format dokumen requirement mengikuti pola EARS (WHEN/IF/WHILE/WHERE … SHALL), dengan bagian Introduction, Requirement 1–3 masing-masing berkriteria penerimaan bernomor (AC1, AC2, …), dan Assumptions.
- **[terlihat langsung]** Ada dua peran yang bergantian: **penulis** (Spec Drafter) dan **peninjau** (critic). Ini sesuai pengaturan "Max critic ↔ reviser rounds" (default 3) di A3 §1.2.

### 5.4 Hasil percobaan

- Tahap Requirements **tidak selesai dalam sekitar sembilan menit**. Setelah drafter menulis dokumen, peninjau menulis daftar keberatan yang terus bertambah (diberi nomor berurutan lewat puluhan label dua huruf: JV, JW, … RZ, TZ, UB, UG) untuk tugas sekecil "cetak hello", dan banyak keberatan yang dinilai sendiri oleh peninjau sebagai tidak material ("Fine.", "Not a gap").
- **[inferensi]** Model gratis yang kecil menjadi peninjau yang terlalu teliti dan berputar tanpa batas. Batas putaran (3) tampaknya hanya membatasi jumlah bolak-balik penulis dan peninjau, bukan panjang satu giliran peninjau. Tidak ada batas waktu atau token per giliran yang terlihat.
- Tidak ada tombol untuk menghentikan worker spec dari UI. Cara yang berhasil: menghapus chat pemilik worker (menu klik-kanan). Palet perintah "Chat: Stop Generating" tidak muncul ("No matching data") karena chat sendiri tidak sedang menghasilkan teks.
- Biaya: model gratis, jadi tidak memakai kuota langganan Claude. Status bilah bawah menunjukkan persentase kuota yang berubah selama pengamatan, tetapi itu kuota akun yang dipakai bersama aplikasi lain, bukan dari tes ini.

### 5.5 Percobaan kedua: Claude Haiku 5.5 (2026-10-31)

Atas arahan pemilik produk, alur diulang dengan `claude-haiku-5-5` (Claude Code CLI). Pesan tunggal: brief akhir (Python 3, satu berkas `hello.py`, tanpa dependensi, exit code 0), minta mulai alur spec, tanpa kode. Bukti: `a5b-*.jpg`.

- **Pemicu**: dengan brief yang eksplisit Haiku langsung memanggil `begin_spec` pada pesan pertama (slug otomatis dari awal pesan: `atm-test`). Tidak perlu pesan lanjutan seperti pada model gratis.
- **Requirements selesai dalam sekitar enam menit** (model gratis: tidak selesai dalam sembilan menit). Isi dokumen berformat User Story dan Acceptance Criteria bergaya "WHEN … THE tool SHALL …" (sama dengan pola EARS percobaan pertama, ditambah User Story), dengan Introduction, glosarium "Terms used throughout this document", Requirement 1–4, dan Constraints.
- **Kartu "Proposed"**: judul tahap, lencana **Proposed**, potongan dua baris isi dokumen, slug di kanan, dan lima aksi: **Approve** (tombol putih penuh, satu-satunya aksi utama), **Request changes**, **View**, **Regenerate** (garis luar), serta **Discard** (teks abu di kanan). Bukti: `a5b-spec-requirements-proposed-actions.jpg`.
- **View** membuka **tab dokumen baru di sidebar kanan** ("Requirements — atm-test") berisi dokumen markdown yang dirender, ikon cari di kanan atas, dan **kolom komentar di bawah** ("Request changes to this document…") berikut pemilih izin dan tombol kirim. Jadi revisi diminta langsung terhadap dokumen, bukan lewat chat. Bukti: `a5b-spec-requirements-document-panel.jpg`.
- **Approve**: kartu diganti satu baris "✓ Requirements approved · atm-test · **View** · **Revise from this version**" dan tahap **Design** otomatis dimulai (kartu baru "Design · Reviewing"). Persetujuan tidak perlu konfirmasi tambahan. "Revise from this version" memungkinkan membuka kembali tahap dari versi yang sudah disetujui. Bukti: `a5b-spec-requirements-approved-design-starts.jpg`.
- **Agent Mode**: tiap tahap punya kartunya sendiri dan kartu baru ditumpuk di atas. Isi log memuat langkah bernama: "Drafting design…", "Consulting Spec Drafter…", "Spec Critic: Review complete", "Revising design (round 3)…", "Spec Reviser failed". Footer kartu: "N events · last Xm ago" dengan tombol **Copy** dan **Clear**. Indikator di kanan header: putaran saat berjalan, centang saat selesai. Dari sini terlihat tiga peran: **Spec Drafter** (menulis), **Spec Critic** (meninjau), **Spec Reviser** (merevisi).
- **Design tidak selesai dalam sekitar delapan menit**. Peninjau terus menulis keberatan substantif (misalnya diagram urutan yang tidak cocok dengan teks, baris tabel uji tanpa kondisi lulus, dokumen persetujuan yang tidak didefinisikan), sampai putaran revisi ke-3. Tahap **Tasks tidak tercapai** dan UI persetujuan tahap Design dan Tasks tidak teramati. Peninjau yang lebih kuat tidak membuat hasil lebih cepat: untuk tugas "cetak hello" tetap berputar lama.
- **Berhenti**: menghapus chat lewat menu klik-kanan menghentikan worker, tapi tidak langsung: tahap berakhir dengan "Spec Reviser failed" dan "Review of design failed" sekitar dua menit kemudian. Pada percobaan pertama indikator berubah jadi centang; pada percobaan kedua tetap tampil putaran. Jadi tidak ada penanda berhenti yang konsisten.
- **Biaya**: kuota sesi pemilik produk turun dari 45% ke sekitar 30% selama percobaan ini (termasuk beberapa menit setelah chat dihapus). Itu pemakaian nyata, jadi percobaan tidak diteruskan ke tahap Tasks.

**[usulan]** tambahan untuk Astra dari percobaan ini:
1. Kartu dokumen dengan **lima aksi** (Approve, Request changes, View, Regenerate, Discard) sudah cukup sebagai pola persetujuan; Approve hanya satu tombol utama.
2. Revisi diminta **di dokumen** (komentar di bawah dokumen), bukan di chat, supaya riwayat revisi menempel pada versi dokumen.
3. Setelah disetujui, simpan versi dan sediakan "Revise from this version".
4. Batas yang tegas wajib: total waktu per tahap, jumlah token per tahap, dan tombol **Batalkan** yang terlihat di kartu. Kedua percobaan menunjukkan peninjau LLM bisa terus menemukan keberatan, dan biaya langsung terbebani ke kuota pemilik.
5. Tampilkan peran (Drafter, Critic, Reviser) dan nomor putaran di kartu, bukan hanya di log.

## 6. Catatan teknis pengamatan

### 6.1 Ukuran jendela

- Setelah pemilik produk mengubah jendela ke mode jendela biasa, ukurannya berubah menjadi kecil (±800×600 pt pada saat dilaporkan). Tepi jendela tidak bisa saya raih tanpa akses Finder, tapi pemilik produk kemudian mengecilkannya sendiri dan melaporkan **tidak ada ukuran minimum**; ukuran terbaca 105×280 pt (lihat `00-peta-navigasi.md`, Butir terbuka 1). Window → Zoom memperbesar jendela memenuhi layar.
- **[terlihat langsung]** Tombol hijau di pojok kiri atas membuka menu macOS "Move & Resize / Fill & Arrange / Full Screen". Sokudo memakai tombol lampu lalu lintas bawaan saat bukan layar penuh.

### 6.2 Keandalan input

- Input latar (`app_type`, `app_key`) di editor composer menjatuhkan sebagian besar karakter (teks berhenti setelah 5–6 karakter); input layar penuh berhasil memasukkan 138 dan 151 karakter. **[inferensi]** Editor teks kaya mengganti fokus setelah beberapa ketukan sehingga ketukan susulan hilang bila datang dari jalur latar.
- Sebagai pelajaran bagi Astra: komposer pesan sebaiknya tahan terhadap input cepat dan otomasi (pakai `<textarea>` biasa untuk komposer utama dan editor kaya hanya untuk deskripsi).

### 6.3 Kontrol layar

- Jendela sering berada di Space lain karena Space aktif pemilik produk adalah app layar penuh, sehingga klik latar ditolak. Pengamatan berjalan lancar bila Sokudo dan sesi ini berada di desktop biasa.

## 7. Arti untuk Astra (bahan Fase B, belum keputusan)

| Temuan Sokudo | Keadaan Astra | **[usulan]** untuk Astra |
| --- | --- | --- |
| Kanban empat kolom (Backlog, In Progress, Under Review, Done) dengan Board dan List | Board Astra memakai kolom per status worker (`DESIGN_SYSTEM.md` §10) | Pertahankan kosakata status Astra; tambahkan tampilan List dan pencarian. Jangan menyalin empat kolom Sokudo, karena status Astra diturunkan dari fakta worker, bukan dipindah manual |
| Tugas dengan ID `TASK-NNN`, deskripsi kaya, assignee, prioritas, label, sub-tugas, tautan, lampiran, komentar, dan History | Belum ada tugas | Tugas minimal untuk P1: judul, deskripsi, assignee (agent), status, History. Sub-tugas, tautan, dan lampiran P2 |
| History mencatat perubahan status tanpa nilai lama dan baru | Prinsip "daemon menyimpan fakta" | Simpan fakta kejadian lengkap (dari, ke, pelaku, waktu) sebagai baris yang tidak bisa diubah; tampilan menurunkan teksnya |
| Reporter dan assignee (agent) per tugas | Worker punya workspace dan agent | Hubungkan tugas dengan worker: satu tugas menjadi satu worker di worktree sendiri |
| Panel Review dengan diff inline atau split, pohon berkas, dan kartu perubahan per giliran dengan Review/Undo untuk perubahan terakhir | Detail worker hanya Terminal | P1: tab Diff pada detail worker (diff dari `git diff` di worktree). Undo = `git checkout` pada worktree (daemon, bukan agent). Tampilkan "Stale" untuk kartu lama |
| Diagnostik bahasa setelah giliran dan auto-fix sampai 2 putaran | Belum ada | P2 (A2 §8) |
| Pemilih agent dan model dengan rel CLI dan pencarian model; Plan mode; izin Full access dan Default permissions | Jenis agent tetap | Pemilih agent di dialog "Task baru" (Fase C). Mode izin Astra: agent selalu berjalan di unit terbatas (ADR 0005); "Full access" **tidak** ditawarkan; ganti dengan tingkat persetujuan per aksi |
| Chat dinamai otomatis oleh agent lewat tool | Belum ada | Opsional P2; judul default dari tugas |
| Giliran agent: status "Working", penghitung waktu, rincian langkah, tombol stop | Status "needs input" dan terminal | Tampilkan langkah dari transkrip (event S2) dan tombol Hentikan yang memanggil `Interrupt`/`Stop` (sudah ada di `agent.Service`) |
| Alur spec: Requirements → Design → Tasks, penulis dan peninjau bergantian, dokumen EARS, slug, folder spec di workspace | Belum ada | P2. **Wajib** batas tegas per tahap: jumlah putaran, **batas waktu dan token per giliran**, dan tombol batalkan yang terlihat; percobaan ini menunjukkan peninjau bisa berputar tanpa akhir dan tidak ada tombol berhenti. Simpan spec di worktree (`docs/specs/<slug>/`) dan di database daemon, bukan di folder tersembunyi |
| Worker latar tetap berjalan walau tab chat ditutup | Worker punya sesi systemd yang hidup sampai dihentikan | Bedakan jelas "tutup tampilan" dan "hentikan worker"; tampilkan worker yang masih hidup tanpa tab |
| Hapus chat meminta konfirmasi yang menyebut nama dan "can't be undone" | | Samakan untuk hapus worker dan workspace (ketik nama untuk menghapus) |
| Agent Mode: log langsung per sub-agent | Terminal PTY | Tab "Aktivitas" yang menampilkan event berstruktur (Plan, Tool, Pesan) dari transkrip; terminal tetap untuk interaksi |

Catatan batas: ini ide fitur dan alur. Tidak ada teks, ikon, atau nama Sokudo yang disalin; kalimat bantuan Astra ditulis ulang.

## Butir terbuka

1. Tahap Design dan Tasks serta UI persetujuannya: tidak tercapai walau dicoba dengan Haiku 5.5 (§5.5); Design berputar lebih dari delapan menit dan memakai sekitar 15 poin kuota sesi. Tidak diulang lagi kecuali pemilik produk meminta, atau dengan setelan Design dan Requirements "max rounds" diturunkan ke 1 di Settings → Agents untuk mempercepat.
2. Isi `.sokudo/specs/hello-cli` dan `.sokudo/specs/atm-test` dan format berkasnya (tidak diperiksa).
3. Alur git: Commit, Push, Pull, dan "Resolve with AI" (butuh repo git uji).
4. Memindah tugas lewat dropdown status dan perilaku Assignee yang mengaitkan tugas ke agent (apakah menugaskan memulai worker otomatis).
5. Sub-tugas, tautan tugas, lampiran, komentar, dan label yang berisi.
6. Opsi lengkap Priority dan pilihan di List view "Group by".
7. (Terselesaikan) Ukuran minimum jendela: tidak ada.
8. Isi tab Review untuk perubahan multi-berkas dan tombol Switch to split diff.
