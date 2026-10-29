package pty

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"testing"
	"time"
)

// The test binary doubles as the session host, as astrad does in production.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == HostCommand {
		os.Exit(HostMainWith(os.Args[2:], os.Environ(), testFollowers))
	}
	os.Exit(m.Run())
}

// socketDir is short on purpose: macOS limits Unix socket paths to 104 bytes, and
// t.TempDir() under /var/folders is longer than that.
func socketDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "astra-pty-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func spawner(t *testing.T) Spawner {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return Spawner{SocketDir: socketDir(t), Executable: exe, Scrollback: 64 << 10, Linger: 5 * time.Second}
}

func shell(id, script string, env ...string) Spec {
	return Spec{
		ID: id, Argv: []string{"/bin/sh", "-c", script}, Dir: "/",
		Env: append([]string{"PATH=/usr/bin:/bin", "TERM=dumb"}, env...), Rows: 24, Cols: 80,
	}
}

// readUntil collects output until it contains every want, in any order, or the session exits.
func readUntil(t *testing.T, c *Conn, want ...string) (string, Event) {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))
	var out strings.Builder
	containsAll := func() bool {
		for _, w := range want {
			if !strings.Contains(out.String(), w) {
				return false
			}
		}
		return true
	}
	for !containsAll() {
		ev, err := c.Next()
		if err != nil {
			t.Fatalf("waiting for %q, got %q: %v", want, out.String(), err)
		}
		out.Write(ev.Output)
		if ev.Exited {
			return out.String(), ev
		}
	}
	return out.String(), Event{}
}

func TestSessionSurvivesDaemonRestart(t *testing.T) {
	ctx := t.Context()
	sp := spawner(t)
	sock, err := sp.Start(ctx, shell("s1", "echo ready; exec cat"))
	if err != nil {
		t.Fatal(err)
	}

	first, err := Dial(ctx, sock)
	if err != nil {
		t.Fatal(err)
	}
	readUntil(t, first, "ready")
	if _, err := first.Write([]byte("before-restart\r")); err != nil {
		t.Fatal(err)
	}
	readUntil(t, first, "before-restart")
	_ = first.Close() // the daemon goes away; the agent keeps running

	ids, err := sp.Sessions(ctx)
	if err != nil || len(ids) != 1 || ids[0] != "s1" {
		t.Fatalf("a restarted daemon should find session s1, got %v (%v)", ids, err)
	}
	second, err := Dial(ctx, sock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = second.Close() }()
	repaint, _ := readUntil(t, second, "before-restart")
	if !strings.Contains(repaint, "ready") {
		t.Fatalf("scrollback lost the start of the session: %q", repaint)
	}
	if _, err := second.Write([]byte("after-restart\r")); err != nil {
		t.Fatal(err)
	}
	readUntil(t, second, "after-restart")

	if err := second.Stop(); err != nil {
		t.Fatal(err)
	}
	if _, ev := readUntil(t, second, "\x00never"); !ev.Exited {
		t.Fatal("expected the session to report its exit")
	}
}

// A daemon may send input and disconnect at once; the input must still be typed.
func TestInputFromAQuicklyClosedConnectionIsTyped(t *testing.T) {
	ctx := t.Context()
	sp := spawner(t)
	sock, err := sp.Start(ctx, shell("quick", "echo ready; exec cat"))
	if err != nil {
		t.Fatal(err)
	}
	for _, word := range []string{"one", "two", "three"} {
		c, err := Dial(ctx, sock)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = c.Write([]byte(word + "\r"))
		_ = c.Close()
	}
	c, err := Dial(ctx, sock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Stop(); _ = c.Close() }()
	// Connections are served concurrently, so the words may arrive in any order.
	readUntil(t, c, "one", "two", "three")
}

func TestResizeReachesTheAgent(t *testing.T) {
	ctx := t.Context()
	sp := spawner(t)
	sock, err := sp.Start(ctx, shell("size", "echo ready; read line; stty size"))
	if err != nil {
		t.Fatal(err)
	}
	c, err := Dial(ctx, sock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()
	readUntil(t, c, "ready")
	if err := c.Resize(30, 100); err != nil {
		t.Fatal(err)
	}
	_, _ = c.Write([]byte("\r"))
	readUntil(t, c, "30 100")
}

func TestAgentGetsOnlyItsSpecEnvironment(t *testing.T) {
	t.Setenv("ASTRA_TEST_DAEMON_SECRET", "must-not-leak")
	ctx := t.Context()
	sp := spawner(t)
	sock, err := sp.Start(ctx, shell("env", "env; echo done", "ASTRA_GIVEN=yes"))
	if err != nil {
		t.Fatal(err)
	}
	c, err := Dial(ctx, sock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()
	out, _ := readUntil(t, c, "done")
	if !strings.Contains(out, "ASTRA_GIVEN=yes") {
		t.Fatalf("agent is missing its spec environment: %q", out)
	}
	if strings.Contains(out, "must-not-leak") {
		t.Fatalf("daemon environment leaked into the agent: %q", out)
	}
}

func TestExitCodeReachesALateDaemon(t *testing.T) {
	ctx := t.Context()
	sp := spawner(t)
	sock, err := sp.Start(ctx, shell("late", "echo bye; exit 7"))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond) // the agent, a separate process, must exit before anyone attaches
	c, err := Dial(ctx, sock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()
	out, ev := readUntil(t, c, "\x00never")
	if !ev.Exited || ev.ExitCode != 7 || !strings.Contains(out, "bye") {
		t.Fatalf("got exit=%v code=%d output=%q", ev.Exited, ev.ExitCode, out)
	}
}

func TestLaunchCommandForAnotherUserResetsTheEnvironment(t *testing.T) {
	t.Setenv("ASTRA_TEST_DAEMON_SECRET", "must-not-leak")
	sp := Spawner{SocketDir: "/s", Executable: "/usr/local/bin/astrad", Scrollback: 1024, Linger: time.Minute}
	spec := Spec{ID: "a", User: "astra-agent", Argv: []string{"claude", "--flag"}, Dir: "/w", Env: []string{"HOME=/h", "TERM=x"}, Rows: 24, Cols: 80}
	cmd, err := sp.launchCommand(t.Context(), spec, "/s/a.sock")
	if err != nil {
		t.Fatal(err)
	}
	line := strings.Join(cmd.Args, " ")
	for _, want := range []string{
		"sudo -n -u astra-agent -- /usr/bin/env -i HOME=/h TERM=x /usr/local/bin/astrad session-host --detach",
		"--socket /s/a.sock", "--socket-mode 660", "-- claude --flag",
	} {
		if !strings.Contains(line, want) {
			t.Errorf("command line lacks %q:\n%s", want, line)
		}
	}
	if strings.Contains(strings.Join(cmd.Env, " "), "must-not-leak") || len(cmd.Env) != 1 {
		t.Errorf("sudo itself must start with a minimal environment, got %v", cmd.Env)
	}
}

func TestLaunchCommandForTheSameUserRunsTheHostDirectly(t *testing.T) {
	sp := Spawner{SocketDir: "/s", Executable: "/usr/local/bin/astrad", Scrollback: 1024, Linger: time.Minute}
	cmd, err := sp.launchCommand(t.Context(), Spec{ID: "a", Argv: []string{"sh"}, Env: []string{"A=b"}}, "/s/a.sock")
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Args[0] != "/usr/local/bin/astrad" || strings.Contains(strings.Join(cmd.Args, " "), "socket-mode") {
		t.Fatalf("unexpected command: %v", cmd.Args)
	}
	if len(cmd.Env) != 1 || cmd.Env[0] != "A=b" {
		t.Fatalf("the agent must get exactly its spec environment, got %v", cmd.Env)
	}
}

func TestLaunchCommandRejectsUserNamesThatCouldInjectArguments(t *testing.T) {
	sp := Spawner{SocketDir: "/s", Executable: "/usr/local/bin/astrad"}
	for _, name := range []string{"root;id", "-u", "A", "../x", "a b", strings.Repeat("a", 33)} {
		if _, err := sp.launchCommand(t.Context(), Spec{ID: "a", User: name, Argv: []string{"sh"}}, "/s/a.sock"); err == nil {
			t.Errorf("user %q should be rejected", name)
		}
	}
}

func TestStartRejectsUnsafeIDs(t *testing.T) {
	sp := spawner(t)
	for _, id := range []string{"", "../x", "A", strings.Repeat("a", 41)} {
		if _, err := sp.Start(t.Context(), shell(id, "true")); err == nil {
			t.Errorf("id %q should be rejected", id)
		}
	}
}

// Serving in-process lets the goroutine leak profile see every goroutine the host owns.
func TestServeLeavesNoGoroutinesBehind(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	sock := filepath.Join(socketDir(t), "leak.sock")
	cfg := HostConfig{
		Socket: sock, Argv: []string{"/bin/sh", "-c", "exec cat"}, Dir: "/",
		Env: []string{"PATH=/usr/bin:/bin"}, Rows: 24, Cols: 80, Scrollback: 4 << 10,
		Linger: time.Second, Logger: slog.New(slog.DiscardHandler),
	}
	served := make(chan error, 1)
	go func() { served <- Serve(ctx, cfg) }()
	if err := WaitReady(ctx, sock); err != nil {
		t.Fatal(err)
	}
	c, err := Dial(ctx, sock)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = c.Write([]byte("hello\r"))
	readUntil(t, c, "hello")
	cancel()
	if err := <-served; err != nil {
		t.Fatal(err)
	}
	_ = c.Close()

	var profile bytes.Buffer
	if err := pprof.Lookup("goroutineleak").WriteTo(&profile, 1); err != nil {
		t.Fatal(err)
	}
	if first, _, _ := strings.Cut(profile.String(), "\n"); !strings.HasSuffix(first, "total 0") {
		t.Fatalf("leaked goroutines:\n%s", profile.String())
	}
	if _, err := os.Stat(sock); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("socket should be removed after the host stops: %v", err)
	}
}

func TestHostStopsWhenItsSocketIsRemoved(t *testing.T) {
	old := socketCheck
	socketCheck = 50 * time.Millisecond
	t.Cleanup(func() { socketCheck = old })

	sock := filepath.Join(socketDir(t), "gone.sock")
	cfg := HostConfig{
		Socket: sock, Argv: []string{"/bin/sh", "-c", "exec cat"}, Dir: "/",
		Env: []string{"PATH=/usr/bin:/bin"}, Rows: 24, Cols: 80, Scrollback: 4 << 10,
		Linger: 100 * time.Millisecond, Logger: slog.New(slog.DiscardHandler),
	}
	served := make(chan error, 1)
	go func() { served <- Serve(t.Context(), cfg) }()
	if err := WaitReady(t.Context(), sock); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(sock); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-served:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("host kept running after its socket was removed")
	}
}

func TestScrollbackKeepsTheNewestBytes(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		writes []string
		want   string
	}{
		{"fits", 8, []string{"abc", "de"}, "abcde"},
		{"wraps", 8, []string{"abcdef", "ghij"}, "cdefghij"},
		{"exactly full", 4, []string{"ab", "cd"}, "abcd"},
		{"one write larger than the buffer", 4, []string{"abcdefgh"}, "efgh"},
		{"many wraps", 3, []string{"ab", "cd", "ef", "g"}, "efg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newScrollback(tt.size)
			for _, w := range tt.writes {
				s.write([]byte(w))
			}
			if got := string(s.snapshot()); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFramesRoundTripAndRejectOversizedPayloads(t *testing.T) {
	var buf bytes.Buffer
	if err := writeChunked(&buf, frameOutput, bytes.Repeat([]byte("x"), maxFrame+10)); err != nil {
		t.Fatal(err)
	}
	total := 0
	for buf.Len() > 0 {
		kind, p, err := readFrame(&buf)
		if err != nil || kind != frameOutput {
			t.Fatalf("kind=%c err=%v", kind, err)
		}
		total += len(p)
	}
	if total != maxFrame+10 {
		t.Fatalf("got %d bytes back", total)
	}
	if err := writeFrame(io.Discard, frameOutput, make([]byte, maxFrame+1)); !errors.Is(err, errFrameTooLarge) {
		t.Fatalf("oversized frame: %v", err)
	}
	header := []byte{frameOutput, 0xff, 0xff, 0xff, 0xff}
	if _, _, err := readFrame(bytes.NewReader(header)); !errors.Is(err, errFrameTooLarge) {
		t.Fatalf("oversized header: %v", err)
	}
}
