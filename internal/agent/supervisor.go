package agent

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"slices"
	"sync"

	"github.com/alfarax9/astra/internal/domain"
)

// Roster lists workers; an empty slug lists every worker on the server.
type Roster interface {
	Workers(ctx context.Context, slug string) ([]domain.Worker, error)
}

// Supervisor keeps one Watch running for every worker whose session is alive, so the
// worker's facts stay true without anyone looking at it. It starts them again after the
// daemon restarts (F3-04, N-07).
type Supervisor struct {
	svc     *Service
	roster  Roster
	log     *slog.Logger
	wg      sync.WaitGroup
	drainMu sync.Mutex // one drain at a time, so a queued worker is never started twice
	mu      sync.Mutex // guards running
	running map[domain.WorkerID]struct{}
}

// NewSupervisor returns a Supervisor for svc.
func NewSupervisor(svc *Service, r Roster, log *slog.Logger) *Supervisor {
	return &Supervisor{svc: svc, roster: r, log: log, running: map[domain.WorkerID]struct{}{}}
}

// Track starts watching worker id until its agent exits or ctx ends. Tracking a worker that is
// already watched does nothing.
func (s *Supervisor) Track(ctx context.Context, id domain.WorkerID) {
	s.mu.Lock()
	if _, ok := s.running[id]; ok {
		s.mu.Unlock()
		return
	}
	s.running[id] = struct{}{}
	s.mu.Unlock()
	s.wg.Go(func() {
		defer func() {
			s.mu.Lock()
			delete(s.running, id)
			s.mu.Unlock()
		}()
		s.watch(ctx, id)
	})
}

func (s *Supervisor) watch(ctx context.Context, id domain.WorkerID) {
	code, err := s.svc.Watch(ctx, id)
	switch {
	case errors.Is(err, ErrNoSession):
		// Nothing runs: whatever the facts said the agent was doing, it is not doing it now.
		if err := s.svc.settle(context.WithoutCancel(ctx), id); err != nil {
			s.log.Warn("settle worker without a session", "worker", id, "err", err)
		}
	case err != nil:
		s.log.Warn("watching worker stopped", "worker", id, "err", err)
	case ctx.Err() == nil:
		s.log.Info("agent exited", "worker", id, "exit_code", code)
	}
	if !errors.Is(err, ErrNoSession) && ctx.Err() == nil {
		s.Drain(ctx) // a session ended, so a slot may be free for a queued worker
	}
}

// Drain starts queued workers, oldest first, for as long as there are free slots (F10-05).
func (s *Supervisor) Drain(ctx context.Context) {
	s.drainMu.Lock()
	defer s.drainMu.Unlock()
	workers, err := s.roster.Workers(ctx, "")
	if err != nil {
		s.log.Warn("list queued workers", "err", err)
		return
	}
	var queued []domain.Worker
	for _, w := range workers {
		if w.Queued {
			queued = append(queued, w)
		}
	}
	slices.SortFunc(queued, func(a, b domain.Worker) int {
		return cmp.Or(a.CreatedAt.Compare(b.CreatedAt.Time), cmp.Compare(a.ID, b.ID))
	})
	for _, w := range queued {
		err := s.svc.Start(ctx, w.ID)
		if errors.Is(err, ErrQueued) {
			return // no slot is free; the next session to end tries again
		}
		if err != nil {
			s.log.Warn("start queued worker", "worker", w.ID, "err", err)
			continue
		}
		s.Track(ctx, w.ID)
	}
}

// Resume watches every worker that has a worktree. Those whose session ended while the daemon
// was away are settled as idle. Sessions that are still alive count against the limit before
// any queued worker is started.
func (s *Supervisor) Resume(ctx context.Context) error {
	workers, err := s.roster.Workers(ctx, "")
	if err != nil {
		return err
	}
	for _, w := range workers {
		if w.Worktree != "" && !w.Queued {
			s.svc.Adopt(ctx, w.ID)
		}
	}
	for _, w := range workers {
		if w.Worktree != "" && !w.Queued {
			s.Track(ctx, w.ID)
		}
	}
	s.Drain(ctx)
	return nil
}

// Wait returns once every watch has ended, which needs the context given to Track to end.
func (s *Supervisor) Wait() { s.wg.Wait() }
