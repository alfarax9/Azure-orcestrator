# VM uji Linux di Mac

Spike yang butuh Linux (S1 systemd, S3 isolasi, S6 cgroups) dijalankan di VM Ubuntu 24.04 arm64 yang dibuat sendiri, tanpa VPS. VM ini dijalankan langsung dengan QEMU dan akselerasi Hypervisor Apple. UTM tidak dipakai, karena sandbox-nya hanya membaca file yang dipilih lewat dialog "Open" dan menolak file dari AppleScript (UTM membuat disk kosong dan mengabaikan `source`).

Hasil di VM ini valid untuk mekanisme kernel dan systemd (izin, user, cgroup, unit). Angka RAM, CPU, dan latensi tetap harus diukur di VPS sungguhan.

## Membuat VM

Semua file ada di `~/astra-vm` (ubah dengan `ASTRA_VM_DIR`).

1. **Alat.** `brew install qemu` (untuk `qemu-system-aarch64` dan `qemu-img`).
2. **Cloud image.** Unduh `ubuntu-24.04-server-cloudimg-arm64.img` dari `cloud-images.ubuntu.com/releases/24.04/release/`, cocokkan dengan `SHA256SUMS`, lalu besarkan salinannya:
   ```sh
   cp ubuntu-24.04-server-cloudimg-arm64.img astra-vm-disk.qcow2
   qemu-img resize astra-vm-disk.qcow2 40G
   ```
3. **Kunci SSH khusus VM.** `ssh-keygen -t ed25519 -N "" -f ~/.ssh/astra_vm -C astra-vm`.
4. **Seed cloud-init** (`seed.iso`, label `cidata`) berisi `user-data` dan `meta-data`. `user-data` membuat user `dev` dengan sudo tanpa password, hanya login dengan kunci publik tadi, memasang `qemu-guest-agent`, dan menyalakan `ssh`. Buat ISO-nya dengan `hdiutil makehybrid -iso -joliet -default-volume-name cidata -o seed.iso seed/`.
5. **Firmware.** `cp /opt/homebrew/share/qemu/edk2-arm-vars.fd vars.fd`.
6. **Jalankan.** `scripts/vm/start-vm.sh`. VM berjalan di latar belakang dan konsolnya masuk ke `console.log`. Matikan dengan `kill $(cat ~/astra-vm/qemu.pid)`.

SSH masuk lewat port yang diteruskan ke host: `scripts/vm/ssh.sh 'uname -a'` (`dev@127.0.0.1:2222`). Cloud-init butuh sekitar satu menit untuk selesai pada boot pertama.

## Menyiapkan isolasi dan menjalankan uji

Salin `astrad` dan `astra-priv` untuk `linux/arm64` ke VM, lalu:

```sh
for c in astrad astra-priv; do CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o /tmp/$c-linux ./cmd/$c; done
scp -i ~/.ssh/astra_vm -P 2222 /tmp/astrad-linux dev@127.0.0.1:/tmp/astrad
scp -i ~/.ssh/astra_vm -P 2222 /tmp/astra-priv-linux dev@127.0.0.1:/tmp/astra-priv
scripts/vm/ssh.sh 'sudo install -m 0755 /tmp/astrad /usr/local/bin/astrad'
```

| Skrip (dijalankan di VM sebagai user yang boleh sudo) | Fungsi |
| --- | --- |
| `scripts/vm/setup-isolation.sh <slug>...` | Memasang Docker, membuat user `astrad` dan satu user `astra-agent-<slug>` per workspace, folder, rahasia palsu, dan aturan sudoers (contoh: `acme rina`) |
| `scripts/vm/redteam-s3.sh [ws] [lain]` | Menjalankan serangan sebagai agent workspace `ws` (termasuk ke workspace `lain`) dan memeriksa semuanya ditolak (spike S3) |
| `scripts/vm/killmode-test.sh [ws]` | Membandingkan nasib sesi saat daemon di-restart: `KillMode` bawaan, `KillMode=process`, dan unit transient per sesi (spike S1 systemd) |
| `scripts/vm/test-priv.sh <astra-priv> <astrad>` | Memasang helper root `astra-priv` dan aturan sudo-nya, lalu menguji ketiga kata kerjanya dan menyerangnya sebagai daemon dan sebagai agent (ADR 0005) |
| `scripts/vm/test-install.sh` (dijalankan **di Mac**, bukan di VM) | Potongan 6 M1 lewat SSH sungguhan: kunci SSH baru, kepercayaan host key, pemeriksaan kesiapan, `astra server install` dari folder binary dan dari URL rilis (termasuk rilis yang dirusak), pairing tanpa menampilkan token, tunnel, web UI dan checklist lewat tunnel, update, host key yang berubah, dan service hidup lagi setelah reboot VM. Membersihkan instalasi sebelumnya di VM dan membersihkan lagi di akhir |
| `scripts/vm/test-serve.sh <astra-priv> <astrad>` | Potongan 4 M1: `astrad serve` sebagai layanan systemd, hanya di loopback, pairing sekali pakai, worker lewat API, restart daemon, dan pencabutan perangkat. Memakai agen tiruan yang sama dengan `test-agent.sh` |
| `scripts/vm/test-agent.sh <astra-priv> <astrad>` | Potongan 3 M1: worker berjalan sebagai unit user agent, host membaca transcript, daemon melihat "Needs you", mengirim tugas, interrupt, stop, dan memutar ulang riwayat setelah restart. Memakai agen tiruan di `/usr/local/bin/claude` (menolak bila sudah ada agent sungguhan di sana) |
| `scripts/vm/test-workspace.sh <astra-priv> <astrad>` | Potongan 2 M1 dari ujung ke ujung: workspace lewat helper, clone, deteksi stack, worktree per worker, dan sesi di worktree worker |

Salin skrip ke VM dengan `scp` seperti di atas, jalankan `sudo bash /tmp/setup-isolation.sh acme rina`, lalu `bash /tmp/redteam-s3.sh acme rina`, `bash /tmp/killmode-test.sh acme`, `bash /tmp/test-priv.sh /tmp/astra-priv /tmp/astrad`, `bash /tmp/test-workspace.sh /tmp/astra-priv /tmp/astrad`, `bash /tmp/test-agent.sh /tmp/astra-priv /tmp/astrad`, dan `bash /tmp/test-serve.sh /tmp/astra-priv /tmp/astrad` (butuh `curl` dan `python3`, bawaan Ubuntu). Hasil dan keputusannya ada di `docs/adr/0005-spike-s3-isolation-and-systemd.md`.
