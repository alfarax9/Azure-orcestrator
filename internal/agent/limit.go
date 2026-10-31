package agent

import (
	"context"
	"errors"

	"github.com/alfarax9/astra/internal/domain"
)

// ErrQueued means the server already runs as many workers as it may (F10-05): the worker was
// marked as waiting, and the Supervisor starts it when a slot frees up.
var ErrQueued = errors.New("waiting for a free slot")

// SetLimit sets how many workers may run at once on this server; zero or less means no limit.
// A worker holds a slot from the moment its session starts until its session ends.
func (s *Service) SetLimit(n int) {
	s.slots.Lock()
	defer s.slots.Unlock()
	s.limit = n
}

// reserve gives the worker a slot, or reports that none is free. A worker that already holds
// one keeps it.
func (s *Service) reserve(id domain.WorkerID) bool {
	s.slots.Lock()
	defer s.slots.Unlock()
	if _, ok := s.active[id]; ok {
		return true
	}
	if s.limit > 0 && len(s.active) >= s.limit {
		return false
	}
	s.active[id] = struct{}{}
	return true
}

// hold records a session that is already running, even beyond the limit: it exists, so it
// counts, and nothing new starts until enough of them end.
func (s *Service) hold(id domain.WorkerID) {
	s.slots.Lock()
	defer s.slots.Unlock()
	s.active[id] = struct{}{}
}

func (s *Service) release(id domain.WorkerID) {
	s.slots.Lock()
	defer s.slots.Unlock()
	delete(s.active, id)
}

// Adopt counts the worker's session against the limit when the daemon starts and finds the
// session still alive. It reports whether there was one.
func (s *Service) Adopt(ctx context.Context, id domain.WorkerID) bool {
	st, err := s.Open(ctx, id)
	if err != nil {
		return false
	}
	_ = st.Close()
	s.hold(id)
	return true
}

// queue marks the worker as waiting for a slot.
func (s *Service) queue(ctx context.Context, id domain.WorkerID) error {
	_, err := s.store.UpdateWorker(ctx, id, func(w *domain.Worker) { w.Queued = true })
	return err
}
