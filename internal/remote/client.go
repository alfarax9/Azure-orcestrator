package remote

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// maxOutput bounds what a remote command may print back.
const maxOutput = 4 << 20

// Target is a server to connect to.
type Target struct {
	Host    string
	Port    int
	User    string
	KeyPath string
	// Passphrase is asked for only when the key file is encrypted.
	Passphrase func() ([]byte, error)
}

// Result is what a remote command did. A command that exits non-zero is a Result with Exit set,
// not an error: whether that is a failure depends on the command.
type Result struct {
	Stdout []byte
	Stderr []byte
	Exit   int
}

// ErrAuth means the server refused the key.
var ErrAuth = errors.New("the server refused this key")

// Client is a connection to one server.
type Client struct {
	c      *ssh.Client
	target Target
}

// connection is a host key the server offered, remembered so a person can trust it.
type connection struct {
	key ssh.PublicKey
}

// Dial connects to the server. An unknown host gives *UnknownHostError and a changed one
// *ChangedHostKeyError; see Trust for what to do about the first.
func Dial(ctx context.Context, t Target, hosts KnownHosts) (*Client, error) {
	client, _, err := dial(ctx, t, hosts)
	return client, err
}

func dial(ctx context.Context, t Target, hosts KnownHosts) (*Client, *connection, error) {
	signer, err := loadKey(t)
	if err != nil {
		return nil, nil, err
	}
	check, err := hosts.callback()
	if err != nil {
		return nil, nil, err
	}
	seen := &connection{}
	cfg := &ssh.ClientConfig{
		User: t.User,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: func(host string, remote net.Addr, key ssh.PublicKey) error {
			seen.key = key
			return check(host, remote, key)
		},
		Timeout: 15 * time.Second,
	}
	addr := net.JoinHostPort(t.Host, fmt.Sprint(t.Port))
	conn, err := (&net.Dialer{Timeout: cfg.Timeout}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("reach %s: %w", addr, err)
	}
	sc, chans, reqs, err := ssh.NewClientConn(conn, addr, cfg)
	if err != nil {
		_ = conn.Close()
		return nil, seen, classify(err)
	}
	return &Client{c: ssh.NewClient(sc, chans, reqs), target: t}, seen, nil
}

// classify keeps the typed host key errors, and names a refused key.
func classify(err error) error {
	var unknown *UnknownHostError
	var changed *ChangedHostKeyError
	switch {
	case errors.As(err, &unknown):
		return unknown
	case errors.As(err, &changed):
		return changed
	case strings.Contains(err.Error(), "unable to authenticate"):
		return fmt.Errorf("%w: add the public key to ~/.ssh/authorized_keys of the user on the server", ErrAuth)
	}
	return fmt.Errorf("connect: %w", err)
}

// TrustHost connects once only to read the host key, and remembers it when the caller
// has confirmed the fingerprint it was shown by the UnknownHostError.
func TrustHost(ctx context.Context, t Target, hosts KnownHosts, confirmed string) error {
	_, seen, err := dial(ctx, t, hosts)
	var unknown *UnknownHostError
	if !errors.As(err, &unknown) {
		if err == nil {
			return nil // already known
		}
		return err
	}
	if unknown.Fingerprint != confirmed || seen.key == nil {
		return fmt.Errorf("the server now offers %s, not the fingerprint you confirmed; nothing was trusted", unknown.Fingerprint)
	}
	return hosts.Trust(hostLabel(t.Host, t.Port), seen.key)
}

func loadKey(t Target) (ssh.Signer, error) {
	info, err := os.Stat(t.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("read the SSH key: %w", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("the SSH key %s is readable by others (mode %o); run chmod 600 on it", t.KeyPath, info.Mode().Perm())
	}
	pem, err := os.ReadFile(t.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("read the SSH key: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(pem)
	var missing *ssh.PassphraseMissingError
	if errors.As(err, &missing) {
		if t.Passphrase == nil {
			return nil, errors.New("the SSH key is protected by a passphrase")
		}
		pass, perr := t.Passphrase()
		if perr != nil {
			return nil, perr
		}
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pem, pass)
	}
	if err != nil {
		return nil, fmt.Errorf("parse the SSH key: %w", err)
	}
	return signer, nil
}

// Run executes cmd on the server with stdin, and returns what it printed. A cancelled ctx ends
// the command.
func (c *Client) Run(ctx context.Context, cmd string, stdin io.Reader) (Result, error) {
	sess, err := c.c.NewSession()
	if err != nil {
		return Result{}, fmt.Errorf("open a session: %w", err)
	}
	defer func() { _ = sess.Close() }()
	var out, errOut limited
	sess.Stdout, sess.Stderr, sess.Stdin = &out, &errOut, stdin
	done := make(chan error, 1)
	go func() { done <- sess.Run(cmd) }()
	var runErr error
	select {
	case runErr = <-done:
	case <-ctx.Done():
		_ = sess.Close()
		return Result{}, ctx.Err()
	}
	res := Result{Stdout: out.buf.Bytes(), Stderr: errOut.buf.Bytes()}
	var exit *ssh.ExitError
	switch {
	case runErr == nil:
	case errors.As(runErr, &exit):
		res.Exit = exit.ExitStatus()
	default:
		return res, fmt.Errorf("run on the server: %w", runErr)
	}
	return res, nil
}

// Forward listens on a loopback address and passes every connection to remote (an address
// as the server sees it), so a client on this machine reaches the daemon (N-12). It returns the
// local address and a function that stops it, which also ends the connections still open.
func (c *Client) Forward(ctx context.Context, listen, remote string) (string, func(), error) {
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", listen)
	if err != nil {
		return "", nil, fmt.Errorf("listen on %s: %w", listen, err)
	}
	t := &tunnel{c: c, remote: remote, open: map[net.Conn]struct{}{}}
	t.wg.Go(func() {
		for {
			local, err := ln.Accept()
			if err != nil {
				return
			}
			t.wg.Go(func() { t.pipe(local) })
		}
	})
	return ln.Addr().String(), func() {
		_ = ln.Close()
		t.closeAll()
		t.wg.Wait()
	}, nil
}

// tunnel is one running Forward: the connections it carries, so stopping it can end them.
type tunnel struct {
	c      *Client
	remote string
	wg     sync.WaitGroup
	mu     sync.Mutex // guards open and stopped
	open   map[net.Conn]struct{}
	closed bool
}

// track remembers a connection; false once the tunnel is stopping.
func (t *tunnel) track(c net.Conn) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return false
	}
	t.open[c] = struct{}{}
	return true
}

func (t *tunnel) closeAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	for c := range t.open {
		_ = c.Close()
	}
}

func (t *tunnel) pipe(local net.Conn) {
	defer func() { _ = local.Close() }()
	far, err := t.c.c.Dial("tcp", t.remote)
	if err != nil {
		return
	}
	defer func() { _ = far.Close() }()
	if !t.track(local) || !t.track(far) {
		return
	}
	var wg sync.WaitGroup
	wg.Go(func() { _, _ = io.Copy(far, local); _ = far.Close() })
	_, _ = io.Copy(local, far)
	_ = local.Close()
	wg.Wait()
}

// Close ends the connection.
func (c *Client) Close() error { return c.c.Close() }

// limited keeps at most maxOutput bytes; the rest is dropped, and a command that prints more
// still runs to its end.
type limited struct{ buf bytes.Buffer }

func (l *limited) Write(p []byte) (int, error) {
	if room := maxOutput - l.buf.Len(); room > 0 {
		l.buf.Write(p[:min(len(p), room)])
	}
	return len(p), nil
}
