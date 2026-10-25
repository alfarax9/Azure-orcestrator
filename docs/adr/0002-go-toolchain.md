# ADR 0002: Toolchain Go, path modul, dan gerbang `make check`

| Hal | Isi |
| --- | --- |
| Status | **Diterima** (2026-10-26) |
| Pemutus | Maulana Alfara Salim |
| Pengusul | Claude Code |
| Milestone | M0 |
| Terkait | `docs/research/go-code-rules.md`, `docs/go-rules.md`, N-19 |

## Konteks

Dokumen riset aturan kode Go meminta versi Go, versi golangci-lint, dan path modul dikunci di ADR saat repo dibuat, dan meminta nilai bertanda "verifikasi" dicek terhadap tool yang benar-benar dipakai. Bagian 9 dokumen riset (daftar verifikasi) tidak ikut terkirim, jadi verifikasi di bawah dilakukan langsung terhadap toolchain yang terpasang.

## Keputusan

### Versi yang dikunci

| Komponen | Versi | Cara pasang |
| --- | --- | --- |
| Go | `go 1.27`, `toolchain go1.27.3` | Homebrew di macOS; `go.mod` mengunci toolchain |
| golangci-lint | 2.15.0 (rilis 2026-10-16) | Binary resmi dari GitHub Releases, checksum SHA-256 dicek, dipasang di `$(go env GOPATH)/bin` |
| deadcode | `golang.org/x/tools` v0.52.0 | `go get -tool -modfile=tools/go.mod` |
| govulncheck | `golang.org/x/vuln` v1.9.0 | `go get -tool -modfile=tools/go.mod` |

Makefile mencari golangci-lint di PATH dulu, lalu di `$(go env GOPATH)/bin`, sehingga tidak perlu mengubah konfigurasi shell. `make lint-version` gagal bila versinya bukan 2.15.0.

### Path modul

`github.com/alfarax9/astra` untuk modul root dan `github.com/alfarax9/astra/tools` untuk modul alat, mengikuti pola `github.com/OWNER/astra` di riset. Repository GitHub saat ini bernama `Azure-orcestrator`, jadi path ini belum bisa dipakai untuk `go install` dari luar. Itu tidak menghalangi build lokal maupun rilis binary. Nama repository diputuskan bersama nama resmi produk (PRD §15, sebelum M1). Bila namanya berubah, path modul diganti di satu commit beserta `.golangci.yml`.

### Verifikasi fitur Go 1.27 yang dipakai riset

Semua dicek dengan `go doc` dan source Go 1.27.3:

| Fitur | Hasil |
| --- | --- |
| Package `uuid` di standard library | Ada |
| `encoding/json/v2` dan `encoding/json/jsontext` | Ada |
| `testing/synctest` | Ada |
| `sync.WaitGroup.Go` | Ada |
| `net/http/httptest.NewTestServer` | Ada |
| Profil `goroutineleak` di `runtime/pprof` | Ada, tanpa `GOEXPERIMENT` |
| `go mod tidy -diff` | Ada |
| `.golangci.yml` dari riset | Lolos `golangci-lint config verify` (skema v2) |

### Perbedaan dari teks riset

File di repo adalah sumber kebenaran. Perubahan dari teks riset, masing-masing karena bug atau kebutuhan repo ini:

1. **`PKGS` dibatasi ke `./cmd/...` dan `./internal/...`, hanya folder yang ada.** `./...` ikut memasukkan file Go di `web/node_modules` (paket `flatted`), dan golangci-lint gagal bila diberi pola untuk folder yang belum ada.
2. **`GOLANGCI_VER` tanpa komentar di baris yang sama.** Di riset tertulis `GOLANGCI_VER := 2.X.Y   # ...`; make menyimpan spasi sebelum komentar ke dalam nilai, sehingga `lint-version` selalu gagal.
3. **`fmt` tidak pernah memanggil `gofmt` tanpa argumen**, dan ikut memeriksa file yang belum di-commit. Versi riset menunggu stdin selamanya bila belum ada file Go yang ter-track.
4. **Skrip anggaran menjalankan `go list` sebelum loop.** Versi riset membaca `go list` lewat `< <(...)`, yang exit code-nya diabaikan `set -e`; import cycle atau paket rusak membuat pengecekan nama, `doc.go`, dan arah impor dilewati diam-diam dan anggaran tetap lulus.
5. **Pengecekan `TODO` memakai `git grep --untracked`**, karena `make check` dijalankan sebelum commit, saat file baru belum ter-track.
6. **Hook `pre-push` ditambahkan** di `.githooks/pre-push` (`exec make check`), sesuai bagian 7 riset; `make setup` mengaktifkan kedua hook.

Gerbang diuji dengan menanam pelanggaran sementara: gofmt, depguard (domain dan capability mengimpor infrastruktur), forbidigo, revive, unused, nama package `utils`, `doc.go` hilang, coverage di bawah 70%, infrastruktur mengimpor capability, import cycle, dan `TODO` tanpa nomor issue semuanya membuat `make check` gagal.

### Belum dipasang

- **Workflow GitHub Actions** (bagian 7 riset). Action harus dikunci ke commit SHA, jadi dipasang tersendiri.
- **`budgets/binary-size.txt`**. Baseline dicatat setelah spike S1, sesuai riset.
- **Modul `desktop/`**. Dibuat bersama spike S10 (Wails).

## Konsekuensi

- `make check` lulus pada repo saat ini dalam sekitar 3 detik, dan menjadi gerbang wajib sebelum commit dan push.
- Setiap mesin pengembang perlu Go 1.27.3 dan golangci-lint 2.15.0 binary. Deadcode dan govulncheck terunduh otomatis lewat `go tool`.
- Upgrade golangci-lint atau Go mengubah ADR ini dan `GOLANGCI_VER` di Makefile pada commit yang sama.
