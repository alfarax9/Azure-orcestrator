package api

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/alfarax9/astra/internal/agent"
	"github.com/alfarax9/astra/internal/domain"
)

// dial opens a WebSocket to path with the given headers and subprotocols.
// It returns the HTTP status of the handshake (0 if there was none).
func dial(t *testing.T, ts string, path string, opts *websocket.DialOptions) (*websocket.Conn, int, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)
	c, resp, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(ts, "http")+path, opts)
	status := 0
	if resp != nil {
		status = resp.StatusCode
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}
	return c, status, err
}

func authHeader() http.Header { return http.Header{"Authorization": {"Bearer " + goodToken}} }

// next reads one JSON text message.
func next(t *testing.T, c *websocket.Conn) map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	kind, data, err := c.Read(ctx)
	if err != nil || kind != websocket.MessageText {
		t.Fatalf("read = %v, %v", kind, err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("not JSON: %s", data)
	}
	return m
}

func TestEventsSendASnapshotThenEveryChange(t *testing.T) {
	ts, f := newServer(t)
	f.workspaces = []domain.Workspace{{ID: "ws-acme", Slug: "acme", Name: "Acme"}}
	f.projects = []domain.Project{{ID: "p1", WorkspaceID: "ws-acme", Name: "site"}}
	f.workers = []domain.Worker{{ID: "w1", ProjectID: "p1", Title: "Fix"}}

	c, _, err := dial(t, ts.URL, "/api/v1/events", &websocket.DialOptions{HTTPHeader: authHeader(), Subprotocols: []string{"astra.v1"}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.CloseNow() }()
	if c.Subprotocol() != "astra.v1" {
		t.Errorf("subprotocol = %q", c.Subprotocol())
	}
	snap := next(t, c)
	data, _ := snap["data"].(map[string]any)
	if snap["type"] != "snapshot" || len(data["workspaces"].([]any)) != 1 || len(data["projects"].([]any)) != 1 || len(data["workers"].([]any)) != 1 {
		t.Fatalf("snapshot = %v", snap)
	}
	if w := data["workers"].([]any)[0].(map[string]any); w["title"] != "Fix" || w["worktree"] != nil {
		t.Errorf("worker in snapshot = %v (the worktree path must not be sent)", w)
	}

	waitSubscribed(t, f)
	f.publish(domain.Change{Worker: &domain.Worker{ID: "w1", Activity: domain.ActivityWorking}})
	f.publish(domain.Change{Project: &domain.Project{ID: "p1", Name: "site"}})
	f.publish(domain.Change{Workspace: &domain.Workspace{ID: "ws-acme", Slug: "acme"}})
	for _, want := range []string{"worker", "project", "workspace"} {
		if m := next(t, c); m["type"] != want || m["data"] == nil {
			t.Errorf("message = %v, want a %s", m, want)
		}
	}
}

func waitSubscribed(t *testing.T, f *fakes) {
	t.Helper()
	for range 200 {
		if f.listenerCount() > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("the events socket never subscribed")
}

func TestEventsOfAnEmptyServerAreEmptyListsNotNull(t *testing.T) {
	ts, _ := newServer(t)
	c, _, err := dial(t, ts.URL, "/api/v1/events", &websocket.DialOptions{HTTPHeader: authHeader(), Subprotocols: []string{"astra.v1"}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.CloseNow() }()
	data := next(t, c)["data"].(map[string]any)
	for _, k := range []string{"workspaces", "projects", "workers"} {
		if list, ok := data[k].([]any); !ok || len(list) != 0 {
			t.Errorf("%s = %v, want []", k, data[k])
		}
	}
}

func TestAWebSocketCanCarryItsTokenInTheSubprotocol(t *testing.T) {
	ts, _ := newServer(t)
	c, _, err := dial(t, ts.URL, "/api/v1/events", &websocket.DialOptions{Subprotocols: []string{"astra.v1", "bearer." + goodToken}})
	if err != nil {
		t.Fatalf("a browser-style handshake failed: %v", err)
	}
	defer func() { _ = c.CloseNow() }()
	if c.Subprotocol() != "astra.v1" {
		t.Errorf("the server answered %q; it must never echo the token", c.Subprotocol())
	}
	for name, opts := range map[string]*websocket.DialOptions{
		"no token":    {Subprotocols: []string{"astra.v1"}},
		"wrong token": {Subprotocols: []string{"astra.v1", "bearer.wrong"}},
	} {
		if _, status, err := dial(t, ts.URL, "/api/v1/events", opts); err == nil || status != 401 {
			t.Errorf("%s: err = %v, status = %d, want a 401", name, err, status)
		}
	}
}

func TestAWebSocketFromAnUnlistedBrowserOriginIsRefused(t *testing.T) {
	ts, _ := newServer(t, "http://localhost:5173")
	header := authHeader()
	header.Set("Origin", "https://evil.example")
	if _, status, err := dial(t, ts.URL, "/api/v1/events", &websocket.DialOptions{HTTPHeader: header, Subprotocols: []string{"astra.v1"}}); err == nil || status != 403 {
		t.Errorf("an unlisted origin: err = %v, status = %d, want a 403", err, status)
	}
	header.Set("Origin", "http://localhost:5173")
	c, _, err := dial(t, ts.URL, "/api/v1/events", &websocket.DialOptions{HTTPHeader: header, Subprotocols: []string{"astra.v1"}})
	if err != nil {
		t.Fatalf("a listed origin: %v", err)
	}
	_ = c.CloseNow()
}

func TestTerminalRelaysBytesBothWaysAndTheExit(t *testing.T) {
	ts, f := newServer(t)
	f.workers = []domain.Worker{{ID: "w1"}}
	f.stream.signals <- agent.Signal{Output: []byte("scrollback ")}
	f.stream.signals <- agent.Signal{Event: &agent.Event{Kind: agent.KindPrompt}} // not for the terminal

	c, _, err := dial(t, ts.URL, "/api/v1/sessions/w1/pty", &websocket.DialOptions{HTTPHeader: authHeader(), Subprotocols: []string{"astra.v1"}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.CloseNow() }()
	read := func() (websocket.MessageType, []byte) {
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()
		kind, data, err := c.Read(ctx)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		return kind, data
	}
	if kind, data := read(); kind != websocket.MessageBinary || string(data) != "scrollback " {
		t.Fatalf("first frame = %v %q", kind, data)
	}

	write := func(kind websocket.MessageType, data string) {
		if err := c.Write(t.Context(), kind, []byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	write(websocket.MessageBinary, "ls\r")
	write(websocket.MessageText, `{"type":"resize","rows":50,"cols":132}`)
	write(websocket.MessageText, `{"type":"resize","rows":0,"cols":132}`) // refused: no such terminal
	write(websocket.MessageText, `not json`)                              // ignored
	waitFor(t, "the input and the resize reach the agent", func() bool {
		f.stream.mu.Lock()
		defer f.stream.mu.Unlock()
		return len(f.stream.writes) == 1 && len(f.stream.resizes) == 1
	})
	if got := f.stream.written(); got[0] != "ls\r" || f.stream.resizes[0] != [2]uint16{50, 132} {
		t.Errorf("writes %q, resizes %v", got, f.stream.resizes)
	}
	if f.lastCall() != "NoteInput w1 ls\r" {
		t.Errorf("the service was not told about the input: %q", f.lastCall())
	}

	f.stream.signals <- agent.Signal{Output: []byte("total 0\r\n")}
	f.stream.signals <- agent.Signal{Exited: true, ExitCode: 3}
	if kind, data := read(); kind != websocket.MessageBinary || string(data) != "total 0\r\n" {
		t.Errorf("output frame = %v %q", kind, data)
	}
	kind, data := read()
	var exit map[string]any
	if kind != websocket.MessageText || json.Unmarshal(data, &exit) != nil || exit["type"] != "exit" || exit["code"] != float64(3) {
		t.Errorf("exit frame = %v %s", kind, data)
	}
}

func waitFor(t *testing.T, what string, ok func() bool) {
	t.Helper()
	for range 300 {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestTerminalAnswersWithAStatusWhenThereIsNothingToOpen(t *testing.T) {
	ts, f := newServer(t)
	f.openErr = agent.ErrNoSession
	if _, status, err := dial(t, ts.URL, "/api/v1/sessions/w1/pty", &websocket.DialOptions{HTTPHeader: authHeader(), Subprotocols: []string{"astra.v1"}}); err == nil || status != 409 {
		t.Errorf("no session: err = %v, status = %d, want 409", err, status)
	}
	f.openErr = domain.ErrNotFound
	if _, status, err := dial(t, ts.URL, "/api/v1/sessions/nobody/pty", &websocket.DialOptions{HTTPHeader: authHeader(), Subprotocols: []string{"astra.v1"}}); err == nil || status != 404 {
		t.Errorf("no worker: err = %v, status = %d, want 404", err, status)
	}
}

func TestTerminalEndsWhenTheClientLeaves(t *testing.T) {
	ts, f := newServer(t)
	c, _, err := dial(t, ts.URL, "/api/v1/sessions/w1/pty", &websocket.DialOptions{HTTPHeader: authHeader(), Subprotocols: []string{"astra.v1"}})
	if err != nil {
		t.Fatal(err)
	}
	_ = c.Close(websocket.StatusNormalClosure, "")
	waitFor(t, "the session connection to be closed", func() bool {
		select {
		case _, open := <-f.stream.signals:
			return !open
		default:
			return false
		}
	})
}
