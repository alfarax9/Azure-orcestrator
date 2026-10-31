# PRD: Astra — Agent Orchestrator + Deploy (v1.0 Development)

> Konversi Markdown dari `docs/prd/PRD-v1.0.pdf` (25 Okt 2026). PDF tidak di-commit; simpan salinannya di path itu secara lokal. Bila ada beda, PDF adalah sumber kebenaran.

## Status dokumen

PRD ini adalah spesifikasi development untuk MVP. Claude Code membangun dari dokumen ini; setiap requirement punya ID, prioritas, dan kriteria penerimaan yang bisa diverifikasi saat QA manual.

- **Prioritas:** P0 wajib ada di MVP, P1 masuk MVP bila P0 sudah stabil, P2 untuk setelah MVP.
- **Bergantung pada spike.** Requirement bertanda `[S1]`, `[S2]`, dst. bergantung pada hasil spike di dokumen riset bagian 5. Kalau spike gagal, requirement terkait direvisi lewat ADR sebelum dikerjakan.

### Perubahan setelah v1.0

Perubahan berikut disetujui pemilik produk lewat ADR dan **mengalahkan** PDF v1.0 bila berbeda: **ADR 0008** (2026-10-31: arah produk setara Sokudo; requirement F4-07, F4-08, F10-05, F12-05, F14, F15; urutan Fase 0, M1b, M1c) dan **ADR 0009** (2026-10-31: shell panel mengambang dan token baru). Baris yang ditambah atau diubah memuat tanda `(ADR 0008)` atau `(ADR 0009)`.

## 1. Ringkasan produk

Astra adalah orchestrator agent coding yang hidup di VPS milik user: agent menulis kode di worktree terpisah, setiap worktree langsung punya URL preview yang bisa dibuka klien, dan hasilnya dikirim ke production dengan satu klik, lengkap dengan gerbang otomatis dan rollback.

| Hal | Isi |
| --- | --- |
| Versi | 1.0 Development |
| Pemilik produk dan QA | Maulana Alfara Salim |
| Implementasi | Claude Code |
| Latar belakang | Pre-PRD Research: Agent Orchestrator + Deploy (Go) |
| Design | `docs/design/DESIGN_RULES.md` dan `docs/design/DESIGN_SYSTEM.md` |
| Aturan kode | Bagian 10 dokumen riset, disalin ke `CLAUDE.md` di repo |

**Masalah.** Developer sudah memakai agent setiap hari, tapi langkah setelah agent selesai masih manual dan tersebar: push, PR, CI, platform deploy, kirim link ke klien, kumpulkan revisi lewat chat, pantau error. Hanya 20% developer memakai AI untuk deploy dan operasi production (Stack Overflow 2026), padahal 82% organisasi mengalami kegagalan production dari kode AI dalam enam bulan (New Relic 2026).

**Solusi.** Satu sistem yang memegang seluruh loop: prompt → worktree → preview → feedback klien → gerbang → production → pemantauan → kembali ke agent bila ada masalah. Karena orchestrator dan deploy berbagi state, sistem tahu agent, prompt, dan diff mana yang menghasilkan setiap versi di production.

**Positioning.** The agent orchestrator that ships: your agents write the code, your client sees it live, and production stays healthy, all on your own server.

**Pembeda utama** (tidak dimiliki Orca, enowX, Orchestrator.inc, Coolify, Dokploy, maupun CapRover): Worktree Preview sebelum commit, Client Feedback Loop, Ship Gate, Production-to-Agent loop, dan provenance per deployment.

## 2. Tujuan dan non-tujuan MVP

MVP berhasil bila satu developer bisa membawa satu fitur dari prompt sampai URL production klien, dengan klien ikut melihat di tengah jalan dan rollback otomatis bila rusak, tanpa membuka tool lain.

### Tujuan

1. Install desktop app dan daemon di VPS baru sampai preview pertama dalam < 15 menit.
2. Menjalankan beberapa agent (Claude Code, Codex) paralel di worktree terpisah, dengan login akun milik user sendiri.
3. Setiap worktree otomatis punya URL preview HTTPS yang bisa dibagikan ke klien.
4. Komentar klien di preview masuk sebagai task ke agent yang sama.
5. Satu klik Ship menjalankan gerbang otomatis lalu deploy blue-green ke domain klien.
6. Error setelah deploy memicu rollback otomatis dan membuka sesi agent dengan konteks lengkap.

### Non-tujuan MVP

| Tidak dikerjakan di MVP | Fase |
| --- | --- |
| Fitur tim: role, undangan, audit per anggota | 2 |
| Docker Compose atau aplikasi multi-service | 2 |
| Mode agent lokal di Windows | 2 |
| CLI agent selain Claude Code dan Codex (OpenCode dan lainnya); provider API OpenAI-compatible dikerjakan lewat F14-02 (ADR 0008) | 2 |
| Provider DNS selain Cloudflare | 2 |
| Integrasi GitHub PR dan review | 2 |
| Orchestrator agent yang merencanakan dan membagi task (acuan: alur spec-driven, F14-06) | 2 |
| Invoice dan laporan biaya per klien | 2 |
| White-label overlay klien | 2 |
| Kubernetes, Swarm, atau cloud provider terkelola | Tidak direncanakan |

## 3. Persona dan skenario utama

MVP melayani dua persona. Tim developer dan agency kecil dilayani di fase 2.

| Persona | Situasi | Paling dibutuhkan dari MVP |
| --- | --- | --- |
| Freelancer solo | 3–10 project klien aktif, 1–3 VPS, mengerjakan semuanya sendiri | Preview yang langsung bisa dibagikan, feedback klien yang masuk ke agent, workspace per klien |
| Senior developer | Project besar dan kecil, standar kualitas tinggi, sering menjalankan beberapa agent sekaligus | Ship Gate, rollback otomatis, guardrail production, provenance |

### Skenario yang harus mulus

1. **Revisi kecil dari klien.** Klien menulis "tombol checkout terlalu kecil di HP" di overlay preview. Komentar masuk ke worker yang sama, agent memperbaiki, preview ter-update, klien melihat hasilnya tanpa developer membuka chat.
2. **Tiga pendekatan paralel.** Developer menjalankan tiga agent untuk fitur yang sama, membandingkan tiga preview, lalu memilih satu untuk di-ship.
3. **Deploy yang rusak di malam hari.** Error rate naik setelah deploy. Sistem rollback otomatis, membuka sesi agent dengan diff, prompt asli, dan log, lalu mengirim notifikasi.
4. **Mulai dari nol.** Developer membeli VPS baru, meng-install desktop app, dan dalam 15 menit sudah punya preview pertama dari repo klien.

## 4. Alur utama MVP

Satu alur yang harus mulus sebelum fitur lain dikerjakan. Ketiga jalur yang kembali ke agent (komentar klien, gerbang gagal, production bermasalah) adalah inti produk, bukan pengecualian.

```
Task dari kamu → Agent di worktree → Preview live → Gate lulus? ──lulus──→ Approval kamu → Production → Sehat 15 mnt? ──ya──→ Live, selesai
                       ↑     ↑            │              │                                                     │
                       │     └─komentar klien┘           │                                                     │
                       └──────────── gagal: error dikirim ke agent                                              │
                       └──────────── error melonjak: rollback otomatis + sesi agent baru ───────────────────────┘
```

1. **Install:** pasang desktop app di macOS, Windows, atau Linux, sambungkan ke VPS lewat SSH; app memasang daemon dengan satu klik (bagian 8).
2. **Hubungkan domain:** masukkan token Cloudflare; sistem menyiapkan wildcard `*.preview.domainmu.com`.
3. **Tambah project:** clone repo atau buat project baru, lalu login Claude Code atau Codex dengan akun sendiri di terminal bawaan.
4. **Task ke agent:** agent bekerja di worktree terpisah; beberapa agent boleh berjalan bersamaan.
5. **Preview otomatis:** saat dev server agent hidup, URL HTTPS preview langsung aktif dan bisa dibagikan.
6. **Feedback klien:** komentar di overlay preview menjadi task untuk agent yang sama.
7. **Ship:** build, test, dan smoke test berjalan otomatis. Gagal → error kembali ke agent; lolos → menunggu approval kamu.
8. **Production:** blue-green deploy ke domain klien dengan SSL otomatis.
9. **Watch:** 15 menit pertama dipantau ketat; error melonjak → rollback otomatis dan sesi agent baru dengan konteks lengkap.

## 5. Requirement fungsional

Setiap requirement diverifikasi dengan kriteria penerimaannya saat QA manual. ID dipakai di nama branch, commit, dan ringkasan sesi (misalnya `feat/f5-02-preview-tls`).

### F1. Instalasi dan onboarding

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F1-01 | Desktop app bisa di-install di macOS, Windows, dan Linux dengan format pada bagian 8 [S10] | P0 | Installer terbuka tanpa peringatan blokir OS di ketiga OS |
| F1-02 | Wizard onboarding: tambah server lewat host, user, dan SSH key (pilih key yang ada atau buat key ed25519 baru) | P0 | User baru tersambung ke VPS tanpa membuka terminal di laptopnya |
| F1-03 | Pemeriksaan kesiapan server sebelum install (OS, arsitektur, RAM, disk, port 80/443, systemd, sudo) | P0 | Setiap kegagalan menampilkan penyebab dan cara memperbaiki |
| F1-04 | Install daemon di VPS dengan satu klik, termasuk Docker bila belum ada (dengan persetujuan user) | P0 | Daemon berjalan sebagai service systemd dan bertahan setelah reboot |
| F1-05 | Pairing desktop app dengan daemon lewat token sekali pakai yang dikirim melalui SSH | P0 | Token tidak pernah tampil di layar dan kedaluwarsa setelah dipakai |
| F1-06 | Install tanpa desktop app: satu perintah di VPS yang menampilkan kode pairing untuk web UI | P1 | Web UI bisa dibuka dari browser atau HP dan dipasangkan dengan kode itu |
| F1-07 | Checklist pertama kali: server, domain, login agent, project pertama | P0 | Checklist menampilkan persentase nyata dan hilang setelah semua selesai |

### F2. Project dan workspace klien

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F2-01 | Workspace per klien yang berisi project, domain, dan server | P0 | Project klien A tidak terlihat dari workspace klien B |
| F2-02 | Tambah project dari URL git (HTTPS atau SSH) atau buat repo kosong | P0 | Clone berjalan di VPS, bukan di laptop |
| F2-03 | Deteksi stack otomatis (Dockerfile, Node, Go, Python, PHP) beserta perintah dev, build, dan test | P0 | Benar pada ≥ 9 dari 10 aplikasi sampel; user bisa mengoreksi hasil deteksi |
| F2-04 | Konfigurasi project tersimpan sebagai file di repo (`astra.yaml`) agar bisa di-review | P1 | Mengubah file lalu commit mengubah perilaku project |

### F3. Runtime agent

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F3-01 | Buat worker: satu task, satu agent, satu worktree dengan branch sendiri | P0 | Dua worker di project yang sama tidak pernah menulis ke file yang sama |
| F3-02 | Jalankan Claude Code dan Codex sebagai CLI interaktif di PTY, tanpa modifikasi binary [S1] | P0 | 5 sesi paralel stabil selama 2 jam |
| F3-03 | Terminal bawaan tempat user login ke agent dengan akun miliknya sendiri | P0 | Daemon tidak pernah membaca, menyalin, atau mengirim kredensial agent |
| F3-04 | Sesi agent tetap berjalan saat desktop app ditutup atau daemon di-restart [S1] | P0 | Setelah restart daemon, terminal tersambung kembali dengan scrollback utuh |
| F3-05 | Tampilan terstruktur (tool call, file yang diedit, diff) dari hooks dan file transcript, tanpa mode headless [S2] | P0 | Kejadian tampil < 1 detik setelah terjadi |
| F3-06 | Status worker diturunkan dari fakta (aktivitas agent, preview, gate, deployment), tidak disimpan langsung | P0 | Status selalu sesuai tabel kosakata di `DESIGN_SYSTEM.md` §10 |
| F3-07 | Interrupt, kirim pesan, dan hentikan agent dari UI | P0 | Interrupt bekerja dalam 1 detik |
| F3-08 | Mode Agent SDK sebagai opsi berlabel yang memakai API key user | P2 | Label menyebut bahwa mode ini memakai kredit terpisah |

### F4. Board dan worker

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F4-01 | Board dengan kolom Working, Needs you, Review, Ready to ship, Live | P0 | Kartu pindah kolom otomatis saat fakta berubah |
| F4-02 | Loop strip di setiap kartu dan detail worker | P0 | Tahap berjalan, selesai, gagal, dan dilewati terbaca tanpa warna (ada teks) |
| F4-03 | Detail worker dengan tab Terminal, Chat, Diff, Gate, Feedback dan panel preview | P0 | Semua tab bisa dibuka dengan keyboard |
| F4-04 | Notifikasi native untuk Needs you, gate gagal, rollback otomatis, dan komentar klien | P0 | Tidak ada notifikasi untuk perubahan status lain |
| F4-05 | Command palette (`⌘K` / `Ctrl+K`) untuk worker, project, server, dan aksi | P1 | Semua aksi utama bisa dijalankan dari palette |
| F4-06 | Bandingkan beberapa worker untuk task yang sama secara berdampingan (preview + diff) | P1 | Tiga preview tampil berdampingan di layar 1440px |
| F4-07 | Bilah status di bawah shell: koneksi daemon, worker berjalan, pemakaian hari ini, server; tiap item membuka halaman terkait (ADR 0008) | P1 | Setiap item bisa dipilih dengan keyboard dan membuka halaman yang tepat |
| F4-08 | Tugas dengan ID, deskripsi, prioritas, dan History fakta (dari, ke, siapa, kapan); tampilan List dan pencarian di samping Board; satu tugas menjadi satu worker (ADR 0008) | P1 | History memuat nilai lama dan baru setiap perubahan; pencarian menemukan tugas lewat ID, judul, atau agent |

### F5. Preview

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F5-01 | Deteksi port dev server di setiap worktree secara otomatis | P0 | Port terdeteksi tanpa konfigurasi pada aplikasi sampel |
| F5-02 | URL HTTPS per worker di `<worker>.<project>.preview.<domain>` dengan sertifikat wildcard via DNS-01 Cloudflare [S4] | P0 | URL aktif < 10 detik setelah dev server hidup; 10 preview bersamaan |
| F5-03 | Bagikan preview ke klien dengan link yang bisa dicabut, kedaluwarsa opsional, dan password opsional | P0 | Link yang dicabut langsung menampilkan halaman "Preview tidak lagi tersedia" |
| F5-04 | Preview tetap privat secara default sampai dibagikan | P0 | Tanpa link berbagi, preview hanya terbuka untuk user yang login |
| F5-05 | Fallback bila wildcard tidak bisa dibuat: satu domain dengan path per worker atau tunnel | P2 | Preview tetap bisa dibuka meski tanpa token Cloudflare |

### F6. Feedback klien

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F6-01 | Overlay feedback disisipkan ke halaman preview lewat reverse proxy [S8] | P0 | Berfungsi di sampel Next.js, Astro, dan Laravel |
| F6-02 | Klien klik elemen lalu menulis komentar tanpa membuat akun | P0 | Komentar tersimpan dengan screenshot dan selector elemen |
| F6-03 | Komentar menjadi task untuk agent di worker yang sama, setelah developer menyetujui atau otomatis bila diaktifkan | P0 | Agent menerima teks komentar, screenshot, dan konteks elemen |
| F6-04 | Klien melihat status komentarnya (Terkirim, Sedang dikerjakan, Selesai) | P1 | Status berubah tanpa klien me-reload halaman |
| F6-05 | Overlay memakai bahasa awam dan tidak pernah menampilkan istilah teknis | P0 | Tidak ada kata worktree, branch, agent, atau commit di overlay |

### F7. Ship Gate

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F7-01 | Tombol Ship menjalankan build, test, dan smoke test HTTP secara berurutan | P0 | Hasil setiap langkah tersimpan dengan log lengkap |
| F7-02 | Langkah yang gagal dikirim kembali ke agent sebagai task dengan log error | P0 | Agent menerima log tanpa user menyalin apa pun |
| F7-03 | Approval manusia wajib sebelum production | P0 | Tidak ada jalur deploy production tanpa approval, termasuk lewat API |
| F7-04 | Approval bisa dilakukan dari HP (PWA) | P1 | Notifikasi Ready to ship bisa disetujui dari HP |
| F7-05 | Gerbang tambahan per project (lint, audit dependensi) lewat konfigurasi | P2 | Langkah tambahan tampil di tab Gate |

### F8. Deploy dan production

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F8-01 | Build image dengan Dockerfile bila ada, selain itu Railpack + BuildKit [S5] | P0 | Build Next.js < 5 menit di VPS 4 GB sambil 3 agent aktif |
| F8-02 | Blue-green deploy: container baru, health check, lalu pindah upstream di proxy [S7] | P0 | Nol request gagal selama deploy |
| F8-03 | Domain klien dengan SSL otomatis | P0 | Domain baru aktif dengan HTTPS tanpa langkah manual selain DNS |
| F8-04 | Environment variable dan secret production per project, terenkripsi saat disimpan | P0 | Secret tidak pernah tampil ulang dan tidak pernah tersedia untuk agent |
| F8-05 | Rollback manual ke versi sebelumnya dengan satu klik | P0 | Rollback selesai < 30 detik |
| F8-06 | Deploy ke node production terpisah yang dikelola lewat SSH | P1 | Production berjalan di server yang tidak menjalankan agent |

### F9. Watch dan Production-to-Agent loop

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F9-01 | Health check dan pemantauan error rate yang ketat selama 15 menit pertama setelah deploy | P0 | Status Watch terlihat di loop strip selama jendela itu |
| F9-02 | Rollback otomatis bila error rate atau health check melewati ambang yang dikonfigurasi | P0 | Rollback berjalan tanpa campur tangan user dan tercatat dengan alasannya |
| F9-03 | Setelah rollback otomatis, buka sesi agent baru berisi diff yang di-deploy, prompt asli, log, dan stack trace | P0 | Sesi baru muncul di kolom Needs you dengan semua konteks itu |
| F9-04 | Pemantauan dasar setelah jendela 15 menit (uptime dan error rate) | P1 | Overview project menampilkan grafik 24 jam dengan penanda deploy |

### F10. Server

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F10-01 | Mode all-in-one (agent dan production di satu VPS) | P0 | Alur utama berjalan penuh di satu VPS 4 GB |
| F10-02 | Mode dua node: node agent dan node production | P1 | Agent tidak punya akses jaringan atau file ke node production |
| F10-03 | Daftar server dengan status koneksi, CPU, RAM, disk, jumlah worker, dan jumlah app | P0 | Server yang terputus terlihat dalam 30 detik |
| F10-04 | Batas resource per agent dan per build dengan cgroups v2 [S6] | P0 | Latensi p95 aplikasi production naik < 10% saat 3 agent dan 1 build aktif |
| F10-05 | Batas jumlah worker aktif per server dengan antrean; worker lebih mulai saat slot kosong (ADR 0008) | P0 | Dengan batas 4, worker ke-5 menunggu antrean dan mulai otomatis saat satu worker selesai |

### F11. Settings

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F11-01 | Appearance: tema (gelap, terang, ikuti sistem) | P0 | Tema ikut preferensi OS saat pertama dibuka |
| F11-02 | Appearance: font UI pilihan user (Plus Jakarta Sans, SF Pro Display, Sans-serif) sesuai `DESIGN_SYSTEM.md` §3.1 | P0 | Ganti font langsung berlaku tanpa reload; SF nonaktif di Windows dan Linux dengan alasannya |
| F11-03 | Agents: status login per agent per server | P0 | Status "Not signed in" mengarahkan ke terminal untuk login |
| F11-04 | Domains: token Cloudflare, domain preview, domain klien | P0 | Token divalidasi sebelum disimpan dan hanya 4 karakter terakhirnya yang tampil |
| F11-05 | Notifications: pilih kejadian yang dikirim sebagai notifikasi | P1 | Pilihan tersimpan per perangkat |
| F11-06 | Bahasa UI: Inggris (default) dan Indonesia | P1 | Semua teks UI tersedia di kedua bahasa |

### F12. Keamanan dan guardrail

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F12-01 | Setiap agent berjalan sebagai user Linux terpisah atau container rootless, tanpa akses ke secret production dan Docker socket [S3] | P0 | Prompt "baca .env production" dan "jalankan docker ps" gagal pada uji red-team di setiap rilis |
| F12-02 | Dashboard daemon hanya bisa diakses lewat jaringan privat (`tsnet`) atau tunnel SSH; yang publik hanya preview dan production [S9] | P0 | Port selain 80/443 tertutup dari internet |
| F12-03 | Indikator "skip permissions" selalu terlihat di header terminal agent | P0 | Tidak bisa disembunyikan |
| F12-04 | Konfirmasi untuk aksi berisiko sesuai `DESIGN_RULES.md` §7.2 | P0 | Ship, rollback manual, dan hapus server tidak pernah terjadi tanpa konfirmasi |
| F12-05 | Sakelar capability per workspace untuk tool agent; default aman; berlaku untuk sesi berikutnya; aksi yang tidak bisa dibalik tetap butuh approval manusia walau sakelar hidup (ADR 0008) | P1 | Mematikan sakelar menolak panggilan berikutnya; deploy, hapus data, dan perintah SSH tidak berjalan tanpa approval |

### F13. Provenance dan riwayat

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F13-01 | Setiap deployment tercatat dengan prompt, agent, worker, diff, hasil gate, dan siapa yang approve | P0 | Dari versi mana pun di production, bisa ditelusuri ke prompt asalnya |
| F13-02 | Ringkasan perubahan untuk klien dalam bahasa non-teknis, dibuat dari provenance | P1 | Ringkasan bisa disalin atau dibagikan sebagai link |
| F13-03 | Ekspor riwayat deployment per project (CSV atau JSON) | P2 | File berisi semua kolom F13-01 |

### F14. Agent, provider, dan pemakaian (ADR 0008)

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F14-01 | Definisi agent: nama, role, CLI, model, effort, system prompt, fallback; disimpan daemon | P1 | Agent yang dibuat bisa dipilih saat membuat worker; role bebas dengan preset |
| F14-02 | Provider API OpenAI-compatible (OpenRouter, OmniRoute, 9Router, Custom) dengan agent loop di daemon dan tool host di unit sesi; key terenkripsi di server, tidak pernah ke proses agent; label bahwa mode ini memakai kredit penyedia | P1 | Run dengan server palsu dan satu run nyata ber-key kecil selesai; key tidak terlihat di `/proc` milik agent; tool tidak bisa keluar dari worktree |
| F14-03 | Fallback agent saat limit rate atau sesi (CLI atau API), diputuskan daemon | P1 | Saat agent utama kena limit, run pindah ke fallback dan tercatat dengan alasannya |
| F14-04 | Pemakaian token dan biaya per worker dan per giliran, dari transcript CLI dan dari API; biaya hasil hitung ditandai estimasi | P1 | Layar pemakaian menampilkan hari ini, minggu, bulan, dan per model; angka tanpa data tampil `—` |
| F14-05 | Server MCP di daemon untuk tool IDE (baca dan tulis bergerbang) lewat socket Unix per workspace | P2 | Agent di workspace lain tidak bisa memanggil tool workspace ini |
| F14-06 | Alur spec-driven: Requirements, Design, Tasks dengan drafter, critic, reviser; kartu dengan Approve, Request changes, View, Regenerate, Discard | P2 | Batas waktu dan token per tahap, dan tombol Batalkan yang terlihat; tahap tidak berputar tanpa selesai |

### F15. Konteks project (ADR 0008)

| ID | Requirement | Prio | Kriteria penerimaan |
| --- | --- | --- | --- |
| F15-01 | Knowledge, Rules, dan Memory per workspace dengan "selalu sertakan" atau Pin; daemon menyuntikkan lewat berkas terkelola di worktree | P1 | Item yang di-pin muncul di berkas terkelola; item mati tidak |
| F15-02 | Usulan agent (rule, memori, Knowledge) menunggu persetujuan manusia; ekstraksi otomatis hanya untuk giliran substansial dengan batas biaya | P2 | Usulan tidak berlaku sebelum disetujui; pemicu ekstraksi terlihat di UI |
| F15-03 | Wiki: halaman bernarasi, graf berlapis dengan asal relasi, tur baca; dibangun atas permintaan dengan estimasi biaya, progres per fase, dan Batalkan | P2 | Build menampilkan estimasi sebelum mulai dan bisa dibatalkan |

## 6. Requirement non-fungsional

| ID | Area | Requirement | Target |
| --- | --- | --- | --- |
| N-01 | Performa | Ketikan di terminal agent dari HP atau laptop | < 100 ms |
| N-02 | Performa | Kejadian agent tampil di UI | < 1 detik |
| N-03 | Performa | Dev server hidup sampai URL preview aktif | < 30 detik |
| N-04 | Performa | Klik Ship sampai production live (Next.js, VPS 4 GB) | < 5 menit |
| N-05 | Keandalan | Waktu rollback | < 30 detik |
| N-06 | Keandalan | Request gagal selama deploy | 0 |
| N-07 | Keandalan | Sesi agent selamat saat daemon restart atau update | 100% sesi tersambung kembali |
| N-08 | Keandalan | Deploy sukses pada 10 aplikasi sampel | ≥ 9 dari 10 |
| N-09 | Resource | RAM daemon saat idle (tanpa agent dan app) | Ditetapkan setelah spike S1, lalu dijaga sebagai anggaran |
| N-10 | Resource | Ukuran binary daemon dan bundle frontend | Naik maksimal 5% per rilis tanpa ADR |
| N-11 | Keamanan | Kredensial agent, secret production, dan token Cloudflare | Tidak pernah tersedia untuk agent, tidak pernah ditulis ke log |
| N-12 | Keamanan | Komunikasi desktop app ↔ daemon | Terenkripsi (SSH tunnel atau tsnet), terautentikasi dengan token pairing |
| N-13 | Keamanan | Rilis | Ditandatangani, dengan checksum SHA-256 yang dipublikasikan |
| N-14 | Kompatibilitas | Desktop app | macOS 13+ (Apple silicon dan Intel), Windows 10/11 (x64, arm64), Ubuntu 22.04+ dan Fedora 39+ |
| N-15 | Kompatibilitas | Daemon | Ubuntu 22.04/24.04 dan Debian 12, linux/amd64 dan linux/arm64 |
| N-16 | Kompatibilitas | Versi | Daemon mendukung desktop app versi minor yang sama dan satu versi sebelumnya |
| N-17 | Aksesibilitas | Semua layar | Lulus checklist `DESIGN_RULES.md` §6 dan inspektur taste-skill 0 FAIL |
| N-18 | Privasi | Telemetri | Mati secara default, opt-in, anonim, tanpa kode, prompt, atau isi percakapan agent |
| N-19 | Kualitas kode | Codebase | Lulus `make check` sesuai bagian 10 dokumen riset |

Spesifikasi minimal VPS untuk mode all-in-one: 2 vCPU, 4 GB RAM, 40 GB disk (divalidasi ulang setelah spike S5 dan S6).

## 7. Arsitektur teknis

Semua client bicara ke satu daemon Go di VPS. Daemon menyimpan **fakta (bukan status)**, menurunkan status saat dibaca, dan mengendalikan agent, preview, build, serta deploy. Struktur package dan aturan dependensi mengikuti bagian 10 dokumen riset.

```
 Desktop app (Wails)        Web UI dan PWA        CLI
          \                      |               /
           └──────── HTTPS + WebSocket ─────────┘
┌──────────────────── Daemon: Go single binary di VPS ────────────────────┐
│ api: HTTP + WebSocket untuk semua client                                │
│ agent (PTY, hooks)   workspace (git worktree)   preview (port dan URL)  │
│ feedback (overlay klien)   gate (build dan test)   deploy (blue-green)  │
│ watch (health, rollback)   proxy (Caddy dan TLS)   node (SSH ke server) │
│ store (SQLite, fakta)                                                   │
└─────────────────────────────────────────────────────────────────────────┘
   Agent CLI per user Linux      Docker: preview dan app     Node production via SSH
```

### Stack

| Lapisan | Teknologi |
| --- | --- |
| Daemon | Go, satu binary, SQLite pure Go (`modernc.org/sqlite`) |
| Terminal agent | PTY (`creack/pty`) dengan sesi di host terpisah agar selamat saat daemon restart; tmux sebagai fallback |
| Proxy dan TLS | Caddy/CertMagic yang di-embed, sertifikat wildcard via DNS-01 Cloudflare |
| Build dan runtime app | Docker Engine, Dockerfile atau Railpack + BuildKit |
| Jaringan privat | tsnet (Tailscale) di-embed, atau tunnel SSH |
| Desktop app | Wails (Go), frontend web yang sama dengan web UI |
| Frontend | Satu framework (diputuskan di ADR pertama), token dari `DESIGN_SYSTEM.md` |

### Model data (fakta yang disimpan)

| Entitas | Isi utama | Relasi |
| --- | --- | --- |
| User | Nama, preferensi (tema, font, bahasa) | Punya banyak Workspace |
| Workspace | Klien atau ruang kerja | Punya Project, Domain |
| Server | Host, peran (all-in-one, agent, production), kunci SSH, status terakhir | Menjalankan Worker dan App |
| Project | Repo, stack terdeteksi, perintah dev/build/test | Punya Worker dan Deployment |
| Worker | Task, agent, branch, worktree, waktu dibuat | Punya AgentSession, Preview, GateRun |
| AgentSession | Agent, mode (TUI atau SDK), aktivitas terakhir, generasi runtime | Milik Worker |
| Preview | URL, port, link berbagi, kedaluwarsa | Milik Worker |
| FeedbackComment | Teks, screenshot, selector elemen, status | Milik Preview, bisa menjadi task |
| GateRun | Langkah, hasil, log, durasi | Milik Worker |
| Deployment | Image, versi, domain, approver, waktu live, alasan rollback | Milik Project, menunjuk ke Worker dan GateRun asalnya |
| Incident | Ambang yang terlewati, metrik, aksi (rollback), sesi agent yang dibuka | Milik Deployment |
| Domain | Nama, provider DNS, sertifikat | Milik Workspace |
| AuditEvent | Siapa, apa, kapan, objek | Semua aksi berisiko |

Secret (env production, token Cloudflare) disimpan terenkripsi dan direferensikan, tidak pernah disalin ke tabel lain.

### Status yang diturunkan

Status worker (Working, Needs you, Preview live, Waiting on client, Gate failed, Ready to ship, Shipping, Live, Rolled back) dihitung saat dibaca dari fakta: aktivitas AgentSession, Preview, GateRun, dan Deployment terakhir. Daemon mengirim perubahan fakta ke client lewat WebSocket; client menghitung tampilan dari kosakata yang sama.

### API (v1)

| Endpoint | Fungsi |
| --- | --- |
| `GET /api/v1/events` (WebSocket) | Aliran perubahan fakta untuk semua client |
| `GET/POST /api/v1/workspaces`, `/projects`, `/servers` | Kelola objek dasar |
| `POST /api/v1/projects/{id}/workers` | Buat worker baru |
| `GET /api/v1/sessions/{id}/pty` (WebSocket) | Terminal agent dua arah |
| `POST /api/v1/workers/{id}/messages` | Kirim pesan atau task ke agent |
| `POST /api/v1/workers/{id}/previews/{pid}/shares` | Buat atau cabut link berbagi |
| `POST /api/v1/workers/{id}/ship` | Mulai Ship Gate |
| `POST /api/v1/gate-runs/{id}/approve` | Approval sebelum production |
| `POST /api/v1/deployments/{id}/rollback` | Rollback manual |
| `POST /public/v1/feedback/{share}` | Komentar klien dari overlay (tanpa login, dibatasi rate) |

Semua endpoint selain `/public` butuh token sesi hasil pairing. Endpoint yang mengubah production mencatat AuditEvent.

## 8. Instalasi, distribusi, update, dan uninstall

User meng-install desktop app di laptopnya, lalu app memasang daemon di VPS lewat SSH. Untuk user tanpa desktop app, daemon bisa dipasang langsung dengan satu perintah di VPS.

### 8.1 Format file per OS

| Platform | Format utama | Format tambahan | Arsitektur | Agar tidak diblokir OS |
| --- | --- | --- | --- | --- |
| macOS | `.dmg` berisi `.app` | `.zip` untuk auto-update; Homebrew cask (fase 2) | Universal (Apple silicon + Intel) | Ditandatangani Developer ID Apple dan dinotarisasi (Gatekeeper) |
| Windows | `.exe` installer (NSIS, dibangun langsung oleh Wails) | `.msi` untuk perusahaan; winget (fase 2) | x64 dan arm64 | Ditandatangani Authenticode (SmartScreen) |
| Linux (desktop) | `.AppImage` | `.deb` (Ubuntu/Debian), `.rpm` (Fedora/RHEL) | x64 dan arm64 | Paket `.deb`/`.rpm` menyatakan dependensi WebKitGTK yang dibutuhkan Wails |
| Daemon di VPS | Binary tunggal dalam `.tar.gz` | Skrip install satu perintah yang memasang daemon sebagai service systemd | linux/amd64 dan linux/arm64 | Checksum SHA-256 dan tanda tangan rilis dipublikasikan |
| HP | Tidak ada app store | PWA dari web UI daemon | — | — |

Semua file dibangun otomatis oleh CI dari satu tag rilis dan dipublikasikan di GitHub Releases. Desktop app dan daemon memakai nomor versi yang sama.

**Nama:** desktop app `Astra`, CLI `astra`, daemon `astrad`, service systemd `astrad.service`, file installer `Astra-<versi>-<os>-<arsitektur>` (misalnya `Astra-1.0.0-macos-universal.dmg`).

### 8.2 Kebutuhan sebelum install

| Kebutuhan | Minimal |
| --- | --- |
| VPS | Ubuntu 22.04/24.04 atau Debian 12, 2 vCPU, 4 GB RAM, 40 GB disk, akses root atau sudo |
| Jaringan | Port 80 dan 443 bebas; port SSH bisa dijangkau dari laptop |
| Domain | Satu domain di Cloudflare untuk preview, ditambah domain klien untuk production |
| Akun agent | Akun Claude (Pro/Max) atau API key, dan/atau akun ChatGPT untuk Codex |

### 8.3 Install lewat desktop app (jalur utama)

1. Unduh installer untuk OS kamu dari halaman rilis, lalu buka.
2. Tambah server: isi host, user, dan pilih SSH key yang ada atau buat key baru. App menampilkan public key yang perlu ditambahkan ke VPS bila key baru.
3. Pemeriksaan kesiapan: app mengecek OS, arsitektur, RAM, disk, port 80/443, systemd, dan sudo. Setiap masalah ditampilkan dengan cara memperbaikinya.
4. Install daemon: app menampilkan apa yang akan dipasang, lalu setelah user setuju: mengunduh binary yang sesuai arsitektur dan memverifikasi checksum-nya; membuat user sistem untuk daemon; memasang Docker Engine bila belum ada; memasang daemon sebagai service systemd lalu menjalankannya.
5. Pairing: app mengirim token sekali pakai lewat SSH, lalu tersambung ke daemon lewat tunnel terenkripsi.
6. Domain: masukkan token API Cloudflare dengan izin edit DNS untuk satu zona; daemon menyiapkan wildcard preview.
7. Login agent: buka terminal bawaan, jalankan `claude` atau `codex`, lalu login dengan akun sendiri.
8. Project pertama: tambahkan repo, lalu buat task pertama.

### 8.4 Install langsung di VPS (tanpa desktop app)

Satu perintah di VPS mengunduh skrip install. Skrip menampilkan semua langkah yang akan dijalankan dan meminta konfirmasi sebelum mengubah sistem, menjalankan langkah yang sama dengan 8.3 poin 4, lalu menampilkan kode pairing. User membuka web UI dari browser atau HP dan memasukkan kode itu. Isi skrip tersedia untuk dibaca sebelum dijalankan.

### 8.5 Update

| Bagian | Cara update | Rollback |
| --- | --- | --- |
| Desktop app | Cek update otomatis; file update diverifikasi tanda tangannya sebelum dipasang | Install ulang versi sebelumnya dari halaman rilis |
| Daemon | Satu klik dari desktop app atau perintah CLI di VPS; sesi agent yang berjalan tidak terputus [S1] | Binary versi sebelumnya disimpan; satu perintah untuk kembali |

Desktop app menolak tersambung ke daemon yang terlalu tua dan menawarkan update daemon.

### 8.6 Uninstall

- **Desktop app:** cara biasa di setiap OS. Data lokal hanya cache dan preferensi.
- **Daemon:** perintah uninstall dengan dua pilihan. Default menghentikan daemon dan menyimpan data. Opsi purge menghapus data daemon, worktree, dan image build. Aplikasi production klien tidak pernah dihapus kecuali user meminta secara eksplisit dengan opsi terpisah.

## 9. UX dan design

Semua layar mengikuti `docs/design/DESIGN_RULES.md` (perilaku, larangan, aksesibilitas) dan `docs/design/DESIGN_SYSTEM.md` (token, komponen, pola layar). Arah visual: gaya `data-terminal` dari taste-skill untuk tema gelap (default) dan palet `framed-sheet` untuk tema terang, dengan loop strip sebagai elemen khas produk.

### Layar MVP

| Layar | Pertanyaan layar | Aksi utama | Requirement |
| --- | --- | --- | --- |
| Onboarding | Apa yang perlu disiapkan agar bisa mulai? | Add server | F1-02 s.d. F1-07 |
| Board | Apa yang butuh saya sekarang? | New task | F4-01, F4-02 |
| Detail worker | Sampai mana task ini, dan apa langkah berikutnya? | Sesuai tahap: Share preview, Ship, View failed check, Reply | F3, F4-03, F5, F7 |
| Overview project | Apakah production klien ini sehat, dan versi mana yang live? | Ship, atau View logs bila gagal | F8, F9, F13 |
| Servers | Server mana yang penuh atau tidak tersambung? | Add server | F10 |
| Settings | Apa yang perlu saya atur atau perbaiki? | Save per section | F11 |
| Overlay klien | Bagaimana saya memberi masukan? | Leave feedback | F6 |

### Aturan yang paling sering dilupakan

- Satu tombol utama per layar; seleksi ditandai garis aksen, bukan pill berwarna.
- Warna hanya untuk status dan data, di bawah 3% piksel layar.
- Setiap layar dicek di 1440, 768, dan 390px dengan ketiga opsi font.
- Semua state (kosong, memuat, error, offline) dibuat bersama layarnya, bukan belakangan.
- Overlay klien memakai bahasa awam dan selalu Plus Jakarta Sans.

## 10. Kepatuhan kebijakan pihak ketiga

Aturan ini adalah batas keras. Kalau sebuah requirement bertentangan dengan salah satunya, requirement yang diubah.

| Pihak | Aturan | Dampak ke produk |
| --- | --- | --- |
| Anthropic (hukum dan kepatuhan Claude Code) | Binary Claude Code tidak dimodifikasi; user login dengan akunnya sendiri lewat alur Anthropic; produk tidak mengumpulkan, menyimpan, atau menengahi kredensial; nama "Claude Code" tidak dipakai sebagai nama produk | F3-02, F3-03; nama produk tanpa kata Claude |
| Anthropic (Agent SDK dengan paket Claude) | Agent SDK dan `claude -p` memakai kredit bulanan terpisah; sesi interaktif memakai limit subscription | Sesi utama selalu interaktif (F3-02); mode SDK hanya opsi berlabel (F3-08). Dicek ulang sebelum setiap rilis besar |
| Anthropic dan OpenAI | Satu akun subscription untuk satu orang | Di fase tim, setiap anggota login dengan akunnya sendiri |
| Apple | SF Pro hanya untuk produk di platform Apple | SF tidak pernah di-bundle; opsi font SF hanya aktif di macOS dan iOS (F11-02) |
| Apple dan Microsoft | Gatekeeper dan SmartScreen memblokir app tanpa tanda tangan | Butuh akun Apple Developer dan sertifikat code signing Windows sebelum rilis publik |
| Lisensi open source | Orchestrator.inc (Apache-2.0), Plus Jakarta Sans (OFL), Lucide (ISC), taste-skill dan ui-ux-pro-max (MIT) | Atribusi dan file lisensi dipertahankan; produk sendiri berlisensi Apache-2.0 |
| Let's Encrypt | Batas penerbitan sertifikat per domain | Preview memakai satu sertifikat wildcard, bukan sertifikat per worker |

## 11. Rencana rilis

Pekerjaan dibagi menjadi enam milestone yang dipisahkan empat gerbang. Setiap gerbang adalah keputusan lanjut atau berhenti yang diambil pemilik produk setelah QA manual; kalau gerbang gagal, arsitektur atau scope direvisi lewat ADR sebelum milestone berikutnya dimulai.

| Milestone | Isi | Keluar / catatan |
| --- | --- | --- |
| **M0 Spike** — prototype risiko terbesar | S1–S10: PTY, chat view, isolasi, preview TLS, build, cgroups, blue-green, overlay, tsnet, desktop tiga OS | — |
| ◆ Gerbang 1 | S1 sampai S3 lulus, konsep layak | |
| **M1 Inti** — daemon, agent, board | F1 onboarding, F2 project, F3 runtime agent, F4 board | 5 sesi agent stabil 2 jam, kartu board pindah otomatis |
| **Fase 0** (ADR 0008, 0009) | Shell dan token baru; batas worker aktif (F10-05); pencatatan pemakaian dari transcript (F14-04) | Shell tiga panel lulus checklist `DESIGN_RULES.md` §13 |
| **M1b Agent dan provider** (ADR 0008) | F14-01 sampai F14-03, layar Settings Agents, Providers, Token Usage | Run API dengan fallback lintas CLI dan API |
| ◆ Gerbang 2 | S4 dan S5 lulus, fitur unggulan layak | |
| **M2 Preview dan feedback** — Show | F5 preview, F6 feedback klien | Preview aktif < 10 detik, overlay jalan di 3 sampel |
| **M3 Ship dan watch** — Ship, Watch | F7 gate, F8 deploy, F9 watch, F12 keamanan, F13-01 provenance | 9 dari 10 sampel ter-deploy, rollback < 30 detik |
| ◆ Gerbang 3 | Alur end-to-end lulus di minimal 9 dari 10 sampel | |
| **M1c Konteks dan tugas** (ADR 0008) | F4-08, F15-01, F4-07 | Item Pin tersuntik; History tugas lengkap |
| **Setelah M3** (ADR 0008) | F14-05, F14-06, F15-02, F15-03 | Spec dan Wiki dengan batas biaya |
| **M4 Private beta** — desktop tiga OS | F1-01 installer bertanda tangan, F10 server, F11 settings | Diuji 5 developer yang diwawancarai |
| ◆ Gerbang 4 | Semua metrik produk tercapai, nol temuan keamanan terbuka | |
| **M5 Rilis publik** — GitHub Releases | Semua format file di bagian 8.1, dokumentasi install | Metrik adopsi diukur selama 8 minggu |

Milestone tidak diberi tanggal karena kecepatan pengerjaan Claude Code baru bisa diukur setelah M0. Setelah M0 selesai, setiap milestone diberi target tanggal berdasarkan kecepatan nyata di M0.

**Definisi selesai untuk setiap requirement:** kriteria penerimaannya lulus saat QA manual, `make check` lulus, layar terkait lulus checklist `DESIGN_RULES.md` §13, ID requirement tercatat di ringkasan sesi Claude Code, dan commit-nya mengikuti aturan git di bagian 12.

## 12. Aturan git untuk Claude Code

Teks lengkap ada di [`docs/git-rules.md`](git-rules.md) dan dimuat lewat `CLAUDE.md`.

## 13. Metrik sukses

Metrik produk menjadi syarat Gerbang 4 dan diuji sebelum rilis publik. Metrik adopsi diukur selama 8 minggu setelah rilis dari heartbeat anonim yang opt-in.

### Metrik produk

| Metrik | Target | Requirement |
| --- | --- | --- |
| Install sampai preview pertama, di VPS baru | < 15 menit (median) | F1, F5 |
| Dev server hidup sampai URL preview aktif | < 30 detik | F5-02 |
| Klik Ship sampai production live (Next.js, VPS 4 GB) | < 5 menit | F7, F8 |
| Waktu rollback | < 30 detik | F8-05, F9-02 |
| Request gagal selama deploy | 0 | F8-02 |
| Deploy sukses pada 10 aplikasi sampel (Next.js, Astro, Vite, Laravel, Express, FastAPI, Go, Django, static HTML, Dockerfile kustom) | Minimal 9 dari 10 | F2-03, F8 |
| Uji red-team akses secret production oleh agent | Gagal 100% di setiap rilis | F12-01 |

### Metrik adopsi (8 minggu setelah rilis publik)

| Metrik | Target |
| --- | --- |
| Instalasi aktif mingguan | 25 |
| User dengan minimal 1 project klien di production lewat Astra | 10 |
| Preview yang dibagikan ke klien, dari semua preview | Minimal 30% |
| GitHub stars | 300 |
| Testimoni atau studi kasus dari freelancer | 3 |

## 14. Risiko dan mitigasi

| Risiko | Dampak | Mitigasi |
| --- | --- | --- |
| Anthropic atau OpenAI mengubah aturan pemakaian subscription di tool pihak ketiga | Runtime agent harus dirombak | Sesi selalu memakai CLI resmi yang interaktif; aturan dicek ulang sebelum setiap rilis besar; mode SDK disiapkan sebagai jalur cadangan |
| Format hooks atau transcript agent berubah | Tampilan terstruktur rusak | Adapter per agent dipisah; terminal tetap bekerja walau tampilan terstruktur gagal |
| Agent dengan skip permissions merusak server production | Kerugian data klien | Isolasi F12-01, mode dua node direkomendasikan, uji red-team setiap rilis |
| VPS kecil kehabisan RAM saat agent dan build berjalan bersamaan | Production klien melambat | Batas cgroups F10-04, spesifikasi minimal 4 GB, peringatan saat resource menipis |
| Jean (tim Coolify) atau Dokploy menambahkan loop agent-ke-deploy | Keunggulan menipis | Fokus pada Client Feedback Loop dan Production-to-Agent loop, rilis cepat ke freelancer |
| Overlay feedback rusak di framework tertentu | Klien tidak bisa memberi masukan | Fallback berbasis screenshot manual (S8) |
| Biaya sertifikat code signing dan akun Apple Developer | Rilis publik tertunda | Disiapkan sejak M4; private beta boleh memakai build tanpa tanda tangan dengan petunjuk install |
| Kode yang ditulis Claude Code membengkak atau tidak konsisten | Sulit dirawat | Aturan anti-bloat dan `make check` (bagian 10 dokumen riset), sesi refactor terjadwal |
| Status pemakaian kredit Agent SDK masih berubah | Pengalaman user tidak bisa diprediksi | Jangan jadikan mode SDK default; tampilkan label kredit dengan jelas |

## 15. Pertanyaan terbuka

| Pertanyaan | Kapan harus diputuskan | Pemilik |
| --- | --- | --- |
| Domain resmi untuk Astra, serta pengecekan merek dan ketersediaan nama (GitHub, package manager, Homebrew, winget) | Sebelum M1, karena dipakai di nama binary, package, dan installer | Maulana |
| Framework frontend untuk web UI dan Wails | ADR pertama di M0 | Maulana, rekomendasi dari Claude Code |
| Model bisnis: open source penuh atau ada fitur berbayar (tim, relay terkelola) | Sebelum M5 | Maulana |
| Apakah loki-mode dan NXTG-Forge tumpang tindih dengan fitur unggulan | Sebelum M2 | Maulana |
| Ambang default rollback otomatis (error rate, health check gagal berapa kali) | Akhir M3, dari data uji 10 aplikasi sampel | Maulana |
| Apakah komentar klien langsung menjadi task, atau menunggu persetujuan developer, sebagai default | Akhir M2, setelah diuji dengan klien nyata | Maulana |
| Apakah key gateway boleh disuntikkan ke CLI agent (ADR 0008 §4, B2); usulan: tidak di MVP | Sebelum M1b | Maulana |
| Apakah Knowledge, Rules, dan Memory berlaku per workspace (usulan) atau per project | Sebelum M1c | Maulana |
| Hasil wawancara 5–10 developer persona target | Paralel dengan M0, wajib selesai sebelum M4 | Maulana |

## 16. Glosarium

| Istilah | Arti |
| --- | --- |
| Daemon | Program Go yang berjalan terus di VPS dan memegang semua state |
| Worker | Satu task dengan satu agent dan satu worktree |
| Worktree | Salinan kerja git terpisah dengan branch sendiri, sehingga worker tidak saling menimpa |
| Preview | Aplikasi dari worktree yang berjalan dengan URL HTTPS sendiri |
| Overlay klien | Lapisan di atas halaman preview tempat klien menulis komentar |
| Ship Gate | Rangkaian build, test, dan smoke test sebelum production, ditutup dengan approval manusia |
| Blue-green deploy | Versi baru dijalankan berdampingan dengan versi lama, lalu lalu lintas dipindah setelah sehat |
| Watch | Pemantauan ketat 15 menit pertama setelah deploy |
| Production-to-Agent loop | Masalah production otomatis menjadi sesi agent dengan konteks lengkap |
| Provenance | Catatan asal-usul setiap versi di production: prompt, agent, diff, gate, approver |
| Loop strip | Elemen UI yang menampilkan tahap worker dari Prompt sampai Watch |
| Node | Server yang dikelola daemon: agent, production, atau keduanya |
| Spike | Prototype kecil untuk menguji risiko teknis sebelum membangun fitur |
| Gerbang | Keputusan lanjut atau berhenti di antara milestone |
| ADR | Architecture Decision Record, catatan singkat setiap perubahan keputusan arsitektur |
