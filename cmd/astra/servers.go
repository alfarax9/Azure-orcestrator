package main

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"github.com/alfarax9/astra/internal/remote"
	"github.com/alfarax9/astra/internal/server"
)

const daemonAddr = "127.0.0.1:7437"

func (a *app) serverCommand(args []string) error {
	if len(args) == 0 {
		return errUsage
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	switch args[0] {
	case "list":
		return a.list()
	case "add", "remove", "check", "install", "pair", "tunnel":
		if len(args) < 2 {
			return errUsage
		}
	default:
		return errUsage
	}
	name, rest := args[1], args[2:]
	switch args[0] {
	case "add":
		return a.add(ctx, name, rest)
	case "remove":
		return a.registry.Remove(name)
	case "check":
		return a.check(ctx, name, rest)
	case "install":
		return a.installOn(ctx, name, rest)
	case "pair":
		return a.pair(ctx, name, rest)
	}
	return a.tunnel(ctx, name, rest)
}

func (a *app) list() error {
	servers, err := a.registry.List()
	if err != nil {
		return err
	}
	if len(servers) == 0 {
		a.errln("No servers yet. Add one with: astra server add <name> --host <host> --user <user>")
	}
	for _, s := range servers {
		paired := "not paired"
		if s.Paired() {
			paired = "paired"
		}
		a.outf("%s\t%s@%s:%d\t%s\t%s\n", s.Name, s.User, s.Host, s.Port, s.Role, paired)
	}
	return nil
}

// dial connects to a saved server. A host seen for the first time is shown with its fingerprint
// and trusted only when the person says so, or names that exact fingerprint with --trust.
func (a *app) dial(ctx context.Context, e server.Entry, trust string) (*remote.Client, error) {
	t := remote.Target{Host: e.Host, Port: e.Port, User: e.User, KeyPath: e.KeyPath, Passphrase: a.passphrase}
	c, err := remote.Dial(ctx, t, a.hosts)
	var unknown *remote.UnknownHostError
	if !errors.As(err, &unknown) {
		return c, err
	}
	if trust == "" {
		a.errf("This is the first connection to %s.\nIts %s key has the fingerprint\n\n  %s\n\nCompare it with what your server provider shows for this server.\n",
			e.Host, unknown.KeyType, unknown.Fingerprint)
		if !a.confirm("Trust this server?") {
			return nil, errors.New("not trusted; nothing was connected")
		}
		trust = unknown.Fingerprint
	}
	if err := remote.TrustHost(ctx, t, a.hosts, trust); err != nil {
		return nil, err
	}
	return remote.Dial(ctx, t, a.hosts)
}

// open connects to a saved server by name.
func (a *app) open(ctx context.Context, name, trust string) (server.Entry, *remote.Client, error) {
	e, err := a.registry.Get(name)
	if err != nil {
		return e, nil, err
	}
	c, err := a.dial(ctx, e, trust)
	return e, c, err
}

func (a *app) add(ctx context.Context, name string, args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", "", "server address")
	port := fs.Int("port", 22, "SSH port")
	login := fs.String("user", "", "user to log in as")
	key := fs.String("key", a.keyPath("astra"), "private key")
	trust := fs.String("trust", "", "fingerprint to trust without asking")
	if err := fs.Parse(args); err != nil || fs.NArg() > 0 || *host == "" {
		return fmt.Errorf("%w: add needs --host", errUsage)
	}
	if *login == "" {
		if me, err := user.Current(); err == nil {
			*login = me.Username
		}
	}
	e := server.Entry{Name: name, Host: *host, Port: *port, User: *login, KeyPath: *key, Role: "all-in-one"}
	if err := e.Validate(); err != nil {
		return err
	}
	c, err := a.dial(ctx, e, *trust)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	if res, err := c.Run(ctx, "true", nil); err != nil || res.Exit != 0 {
		return fmt.Errorf("the server accepted the key but could not run a command: %w", err)
	}
	if err := a.registry.Save(e); err != nil {
		return err
	}
	a.errf("Server %s added. Next: astra server check %s\n", name, name)
	return nil
}

func (a *app) check(ctx context.Context, name string, args []string) error {
	trust, err := trustFlag(args)
	if err != nil {
		return err
	}
	_, c, err := a.open(ctx, name, trust)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	report, err := server.Probe(ctx, sshRunner{c})
	if err != nil {
		return err
	}
	a.printReport(report)
	if !report.OK() {
		return errors.New("this server cannot run Astra yet; fix the failed items above")
	}
	return nil
}

func trustFlag(args []string) (string, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	trust := fs.String("trust", "", "fingerprint to trust without asking")
	if err := fs.Parse(args); err != nil || fs.NArg() > 0 {
		return "", errUsage
	}
	return *trust, nil
}

func (a *app) printReport(r server.Report) {
	for _, c := range r.Checks {
		a.outf("%-5s %-30s %s\n", map[server.Status]string{server.Pass: "ok", server.Warn: "warn", server.Fail: "FAIL"}[c.Status], c.Name, c.Detail)
		if c.Fix != "" {
			a.outf("      %s\n", c.Fix)
		}
	}
	switch {
	case r.Installed != "" && r.Running:
		a.outf("\nastrad %s is installed and running.\n", r.Installed)
	case r.Installed != "":
		a.outf("\nastrad %s is installed but not running.\n", r.Installed)
	}
}

// pair makes a pairing token on the server and trades it through an SSH tunnel; the token is
// never printed. The device token is kept in this machine's config.
func (a *app) pair(ctx context.Context, name string, args []string) error {
	trust, err := trustFlag(args)
	if err != nil {
		return err
	}
	e, c, err := a.open(ctx, name, trust)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	report, err := server.Probe(ctx, sshRunner{c})
	if err != nil {
		return err
	}
	host, _ := os.Hostname()
	device, err := server.Pair(ctx, sshRunner{c}, report.Root, "astra on "+host, func(ctx context.Context, pairing, deviceName string) (string, error) {
		addr, stop, err := c.Forward(ctx, "127.0.0.1:0", daemonAddr)
		if err != nil {
			return "", err
		}
		defer stop()
		return exchange(ctx, "http://"+addr, pairing, deviceName)
	})
	if err != nil {
		return err
	}
	e.DeviceToken = device.Reveal()
	if err := a.registry.Save(e); err != nil {
		return err
	}
	a.errf("Paired with %s. The device token is kept in %s (readable only by you).\n", name, a.registry.Path)
	return nil
}

// exchange calls the daemon's pairing endpoint at base and returns the device token.
func exchange(ctx context.Context, base, pairing, deviceName string) (string, error) {
	body, err := json.Marshal(map[string]string{"token": pairing, "name": deviceName})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/v1/pairing", bytes.NewReader(body)) //nolint:gosec // base is the local end of our own SSH tunnel
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req) //nolint:gosec // base is the local end of our own SSH tunnel
	if err != nil {
		return "", fmt.Errorf("reach the daemon through the tunnel: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("the daemon answered %d", resp.StatusCode)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.UnmarshalRead(io.LimitReader(resp.Body, 1<<16), &out); err != nil {
		return "", fmt.Errorf("read the daemon's answer: %w", err)
	}
	return out.Token, nil
}

// tunnel keeps a local port open to the daemon until it is stopped (N-12).
func (a *app) tunnel(ctx context.Context, name string, args []string) error {
	fs := flag.NewFlagSet("tunnel", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	port := fs.Int("port", 7437, "local port")
	trust := fs.String("trust", "", "fingerprint to trust without asking")
	if err := fs.Parse(args); err != nil || fs.NArg() > 0 {
		return errUsage
	}
	_, c, err := a.open(ctx, name, *trust)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	addr, stop, err := c.Forward(ctx, fmt.Sprintf("127.0.0.1:%d", *port), daemonAddr)
	if err != nil {
		return err
	}
	defer stop()
	a.errf("The daemon of %s is at http://%s (Ctrl-C closes the tunnel).\n", name, addr)
	<-ctx.Done()
	return nil
}
