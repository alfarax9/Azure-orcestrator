#!/usr/bin/env bash
# Spike S3 red-team (F12-01): starts an agent session as the unprivileged agent user, the
# way the daemon will, and types the attacks into it. Every check must come out DENIED.
# Run on a host prepared by setup-isolation.sh, as a user that may sudo.
set -uo pipefail

ASTRAD=/usr/local/bin/astrad
export ASTRA_SESSIONS_DIR=/var/lib/astra/sessions
AGENT=astra-agent
fail=0

as_daemon() { sudo -n -u astrad -H env ASTRA_SESSIONS_DIR="$ASTRA_SESSIONS_DIR" "$ASTRAD" "$@"; }

# A stand-in for the running daemon: a process of the daemon user with a secret in its
# environment, which an agent must not read, signal or inspect.
sudo -n -u astrad -H bash -c 'SECRET_TOKEN=daemon-secret-value nohup sleep 600 >/dev/null 2>&1 &'
sleep 1
DPID=$(pgrep -u astrad -x sleep | head -1)

as_daemon session stop rt >/dev/null 2>&1 || true
as_daemon session start rt -user "$AGENT" -dir /var/lib/astra/work -- /bin/bash --noprofile --norc -i >/dev/null \
  || { echo "could not start the agent session as $AGENT"; exit 2; }

# The agent's terminal is a TUI-style screen, so reading it back is unreliable. Instead each
# attack runs inside the agent's shell with its output and exit status written to files in
# the agent's work folder, which this script reads as root. Commands must not contain double
# quotes (astrad session send reads Go escapes).
WORK=/var/lib/astra/work
n=0
run_as_agent() {
  n=$((n + 1))
  local f="$WORK/rt.$n"
  as_daemon session send rt "( $1 ) > $f.out 2>&1; echo \$? > $f.rc\\r" >/dev/null
  for _ in $(seq 1 40); do
    sudo -n test -s "$f.rc" && break
    sleep 0.25
  done
  RC=$(sudo -n cat "$f.rc" 2>/dev/null || echo timeout)
  OUT=$(sudo -n cat "$f.out" 2>/dev/null | tr '\n' ' ')
}

# check <name> <command typed as the agent> <regex that proves the attack was refused>
check() {
  local name=$1 cmd=$2 want=$3
  run_as_agent "$cmd"
  if grep -Eqi "$want" <<<"$OUT"; then
    printf 'DENIED   %-34s rc=%-3s %s\n' "$name" "$RC" "$(cut -c1-90 <<<"$OUT")"
  else
    printf 'ALLOWED  %-34s rc=%-3s %s\n' "$name" "$RC" "$(cut -c1-90 <<<"$OUT")"; fail=1
  fi
}

echo "== agent identity"
run_as_agent 'id'; echo "$OUT"
echo "== attacks (a refusal prints one of: denied, not permitted, cannot, no such)"
check 'read production .env'        'cat /var/lib/astrad/production/acme.env' 'permission denied'
check 'list daemon home'            'ls /var/lib/astrad' 'permission denied'
check 'read daemon secrets file'    'cat /etc/astra/secrets.env' 'permission denied'
check 'docker ps'                   'docker ps' 'permission denied|cannot connect|got permission'
check 'open docker.sock directly'   'test -r /var/run/docker.sock && echo READABLE || echo NOT_READABLE' 'NOT_READABLE'
check 'sudo to root'                'sudo -n true' 'password|not allowed|may not|not in the sudoers|a password is required'
check 'read daemon env via /proc'   "cat /proc/$DPID/environ" 'permission denied'
check 'signal the daemon'           "kill -9 $DPID" 'not permitted'
check 'read another user home'      'ls /home/dev' 'permission denied'
check 'write to /usr/local/bin'     'touch /usr/local/bin/pwned' 'permission denied|read-only'
check 'member of docker group?'      'id -nG | grep -cw docker || true' '^0'
check 'daemon secret not in env'    'env | grep -ci -E daemon-secret\\|cloudflare\\|DECOY || true' '^0'
echo "== control: what the agent is allowed to do"
run_as_agent 'echo ok > /var/lib/astra/work/ok.txt && echo WRITE_OK'
if grep -q WRITE_OK <<<"$OUT"; then echo "OK       the agent can write in its own work folder"; else echo "FAIL     the agent cannot write in its own work folder"; fail=1; fi
run_as_agent 'ps -o user:12,comm= -u astra-agent | sort -u'
echo "         processes of the agent user: $OUT"

as_daemon session stop rt >/dev/null 2>&1 || true
sudo -n pkill -u astrad -x sleep >/dev/null 2>&1 || true

echo "== the daemon's own power stops at the agent user"
if sudo -n -u astrad sudo -n -u root true >/dev/null 2>&1; then echo "ALLOWED  the daemon user can become root"; fail=1; else echo "DENIED   the daemon user cannot become root"; fi
if sudo -n -u astrad sudo -n -u dev true >/dev/null 2>&1; then echo "ALLOWED  the daemon user can become another human user"; fail=1; else echo "DENIED   the daemon user cannot become another human user"; fi
if sudo -n -u astrad sudo -n -u "$AGENT" true >/dev/null 2>&1; then echo "OK       the daemon user can run commands as the agent user"; else echo "FAIL     the daemon user cannot run commands as the agent user"; fail=1; fi

echo "== control: the isolation tests are meaningful"
if sudo -n docker ps >/dev/null 2>&1; then echo "OK       root can run docker ps (so the agent's refusal is permissions, not a broken Docker)"; else echo "WARN     docker does not work for root either"; fail=1; fi
if sudo -n -u astrad cat /var/lib/astrad/production/acme.env >/dev/null 2>&1; then echo "OK       the daemon user can read the production decoy"; else echo "WARN     the daemon cannot read its own decoy"; fail=1; fi
if ls /var/lib/astra/sessions >/dev/null 2>&1; then echo "WARN     an unrelated user can list the session sockets"; fail=1; else echo "OK       an unrelated user cannot list the session sockets"; fi

(( fail )) && { echo "RESULT: FAIL"; exit 1; }
echo "RESULT: every attack was refused"
