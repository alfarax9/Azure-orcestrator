package pty

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// HostCommand is the subcommand that runs a session host. The daemon binary and the
// tests both dispatch it to HostMain.
const HostCommand = "session-host"

const readyTimeout = 5 * time.Second

var (
	validID   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)
	validUser = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)
	// validFollow names an event source registered with HostMainWith.
	validFollow = regexp.MustCompile(`^[a-z]{1,16}$`)
)

// sharedMode lets the daemon's group use the socket of a session that runs as another user.
const sharedMode = 0o660

// Spec describes the agent a new session runs.
type Spec struct {
	ID         string
	User       string // run the agent as this Linux user, through sudo; empty: as the caller
	Argv       []string
	Dir        string
	Env        []string // the agent's whole environment; nothing is inherited from the daemon
	Follow     string   // name of the event source the host starts, such as "claude"; empty: none
	Rows, Cols uint16
}

// Spawner starts session hosts as processes that outlive the daemon.
type Spawner struct {
	SocketDir  string        // holds <id>.sock and <id>.log for every session
	Executable string        // binary that handles HostCommand
	Scrollback int           // bytes of output each host keeps
	Linger     time.Duration // how long an exited session waits for its exit code to be read
}

// Socket is where the host of session id listens.
func (s Spawner) Socket(id string) string {
	return filepath.Join(s.SocketDir, id+".sock")
}

// Start launches a detached host for spec and returns its socket once it accepts
// connections. The host is not a child of the caller, so it survives a daemon restart.
func (s Spawner) Start(ctx context.Context, spec Spec) (string, error) {
	if !validID.MatchString(spec.ID) {
		return "", fmt.Errorf("invalid session id %q", spec.ID)
	}
	if len(spec.Argv) == 0 {
		return "", errors.New("session needs a command")
	}
	if spec.Follow != "" && !validFollow.MatchString(spec.Follow) {
		return "", fmt.Errorf("invalid event source %q", spec.Follow)
	}
	if err := os.MkdirAll(s.SocketDir, 0o700); err != nil {
		return "", fmt.Errorf("create socket dir: %w", err)
	}
	sock := s.Socket(spec.ID)
	cmd, err := s.launchCommand(ctx, spec, sock)
	if err != nil {
		return "", err
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("launch session host: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return sock, WaitReady(ctx, sock)
}

// launchCommand builds the command that starts a detached host for spec. For another
// user it goes through sudo, which resets the environment; env -i then hands the agent
// exactly spec.Env, so nothing of the daemon's environment reaches it (F12-01, N-11).
func (s Spawner) launchCommand(ctx context.Context, spec Spec, sock string) (*exec.Cmd, error) {
	host := []string{
		s.Executable, HostCommand, "--detach", "--socket", sock, "--dir", spec.Dir,
		"--rows", strconv.Itoa(int(spec.Rows)), "--cols", strconv.Itoa(int(spec.Cols)),
		"--scrollback", strconv.Itoa(s.Scrollback), "--linger", s.Linger.String(),
	}
	if spec.Follow != "" {
		host = append(host, "--follow", spec.Follow)
	}
	if spec.User == "" {
		host = append(host, "--")
		cmd := exec.CommandContext(ctx, host[0], append(host[1:], spec.Argv...)...) //nolint:gosec // our own binary
		cmd.Env = spec.Env
		return cmd, nil
	}
	if !validUser.MatchString(spec.User) {
		return nil, fmt.Errorf("invalid user name %q", spec.User)
	}
	host = append(host, "--socket-mode", strconv.FormatUint(sharedMode, 8), "--")
	argv := []string{"-n", "-u", spec.User, "--", "/usr/bin/env", "-i"}
	argv = append(append(append(argv, spec.Env...), host...), spec.Argv...)
	cmd := exec.CommandContext(ctx, "sudo", argv...) //nolint:gosec // user name validated above
	cmd.Env = []string{"PATH=/usr/bin:/bin"}
	cmd.Dir = "/" // the caller's folder may be out of the agent user's reach
	return cmd, nil
}

// WaitReady waits until the host listening on sock answers, for a few seconds. A host started
// by another process, such as a systemd unit, takes a moment to open its socket.
func WaitReady(ctx context.Context, sock string) error {
	ctx, cancel := context.WithTimeout(ctx, readyTimeout)
	defer cancel()
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	for {
		if ping(ctx, sock) == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			log := strings.TrimSuffix(sock, ".sock") + ".log"
			return fmt.Errorf("session host did not start on %s (see %s): %w", sock, log, ctx.Err())
		case <-tick.C:
		}
	}
}

// Sessions lists the sessions whose hosts still answer, for a daemon that just restarted.
func (s Spawner) Sessions(ctx context.Context) ([]string, error) {
	socks, err := filepath.Glob(filepath.Join(s.SocketDir, "*.sock"))
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	var ids []string
	for _, sock := range socks {
		if ping(ctx, sock) != nil {
			continue // a host that crashed; its socket file is stale
		}
		ids = append(ids, strings.TrimSuffix(filepath.Base(sock), ".sock"))
	}
	return ids, nil
}

// HostMainWith runs HostCommand with its arguments and environment and returns the exit
// status. With --detach it re-executes itself in a new session and returns at once, so the
// caller never owns the host process. followers are the event sources a host may be told to
// start with --follow <name>; an unknown name is refused, so a host never runs anything not
// registered.
func HostMainWith(args, env []string, followers map[string]Follower) int {
	fs := flag.NewFlagSet(HostCommand, flag.ContinueOnError)
	follow := fs.String("follow", "", "event source to start with the session")
	detach := fs.Bool("detach", false, "start a detached host and return")
	cfg := HostConfig{Env: env}
	fs.StringVar(&cfg.Socket, "socket", "", "unix socket to serve")
	fs.StringVar(&cfg.Dir, "dir", "", "working directory of the agent")
	mode := fs.String("socket-mode", "600", "octal permissions of the socket and log")
	rows := fs.Uint("rows", 24, "terminal rows")
	cols := fs.Uint("cols", 80, "terminal columns")
	fs.IntVar(&cfg.Scrollback, "scrollback", 1<<20, "bytes of output to keep")
	fs.DurationVar(&cfg.Linger, "linger", 10*time.Minute, "wait for the exit code to be read")
	if err := fs.Parse(args); err != nil || cfg.Socket == "" || fs.NArg() == 0 {
		return 2
	}
	perm, err := strconv.ParseUint(*mode, 8, 32)
	if err != nil || perm&^0o770 != 0 {
		return 2 // owner and group bits only: a socket is never opened to everyone
	}
	if *follow != "" {
		if cfg.Follow = followers[*follow]; cfg.Follow == nil {
			return 2
		}
	}
	cfg.SocketMode = os.FileMode(perm)
	cfg.Argv, cfg.Rows, cfg.Cols = fs.Args(), uint16(*rows), uint16(*cols) //nolint:gosec // terminal sizes are small
	cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	if *detach {
		return detachHost(args, env, cfg)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	if err := Serve(ctx, cfg); err != nil {
		cfg.Logger.Error("session host failed", "err", err)
		return 1
	}
	return 0
}

// detachHost starts the real host in its own session with its log next to the socket,
// then returns without waiting. The host is reparented to init/launchd.
func detachHost(args, env []string, cfg HostConfig) int {
	self, err := os.Executable()
	if err != nil {
		cfg.Logger.Error("find own binary", "err", err)
		return 1
	}
	logFile, err := os.OpenFile(strings.TrimSuffix(cfg.Socket, ".sock")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, cfg.SocketMode)
	if err != nil {
		cfg.Logger.Error("open host log", "err", err)
		return 1
	}
	defer func() { _ = logFile.Close() }()
	// Drop --detach from the host flags only; the agent's own arguments follow "--".
	rest := make([]string, 0, len(args))
	for i, a := range args {
		if a == "--" {
			rest = append(rest, args[i:]...)
			break
		}
		if a != "--detach" {
			rest = append(rest, a)
		}
	}
	// Background: the host must outlive this short-lived process.
	cmd := exec.CommandContext(context.Background(), self, append([]string{HostCommand}, rest...)...) //nolint:gosec // our own binary
	cmd.Env, cmd.Stdout, cmd.Stderr = env, logFile, logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		cfg.Logger.Error("start session host", "err", err)
		return 1
	}
	_ = cmd.Process.Release()
	return 0
}
