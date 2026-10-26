#!/usr/bin/env bash
# Prepares a Debian or Ubuntu host for agent isolation (F12-01, ADR 0005): the daemon user,
# one agent user per client workspace, the folders between them, a decoy production secret,
# and the one sudo rule per agent user the daemon needs. Safe to run again.
#
#   sudo ./setup-isolation.sh [workspace-slug ...]      (default: acme)
set -euo pipefail

DAEMON=astrad          # runs the daemon; owns production secrets
STATE=/var/lib/astra

[[ $EUID -eq 0 ]] || { echo "run as root" >&2; exit 1; }
SLUGS=("$@")
(( ${#SLUGS[@]} )) || SLUGS=(acme)

# docker.sock must exist for the "agent cannot reach Docker" check to mean anything.
if ! command -v docker >/dev/null; then
  apt-get update -qq
  DEBIAN_FRONTEND=noninteractive apt-get install -y -qq docker.io >/dev/null
fi

getent group "$DAEMON" >/dev/null || groupadd --system "$DAEMON"
id "$DAEMON" >/dev/null 2>&1 ||
  useradd --system --gid "$DAEMON" --home-dir /var/lib/astrad --create-home --shell /usr/sbin/nologin "$DAEMON"

# Daemon home and decoy production secret: the daemon and nobody else.
chown "$DAEMON:$DAEMON" /var/lib/astrad && chmod 0700 /var/lib/astrad
install -d -o "$DAEMON" -g "$DAEMON" -m 0700 /var/lib/astrad/production
printf 'DATABASE_URL=postgres://acme:DECOY-not-a-real-secret@db.internal/acme\n' >/var/lib/astrad/production/acme.env
chown "$DAEMON:$DAEMON" /var/lib/astrad/production/acme.env && chmod 0600 /var/lib/astrad/production/acme.env
install -d -o root -g "$DAEMON" -m 0750 /etc/astra
printf 'cloudflare_token=DECOY-not-a-real-token\n' >/etc/astra/secrets.env
chown root:"$DAEMON" /etc/astra/secrets.env && chmod 0640 /etc/astra/secrets.env

install -d -o root -g root -m 0755 "$STATE" "$STATE/agents"

# One agent user per client workspace. Workspaces cannot see each other (F2-01), because
# every folder below belongs to its own agent user and to no shared group.
for slug in "${SLUGS[@]}"; do
  [[ $slug =~ ^[a-z][a-z0-9-]{0,18}[a-z0-9]$ ]] || { echo "bad workspace slug: $slug" >&2; exit 1; }
  agent="astra-agent-$slug"
  base="$STATE/agents/$slug"

  getent group "$agent" >/dev/null || groupadd --system "$agent"
  id "$agent" >/dev/null 2>&1 ||
    useradd --system --gid "$agent" --home-dir "$base/home" --no-create-home --shell /bin/bash "$agent"
  # An agent user must not be in docker, sudo, adm or the daemon's group, however it was made.
  for g in docker sudo adm "$DAEMON"; do
    if id -nG "$agent" | tr ' ' '\n' | grep -qx "$g"; then gpasswd -d "$agent" "$g" >/dev/null; fi
  done

  # The daemon's group may walk through $base to the sessions folder, and nothing else:
  # home and work stay unreadable to it, which keeps agent logins out of its reach (F3-03).
  install -d -o "$agent" -g "$DAEMON" -m 0710 "$base"
  # The agent owns the sockets and the daemon's group may use them; setgid makes every
  # socket the host creates belong to that group.
  install -d -o "$agent" -g "$DAEMON" -m 2770 "$base/sessions"
  install -d -o "$agent" -g "$agent" -m 0700 "$base/work" "$base/home"

  # The daemon may run commands as this agent user, and only as this user.
  cat >"/etc/sudoers.d/astrad-agent-$slug" <<EOF
# Managed by scripts/vm/setup-isolation.sh
$DAEMON ALL=($agent) NOPASSWD: ALL
Defaults:$DAEMON !requiretty
EOF
  chmod 0440 "/etc/sudoers.d/astrad-agent-$slug"
  visudo -cf "/etc/sudoers.d/astrad-agent-$slug" >/dev/null
done

# Leftovers of the first single-agent layout, if this host had it.
rm -f /etc/sudoers.d/astrad
if id astra-agent >/dev/null 2>&1; then userdel astra-agent 2>/dev/null || true; fi
rm -rf "$STATE/sessions" "$STATE/work" "$STATE/agent-home"

echo "isolation set up for: ${SLUGS[*]}"
id "$DAEMON"
for slug in "${SLUGS[@]}"; do id "astra-agent-$slug"; done
ls -ld /var/lib/astrad /var/lib/astrad/production /etc/astra "$STATE"/agents/* "$STATE"/agents/*/* /var/run/docker.sock 2>/dev/null
