package server

// The privileged steps of an installation, as the scripts `sudo sh` reads. Each is safe to run
// again, which is what makes an update the same as a first install. @DIR@ is the temporary
// folder on the server; everything else is fixed text.

// installBinaries puts the daemon and the root helper in place, keeping what was there.
const installBinaries = `install -d -m 0755 /usr/libexec/astra
for f in /usr/local/bin/astrad /usr/libexec/astra/astra-priv; do
  if [ -f "$f" ]; then cp -p "$f" "$f.previous"; fi
done
install -m 0755 -o root -g root @DIR@/astrad /usr/local/bin/astrad.new
mv -f /usr/local/bin/astrad.new /usr/local/bin/astrad
install -m 0755 -o root -g root @DIR@/astra-priv /usr/libexec/astra/astra-priv.new
mv -f /usr/libexec/astra/astra-priv.new /usr/libexec/astra/astra-priv
`

// createUser makes the daemon's own user, which owns the database and the secrets, and the
// folders the agent users' workspaces live in (ADR 0005).
const createUser = `getent group astrad >/dev/null || groupadd --system astrad
id astrad >/dev/null 2>&1 || useradd --system --gid astrad --home-dir /var/lib/astrad --create-home --shell /usr/sbin/nologin astrad
install -d -o astrad -g astrad -m 0700 /var/lib/astrad
install -d -o root -g root -m 0755 /var/lib/astra /var/lib/astra/agents
install -d -o root -g astrad -m 0750 /etc/astra
`

// allowHelper lets the daemon user call astra-priv as root. The rule is checked by visudo
// before it goes live, so a mistake can never lock anyone out of sudo.
const allowHelper = `tmp=/etc/sudoers.d/.astrad-priv.tmp
printf '%s\n' 'astrad ALL=(root) NOPASSWD: /usr/libexec/astra/astra-priv' > "$tmp"
chmod 0440 "$tmp"
if ! visudo -cf "$tmp" >/dev/null; then rm -f "$tmp"; echo "visudo rejected the rule" >&2; exit 1; fi
mv -f "$tmp" /etc/sudoers.d/astrad-priv
`

const installWeb = `rm -rf /var/lib/astra/web.new
mkdir -p /var/lib/astra/web.new
if [ -d @DIR@/web ]; then cp -R @DIR@/web/. /var/lib/astra/web.new/; fi
chown -R root:root /var/lib/astra/web.new
chmod -R go-w /var/lib/astra/web.new
rm -rf /var/lib/astra/web
mv /var/lib/astra/web.new /var/lib/astra/web
`

const installDocker = `if ! command -v docker >/dev/null 2>&1; then
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -qq
  apt-get install -y -qq docker.io
fi
systemctl enable --now docker
`

// startDaemon enables the service for every boot and starts it, or restarts it on an update.
// Agent sessions are units of their own, so a restart does not touch them (ADR 0005).
const startDaemon = `systemctl enable astrad
systemctl restart astrad
`

// serviceUnit is the daemon's unit. It is not hardened with NoNewPrivileges, because the daemon
// has to reach astra-priv through sudo.
const serviceUnit = `[Unit]
Description=Astra daemon
After=network-online.target
Wants=network-online.target

[Service]
User=astrad
Environment=ASTRA_DB=/var/lib/astrad/astra.db
@WEB@ExecStart=/usr/local/bin/astrad serve
Restart=on-failure
RestartSec=2

[Install]
WantedBy=multi-user.target
`

// waitScript waits up to 30 seconds for the daemon to answer, and prints what it said.
const waitScript = `for i in $(seq 1 30); do
  if out=$(curl -fsS http://127.0.0.1:7437/api/v1/version 2>/dev/null); then echo "$out"; exit 0; fi
  sleep 1
done
echo "the daemon did not answer on 127.0.0.1:7437 within 30 seconds" >&2
journalctl -u astrad -n 20 --no-pager >&2 2>/dev/null || true
exit 1
`
