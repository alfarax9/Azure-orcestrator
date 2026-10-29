#!/bin/sh
# A stand-in for Claude Code, for trying Astra without a login or quota: it asks to trust the
# folder, takes a prompt, edits a file and writes a transcript where Claude Code keeps it.
dir=$(pwd -P | sed 's/[^A-Za-z0-9]/-/g')
proj="$HOME/.claude/projects/$dir"
mkdir -p "$proj"
printf 'Do you trust the files in this folder?\r\n'
read -r answer
printf '> \r\n'
while read -r line; do
  printf 'a\nb\n' > changed.txt
  now=$(date -u +%Y-%m-%dT%H:%M:%S.000Z)
  printf '{"type":"user","timestamp":"%s","message":{"role":"user","content":"task"}}\n' "$now" >> "$proj/fake.jsonl"
  printf '{"type":"assistant","timestamp":"%s","message":{"role":"assistant","content":[{"type":"text","text":"done"}],"stop_reason":"end_turn"}}\n' "$now" >> "$proj/fake.jsonl"
  printf 'Done: wrote changed.txt (2 lines)\r\n> \r\n'
done
