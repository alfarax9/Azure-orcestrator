package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alfarax9/astra/internal/agent"
	"github.com/alfarax9/astra/internal/api"
	"github.com/alfarax9/astra/internal/domain"
)

const (
	defaultListen   = "127.0.0.1:7437"
	shutdownTimeout = 10 * time.Second
)

// trackerFunc lets the API hand a started worker to the supervisor, which runs for the whole
// life of the daemon rather than for one request.
type trackerFunc func(id domain.WorkerID)

func (f trackerFunc) Track(id domain.WorkerID) { f(id) }

// runServe is the daemon: the API on a loopback address, and a watch on every worker whose
// agent is running. SIGTERM or Ctrl-C stops it; agent sessions are separate units and carry on.
func runServe(_ []string) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := serve(ctx, log); err != nil {
		log.Error("astrad stopped", "err", err)
		return 1
	}
	return 0
}

func serve(ctx context.Context, log *slog.Logger) error {
	addr := envOr("ASTRA_LISTEN", defaultListen)
	if err := checkListen(addr, os.Getenv("ASTRA_ALLOW_REMOTE") == "1"); err != nil {
		return err
	}
	a, err := open(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = a.store.Close() }()
	if err := applyLimit(a.ag); err != nil {
		return err
	}

	sup := agent.NewSupervisor(a.ag, a.store, log)
	tracker := trackerFunc(func(id domain.WorkerID) { sup.Track(ctx, id) })
	cfg := api.Config{Version: version, Origins: splitList(os.Getenv("ASTRA_ORIGINS")), Log: log}
	if dir := os.Getenv("ASTRA_WEB_DIR"); dir != "" {
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil { //nolint:gosec // the operator named this folder
			return fmt.Errorf("ASTRA_WEB_DIR %s has no index.html: %w", dir, err)
		}
		cfg.Static = os.DirFS(dir)
	}
	handler := api.New(cfg, a.auth, a.ws, a.ag, a.store, tracker, a.list).Handler()

	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	srv := &http.Server{
		Handler:           handler,
		BaseContext:       func(net.Listener) context.Context { return ctx }, // sockets end when the daemon does
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       2 * time.Minute,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
	}
	if err := sup.Resume(ctx); err != nil {
		return fmt.Errorf("resume workers: %w", err)
	}
	log.Info("astrad listening", "addr", ln.Addr().String(), "version", version)

	served := make(chan error, 1)
	go func() { served <- srv.Serve(ln) }()
	select {
	case err := <-served:
		return err
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	err = srv.Shutdown(shutdown)
	sup.Wait()
	if err != nil {
		return fmt.Errorf("shut down: %w", err)
	}
	return nil
}

// checkListen refuses an address other than loopback unless the operator said remote access is
// meant: the API has no TLS of its own, so a tunnel or proxy must be in front of it (ADR 0006).
func checkListen(addr string, allowRemote bool) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("ASTRA_LISTEN %q: %w", addr, err)
	}
	if ip := net.ParseIP(host); host == "localhost" || (ip != nil && ip.IsLoopback()) || allowRemote {
		return nil
	}
	return errors.New("ASTRA_LISTEN is not a loopback address: the API has no TLS, so put an SSH tunnel or a proxy in front of it, or set ASTRA_ALLOW_REMOTE=1 if one already is")
}

// defaultMaxWorkers is how many workers run at once when the operator says nothing: enough for
// parallel work on the smallest server the PRD names, and few enough to leave room for builds.
const defaultMaxWorkers = 4

// applyLimit sets how many workers may run at once from ASTRA_MAX_WORKERS.
func applyLimit(svc *agent.Service) error {
	n, err := maxWorkers(os.Getenv("ASTRA_MAX_WORKERS"))
	if err == nil {
		svc.SetLimit(n)
	}
	return err
}

// maxWorkers reads ASTRA_MAX_WORKERS: a whole number, where 0 means no limit (F10-05).
func maxWorkers(value string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return defaultMaxWorkers, nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n < 0 {
		return 0, fmt.Errorf("ASTRA_MAX_WORKERS %q is not a whole number of 0 or more (0 means no limit)", value)
	}
	return n, nil
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

// runPair makes a pairing token and prints it, and nothing else, to stdout: the desktop app
// runs this over SSH and reads the token from there, so it is never shown (F1-05).
func runPair() int {
	ctx := context.Background()
	a, err := open(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "astrad:", err)
		return 1
	}
	defer func() { _ = a.store.Close() }()
	token, err := a.auth.NewPairing(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "astrad:", err)
		return 1
	}
	fmt.Println(token.Reveal())
	return 0
}

// runDevices lists paired devices, or revokes one so its token stops working at once.
func runDevices(args []string) int {
	ctx := context.Background()
	a, err := open(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "astrad:", err)
		return 1
	}
	defer func() { _ = a.store.Close() }()
	switch {
	case len(args) == 0:
		list, err := a.store.Devices(ctx)
		if list == nil {
			list = []domain.Device{}
		}
		return report(list, err)
	case len(args) == 2 && args[0] == "revoke":
		return report(map[string]string{"revoked": args[1]}, a.store.RevokeDevice(ctx, args[1]))
	}
	fmt.Fprintln(os.Stderr, "usage: astrad devices [revoke <device-id>]")
	return 2
}
