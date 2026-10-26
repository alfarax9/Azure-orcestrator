#!/usr/bin/env bash
# Prepares a Debian or Ubuntu host for spike S3 (agent isolation, F12-01): two system users,
# the folders between them, a decoy production secret, and the one sudo rule the daemon needs.
# Safe to run again. Run as root: sudo ./setup-isolation.sh
set -euo pipefail

DAEMON=astrad          # runs the daemon; owns production secrets
AGENT=astra-agent      # runs every agent CLI; owns nothing else
STATE=/var/lib/astra

[[ $EUID -eq 0 ]] || { echo "run as root" >&2; exit 1; }

# docker.sock must exist for the "agent cannot reach Docker" check to mean anything.
if ! command -v docker >/dev/null; then
  apt-get update -qq
  DEBIAN_FRONTEND=noninteractive apt-get install -y -qq docker.io >/dev/null
fi

getent group "$DAEMON" >/dev/null || groupadd --system "$DAEMON"
id "$DAEMON" >/dev/null 2>&1 ||
  useradd --system --gid "$DAEMON" --home-dir /var/lib/astrad --create-home --shell /usr/sbin/nologin "$DAEMON"
getent group "$AGENT" >/dev/null || groupadd --system "$AGENT"
id "$AGENT" >/dev/null 2>&1 ||
  useradd --system --gid "$AGENT" --home-dir "$STATE/agent-home" --create-home --shell /bin/bash "$AGENT"

# The agent must not be in docker, sudo, adm or the daemon's group, however it was created.
for g in docker sudo adm "$DAEMON"; do
  if id -nG "$AGENT" | tr ' ' '\n' | grep -qx "$g"; then gpasswd -d "$AGENT" "$g" >/dev/null; fi
done

# Daemon home and decoy production secret: the daemon and nobody else.
chown "$DAEMON:$DAEMON" /var/lib/astrad && chmod 0700 /var/lib/astrad
install -d -o "$DAEMON" -g "$DAEMON" -m 0700 /var/lib/astrad/production
install -m 0600 -o "$DAEMON" -g "$DAEMON" /dev/null /var/lib/astrad/production/acme.env
printf 'DATABASE_URL=postgres://acme:DECOY-not-a-real-secret@db.internal/acme\n' >/var/lib/astrad/production/acme.env
install -d -o root -g "$DAEMON" -m 0750 /etc/astra
printf 'cloudflare_token=DECOY-not-a-real-token\n' >/etc/astra/secrets.env
chown root:"$DAEMON" /etc/astra/secrets.env && chmod 0640 /etc/astra/secrets.env

# Where sessions live. The agent owns the sockets and the daemon's group may use them; the
# setgid bit makes every socket the host creates belong to that group.
install -d -o root -g root -m 0755 "$STATE"
install -d -o "$AGENT" -g "$DAEMON" -m 2770 "$STATE/sessions"
install -d -o "$AGENT" -g "$AGENT" -m 0700 "$STATE/work"
# The agent's home and work folders are its own; the daemon's group cannot enter them.
chown "$AGENT:$AGENT" "$STATE/agent-home" && chmod 0700 "$STATE/agent-home"

# The daemon may run commands as the agent user, and only as that user.
cat >/etc/sudoers.d/astrad <<EOF
# Managed by scripts/vm/setup-isolation.sh
$DAEMON ALL=($AGENT) NOPASSWD: ALL
Defaults:$DAEMON !requiretty
EOF
chmod 0440 /etc/sudoers.d/astrad
visudo -cf /etc/sudoers.d/astrad >/dev/null

echo "isolation set up: daemon=$DAEMON agent=$AGENT state=$STATE"
id "$DAEMON" "$AGENT"
ls -ld /var/lib/astrad /var/lib/astrad/production /etc/astra "$STATE/sessions" "$STATE/work" /var/run/docker.sock 2>/dev/null
