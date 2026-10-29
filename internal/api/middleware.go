package api

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/alfarax9/astra/internal/auth"
)

// protect wraps every route: panics become a 500, answers are never cached or sniffed, a
// listed browser origin gets CORS, and each request is logged without its query or headers.
func (s *Server) protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &recorder{ResponseWriter: w, status: http.StatusOK}
		start := s.cfg.Now()
		defer func() {
			if v := recover(); v != nil {
				s.cfg.Log.Error("handler panicked", "path", r.URL.Path, "panic", v)
				if !rec.wrote {
					writeJSON(rec, http.StatusInternalServerError, map[string]any{"error": map[string]string{"code": "internal", "message": "internal error"}})
				}
			}
			s.cfg.Log.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status,
				"took", s.cfg.Now().Sub(start).Round(time.Millisecond))
		}()
		h := rec.Header()
		h.Set("Cache-Control", "no-store")
		h.Set("X-Content-Type-Options", "nosniff")
		if origin := r.Header.Get("Origin"); origin != "" && slices.Contains(s.cfg.Origins, origin) {
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Vary", "Origin")
			if r.Method == http.MethodOptions {
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				h.Set("Access-Control-Max-Age", "600")
				rec.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(rec, r)
	})
}

// recorder remembers the status for the log, and still lets a handler take over the
// connection, which a WebSocket needs.
type recorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *recorder) WriteHeader(status int) {
	if !r.wrote {
		r.status, r.wrote = status, true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *recorder) Write(b []byte) (int, error) {
	r.wrote = true
	return r.ResponseWriter.Write(b)
}

// Hijack hands the connection to a WebSocket and records its upgrade.
func (r *recorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := http.NewResponseController(r.ResponseWriter).Hijack()
	if err == nil {
		r.status, r.wrote = http.StatusSwitchingProtocols, true
	}
	return conn, rw, err
}

// Unwrap lets http.ResponseController reach the real writer for flushing and deadlines.
func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// subprotocolPrefix marks a token carried in Sec-WebSocket-Protocol, because a browser cannot
// set the Authorization header on a WebSocket.
const subprotocolPrefix = "bearer."

// requireDevice runs h only for a paired device. A WebSocket route may also carry the token in
// its subprotocol list.
func (s *Server) requireDevice(h http.HandlerFunc, socket bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearer(r.Header.Get("Authorization"))
		if token == "" && socket {
			token = subprotocolToken(r.Header.Values("Sec-WebSocket-Protocol"))
		}
		if _, err := s.auth.Authenticate(r.Context(), token); err != nil {
			if !errors.Is(err, auth.ErrUnauthorized) {
				s.cfg.Log.Error("authenticate", "err", err)
			}
			w.Header().Set("WWW-Authenticate", "Bearer")
			s.fail(w, r, auth.ErrUnauthorized)
			return
		}
		h(w, r)
	})
}

func bearer(header string) string {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

func subprotocolToken(values []string) string {
	for _, v := range values {
		for _, p := range strings.Split(v, ",") {
			if token, ok := strings.CutPrefix(strings.TrimSpace(p), subprotocolPrefix); ok {
				return token
			}
		}
	}
	return ""
}
