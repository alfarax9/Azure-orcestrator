package api

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// pageHeaders protect the page that holds the device token: nothing loads from elsewhere, it
// cannot be framed, and it sends no referrer. Styles may be inline because the terminal sets
// them on its own elements.
var pageHeaders = map[string]string{
	"Content-Security-Policy": "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' ws: wss:; frame-ancestors 'none'; base-uri 'none'",
	"X-Frame-Options":         "DENY",
	"Referrer-Policy":         "no-referrer",
}

// static serves the web UI: a file that exists is sent as it is, and any other path that is not
// the API gets index.html, so a reload on #/workers/<id> or a future route still opens the app.
func (s *Server) static() http.Handler {
	if s.cfg.Static == nil {
		return http.NotFoundHandler()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		info, err := fs.Stat(s.cfg.Static, name)
		switch {
		case name != "" && err == nil && !info.IsDir():
			if strings.HasPrefix(name, "assets/") {
				// Built assets carry a hash in their names, so they can be kept for ever.
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
		case name != "" && strings.Contains(path.Base(name), "."):
			http.NotFound(w, r) // a missing asset is a 404, not the app
			return
		default:
			name = "index.html"
		}
		if name == "index.html" {
			for k, v := range pageHeaders {
				w.Header().Set(k, v)
			}
		}
		http.ServeFileFS(w, r, s.cfg.Static, name) //nolint:gosec // a rooted fs.FS cannot leave its folder, and fs refuses ".."
	})
}
