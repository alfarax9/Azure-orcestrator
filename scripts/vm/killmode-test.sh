#!/usr/bin/env bash
# Spike S1 (systemd): does an agent session survive "systemctl restart" of the daemon?
# A stand-in daemon unit starts one session and then sleeps. Three ways to keep the
# session alive are compared:
#   A  default KillMode=control-group        (expected: the session dies with the daemon)
#   B  KillMode=process                       (the session stays inside the daemon's cgroup)
#   D  each session as its own transient unit (systemd-run, in a slice with resource limits)
# Run on a host prepared by setup-isolation.sh, as a user that may sudo.
set -uo pipefail

WS=${1:-acme}
AGENT=astra-agent-$WS
BASE=/var/lib/astra/agents/$WS
ASTRAD=/usr/local/bin/astrad
SESSIONS=$BASE/sessions
UNIT=astrad-ks
fail=0

as_daemon() { sudo -n -u astrad -H env ASTRA_SESSIONS_DIR="$SESSIONS" "$ASTRAD" "$@"; }
host_pid() { pgrep -f "session-host --socket $SESSIONS/ks1.sock" | head -1; }
# still_running <pid>: the very same process, not a new session started by a restarted unit.
still_running() { [[ -n ${1:-} ]] && sudo -n test -d /proc/"$1" && [[ $(host_pid) == "$1" ]]; }
alive_sessions() { as_daemon session ls 2>/dev/null | tr '\n' ' '; }

cleanup() {
  sudo -n systemctl stop "$UNIT" astra-s-ks1 2>/dev/null || true
  sudo -n pkill -f "session-host --socket $SESSIONS/ks1.sock" 2>/dev/null || true
  sudo -n rm -f /etc/systemd/system/$UNIT.service /etc/sudoers.d/astrad-ks "$SESSIONS"/ks1.* 2>/dev/null || true
  sudo -n systemctl daemon-reload
  sudo -n systemctl reset-failed 2>/dev/null || true
}
trap cleanup EXIT
cleanup

write_unit() { # $1 = extra [Service] lines, $2 = ExecStart
  sudo -n tee /etc/systemd/system/$UNIT.service >/dev/null <<EOF
[Unit]
Description=Astra KillMode test (stand-in daemon)
[Service]
Type=simple
User=astrad
$1
ExecStart=/bin/sh -c '$2; exec /bin/sleep infinity'
EOF
  sudo -n systemctl daemon-reload
}

START_IN_UNIT="env ASTRA_SESSIONS_DIR=$SESSIONS $ASTRAD session start ks1 -user $AGENT -dir $BASE/work -- /bin/sleep 3600"

report() { # $1 = label
  local pid cg
  pid=$(host_pid)
  cg=$([[ -n $pid ]] && sudo -n cat /proc/"$pid"/cgroup | sed 's#0::##')
  printf '  %-26s host=%-6s sessions=[%s] cgroup=%s\n' "$1" "${pid:-gone}" "$(alive_sessions)" "${cg:--}"
}

run_variant() { # $1 = name, $2 = [Service] extra lines
  echo "== variant $1"
  write_unit "$2" "$START_IN_UNIT"
  sudo -n systemctl start $UNIT; sleep 2
  report "after start"
  FIRST=$(host_pid)
  sudo -n systemctl restart $UNIT; sleep 2
  report "after restart"
  SURVIVED_RESTART=$(still_running "$FIRST" && echo yes || echo no)
  sudo -n systemctl stop $UNIT; sleep 2
  report "after stop"
  SURVIVED_STOP=$(still_running "$FIRST" && echo yes || echo no)
  echo "  journal: $(sudo -n journalctl -u $UNIT --no-pager -o cat 2>/dev/null | grep -iE 'left-over|leftover|killing|kill' | tail -3 | tr '\n' '|')"
  as_daemon session stop ks1 >/dev/null 2>&1; sudo -n pkill -f "session-host --socket $SESSIONS/ks1.sock" 2>/dev/null; sleep 1
  sudo -n systemctl stop $UNIT 2>/dev/null; sudo -n rm -f "$SESSIONS"/ks1.*
  printf '  RESULT %s: survives restart=%s, survives stop=%s\n' "$1" "$SURVIVED_RESTART" "$SURVIVED_STOP"
}

run_variant A 'KillMode=control-group'
A_RESTART=$SURVIVED_RESTART
run_variant B 'KillMode=process'
B_RESTART=$SURVIVED_RESTART; B_STOP=$SURVIVED_STOP

echo "== variant D (a transient unit per session, in a resource-limited slice)"
write_unit 'KillMode=control-group' 'sleep 0'
# The daemon user may start sessions as units. This rule is wide on purpose for the test:
# production needs a narrower broker (see docs/adr/0003).
echo 'astrad ALL=(root) NOPASSWD: /usr/bin/systemd-run' | sudo -n tee /etc/sudoers.d/astrad-ks >/dev/null
sudo -n chmod 0440 /etc/sudoers.d/astrad-ks
sudo -n systemctl start $UNIT; sleep 1
sudo -n -u astrad sudo -n systemd-run --quiet --unit=astra-s-ks1 --slice=astra-agents.slice \
  --uid=$AGENT --gid=$AGENT --working-directory=$BASE/work \
  --property=MemoryMax=512M --property=CPUQuota=100% --property=TasksMax=256 \
  --setenv=HOME=$BASE/home --setenv=TERM=xterm-256color \
  "$ASTRAD" session-host --socket "$SESSIONS/ks1.sock" --socket-mode 660 --dir $BASE/work -- /bin/sleep 3600
sleep 2
report "after start"
FIRST=$(host_pid)
echo "  limits: $(systemctl show astra-s-ks1 -p MemoryMax -p CPUQuotaPerSecUSec -p TasksMax -p Slice | tr '\n' ' ')"
echo "  agent env keys: $(sudo -n cat /proc/"$(pgrep -u "$AGENT" -x sleep | head -1)"/environ | tr '\0' '\n' | cut -d= -f1 | tr '\n' ' ')"
sudo -n systemctl restart $UNIT; sleep 2
report "after daemon restart"
D_RESTART=$(still_running "$FIRST" && echo yes || echo no)
sudo -n systemctl stop $UNIT; sleep 2
report "after daemon stop"
D_STOP=$(still_running "$FIRST" && echo yes || echo no)
as_daemon session ls | grep -q ks1 && echo "  the daemon user still finds the session socket" || echo "  the daemon user cannot find the session"
sudo -n systemctl stop astra-s-ks1 2>/dev/null
printf '  RESULT D: survives restart=%s, survives stop=%s\n' "$D_RESTART" "$D_STOP"

echo "== summary"
[[ $A_RESTART == no ]] && echo "A: sessions die with the daemon, as expected" || { echo "A: UNEXPECTED, the session survived"; fail=1; }
[[ $B_RESTART == yes ]] && echo "B: sessions survive a restart (KillMode=process)" || { echo "B: the session did NOT survive"; fail=1; }
[[ $D_RESTART == yes && $D_STOP == yes ]] && echo "D: sessions survive restart and stop as separate units" || { echo "D: the session did NOT survive"; fail=1; }
(( fail )) && exit 1 || exit 0
