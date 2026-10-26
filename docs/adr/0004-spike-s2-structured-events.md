# ADR 0004: Spike S2, tampilan terstruktur dari transcript agent

| Hal | Isi |
| --- | --- |
| Status | **Sebagian** (2026-10-26): Claude Code terbukti live; Codex terbukti pada rollout nyata, latensi live belum diukur |
| Pemutus | Maulana Alfara Salim |
| Pengusul | Claude Code |
| Milestone | M0, syarat Gerbang 1 |
| Terkait | F3-05, N-02, PRD §14 (risiko format agent berubah), `internal/agent`, ADR 0003 |

## Pertanyaan spike

Bisakah daemon menampilkan apa yang dilakukan agent (prompt, pesan, perintah, tool, file yang berubah beserta diff-nya) kurang dari 1 detik setelah terjadi, hanya dengan membaca file yang ditulis agent sendiri, tanpa mode headless dan tanpa mengubah agent?

## Desain yang diuji

- **Satu adapter per agent** di `internal/agent` (`ParseClaude`, `ParseCodex`), keduanya mengubah satu baris transcript menjadi `[]Event`. Model event sama untuk semua agent: `prompt`, `message`, `command` (dengan exit code bila ada), `tool` (dengan target), `file` (path, `add`/`update`/`delete`, hunk diff), dan `turn_end`.
- **Diff selalu berbentuk hunk** (`oldStart`, `oldLines`, `newStart`, `newLines`, baris berawalan ` `, `+`, `-`), dari mana pun asalnya: `structuredPatch` dan `bashEditDiff` Claude dipakai langsung, file baru dari Write atau Codex dijadikan satu hunk, dan `unified_diff` Codex diurai.
- **Adapter melewati apa yang tidak dikenalnya.** Baris yang bukan JSON, jenis record baru, atau field yang berubah tidak menghentikan tampilan; terminal tetap jalan walau tampilan terstruktur tertinggal (mitigasi PRD §14). `encoding/json/v2` mengabaikan field yang tidak dideklarasikan.
- **`Follow` membaca baris baru setiap 100 ms** dari posisi terakhir, menahan baris yang belum lengkap sampai ditutup newline.
- **Pencari transcript per folder kerja:** Claude Code di `~/.claude/projects/<slug>/<sesi>.jsonl` (slug = path fisik dengan setiap karakter selain huruf dan angka menjadi `-`); Codex di `~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl`, dicocokkan lewat `cwd` di record `session_meta`. Yang dipilih adalah file terbaru yang berubah setelah sesi dimulai.
- **Dependensi baru: tidak ada.** Hanya standard library.

## Format yang ditemukan (Claude Code 2.1.310, Codex 0.164.0)

| Event | Claude Code | Codex |
| --- | --- | --- |
| Prompt | entri `user` berisi string; bash mode tercatat sebagai `<bash-input>` | item `UserMessage` |
| Pesan agent | blok `text` di entri `assistant` | item `AgentMessage` |
| Perintah | blok `tool_use` bernama `Bash` (dicatat saat dipanggil) | item `CommandExecution` dengan `exit_code` (dicatat saat selesai) |
| File berubah | hasil tool: `structuredPatch` (Edit), `type: create` + isi (Write), `bashEditDiff` (Bash) | item `FileChange`: `add` + isi, `update` + `unified_diff`, `delete` + isi |
| Tool lain | blok `tool_use` lain (Read, Grep, …) | tidak dipakai: tool call Codex berupa potongan JavaScript yang memanggil tool internal |
| Giliran selesai | `stop_reason: end_turn` pada entri yang berisi teks | `task_complete` |

## Hasil

| Uji | Hasil |
| --- | --- |
| Transcript Claude Code nyata (uji kerja S1) | Prompt, perintah heredoc, `count.sh` +4 dan `notes.md` +5, pesan, dan akhir giliran terbaca benar |
| Transcript besar repo ini sendiri | 36 prompt dan tepat 36 akhir giliran, 245 perubahan file, 204 perintah, 277 tool; tidak ada baris yang gagal |
| Rollout Codex nyata (uji kerja S1) | Prompt, pesan, dua perintah dengan `exit 0`, `README.md` +3, pesan akhir, akhir giliran terbaca benar |
| **Latensi live Claude Code** (5 probe) | Lag dari timestamp agent sampai event terbaca 139–236 ms; dari Enter sampai event di tangan 147–230 ms. Target F3-05 dan N-02: < 1 detik |
| Golden file kedua agent | Lulus; fixture meniru struktur transcript asli dengan isi netral, termasuk baris rusak dan record yang harus dilewati |
| `Follow` | Baris setengah jadi tidak diurai; baris rusak dilewati; event tiba satu poll setelah barisnya lengkap (diuji dengan `testing/synctest`) |
| `make check` | Lulus; coverage `internal/agent` 95.7% |

Alat percobaan: `astrad events -agent claude|codex -dir DIR` mengikuti transcript sebuah folder dan mencetak setiap event beserta lag-nya.

## Temuan

1. **Codex tetap menulis rollout JSONL.** State Codex ada di SQLite (`~/.codex/state_5.sqlite`), tetapi setiap sesi juga ditulis ke `~/.codex/sessions/…/rollout-*.jsonl`. Adapter tidak perlu SQLite, jadi `modernc.org/sqlite` tidak ditarik untuk S2.
2. **Codex hanya mencatat item yang sudah selesai.** Perintah yang berjalan lama baru tampil setelah selesai; rollout tidak berisi event "mulai". Claude mencatat `tool_use` saat dipanggil dan hasilnya saat selesai. Untuk menampilkan "Codex sedang menjalankan X" secara live, perlu sumber lain (hooks Codex atau layar terminal), diputuskan di M1.
3. **Claude menulis satu pesan sebagai beberapa entri** (thinking, lalu teks), masing-masing membawa `stop_reason`. Tanpa penanganan khusus, akhir giliran tercatat ganda; adapter hanya mengakhiri giliran dari entri yang menampilkan sesuatu.
4. **Bash mode Claude Code memicu jawaban model.** Probe latensi memakai `!echo …` dengan anggapan tidak memakai kuota; ternyata Claude mengomentari setiap output, sehingga lima probe memakai lima giliran model yang sangat pendek. Probe tanpa kuota untuk Claude belum ada.
5. **Sesi dipetakan ke transcript lewat folder dan waktu.** Dua sesi agent yang sama di folder yang sama tidak bisa dibedakan. Di Astra setiap worker punya worktree sendiri (F3-01), jadi foldernya unik; aturan ini harus tetap dijaga.
6. **Hooks belum dipakai.** Transcript sudah cukup untuk Claude Code. Hooks berarti menulis konfigurasi agent ke setiap worktree, jadi hanya dipakai bila ada kebutuhan yang tidak bisa dipenuhi transcript (temuan 2), dengan ADR.
7. **Format tidak didokumentasikan.** Fixture dan golden file per versi agent menjadi alarm dini: versi agent baru diuji terhadap fixture lamanya, dan fixture baru ditambahkan dari transcript versi itu (dengan isi netral, tanpa data pribadi).

## Belum terbukti

1. **Latensi live Codex.** Butuh satu giliran model kecil; dijalankan dengan persetujuan pemilik produk. Karena Codex menulis item saat selesai, yang diukur adalah jeda dari selesainya item sampai event terbaca.
2. **Path sangat panjang.** Claude Code mungkin memotong nama folder transcript yang panjang; belum diuji.
3. **Rotasi atau pemotongan file transcript** (misalnya `/compact` di Claude Code atau sesi baru di folder yang sama) belum ditangani; follower mengikuti satu file sampai dihentikan.

## Keputusan

Tampilan terstruktur M1 dibangun dari transcript lewat adapter per agent ini, tanpa mode headless dan tanpa hooks. Status worker (F3-06) juga membaca `turn_end` dari sini. Keputusan direvisi bila latensi live Codex melewati 1 detik atau bila temuan 2 terbukti mengganggu developer.
