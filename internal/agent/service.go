package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

// Signal is one thing a running session reports to the daemon.
type Signal struct {
	Output   []byte // bytes the agent wrote to its terminal
	Event    *Event // a structured event from the agent's transcript
	Exited   bool
	ExitCode int
}

// Stream is the daemon's end of one session.
type Stream interface {
	Next() (Signal, error)
	Write(p []byte) error
	Resize(rows, cols uint16) error
	Close() error
}

// SessionSpec says which session to start: the agent of one worker, in its worktree.
type SessionSpec struct {
	Slug  string // the workspace, and so the Linux user the agent runs as
	ID    string // the session, which is the worker's id
	Agent domain.AgentKind
	Dir   string // the worker's worktree, relative to the workspace's work folder
}

// Sessions starts, reaches and stops agent sessions. The implementation decides how: as the
// workspace's agent user through astra-priv on a server, or as the developer on a laptop.
type Sessions interface {
	Start(ctx context.Context, spec SessionSpec) error
	Attach(ctx context.Context, slug, id string) (Stream, error)
	Stop(ctx context.Context, slug, id string) error
}

// Store keeps the facts the service reads and changes.
type Store interface {
	Worker(ctx context.Context, id domain.WorkerID) (domain.Worker, error)
	UpdateWorker(ctx context.Context, id domain.WorkerID, change func(*domain.Worker)) (domain.Worker, error)
	WorkspaceOf(ctx context.Context, project domain.ProjectID) (domain.Workspace, error)
	RecordUsage(ctx context.Context, id domain.WorkerID, at time.Time, u domain.Usage) error
}

// Diffs brings a worker's added and removed line counts up to date.
type Diffs interface {
	Refresh(ctx context.Context, id domain.WorkerID) error
}

var (
	// ErrInvalid means the text or keys to send were refused before anything was written.
	ErrInvalid = errors.New("invalid")
	// ErrNoSession means the worker has no running session: it was never started, or ended.
	ErrNoSession = errors.New("no running session")
)

const (
	// pastePause separates a pasted message from the Enter that submits it: Codex treats an
	// Enter that arrives with the paste as part of the text (docs/adr/0003).
	pastePause = 200 * time.Millisecond
	// diffEvery is how often file events may trigger a diff count.
	diffEvery = 2 * time.Second

	maxMessage = 20000
	maxKeys    = 64
	escape     = "\x1b"
)

// Service runs the agents of workers.
type Service struct {
	sessions Sessions
	store    Store
	diffs    Diffs // optional
	pause    time.Duration
	now      func() time.Time

	mu     sync.Mutex // guards screens
	screen map[domain.WorkerID]*needsDetector

	slots  sync.Mutex // guards limit and active
	limit  int
	active map[domain.WorkerID]struct{} // workers whose session runs
}

// New returns a Service. diffs may be nil.
func New(s Sessions, st Store, d Diffs, now func() time.Time) *Service {
	return &Service{sessions: s, store: st, diffs: d, pause: pastePause, now: now, screen: map[domain.WorkerID]*needsDetector{}, active: map[domain.WorkerID]struct{}{}}
}

// target finds the workspace and session of a worker.
func (s *Service) target(ctx context.Context, id domain.WorkerID) (domain.Worker, string, error) {
	w, err := s.store.Worker(ctx, id)
	if err != nil {
		return w, "", err
	}
	ws, err := s.store.WorkspaceOf(ctx, w.ProjectID)
	return w, ws.Slug, err
}

// Start runs the worker's agent in its worktree. The task is not typed for the person: the
// agent first asks whether to trust the folder, and Enter would answer yes. The worker waits
// as "Needs you" until the person has answered, then gets its task through Send. When the
// server already runs as many workers as it may, the worker is queued and ErrQueued returned.
func (s *Service) Start(ctx context.Context, id domain.WorkerID) error {
	w, slug, err := s.target(ctx, id)
	if err != nil {
		return err
	}
	if w.Worktree == "" {
		return fmt.Errorf("worker %s has no worktree", id)
	}
	if !s.reserve(id) {
		if err := s.queue(ctx, id); err != nil {
			return err
		}
		return fmt.Errorf("worker %s: %w", id, ErrQueued)
	}
	if err := s.sessions.Start(ctx, SessionSpec{Slug: slug, ID: string(id), Agent: w.Agent, Dir: w.Worktree}); err != nil {
		s.release(id)
		return fmt.Errorf("start the agent of worker %s: %w", id, err)
	}
	_, err = s.store.UpdateWorker(ctx, id, func(w *domain.Worker) { w.Queued, w.Activity = false, domain.ActivityIdle })
	return err
}

// Watch follows the worker's session and keeps its facts true until the agent exits (its
// exit code is returned) or ctx ends. Events replayed from the host after a daemon restart
// walk the activity to where it should be.
func (s *Service) Watch(ctx context.Context, id domain.WorkerID) (int, error) {
	w, err := s.store.Worker(ctx, id)
	if err != nil {
		return 0, err
	}
	st, err := s.Open(ctx, id)
	if err != nil {
		return 0, err
	}
	s.hold(id)
	defer s.release(id)
	var closers sync.WaitGroup
	done := make(chan struct{})
	closers.Go(func() {
		select {
		case <-ctx.Done():
		case <-done:
		}
		_ = st.Close() // unblocks Next
	})
	defer func() { close(done); closers.Wait() }()

	wt := watcher{s: s, id: id, cur: w.Activity}
	for {
		sig, err := st.Next()
		if err != nil {
			if ctx.Err() != nil {
				return 0, nil
			}
			return 0, fmt.Errorf("read session of worker %s: %w", id, err)
		}
		if exited, code, err := wt.apply(ctx, sig); exited || err != nil {
			return code, err
		}
	}
}

// watcher is the state of one Watch: the activity last stored and when the diff was last counted.
type watcher struct {
	s        *Service
	id       domain.WorkerID
	cur      domain.Activity
	lastDiff time.Time
}

// apply takes in one signal and stores the activity it leads to. It reports an exit.
func (wt *watcher) apply(ctx context.Context, sig Signal) (exited bool, code int, err error) {
	s := wt.s
	next := wt.cur
	switch {
	case sig.Exited:
		return true, sig.ExitCode, s.setActivity(context.WithoutCancel(ctx), wt.id, domain.ActivityIdle)
	case sig.Event != nil && sig.Event.Kind == KindUsage:
		// Tokens are bookkeeping: the agent has not moved on, so the activity and screen stay.
		return false, 0, s.recordUsage(ctx, wt.id, *sig.Event)
	case sig.Event != nil:
		s.screenOf(wt.id).Reset() // the agent moved on, so any question was answered
		next = ActivityAfter(wt.cur, *sig.Event)
		wt.lastDiff = s.refreshDiff(ctx, wt.id, *sig.Event, wt.lastDiff)
	case len(sig.Output) > 0 && s.screenOf(wt.id).Feed(sig.Output):
		next = domain.ActivityNeedsInput
	}
	if next == wt.cur {
		return false, 0, nil
	}
	wt.cur = next
	return false, 0, s.setActivity(ctx, wt.id, next)
}

// recordUsage stores the tokens a usage event reports. After a daemon restart the host replays
// its events, so the same message arrives again and replaces its own row.
func (s *Service) recordUsage(ctx context.Context, id domain.WorkerID, e Event) error {
	if e.Usage == nil {
		return nil
	}
	at := e.At
	if at.IsZero() {
		at = s.now()
	}
	return s.store.RecordUsage(ctx, id, at, *e.Usage)
}

// Open attaches to the worker's running session, for the terminal or for watching it. A
// worker without a session gives ErrNoSession.
func (s *Service) Open(ctx context.Context, id domain.WorkerID) (Stream, error) {
	_, slug, err := s.target(ctx, id)
	if err != nil {
		return nil, err
	}
	st, err := s.sessions.Attach(ctx, slug, string(id))
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ECONNREFUSED) {
		return nil, fmt.Errorf("worker %s: %w", id, ErrNoSession)
	}
	return st, err
}

// NoteInput tells the service that a person typed into the worker's terminal. When they pressed
// Enter, a question that was waiting for them is taken as answered; if it is still on the
// screen, the next output brings "Needs you" back.
func (s *Service) NoteInput(ctx context.Context, id domain.WorkerID, input []byte) error {
	if !bytes.ContainsAny(input, "\r\n") {
		return nil
	}
	s.screenOf(id).Reset()
	w, err := s.store.Worker(ctx, id)
	if err != nil || w.Activity != domain.ActivityNeedsInput {
		return err
	}
	return s.setActivity(ctx, id, domain.ActivityIdle)
}

// refreshDiff recounts a worker's changed lines after files change and when a turn ends,
// at most once per diffEvery. It returns the time of the last count.
func (s *Service) refreshDiff(ctx context.Context, id domain.WorkerID, e Event, last time.Time) time.Time {
	if s.diffs == nil || (e.Kind != KindTurnEnd && e.Kind != KindFile) {
		return last
	}
	if now := s.now(); e.Kind == KindTurnEnd || now.Sub(last) >= diffEvery {
		_ = s.diffs.Refresh(ctx, id) // a failed count only leaves the old numbers; the next event retries
		return now
	}
	return last
}

func (s *Service) screenOf(id domain.WorkerID) *needsDetector {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.screen[id] == nil {
		s.screen[id] = &needsDetector{}
	}
	return s.screen[id]
}

// settle marks a worker idle when it was thought to be working or waiting but has no session.
func (s *Service) settle(ctx context.Context, id domain.WorkerID) error {
	w, err := s.store.Worker(ctx, id)
	if err != nil || w.Activity == domain.ActivityIdle {
		return err
	}
	return s.setActivity(ctx, id, domain.ActivityIdle)
}

func (s *Service) setActivity(ctx context.Context, id domain.WorkerID, a domain.Activity) error {
	_, err := s.store.UpdateWorker(ctx, id, func(w *domain.Worker) { w.Activity = a })
	return err
}

// Send gives the agent a message or task: pasted as one block so line breaks stay in the
// text, then submitted with a separate Enter (F3-07).
func (s *Service) Send(ctx context.Context, id domain.WorkerID, text string) error {
	if strings.TrimSpace(text) == "" || len(text) > maxMessage || strings.Contains(text, escape) {
		return fmt.Errorf("%w: a message is 1 to %d characters without escape sequences, which could end the paste early", ErrInvalid, maxMessage)
	}
	return s.typeInto(ctx, id, domain.ActivityWorking, func(st Stream) error {
		if err := st.Write([]byte(escape + "[200~" + text + escape + "[201~")); err != nil {
			return err
		}
		select {
		case <-time.After(s.pause):
		case <-ctx.Done():
			return ctx.Err()
		}
		return st.Write([]byte("\r"))
	})
}

// Press types raw keys, such as Enter to accept the folder-trust dialog the person chose to
// accept (docs/m1-plan.md).
func (s *Service) Press(ctx context.Context, id domain.WorkerID, keys string) error {
	if keys == "" || len(keys) > maxKeys {
		return fmt.Errorf("%w: keys are 1 to %d bytes", ErrInvalid, maxKeys)
	}
	return s.typeInto(ctx, id, domain.ActivityIdle, func(st Stream) error { return st.Write([]byte(keys)) })
}

// Interrupt stops what the agent is doing in its current turn (Escape, F3-07).
func (s *Service) Interrupt(ctx context.Context, id domain.WorkerID) error {
	return s.typeInto(ctx, id, domain.ActivityIdle, func(st Stream) error { return st.Write([]byte(escape)) })
}

// typeInto writes to the worker's session through a short attachment. Whatever was on the
// screen is answered, so the worker is no longer waiting: it is after (working for a message,
// idle for keys and an interrupt) until its events say otherwise.
func (s *Service) typeInto(ctx context.Context, id domain.WorkerID, after domain.Activity, write func(Stream) error) error {
	st, err := s.Open(ctx, id)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()
	if err := write(st); err != nil {
		return fmt.Errorf("type into the session of worker %s: %w", id, err)
	}
	s.screenOf(id).Reset()
	return s.setActivity(ctx, id, after)
}

// Stop ends the worker's session and everything the agent started in it.
// A queued worker has no session: it only leaves the queue.
func (s *Service) Stop(ctx context.Context, id domain.WorkerID) error {
	w, slug, err := s.target(ctx, id)
	if err != nil {
		return err
	}
	if !w.Queued {
		if err := s.sessions.Stop(ctx, slug, string(id)); err != nil {
			return err
		}
		s.release(id)
	}
	_, err = s.store.UpdateWorker(ctx, id, func(w *domain.Worker) { w.Queued, w.Activity = false, domain.ActivityIdle })
	return err
}
