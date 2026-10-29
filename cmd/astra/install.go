package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/alfarax9/astra/internal/server"
)

// installFlags are the choices every installation takes.
type installFlags struct {
	opts  server.Options
	yes   bool
	trust string
}

func parseInstall(args []string) (installFlags, error) {
	var f installFlags
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&f.opts.Source.BinaryDir, "binary-dir", "", "folder with astrad-linux-<arch>, astra-priv-linux-<arch> and optionally web/")
	fs.StringVar(&f.opts.Source.ReleaseURL, "release-url", "", "folder of a release: SHA256SUMS and astra-linux-<arch>.tar.gz")
	fs.BoolVar(&f.opts.Source.AllowHTTP, "insecure-http", false, "allow a release URL that is not https")
	fs.BoolVar(&f.opts.Docker, "docker", false, "install Docker Engine if it is missing")
	fs.BoolVar(&f.yes, "yes", false, "do not ask for approval")
	fs.StringVar(&f.trust, "trust", "", "fingerprint to trust without asking")
	if err := fs.Parse(args); err != nil || fs.NArg() > 0 {
		return f, errUsage
	}
	return f, nil
}

// installOn installs the daemon on a saved server over SSH (F1-04).
func (a *app) installOn(ctx context.Context, name string, args []string) error {
	f, err := parseInstall(args)
	if err != nil {
		return err
	}
	_, c, err := a.open(ctx, name, f.trust)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	return a.install(ctx, sshRunner{c}, f)
}

// installHere installs on this machine, for the person who is logged in to the server itself
// (F1-06), and then prints the one pairing token for the web UI.
func (a *app) installHere(args []string) error {
	f, err := parseInstall(args)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := a.install(ctx, localRunner{}, f); err != nil {
		return err
	}
	report, err := server.Probe(ctx, localRunner{})
	if err != nil {
		return err
	}
	token, err := server.PairingToken(ctx, localRunner{}, report.Root)
	if err != nil {
		return err
	}
	a.errln("\nOpen the web UI (through an SSH tunnel to port 7437, or astra server tunnel from your laptop) and paste this pairing token. It works once and expires in 10 minutes:")
	a.outln(token.Reveal())
	return nil
}

// install checks the machine, shows what will be done, and does it once the person agrees.
func (a *app) install(ctx context.Context, r server.Runner, f installFlags) error {
	report, err := server.Probe(ctx, r)
	if err != nil {
		return err
	}
	a.printReport(report)
	plan, err := server.NewPlan(r, report, f.opts)
	if err != nil {
		return err
	}
	a.errln("\nThis will:")
	for i, title := range plan.Titles() {
		a.errf("  %d. %s\n", i+1, title)
	}
	for _, n := range plan.Notes {
		a.errf("\n%s\n", n)
	}
	if !f.yes && !a.confirm("\nGo ahead?") {
		return errors.New("cancelled; nothing was changed")
	}
	total := len(plan.Steps)
	out, err := plan.Run(ctx, func(i int, title string) { a.errf("[%d/%d] %s\n", i+1, total, title) })
	if err != nil {
		return err
	}
	a.errf("\nThe daemon is running: %s\n", out.Version)
	return nil
}
