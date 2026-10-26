#!/bin/bash
# Boots the Astra test VM (Ubuntu 24.04 cloud image) with QEMU and Apple's Hypervisor framework.
# Files live in $ASTRA_VM_DIR (default ~/astra-vm); see docs/vm-testing.md. SSH: scripts/vm/ssh.sh
set -euo pipefail
VM_DIR="${ASTRA_VM_DIR:-$HOME/astra-vm}"; cd "$VM_DIR"
[ -f qemu.pid ] && kill -0 "$(cat qemu.pid)" 2>/dev/null && { echo "already running (pid $(cat qemu.pid))"; exit 0; }
exec qemu-system-aarch64 \
  -name astra-vm -machine virt,highmem=on -accel hvf -cpu host -smp 2 -m 4096 \
  -drive if=pflash,format=raw,readonly=on,file=/opt/homebrew/share/qemu/edk2-aarch64-code.fd \
  -drive if=pflash,format=raw,file=vars.fd \
  -drive file=astra-vm-disk.qcow2,format=qcow2,if=virtio,cache=writeback \
  -drive file=seed.iso,format=raw,if=virtio,readonly=on \
  -netdev user,id=n0,hostfwd=tcp:127.0.0.1:2222-:22 -device virtio-net-pci,netdev=n0 \
  -device virtio-rng-pci \
  -display none -serial file:console.log -monitor none \
  -daemonize -pidfile qemu.pid
