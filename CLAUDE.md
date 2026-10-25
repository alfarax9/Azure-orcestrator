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
| `docs/go-rules.md` (lengkap: `docs/research/go-code-rules.md`) | Lapisan package, aturan kode Go, anggaran, `make check` (dimuat di bawah) | Sebelum menulis kode Go |
| `docs/monorepo.md` | Tiga modul Go (root, `desktop/`, `tools/`) dan `web/`, `replace` desktop, `go.work`, pemeriksaan per modul | Sebelum menyentuh `desktop/`, `tools/`, atau `go.mod` mana pun |
| `docs/adr/` | Keputusan arsitektur: 0001 frontend, 0002 toolchain Go | Saat keputusan terkait berubah |

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

## Stack yang sudah diputuskan

- **Daemon dan CLI:** Go 1.27, modul `github.com/alfarax9/astra`, `CGO_ENABLED=0` (`docs/adr/0002-go-toolchain.md`).
- **Frontend:** React + TypeScript + Vite (SPA statis), React Aria Components untuk komponen headless (`docs/adr/0001-frontend-framework.md`).

## Setup setelah clone

1. Pasang Go 1.27.3 dan golangci-lint 2.15.0 (binary resmi, lihat ADR 0002).
2. `make setup` sekali, untuk mengaktifkan hook `commit-msg` dan `pre-push` di `.githooks/`.
3. `npm ci` di `web/`, lalu `npx skills experimental_install` di root untuk memasang taste-skill dari `skills-lock.json`.
4. Opsional: `go work init . ./desktop ./tools` untuk editor. `go.work` tidak di-commit, dan Makefile selalu memakai `GOWORK=off`.

## Aturan git

@docs/git-rules.md

## Aturan kode Go

@docs/go-rules.md
