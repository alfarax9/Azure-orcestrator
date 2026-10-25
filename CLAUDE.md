# Astra — Agent Orchestrator + Deploy

Orchestrator agent coding yang hidup di VPS milik user: agent menulis kode di worktree terpisah, setiap worktree punya URL preview untuk klien, lalu di-ship ke production dengan gerbang otomatis dan rollback.

Pemilik produk dan QA: Maulana Alfara Salim. Implementasi: Claude Code.

## Dokumen sumber

| Dokumen | Isi | Kapan dibaca |
| --- | --- | --- |
| `docs/PRD.md` (asli: `docs/prd/PRD-v1.0.pdf`, hanya lokal, tidak di-commit) | Requirement F1–F13, N-01–N-19, arsitektur, API, milestone | Sebelum mengerjakan requirement apa pun |
| `docs/design/DESIGN_RULES.md` | Perilaku, larangan, aksesibilitas, checklist §13 | Sebelum membuat atau mengubah UI |
| `docs/design/DESIGN_SYSTEM.md` | Token, tipografi, komponen, pola layar, kosakata status §10 | Sebelum membuat atau mengubah UI |
| `docs/git-rules.md` | Aturan commit, branch, PR (dimuat di bawah) | Selalu |

Urutan prioritas saat bertentangan: permintaan eksplisit pemilik produk → `DESIGN_RULES.md` → `DESIGN_SYSTEM.md` → default skill.

## Aturan inti

- **Prioritas:** P0 wajib di MVP, P1 bila P0 stabil, P2 setelah MVP. Requirement bertanda `[S1]`–`[S10]` bergantung pada spike; bila spike gagal, revisi lewat ADR sebelum dikerjakan.
- **Definisi selesai** per requirement: kriteria penerimaan lulus saat QA manual, `make check` lulus, layar terkait lulus `DESIGN_RULES.md` §13, ID requirement tercatat di ringkasan sesi, commit sesuai `docs/git-rules.md`.
- **Daemon menyimpan fakta, bukan status.** Status worker diturunkan saat dibaca, memakai kosakata `DESIGN_SYSTEM.md` §10.
- **Kredensial agent, secret production, dan token Cloudflare** tidak pernah tersedia untuk agent dan tidak pernah ditulis ke log.
- **Tidak ada jalur deploy production tanpa approval manusia**, termasuk lewat API.
- **Kepatuhan pihak ketiga (PRD §10) adalah batas keras:** binary Claude Code tidak dimodifikasi, user login dengan akunnya sendiri, nama produk tanpa kata "Claude", SF Pro tidak pernah di-bundle.
- **UI:** komponen hanya memakai token dari `web/src/styles/tokens.css` (turunan `DESIGN_SYSTEM.md` §11); ikon hanya Lucide; satu tombol utama per layar; cek di 1440/768/390 dengan ketiga opsi font.

## Nama

Desktop app `Astra`, CLI `astra`, daemon `astrad`, service `astrad.service`, config project `astra.yaml`.

## Belum tersedia di repo

- **Bagian 10 dokumen riset** (aturan kode, struktur package, `make check`) dirujuk PRD tapi belum disalin ke sini. Minta ke pemilik produk sebelum menulis kode Go.
- **Framework frontend** belum diputuskan; usulan ada di `docs/adr/0001-frontend-framework.md` (status: diusulkan).

## Setup setelah clone

Jalankan `make setup` sekali untuk mengaktifkan hook `commit-msg` di `.githooks/`.

## Aturan git

@docs/git-rules.md
