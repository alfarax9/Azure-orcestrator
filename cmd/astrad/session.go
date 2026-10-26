package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alfarax9/astra/internal/pty"
)

// agentEnv is everything an agent inherits from the daemon's environment. Anything else,
// tokens included, stays in the daemon (docs/go-rules.md, secrets).
var agentEnv = []string{"HOME", "USER", "LOGNAME", "PATH", "SHELL", "LANG", "LC_ALL", "TMPDIR"}

const sessionUsage = `usage: astrad session <command>
  start <id> [-dir DIR] [-user NAME] [-rows N] [-cols N] -- <command> [args]
  ls
  tail  <id> [-for 2s] [-strip]
  send  <id> <text>    (Go escapes allowed: \r \x03)
  stop  <id>
sessions live in $ASTRA_SESSIONS_DIR, default ~/.astrad/sessions`

// runSession is the spike S1 command line for session hosts, until the daemon's API exists.
func runSession(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, sessionUsage)
		return 2
	}
	sp, err := spawner()
	if err != nil {
		fmt.Fprintln(os.Stderr, "astrad:", err)
		return 1
	}
	cmds := map[string]func(context.Context, pty.Spawner, []string) error{
		"start": sessionStart, "ls": sessionList, "tail": sessionTail, "send": sessionSend, "stop": sessionStop,
	}
	run, ok := cmds[args[0]]
	if !ok {
		fmt.Fprintln(os.Stderr, sessionUsage)
		return 2
	}
	if err := run(context.Background(), sp, args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "astrad:", err)
		return 1
	}
	return 0
}

func spawner() (pty.Spawner, error) {
	dir := os.Getenv("ASTRA_SESSIONS_DIR")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return pty.Spawner{}, fmt.Errorf("find home directory: %w", err)
		}
		dir = filepath.Join(home, ".astrad", "sessions")
	}
	exe, err := os.Executable()
	if err != nil {
		return pty.Spawner{}, fmt.Errorf("find own binary: %w", err)
	}
	return pty.Spawner{SocketDir: dir, Executable: exe, Scrollback: 1 << 20, Linger: 10 * time.Minute}, nil
}

// environment is what an agent receives: the allowlisted variables of the daemon, except
// that an agent running as another user gets that user's identity instead of the daemon's.
func environment(asUser string) ([]string, error) {
	env := []string{"TERM=xterm-256color", "COLORTERM=truecolor"}
	for _, k := range agentEnv {
		if v, ok := os.LookupEnv(k); ok {
			env = append(env, k+"="+v)
		}
	}
	if asUser == "" {
		return env, nil
	}
	u, err := user.Lookup(asUser)
	if err != nil {
		return nil, fmt.Errorf("look up user %s: %w", asUser, err)
	}
	identity := map[string]string{"HOME": u.HomeDir, "USER": u.Username, "LOGNAME": u.Username, "SHELL": "/bin/bash"}
	kept := env[:0]
	for _, kv := range env {
		if k, _, _ := strings.Cut(kv, "="); identity[k] == "" {
			kept = append(kept, kv)
		}
	}
	for _, k := range []string{"HOME", "USER", "LOGNAME", "SHELL"} {
		kept = append(kept, k+"="+identity[k])
	}
	return kept, nil
}

func sessionStart(ctx context.Context, sp pty.Spawner, args []string) error {
	if len(args) == 0 {
		return errors.New(sessionUsage)
	}
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	wd, _ := os.Getwd()
	dir := fs.String("dir", wd, "working directory of the agent")
	asUser := fs.String("user", "", "run the agent as this Linux user (needs a sudo rule)")
	rows := fs.Uint("rows", 40, "terminal rows")
	cols := fs.Uint("cols", 120, "terminal columns")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	env, err := environment(*asUser)
	if err != nil {
		return err
	}
	spec := pty.Spec{
		ID: args[0], User: *asUser, Argv: fs.Args(), Dir: *dir, Env: env,
		Rows: uint16(*rows), Cols: uint16(*cols), //nolint:gosec // terminal sizes are small
	}
	sock, err := sp.Start(ctx, spec)
	if err != nil {
		return err
	}
	fmt.Println(sock)
	return nil
}

func sessionList(ctx context.Context, sp pty.Spawner, _ []string) error {
	ids, err := sp.Sessions(ctx)
	for _, id := range ids {
		fmt.Println(id)
	}
	return err
}

// ansi matches terminal escape sequences, so -strip prints a TUI's text only.
var ansi = regexp.MustCompile(`\x1b\[[0-9;?<>=]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(\x07|\x1b\\)|\x1b[()][0-9A-Za-z]|\x1b[=>78]`)

func sessionTail(ctx context.Context, sp pty.Spawner, args []string) error {
	if len(args) == 0 {
		return errors.New(sessionUsage)
	}
	fs := flag.NewFlagSet("tail", flag.ContinueOnError)
	wait := fs.Duration("for", 2*time.Second, "how long to read before detaching")
	strip := fs.Bool("strip", false, "remove terminal escape sequences")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	c, err := pty.Dial(ctx, sp.Socket(args[0]))
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	_ = c.SetReadDeadline(time.Now().Add(*wait))
	for {
		ev, err := c.Next()
		if err != nil {
			return nil // the deadline ends tail; the session keeps running
		}
		out := ev.Output
		if *strip {
			out = ansi.ReplaceAll(out, nil)
		}
		_, _ = os.Stdout.Write(out)
		if ev.Exited {
			fmt.Printf("\n[exited %d]\n", ev.ExitCode)
			return nil
		}
	}
}

func sessionSend(ctx context.Context, sp pty.Spawner, args []string) error {
	if len(args) != 2 {
		return errors.New(sessionUsage)
	}
	text, err := strconv.Unquote(`"` + args[1] + `"`)
	if err != nil {
		return fmt.Errorf("text: %w", err)
	}
	c, err := pty.Dial(ctx, sp.Socket(args[0]))
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	_, err = c.Write([]byte(text))
	return err
}

func sessionStop(ctx context.Context, sp pty.Spawner, args []string) error {
	if len(args) != 1 {
		return errors.New(sessionUsage)
	}
	c, err := pty.Dial(ctx, sp.Socket(args[0]))
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	if err := c.Stop(); err != nil {
		return err
	}
	_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		ev, err := c.Next()
		if err != nil {
			return fmt.Errorf("waiting for exit: %w", err)
		}
		if ev.Exited {
			fmt.Println("exited", ev.ExitCode)
			return nil
		}
	}
}
