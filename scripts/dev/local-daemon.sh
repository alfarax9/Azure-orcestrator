#!/usr/bin/env bash
# Runs a daemon on this machine for working on the web UI (npm run dev in web/): a throwaway
# data folder, the API on 127.0.0.1:7437, one workspace and project to start tasks in, and a
# pairing token printed once. Agents run as you, in worktrees under the data folder.
#
#   scripts/dev/local-daemon.sh            a stand-in agent (no login, no quota)
#   REAL_AGENT=1 scripts/dev/local-daemon.sh   the claude on your PATH, with your own login
#   NO_SEED=1 scripts/dev/local-daemon.sh      no workspace or project yet (the first-run checklist)
#
# Stop it with Ctrl-C. ASTRA_DEV_DIR changes where the data goes (default /tmp/astra-dev).
set -euo pipefail

here=$(cd "$(dirname "$0")" && pwd)
root=${ASTRA_DEV_DIR:-/tmp/astra-dev}
mkdir -p "$root/bin" "$root/home"

(cd "$here/../.." && go build -o "$root/astrad" ./cmd/astrad)

if [[ ${REAL_AGENT:-0} != 1 ]]; then
  install -m 0755 "$here/fake-claude.sh" "$root/bin/claude"
  export PATH="$root/bin:$PATH" HOME="$root/home"
fi

export ASTRA_LOCAL=1 ASTRA_DB="$root/astra.db" ASTRA_AGENTS_DIR="$root/agents"
# The web dev server proxies to the daemon and passes the browser's Origin on (web/vite.config.ts).
export ASTRA_ORIGINS=http://localhost:5173

# A first workspace and project, so New task has somewhere to go. Running this again keeps them.
# NO_SEED=1 starts with nothing, the way a new user sees the first-run checklist.
if [[ ${NO_SEED:-0} != 1 ]]; then
  "$root/astrad" workspace add dev "Dev" >/dev/null 2>&1 || true
  "$root/astrad" project add dev site >/dev/null 2>&1 || true
fi

echo "Paste this pairing token into the web page (valid 10 minutes, works once):" >&2
"$root/astrad" pair >&2
echo "Data in $root. API on http://127.0.0.1:7437; web UI: cd web && npm run dev" >&2
exec "$root/astrad" serve
