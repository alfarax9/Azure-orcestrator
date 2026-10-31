# Ringkasan gap: Sokudo dibandingkan Astra

Disusun 2026-10-31 dari catatan observasi A1–A8 (`00`–`90` di folder ini) dan keadaan Astra setelah M1 (`docs/m1-plan.md`). Ini bahan keputusan untuk ADR 0008 (arah produk) dan ADR 0009 (design system); **belum ada yang diputuskan**.

## Cara membaca

- **Astra sekarang**: **Ada** (sudah dibangun di M1), **Sebagian**, **Rencana** (sudah ada di PRD, belum dibangun), **Belum** (tidak ada di PRD).
- **Usulan**: **P0** wajib di MVP, **P1** setelah P0 stabil, **P2** setelah MVP, **Tolak** (tidak cocok untuk Astra, alasan di kolom terakhir).
- **PRD**: ID yang sudah ada, atau ID baru yang **diusulkan** (ditandai `baru:`). ID baru baru berlaku setelah ADR 0008 disetujui dan PRD diperbarui.
- Prioritas mengikuti PRD §2: hal yang PRD sebut "non-tujuan MVP" tetap P2 kecuali pemilik produk memutuskan lain.

## 1. Shell, navigasi, dan tampilan

| Fitur Sokudo | Astra sekarang | Usulan | PRD | Catatan |
| --- | --- | --- | --- | --- |
| Shell tiga kolom: sidebar kiri, editor bertab dengan split, sidebar kanan berpanel, panel bawah, bilah status (A1 §3) | Sebagian: Board dan detail worker | P0 | F4-01, F4-03 | Shell baru menggantikan rail + daftar worker; detail worker menjadi tab di area tengah (ADR 0009) |
| Bilah status sebagai pintu ke halaman terkait (A1 §3.6) | Belum | P1 | baru: F4-07 | Isi Astra: koneksi daemon, jumlah worker berjalan, usage hari ini, Docker, server |
| Command palette dengan grup, deskripsi, keycap (A1 §5) | Rencana | P1 | F4-05 | Palet menjadi satu sumber perintah dan shortcut (A4 §2) |
| Rangkai ulang shortcut per perangkat, Import/Export (A4 §2) | Belum | P2 | F4-05 | Simpan di localStorage |
| Panel notifikasi in-app per sumber (A1 §6) | Belum | P1 | F4-04 | Notifikasi native sudah P0; panel in-app pelengkapnya |
| Mode warna System/Light/Dark (A4 §1.2) | Rencana | P0 | F11-01 | Token kedua tema dari ADR 0009 |
| Indikator "agent bekerja" dengan banyak pilihan animasi (A4 §1.1) | Belum | Tolak | — | Satu indikator bawaan saja; hormati `prefers-reduced-motion` |
| Zoom seluruh UI (A4 §1.3) | Belum | Tolak | — | Web UI memakai zoom browser |
| Tidak ada ukuran minimum, UI terpotong (A1 butir 1) | — | — | — | Astra wajib punya tata letak ringkas di < 768 px |

## 2. Agent, provider, dan orkestrasi

| Fitur Sokudo | Astra sekarang | Usulan | PRD | Catatan |
| --- | --- | --- | --- | --- |
| Definisi agent: nama, role, CLI, model, effort, system prompt (A3 §1.6) | Belum (jenis agent tetap) | P1 | baru: F14-01 | Disimpan daemon; role bebas dengan preset |
| Fallback agent saat limit rate atau sesi, lintas CLI (A3 §1.6) | Belum | P1 | baru: F14-03 | Daemon pemilik tunggal keputusan pindah |
| Provider hanya CLI lokal, 8 jenis, Install dari UI (A3 §2) | Sebagian: Claude Code dan Codex | P2 | PRD §2 non-tujuan | Agent lain tetap fase 2. Install dari UI ditolak: pemasangan lewat installer |
| **Provider API OpenAI-compatible** (OpenRouter, OmniRoute, 9Router, Custom) — tidak ada di Sokudo | Belum | P1 | baru: F14-02 | **Pembeda Astra.** Agent loop di daemon, tool host di unit sesi, key terenkripsi (rencana Fase C). Lihat ADR 0008 §4 untuk batas kepatuhan |
| Max concurrent workers dengan antrean (A3 §1.3) | Belum | P0 | baru: F10-05 | Melindungi VPS 4 GB; F10-04 (cgroups) tidak membatasi jumlah |
| Project Scripts "Run on worktree create" (A3 §1.4) | Belum | P1 | F2-04 | Bagian `scripts:` di `astra.yaml`, dijalankan sebagai user agent |
| Token Usage: kartu, grafik harian, per model, aktivitas (A3 §4) | Belum | P1 | baru: F14-04 | Parser transcript membaca usage; biaya estimasi ditandai |
| Sisa kuota langganan CLI (A3 §7.3) | Belum | P2 | baru: F14-04 | Hanya bila CLI melaporkannya |
| Sakelar capability per workspace, default aman, berlaku untuk sesi berikutnya (A3 §3.3) | Sebagian (isolasi ADR 0005) | P1 | F12-01, baru: F12-05 | Aksi tidak bisa dibalik tetap butuh approval walau sakelar hidup |
| Server MCP berisi tool IDE untuk agent (A3 §3) | Belum | P2 | baru: F14-05 | Socket Unix per workspace, bukan TCP loopback bersama |
| Plan mode dan izin per pesan (Full access / Default) (A5 §3) | Belum | P2 | F3-07 | "Full access" tidak ditawarkan; agent selalu di unit terbatas |
| Agent menamai chat otomatis (A5 §4) | Belum | P2 | — | Judul default dari tugas |

## 3. Tugas dan alur kerja

| Fitur Sokudo | Astra sekarang | Usulan | PRD | Catatan |
| --- | --- | --- | --- | --- |
| Kanban 4 kolom manual + List, cari, filter (A5 §1) | Ada: Board 5 kolom dari fakta | — | F4-01 | Tetap: kolom Astra diturunkan dari fakta (PRD §7). Tambah tampilan List dan pencarian (P1) |
| Tugas dengan ID, deskripsi kaya, prioritas, label, sub-tugas, lampiran, komentar, History (A5 §1.2–1.3) | Sebagian: worker punya task teks | P1 | baru: F4-08 | History = fakta kejadian lengkap (dari, ke, siapa, kapan) |
| Kartu perubahan per giliran dengan Review dan Undo; panel Review diff inline/split (A5 §2) | Rencana (tab Diff) | P0 | F4-03, F3-05 | Undo = daemon mengembalikan worktree, bukan agent |
| Diagnostik bahasa dan auto-fix setelah giliran (A2 §2.1) | Belum | P2 | — | Butuh batas biaya |
| Spec-driven: Requirements → Design → Tasks, drafter/critic/reviser, kartu dengan Approve/Request changes/View/Regenerate/Discard (A5 §5) | Belum | P2 | PRD §2 non-tujuan ("orchestrator yang merencanakan") | **Wajib** batas waktu dan token per tahap dan tombol Batalkan; kedua percobaan berputar tanpa selesai |
| Git: commit message dan resolusi konflik dengan AI (A3 §1.5) | Belum | P2 | — | |

## 4. Konteks proyek

| Fitur Sokudo | Astra sekarang | Usulan | PRD | Catatan |
| --- | --- | --- | --- | --- |
| Knowledge Base, Rules, Memory dengan "Include in every session" / Pin (A6 §2–4) | Belum | P1 | baru: F15-01 | Daemon menulis bagian terkelola di `CLAUDE.md`/`AGENTS.md` worktree |
| Self-improvement dan ekstraksi memori otomatis; usulan agent menunggu approval (A2 §3.1, A6) | Belum | P2 | baru: F15-02 | Antrean persetujuan yang sama dengan deploy |
| Wiki: halaman, graf berlapis, tur, estimasi biaya, auto-update (A6 §1) | Belum | P2 | baru: F15-03 | Bangun atas permintaan; batas token |
| Skills dari marketplace pihak ketiga (A2 §3.5) | Belum | Tolak (bentuk ini) | — | Hanya dari repo yang diizinkan admin, dengan pratinjau isi |
| History compaction (A2 §3.1) | — | Tolak | — | Urusan CLI agent, bukan Astra |

## 5. Lingkungan dan alat

| Fitur Sokudo | Astra sekarang | Usulan | PRD | Catatan |
| --- | --- | --- | --- | --- |
| Environment bernama dengan variabel plain/secret (A2 §5) | Rencana | P0 | F8-04 | Secret terenkripsi di daemon; tidak pernah ke agent (N-11) |
| Docker engine terkelola, Compose, volume, image (A2 §4, A7) | Rencana (build dan deploy) | P2 untuk UI Docker | F8-01, PRD §2 | Compose multi-service fase 2 |
| Domain lokal `*.sokudo.local` dengan CA lokal | Rencana | — | F5-02 | Astra memakai wildcard TLS publik |
| Database Manager, konsol SQL, transaksi menunggu Commit (A4 §4, A7) | Belum | P2 | — | Pola "tahan dalam transaksi sampai disetujui" layak ditiru bila ada |
| Browser dalam aplikasi, CDP untuk agent (A4 §3, A7) | Belum | Tolak | — | Web UI memakai browser pengguna; preview dibuka sebagai tautan |
| Canvas desain dan presentasi dengan AI, Publish ke Cloudflare Pages (A7, A4 §5) | Belum | Tolak untuk MVP | — | Di luar inti Astra (agent → preview → production) |
| Remote SSH dengan profil tersinkron terenkripsi (A2 §6) | Ada: `astra server add` | — | F1-02, F10 | Daemon di VPS adalah sumber kebenaran; tidak perlu sinkron vendor |
| Terminal multi-sesi, "mention terminal di chat" (A1 §3.5, A7) | Sebagian: tab Terminal | P1 | F4-03 | Aksi "kirim keluaran ke agent" |

## 6. Privasi, pembaruan, dan pengaturan

| Fitur Sokudo | Astra sekarang | Usulan | PRD | Catatan |
| --- | --- | --- | --- | --- |
| Telemetri mati, crash report hidup, capture diagnostik mati (A2 §2.2) | Rencana | P0 | N-18 | Astra: telemetri tidak ada; capture transcript mentah mati secara default |
| Notifikasi desktop untuk approval, pertanyaan, gagal, selesai (A2 §2.1) | Rencana | P0 | F4-04, F11-05 | Notification API browser + app desktop |
| Archive dan Delete workspace (A2 §2.1) | Belum | P1 | F2-01 | Delete = konfirmasi mengetik nama (PRD F12-04) |
| Update track Stable/Nightly (A2 §2.3) | Belum | P2 | PRD §8 | |

## 7. Pola desain lintas layar

Ditarik dari A1–A8; masuk ADR 0009 dan `DESIGN_RULES.md`:

1. Satu tombol utama per keadaan; aksi merusak merah penuh dengan dialog yang menyebut nama objek dan apa yang tidak ikut terhapus.
2. Teks bantuan menjelaskan konsekuensi dan biaya, bukan hanya fungsi.
3. Keadaan kosong selalu bertulis dan menunjuk tombol yang harus ditekan.
4. Operasi mahal (build wiki, spec) memberi estimasi biaya, progres per fase, dan Cancel.
5. Usulan agent tidak berlaku sebelum disetujui manusia.
6. Nilai teknis selalu monospace.
7. Validasi form sebaris dengan garis merah dan pesan; tombol simpan memberi petunjuk.
8. Yang **tidak** ditiru: teks redup di bawah 4,5:1, teks bantuan usang, dan tidak adanya tombol berhenti untuk pekerjaan latar.

## 8. Ringkasan angka

| Usulan | Jumlah baris |
| --- | --- |
| P0 | 7 |
| P1 | 13 |
| P2 | 14 |
| Tolak | 6 |
| Tanpa prioritas (sudah ada atau catatan saja) | 4 |

Usulan ID baru: F4-07, F4-08, F10-05, F12-05, F14-01 sampai F14-05, F15-01 sampai F15-03.
