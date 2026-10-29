package api

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"

	"github.com/alfarax9/astra/internal/domain"
)

const (
	// subprotocol is what the server answers; the client must offer it, next to the optional
	// bearer.<token> entry that carries its token.
	subprotocol = "astra.v1"
	// eventQueue is how many changes wait for a slow client before it is cut off.
	eventQueue   = 256
	writeTimeout = 10 * time.Second
)

// event is one message on the events socket.
type event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// snapshot is every fact, sent first so a client starts from the whole picture.
type snapshot struct {
	Workspaces []domain.Workspace `json:"workspaces"`
	Projects   []domain.Project   `json:"projects"`
	Workers    []domain.Worker    `json:"workers"`
}

// accept upgrades the request to a WebSocket that checks origins against the configured list.
func (s *Server) accept(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols:   []string{subprotocol},
		OriginPatterns: s.originHosts(),
	})
}

// events streams the facts: a snapshot, then every change as it happens (ADR 0006). A client
// that cannot keep up is cut off and reconnects to a fresh snapshot, so the queue is bounded.
func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	conn, err := s.accept(w, r)
	if err != nil {
		return // Accept already answered
	}
	defer func() { _ = conn.CloseNow() }()
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	ctx = conn.CloseRead(ctx) // a client sends nothing; this answers its pings and sees it leave

	changes := make(chan domain.Change, eventQueue)
	var overflow atomic.Bool
	unsubscribe := s.reads.Subscribe(func(c domain.Change) {
		select {
		case changes <- c:
		default:
			overflow.Store(true)
			cancel()
		}
	})
	defer unsubscribe()

	// Subscribed before the snapshot is read, so no change falls between the two; a change
	// that is in both arrives twice and replaces itself.
	snap, err := s.snapshot(ctx)
	if err != nil {
		s.cfg.Log.Error("events snapshot", "err", err)
		_ = conn.Close(websocket.StatusInternalError, "internal error")
		return
	}
	if !send(ctx, conn, event{Type: "snapshot", Data: snap}) {
		return
	}
	for {
		select {
		case <-ctx.Done():
			if overflow.Load() {
				_ = conn.Close(websocket.StatusTryAgainLater, "too slow; reconnect for a fresh snapshot")
			}
			return
		case c := <-changes:
			if !send(ctx, conn, changeEvent(c)) {
				return
			}
		}
	}
}

func (s *Server) snapshot(ctx context.Context) (snapshot, error) {
	snap := snapshot{Workspaces: []domain.Workspace{}, Projects: []domain.Project{}, Workers: []domain.Worker{}}
	var err error
	if snap.Workspaces, err = s.reads.Workspaces(ctx); err != nil {
		return snap, err
	}
	for _, ws := range snap.Workspaces {
		ps, err := s.reads.Projects(ctx, ws.ID)
		if err != nil {
			return snap, err
		}
		snap.Projects = append(snap.Projects, ps...)
	}
	snap.Workers, err = s.reads.Workers(ctx, "")
	snap.Workspaces, snap.Projects, snap.Workers = nonNil(snap.Workspaces), nonNil(snap.Projects), nonNil(snap.Workers)
	return snap, err
}

func changeEvent(c domain.Change) event {
	switch {
	case c.Workspace != nil:
		return event{Type: "workspace", Data: c.Workspace}
	case c.Project != nil:
		return event{Type: "project", Data: c.Project}
	}
	return event{Type: "worker", Data: c.Worker}
}

// send writes one JSON message, and reports whether the client is still there.
func send(ctx context.Context, conn *websocket.Conn, v any) bool {
	b, err := json.Marshal(v)
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	return conn.Write(ctx, websocket.MessageText, b) == nil
}
