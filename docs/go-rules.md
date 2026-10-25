# Aturan kode Go untuk Claude Code

Ringkasan dari `docs/research/go-code-rules.md`. Bila ragu, dokumen lengkap dan `docs/adr/` yang menang. Kode Go belum selesai sampai `make check` lulus.

## Modul dan versi

- Go 1.27, toolchain dan versi alat dikunci di `docs/adr/0002-go-toolchain.md`.
- Tiga modul: root (daemon `astrad` dan CLI `astra`, `CGO_ENABLED=0`), `desktop/` (Wails, CGO), `tools/` (deadcode, govulncheck lewat `go tool -modfile=tools/go.mod`).
- golangci-lint dipasang sebagai binary dengan versi terkunci, bukan `go install`.

## Lapisan package (arah impor hanya ke bawah)

| Lapis | Package | Boleh mengimpor |
| --- | --- | --- |
| Protokol | `cmd/astrad`, `cmd/astra`, `internal/api` | Capability, domain; `cmd/astrad` juga infrastruktur untuk merakit |
| Capability | `agent`, `workspace`, `preview`, `feedback`, `gate`, `deploy`, `watch`, `server` | Domain, stdlib; capability lain hanya lewat ADR |
| Domain | `domain` | Stdlib saja |
| Infrastruktur (daun) | `store`, `proxy`, `pty`, `agentcli`, `docker`, `dns`, `remote` | Domain, stdlib, dependensi yang disetujui |

- Capability tidak mengimpor infrastruktur. Capability mendefinisikan interface kecil (1–3 method) yang dibutuhkannya; infrastruktur mengimplementasinya; `cmd/astrad` menyambungkan.
- Constructor mengembalikan tipe konkret. Interface hanya untuk batas ke sistem luar yang perlu di-fake (Docker, SSH, DNS, CLI agent, waktu).
- Ditegakkan oleh `depguard` dan langkah 8 `scripts/check-budgets.sh`.

## Kode

- **Error:** bungkus dengan `%w` dan konteks operasi; cek dengan `errors.Is`/`errors.As`; tangani sekali (kembalikan atau log); huruf kecil tanpa titik; `panic` hanya di startup `cmd/`.
- **Context:** parameter pertama untuk I/O atau operasi lama; tidak disimpan di struct; selalu dihormati.
- **Konkurensi:** setiap goroutine punya pemilik (`sync.WaitGroup.Go` atau `errgroup`); channel ditutup pengirim; field ber-mutex dikelompokkan di bawah mutex-nya; `pty`, `watch`, `agent` punya test `goroutineleak`; timeout diuji dengan `testing/synctest`, bukan `time.Sleep`.
- **Logging:** `log/slog` terstruktur lewat constructor; tidak ada `fmt.Print*`/`log.Print*`/`os.Getenv` di `internal/`.
- **Secret:** tipe `domain.Secret` (`String()` dan `LogValue()` menyembunyikan isi, `Reveal()` eksplisit); tidak pernah ke log, tidak pernah ke proses agent.
- **HTTP:** `net/http.ServeMux` dengan pola method dan path, tanpa framework; handler hanya decode, validasi, panggil capability, encode; timeout server eksplisit.
- **Test:** lewat API publik; table-driven; fake buatan tangan; build tag `integration` untuk Docker, SSH, jaringan nyata; golden file di `testdata/` untuk parser output pihak ketiga.
- **Penamaan:** package satu kata tanpa `util`/`common`/`helpers`/`misc`/`shared`/`base`; `deploy.Service` bukan `deploy.DeployService`; akronim `URL`, `ID`, `HTTP`, `SSH`, `PTY`; setiap package di `internal/` punya `doc.go`; setiap yang diekspor punya komentar doc; `//nolint` wajib menyebut linter dan alasan.
- **Dependensi:** stdlib dulu (`uuid`, `slog`, `encoding/json/v2`, routing HTTP, `synctest` sudah ada). Dependensi baru butuh ADR. Disetujui sejak awal, bergantung hasil spike: `modernc.org/sqlite`, `github.com/creack/pty`, Caddy/CertMagic, `tailscale.com/tsnet`, `golang.org/x/crypto/ssh`, Docker Engine SDK, Railpack.

## Anggaran (gagal = `make check` gagal; menaikkan batas butuh ADR)

Fungsi ≤ 60 baris dan ≤ 40 statement; kompleksitas kognitif ≤ 15; file ≤ 400 baris; nol kode mati, simbol tak terpakai, dan duplikasi > 100 token; ≤ 25 dependensi langsung; binary `astrad` naik ≤ 5% dari `budgets/binary-size.txt`; coverage capability dan domain ≥ 70%; `TODO(#123)` saja; tanpa FIXME, HACK, XXX.

## `make check`

`fmt` → `tidy` → `lint-version` → `lint` → `deadcode` → `test` (`-race`) → `vuln` → `budgets`. Hook `pre-push` menjalankannya; aktifkan dengan `make setup`. `make fix` menjalankan `go fix`; `make integration` menjalankan test bertag `integration`.
