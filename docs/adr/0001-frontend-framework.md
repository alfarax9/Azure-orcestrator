# ADR 0001: Framework frontend untuk web UI dan desktop app

| Hal | Isi |
| --- | --- |
| Status | **Diterima** (2026-10-25): React + TypeScript + Vite |
| Tanggal | 2026-10-25 |
| Pemutus | Maulana Alfara Salim |
| Pengusul | Claude Code |
| Milestone | M0 (PRD §15: "ADR pertama di M0") |
| Terkait | PRD §7 (Stack), F4, F11-02, N-01, N-02, N-10, N-17; `DESIGN_RULES.md` §6, §8, §9; `DESIGN_SYSTEM.md` §8 |

## Konteks

Astra punya satu frontend yang dipakai di tiga tempat: desktop app (Wails, WebView OS), web UI yang dilayani daemon, dan PWA di HP (PRD §7). Frontend di-embed ke binary Go, jadi hasilnya harus SPA statis tanpa server Node saat runtime.

Kebutuhan yang menentukan pilihan:

1. **Aksesibilitas berat** (`DESIGN_RULES.md` §6, N-17): dialog dengan focus trap, command palette (combobox), tabs, radio group, tooltip yang bisa dijangkau keyboard, satu region `aria-live`, alternatif keyboard untuk drag Kanban. Semua harus lulus `inspect.mjs --app` 0 FAIL.
2. **Tanpa gaya pihak ketiga** (`DESIGN_RULES.md` §9): komponen hanya memakai token; library headless boleh lewat ADR. Jadi yang dicari adalah perilaku dan a11y siap pakai, bukan tampilan.
3. **Data real-time**: daemon mengirim perubahan fakta lewat WebSocket, client menurunkan status worker dari kosakata `DESIGN_SYSTEM.md` §10 (PRD §7). Kejadian tampil < 1 detik (N-02).
4. **Terminal**: banyak terminal xterm.js sekaligus, render WebGL di atas 16 terminal (`DESIGN_SYSTEM.md` §8.10), ketikan < 100 ms (N-01).
5. **List panjang**: virtualisasi di atas 50 baris (§8.6). **Motion**: kartu Board pindah kolom sebagai shared element (§8).
6. **Anggaran bundle**: naik maksimal 5% per rilis tanpa ADR (N-10), dicek di `make check`.
7. **Gerbang kualitas** (`docs/git-rules.md` §12.4): `npm run lint`, `npx knip`, `npm test`, `npm run build`; inspektur taste-skill di `http://localhost:5173`, yaitu port default Vite.
8. **Implementer adalah Claude Code.** Kode yang konsisten dan benar sejak percobaan pertama lebih berharga daripada beberapa KB bundle.

## Opsi yang dipertimbangkan

| Kriteria | React + Vite | Svelte 5 + Vite | SolidJS + Vite | Vue 3 + Vite |
| --- | --- | --- | --- | --- |
| Headless a11y (dialog, combobox, tabs, tooltip, drag dengan keyboard) | **Terkuat**: React Aria Components (Adobe), diuji dengan screen reader, termasuk drag-and-drop yang punya alternatif keyboard | Bits UI / Melt UI; cukup, belum sematang React Aria | Kobalte; cakupan lebih kecil | Reka UI (Radix Vue); cukup |
| Data real-time dan status turunan | Store eksternal + `useSyncExternalStore`; perlu disiplin memoization | **Runes (`$derived`)** cocok sekali dengan "fakta → status" | **Signals**, paling efisien | `computed`, cocok |
| xterm.js + WebGL | Framework-agnostic, dibungkus komponen tipis | Sama | Sama | Sama |
| Virtualisasi | TanStack Virtual (matang) | TanStack Virtual (adapter) | TanStack Virtual (adapter) | TanStack Virtual (adapter) |
| Ukuran runtime (gzip, kira-kira) | ~45 KB | ~5–15 KB | ~7 KB | ~35 KB |
| Template Wails resmi | Ya | Ya | Tidak (komunitas) | Ya |
| Keandalan kode yang ditulis Claude Code | **Tertinggi**: data latih dan pola terbanyak | Tinggi, tapi runes (Svelte 5) masih sering tercampur sintaks Svelte 4 | Sedang; sering salah kaprah sebagai "React" | Tinggi |
| Lucide | `lucide-react` | `@lucide/svelte` | `lucide-solid` | `lucide-vue-next` |

Opsi yang disingkirkan:

- **Next.js** (dan Nuxt, SvelteKit mode server). Next.js bisa diekspor statis (`output: 'export'`), tetapi mode itu mematikan SSR, API routes, middleware, dan optimasi gambar, sehingga yang tersisa hanya React dengan toolchain lebih berat. Alasan lain: tidak ada kebutuhan SSR atau SEO (dashboard di balik pairing; halaman publik klien dilayani Go); menjalankan server Node di VPS menambah proses, RAM (F10-04, N-09), dan permukaan serangan (F12-02); template Wails resmi dan inspektur taste-skill (`localhost:5173`) memakai Vite. Next.js tetap layak untuk landing page atau situs dokumentasi, sebagai proyek terpisah dari app.
- **Library komponen bergaya** (MUI, Chakra, shadcn dengan tema bawaan, Vuetify), karena dilarang `DESIGN_RULES.md` §9.

## Rekomendasi

**React + TypeScript + Vite, SPA statis**, dengan pustaka pendukung yang sengaja sedikit:

| Kebutuhan | Pilihan | Alasan |
| --- | --- | --- |
| Komponen headless | React Aria Components | Seluruh daftar a11y §6 tersedia tanpa gaya; drag-and-drop Kanban sudah punya alternatif keyboard. Ini sekaligus ADR yang diwajibkan `DESIGN_RULES.md` §9 untuk komponen headless |
| State dari WebSocket | Store fakta milik sendiri (modul TS kecil) + `useSyncExternalStore`, status diturunkan oleh satu fungsi murni `deriveWorkerStatus(facts)` | Fungsi yang sama dipakai daemon dan client; mudah diuji tanpa UI |
| Routing | Router kecil berbasis hash atau TanStack Router | Wails dan PWA sama-sama butuh routing tanpa server |
| Terminal | `@xterm/xterm` + addon WebGL dan fit | Diwajibkan §8.10 |
| Virtualisasi | TanStack Virtual | §8.6 |
| Ikon | `lucide-react` | Satu set ikon (§9) |
| Gaya | CSS biasa dengan `web/src/styles/tokens.css`, tanpa CSS-in-JS | Token jadi satu-satunya sumber nilai, gampang di-lint |
| Test | Vitest + Testing Library | Satu toolchain dengan Vite |

**Kenapa React, bukan Svelte 5 yang lebih kecil:** risiko terbesar frontend Astra adalah aksesibilitas dan interaksi yang rumit (palette, dialog, drag dengan keyboard, terminal), bukan ukuran bundle. React Aria memberi semua perilaku itu dalam keadaan sudah teruji, dan Claude Code menulis React dengan kesalahan paling sedikit. Selisih ~30–40 KB gzip tidak berarti untuk app yang dimuat dari disk lokal (Wails) atau dari daemon di jaringan privat. Anggaran N-10 diukur relatif per rilis, jadi tetap terjaga.

**Opsi kedua:** Svelte 5 + Bits UI. Pilih ini bila pemilik produk menganggap ukuran bundle dan model reaktif lebih penting daripada kematangan a11y. Store fakta dan `deriveWorkerStatus` tetap sama, jadi biaya pindah di awal M1 kecil.

## Konsekuensi

**Positif**

- Checklist a11y §6 sebagian besar dipenuhi oleh komponen yang sudah teruji, bukan ditulis ulang.
- Logika status terpisah dari framework, jadi bisa diuji dan dipakai ulang di daemon.
- Toolchain tunggal (Vite, Vitest, ESLint, knip) cocok dengan gerbang `npm run lint`, `npx knip`, `npm test`, `npm run build`.

**Negatif dan cara menahannya**

- React Aria cukup besar. Import per komponen saja, dan pantau anggaran bundle di `make check` sejak commit pertama di `web/`.
- Render ulang berlebihan saat banyak fakta masuk. Pakai selector per komponen di `useSyncExternalStore`; terminal tidak pernah dirender lewat state React (xterm.js menulis langsung ke DOM-nya).
- Shared-element transition untuk Board butuh kode sendiri (FLIP dengan `transform`), karena hanya `transform` dan `opacity` yang boleh dianimasikan (§8).

## Validasi di M0

Rekomendasi ini dianggap benar bila prototype kecil di M0 lulus semua poin berikut. Kalau ada yang gagal, ADR ini direvisi sebelum M1.

1. Board dengan 5 kolom dan 30 kartu, kartu pindah kolom dengan mouse dan keyboard; `inspect.mjs --app` 0 FAIL di 1440, 768, dan 390.
2. 20 terminal xterm.js dengan WebGL terbuka bersamaan di WebView Wails macOS, Windows, dan Linux (bersama spike S10); ketikan < 100 ms.
3. 100 perubahan fakta per detik lewat WebSocket tanpa frame jatuh di Board.
4. Bundle produksi awal dicatat sebagai garis dasar anggaran N-10.

## Keputusan

Diterima oleh Maulana Alfara Salim pada 2026-10-25: **React + TypeScript + Vite** sebagai SPA statis, dengan pustaka pendukung pada tabel Rekomendasi. Svelte 5 + Bits UI tetap menjadi opsi cadangan bila validasi M0 gagal. Pemakaian React Aria Components sebagai komponen headless (syarat `DESIGN_RULES.md` §9) disetujui lewat ADR ini.
