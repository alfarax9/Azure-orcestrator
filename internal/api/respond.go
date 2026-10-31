package api

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/alfarax9/astra/internal/agent"
	"github.com/alfarax9/astra/internal/auth"
	"github.com/alfarax9/astra/internal/domain"
	"github.com/alfarax9/astra/internal/workspace"
)

// maxBody bounds a request body; nothing a client sends is large.
const maxBody = 1 << 20

var (
	errNoRoute  = errors.New("no such route")
	errBadBody  = errors.New("the request body is not valid for this route")
	errBadQuery = errors.New("the query is not valid for this route")
)

// problem is how a failed request is answered.
type problem struct {
	status  int
	code    string
	message string
}

// classify turns an error into the answer a client gets. Only errors written for people reach
// the message; anything else is a plain "internal error", and the cause goes to the log.
func classify(err error) (problem, bool) {
	switch {
	case errors.Is(err, errNoRoute), errors.Is(err, domain.ErrNotFound):
		return problem{http.StatusNotFound, "not_found", "not found"}, false
	case errors.Is(err, auth.ErrUnauthorized):
		return problem{http.StatusUnauthorized, "unauthorized", "pair this device first"}, false
	case errors.Is(err, domain.ErrConflict):
		return problem{http.StatusConflict, "conflict", "that name is already taken, or its parent does not exist"}, false
	case errors.Is(err, agent.ErrNoSession):
		return problem{http.StatusConflict, "no_session", "the worker's agent is not running"}, false
	case errors.Is(err, workspace.ErrClone):
		return problem{http.StatusUnprocessableEntity, "clone_failed", gitMessage(err)}, false
	case errors.Is(err, workspace.ErrInvalid), errors.Is(err, agent.ErrInvalid), errors.Is(err, auth.ErrInvalid), errors.Is(err, errBadBody), errors.Is(err, errBadQuery):
		return problem{http.StatusBadRequest, "invalid", err.Error()}, false
	}
	return problem{http.StatusInternalServerError, "internal", "internal error"}, true
}

// gitMessage keeps what git said about why a clone failed (its "fatal:" line), which is for
// the person adding the project, and leaves out the paths around it.
func gitMessage(err error) string {
	for _, line := range strings.Split(err.Error(), "\n") {
		if _, msg, ok := strings.Cut(line, "fatal: "); ok {
			return "git: " + strings.TrimSpace(msg)
		}
	}
	return "the repository could not be cloned"
}

// fail answers a failed request and logs the cause when it was not the client's doing.
func (s *Server) fail(w http.ResponseWriter, r *http.Request, err error) {
	p, internal := classify(err)
	if internal {
		s.cfg.Log.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	}
	writeJSON(w, p.status, map[string]any{"error": map[string]string{"code": p.code, "message": p.message}})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(append(b, '\n')) //nolint:gosec // JSON from json.Marshal, served as application/json
}

// decode reads a JSON body into v, refusing unknown fields and bodies over maxBody.
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	if err := json.UnmarshalRead(r.Body, v, json.RejectUnknownMembers(true)); err != nil {
		return fmt.Errorf("%w: %w", errBadBody, err)
	}
	return nil
}
