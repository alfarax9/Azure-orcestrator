#!/usr/bin/env bash
# M1 slice 2 end to end on a real host (F2-01, F2-02, F2-03, F3-01, ADR 0005): the daemon
# user creates workspaces through astra-priv, clones projects and makes worktrees as each
# workspace's agent user, and starts a session in a worker's worktree. Run on a host
# prepared by setup-isolation.sh, as a user that may sudo:
#
#   bash test-workspace.sh /tmp/astra-priv /tmp/astrad
set -uo pipefail

PRIV_SRC=${1:?path of the astra-priv binary}
ASTRAD_SRC=${2:?path of the astrad binary}
PRIV=/usr/libexec/astra/astra-priv
ASTRAD=/usr/local/bin/astrad
AGENTS=/var/lib/astra/agents
DB=/var/lib/astrad/test-workspace.db
REMOTE=$AGENTS/betaco/home/remote
fail=0

pass() { printf '  ok    %s\n' "$1"; }
bad()  { printf '  FAIL  %s\n' "$1"; fail=1; }
daemon() { sudo -n -u astrad env ASTRA_DB=$DB "$ASTRAD" "$@"; }
field() { python3 -c "import json,sys; print(json.load(sys.stdin)$1)"; }

cleanup() {
  for ws in betaco gamma; do
    sudo -n systemctl stop "astra-s-$ws-w1.service" 2>/dev/null
    sudo -n rm -f "/etc/sudoers.d/astra-agent-$ws"
    id "astra-agent-$ws" >/dev/null 2>&1 && sudo -n userdel "astra-agent-$ws"
    getent group "astra-agent-$ws" >/dev/null && sudo -n groupdel "astra-agent-$ws"
    sudo -n rm -rf "${AGENTS:?}/$ws"
  done
  sudo -n rm -f $DB $DB-wal $DB-shm
  return 0
}
trap cleanup EXIT

echo "== install"
getent passwd astrad >/dev/null || { echo "run setup-isolation.sh first" >&2; exit 1; }
sudo -n install -D -o root -g root -m 0755 "$PRIV_SRC" "$PRIV"
sudo -n install -o root -g root -m 0755 "$ASTRAD_SRC" "$ASTRAD"
echo "astrad ALL=(root) NOPASSWD: $PRIV" | sudo -n tee /etc/sudoers.d/astrad-priv >/dev/null
sudo -n chmod 0440 /etc/sudoers.d/astrad-priv
cleanup

echo "== workspaces (F2-01)"
out=$(daemon workspace add betaco "Beta Co" 2>&1) && pass "workspace betaco added" || bad "workspace add: $out"
id astra-agent-betaco >/dev/null 2>&1 && pass "its agent user was made by astra-priv" || bad "no agent user"
daemon workspace add gamma "Gamma" >/dev/null 2>&1 && pass "workspace gamma added" || bad "workspace add gamma"
daemon workspace add betaco "Again" >/dev/null 2>&1 && bad "a taken slug was accepted" || pass "a taken slug is refused"

# A remote holding a Next.js app on branch trunk. The agent user owns it: git refuses a
# repository that belongs to another user (safe.directory).
as_betaco() { sudo -n -u astra-agent-betaco env HOME=$AGENTS/betaco/home "$@"; }
as_betaco git init -q -b trunk $REMOTE
echo '{"dependencies":{"next":"15"},"scripts":{"dev":"next dev","build":"next build","test":"vitest"}}' | as_betaco tee $REMOTE/package.json >/dev/null
as_betaco git -C $REMOTE add . && as_betaco git -C $REMOTE -c user.name=t -c user.email=t@t commit -qm first

echo "== projects (F2-02, F2-03)"
out=$(daemon project add betaco site file://$REMOTE 2>&1); rc=$?
pid=$(echo "$out" | field '["id"]' 2>/dev/null)
[[ $rc == 0 && -n $pid ]] && pass "project cloned on the server" || bad "project add: $out"
[[ $(echo "$out" | field '["stack"]') == nextjs && $(echo "$out" | field '["defaultBranch"]') == trunk ]] \
  && pass "stack nextjs, default branch trunk, commands: $(echo "$out" | field '["commands"]')" || bad "detection: $out"
owner=$(sudo -n stat -c %U $AGENTS/betaco/work/site/repo/package.json)
[[ $owner == astra-agent-betaco ]] && pass "the clone belongs to the agent user" || bad "clone owner: $owner"
hello=$(daemon project add betaco hello https://github.com/octocat/Hello-World.git 2>&1)
[[ $(echo "$hello" | field '["defaultBranch"]' 2>/dev/null) == master ]] && pass "an HTTPS clone works (Hello-World, master)" || bad "https clone: $hello"
daemon project add betaco evil 'ext::sh -c id' >/dev/null 2>&1 && bad "ext:: URL accepted" || pass "an ext:: URL is refused"
daemon project add betaco site "" >/dev/null 2>&1 && bad "taken project name accepted" || pass "a taken project name is refused"
out=$(daemon project commands "$pid" -dev "npm run dev -- -p 3001" 2>&1)
[[ $(echo "$out" | field '["commands"]["dev"]') == "npm run dev -- -p 3001" ]] && pass "the user can correct the commands" || bad "commands: $out"

echo "== workers (F3-01)"
w1=$(daemon worker new "$pid" "Fix the header" 2>/dev/null)
w2=$(daemon worker new "$pid" "Fix the header" -agent codex 2>/dev/null)
id1=$(echo "$w1" | field '["id"]'); id2=$(echo "$w2" | field '["id"]')
b1=$(echo "$w1" | field '["branch"]'); b2=$(echo "$w2" | field '["branch"]')
[[ -n $id1 && -n $id2 && $b1 != "$b2" ]] && pass "two workers, branches $b1 and $b2" || bad "workers: $w1 / $w2"
wt1=$AGENTS/betaco/work/site/workers/$id1; wt2=$AGENTS/betaco/work/site/workers/$id2
sudo -n -u astra-agent-betaco sh -c "echo a > $wt1/header.txt"
[[ -n $id2 ]] && sudo -n test -e "$wt2/header.txt" && bad "a file of worker 1 appeared in worker 2" || pass "the worktrees do not share files"
[[ -n $b1 && $(sudo -n -u astra-agent-betaco git -C "$wt1" branch --show-current) == "$b1" ]] && pass "worker 1 is on its own branch" || bad "branch of worktree 1"

echo "== isolation between workspaces"
sudo -n -u astrad ls $AGENTS/betaco/work >/dev/null 2>&1 && bad "the daemon can open work/" || pass "the daemon user cannot open work/ itself"
sudo -n -u astra-agent-gamma ls $AGENTS/betaco/work >/dev/null 2>&1 && bad "gamma reads betaco" || pass "gamma's agent cannot read betaco's work"
sudo -n -u astrad sudo -n -u astra-agent-gamma ls $AGENTS/betaco/work >/dev/null 2>&1 && bad "commands as gamma reach betaco" || pass "commands run as gamma cannot reach betaco either"

echo "== a session in the worker's worktree"
sock=$(sudo -n -u astrad sudo -n $PRIV session-start betaco w1 --agent shell --dir "site/workers/$id1" 2>&1)
sleep 2
sudo -n -u astrad env ASTRA_SESSIONS_DIR=$AGENTS/betaco/sessions $ASTRAD session send w1 'pwd; git branch --show-current\r' >/dev/null 2>&1
sleep 1
tail=$(sudo -n -u astrad env ASTRA_SESSIONS_DIR=$AGENTS/betaco/sessions $ASTRAD session tail w1 -for 1s -strip 2>&1)
[[ $tail == *"$wt1"* && $tail == *"$b1"* ]] && pass "the agent's shell starts in the worktree, on the worker's branch" || bad "session: $sock / $tail"
sudo -n -u astrad sudo -n $PRIV session-stop betaco w1 && pass "session stopped" || bad "session-stop"

echo
(( fail )) && { echo "workspace slice: FAILED"; exit 1; }
echo "workspace slice: all checks passed"
