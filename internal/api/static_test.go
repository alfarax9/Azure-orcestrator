package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func serveWith(t *testing.T, files fstest.MapFS) *httptest.Server {
	t.Helper()
	f := &fakes{stream: newFakeStream(), static: files}
	srv := New(Config{Version: "test", Log: slog.New(slog.DiscardHandler), Static: f.static}, f, f, f, f, f, f)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func get(t *testing.T, url string) (int, http.Header, string) {
	t.Helper()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, string(body)
}

const app = "<h1>Astra</h1>"

func newApp(t *testing.T) *httptest.Server {
	t.Helper()
	return serveWith(t, fstest.MapFS{"index.html": {Data: []byte(app)}, "assets/app-1a2b.js": {Data: []byte("console.log(1)")}})
}

func TestThePageIsServedWithItsProtections(t *testing.T) {
	ts := newApp(t)
	status, h, body := get(t, ts.URL+"/")
	if status != 200 || body != app {
		t.Fatalf("/ = %d %q", status, body)
	}
	for _, k := range []string{"Content-Security-Policy", "X-Frame-Options", "Referrer-Policy"} {
		if h.Get(k) == "" {
			t.Errorf("the page has no %s", k)
		}
	}
	for _, p := range []string{"/#/workers/x", "/workers/abc"} {
		if status, _, body := get(t, ts.URL+p); status != 200 || body != app {
			t.Errorf("%s = %d %q; a path the app owns is the app", p, status, body)
		}
	}
}

func TestAssetsAreCachedForeverAndMissingOnesAre404(t *testing.T) {
	ts := newApp(t)
	status, h, _ := get(t, ts.URL+"/assets/app-1a2b.js")
	if status != 200 || h.Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Errorf("an asset = %d, cache %q", status, h.Get("Cache-Control"))
	}
	if h.Get("Content-Security-Policy") != "" {
		t.Error("only the page carries the page policy")
	}
	if status, _, _ := get(t, ts.URL+"/assets/missing.js"); status != 404 {
		t.Errorf("a missing asset = %d, want 404", status)
	}
}

func TestNothingOutsideTheWebFolderIsServed(t *testing.T) {
	ts := newApp(t)
	for _, p := range []string{"/../../etc/passwd", "/%2e%2e/%2e%2e/etc/passwd", "/assets/../../etc/passwd"} {
		// Either the app (the path was cleaned) or a refusal; never anything else.
		if status, _, body := get(t, ts.URL+p); status == 200 && body != app {
			t.Errorf("%s served %q, want only the app", p, body)
		}
	}
}

func TestTheAPIKeepsItsOwnAnswersWhenThePageIsServed(t *testing.T) {
	ts := serveWith(t, fstest.MapFS{"index.html": {Data: []byte("app")}})
	if status, _, body := get(t, ts.URL+"/api/v1/version"); status != 200 || body == "app" {
		t.Errorf("version = %d %q", status, body)
	}
	if status, _, body := get(t, ts.URL+"/api/v1/nothing"); status != 404 || body == "app" {
		t.Errorf("an unknown API route = %d %q; it must not be the app", status, body)
	}
}

func TestThePageAnswersHEADToo(t *testing.T) {
	ts := newApp(t)
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodHead, ts.URL+"/", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 || resp.Header.Get("Content-Security-Policy") == "" {
		t.Errorf("HEAD / = %d, policy %q", resp.StatusCode, resp.Header.Get("Content-Security-Policy"))
	}
	post, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, ts.URL+"/", nil)
	if r, err := http.DefaultClient.Do(post); err == nil {
		_ = r.Body.Close()
		if r.StatusCode == 200 {
			t.Error("POST / must not be the app")
		}
	}
}

func TestWithoutAWebUIThePathsAreNotFound(t *testing.T) {
	ts, _ := newServer(t)
	if status, _, _ := get(t, ts.URL+"/"); status != 404 {
		t.Errorf("/ = %d", status)
	}
}

func TestTheChecklistIsServedToPairedDevicesOnly(t *testing.T) {
	ts, _ := newServer(t)
	status, body := call(t, "GET", ts.URL+"/api/v1/checklist", goodToken, "")
	if status != 200 || body["percent"] != float64(100) || body["complete"] != true {
		t.Errorf("checklist = %d %v", status, body)
	}
}
