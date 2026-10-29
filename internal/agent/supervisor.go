package agent

import (
	"context"
	"errors"
	"log/slog"
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
}

// Resume watches every worker that has a worktree. Those whose session ended while the daemon
// was away are settled as idle.
func (s *Supervisor) Resume(ctx context.Context) error {
	workers, err := s.roster.Workers(ctx, "")
	if err != nil {
		return err
	}
	for _, w := range workers {
		if w.Worktree != "" {
			s.Track(ctx, w.ID)
		}
	}
	return nil
}

// Wait returns once every watch has ended, which needs the context given to Track to end.
func (s *Supervisor) Wait() { s.wg.Wait() }
