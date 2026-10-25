# Riset: Aturan Kode Go, Struktur Package, dan `make check` untuk Astra

Oct 26, 2026 · @azura

> Dokumen riset bagian 10 yang dirujuk PRD. Bagian 1–5 disimpan apa adanya. Isi file di bagian 6 (Makefile, `.golangci.yml`, skrip anggaran, `tools/go.mod`) sudah diterapkan di root repo, dan workflow CI di bagian 7 diringkas karena belum dipasang. Ringkasan yang dimuat di setiap sesi ada di `docs/go-rules.md`; versi tool, path modul, dan perbedaan dari teks riset dicatat di `docs/adr/0002-go-toolchain.md`. Bagian 9 ("verifikasi") tidak ikut terkirim; hasil verifikasinya juga dicatat di ADR 0002.

## Ringkasan

Astra dibangun dengan Go 1.27, sebagai satu modul daemon yang dikompilasi statis tanpa CGO, dengan arsitektur berlapis yang arah dependensinya hanya ke bawah dan ditegakkan otomatis oleh linter. `make check` adalah satu-satunya gerbang kualitas: kalau tidak lulus, pekerjaan Claude Code belum selesai.

| Keputusan | Pilihan |
| --- | --- |
| Versi Go | `go 1.27` di `go.mod`, toolchain dikunci ke patch terbaru 1.27.x |
| Modul | Tiga modul terpisah: daemon dan CLI (root), desktop app Wails (`desktop/`), dan alat pengembangan (`tools/`) |
| Arsitektur | Empat lapis: protokol → capability → domain, dengan infrastruktur sebagai daun yang mengimplementasi interface milik capability |
| Penegakan batas package | `depguard` di golangci-lint v2 |
| Gerbang kualitas | `make check`: gofmt, `go mod tidy -diff`, golangci-lint, deadcode, `go test -race`, govulncheck, dan skrip anggaran |
| Acuan gaya | Effective Go, Go Code Review Comments, dan Google Go Style Guide; dokumen ini menang bila bertentangan |

Nilai-nilai yang ditandai **verifikasi** di bagian 9 harus dicek terhadap versi tool yang benar-benar dipakai saat repo dibuat, sebelum dikunci.

## 1. Acuan yang dipakai

Aturan Astra tidak menciptakan gaya baru. Aturan di sini diambil dari panduan resmi Go dan style guide yang sudah teruji, lalu dipersempit menjadi batas yang bisa dicek mesin.

| Acuan | Dipakai untuk | Prioritas |
| --- | --- | --- |
| Dokumen ini dan `docs/adr/` | Keputusan khusus Astra | 1 (menang bila bertentangan) |
| [Google Go Style Guide](https://google.github.io/styleguide/go/) | Penamaan, komentar, error, struktur kode | 2 |
| [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) | Checklist review: context, interface, receiver, pesan error | 2 |
| [Effective Go](https://go.dev/doc/effective_go) | Idiom dasar bahasa | 3 |
| [Organizing a Go module](https://go.dev/doc/modules/layout) | Pemakaian `cmd/` dan `internal/` | 3 |
| [Uber Go Style Guide](https://github.com/uber-go/guide) | Konkurensi, startup, dan pola performa | 4 (pelengkap) |
| [Backend Code Structure Orchestrator.inc](https://github.com/OrchestratorInc/agent-orchestrator/blob/main/docs/backend-code-structure.md) (Apache-2.0) | Contoh nyata daemon orchestrator berbahasa Go: lapisan, tempat interface, arah import | Referensi arsitektur |
| [Go 1.27 Release Notes](https://go.dev/doc/go1.27) | Fitur bahasa dan library yang boleh dipakai | Referensi versi |

Dari Orchestrator.inc, ada tiga prinsip yang diadopsi langsung: domain tetap murni tanpa dependensi infrastruktur, adapter adalah daun yang tidak mengimpor lapisan inti, dan lapisan CLI maupun HTTP tetap tipis karena semua logika ada di daemon. Aturan penempatan interface mereka juga dipakai: interface didefinisikan di package yang memakainya, dan constructor mengembalikan tipe konkret kecuali memang dibutuhkan.

## 2. Versi Go dan toolchain

Go 1.27 dirilis pada Agustus 2026 dan menjadi versi minimal Astra. Versi ini cocok dengan target PRD karena Go 1.27 sendiri mensyaratkan macOS 13 Ventura atau lebih baru, sama dengan batas desktop app di N-14.

### Fitur Go 1.27 yang relevan untuk Astra

| Fitur | Dipakai untuk |
| --- | --- |
| Package `uuid` di standard library | ID worker, sesi, dan deployment tanpa dependensi pihak ketiga |
| `encoding/json/v2` dan `encoding/json/jsontext` | API Astra sendiri dan pembacaan transcript agent secara streaming |
| Generic method | Hanya bila benar-benar menghapus duplikasi (lihat 4.9) |
| Profil `goroutineleak` di `runtime/pprof` | Mendeteksi goroutine bocor di sesi PTY dan watcher saat test |
| `go test` menjalankan cek `stdversion` secara default | Mencegah pemakaian API yang lebih baru dari versi di `go.mod` |
| `go fix` dengan modernizer | Merapikan kode ke idiom terbaru secara otomatis |
| `go mod tidy` menggabungkan blok require ganda | `go.mod` tetap rapi tanpa campur tangan manual |
| `net/http/httptest.NewTestServer` + `testing/synctest` | Test HTTP dan timeout tanpa jaringan nyata dan tanpa menunggu waktu nyata |

### Pengaturan `go.mod`

```go
module github.com/OWNER/astra

go 1.27

toolchain go1.27.x // ganti x dengan patch terbaru saat repo dibuat
```

- **Daemon dikompilasi dengan `CGO_ENABLED=0`** agar menjadi satu binary statis yang jalan di semua VPS Linux. Ini mungkin karena SQLite memakai driver pure Go (`modernc.org/sqlite`).
- **Desktop app adalah modul terpisah** (`desktop/go.mod`), karena Wails butuh CGO dan WebView sistem di macOS dan Linux. Pemisahan ini menjaga daemon tetap statis dan dependensinya tetap kecil.
- **Alat pengembangan ada di modul `tools/`** (deadcode, govulncheck) memakai direktif `tool` di `tools/go.mod`, lalu dijalankan dengan `go tool -modfile=tools/go.mod <alat>`. Dependensi alat tidak masuk ke `go.mod` daemon, sehingga tidak dihitung di anggaran dependensi.
- **golangci-lint dipasang sebagai binary dengan versi dikunci**, bukan lewat `go install` atau direktif `tool`. Tim golangci-lint sendiri menyatakan instalasi dari source tidak dijamin berfungsi dan merekomendasikan binary.

### Kebijakan upgrade

- Patch Go (1.27.x) di-upgrade begitu rilis, terutama untuk perbaikan keamanan.
- Versi mayor berikutnya (1.28) diadopsi paling lambat satu bulan setelah rilis, setelah `make check` lulus di CI. Astra tidak boleh tertinggal lebih dari satu versi, karena Go hanya mendukung dua rilis terbaru.
- Setiap upgrade Go disertai `go fix ./...` di commit terpisah (`chore(build): apply go fix modernizers`).

## 3. Struktur package Astra

Satu package mewakili satu kemampuan produk, lapisannya hanya mengimpor ke bawah, dan infrastruktur menjadi daun yang mengimplementasi interface milik capability. Hanya `cmd/astrad` yang mengimpor semuanya, karena di sanalah semua bagian dirakit.

[embedded content: lapisan package · 4 lapis, 19 package, satu arah]

### Layout repository

```text
astra/
  cmd/
    astrad/         daemon: membaca config, merakit semua package, menjalankan server
    astra/          CLI: install, update, uninstall, status
  internal/
    api/            HTTP + WebSocket; tipis, hanya protokol
    agent/          sesi agent, adapter hooks dan transcript per agent
    workspace/      git worktree dan workspace klien
    preview/        deteksi port, URL preview, link berbagi
    feedback/       komentar klien dan konversi jadi task
    gate/           build, test, smoke test, approval
    deploy/         rilis blue-green dan rollback
    watch/          health check, ambang error, insiden
    server/         node yang dikelola: peran, status, resource
    domain/         tipe fakta, kosakata status, fungsi turunan status
    store/          SQLite: tabel fakta dan query
    proxy/          Caddy/CertMagic yang di-embed
    pty/            PTY dan host sesi terpisah
    agentcli/       menjalankan CLI agent resmi (Claude Code, Codex)
    docker/         build dan container lewat Docker Engine
    dns/            Cloudflare DNS-01
    remote/         SSH ke node lain
  desktop/          modul Go terpisah untuk Wails (butuh CGO)
  web/              frontend untuk web UI dan desktop app
  tools/            modul Go untuk alat pengembangan
  scripts/          skrip anggaran dan rilis
  docs/
    adr/            catatan keputusan arsitektur
    git-rules.md
    go-rules.md     ringkasan dokumen ini untuk Claude Code
    design/         DESIGN_RULES.md dan DESIGN_SYSTEM.md
  Makefile
  .golangci.yml
  CLAUDE.md
```

### Aturan per lapis

| Lapis | Package | Boleh mengimpor | Dilarang mengimpor |
| --- | --- | --- | --- |
| Protokol | `cmd/astrad`, `cmd/astra`, `internal/api` | Capability, domain; `cmd/astrad` juga infrastruktur untuk merakit | — |
| Capability | `agent`, `workspace`, `preview`, `feedback`, `gate`, `deploy`, `watch`, `server` | Domain, standard library, capability lain hanya bila tercatat di ADR | `api`, `cmd`, semua package infrastruktur |
| Domain | `domain` | Standard library saja | Semua package Astra lain dan semua dependensi pihak ketiga |
| Infrastruktur | `store`, `proxy`, `pty`, `agentcli`, `docker`, `dns`, `remote` | Domain, standard library, dependensi pihak ketiga yang disetujui | Capability, `api`, `cmd`, dan sesama infrastruktur |

### Cara capability memakai infrastruktur

Capability tidak mengimpor infrastruktur. Capability mendefinisikan interface kecil yang dibutuhkannya, infrastruktur mengimplementasinya, dan `cmd/astrad` menyambungkan keduanya. Contoh untuk `deploy`:

```go
// internal/deploy/deploy.go
package deploy

// Runtime is what deploy needs from the container engine.
type Runtime interface {
	Start(ctx context.Context, spec domain.ReleaseSpec) (domain.ContainerID, error)
	Stop(ctx context.Context, id domain.ContainerID) error
}

// Router switches live traffic between containers.
type Router interface {
	Switch(ctx context.Context, host string, to domain.ContainerID) error
}

type Service struct {
	runtime Runtime
	router  Router
	facts   FactWriter
}

func New(rt Runtime, r Router, f FactWriter) *Service {
	return &Service{runtime: rt, router: r, facts: f}
}
```

```go
// cmd/astrad/main.go (cuplikan)
engine := docker.New(dockerClient)
router := proxy.New(caddyApp)
deploys := deploy.New(engine, router, facts)
```

Package `docker` dan `proxy` tidak tahu `deploy` ada, sehingga bisa diuji dan diganti sendiri-sendiri, dan `deploy` bisa diuji dengan fake kecil tanpa Docker sungguhan.

### Penegakan otomatis

Aturan per lapis dicek oleh `depguard` di golangci-lint (konfigurasi lengkap di bagian 6). Pelanggaran impor membuat `make check` gagal, jadi Claude Code tidak bisa tanpa sengaja membuat capability bergantung langsung pada Docker atau SQLite.

## 4. Aturan kode Go

Setiap aturan di bawah punya satu alasan: kode yang ditulis Claude Code harus bisa dibaca, diuji, dan dihapus dengan aman oleh orang lain. Aturan yang bisa dicek mesin ditandai dengan nama linter-nya.

### 4.1 Error

- Bungkus error dengan konteks operasi memakai `%w`: `fmt.Errorf("start preview %s: %w", workerID, err)`. *(errorlint)*
- Periksa error dengan `errors.Is` dan `errors.As`, bukan perbandingan string atau `==` pada error yang dibungkus. *(errorlint)*
- Error ditangani satu kali: dikembalikan **atau** di-log, tidak keduanya.
- Pesan error huruf kecil, tanpa titik di akhir. *(revive: error-strings)*
- Sentinel error (`var ErrPreviewNotFound = errors.New(...)`) hanya diekspor bila pemanggil memang bercabang karenanya.
- `panic` hanya untuk kesalahan programmer saat startup di `cmd/`. Package di `internal/` tidak pernah panic untuk input dari luar.
- Jangan membungkus pointer ke tipe error dengan `%w`; Go 1.27 vet sudah melaporkannya karena merusak `errors.Is`.

### 4.2 Context

- `ctx context.Context` selalu parameter pertama untuk fungsi yang melakukan I/O atau bisa lama. *(revive: context-as-argument)*
- Context tidak pernah disimpan di struct.
- Setiap loop, tunggu, dan panggilan jaringan menghormati pembatalan context.
- Request HTTP keluar selalu dibuat dengan context. *(noctx)*

### 4.3 Interface

- Interface didefinisikan di package yang memakainya, sekecil mungkin (1–3 method).
- Constructor mengembalikan tipe konkret (`*Service`), bukan interface.
- Jangan membuat interface untuk tipe yang implementasinya hanya satu, kecuali sebagai batas ke sistem luar yang perlu di-fake saat test (Docker, SSH, DNS, CLI agent, waktu).
- Tidak ada package `interfaces` atau `ports` bersama di MVP. Bila dua capability butuh interface yang sama, catat di ADR dulu.

### 4.4 Konkurensi

- Setiap goroutine punya pemilik yang menunggunya selesai (`sync.WaitGroup.Go` atau `errgroup`). Tidak ada goroutine "lepas".
- Channel ditutup oleh pengirimnya. Penerima tidak pernah menutup channel.
- Field yang dilindungi mutex dikelompokkan tepat di bawah mutex-nya dengan komentar.
- Semua test berjalan dengan `-race`.
- Package dengan goroutine panjang (`pty`, `watch`, `agent`) punya test yang mengecek profil `goroutineleak` setelah shutdown.
- Timeout dan retry diuji dengan `testing/synctest`, bukan `time.Sleep`.

### 4.5 Logging

- Memakai `log/slog` dengan field terstruktur. `*slog.Logger` masuk lewat constructor. *(sloglint)*
- Tidak ada `fmt.Print*` atau `log.Print*` di `internal/`. *(forbidigo)*
- Secret tidak pernah ditulis ke log. Tipe secret punya method `LogValue()` yang mengembalikan `[redacted]` (lihat 4.6).
- Level: `Debug` untuk detail pengembangan, `Info` untuk perubahan state penting, `Warn` untuk masalah yang pulih sendiri, `Error` untuk masalah yang butuh tindakan.

### 4.6 Konfigurasi dan secret

- Konfigurasi dibaca sekali di `cmd/astrad` (file + environment), divalidasi, lalu dikirim sebagai struct ke constructor.
- `os.Getenv` dilarang di `internal/`. *(forbidigo)*
- Secret (token Cloudflare, env production, token pairing) memakai tipe `domain.Secret` yang `String()` dan `LogValue()`-nya selalu menyembunyikan isi. Nilai asli hanya bisa diambil lewat method eksplisit `Reveal()`.
- Secret tidak pernah dikirim ke proses agent, termasuk lewat environment.

### 4.7 HTTP dan API

- Routing memakai `net/http.ServeMux` dengan pola method dan path. Tidak ada web framework.
- Handler di `internal/api` hanya: decode, validasi bentuk, panggil capability, encode. Tidak ada logika bisnis.
- Body response ditutup. *(bodyclose)*
- Server punya timeout baca, tulis, dan idle yang eksplisit.

### 4.8 Test

- Test menguji perilaku lewat API publik package, bukan detail internal.
- Table-driven test untuk fungsi dengan banyak kasus.
- Fake buatan tangan, bukan framework mock. Fake tinggal di file `_test.go` atau package `xxxtest` di sebelahnya.
- Test yang butuh Docker, SSH, atau jaringan nyata diberi build tag `integration` dan dijalankan terpisah dari `make check`.
- Parser untuk output pihak ketiga (transcript agent, log build) memakai golden file di `testdata/`.
- HTTP diuji dengan `httptest.NewTestServer` + `testing/synctest` agar tidak memakai jaringan dan waktu nyata.

### 4.9 Generics

- Dipakai hanya bila saat ini sudah ada minimal dua tipe konkret yang memakainya, bukan untuk kemungkinan di masa depan.
- Generic method (baru di Go 1.27) boleh, dengan aturan yang sama.

### 4.10 Penamaan dan komentar

- Nama package pendek, huruf kecil, satu kata, tanpa `util`, `common`, `helpers`, `misc`, `shared`, atau `base`. *(skrip anggaran)*
- Hindari pengulangan: `deploy.Service`, bukan `deploy.DeployService`.
- Akronim konsisten: `URL`, `ID`, `HTTP`, `SSH`, `PTY`.
- Receiver satu atau dua huruf, konsisten dalam satu tipe. *(revive: receiver-naming)*
- Setiap package punya `doc.go` dengan satu paragraf tanggung jawab. Setiap identifier yang diekspor punya komentar doc. *(revive: package-comments, exported)*
- Komentar menjelaskan alasan, bukan mengulang kode. Tidak ada kode yang di-comment out.
- `//nolint` wajib menyebut linter dan alasannya. *(nolintlint)*

### 4.11 Dependensi

- Standard library dulu. Go 1.27 sudah punya `uuid`, `log/slog`, `encoding/json/v2`, routing HTTP, dan `testing/synctest`, jadi dependensi pihak ketiga untuk hal-hal itu dilarang.
- Dependensi baru butuh ADR: alasan, alternatif standard library, lisensi, dan ukuran.
- Dependensi yang disetujui sejak awal: `modernc.org/sqlite`, `github.com/creack/pty`, Caddy/CertMagic, `tailscale.com/tsnet`, `golang.org/x/crypto/ssh`, Docker Engine SDK, dan Railpack (semua bergantung pada hasil spike).

## 5. Anggaran anti-bloat

"Tidak bloated" diterjemahkan menjadi angka yang dicek di setiap commit. Melewati batas membuat `make check` gagal; menaikkan batas butuh ADR.

| Aturan | Batas | Dicek oleh |
| --- | --- | --- |
| Panjang fungsi | ≤ 60 baris dan ≤ 40 statement | `funlen` |
| Kompleksitas kognitif fungsi | ≤ 15 | `gocognit` |
| Panjang file | ≤ 400 baris, di luar `_test.go`, `testdata/`, dan kode generate | `scripts/check-budgets.sh` |
| Kode mati | Nol fungsi yang tidak terjangkau dari `main` maupun test | `deadcode -test` |
| Simbol tidak terpakai | Nol | `unused` |
| Parameter atau return yang tidak pernah berubah | Nol | `unparam` |
| Duplikasi | Nol blok di atas ambang 100 token | `dupl` |
| Nama package serba-guna | Tidak ada `util`, `utils`, `common`, `helpers`, `misc`, `shared`, `base` | `scripts/check-budgets.sh` |
| Package tanpa `doc.go` | Nol di `internal/` | `scripts/check-budgets.sh` |
| Dependensi langsung modul daemon | ≤ 25 | `scripts/check-budgets.sh` |
| Ukuran binary daemon (`linux/amd64`, `-trimpath -ldflags="-s -w"`) | Naik ≤ 5% dari `budgets/binary-size.txt` | `scripts/check-budgets.sh` |
| Coverage package capability dan domain | ≥ 70% per package | `scripts/check-budgets.sh` |
| `//nolint` tanpa alasan | Nol | `nolintlint` |
| TODO tanpa nomor issue | Nol (`TODO(#123): ...` saja yang boleh) | `scripts/check-budgets.sh` |
| FIXME, HACK, XXX | Nol | `godox` |

Baseline ukuran binary dicatat setelah spike S1 selesai. Setiap kali baseline sengaja dinaikkan (misalnya karena dependensi baru yang disetujui), file `budgets/binary-size.txt` diperbarui di commit yang sama dengan ADR-nya.

## 6. Isi `make check`

`make check` menjalankan delapan langkah berurutan dan berhenti di langkah pertama yang gagal. File di bawah sudah diuji sintaksnya (Makefile, YAML, dan bash); nilai yang ditandai **verifikasi** di bagian 9 dicek ulang saat repo dibuat.

| Langkah | Perintah | Yang dicek |
| --- | --- | --- |
| `fmt` | `gofmt -l` | Format standar Go |
| `tidy` | `go mod tidy -diff` | `go.mod` dan `go.sum` sudah rapi |
| `lint-version` | `golangci-lint --version` | Versi golangci-lint sama dengan yang dikunci |
| `lint` | `golangci-lint run` | Semua linter dan batas package (depguard) |
| `deadcode` | `go tool -modfile=tools/go.mod deadcode -test` | Fungsi yang tidak terjangkau |
| `test` | `go test -race -count=1 -coverprofile` | Test, race condition, coverage |
| `vuln` | `go tool -modfile=tools/go.mod govulncheck` | Kerentanan yang benar-benar terpanggil di kode |
| `budgets` | `scripts/check-budgets.sh` | Anggaran di bagian 5 dan aturan infrastruktur sebagai daun |

Isi Makefile, `.golangci.yml`, `scripts/check-budgets.sh`, dan `tools/go.mod` versi dokumen riset sudah diterapkan di root repo. File di repo adalah sumber kebenaran; perbedaan dari teks riset aslinya dicatat di ADR 0002.

## 7. CI dan hook lokal

CI menjalankan perintah yang sama persis dengan yang dijalankan Claude Code di lokal, yaitu `make check`. Tidak ada pemeriksaan yang hanya ada di CI.

### Hook lokal

| Hook | Isi | Sumber aturan |
| --- | --- | --- |
| `commit-msg` | Menghapus baris atribusi AI dari pesan commit | Aturan git PRD bagian 12.9 |
| `pre-push` | Menjalankan `make check`; push dibatalkan bila gagal | Dokumen ini |

Keduanya ada di `.githooks/` dan diaktifkan dengan `make setup`. Hook tidak boleh dilewati dengan `--no-verify` (aturan git 12.4).

### Workflow GitHub Actions

Rencana dari riset (belum dipasang di repo):

- Job `go`: `make check` di `ubuntu-24.04` dengan golangci-lint versi yang sama dengan Makefile.
- Job `web`: `npm ci`, `npm run lint`, `npx knip`, `npm test`, `npm run build` di folder `web/`.
- Job `desktop`: `go vet` dan `go test` modul `desktop/` di macOS, Windows, dan Ubuntu, hanya pada push ke `main`.
- Jadwal harian untuk govulncheck.
- Action dikunci ke commit SHA, bukan tag; pemeliharaan lewat Dependabot.
- Test integrasi (`make integration`, butuh Docker) di job terpisah sebelum rilis; build rilis dipicu tag versi.
