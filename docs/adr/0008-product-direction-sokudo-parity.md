# ADR 0008: Arah produk: fitur dan alur setara Sokudo, di arsitektur VPS

| Hal | Isi |
| --- | --- |
| Status | **Disetujui** pemilik produk (2026-10-31) |
| Pemutus | Maulana Alfara Salim |
| Pengusul | Claude Code |
| Milestone | Setelah M1; mengubah PRD (§2, §5, §11) |
| Terkait | `docs/sokudo-observation/` (A1–A8, `99-ringkasan-gap.md`); PRD F3, F4, F10, F12, §2, §10; ADR 0005, 0006; ADR 0009 |

## Konteks

Pemilik produk meminta Astra mendapat fitur dan alur yang semirip mungkin dengan Sokudo 0.19.6, sebuah IDE orkestrator agent berbasis Tauri yang berjalan lokal, lalu menambah fitur yang belum dimiliki Sokudo. Keputusan yang sudah diambil pemilik produk (2026-10-30):

- Tiru **fitur dan tampilan**; `DESIGN_RULES.md` dan `DESIGN_SYSTEM.md` disesuaikan (ADR 0009).
- Daemon tetap di VPS; web UI tetap klien (ADR 0005, 0006).
- Observasi dulu, lalu daftar gap, lalu implementasi.

Observasi selesai (A1–A8). Perbedaan mendasarnya: Sokudo adalah **IDE lokal** (folder di laptop, Docker lokal, agent berjalan sebagai proses pengguna), sedangkan Astra adalah **orchestrator di VPS** dengan jalur agent → preview → klien → production, yang tidak dimiliki Sokudo. Meniru berarti memetakan fitur Sokudo ke arsitektur itu, bukan memindahkannya apa adanya.

## Keputusan yang diusulkan

### 1. Batas peniruan (berlaku untuk semua ADR dan kode turunan)

- Yang ditiru: **ide, fitur, alur kerja, dan tampilan umum** hasil pengamatan layar.
- Yang tidak: kode, aset, logo, ikon, ilustrasi, nama, teks bantuan, dan animasi milik Sokudo; tidak mengekstrak frontend dari binary mereka; tidak mengakses backend mereka.
- Semua kalimat UI Astra ditulis sendiri. Nilai ukuran dan warna dari A8 adalah bahan perbandingan, bukan nilai yang disalin (lihat ADR 0009 §2).
- Nama produk tetap tanpa kata "Claude" dan tanpa nama Sokudo (PRD §10).

### 2. Inti Astra tidak diganti

Preview per worker, umpan balik klien, Ship Gate, deploy blue-green, Watch, dan provenance (F5–F9, F13) tetap inti dan **tidak ditunda** demi fitur ala Sokudo. Fitur baru dikerjakan di sela dan di sekitarnya, dengan urutan di §6.

### 3. Fitur yang diadopsi, diadaptasi, dan ditolak

Rincian per fitur ada di `99-ringkasan-gap.md`. Ringkasan keputusan:

| Kelompok | Keputusan |
| --- | --- |
| **Shell** | Diadopsi: shell tiga kolom dengan panel kanan berpanel, panel bawah, bilah status, dan command palette (P0 untuk shell, P1 untuk palet dan bilah status). Rincian visual di ADR 0009 |
| **Agent dan provider CLI** | Diadaptasi: agent sebagai konfigurasi (nama, role, CLI, model, effort, system prompt, fallback) disimpan daemon. Install CLI dari UI **ditolak**; pemasangan CLI agent tetap lewat installer karena menjalankan perintah sebagai root |
| **Provider API** | **Ditambahkan sebagai pembeda** (§4). Tidak ada di Sokudo |
| **Concurrency** | Diadopsi: batas jumlah worker aktif dengan antrean, P0 karena VPS 4 GB |
| **Tugas dan Kanban** | Dipertahankan: kolom Board diturunkan dari fakta (PRD §7), tidak dipindah manual. Diadaptasi: tugas dengan ID, deskripsi, prioritas, History fakta, tampilan List, pencarian |
| **Review dan Undo** | Diadopsi: tab Diff pada detail worker dengan kartu perubahan per giliran; Undo dijalankan daemon pada worktree |
| **Spec-driven** | Diadaptasi, **P2**, dengan batas waktu dan token per tahap dan tombol Batalkan wajib (dua percobaan Sokudo berputar tanpa selesai) |
| **Konteks (Knowledge, Rules, Memory, Wiki)** | Diadaptasi, P1 untuk Knowledge, Rules, Memory; P2 untuk Wiki dan ekstraksi otomatis. Daemon yang menyuntikkan konteks lewat berkas terkelola di worktree; usulan agent menunggu persetujuan manusia |
| **Sakelar capability per workspace** | Diadopsi dan diperkuat: aksi yang tidak bisa dibalik tetap butuh approval manusia walau sakelar hidup |
| **Server MCP untuk tool IDE** | P2; bila dibuat, lewat socket Unix per workspace, bukan TCP loopback bersama |
| **Notifikasi** | Diadopsi (approval, pertanyaan agent, gagal, selesai), per perangkat |
| **Ditolak** | Browser dalam aplikasi dengan CDP untuk agent; Canvas dan presentasi; marketplace skill terbuka; indikator animasi yang bisa dipilih; zoom UI sendiri; sinkron profil SSH lewat server vendor; history compaction |
| **Ditunda (P2 atau di luar MVP)** | UI Docker dan Docker Compose, Database Manager, Git dengan AI, update track Stable/Nightly, rangkai ulang shortcut |

### 4. Provider API OpenAI-compatible (pembeda Astra)

Agent bisa dijalankan bukan hanya lewat CLI interaktif (Claude Code, Codex), tetapi juga lewat **provider API** yang kompatibel dengan OpenAI (preset OpenRouter, OmniRoute, 9Router, dan Custom dengan base URL sendiri).

Arsitektur (sesuai ADR 0005 dan N-11):

- **Agent loop berjalan di daemon**, karena daemon yang memegang API key. Event dikirim ke bentuk `agent.Event` yang sama, sehingga Board dan status tidak berubah.
- **Tool agent dijalankan oleh "tool host"** di dalam unit sesi milik user agent, lewat socket seperti session host. Tool terkurung di worktree dan di batas cgroup. API key tidak pernah sampai ke proses agent.
- **API key disimpan terenkripsi** di SQLite (AES-GCM, stdlib) dengan kunci `/etc/astra/secret.key` (`root:astrad`, `0640`, dibuat installer). Memakai `domain.Secret`; tidak pernah ke log.
- **Fallback**: Astra satu-satunya pemilik keputusan pindah agent (429 atau 402 dari API, atau limit yang terdeteksi di CLI). Gateway dianggap satu endpoint; fallback internal gateway menjadi tanggung jawab pengguna dan didokumentasikan.
- **Usage dan biaya** dicatat per worker dan per giliran; biaya hasil hitung ditandai estimasi.
- Dependensi baru: tidak ada selain stdlib (`net/http`, `crypto/aes`, `crypto/cipher`). Paket baru `internal/llm` (infrastruktur) dan `internal/toolhost`; migrasi `004`.

Batas kepatuhan (PRD §10 tetap berlaku):

1. Key adalah **milik pengguna**. Astra tidak membundel key, tidak menjual atau membagikan kredit, dan tidak menjalankan layanan gateway sendiri.
2. Kredensial langganan (Claude, ChatGPT) **tidak pernah** dipakai untuk memanggil API atau digunakan ulang di luar CLI resminya. Provider API hanya menerima API key.
3. Mode ini diberi label jelas (F3-08): memakai kredit penyedia, terpisah dari langganan.
4. Gateway pihak ketiga ditampilkan netral, tanpa promosi, dan tidak ada fitur yang mendorong penghindaran batas atau syarat layanan penyedia.

**Pertanyaan yang butuh keputusan pemilik produk (B2):** menyuntikkan key gateway ke CLI (misalnya Codex lewat `OPENAI_BASE_URL`) membuat key terlihat oleh proses agent. Usulan: **tidak** didukung di MVP; CLI memakai login akun sendiri, API hanya lewat agent loop daemon.

### 5. Perubahan PRD yang menyertai

Dikerjakan setelah ADR ini disetujui, di berkas `docs/PRD.md` (tidak dalam ADR ini):

- Requirement baru (usulan ID): F4-07 (bilah status), F4-08 (tugas dan History), F10-05 (batas worker aktif), F12-05 (sakelar capability), F14-01 sampai F14-05 (agent, provider API, fallback, usage, server MCP), F15-01 sampai F15-03 (konteks, ekstraksi otomatis, Wiki).
- §2 non-tujuan: baris "Agent selain Claude Code dan Codex" dipersempit menjadi "CLI agent selain Claude Code dan Codex"; baris "Orchestrator agent yang merencanakan dan membagi task" tetap fase 2 dan menjadi acuan spec-driven.
- §11: milestone baru atau penyisipan (lihat §6).
- §15: satu pertanyaan terbuka baru (B2 di atas).

### 6. Urutan kerja yang diusulkan

Tanpa mengubah gerbang yang sudah ada:

| Langkah | Isi | Requirement |
| --- | --- | --- |
| **Fase 0 (sebelum M2)** | Shell dan token baru (ADR 0009); batas worker aktif; pencatatan usage dari transcript CLI | F4-03, F10-05, F14-04 |
| **M1b: agent dan provider** | Definisi agent, fallback, provider API dengan agent loop dan tool host, layar Settings (Agents, Providers, Token Usage) | F14-01 sampai F14-03 |
| **M2 dan M3** | Tetap seperti PRD: preview, feedback, Ship Gate, deploy, Watch | F5–F9, F12 |
| **M1c: konteks dan tugas** | Knowledge, Rules, Memory, tugas dengan History, tampilan List | F4-08, F15-01 |
| **Setelah M3** | Spec-driven, Wiki, ekstraksi otomatis, server MCP | F14-05, F15-02, F15-03 |

Alasan: provider API dan batas worker mengubah cara agent dijalankan, jadi lebih murah dikerjakan sebelum preview dan deploy bergantung pada bentuk agent; fitur konteks dan spec tidak menghalangi alur klien dan production.

## Konsekuensi

**Positif**
- Astra mendapat shell, alur agent, dan konteks yang dikenal pengguna Sokudo, sambil mempertahankan jalur preview sampai production yang tidak dimiliki Sokudo.
- Provider API menambah pilihan biaya dan ketahanan (fallback) tanpa menyentuh kredensial langganan.
- Kebijakan "usulan agent perlu persetujuan manusia" konsisten di Rules, Memory, spec, dan deploy.

**Negatif dan risiko**
- Lingkup bertambah besar (≈15 requirement baru) sementara MVP belum punya preview dan deploy. Risiko jadwal; mitigasi: urutan §6 dan prioritas P0/P1/P2.
- Agent loop di daemon menambah permukaan serangan (key terenkripsi di server, tool host). Mitigasi: kunci di luar jangkauan agent, uji VM bahwa key tidak terlihat di `/proc` milik agent, uji path traversal.
- Fitur spec-driven dan ekstraksi otomatis membakar biaya tanpa batas bila tidak dibatasi. Mitigasi: batas waktu, token, dan biaya per tahap; tombol Batalkan wajib.
- `DESIGN_RULES.md` perlu direvisi (ADR 0009); layar yang sudah lulus §13 harus diperiksa ulang.

## Alternatif yang ditolak

- **Menyalin Sokudo apa adanya (desktop lokal).** Bertentangan dengan keputusan arsitektur VPS dan jalur preview ke production.
- **Menunda seluruh fitur Sokudo sampai M3.** Tidak memenuhi permintaan pemilik produk dan menunda umpan balik pengguna atas shell dan alur agent.
- **Menjadikan Astra gateway API sendiri.** Melanggar batas kepatuhan §4 (membundel key, risiko syarat layanan).

## Pertanyaan terbuka

1. Persetujuan urutan §6, terutama apakah M1b mendahului M2.
2. B2 (key gateway ke CLI): usulan tidak didukung.
3. Apakah Knowledge, Rules, dan Memory berlaku per workspace (usulan) atau per project.
4. Apakah tugas Astra berdiri sendiri atau selalu menjadi worker (usulan: satu tugas menjadi satu worker).
