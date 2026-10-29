package remote

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// sshServer is a small SSH server for tests: it accepts one client key, answers exec requests
// from a table, and forwards direct-tcpip channels to the address they ask for.
type sshServer struct {
	addr    string
	hostKey ssh.Signer
	mu      sync.Mutex
	ran     []string
	stdins  []string
}

type reply struct {
	stdout, stderr string
	exit           int
}

func newSSHServer(t *testing.T, clientKey ssh.PublicKey, commands map[string]reply) *sshServer {
	t.Helper()
	_, hostPriv, _ := ed25519.GenerateKey(rand.Reader)
	hostSigner, err := ssh.NewSignerFromKey(hostPriv)
	if err != nil {
		t.Fatal(err)
	}
	s := &sshServer{hostKey: hostSigner}
	cfg := &ssh.ServerConfig{PublicKeyCallback: func(_ ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if bytes.Equal(key.Marshal(), clientKey.Marshal()) {
			return &ssh.Permissions{}, nil
		}
		return nil, errors.New("unknown key")
	}}
	cfg.AddHostKey(hostSigner)
	ln, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s.addr = ln.Addr().String()
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go s.serve(c, cfg, commands)
		}
	}()
	return s
}

func (s *sshServer) serve(c net.Conn, cfg *ssh.ServerConfig, commands map[string]reply) {
	conn, chans, reqs, err := ssh.NewServerConn(c, cfg)
	if err != nil {
		return
	}
	go ssh.DiscardRequests(reqs)
	for ch := range chans {
		switch ch.ChannelType() {
		case "session":
			go s.session(ch, commands)
		case "direct-tcpip":
			go forward(ch)
		default:
			_ = ch.Reject(ssh.UnknownChannelType, "no")
		}
	}
	_ = conn.Close()
}

func (s *sshServer) session(nc ssh.NewChannel, commands map[string]reply) {
	ch, reqs, err := nc.Accept()
	if err != nil {
		return
	}
	defer func() { _ = ch.Close() }()
	for req := range reqs {
		if req.Type != "exec" {
			_ = req.Reply(false, nil)
			continue
		}
		var payload struct{ Command string }
		_ = ssh.Unmarshal(req.Payload, &payload)
		_ = req.Reply(true, nil)
		stdin, _ := io.ReadAll(io.LimitReader(ch, 1<<20))
		s.mu.Lock()
		s.ran = append(s.ran, payload.Command)
		s.stdins = append(s.stdins, string(stdin))
		s.mu.Unlock()
		r, ok := commands[payload.Command]
		if !ok {
			r = reply{stderr: "no such command", exit: 127}
		}
		_, _ = io.WriteString(ch, r.stdout)
		_, _ = io.WriteString(ch.Stderr(), r.stderr)
		_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{uint32(r.exit)}))
		return
	}
}

func forward(nc ssh.NewChannel) {
	var d struct {
		Host     string
		Port     uint32
		OrigHost string
		OrigPort uint32
	}
	_ = ssh.Unmarshal(nc.ExtraData(), &d)
	target, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", net.JoinHostPort(d.Host, fmt.Sprint(d.Port)))
	if err != nil {
		_ = nc.Reject(ssh.ConnectionFailed, err.Error())
		return
	}
	ch, _, err := nc.Accept()
	if err != nil {
		_ = target.Close()
		return
	}
	go func() { _, _ = io.Copy(target, ch); _ = target.Close() }()
	_, _ = io.Copy(ch, target)
	_ = ch.Close()
}

// fixture is a server plus a client key and a known hosts file.
type fixture struct {
	srv    *sshServer
	target Target
	hosts  KnownHosts
	pub    string
}

func newFixture(t *testing.T, commands map[string]reply) fixture {
	t.Helper()
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "id_ed25519")
	pub, err := GenerateKey(keyPath, "test")
	if err != nil {
		t.Fatal(err)
	}
	clientKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pub))
	if err != nil {
		t.Fatal(err)
	}
	srv := newSSHServer(t, clientKey, commands)
	host, portText, _ := net.SplitHostPort(srv.addr)
	var port int
	_, _ = fmt.Sscan(portText, &port)
	return fixture{srv: srv, target: Target{Host: host, Port: port, User: "dev", KeyPath: keyPath}, hosts: KnownHosts{Path: filepath.Join(dir, "known_hosts")}, pub: pub}
}

// trusted dials a fixture's server after trusting its host key, the way a person would.
func (f fixture) trusted(t *testing.T) *Client {
	t.Helper()
	var unknown *UnknownHostError
	if _, err := Dial(t.Context(), f.target, f.hosts); !errors.As(err, &unknown) {
		t.Fatalf("first dial = %v, want an unknown host", err)
	}
	if err := TrustHost(t.Context(), f.target, f.hosts, ssh.FingerprintSHA256(f.srv.hostKey.PublicKey())); err != nil {
		t.Fatal(err)
	}
	c, err := Dial(t.Context(), f.target, f.hosts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestARunReturnsOutputExitCodeAndPassesStdin(t *testing.T) {
	f := newFixture(t, map[string]reply{
		"echo hi":  {stdout: "hi\n"},
		"false":    {exit: 1, stderr: "nope"},
		"cat > /x": {stdout: "stored\n"},
	})
	c := f.trusted(t)
	if res, err := c.Run(t.Context(), "echo hi", nil); err != nil || string(res.Stdout) != "hi\n" || res.Exit != 0 {
		t.Errorf("echo = %+v, %v", res, err)
	}
	res, err := c.Run(t.Context(), "false", nil)
	if err != nil || res.Exit != 1 || string(res.Stderr) != "nope" {
		t.Errorf("a failing command is a result, not an error: %+v, %v", res, err)
	}
	if res, err := c.Run(t.Context(), "cat > /x", strings.NewReader("payload")); err != nil || string(res.Stdout) != "stored\n" {
		t.Errorf("stdin command = %+v, %v", res, err)
	}
	f.srv.mu.Lock()
	defer f.srv.mu.Unlock()
	if f.srv.stdins[2] != "payload" {
		t.Errorf("the server received stdin %q", f.srv.stdins[2])
	}
}

func TestAnUnknownHostStopsWithItsFingerprintUntilAPersonTrustsIt(t *testing.T) {
	f := newFixture(t, nil)
	_, err := Dial(t.Context(), f.target, f.hosts)
	var unknown *UnknownHostError
	if !errors.As(err, &unknown) {
		t.Fatalf("err = %v, want UnknownHostError", err)
	}
	want := ssh.FingerprintSHA256(f.srv.hostKey.PublicKey())
	if unknown.Fingerprint != want || unknown.KeyType != "ssh-ed25519" {
		t.Errorf("unknown = %+v, want fingerprint %s", unknown, want)
	}
	// Nothing was remembered by looking.
	if data, _ := os.ReadFile(f.hosts.Path); len(data) != 0 {
		t.Errorf("the known hosts file was written without consent: %q", data)
	}
	// A fingerprint other than the one the server shows is refused.
	if err := TrustHost(t.Context(), f.target, f.hosts, "SHA256:somethingelse"); err == nil {
		t.Error("a wrong fingerprint was trusted")
	}
	if err := TrustHost(t.Context(), f.target, f.hosts, want); err != nil {
		t.Fatal(err)
	}
	c, err := Dial(t.Context(), f.target, f.hosts)
	if err != nil {
		t.Fatalf("after trusting: %v", err)
	}
	_ = c.Close()
	if err := TrustHost(t.Context(), f.target, f.hosts, "anything"); err != nil {
		t.Errorf("trusting a host that is already known should do nothing: %v", err)
	}
	if info, _ := os.Stat(f.hosts.Path); info.Mode().Perm() != 0o600 {
		t.Errorf("known hosts mode = %v", info.Mode().Perm())
	}
}

func TestAChangedHostKeyIsRefusedAndNeverOverridden(t *testing.T) {
	f := newFixture(t, nil)
	f.trusted(t)
	// Another server answers on the same address: the file still remembers the first key.
	clientKey, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(f.pub))
	other := newSSHServer(t, clientKey, nil)
	host, portText, _ := net.SplitHostPort(other.addr)
	var port int
	_, _ = fmt.Sscan(portText, &port)
	// The file remembers the first server's key for the second server's address.
	if err := os.WriteFile(f.hosts.Path, []byte(knownhosts.Line([]string{hostLabel(host, port)}, f.srv.hostKey.PublicKey())+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := f.target
	target.Port = port
	_, err := Dial(t.Context(), target, f.hosts)
	var changed *ChangedHostKeyError
	if !errors.As(err, &changed) {
		t.Fatalf("err = %v, want ChangedHostKeyError", err)
	}
	if changed.Fingerprint != ssh.FingerprintSHA256(other.hostKey.PublicKey()) || changed.Known != ssh.FingerprintSHA256(f.srv.hostKey.PublicKey()) {
		t.Errorf("changed = %+v", changed)
	}
	if !strings.Contains(changed.Error(), "do not continue") {
		t.Errorf("the message must warn: %v", changed)
	}
	if err := TrustHost(t.Context(), target, f.hosts, changed.Fingerprint); err == nil {
		t.Error("a changed key must not be trusted by TrustHost")
	}
}

func TestAKeyTheServerDoesNotKnowIsAnAuthError(t *testing.T) {
	f := newFixture(t, nil)
	f.trusted(t)
	other := filepath.Join(t.TempDir(), "other")
	if _, err := GenerateKey(other, ""); err != nil {
		t.Fatal(err)
	}
	target := f.target
	target.KeyPath = other
	if _, err := Dial(t.Context(), target, f.hosts); !errors.Is(err, ErrAuth) {
		t.Errorf("err = %v, want ErrAuth", err)
	}
}

func TestKeyFilesAreCheckedBeforeUse(t *testing.T) {
	f := newFixture(t, nil)
	if err := os.Chmod(f.target.KeyPath, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Dial(t.Context(), f.target, f.hosts); err == nil || !strings.Contains(err.Error(), "chmod 600") {
		t.Errorf("a key readable by others: err = %v", err)
	}
	target := f.target
	target.KeyPath = filepath.Join(t.TempDir(), "missing")
	if _, err := Dial(t.Context(), target, f.hosts); err == nil {
		t.Error("a missing key must be an error")
	}
	garbage := filepath.Join(t.TempDir(), "garbage")
	_ = os.WriteFile(garbage, []byte("not a key"), 0o600)
	target.KeyPath = garbage
	if _, err := Dial(t.Context(), target, f.hosts); err == nil {
		t.Error("a file that is not a key must be an error")
	}
}

func TestAPassphraseProtectedKeyAsksForItsPassphrase(t *testing.T) {
	f := newFixture(t, nil)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	block, err := ssh.MarshalPrivateKeyWithPassphrase(priv, "enc", []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "enc")
	var buf bytes.Buffer
	_ = writePEM(&buf, block)
	_ = os.WriteFile(path, buf.Bytes(), 0o600)
	target := f.target
	target.KeyPath = path
	if _, err := Dial(t.Context(), target, f.hosts); err == nil || !strings.Contains(err.Error(), "passphrase") {
		t.Errorf("without a passphrase function: %v", err)
	}
	asked := 0
	target.Passphrase = func() ([]byte, error) { asked++; return []byte("wrong"), nil }
	if _, err := Dial(t.Context(), target, f.hosts); err == nil || asked != 1 {
		t.Errorf("a wrong passphrase: err = %v, asked %d", err, asked)
	}
	target.Passphrase = func() ([]byte, error) { return nil, errors.New("cancelled") }
	if _, err := Dial(t.Context(), target, f.hosts); err == nil || !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("a cancelled prompt: %v", err)
	}
}

func TestGenerateKeyMakesAnOwnerOnlyKeyAndNeverReplacesOne(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys", "astra")
	pub, err := GenerateKey(path, "astra on laptop")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(pub, "ssh-ed25519 ") || !strings.HasSuffix(pub, " astra on laptop\n") {
		t.Errorf("public key = %q", pub)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Errorf("private key mode = %v", info.Mode().Perm())
	}
	if onDisk, _ := os.ReadFile(path + ".pub"); string(onDisk) != pub {
		t.Error("the .pub file differs from what was returned")
	}
	before, _ := os.ReadFile(path)
	if _, err := GenerateKey(path, ""); err == nil {
		t.Error("an existing key was replaced")
	}
	if after, _ := os.ReadFile(path); !bytes.Equal(before, after) {
		t.Error("the existing key changed")
	}
}

func TestForwardCarriesTrafficToTheServersLoopback(t *testing.T) {
	daemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "from the daemon") }))
	t.Cleanup(daemon.Close)
	f := newFixture(t, nil)
	c := f.trusted(t)
	addr, stop, err := c.Forward(t.Context(), "127.0.0.1:0", strings.TrimPrefix(daemon.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+addr+"/", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) != "from the daemon" {
		t.Errorf("through the tunnel: %q", body)
	}
	stop()
	if conn, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", addr); err == nil {
		_ = conn.Close()
		t.Error("the tunnel still listens after stop")
	}
	if _, _, err := c.Forward(t.Context(), "256.0.0.1:0", "x:1"); err == nil {
		t.Error("an invalid listen address must fail")
	}
}

func TestRunStopsWhenItsContextEnds(t *testing.T) {
	f := newFixture(t, map[string]reply{"sleep": {}})
	c := f.trusted(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := c.Run(ctx, "sleep", blockingReader{}); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

// blockingReader never gives any input, so a command waiting for stdin waits for ever.
type blockingReader struct{}

func (blockingReader) Read([]byte) (int, error) { select {} }

func TestUnreachableHostsAreAnError(t *testing.T) {
	f := newFixture(t, nil)
	target := f.target
	ln, _ := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	_, portText, _ := net.SplitHostPort(ln.Addr().String())
	_ = ln.Close()
	_, _ = fmt.Sscan(portText, &target.Port)
	if _, err := Dial(t.Context(), target, f.hosts); err == nil || !strings.Contains(err.Error(), "reach") {
		t.Errorf("err = %v", err)
	}
}
