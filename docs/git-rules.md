# Aturan git untuk Claude Code

Sumber: PRD Astra v1.0, bagian 12.

Aturan ini berlaku untuk setiap commit, push, branch, pull request, dan repository yang dibuat atau diubah Claude Code di project Astra, dan **mengalahkan perilaku default Claude Code**. Kalau sebuah aturan tidak bisa diikuti, Claude Code berhenti dan bertanya, bukan mencari jalan pintas.

**Penyesuaian untuk Astra.** Aturan aslinya ditulis untuk project Node. Karena backend Astra adalah Go, pemeriksaan sebelum commit memakai `make check`, dan pemeriksaan npm hanya berlaku untuk folder `web/`. Larangan menyebut Claude (12.1) berlaku untuk atribusi, yaitu siapa yang menulis kode. Menyebut Claude Code atau Codex sebagai agent yang didukung Astra tetap boleh di kode, commit, dan dokumentasi, karena itu bagian dari fitur produk.

## 12.1 Tanpa atribusi Claude atau AI

Repository tidak boleh berisi atribusi Claude atau AI dalam bentuk apa pun:

- trailer `Co-Authored-By: Claude` atau trailer co-author untuk AI;
- trailer `Claude-Session:`;
- baris, badge, atau link "Generated with Claude Code";
- pernyataan bahwa kode ditulis dengan bantuan Claude atau AI, di commit, judul atau deskripsi PR, README, release notes, tag, maupun deskripsi repository.

## 12.2 Identitas

- Commit tercatat atas identitas pemilik repository dari `git config user.name` dan `git config user.email`.
- Claude Code tidak pernah mengubah pengaturan identitas git.
- Kalau salah satu nilai kosong, Claude Code berhenti dan meminta user mengisinya.

## 12.3 Pesan commit

Ditulis dalam bahasa Inggris dengan format:

```
<type>(<scope>): <summary>

<body: what changed and why, wrapped at 72 characters>
```

| Bagian | Aturan |
| --- | --- |
| type | `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `style`, `perf`, atau `build` |
| scope | Nama package atau area Astra: `agent`, `workspace`, `preview`, `feedback`, `gate`, `deploy`, `watch`, `proxy`, `node`, `store`, `api`, `cli`, `desktop`, `web`, `installer`, `design` |
| summary | Kalimat perintah, huruf kecil di awal, tanpa titik, maksimal 72 karakter, misalnya `add wildcard tls for worker previews` |
| body | Opsional untuk perubahan kecil; wajib bila alasannya tidak jelas dari summary. Sebutkan ID requirement, misalnya `Implements F5-02.` |

Tanpa emoji, dan tanpa trailer selain yang diminta user secara eksplisit.

## 12.4 Sebelum setiap commit

Commit hanya boleh dibuat bila semua pemeriksaan ini lulus:

1. `make check` di root repo: gofmt, go mod tidy, golangci-lint, deadcode, `go test -race`, govulncheck, dan skrip anggaran.
2. Bila ada perubahan di `web/`: `npm run lint`, `npx knip`, `npm test`, dan `npm run build` di folder `web/`.
3. Untuk perubahan UI: `lint.mjs` dan `inspect.mjs --app` dari taste-skill (lihat `docs/design/DESIGN_RULES.md` §1).

Lalu Claude Code meninjau diff-nya sendiri:

- Hapus apa pun yang tidak dibutuhkan perubahan itu: kode debug, `fmt.Println` atau `console.log` sementara, kode yang di-comment out, file yang tidak dipakai.
- Pastikan tidak ada secret yang ter-stage. Kalau ada, unstage, beri tahu user, dan jangan commit.
- Satu perubahan logis per commit; perubahan yang tidak berhubungan dipisah.
- Satu commit per fase eksekusi. Satu fase eksekusi adalah satu requirement (satu ID, misalnya F5-02) yang kriteria penerimaannya sudah lulus.
- Bila pemeriksaan atau hook gagal, perbaiki penyebabnya. Jangan pernah memakai `--no-verify`.

## 12.5 Tidak boleh di-commit

- Secret apa pun: file `.env` dan `.env.*` (kecuali `.env.example` berisi placeholder), token Cloudflare, API key, private key SSH (`id_*`), sertifikat dan kunci (`*.pem`, `*.key`, `*.p12`), sertifikat code signing Apple dan Windows.
- Hasil build: binary Go (`bin/`, `dist/`), output Wails (`build/bin/`), installer (`*.dmg`, `*.exe`, `*.msi`, `*.AppImage`, `*.deb`, `*.rpm`).
- `node_modules/`, `web/dist/`, laporan coverage.
- Log, database lokal (`*.db`, `*.sqlite`), data daemon lokal, folder kerja `.taste-skill/`.
- File OS (`.DS_Store`, `Thumbs.db`).
- File biner besar tanpa persetujuan user. Font Plus Jakarta Sans di `web/assets/fonts/` sudah disetujui.

`.gitignore` harus mencakup semua ini sebelum commit pertama.

## 12.6 Branch dan push

- Push hanya bila user memintanya di sesi itu. Commit tidak berarti push.
- Tidak pernah push langsung ke `main` kecuali user memintanya secara eksplisit. Kerja dilakukan di branch `<type>/<deskripsi-singkat>`, dengan ID requirement di deskripsinya, misalnya `feat/f5-02-preview-tls`.
- Tidak pernah force-push ke `main`. Di branch lain, hanya `--force-with-lease` dan hanya dengan persetujuan user.
- Tidak pernah menulis ulang, amend, atau rebase commit yang sudah di-push, kecuali diminta user.
- Tidak pernah menghapus branch atau tag remote tanpa persetujuan user.

## 12.7 Pull request

- Judul mengikuti format commit di 12.3.
- Deskripsi dalam bahasa Inggris berisi: apa yang berubah, kenapa, bagaimana diuji, ID requirement yang diselesaikan, dan hal yang perlu diperiksa reviewer.
- Tanpa atribusi (12.1).
- Judul dan deskripsi ditunjukkan ke user untuk disetujui sebelum PR dibuat.

## 12.8 Membuat repository

- Nama, deskripsi, topik, dan README repository mengikuti 12.1.
- `.gitignore`, hook `commit-msg`, dan `.claude/settings.json` (12.9) dipasang sebelum commit pertama.

## 12.9 Penegakan tiga lapis

Aturan 12.1 ditegakkan di tiga lapis, karena pengaturan Claude Code saja dilaporkan tidak selalu andal.

**Lapis 1: pengaturan Claude Code.** Di `.claude/settings.json` pada root repo dan di `~/.claude/settings.json` milik user:

```json
{
  "attribution": {
    "commit": "",
    "pr": ""
  }
}
```

**Lapis 2: hook commit-msg.** Disimpan sebagai `.githooks/commit-msg` dan dibuat executable (`chmod +x .githooks/commit-msg`):

```sh
#!/bin/sh
# Strip AI attribution from commit messages
grep -v -iE '^Co-Authored-By: *Claude|^Claude-Session:|Generated with \[?Claude Code' "$1" > "$1.tmp"
mv "$1.tmp" "$1"
```

Karena root repo Astra adalah project Go tanpa `package.json`, hook diaktifkan lewat target `make setup` yang wajib dijalankan sekali setelah clone (disebut di README):

```make
setup:
	git config core.hooksPath .githooks
```

**Lapis 3: file aturan.** `docs/git-rules.md` direferensikan dari `CLAUDE.md` di root repo agar dimuat di setiap sesi.

Hook hanya melindungi pesan commit. Deskripsi PR, release notes, dan deskripsi repository diperiksa manual sebelum dikirim.
