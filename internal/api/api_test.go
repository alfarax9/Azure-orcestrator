package api

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alfarax9/astra/internal/agent"
	"github.com/alfarax9/astra/internal/domain"
	"github.com/alfarax9/astra/internal/workspace"
)

// call sends a request and returns the status and the decoded JSON body (nil when empty).
func call(t *testing.T, method, url, token, body string, header ...string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for i := 0; i+1 < len(header); i += 2 {
		req.Header.Set(header[i], header[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var out map[string]any
	_ = json.UnmarshalRead(resp.Body, &out)
	return resp.StatusCode, out
}

func errCode(body map[string]any) string {
	e, _ := body["error"].(map[string]any)
	code, _ := e["code"].(string)
	return code
}

func TestVersionIsPublicAndEverythingElseNeedsADevice(t *testing.T) {
	ts, _ := newServer(t)
	if status, body := call(t, "GET", ts.URL+"/api/v1/version", "", ""); status != 200 || body["version"] != "test" || body["api"] != float64(1) {
		t.Errorf("version = %d %v", status, body)
	}
	for _, route := range []string{
		"GET /api/v1/workspaces", "GET /api/v1/checklist", "POST /api/v1/workspaces", "GET /api/v1/projects?workspace=x", "POST /api/v1/projects",
		"PUT /api/v1/projects/p/commands", "GET /api/v1/workers", "GET /api/v1/workers/w", "GET /api/v1/usage", "POST /api/v1/projects/p/workers",
		"POST /api/v1/workers/w/start", "POST /api/v1/workers/w/stop", "POST /api/v1/workers/w/interrupt",
		"POST /api/v1/workers/w/messages", "GET /api/v1/events", "GET /api/v1/sessions/w/pty",
	} {
		method, path, _ := strings.Cut(route, " ")
		for _, token := range []string{"", "wrong"} {
			status, body := call(t, method, ts.URL+path, token, "{}")
			if status != 401 || errCode(body) != "unauthorized" {
				t.Errorf("%s with token %q = %d %v, want 401", route, token, status, body)
			}
		}
	}
	// A token in the wrong place does not count: only WebSocket routes read the subprotocol.
	status, _ := call(t, "GET", ts.URL+"/api/v1/workspaces", "", "", "Sec-WebSocket-Protocol", "astra.v1, bearer."+goodToken)
	if status != 401 {
		t.Errorf("a subprotocol token on a plain route = %d, want 401", status)
	}
}

func TestPairingGivesTheTokenOnceAndNeverLogsIt(t *testing.T) {
	ts, f := newServer(t)
	status, body := call(t, "POST", ts.URL+"/api/v1/pairing", "", `{"token":"pairing-token","name":"Laptop"}`)
	if status != 200 || body["token"] != goodToken {
		t.Fatalf("pairing = %d %v", status, body)
	}
	if dev, _ := body["device"].(map[string]any); dev["name"] != "Laptop" {
		t.Errorf("device = %v", body["device"])
	}
	if status, _ := call(t, "GET", ts.URL+"/api/v1/workspaces", goodToken, ""); status != 200 {
		t.Errorf("the new token is refused: %d", status)
	}
	for _, bad := range []string{`{"token":"wrong","name":"x"}`, `{"token":"","name":"x"}`} {
		if status, body := call(t, "POST", ts.URL+"/api/v1/pairing", "", bad); status != 401 || errCode(body) != "unauthorized" {
			t.Errorf("%s = %d %v", bad, status, body)
		}
	}
	if status, body := call(t, "POST", ts.URL+"/api/v1/pairing", "", `{"token":"pairing-token","name":""}`); status != 400 || errCode(body) != "invalid" {
		t.Errorf("a missing name = %d %v", status, body)
	}
	if logs := f.logged(); strings.Contains(logs, goodToken) || strings.Contains(logs, "pairing-token") {
		t.Errorf("a token reached the log: %s", logs)
	}
}

// apiDo sends an authenticated request under /api/v1.
func apiDo(t *testing.T, ts *httptest.Server, method, path, body string) (int, map[string]any) {
	t.Helper()
	return call(t, method, ts.URL+"/api/v1"+path, goodToken, body)
}

func TestWorkspacesAndProjectsThroughTheAPI(t *testing.T) {
	ts, f := newServer(t)
	if status, body := apiDo(t, ts, "POST", "/workspaces", `{"slug":"acme","name":"Acme"}`); status != 201 || body["slug"] != "acme" {
		t.Fatalf("add workspace = %d %v", status, body)
	}
	if status, body := apiDo(t, ts, "POST", "/projects", `{"workspace":"acme","name":"site","repoUrl":"https://x/y.git"}`); status != 201 || body["name"] != "site" {
		t.Fatalf("add project = %d %v", status, body)
	}
	if f.lastCall() != "AddProject acme site https://x/y.git" {
		t.Errorf("last call = %q", f.lastCall())
	}
	if status, _ := apiDo(t, ts, "GET", "/projects?workspace=acme", ""); status != 200 {
		t.Errorf("list projects = %d", status)
	}
	if status, body := apiDo(t, ts, "GET", "/projects?workspace=nobody", ""); status != 404 || errCode(body) != "not_found" {
		t.Errorf("projects of an unknown workspace = %d %v", status, body)
	}
	status, body := apiDo(t, ts, "PUT", "/projects/p-site/commands", `{"dev":"pnpm dev"}`)
	if cmds, _ := body["commands"].(map[string]any); status != 200 || cmds["dev"] != "pnpm dev" {
		t.Errorf("set commands = %d %v", status, body)
	}
}

func TestWorkersThroughTheAPI(t *testing.T) {
	ts, f := newServer(t)
	// Client review is on unless the client says otherwise.
	if status, _ := apiDo(t, ts, "POST", "/projects/p-site/workers", `{"title":"Fix header","agent":"claude"}`); status != 201 || !strings.HasSuffix(f.lastCall(), "review=true") {
		t.Errorf("new worker = %d, call %q", status, f.lastCall())
	}
	if status, _ := apiDo(t, ts, "POST", "/projects/p-site/workers", `{"title":"Internal","agent":"codex","clientReview":false}`); status != 201 || !strings.HasSuffix(f.lastCall(), "review=false") {
		t.Errorf("internal worker = %d, call %q", status, f.lastCall())
	}
	if status, body := apiDo(t, ts, "GET", "/workers/w1", ""); status != 200 || body["id"] != "w1" {
		t.Errorf("get worker = %d %v", status, body)
	}
	if status, _ := apiDo(t, ts, "GET", "/workers", ""); status != 200 {
		t.Errorf("list workers = %d", status)
	}
	if status, body := apiDo(t, ts, "GET", "/workers/nobody", ""); status != 404 || errCode(body) != "not_found" {
		t.Errorf("an unknown worker = %d %v", status, body)
	}
}

func TestStartTracksTheWorkerAndTheRestActsOnTheAgent(t *testing.T) {
	ts, f := newServer(t)
	f.workers = []domain.Worker{{ID: "w1"}}
	do := func(path, body string) int {
		status, _ := call(t, "POST", ts.URL+"/api/v1/workers/w1"+path, goodToken, body)
		return status
	}
	if do("/start", "") != 200 || len(f.tracked) != 1 || f.tracked[0] != "w1" {
		t.Errorf("start: tracked %v", f.tracked)
	}
	f.failWith = errors.New("unit exists")
	if status := do("/start", ""); status != 500 || len(f.tracked) != 1 {
		t.Errorf("a failed start = %d, tracked %v; it must not be tracked", status, f.tracked)
	}
	if do("/stop", "") != 204 || f.lastCall() != "Stop w1" {
		t.Errorf("stop: %q", f.lastCall())
	}
	if do("/interrupt", "") != 204 || f.lastCall() != "Interrupt w1" {
		t.Errorf("interrupt: %q", f.lastCall())
	}
	if do("/messages", `{"text":"Fix it"}`) != 204 || f.lastCall() != "Send w1 Fix it" {
		t.Errorf("message: %q", f.lastCall())
	}
}

func TestEveryFailureHasItsOwnAnswerAndInternalOnesStayInTheLog(t *testing.T) {
	ts, f := newServer(t)
	for _, c := range []struct {
		err    error
		status int
		code   string
	}{
		{domain.ErrNotFound, 404, "not_found"},
		{domain.ErrConflict, 409, "conflict"},
		{agent.ErrNoSession, 409, "no_session"},
		{agent.ErrInvalid, 400, "invalid"},
		{fmt.Errorf("%w: bad slug", workspace.ErrInvalid), 400, "invalid"},
		{fmt.Errorf("%w: /var/lib/astra/agents/acme: fatal: repository 'https://x/y.git/' not found", workspace.ErrClone), 422, "clone_failed"},
		{errors.New("sql: database is locked at /var/lib/astrad/astra.db"), 500, "internal"},
	} {
		f.failWith = c.err
		status, body := call(t, "POST", ts.URL+"/api/v1/workers/w1/stop", goodToken, "")
		if status != c.status || errCode(body) != c.code {
			t.Errorf("%v = %d %v, want %d %s", c.err, status, body, c.status, c.code)
		}
		if msg := fmt.Sprint(body); c.status == 500 && (strings.Contains(msg, "sql") || strings.Contains(msg, "/var/lib")) {
			t.Errorf("an internal detail reached the client: %s", msg)
		}
	}
	f.failWith = fmt.Errorf("%w: /var/lib/astra/agents/acme/work: fatal: repository 'https://x/y.git/' not found", workspace.ErrClone)
	_, body := call(t, "POST", ts.URL+"/api/v1/workers/w1/stop", goodToken, "")
	if msg := fmt.Sprint(body); !strings.Contains(msg, "repository 'https://x/y.git/' not found") || strings.Contains(msg, "/var/lib") {
		t.Errorf("a clone failure should say what git said and nothing about paths: %s", msg)
	}
	if !strings.Contains(f.logged(), "database is locked") {
		t.Error("the internal cause was not logged")
	}
}

func TestBodiesAreCheckedStrictly(t *testing.T) {
	ts, _ := newServer(t)
	for name, body := range map[string]string{
		"an unknown field": `{"slug":"a","name":"b","admin":true}`,
		"not json":         `slug=a`,
		"a wrong type":     `{"slug":1,"name":"b"}`,
		"empty":            ``,
		"too big":          `{"slug":"` + strings.Repeat("a", 2<<20) + `","name":"b"}`,
	} {
		if status, body := call(t, "POST", ts.URL+"/api/v1/workspaces", goodToken, body); status != 400 || errCode(body) != "invalid" {
			t.Errorf("%s = %d %v, want 400 invalid", name, status, body)
		}
	}
}

func TestUnknownRoutesAreJSON404sAndAnswersAreNotCached(t *testing.T) {
	ts, _ := newServer(t)
	req, _ := http.NewRequestWithContext(t.Context(), "GET", ts.URL+"/api/v1/nothing", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 404 || resp.Header.Get("Content-Type") != "application/json" ||
		resp.Header.Get("Cache-Control") != "no-store" || resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("status %d, headers %v", resp.StatusCode, resp.Header)
	}
}

func TestOnlyListedBrowserOriginsGetCORS(t *testing.T) {
	ts, _ := newServer(t, "http://localhost:5173")
	status, _ := call(t, "OPTIONS", ts.URL+"/api/v1/workspaces", "", "", "Origin", "http://localhost:5173")
	if status != 204 {
		t.Errorf("preflight = %d", status)
	}
	headers := func(origin string) http.Header {
		req, _ := http.NewRequestWithContext(t.Context(), "GET", ts.URL+"/api/v1/version", nil)
		req.Header.Set("Origin", origin)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		return resp.Header
	}
	if got := headers("http://localhost:5173").Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("a listed origin got %q", got)
	}
	if got := headers("https://evil.example").Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("an unlisted origin got %q", got)
	}
}
