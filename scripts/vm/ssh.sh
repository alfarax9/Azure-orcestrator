#!/bin/bash
# Runs a command in the Astra test VM: ~/astra-vm/ssh.sh 'uname -a'
exec ssh -i ~/.ssh/astra_vm -p 2222 -o BatchMode=yes -o StrictHostKeyChecking=accept-new -o UserKnownHostsFile="${ASTRA_VM_DIR:-$HOME/astra-vm}/known_hosts" dev@127.0.0.1 "$@"
