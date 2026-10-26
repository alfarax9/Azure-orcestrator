package main

import (
	"context"
	"encoding/json/v2"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/alfarax9/astra/internal/asuser"
	"github.com/alfarax9/astra/internal/domain"
	"github.com/alfarax9/astra/internal/store"
	"github.com/alfarax9/astra/internal/workspace"
)

const workspaceUsage = `usage:
  astrad workspace add <slug> <name>
  astrad project add <slug> <name> [repo-url]
  astrad project commands <project-id> [-dev CMD] [-build CMD] [-test CMD]
  astrad worker new <project-id> <title> [-agent claude|codex] [-internal]
the database is $ASTRA_DB (default ~/.astrad/astra.db). ASTRA_LOCAL=1 runs without agent
users, under $ASTRA_AGENTS_DIR (default ~/.astrad/agents; on a server /var/lib/astra/agents)`

// runObjects is the command line for workspaces, projects and workers until the API exists
// (M1 slice 4). It prints what it created as JSON.
func runObjects(kind string, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, workspaceUsage)
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	st, svc, err := openWorkspaces(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "astrad:", err)
		return 1
	}
	defer func() { _ = st.Close() }()
	var out any
	switch kind + " " + args[0] {
	case "workspace add":
		out, err = need(args, 3, func() (any, error) { return svc.AddWorkspace(ctx, args[1], args[2]) })
	case "project add":
		out, err = need(args, 3, func() (any, error) { return svc.AddProject(ctx, args[1], args[2], optional(args, 3)) })
	case "project commands":
		out, err = setCommands(ctx, st, svc, args[1:])
	case "worker new":
		out, err = newWorker(ctx, svc, args[1:])
	default:
		err = errUsage
	}
	return report(out, err)
}

var errUsage = errors.New("wrong arguments")

func report(out any, err error) int {
	switch {
	case errors.Is(err, errUsage):
		fmt.Fprintln(os.Stderr, workspaceUsage)
		return 2
	case err != nil:
		fmt.Fprintln(os.Stderr, "astrad:", err)
		return 1
	}
	b, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "astrad:", err)
		return 1
	}
	fmt.Println(string(b))
	return 0
}

func need(args []string, n int, f func() (any, error)) (any, error) {
	if len(args) < n {
		return nil, errUsage
	}
	return f()
}

func optional(args []string, i int) string {
	if len(args) > i {
		return args[i]
	}
	return ""
}

func openWorkspaces(ctx context.Context) (*store.Store, *workspace.Service, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("find home directory: %w", err)
	}
	dbPath := envOr("ASTRA_DB", filepath.Join(home, ".astrad", "astra.db"))
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, nil, fmt.Errorf("create database folder: %w", err)
	}
	st, err := store.Open(ctx, dbPath)
	if err != nil {
		return nil, nil, err
	}
	local := os.Getenv("ASTRA_LOCAL") == "1"
	agents := "/var/lib/astra/agents"
	if local {
		agents = filepath.Join(home, ".astrad", "agents")
	}
	run := asuser.Runner{
		Root:   envOr("ASTRA_AGENTS_DIR", agents),
		Helper: envOr("ASTRA_PRIV", "/usr/libexec/astra/astra-priv"),
		Local:  local,
	}
	return st, workspace.New(run, run, st, st, st, time.Now), nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// setCommands changes only the commands given as flags and keeps the others.
func setCommands(ctx context.Context, st *store.Store, svc *workspace.Service, args []string) (any, error) {
	if len(args) == 0 {
		return nil, errUsage
	}
	id := domain.ProjectID(args[0])
	p, err := st.Project(ctx, id)
	if err != nil {
		return nil, err
	}
	fs := flag.NewFlagSet("commands", flag.ContinueOnError)
	c := p.Commands
	fs.StringVar(&c.Dev, "dev", c.Dev, "command that runs the app for development")
	fs.StringVar(&c.Build, "build", c.Build, "command that builds the app")
	fs.StringVar(&c.Test, "test", c.Test, "command that runs the tests")
	if err := fs.Parse(args[1:]); err != nil || fs.NArg() > 0 {
		return nil, errUsage
	}
	if err := svc.SetCommands(ctx, id, c); err != nil {
		return nil, err
	}
	p.Commands = c
	return p, nil
}

func newWorker(ctx context.Context, svc *workspace.Service, args []string) (any, error) {
	if len(args) < 2 {
		return nil, errUsage
	}
	fs := flag.NewFlagSet("worker", flag.ContinueOnError)
	agentName := fs.String("agent", "claude", "claude or codex")
	internal := fs.Bool("internal", false, "an internal project: no client review")
	if err := fs.Parse(args[2:]); err != nil || fs.NArg() > 0 {
		return nil, errUsage
	}
	w, err := svc.CreateWorker(ctx, domain.ProjectID(args[0]), args[1], domain.AgentKind(*agentName), !*internal)
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(os.Stderr, "worktree:", w.Worktree)
	return w, nil
}
