package api

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"sync"

	"github.com/coder/websocket"

	"github.com/alfarax9/astra/internal/agent"
	"github.com/alfarax9/astra/internal/domain"
)

// maxInput bounds one frame a person types or pastes into the terminal.
const maxInput = 1 << 20

// resize is the text frame a client sends when its terminal changes size.
type resize struct {
	Type string `json:"type"`
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

// terminal connects a client to a worker's agent terminal (F3-02): binary frames are the
// terminal's bytes in both directions, the first ones being the host's scrollback; text frames
// carry resizes from the client and the agent's exit to it (ADR 0006).
func (s *Server) terminal(w http.ResponseWriter, r *http.Request) {
	id := domain.WorkerID(r.PathValue("id"))
	st, err := s.agents.Open(r.Context(), id)
	if err != nil {
		s.fail(w, r, err) // before the upgrade, so a client sees a status, not a closed socket
		return
	}
	defer func() { _ = st.Close() }()
	conn, err := s.accept(w, r)
	if err != nil {
		return
	}
	defer func() { _ = conn.CloseNow() }()
	conn.SetReadLimit(maxInput)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var wg sync.WaitGroup
	wg.Go(func() {
		defer cancel()
		s.relayOutput(ctx, conn, st)
	})
	s.relayInput(ctx, conn, st, id)
	cancel()
	_ = st.Close() // unblocks the output relay
	wg.Wait()
	_ = conn.Close(websocket.StatusNormalClosure, "")
}

// relayOutput copies what the agent writes to the client until the agent exits or the client
// leaves. Events from the agent's transcript are not for the terminal and are skipped.
func (s *Server) relayOutput(ctx context.Context, conn *websocket.Conn, st agent.Stream) {
	for {
		sig, err := st.Next()
		if err != nil {
			return
		}
		switch {
		case sig.Exited:
			send(ctx, conn, map[string]any{"type": "exit", "code": sig.ExitCode})
			return
		case len(sig.Output) > 0:
			wctx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := conn.Write(wctx, websocket.MessageBinary, sig.Output)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

// relayInput types the client's frames into the agent until the client leaves.
func (s *Server) relayInput(ctx context.Context, conn *websocket.Conn, st agent.Stream, id domain.WorkerID) {
	for {
		kind, data, err := conn.Read(ctx)
		if err != nil || !s.handleFrame(ctx, st, id, kind, data) {
			return
		}
	}
}

// handleFrame acts on one frame from the client and reports whether the terminal goes on.
func (s *Server) handleFrame(ctx context.Context, st agent.Stream, id domain.WorkerID, kind websocket.MessageType, data []byte) bool {
	switch kind {
	case websocket.MessageBinary:
		if err := st.Write(data); err != nil {
			return false
		}
		if err := s.agents.NoteInput(ctx, id, data); err != nil && !errors.Is(err, context.Canceled) {
			s.cfg.Log.Warn("note terminal input", "worker", id, "err", err)
		}
	case websocket.MessageText:
		var rs resize
		if json.Unmarshal(data, &rs) == nil && rs.Type == "resize" && rs.Rows > 0 && rs.Cols > 0 {
			_ = st.Resize(rs.Rows, rs.Cols)
		}
	}
	return true
}
