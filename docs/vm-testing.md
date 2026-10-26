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

Salin `astrad` untuk `linux/arm64` ke VM, lalu:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o astrad-linux-arm64 ./cmd/astrad
scp -i ~/.ssh/astra_vm -P 2222 astrad-linux-arm64 dev@127.0.0.1:/tmp/astrad
scripts/vm/ssh.sh 'sudo install -m 0755 /tmp/astrad /usr/local/bin/astrad'
```

| Skrip (dijalankan di VM sebagai user yang boleh sudo) | Fungsi |
| --- | --- |
| `scripts/vm/setup-isolation.sh` | Memasang Docker, membuat user `astrad` dan `astra-agent`, folder, rahasia palsu, dan aturan sudoers |
| `scripts/vm/redteam-s3.sh` | Menjalankan serangan sebagai agent dan memeriksa semuanya ditolak (spike S3) |
| `scripts/vm/killmode-test.sh` | Membandingkan nasib sesi saat daemon di-restart: `KillMode` bawaan, `KillMode=process`, dan unit transient per sesi (spike S1 systemd) |

Salin skrip ke VM dengan `scp` seperti di atas, lalu jalankan dengan `bash /tmp/<nama>.sh`. Hasil dan keputusannya ada di `docs/adr/0005-spike-s3-isolation-and-systemd.md`.
