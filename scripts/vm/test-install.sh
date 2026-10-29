#!/usr/bin/env bash
# M1 slice 6 (F1-02 to F1-07, ADR 0007), run on the Mac against the Linux test VM over real SSH:
# a new SSH key, host key trust, the readiness check, installing the daemon with the astra CLI,
# pairing without ever showing a token, the tunnel, the web UI and the first-run checklist through
# it, an update, a release download with a tampered file, a changed host key, and the service
# coming back after a reboot. The VM is cleaned of earlier installs first and afterwards.
#
#   scripts/vm/test-install.sh
set -uo pipefail

REPO=$(cd "$(dirname "$0")/../.." && pwd)
VM="$REPO/scripts/vm/ssh.sh"
WORK=$(mktemp -d /tmp/astra-install-test.XXXXXX)
export ASTRA_CONFIG_DIR=$WORK/cfg
ASTRA=$WORK/astra
DIST=$WORK/dist
PORT_TUNNEL=17437
fail=0

pass() { printf '  ok    %s\n' "$1"; }
bad()  { printf '  FAIL  %s\n' "$1"; fail=1; }
field() { python3 -c "import json,sys; print(json.load(sys.stdin)$1)"; }
device_token() { python3 -c "import json; print(json.load(open('$ASTRA_CONFIG_DIR/servers.json'))[0].get('deviceToken',''))"; }
api() { curl -s -m 10 -H "Authorization: Bearer $(device_token)" "http://127.0.0.1:$PORT_TUNNEL$1"; }

vm_clean() {
  $VM 'sudo systemctl stop astrad 2>/dev/null; sudo systemctl disable astrad 2>/dev/null
       for u in $(systemctl list-units --plain --no-legend "astra-s-*" | cut -d" " -f1); do sudo systemctl stop "$u"; done
       sudo rm -f /etc/systemd/system/astrad.service /etc/systemd/system/astrad-test.service /etc/sudoers.d/astrad* /etc/sudoers.d/astra-agent-* /usr/local/bin/astrad* /usr/libexec/astra/astra-priv*
       for u in $(getent passwd | cut -d: -f1 | grep "^astra-agent-"); do sudo userdel "$u"; done
       sudo userdel astrad 2>/dev/null; sudo groupdel astrad 2>/dev/null
       for g in $(getent group | cut -d: -f1 | grep "^astra-agent-"); do sudo groupdel "$g"; done
       sudo rm -rf /var/lib/astra /var/lib/astrad /etc/astra; sudo systemctl daemon-reload' >/dev/null 2>&1
}
cleanup() {
  [[ -n ${TUNNEL_PID:-} ]] && kill "$TUNNEL_PID" 2>/dev/null
  [[ -n ${HTTP_PID:-} ]] && kill "$HTTP_PID" 2>/dev/null
  vm_clean
  $VM "sed -i '/astra-test-key/d' ~/.ssh/authorized_keys" >/dev/null 2>&1
  rm -rf "$WORK"
}
trap cleanup EXIT

echo "== build"
(cd "$REPO" && go build -o "$ASTRA" ./cmd/astra) || { echo "build failed"; exit 1; }
mkdir -p "$DIST"
for c in astrad astra-priv; do
  (cd "$REPO" && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o "$DIST/$c-linux-arm64" ./cmd/$c) || exit 1
done
[[ -f $REPO/web/dist/index.html ]] || (cd "$REPO/web" && npm run build >/dev/null 2>&1)
cp -R "$REPO/web/dist" "$DIST/web"
vm_clean
pass "binaries and the web UI are built; the VM has no earlier install"

echo "== an SSH key (F1-02)"
$ASTRA key new vm 2>"$WORK/key.err" >"$WORK/vm.pub" && pass "astra key new made a key" || bad "key new"
[[ $(stat -f %Lp "$ASTRA_CONFIG_DIR/keys/vm" 2>/dev/null || stat -c %a "$ASTRA_CONFIG_DIR/keys/vm") == 600 ]] && pass "the private key is mode 600" || bad "key mode"
grep -q '^ssh-ed25519 ' "$WORK/vm.pub" && pass "the public key to add to the server was printed" || bad "public key output"
$ASTRA key new vm >/dev/null 2>&1 && bad "an existing key was replaced" || pass "an existing key is never replaced"
{ cat "$WORK/vm.pub"; echo; } | sed 's/ astra$/ astra-test-key/' | $VM 'cat >> ~/.ssh/authorized_keys'

echo "== adding the server and trusting its host key"
mkdir -p "$WORK/cfg2" && cp -R "$ASTRA_CONFIG_DIR/keys" "$WORK/cfg2/"
echo n | ASTRA_CONFIG_DIR=$WORK/cfg2 $ASTRA server add vm --host 127.0.0.1 --port 2222 --user dev --key "$WORK/cfg2/keys/vm" >/dev/null 2>"$WORK/trust.err" \
  && bad "a host was added without trusting it" || pass "answering no adds nothing"
grep -q "SHA256:" "$WORK/trust.err" && pass "the host key's fingerprint was shown for comparison" || bad "no fingerprint shown"
[[ ! -s $WORK/cfg2/known_hosts ]] && pass "nothing was remembered" || bad "known_hosts written without consent"
echo y | $ASTRA server add vm --host 127.0.0.1 --port 2222 --user dev --key "$ASTRA_CONFIG_DIR/keys/vm" >/dev/null 2>&1 \
  && pass "answering yes adds the server" || bad "server add"
$ASTRA server list | grep -q "^vm	dev@127.0.0.1:2222	all-in-one	not paired" && pass "it is listed, not paired yet" || bad "server list"

echo "== readiness (F1-03)"
check=$($ASTRA server check vm 2>&1); rc=$?
[[ $rc == 0 ]] && pass "the VM passes the check" || bad "check exit $rc: $check"
for want in "Operating system" "systemd" "Resource limits" "Administrator access" "Memory" "git"; do
  grep -q "$want" <<<"$check" && continue; bad "the report has no $want"
done
echo "$check" | sed 's/^/        /' | head -14

echo "== install (F1-04)"
install_out=$($ASTRA server install vm --binary-dir "$DIST" --yes 2>&1); rc=$?
[[ $rc == 0 ]] && pass "astra server install finished" || { bad "install exit $rc"; echo "$install_out" | tail -8; }
grep -q "There is no SHA256SUMS file" <<<"$install_out" && pass "it said where the checksums came from" || bad "no checksum note"
[[ $($VM 'systemctl is-active astrad') == active ]] && pass "astrad.service is active" || bad "service not active"
[[ $($VM 'systemctl is-enabled astrad') == enabled ]] && pass "and enabled for every boot" || bad "service not enabled"
[[ $($VM 'ps -o user= -p $(systemctl show astrad -p MainPID --value)') == astrad ]] && pass "the daemon runs as the astrad user, not root" || bad "daemon user"
$VM 'sudo visudo -c' >/dev/null 2>&1 && pass "sudoers is valid" || bad "sudoers invalid"
[[ $($VM 'sudo cat /etc/sudoers.d/astrad-priv') == *"/usr/libexec/astra/astra-priv"* ]] && pass "astrad may run astra-priv, and only that" || bad "sudoers rule"
[[ $($VM 'stat -c %a:%U /var/lib/astrad /usr/libexec/astra/astra-priv | tr "\n" " "') == "700:astrad 755:root " ]] && pass "folders and the helper have the right owners and modes" || bad "owners: $($VM 'stat -c %a:%U /var/lib/astrad /usr/libexec/astra/astra-priv')"

echo "== pairing without showing a token (F1-05)"
pair_out=$($ASTRA server pair vm 2>&1); rc=$?
[[ $rc == 0 ]] && pass "astra server pair finished" || bad "pair exit $rc: $pair_out"
tok=$(device_token)
[[ -n $tok ]] && pass "the device token is kept for this laptop" || bad "no device token saved"
grep -q "$tok" <<<"$pair_out" && bad "the device token was printed" || pass "no token was printed"
[[ $(stat -f %Lp "$ASTRA_CONFIG_DIR/servers.json" 2>/dev/null || stat -c %a "$ASTRA_CONFIG_DIR/servers.json") == 600 ]] && pass "servers.json is readable only by its owner" || bad "servers.json mode"
$VM 'sudo journalctl -u astrad --no-pager -o cat' | grep -q "$tok" && bad "the token reached the journal" || pass "no token in the daemon's journal"
$ASTRA server list | grep -q "	paired$" && pass "the server now reads as paired" || bad "not paired in the list"

echo "== the tunnel, the web UI and the checklist (N-12, F1-06, F1-07)"
$ASTRA server tunnel vm --port $PORT_TUNNEL >/dev/null 2>&1 &
TUNNEL_PID=$!
for _ in $(seq 20); do curl -sf "http://127.0.0.1:$PORT_TUNNEL/api/v1/version" >/dev/null 2>&1 && break; sleep 0.5; done
[[ $(curl -s "http://127.0.0.1:$PORT_TUNNEL/api/v1/version" | field '["api"]') == 1 ]] && pass "the daemon answers through the SSH tunnel" || bad "no answer through the tunnel"
curl -s "http://127.0.0.1:$PORT_TUNNEL/" | grep -q '<div id="root">' && pass "the web UI is served from the same address" || bad "no web UI"
curl -sI "http://127.0.0.1:$PORT_TUNNEL/" | grep -qi "^content-security-policy:" && pass "with a content security policy" || bad "no CSP"
[[ $(curl -s -o /dev/null -w '%{http_code}' "http://127.0.0.1:$PORT_TUNNEL/api/v1/workspaces") == 401 ]] && pass "the API still needs the device token" || bad "API open without a token"
[[ $(api /api/v1/workspaces) == "[]" ]] && pass "the device token is accepted" || bad "device token rejected: $(api /api/v1/workspaces)"
[[ $(api /api/v1/checklist | field '["percent"]') == 33 ]] && pass "the checklist starts at 33% (the server is the only done step)" || bad "checklist: $(api /api/v1/checklist)"
curl -s -X POST -H "Authorization: Bearer $tok" -d '{"slug":"warung-sari","name":"Warung Sari"}' "http://127.0.0.1:$PORT_TUNNEL/api/v1/workspaces" >/dev/null
[[ $(curl -s -o /dev/null -w '%{http_code}' -X POST -H "Authorization: Bearer $tok" -d '{"workspace":"warung-sari","name":"web","repoUrl":""}' "http://127.0.0.1:$PORT_TUNNEL/api/v1/projects") == 201 ]] && pass "a workspace and an empty project were made through astra-priv" || bad "project creation"
checklist=$(api /api/v1/checklist)
[[ $(echo "$checklist" | field '["percent"]') == 66 ]] && pass "the checklist moved to 66% from the facts" || bad "checklist after a project: $checklist"
echo "$checklist" | python3 -c "import json,sys; d=json.load(sys.stdin); print('        agent:', [i['detail'] for i in d['items'] if i['id']=='agent'][0])"

echo "== an update (F1-04, §8.5)"
update_out=$($ASTRA server install vm --binary-dir "$DIST" --yes 2>&1); rc=$?
[[ $rc == 0 ]] && pass "installing again works" || bad "update exit $rc"
grep -q "installed already" <<<"$update_out" && pass "it said what is replaced" || bad "no update note"
$VM 'test -f /usr/local/bin/astrad.previous && test -f /usr/libexec/astra/astra-priv.previous' && pass "the previous binaries are kept" || bad "no .previous files"
[[ $($VM 'systemctl is-active astrad') == active ]] && pass "the daemon is running after the update" || bad "daemon down after the update"
[[ $(api /api/v1/workspaces | field '[0]["slug"]') == warung-sari ]] && pass "the data survived, and the tunnel and token still work" || bad "data after update"

echo "== a release downloaded on the server, and a tampered one"
REL=$WORK/release; STAGE=$WORK/stage; mkdir -p "$REL" "$STAGE"
cp "$DIST/astrad-linux-arm64" "$STAGE/astrad"; cp "$DIST/astra-priv-linux-arm64" "$STAGE/astra-priv"; cp -R "$DIST/web" "$STAGE/web"
tar -czf "$REL/astra-linux-arm64.tar.gz" -C "$STAGE" astrad astra-priv web
(cd "$REL" && shasum -a 256 astra-linux-arm64.tar.gz > SHA256SUMS)
pkill -f "http.server 18080" 2>/dev/null
(cd "$REL" && exec python3 -m http.server 18080 --bind 0.0.0.0 >/dev/null 2>&1) &
HTTP_PID=$!
sleep 1
$ASTRA server install vm --release-url http://10.0.2.2:18080 --yes >/dev/null 2>&1 && bad "plain http was accepted without opting in" || pass "an http release URL is refused without --insecure-http"
rel_out=$($ASTRA server install vm --release-url http://10.0.2.2:18080 --insecure-http --yes 2>&1); rc=$?
[[ $rc == 0 ]] && pass "installing from a release URL works" || { bad "release install exit $rc"; echo "$rel_out" | tail -5; }
before=$($VM 'sha256sum /usr/local/bin/astrad | cut -d" " -f1')
printf 'tampered' >> "$REL/astra-linux-arm64.tar.gz"
tamper_out=$($ASTRA server install vm --release-url http://10.0.2.2:18080 --insecure-http --yes 2>&1); rc=$?
[[ $rc != 0 ]] && grep -q "nothing was installed" <<<"$tamper_out" && pass "a tampered release is refused: nothing was installed" || bad "tampered release: exit $rc"
[[ $($VM 'sha256sum /usr/local/bin/astrad | cut -d" " -f1') == "$before" ]] && pass "the installed daemon is untouched" || bad "the daemon changed"
left=$($VM 'ls /tmp | grep astra-install' 2>/dev/null)
[[ -z $left ]] && pass "no temporary files are left on the server" || bad "temp files left: $left"

echo "== a changed host key is refused"
cp "$ASTRA_CONFIG_DIR/known_hosts" "$WORK/known_hosts.good"
ssh-keygen -q -t ed25519 -N "" -f "$WORK/other" && printf '[127.0.0.1]:2222 %s\n' "$(cut -d' ' -f1,2 "$WORK/other.pub")" > "$ASTRA_CONFIG_DIR/known_hosts"
changed=$(echo y | $ASTRA server check vm 2>&1); rc=$?
[[ $rc != 0 ]] && grep -q "changed" <<<"$changed" && grep -q "do not continue" <<<"$changed" && pass "a different host key stops everything, with a warning" || bad "changed key: exit $rc $changed"
grep -q "Trust this server" <<<"$changed" && bad "it offered to trust the changed key" || pass "it never offers to trust a changed key"
cp "$WORK/known_hosts.good" "$ASTRA_CONFIG_DIR/known_hosts"

echo "== the service comes back after a reboot (F1-04)"
$VM 'sudo systemctl reboot' >/dev/null 2>&1
sleep 5
for _ in $(seq 60); do $VM 'true' >/dev/null 2>&1 && break; sleep 2; done
for _ in $(seq 30); do [[ $($VM 'systemctl is-active astrad' 2>/dev/null) == active ]] && break; sleep 2; done
[[ $($VM 'systemctl is-active astrad' 2>/dev/null) == active ]] && pass "astrad is running again after the VM rebooted" || bad "daemon not running after reboot"
kill "$TUNNEL_PID" 2>/dev/null; wait "$TUNNEL_PID" 2>/dev/null
$ASTRA server tunnel vm --port $PORT_TUNNEL >/dev/null 2>&1 &
TUNNEL_PID=$!
for _ in $(seq 20); do [[ $(api /api/v1/workspaces 2>/dev/null | field '[0]["slug"]' 2>/dev/null) == warung-sari ]] && break; sleep 1; done
[[ $(api /api/v1/workspaces | field '[0]["slug"]' 2>/dev/null) == warung-sari ]] && pass "with its data, through a new tunnel connection" || bad "no data after reboot"

echo
(( fail )) && { echo "install slice: FAILED"; exit 1; }
echo "install slice: all checks passed"
