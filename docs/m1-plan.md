# Rencana M1: inti (daemon, agent, board)

Sumber: PRD §11 dan §5 (F1 onboarding, F2 project, F3 runtime agent, F4 board). Keluar dari M1: **5 sesi agent stabil 2 jam, dan kartu Board pindah kolom otomatis dari fakta.** Keputusan yang sudah diambil di M0 ada di `docs/adr/` (0001 sampai 0005) dan tidak diulang di sini.

## Potongan kerja

Setiap potongan selesai bila `make check` dan pemeriksaan web lulus, dan requirement yang disebut punya kriteria penerimaannya teruji.

| # | Potongan | Requirement | Status |
| --- | --- | --- | --- |
| 1 | **Fondasi:** `internal/domain` (fakta, status, tahap loop strip, `Secret`, ID), `internal/store` (SQLite, migrasi), kontrak status bersama Go dan web (`contract/status-cases.json`) | F3-06, F2-01 (data) | Selesai |
| 2 | **Workspace:** helper `astra-priv`, workspace, project, dan `git worktree` per worker, semuanya dijalankan sebagai user agent workspace (ADR 0005); deteksi stack dan perintah dev/build/test | F2-01, F2-02, F2-03, F3-01 | Lulus di VM (`scripts/vm/test-workspace.sh`); menunggu QA manual |
| 3 | **Agent:** siklus hidup worker (mulai sesi lewat `internal/pty`, kirim pesan sebagai bracketed paste, interrupt, hentikan), pembaca transcript di dalam session host yang mengirim event lewat socket, `Activity` dari event | F3-02, F3-04, F3-05, F3-07 | Lulus di VM dengan agen tiruan (`scripts/vm/test-agent.sh`); belum diuji dengan Claude Code dan Codex sungguhan; menunggu QA manual |
| 4 | **API:** HTTP dan WebSocket `/api/v1/events` (aliran perubahan fakta), pairing dengan token sekali pakai, PTY dua arah; `astrad serve` yang melayani API dan mengawasi worker (ADR 0006) | F1-05, F3-02 | Lulus di VM (`scripts/vm/test-serve.sh`) dan di uji Go dengan WebSocket sungguhan; belum diuji dengan browser dan agent sungguhan; menunggu QA manual |
| 5 | **Web:** Board tersambung ke daemon (store fakta dari WebSocket menggantikan data contoh), detail worker dengan tab Terminal (xterm.js) | F4-01, F4-02, F4-03 | Lulus di browser dengan daemon lokal dan agen tiruan (`scripts/dev/local-daemon.sh`), uji Vitest, dan inspektur taste-skill (0 FAIL); belum diuji dengan agent sungguhan, VPS, dan di tiga font; menunggu QA manual. Tab Chat, Diff, Gate, Feedback dan panel preview menyusul bersama fiturnya |
| 6 | **Onboarding dan server:** tambah server lewat SSH, pemeriksaan kesiapan, pasang daemon dan user agent | F1-01 sampai F1-07 | Lulus lewat SSH sungguhan ke VM (`scripts/vm/test-install.sh`, 58 pemeriksaan) dan di browser; F1-02 sampai F1-05 dan F1-07 selesai di sisi kode, F1-06 sebagian (tanpa web UI yang di-embed ke binary), F1-01 belum (installer bertanda tangan butuh pipeline rilis dan app desktop); belum diuji di VPS sungguhan, Debian 12, atau dengan agent sungguhan; menunggu QA manual (ADR 0007) |

## Keputusan yang masih terbuka

Dari ADR 0003 dan 0005; masing-masing harus diputuskan sebelum potongan yang bergantung padanya. Hak membuat user dan unit sesi sudah diputuskan: helper milik root `astra-priv` (ADR 0005). Onboarding: ADR 0007.

| Keputusan | Dibutuhkan oleh | Status |
| --- | --- | --- |
| Pembaca transcript di dalam session host | Potongan 3 | **Diputuskan (2026-10-29): ya.** Host menjalankan pembaca S2 dan mengirim event lewat frame baru; daemon tidak membuka home agent (ADR 0005, temuan 2) |
| Kebijakan trust folder | Potongan 3 | **Diputuskan (2026-10-29): pengguna menyetujui per worker.** Daemon tidak menulis konfigurasi agent dan tidak mengirim input apa pun sebelum pengguna menjawab dialog; worker berstatus "Needs you" sampai itu (ADR 0003, temuan 8) |
| Event "berjalan" untuk Codex | Potongan 3 | Belum diputuskan: rollout Codex hanya mencatat item yang selesai. Sementara, aktivitas Codex berubah ke "working" saat pesan dikirim dan ke "idle" saat giliran selesai |

## Aturan kerja M1

- Kode Go mengikuti `docs/go-rules.md`: capability mendefinisikan interface kecil, `internal/store` dan `internal/pty` mengimplementasikannya, dan hanya `cmd/astrad` yang merakit semuanya.
- Perubahan bentuk data yang dilihat web (`domain.Worker`) memperbarui `contract/status-cases.json` dan kedua sisi dalam satu rangkaian commit.
- Daemon menyimpan fakta, bukan status; web menurunkan status dari fakta dengan aturan yang sama.
