# Monorepo — Astra

Pelengkap untuk `docs/research/go-code-rules.md`. Dirujuk dari `CLAUDE.md`.

Astra memakai **satu repository** yang berisi daemon, CLI, desktop app, web UI, dan alat pengembangan. Monorepo ini sengaja dibuat ringan: cukup Makefile dan fitur bawaan Go, tanpa perkakas monorepo tambahan.

---

## 1. Kenapa satu repository

| Alasan | Dampak untuk Astra |
| --- | --- |
| Versi selalu sama | Desktop app dan daemon wajib memakai nomor versi yang sama (PRD N-16). Satu tag rilis menghasilkan semua installer sekaligus. |
| Perubahan API cukup satu PR | Endpoint baru di daemon langsung disertai perubahan di web UI dan desktop app, tanpa koordinasi antar-repo. |
| Frontend dipakai bersama | Web UI dan desktop app (Wails) memakai frontend yang sama dari `web/`. |
| Satu sumber aturan | Satu `CLAUDE.md`, satu set aturan git, dan satu CI, sehingga aturan tidak bisa berbeda antar-bagian. |

**Tidak dipakai:** Nx, Turborepo, Bazel, atau perkakas monorepo lain. Perkakas itu berguna untuk puluhan package dan tim besar; untuk Astra, menambahkannya bertentangan dengan aturan anti-bloat.

---

## 2. Isi repository

| Bagian | Bentuk | Path modul / package | Alasan dipisah |
| --- | --- | --- | --- |
| Root (`cmd/`, `internal/`) | Modul Go utama | `github.com/alfarax9/astra` | Daemon dan CLI, dikompilasi tanpa CGO |
| `desktop/` | Modul Go terpisah | `github.com/alfarax9/astra/desktop` | Wails butuh CGO dan dependensi WebView |
| `tools/` | Modul Go terpisah | `github.com/alfarax9/astra/tools` | Dependensi alat (deadcode, govulncheck) tidak masuk ke `go.mod` daemon |
| `web/` | Project npm | — | Frontend untuk web UI dan desktop app |

Path modul memakai pemilik `alfarax9`; alasannya dicatat di `docs/adr/0002-go-toolchain.md`.

---

## 3. Hubungan antar-modul

### 3.1 Desktop memakai kode dari modul root

Modul `desktop/` butuh kode dari modul root, misalnya tipe API dan client untuk berbicara dengan daemon. Hubungan ini diatur dengan direktif `replace` di `desktop/go.mod`:

```go
module github.com/alfarax9/astra/desktop

go 1.27

require github.com/alfarax9/astra v0.0.0

replace github.com/alfarax9/astra => ../
```

Dengan cara ini:

- Desktop **selalu memakai kode daemon dari commit yang sama**, sehingga tidak ada kemungkinan versi tidak cocok.
- **Tidak perlu menerbitkan versi modul root secara terpisah** atau membuat tag khusus seperti `desktop/v1.0.0`.
- Versi `v0.0.0` di `require` hanya placeholder; `replace` yang menentukan kode mana yang dipakai.

Direktif `replace` hanya berlaku saat membangun modul `desktop/` itu sendiri. Ini aman karena modul desktop tidak pernah dipakai sebagai dependensi oleh project lain.

### 3.2 Package yang boleh dipakai desktop

Desktop hanya boleh mengimpor package root yang memang ditujukan untuk client:

| Boleh | Tidak boleh |
| --- | --- |
| `internal/api/client` (client HTTP + WebSocket ke daemon) | Package capability (`agent`, `deploy`, dan lainnya) |
| `internal/domain` (tipe fakta dan kosakata status) | Package infrastruktur (`store`, `docker`, dan lainnya) |

Karena path modul desktop (`github.com/alfarax9/astra/desktop`) berada di dalam path modul root, aturan `internal` Go mengizinkan desktop mengimpor `github.com/alfarax9/astra/internal/...`.

**Sudah diverifikasi (2026-10-26, Go 1.27.3)** dengan eksperimen di luar repo: modul `desktop` dengan `replace ../` dan `GOWORK=off` berhasil mengimpor `internal/api/client` dan `internal/domain`, sedangkan modul dengan path di luar `github.com/alfarax9/astra` ditolak dengan `use of internal package ... not allowed`. Package client tetap di `internal/api/client`; fallback ke package non-internal tidak diperlukan. Wails sendiri belum ikut diuji; itu bagian dari spike S10.

**Ditegakkan otomatis:** rule depguard `desktop-is-a-client` di `.golangci.yml` menolak impor package capability dan infrastruktur dari `desktop/`. golangci-lint yang dijalankan di dalam `desktop/` memakai `.golangci.yml` di root.

Desktop tidak boleh menjalankan logika bisnis. Semua keputusan tetap di daemon, dan desktop hanya menampilkan serta mengirim perintah.

### 3.3 Pengembangan lokal dengan `go.work`

Untuk pengembangan lokal, `go.work` boleh dipakai agar editor dan `go` command melihat ketiga modul sekaligus:

```sh
go work init . ./desktop ./tools
```

Aturan:

- `go.work` dan `go.work.sum` **tidak di-commit**; keduanya ada di `.gitignore`.
- Build, test, dan CI tidak boleh bergantung pada `go.work`. Makefile mengekspor `GOWORK=off`, sehingga setiap target sudah memeriksa modulnya sendiri tanpa `go.work`.

---

## 4. Pemeriksaan per modul

Setiap modul punya `go.mod` sendiri, jadi `go mod tidy` dan linter dijalankan di masing-masing modul.

| Modul | Pemeriksaan | Dijalankan oleh |
| --- | --- | --- |
| Root | `make check` lengkap | Lokal (sebelum commit dan lewat hook `pre-push`) dan CI |
| `desktop/` | `go mod tidy -diff`, `go vet ./...`, `go test ./...`, `golangci-lint run` | `make check-desktop` (dilewati selama `desktop/go.mod` belum ada) dan CI saat push ke `main` |
| `tools/` | `go mod tidy -diff` | `make check-tools` |
| `web/` | `npm run lint`, `npx knip`, `npm test`, `npm run build` | Lokal bila `web/` berubah dan CI |

---

## 5. Rilis dan versi

- Satu tag `vX.Y.Z` di root repository untuk seluruh rilis: daemon, CLI, desktop app, dan web UI.
- Workflow rilis membangun semua format file per OS (PRD bagian 8.1) dari tag yang sama.
- Nomor versi disuntikkan saat build lewat `-ldflags "-X ..."`, baik ke daemon maupun desktop app, sehingga keduanya melaporkan versi yang sama.
- Tidak ada tag terpisah untuk modul `desktop/` atau `tools/`.

---

## 6. Aturan untuk sesi Claude Code

- **Perubahan tipe API ikut diperbarui di semua pemakainya dalam satu commit.** Bila `internal/api/client` atau `internal/domain` berubah, desktop dan web UI yang memakainya diperbarui di commit yang sama. Bila pemilik produk meminta commit per file (`docs/git-rules.md` §12.4), rangkaian commit itu yang dihitung sebagai satu perubahan, dan commit terakhirnya wajib lulus semua pemeriksaan.
- **Setiap modul tetap rapi sendiri.** Setelah menambah atau menghapus dependensi, jalankan `go mod tidy` di modul yang bersangkutan, bukan hanya di root.
- **Jangan commit `go.work`.**
- **Jangan menambah modul Go baru** tanpa ADR. Tiga modul (root, `desktop/`, `tools/`) sudah cukup untuk MVP.
