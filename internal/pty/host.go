package pty

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
)

const (
	writeTimeout = 5 * time.Second // a stuck daemon is dropped instead of stalling the agent
	drainTimeout = 2 * time.Second // output still buffered after the agent exits
	killDelay    = 3 * time.Second // SIGHUP first, SIGKILL if the agent ignores it
)

// socketCheck is how often a host checks its socket file still exists. A removed socket
// means nobody can reach the session again, so the host stops instead of running forever.
var socketCheck = 10 * time.Second

// HostConfig describes one session host.
type HostConfig struct {
	Socket     string
	SocketMode os.FileMode // permissions of the socket and log; zero means 0600
	Argv       []string
	Dir        string
	Env        []string // the agent's whole environment
	Rows, Cols uint16
	Scrollback int           // bytes of output kept for a reconnecting daemon
	Linger     time.Duration // how long an exited session waits for a daemon to read its exit code
	Logger     *slog.Logger
}

type host struct {
	cfg  HostConfig
	cmd  *exec.Cmd
	ptmx *os.File

	collectOnce sync.Once
	collected   chan struct{} // closed once a daemon has received the exit code

	mu        sync.Mutex // guards the fields below
	history   *scrollback
	attached  map[net.Conn]struct{} // connections that receive output; a new one never evicts another
	conns     map[net.Conn]struct{} // every open connection, closed when the host stops
	rows      uint16
	cols      uint16
	exited    bool
	exitCode  int
	killTimer *time.Timer // set once the agent was asked to stop
	closing   bool        // no new connections once the host is stopping
}

// Serve runs cfg.Argv in a PTY and serves it on cfg.Socket. It returns once the agent has
// exited and a daemon has read the exit code (or cfg.Linger has passed), or when ctx is
// cancelled, in which case the agent is stopped first.
func Serve(ctx context.Context, cfg HostConfig) error {
	if cfg.SocketMode == 0 {
		cfg.SocketMode = 0o600
	}
	ln, err := listen(ctx, cfg.Socket, cfg.SocketMode)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(cfg.Socket) }()

	// WithoutCancel: a cancelled ctx must not SIGKILL the agent; stop hangs it up first.
	cmd := exec.CommandContext(context.WithoutCancel(ctx), cfg.Argv[0], cfg.Argv[1:]...) //nolint:gosec // running the agent CLI is this package's job
	cmd.Dir, cmd.Env = cfg.Dir, cfg.Env
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: cfg.Rows, Cols: cfg.Cols})
	if err != nil {
		_ = ln.Close()
		return fmt.Errorf("start %s: %w", cfg.Argv[0], err)
	}
	h := &host{
		cfg: cfg, cmd: cmd, ptmx: ptmx, collected: make(chan struct{}),
		attached: map[net.Conn]struct{}{}, conns: map[net.Conn]struct{}{},
		history: newScrollback(cfg.Scrollback), rows: cfg.Rows, cols: cfg.Cols,
	}
	cfg.Logger.Info("session started", "socket", cfg.Socket, "agent", cfg.Argv[0], "pid", cmd.Process.Pid)

	var wg sync.WaitGroup
	pumped := make(chan struct{})
	exited := make(chan error, 1)
	done := make(chan struct{})
	wg.Go(func() { defer close(pumped); h.pump() })
	wg.Go(func() { h.accept(ln) })
	wg.Go(func() { exited <- cmd.Wait() })
	wg.Go(func() { h.watchSocket(done) })

	code := h.waitAgent(ctx, exited, pumped)
	close(done)
	h.finish(code)
	h.linger(ctx)

	_ = ln.Close()
	h.closeConns()
	_ = ptmx.Close()
	wg.Wait()
	cfg.Logger.Info("session closed", "socket", cfg.Socket, "exit_code", code)
	return nil
}

func listen(ctx context.Context, socket string, mode os.FileMode) (net.Listener, error) {
	if ping(ctx, socket) == nil {
		return nil, fmt.Errorf("session %s is already running", socket)
	}
	if err := os.Remove(socket); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("remove stale socket: %w", err)
	}
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "unix", socket)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", socket, err)
	}
	if err := os.Chmod(socket, mode); err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("restrict socket: %w", err)
	}
	return ln, nil
}

// waitAgent waits for the agent to exit, stopping it if ctx ends first, then lets the
// pump drain the last output so it reaches the scrollback before the exit code.
func (h *host) waitAgent(ctx context.Context, exited <-chan error, pumped <-chan struct{}) int {
	var err error
	select {
	case err = <-exited:
	case <-ctx.Done():
		h.stop()
		err = <-exited
	}
	select {
	case <-pumped:
	case <-time.After(drainTimeout):
	}
	return exitCode(err)
}

// watchSocket stops the agent if the socket file disappears, for example when the
// daemon's data directory is purged, so no unreachable session is left running.
func (h *host) watchSocket(done <-chan struct{}) {
	tick := time.NewTicker(socketCheck)
	defer tick.Stop()
	for {
		select {
		case <-done:
			return
		case <-tick.C:
			if _, err := os.Stat(h.cfg.Socket); errors.Is(err, os.ErrNotExist) {
				h.cfg.Logger.Warn("socket removed, stopping session", "socket", h.cfg.Socket)
				h.stop()
				return
			}
		}
	}
}

func exitCode(err error) int {
	var ee *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &ee):
		return ee.ExitCode()
	default:
		return -1
	}
}

// stop hangs up the agent's process group, then kills it if it is still there after
// killDelay. finish cancels the kill, so a reused process ID is never signalled.
func (h *host) stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.exited || h.killTimer != nil {
		return
	}
	pgid := -h.cmd.Process.Pid
	_ = syscall.Kill(pgid, syscall.SIGHUP)
	h.killTimer = time.AfterFunc(killDelay, func() { _ = syscall.Kill(pgid, syscall.SIGKILL) })
}

// pump copies agent output into the scrollback and to the attached daemon.
func (h *host) pump() {
	buf := make([]byte, 32<<10)
	for {
		n, err := h.ptmx.Read(buf)
		if n > 0 {
			h.output(buf[:n])
		}
		if err != nil {
			return
		}
	}
}

func (h *host) output(p []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.history.write(p)
	for c := range h.attached {
		h.send(c, frameOutput, p)
	}
}

// send writes to one attached daemon and reports whether it is still attached; the
// caller holds h.mu. A daemon that is gone or cannot keep up is detached and repaints
// from the scrollback when it reconnects. Only the write side is shut: input it sent
// before leaving is still read and typed.
func (h *host) send(c net.Conn, kind byte, p []byte) bool {
	_ = c.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err := writeChunked(c, kind, p); err != nil {
		h.cfg.Logger.Warn("daemon detached", "socket", h.cfg.Socket, "err", err)
		if uc, ok := c.(*net.UnixConn); ok {
			_ = uc.CloseWrite()
		}
		delete(h.attached, c)
		return false
	}
	return true
}

func (h *host) finish(code int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.exited, h.exitCode = true, code
	if h.killTimer != nil {
		h.killTimer.Stop()
	}
	for c := range h.attached {
		h.sendExit(c)
	}
}

// sendExit tells one attached daemon the exit code; the caller holds h.mu.
func (h *host) sendExit(c net.Conn) {
	var p [4]byte
	binary.BigEndian.PutUint32(p[:], uint32(int32(h.exitCode))) //nolint:gosec // exit codes fit in int32
	if h.send(c, frameExit, p[:]) {
		h.collectOnce.Do(func() { close(h.collected) })
	}
}

func (h *host) linger(ctx context.Context) {
	timer := time.NewTimer(h.cfg.Linger)
	defer timer.Stop()
	select {
	case <-h.collected:
	case <-timer.C:
	case <-ctx.Done():
	}
}

// closeConns ends every connection so their readers return and the host can stop.
func (h *host) closeConns() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closing = true
	for c := range h.conns {
		_ = c.Close()
	}
	clear(h.attached)
}
