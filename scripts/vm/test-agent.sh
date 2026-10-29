#!/usr/bin/env bash
# M1 slice 3 end to end on a real host (F3-04, F3-05, F3-07, ADR 0005): a worker's agent runs
# as a unit of its workspace's agent user, the host reads the agent's transcript inside the
# unit, and the daemon user follows it through the socket without ever reading the agent's
# home. The agent is a stand-in for Claude Code (it asks to trust the folder, takes a prompt,
# edits a file and writes a transcript where Claude Code does), so no login or quota is needed.
# Run on a host prepared by setup-isolation.sh, as a user that may sudo:
#
#   bash test-agent.sh /tmp/astra-priv /tmp/astrad
set -uo pipefail

PRIV_SRC=${1:?path of the astra-priv binary}
ASTRAD_SRC=${2:?path of the astrad binary}
PRIV=/usr/libexec/astra/astra-priv
ASTRAD=/usr/local/bin/astrad
FAKE=/usr/local/bin/claude
AGENTS=/var/lib/astra/agents
DB=/var/lib/astrad/test-agent.db
WATCH_OUT=/tmp/astra-watch.out
fail=0

pass() { printf '  ok    %s\n' "$1"; }
bad()  { printf '  FAIL  %s\n' "$1"; fail=1; }
daemon() { sudo -n -u astrad env ASTRA_DB=$DB "$ASTRAD" "$@"; }
field() { python3 -c "import json,sys; print(json.load(sys.stdin)$1)"; }
status() { daemon worker show "$1" 2>/dev/null | field '["status"]' 2>/dev/null; }
wait_for() { # wait_for <description> <command...>: polls for up to 25 seconds
  local d=$1; shift
  for _ in $(seq 50); do "$@" >/dev/null 2>&1 && { pass "$d"; return 0; }; sleep 0.5; done
  bad "$d"; return 1
}

cleanup() {
  [[ -n ${WATCH_PID:-} ]] && sudo -n kill "$WATCH_PID" 2>/dev/null
  sudo -n systemctl stop "astra-s-betaco-${WID:-none}.service" 2>/dev/null
  sudo -n rm -f "/etc/sudoers.d/astra-agent-betaco" $DB $DB-wal $DB-shm $WATCH_OUT
  [[ -f $FAKE.marker ]] && sudo -n rm -f $FAKE $FAKE.marker
  id astra-agent-betaco >/dev/null 2>&1 && sudo -n userdel astra-agent-betaco
  getent group astra-agent-betaco >/dev/null && sudo -n groupdel astra-agent-betaco
  sudo -n rm -rf "${AGENTS:?}/betaco"
  return 0
}
trap cleanup EXIT

echo "== install"
getent passwd astrad >/dev/null || { echo "run setup-isolation.sh first" >&2; exit 1; }
sudo -n install -D -o root -g root -m 0755 "$PRIV_SRC" "$PRIV"
sudo -n install -o root -g root -m 0755 "$ASTRAD_SRC" "$ASTRAD"
echo "astrad ALL=(root) NOPASSWD: $PRIV" | sudo -n tee /etc/sudoers.d/astrad-priv >/dev/null
sudo -n chmod 0440 /etc/sudoers.d/astrad-priv
[[ -e $FAKE && ! -f $FAKE.marker ]] && { echo "$FAKE exists and is not ours; refusing to replace a real agent" >&2; exit 1; }
cleanup
cat <<'AGENT' | sudo -n tee $FAKE >/dev/null
#!/bin/sh
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
done
AGENT
sudo -n chmod 0755 $FAKE; sudo -n touch $FAKE.marker

echo "== a worker"
daemon workspace add betaco "Beta Co" >/dev/null 2>&1 && pass "workspace betaco" || bad "workspace add"
pid=$(daemon project add betaco site "" 2>/dev/null | field '["id"]')
WID=$(daemon worker new "$pid" "Fix the header" 2>/dev/null | field '["id"]')
[[ -n $pid && -n $WID ]] && pass "project and worker $WID" || bad "project or worker"

echo "== start (F3-02, F3-04)"
daemon worker start "$WID" >/dev/null 2>&1 && pass "the agent started as a unit of the agent user" || bad "worker start"
unit=astra-s-betaco-$WID.service
[[ $(systemctl is-active $unit) == active ]] && pass "unit $unit is active" || bad "unit state: $(systemctl is-active $unit)"
host=$(systemctl show $unit -p MainPID --value)
[[ $(ps -o user= -p "$host") == astra-agent-betaco ]] && pass "the host runs as the agent user" || bad "host user"
tr '\0' ' ' < <(sudo -n cat /proc/"$host"/cmdline) | grep -q -- '--follow claude' && pass "the host was told to follow Claude's transcript" || bad "no --follow"
[[ -n $(status "$WID") ]] && pass "status before the first prompt: $(status "$WID")" || bad "no status"

echo "== watch, trust question, task (F3-05, F3-07)"
sudo -n -u astrad env ASTRA_DB=$DB "$ASTRAD" worker watch "$WID" >$WATCH_OUT 2>&1 &
WATCH_PID=$!
wait_for "the agent asks to trust the folder and the daemon saw it (needs_input)" bash -c "[[ \$(sudo -n -u astrad env ASTRA_DB=$DB $ASTRAD worker show $WID | python3 -c 'import json,sys;print(json.load(sys.stdin)[\"agentActivity\"])') == needs_input ]]"
daemon worker press "$WID" '\r' >/dev/null 2>&1 && pass "the person answered the question" || bad "press"
daemon worker send "$WID" "Fix the header" >/dev/null 2>&1 && pass "the task was sent as a paste plus Enter" || bad "send"
wait_for "the turn finished: idle with 2 lines added" bash -c "daemon_show=\$(sudo -n -u astrad env ASTRA_DB=$DB $ASTRAD worker show $WID); [[ \$(echo \"\$daemon_show\" | python3 -c 'import json,sys;d=json.load(sys.stdin);print(d[\"agentActivity\"], d[\"added\"])') == 'idle 2' ]]"
w=$(daemon worker show "$WID")
echo "  worker: $(echo "$w" | field '["status"]'), +$(echo "$w" | field '["added"]') -$(echo "$w" | field '["removed"]')"

echo "== the daemon never reads the agent's files"
sudo -n -u astrad cat "$AGENTS/betaco/home/.claude/projects"/*/fake.jsonl >/dev/null 2>&1 && bad "the daemon user read the transcript" || pass "the daemon user cannot read the agent's transcript"
sudo -n test -f "$AGENTS/betaco/home/.claude/projects/$(echo "$AGENTS/betaco/work/site/workers/$WID" | sed 's/[^A-Za-z0-9]/-/g')/fake.jsonl" && pass "the transcript exists, in the agent's home" || bad "no transcript"

echo "== a daemon restart (F3-04)"
sudo -n kill "$WATCH_PID"; wait "$WATCH_PID" 2>/dev/null; WATCH_PID=
[[ $(systemctl is-active $unit) == active ]] && pass "the session kept running without a daemon" || bad "the session died with the watcher"
daemon worker send "$WID" "Once more" >/dev/null 2>&1
sleep 2
sudo -n -u astrad env ASTRA_DB=$DB "$ASTRAD" worker watch "$WID" >$WATCH_OUT 2>&1 &
WATCH_PID=$!
wait_for "a new daemon replays the history and finds the worker idle again" bash -c "[[ \$(sudo -n -u astrad env ASTRA_DB=$DB $ASTRAD worker show $WID | python3 -c 'import json,sys;print(json.load(sys.stdin)[\"agentActivity\"])') == idle ]]"

echo "== interrupt and stop (F3-07)"
daemon worker interrupt "$WID" >/dev/null 2>&1 && pass "interrupt sent" || bad "interrupt"
daemon worker stop "$WID" >/dev/null 2>&1 && pass "stop sent" || bad "stop"
sleep 2
[[ $(systemctl is-active $unit 2>&1) != active ]] && pass "the unit is gone" || bad "unit still active"
wait "$WATCH_PID" 2>/dev/null; WATCH_PID=
[[ $(sudo -n -u astrad env ASTRA_DB=$DB $ASTRAD worker show "$WID" | field '["agentActivity"]') == idle ]] && pass "the worker is idle after stop" || bad "activity after stop"

echo
(( fail )) && { echo "agent slice: FAILED"; exit 1; }
echo "agent slice: all checks passed"
