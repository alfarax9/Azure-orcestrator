package main

import (
	"context"
	"encoding/json/v2"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// freeAddr returns a loopback address nobody is listening on.
func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	return ln.Addr().String()
}

type client struct {
	t     *testing.T
	base  string
	token string
}

// do sends a JSON request and decodes the answer.
func (c *client) do(method, path, body string, out any) int {
	c.t.Helper()
	req, err := http.NewRequestWithContext(c.t.Context(), method, "http://"+c.base+"/api/v1"+path, strings.NewReader(body))
	if err != nil {
		c.t.Fatal(err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if out != nil {
		_ = json.UnmarshalRead(resp.Body, out)
	}
	return resp.StatusCode
}

// startDaemon runs serve on a free loopback port until the test ends.
func startDaemon(t *testing.T) *client {
	t.Helper()
	localEnv(t)
	addr := freeAddr(t)
	t.Setenv("ASTRA_LISTEN", addr)
	ctx, stop := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() { served <- serve(ctx, slog.New(slog.DiscardHandler)) }()
	t.Cleanup(func() {
		stop()
		if err := <-served; err != nil {
			t.Errorf("serve = %v", err)
		}
	})
	for range 100 { // wait for the listener
		conn, err := (&net.Dialer{Timeout: time.Second}).DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	return &client{t: t, base: addr}
}

// pairDevice makes a pairing token the way `astrad pair` does and trades it for a device token.
func pairDevice(t *testing.T, c *client) {
	t.Helper()
	a, err := open(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	pairing, err := a.auth.NewPairing(t.Context())
	_ = a.store.Close()
	if err != nil {
		t.Fatal(err)
	}
	var paired struct {
		Token string `json:"token"`
	}
	body := `{"token":"` + pairing.Reveal() + `","name":"Test laptop"}`
	if status := c.do("POST", "/pairing", body, &paired); status != 200 || paired.Token == "" {
		t.Fatalf("pairing = %d %+v", status, paired)
	}
	if status := c.do("POST", "/pairing", body, nil); status != 401 {
		t.Errorf("the pairing token worked twice: %d", status)
	}
	if status := c.do("GET", "/workspaces", "", nil); status != 401 {
		t.Errorf("without a token = %d", status)
	}
	c.token = paired.Token
}

// dialSocket opens a WebSocket carrying the device token in the subprotocol, as a browser does.
func dialSocket(t *testing.T, c *client, path string) *websocket.Conn {
	t.Helper()
	conn, resp, err := websocket.Dial(t.Context(), "ws://"+c.base+"/api/v1"+path, &websocket.DialOptions{
		Subprotocols: []string{"astra.v1", "bearer." + c.token}})
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.CloseNow() })
	return conn
}

type object struct {
	ID string `json:"id"`
}

func TestTheDaemonServesTheWholeFlowOverHTTPAndWebSockets(t *testing.T) {
	c := startDaemon(t)
	pairDevice(t, c)
	watch := newEventWatch(t, dialSocket(t, c, "/events"))
	watch.waitFor("the empty snapshot", func(m map[string]any) bool { return m["type"] == "snapshot" })

	var ws, project, worker object
	if c.do("POST", "/workspaces", `{"slug":"acme","name":"Acme"}`, &ws) != 201 ||
		c.do("POST", "/projects", `{"workspace":"acme","name":"site"}`, &project) != 201 ||
		c.do("POST", "/projects/"+project.ID+"/workers", `{"title":"Fix the header","agent":"claude"}`, &worker) != 201 {
		t.Fatal("creating the objects failed")
	}
	watch.waitFor("the worker appearing", func(m map[string]any) bool {
		d, _ := m["data"].(map[string]any)
		return m["type"] == "worker" && d["id"] == worker.ID
	})

	if status := c.do("POST", "/workers/"+worker.ID+"/start", "", nil); status != 200 {
		t.Fatalf("start = %d", status)
	}
	watch.waitFor("the agent asking to trust the folder", func(m map[string]any) bool { return activity(m, worker.ID) == "needs_input" })
	answerInTheTerminal(t, c, worker.ID)

	if status := c.do("POST", "/workers/"+worker.ID+"/messages", `{"text":"Fix the header"}`, nil); status != 204 {
		t.Fatalf("message = %d", status)
	}
	watch.waitFor("the finished turn with its two added lines", func(m map[string]any) bool {
		d, _ := m["data"].(map[string]any)
		return activity(m, worker.ID) == "idle" && d["added"] == float64(2)
	})
	if status := c.do("POST", "/workers/"+worker.ID+"/stop", "", nil); status != 204 {
		t.Errorf("stop = %d", status)
	}
}

// answerInTheTerminal is the person pressing Enter on the folder-trust question in the
// terminal tab: the question is repainted from the scrollback, and a resize and the answer go in.
func answerInTheTerminal(t *testing.T, c *client, id string) {
	t.Helper()
	term := dialSocket(t, c, "/sessions/"+id+"/pty")
	if screen := readTerminalUntil(t, term, "trust the files"); !strings.Contains(screen, "Do you trust the files in this folder?") {
		t.Errorf("the terminal did not repaint the question: %q", screen)
	}
	if err := term.Write(t.Context(), websocket.MessageText, []byte(`{"type":"resize","rows":30,"cols":100}`)); err != nil {
		t.Fatal(err)
	}
	if err := term.Write(t.Context(), websocket.MessageBinary, []byte("\r")); err != nil {
		t.Fatal(err)
	}
	readTerminalUntil(t, term, "> ")
}

func activity(m map[string]any, id string) any {
	d, _ := m["data"].(map[string]any)
	if m["type"] != "worker" || d["id"] != id {
		return nil
	}
	return d["agentActivity"]
}

// eventWatch reads the events socket in the background and lets the test wait for a message.
type eventWatch struct {
	t    *testing.T
	msgs chan map[string]any
}

func newEventWatch(t *testing.T, c *websocket.Conn) *eventWatch {
	w := &eventWatch{t: t, msgs: make(chan map[string]any, 256)}
	go func() {
		for {
			_, data, err := c.Read(context.Background())
			if err != nil {
				close(w.msgs)
				return
			}
			var m map[string]any
			if json.Unmarshal(data, &m) == nil {
				w.msgs <- m
			}
		}
	}()
	return w
}

func (w *eventWatch) waitFor(what string, ok func(map[string]any) bool) {
	w.t.Helper()
	timeout := time.After(20 * time.Second)
	for {
		select {
		case m, open := <-w.msgs:
			if !open {
				w.t.Fatalf("the events socket closed while waiting for %s", what)
			}
			if ok(m) {
				return
			}
		case <-timeout:
			w.t.Fatalf("timed out waiting for %s", what)
		}
	}
}

// readTerminalUntil reads terminal output until it holds want, ignoring text frames.
func readTerminalUntil(t *testing.T, c *websocket.Conn, want string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	var out strings.Builder
	for !strings.Contains(out.String(), want) {
		kind, data, err := c.Read(ctx)
		if err != nil {
			t.Fatalf("waiting for %q on the terminal, got %q: %v", want, out.String(), err)
		}
		if kind == websocket.MessageBinary {
			out.Write(data)
		}
	}
	return out.String()
}

func TestOnlyLoopbackAddressesAreServedUnlessRemoteIsAllowed(t *testing.T) {
	for addr, remote := range map[string]bool{
		"127.0.0.1:7437": false, "localhost:7437": false, "[::1]:7437": false,
		"0.0.0.0:7437": true, "10.0.0.5:7437": true, "example.com:7437": true,
	} {
		err := checkListen(addr, false)
		if (err != nil) != remote {
			t.Errorf("checkListen(%q, false) = %v", addr, err)
		}
		if err := checkListen(addr, true); err != nil {
			t.Errorf("checkListen(%q, true) = %v", addr, err)
		}
	}
	if err := checkListen("no-port", false); err == nil {
		t.Error("an address without a port must be refused")
	}
	if got := splitList(" http://a , ,http://b "); len(got) != 2 || got[0] != "http://a" || got[1] != "http://b" {
		t.Errorf("splitList = %q", got)
	}
}
