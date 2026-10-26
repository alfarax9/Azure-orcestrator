#!/usr/bin/env bash
# Proves the root helper astra-priv (ADR 0005): that it does the three jobs, that the
# daemon user can do nothing else as root, and that hostile requests and a hostile agent
# are refused. Run on a host prepared by setup-isolation.sh, as a user that may sudo, with
# astrad and astra-priv already built for the host:
#
#   bash test-priv.sh /tmp/astra-priv /tmp/astrad
set -uo pipefail

PRIV_SRC=${1:?path of the astra-priv binary}
ASTRAD_SRC=${2:?path of the astrad binary}
WS=betaco
AGENT=astra-agent-$WS
BASE=/var/lib/astra/agents/$WS
PRIV=/usr/libexec/astra/astra-priv
ASTRAD=/usr/local/bin/astrad
fail=0

pass() { printf '  ok    %s\n' "$1"; }
bad()  { printf '  FAIL  %s\n' "$1"; fail=1; }
check() { # check <description> <command...>: passes when the command succeeds
  local d=$1; shift
  if "$@" >/dev/null 2>&1; then pass "$d"; else bad "$d"; fi
}
refused() { # refused <description> <command...>: passes when the command fails
  local d=$1; shift
  if "$@" >/dev/null 2>&1; then bad "$d"; else pass "$d"; fi
}
priv() { sudo -n -u astrad sudo -n "$PRIV" "$@"; }
as_agent() { sudo -n -u "$AGENT" "$@"; }
as_daemon() { sudo -n -u astrad env ASTRA_SESSIONS_DIR="$BASE/sessions" "$ASTRAD" "$@"; }

cleanup() {
  sudo -n systemctl stop "astra-s-$WS-t1.service" "astra-s-$WS-t2.service" 2>/dev/null || true
  sudo -n rm -f "/etc/sudoers.d/$AGENT" "/etc/sudoers.d/.$AGENT.tmp"
  id "$AGENT" >/dev/null 2>&1 && sudo -n userdel "$AGENT" 2>/dev/null
  getent group "$AGENT" >/dev/null && sudo -n groupdel "$AGENT" 2>/dev/null
  sudo -n rm -rf "$BASE"
  return 0
}
trap cleanup EXIT

echo "== install"
getent passwd astrad >/dev/null || { echo "run setup-isolation.sh first" >&2; exit 1; }
sudo -n install -D -o root -g root -m 0755 "$PRIV_SRC" "$PRIV"
sudo -n install -o root -g root -m 0755 "$ASTRAD_SRC" "$ASTRAD"
# The daemon may run this one program as root, and nothing else. The wide systemd-run rule
# of the first experiment goes away.
sudo -n rm -f /etc/sudoers.d/astrad-ks
echo "astrad ALL=(root) NOPASSWD: $PRIV" | sudo -n tee /etc/sudoers.d/astrad-priv >/dev/null
sudo -n chmod 0440 /etc/sudoers.d/astrad-priv
check "sudoers rule is valid" sudo -n visudo -cf /etc/sudoers.d/astrad-priv
cleanup

echo "== workspace-add"
check "creates the workspace" priv workspace-add $WS
check "the agent user exists" id "$AGENT"
groups=$(id -nG "$AGENT")
[[ $groups == "$AGENT" ]] && pass "the agent user is in its own group only ($groups)" || bad "groups: $groups"
for want in "$BASE:710:$AGENT:astrad" "$BASE/sessions:2770:$AGENT:astrad" "$BASE/work:700:$AGENT:$AGENT" "$BASE/home:700:$AGENT:$AGENT"; do
  path=${want%%:*}; rest=${want#*:}; got=$(sudo -n stat -c '%a:%U:%G' "$path")
  [[ $got == "$rest" ]] && pass "$path is $got" || bad "$path is $got, want $rest"
done
check "the sudoers file is valid" sudo -n visudo -cf "/etc/sudoers.d/$AGENT"
[[ $(sudo -n stat -c '%a:%U' "/etc/sudoers.d/$AGENT") == "440:root" ]] && pass "sudoers file is 0440 root" || bad "sudoers file permissions"
uid=$(id -u "$AGENT")
check "creating it again is fine" priv workspace-add $WS
[[ $(id -u "$AGENT") == "$uid" ]] && pass "the user was kept, not recreated" || bad "the user was recreated"
[[ $(sudo -n -u astrad sudo -n -u "$AGENT" id -un) == "$AGENT" ]] && pass "the daemon can run commands as the agent user" || bad "daemon cannot run as the agent user"

echo "== what the daemon user cannot do"
refused "no other command as root" sudo -n -u astrad sudo -n id
refused "no systemd-run as root" sudo -n -u astrad sudo -n /usr/bin/systemd-run --quiet true
refused "no useradd as root" sudo -n -u astrad sudo -n /usr/sbin/useradd evil
refused "no shell as another workspace user" sudo -n -u astrad sudo -n -u nobody id
refused "an agent cannot call the helper" sudo -n -u "$AGENT" sudo -n "$PRIV" workspace-add $WS
refused "an agent cannot read the daemon's secrets" sudo -n -u "$AGENT" cat /var/lib/astrad/production/acme.env

echo "== hostile requests are refused with a usage error and change nothing"
before=$(getent passwd | wc -l)
for args in "workspace-add x;id" "workspace-add ../etc" "workspace-add -rf" "workspace-add ACME" "format-disk" \
  "session-start $WS t1 --agent bash" "session-start $WS T1 --agent shell" "session-start $WS t1;id --agent shell" \
  "session-start $WS t1 --agent shell --dir ../../.." "session-start $WS t1 --agent shell --dir /etc" \
  "session-start $WS t1 --agent shell --env X=1" "session-start $WS t1 --agent shell --memory-max 999999" \
  "session-stop $WS t1;reboot"; do
  # shellcheck disable=SC2086
  priv $args >/dev/null 2>&1; rc=$?
  [[ $rc == 2 ]] && pass "refused (exit 2): $args" || bad "exit $rc for: $args"
done
[[ $(getent passwd | wc -l) == "$before" ]] && pass "no user was created by them" || bad "the user list changed"

echo "== session-start"
as_agent mkdir -p "$BASE/work/proj"
as_agent ln -sfn /etc "$BASE/work/escape"
priv session-start $WS t1 --agent shell --dir proj --memory-max 512 --cpu-quota 50 >/tmp/priv-sock 2>/tmp/priv-err; rc=$?
[[ $rc == 0 && $(cat /tmp/priv-sock) == "$BASE/sessions/t1.sock" ]] && pass "starts a session and prints its socket" || { bad "session-start rc=$rc: $(cat /tmp/priv-err)"; }
sleep 2
unit=astra-s-$WS-t1.service
[[ $(systemctl is-active $unit) == active ]] && pass "the unit is active" || bad "unit state: $(systemctl is-active $unit)"
show=$(systemctl show $unit -p Slice -p MemoryMax -p TasksMax -p CPUQuotaPerSecUSec -p NoNewPrivileges | sort | tr '\n' ' ')
echo "  unit: $show"
[[ $show == *"Slice=astra-agents-$WS.slice"* && $show == *"MemoryMax=536870912"* && $show == *"TasksMax=512"* && $show == *"CPUQuotaPerSecUSec=500ms"* && $show == *"NoNewPrivileges=yes"* ]] \
  && pass "slice and limits are set" || bad "unit properties differ"
pid=$(systemctl show $unit -p MainPID --value)
[[ $(ps -o user= -p "$pid") == "$AGENT" ]] && pass "the session host runs as the agent user" || bad "host user: $(ps -o user= -p "$pid")"
[[ $(sudo -n readlink /proc/"$pid"/cwd) == "$BASE/work/proj" ]] && pass "working directory is work/proj" || bad "cwd: $(sudo -n readlink /proc/"$pid"/cwd)"
as_daemon session ls 2>/dev/null | grep -qx t1 && pass "the daemon user finds the session" || bad "the daemon user cannot find the session"
as_daemon session send t1 'echo marker-$((6*7)) $HOME\r' >/dev/null 2>&1; sleep 1
out=$(as_daemon session tail t1 -for 1s -strip 2>/dev/null)
[[ $out == *"marker-42 $BASE/home"* ]] && pass "the daemon drove the shell; HOME is the agent's home" || bad "tail: $out"
shellpid=$(pgrep -u "$AGENT" -x bash | head -1)
keys=$(sudo -n cat /proc/"$shellpid"/environ | tr '\0' '\n' | cut -d= -f1 | sort | tr '\n' ' ')
echo "  agent environment: $keys"
if sudo -n cat /proc/"$shellpid"/environ | tr '\0' '\n' | grep -qE '^(SUDO_|SSH_|ASTRA_|XDG_|DBUS_)'; then bad "daemon environment leaked into the agent"; else pass "no variable of the daemon, sudo or ssh reached the agent"; fi
refused "a working directory that is a link out of work/ is refused" priv session-start $WS t2 --agent shell --dir escape
refused "the same session cannot be started twice" priv session-start $WS t1 --agent shell

echo "== session-stop"
check "stops the session" priv session-stop $WS t1
sleep 1
[[ $(systemctl is-active $unit 2>&1) != active ]] && pass "the unit is gone" || bad "unit still active"
[[ -z $(pgrep -u "$AGENT" -x bash) ]] && pass "nothing of the agent is left running" || bad "processes left"
refused "stopping a session that is not there fails" priv session-stop $WS t1

echo "== a user with extra groups"
sudo -n usermod -aG docker "$AGENT"
refused "workspace-add refuses an agent user in docker" priv workspace-add $WS
sudo -n gpasswd -d "$AGENT" docker >/dev/null
check "and works again once it is removed" priv workspace-add $WS

echo "== an agent that swaps its folder for a link"
as_agent mv "$BASE/sessions" "$BASE/sessions.old"
as_agent ln -s /etc/ssh "$BASE/sessions"
before=$(sudo -n stat -c '%a:%U:%G' /etc/ssh)
refused "workspace-add does not follow the link" priv workspace-add $WS
[[ $(sudo -n stat -c '%a:%U:%G' /etc/ssh) == "$before" ]] && pass "/etc/ssh was not touched ($before)" || bad "/etc/ssh changed"
as_agent rm "$BASE/sessions"
as_agent mv "$BASE/sessions.old" "$BASE/sessions"
check "put the folder back and it works again" priv workspace-add $WS

echo
(( fail )) && { echo "astra-priv: FAILED"; exit 1; }
echo "astra-priv: all checks passed"
