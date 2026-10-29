#!/usr/bin/env bash
# M1 slice 4 on a real host (F1-05, F3-02, ADR 0006): astrad runs as a systemd service of the
# daemon user, serves the API on loopback only, pairs a device with a one-time token, runs a
# worker's agent as the agent user, keeps the worker's facts true from the agent's events, and
# picks its watches up again after a restart. The agent is the stand-in of test-agent.sh.
# WebSocket behaviour is covered by the Go tests (cmd/astrad), which run the same code.
#
#   bash test-serve.sh /tmp/astra-priv /tmp/astrad
set -uo pipefail

PRIV_SRC=${1:?path of the astra-priv binary}
ASTRAD_SRC=${2:?path of the astrad binary}
PRIV=/usr/libexec/astra/astra-priv
ASTRAD=/usr/local/bin/astrad
FAKE=/usr/local/bin/claude
AGENTS=/var/lib/astra/agents
DB=/var/lib/astrad/test-serve.db
UNIT=astrad-test
API=http://127.0.0.1:7437/api/v1
fail=0

pass() { printf '  ok    %s\n' "$1"; }
bad()  { printf '  FAIL  %s\n' "$1"; fail=1; }
daemon() { sudo -n -u astrad env ASTRA_DB=$DB "$ASTRAD" "$@"; }
field() { python3 -c "import json,sys; print(json.load(sys.stdin)$1)"; }
api() { # api <method> <path> [json]: prints "<status> <body>"
  local out; out=$(curl -s -m 10 -w '\n%{http_code}' -X "$1" "$API$2" -H "Authorization: Bearer ${TOKEN:-}" -H 'Content-Type: application/json' ${3:+-d "$3"})
  printf '%s %s' "${out##*$'\n'}" "${out%$'\n'*}"
}
status_of() { api "$@" | cut -d' ' -f1; }
body_of() { api "$@" | cut -d' ' -f2-; }
worker_field() { body_of GET "/workers/$WID" | field "$1" 2>/dev/null; }
wait_for() { # wait_for <description> <command...>: polls for up to 25 seconds
  local d=$1; shift
  for _ in $(seq 100); do "$@" >/dev/null 2>&1 && { pass "$d"; return 0; }; sleep 0.25; done
  bad "$d"; return 1
}
is_activity() { [[ $(worker_field '["agentActivity"]') == "$1" ]]; }
has_added() { [[ $(worker_field '["added"]') == "$1" ]]; }

cleanup() {
  sudo -n systemctl stop $UNIT 2>/dev/null
  sudo -n systemctl stop "astra-s-betaco-${WID:-none}.service" 2>/dev/null
  sudo -n rm -f /etc/systemd/system/$UNIT.service /etc/sudoers.d/astra-agent-betaco $DB $DB-wal $DB-shm
  sudo -n systemctl daemon-reload
  [[ -f $FAKE.marker ]] && sudo -n rm -f $FAKE $FAKE.marker
  id astra-agent-betaco >/dev/null 2>&1 && sudo -n userdel astra-agent-betaco
  getent group astra-agent-betaco >/dev/null && sudo -n groupdel astra-agent-betaco
  sudo -n rm -rf "${AGENTS:?}/betaco"
  return 0
}
trap cleanup EXIT

echo "== install"
getent passwd astrad >/dev/null || { echo "run setup-isolation.sh first" >&2; exit 1; }
[[ -e $FAKE && ! -f $FAKE.marker ]] && { echo "$FAKE exists and is not ours; refusing to replace a real agent" >&2; exit 1; }
cleanup
sudo -n install -D -o root -g root -m 0755 "$PRIV_SRC" "$PRIV"
sudo -n install -o root -g root -m 0755 "$ASTRAD_SRC" "$ASTRAD"
echo "astrad ALL=(root) NOPASSWD: $PRIV" | sudo -n tee /etc/sudoers.d/astrad-priv >/dev/null
sudo -n chmod 0440 /etc/sudoers.d/astrad-priv
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
sudo -n tee /etc/systemd/system/$UNIT.service >/dev/null <<UNITFILE
[Unit]
Description=Astra daemon (test)
[Service]
User=astrad
Environment=ASTRA_DB=$DB
ExecStart=$ASTRAD serve
UNITFILE
sudo -n systemctl daemon-reload

echo "== the daemon as a service"
sudo -n -u astrad env ASTRA_DB=$DB ASTRA_LISTEN=0.0.0.0:7438 "$ASTRAD" serve >/dev/null 2>&1 && bad "a non-loopback address was accepted" || pass "a non-loopback address is refused"
sudo -n systemctl start $UNIT
wait_for "the API answers" bash -c "curl -sf $API/version | grep -q '\"api\":1'"
[[ $(ss -ltn | grep -c ":7437 ") == 1 && -n $(ss -ltn | grep "127.0.0.1:7437") ]] && pass "it listens on 127.0.0.1 only" || bad "listening: $(ss -ltn | grep 7437)"
[[ $(status_of GET /workspaces) == 401 ]] && pass "no token, no access" || bad "an unauthenticated request was answered"

echo "== pairing (F1-05)"
PAIRING=$(daemon pair)
[[ ${#PAIRING} -ge 40 ]] && pass "astrad pair printed a token (not shown here)" || bad "pair output: ${#PAIRING} characters"
resp=$(curl -s -X POST $API/pairing -d "{\"token\":\"$PAIRING\",\"name\":\"VM laptop\"}")
TOKEN=$(echo "$resp" | field '["token"]' 2>/dev/null)
[[ -n $TOKEN ]] && pass "the pairing token was traded for a device token" || bad "pairing: $resp"
[[ $(curl -s -o /dev/null -w '%{http_code}' -X POST $API/pairing -d "{\"token\":\"$PAIRING\",\"name\":\"again\"}") == 401 ]] && pass "the pairing token works once" || bad "the pairing token worked twice"
grep -q "$TOKEN" < <(sudo -n journalctl -u $UNIT --no-pager -o cat 2>/dev/null) && bad "the device token reached the journal" || pass "no token in the journal"
[[ $(daemon devices | field '[0]["name"]') == "VM laptop" ]] && pass "the device is listed" || bad "devices"

echo "== a worker through the API (F3-02)"
[[ $(status_of POST /workspaces '{"slug":"betaco","name":"Beta Co"}') == 201 ]] && pass "workspace betaco" || bad "workspace"
PID=$(body_of POST /projects '{"workspace":"betaco","name":"site"}' | field '["id"]')
WID=$(body_of POST "/projects/$PID/workers" '{"title":"Fix the header","agent":"claude"}' | field '["id"]')
[[ -n $PID && -n $WID ]] && pass "project and worker $WID" || bad "project or worker"
[[ $(status_of POST "/workers/$WID/start") == 200 ]] && pass "the agent started" || bad "start"
unit=astra-s-betaco-$WID.service
[[ $(systemctl is-active $unit) == active ]] && pass "its unit is active, apart from the daemon's" || bad "unit state"
wait_for "the daemon saw the trust question (needs_input)" is_activity needs_input
daemon worker press "$WID" '\r' >/dev/null 2>&1; daemon worker send "$WID" "Fix the header" >/dev/null 2>&1
wait_for "the turn finished, found by the daemon: idle with 2 lines added" bash -c "$(declare -f api worker_field body_of field is_activity has_added); API=$API TOKEN=$TOKEN WID=$WID; is_activity idle && has_added 2"

echo "== a daemon restart (F3-04, N-07)"
sudo -n systemctl restart $UNIT
wait_for "the API is back" bash -c "curl -sf $API/version >/dev/null"
[[ $(systemctl is-active $unit) == active ]] && pass "the agent's session was not touched" || bad "the session died with the daemon"
[[ $(status_of GET /workers) == 200 ]] && pass "the same device token still works" || bad "the token stopped working"
daemon worker send "$WID" "Once more" >/dev/null 2>&1
wait_for "the new daemon was watching: the worker went back to idle on its own" bash -c "$(declare -f api worker_field body_of field is_activity); API=$API TOKEN=$TOKEN WID=$WID; is_activity idle"

echo "== revoking a device"
dev=$(daemon devices | field '[0]["id"]')
daemon devices revoke "$dev" >/dev/null 2>&1 && pass "the device was revoked" || bad "revoke"
[[ $(status_of GET /workspaces) == 401 ]] && pass "its token stopped working at once" || bad "a revoked token still works"
[[ $(daemon devices) == "[]" ]] && pass "no devices are left" || bad "devices after revoking"

echo "== stop"
daemon worker stop "$WID" >/dev/null 2>&1; sleep 2
[[ $(systemctl is-active $unit 2>&1) != active ]] && pass "the agent's unit is gone" || bad "the unit is still active"
sudo -n systemctl stop $UNIT && [[ $(systemctl is-active $UNIT 2>&1) != active ]] && pass "the daemon stops on SIGTERM" || bad "the daemon did not stop"

echo
(( fail )) && { echo "serve slice: FAILED"; exit 1; }
echo "serve slice: all checks passed"
