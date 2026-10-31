package agent

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

// poolStore keeps several workers, so the limit has something to count.
type poolStore struct {
	mu      sync.Mutex
	workers map[domain.WorkerID]domain.Worker
}

func newPool(ids ...domain.WorkerID) *poolStore {
	p := &poolStore{workers: map[domain.WorkerID]domain.Worker{}}
	for i, id := range ids {
		p.workers[id] = domain.Worker{ID: id, Worktree: "w/" + string(id), Agent: domain.AgentClaude, CreatedAt: domain.Timestamp{Time: time.UnixMilli(int64(1000 + i))}}
	}
	return p
}

func (p *poolStore) Worker(_ context.Context, id domain.WorkerID) (domain.Worker, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	w, ok := p.workers[id]
	if !ok {
		return w, domain.ErrNotFound
	}
	return w, nil
}

func (p *poolStore) UpdateWorker(_ context.Context, id domain.WorkerID, change func(*domain.Worker)) (domain.Worker, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	w, ok := p.workers[id]
	if !ok {
		return w, domain.ErrNotFound
	}
	change(&w)
	p.workers[id] = w
	return w, nil
}

func (p *poolStore) WorkspaceOf(context.Context, domain.ProjectID) (domain.Workspace, error) {
	return domain.Workspace{Slug: "acme"}, nil
}

func (p *poolStore) RecordUsage(context.Context, domain.WorkerID, time.Time, domain.Usage) error {
	return nil
}

func (p *poolStore) Workers(context.Context, string) ([]domain.Worker, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var out []domain.Worker
	for _, w := range p.workers {
		out = append(out, w)
	}
	return out, nil
}

// pool records the sessions that were started and lets a test end them.
type poolSessions struct {
	mu      sync.Mutex
	started []string
	streams map[string]*fakeStream
}

func (s *poolSessions) Start(_ context.Context, spec SessionSpec) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = append(s.started, spec.ID)
	s.streams[spec.ID] = newStream()
	return nil
}

func (s *poolSessions) Attach(_ context.Context, _, id string) (Stream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st := s.streams[id]; st != nil {
		return &attachment{session: st, closed: make(chan struct{})}, nil
	}
	return nil, fs.ErrNotExist // no such session
}

// attachment is one connection to a session: closing it ends that connection, not the session.
type attachment struct {
	session *fakeStream
	once    sync.Once
	closed  chan struct{}
}

func (a *attachment) Next() (Signal, error) {
	select {
	case sig, ok := <-a.session.signals:
		if !ok {
			return Signal{}, io.EOF
		}
		return sig, nil
	case <-a.closed:
		return Signal{}, io.EOF
	}
}

func (a *attachment) Write(p []byte) error           { return a.session.Write(p) }
func (a *attachment) Resize(rows, cols uint16) error { return a.session.Resize(rows, cols) }
func (a *attachment) Close() error                   { a.once.Do(func() { close(a.closed) }); return nil }

func (s *poolSessions) Stop(context.Context, string, string) error { return nil }

func (s *poolSessions) startedIDs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.started...)
}

func (s *poolSessions) end(id string) {
	s.mu.Lock()
	st := s.streams[id]
	s.mu.Unlock()
	st.signals <- Signal{Exited: true}
}

func limitFixture(limit int, ids ...domain.WorkerID) (*Service, *poolStore, *poolSessions) {
	store := newPool(ids...)
	sessions := &poolSessions{streams: map[string]*fakeStream{}}
	svc := New(sessions, store, nil, time.Now)
	svc.SetLimit(limit)
	return svc, store, sessions
}

func TestWorkersBeyondTheLimitAreQueuedWithoutStartingTheirAgent(t *testing.T) {
	ctx := t.Context()
	svc, store, sessions := limitFixture(2, "a", "b", "c")
	for _, id := range []domain.WorkerID{"a", "b"} {
		if err := svc.Start(ctx, id); err != nil {
			t.Fatalf("start %s: %v", id, err)
		}
	}
	err := svc.Start(ctx, "c")
	if !errors.Is(err, ErrQueued) {
		t.Fatalf("third start = %v, want ErrQueued", err)
	}
	c, _ := store.Worker(ctx, "c")
	if !c.Queued || c.Status() != domain.StatusQueued {
		t.Errorf("queued worker: %+v, status %s", c, c.Status())
	}
	if got := sessions.startedIDs(); len(got) != 2 {
		t.Errorf("sessions started %v: the queued worker must not have one", got)
	}
}

func TestStartingARunningWorkerAgainKeepsItsSlot(t *testing.T) {
	ctx := t.Context()
	svc, _, _ := limitFixture(1, "a", "b")
	if err := svc.Start(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Start(ctx, "a"); err != nil {
		t.Errorf("a worker that holds the slot is not queued behind itself: %v", err)
	}
	if err := svc.Start(ctx, "b"); !errors.Is(err, ErrQueued) {
		t.Errorf("second worker = %v, want ErrQueued", err)
	}
}

func TestZeroMeansNoLimit(t *testing.T) {
	ctx := t.Context()
	svc, _, _ := limitFixture(0, "a", "b", "c", "d", "e")
	for _, id := range []domain.WorkerID{"a", "b", "c", "d", "e"} {
		if err := svc.Start(ctx, id); err != nil {
			t.Fatalf("start %s: %v", id, err)
		}
	}
}

func TestAFailedStartFreesItsSlot(t *testing.T) {
	ctx := t.Context()
	svc, _, _ := limitFixture(1, "a", "b")
	svc.sessions = failingSessions{}
	if err := svc.Start(ctx, "a"); err == nil || errors.Is(err, ErrQueued) {
		t.Fatalf("start = %v, want a real failure", err)
	}
	svc.sessions = &poolSessions{streams: map[string]*fakeStream{}}
	if err := svc.Start(ctx, "b"); err != nil {
		t.Errorf("the slot of the failed start is free again: %v", err)
	}
}

type failingSessions struct{}

func (failingSessions) Start(context.Context, SessionSpec) error { return errors.New("unit exists") }
func (failingSessions) Attach(context.Context, string, string) (Stream, error) {
	return nil, errors.New("none")
}
func (failingSessions) Stop(context.Context, string, string) error { return nil }

func TestStoppingAQueuedWorkerTakesItOutOfTheQueue(t *testing.T) {
	ctx := t.Context()
	svc, store, _ := limitFixture(1, "a", "b")
	_ = svc.Start(ctx, "a")
	_ = svc.Start(ctx, "b")
	if err := svc.Stop(ctx, "b"); err != nil {
		t.Fatal(err)
	}
	b, _ := store.Worker(ctx, "b")
	if b.Queued || b.Activity != domain.ActivityIdle {
		t.Errorf("stopped queued worker: %+v", b)
	}
}

func TestDrainStartsQueuedWorkersOldestFirstAsSlotsFree(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		svc, store, sessions := limitFixture(1, "a", "b", "c")
		sup := NewSupervisor(svc, store, slog.New(slog.DiscardHandler))
		for _, id := range []domain.WorkerID{"a", "b", "c"} {
			_ = svc.Start(ctx, id) // a runs; b and c queue
		}
		sup.Track(ctx, "a")
		synctest.Wait()
		sessions.end("a")
		synctest.Wait()
		if got := sessions.startedIDs(); len(got) != 2 || got[1] != "b" {
			t.Fatalf("started %v: the oldest queued worker (b) must go first", got)
		}
		b, _ := store.Worker(ctx, "b")
		c, _ := store.Worker(ctx, "c")
		if b.Queued || !c.Queued {
			t.Errorf("after one slot freed: b queued=%v, c queued=%v", b.Queued, c.Queued)
		}
		sessions.end("b")
		synctest.Wait()
		if got := sessions.startedIDs(); len(got) != 3 || got[2] != "c" {
			t.Errorf("started %v: c goes when b ends", got)
		}
		cancel()
		sup.Wait()
	})
}

func TestResumeCountsSessionsThatSurvivedTheRestartBeforeStartingQueuedWorkers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		svc, store, sessions := limitFixture(1, "a", "b")
		sessions.streams["a"] = newStream() // a's session is still alive after the daemon restarted
		store.workers["b"] = func() domain.Worker { w := store.workers["b"]; w.Queued = true; return w }()
		sup := NewSupervisor(svc, store, slog.New(slog.DiscardHandler))
		if err := sup.Resume(ctx); err != nil {
			t.Fatal(err)
		}
		synctest.Wait()
		if got := sessions.startedIDs(); len(got) != 0 {
			t.Errorf("started %v: the one allowed slot is taken by the session that survived", got)
		}
		sessions.end("a")
		synctest.Wait()
		if got := sessions.startedIDs(); len(got) != 1 || got[0] != "b" {
			t.Errorf("started %v, want b once a ends", got)
		}
		cancel()
		sup.Wait()
	})
}
