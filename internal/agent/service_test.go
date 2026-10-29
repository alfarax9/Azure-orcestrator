package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

type fakeStream struct {
	signals chan Signal
	mu      sync.Mutex
	writes  []string
	once    sync.Once
}

func newStream(signals ...Signal) *fakeStream {
	s := &fakeStream{signals: make(chan Signal, len(signals)+1)}
	for _, sig := range signals {
		s.signals <- sig
	}
	return s
}

func (s *fakeStream) Next() (Signal, error) {
	sig, ok := <-s.signals
	if !ok {
		return Signal{}, io.EOF
	}
	return sig, nil
}

func (s *fakeStream) Write(p []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writes = append(s.writes, string(p))
	return nil
}

func (s *fakeStream) Resize(rows, cols uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writes = append(s.writes, fmt.Sprintf("resize %dx%d", rows, cols))
	return nil
}

func (s *fakeStream) Close() error {
	s.once.Do(func() { close(s.signals) })
	return nil
}

type fakeSessions struct {
	started   []SessionSpec
	stopped   []string
	stream    *fakeStream
	startErr  error
	attachErr error
}

func (f *fakeSessions) Start(_ context.Context, spec SessionSpec) error {
	f.started = append(f.started, spec)
	return f.startErr
}

func (f *fakeSessions) Attach(context.Context, string, string) (Stream, error) {
	if f.attachErr != nil {
		return nil, f.attachErr
	}
	return f.stream, nil
}

func (f *fakeSessions) Stop(_ context.Context, slug, id string) error {
	f.stopped = append(f.stopped, slug+"/"+id)
	return nil
}

type memStore struct {
	mu      sync.Mutex
	worker  domain.Worker
	history []domain.Activity
}

func (m *memStore) Worker(_ context.Context, id domain.WorkerID) (domain.Worker, error) {
	if id != m.worker.ID {
		return domain.Worker{}, domain.ErrNotFound
	}
	return m.worker, nil
}

func (m *memStore) UpdateWorker(_ context.Context, _ domain.WorkerID, change func(*domain.Worker)) (domain.Worker, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	change(&m.worker)
	m.history = append(m.history, m.worker.Activity)
	return m.worker, nil
}

func (m *memStore) WorkspaceOf(context.Context, domain.ProjectID) (domain.Workspace, error) {
	return domain.Workspace{Slug: "acme"}, nil
}

type countDiffs struct{ n int }

func (c *countDiffs) Refresh(context.Context, domain.WorkerID) error { c.n++; return nil }

type fixture struct {
	svc      *Service
	sessions *fakeSessions
	store    *memStore
	diffs    *countDiffs
	clock    time.Time
}

func newFixture(signals ...Signal) *fixture {
	f := &fixture{
		sessions: &fakeSessions{stream: newStream(signals...)},
		store:    &memStore{worker: domain.NewWorker("p1", "Fix header", domain.AgentClaude, "astra/x", true, time.Unix(0, 0))},
		diffs:    &countDiffs{},
		clock:    time.Unix(1000, 0),
	}
	f.store.worker.Worktree = "site/workers/w1"
	f.svc = New(f.sessions, f.store, f.diffs, func() time.Time { return f.clock })
	f.svc.pause = 0
	return f
}

func (f *fixture) id() domain.WorkerID { return f.store.worker.ID }

func event(k Kind) Signal { return Signal{Event: &Event{Kind: k}} }

func TestActivityFollowsTheAgentsEvents(t *testing.T) {
	for _, c := range []struct {
		cur  domain.Activity
		kind Kind
		want domain.Activity
	}{
		{domain.ActivityIdle, KindPrompt, domain.ActivityWorking},
		{domain.ActivityIdle, KindMessage, domain.ActivityWorking},
		{domain.ActivityIdle, KindCommand, domain.ActivityWorking},
		{domain.ActivityIdle, KindTool, domain.ActivityWorking},
		{domain.ActivityIdle, KindFile, domain.ActivityWorking},
		{domain.ActivityWorking, KindTurnEnd, domain.ActivityIdle},
		{domain.ActivityNeedsInput, KindMessage, domain.ActivityWorking},
		{domain.ActivityNeedsInput, "unknown", domain.ActivityNeedsInput},
	} {
		if got := ActivityAfter(c.cur, Event{Kind: c.kind}); got != c.want {
			t.Errorf("%s then %s = %s, want %s", c.cur, c.kind, got, c.want)
		}
	}
}

func TestQuestionsForAPersonAreSpottedThroughTerminalFormatting(t *testing.T) {
	for name, out := range map[string]string{
		"claude trust":    "\x1b[1mDo you trust the files in this folder?\x1b[0m",
		"split by cursor": "Do\x1b[1Cyou\x1b[1Ctrust\x1b[1Cthe\x1b[1Cfiles\x1b[1Cin\x1b[1Cthis\x1b[1Cfolder?",
		"codex trust":     "Do you trust the contents of this directory?",
		"claude approval": "Do you want to proceed?\n❯ 1. Yes",
		"codex approval":  "Would you like to run the following command?",
	} {
		var d needsDetector
		if !d.Feed([]byte(out)) {
			t.Errorf("%s: not spotted", name)
		}
	}
	var d needsDetector
	if d.Feed([]byte("Fixing the header\nrunning go test ./...\n")) {
		t.Error("ordinary output was taken for a question")
	}
	// The question stays on screen across chunks, then scrolls away and can be reset.
	d.Feed([]byte("Do you want"))
	if !d.Feed([]byte(" to proceed?")) {
		t.Error("a question split across chunks was missed")
	}
	d.Reset()
	if d.Feed([]byte("working")) {
		t.Error("a reset detector still sees the old question")
	}
	d.Feed([]byte("Do you want to proceed?"))
	if d.Feed([]byte(strings.Repeat("x", tailLimit+10))) {
		t.Error("an old question did not scroll away")
	}
}

func TestWatchKeepsActivityTrueFromTheSession(t *testing.T) {
	f := newFixture(
		Signal{Output: []byte("Do you trust the files in this folder?")},
		event(KindPrompt), event(KindFile), event(KindTurnEnd),
		Signal{Exited: true, ExitCode: 7},
	)
	f.store.worker.Activity = domain.ActivityIdle
	code, err := f.svc.Watch(t.Context(), f.id())
	if err != nil || code != 7 {
		t.Fatalf("Watch = %d, %v", code, err)
	}
	want := []domain.Activity{domain.ActivityNeedsInput, domain.ActivityWorking, domain.ActivityIdle, domain.ActivityIdle}
	if fmt.Sprint(f.store.history) != fmt.Sprint(want) {
		t.Errorf("activity history = %v, want %v", f.store.history, want)
	}
	if f.diffs.n != 2 {
		t.Errorf("diffs counted %d times, want after the file change and the turn end", f.diffs.n)
	}
}

func TestWatchCountsDiffsAtMostOnceInAWhileForFileEvents(t *testing.T) {
	f := newFixture(event(KindFile), event(KindFile), event(KindFile), Signal{Exited: true})
	if _, err := f.svc.Watch(t.Context(), f.id()); err != nil {
		t.Fatal(err)
	}
	if f.diffs.n != 1 {
		t.Errorf("diffs counted %d times for three quick file events", f.diffs.n)
	}
}

func TestWatchEndsQuietlyWhenItsContextEnds(t *testing.T) {
	f := newFixture()
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { _, err := f.svc.Watch(ctx, f.id()); done <- err }()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("err = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Watch did not return after its context ended")
	}
}

func TestWatchReportsAConnectionThatBreaks(t *testing.T) {
	f := newFixture()
	_ = f.sessions.stream.Close()
	if _, err := f.svc.Watch(t.Context(), f.id()); err == nil || !errors.Is(err, io.EOF) {
		t.Errorf("err = %v, want the broken stream", err)
	}
	if _, err := f.svc.Watch(t.Context(), "nobody"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown worker: err = %v", err)
	}
}

func TestSendPastesTheMessageThenPressesEnterSeparately(t *testing.T) {
	f := newFixture()
	if err := f.svc.Send(t.Context(), f.id(), "Fix the header\nand the footer"); err != nil {
		t.Fatal(err)
	}
	want := []string{"\x1b[200~Fix the header\nand the footer\x1b[201~", "\r"}
	if fmt.Sprint(f.sessions.stream.writes) != fmt.Sprint(want) {
		t.Errorf("writes = %q, want %q", f.sessions.stream.writes, want)
	}
	if f.store.worker.Activity != domain.ActivityWorking {
		t.Errorf("activity = %s", f.store.worker.Activity)
	}
}

func TestSendRefusesTextThatCouldEscapeThePaste(t *testing.T) {
	for _, text := range []string{"", "  \n", "ok\x1b[201~rm -rf ~", strings.Repeat("x", maxMessage+1)} {
		f := newFixture()
		if err := f.svc.Send(t.Context(), f.id(), text); !errors.Is(err, ErrInvalid) {
			t.Errorf("Send(%.20q) err = %v, want ErrInvalid", text, err)
		}
		if len(f.sessions.stream.writes) != 0 {
			t.Errorf("%.20q: something was typed", text)
		}
	}
}

func TestAnsweringAQuestionClearsNeedsYou(t *testing.T) {
	f := newFixture()
	f.store.worker.Activity = domain.ActivityNeedsInput
	f.svc.screenOf(f.id()).Feed([]byte("Do you trust the files in this folder?"))
	if err := f.svc.Press(t.Context(), f.id(), "\r"); err != nil {
		t.Fatal(err)
	}
	if f.store.worker.Activity != domain.ActivityIdle || f.svc.screenOf(f.id()).Feed([]byte("ok")) {
		t.Errorf("activity = %s; the old question should be gone", f.store.worker.Activity)
	}
	for _, keys := range []string{"", strings.Repeat("a", maxKeys+1)} {
		if err := f.svc.Press(t.Context(), f.id(), keys); !errors.Is(err, ErrInvalid) {
			t.Errorf("Press(%d bytes) err = %v", len(keys), err)
		}
	}
}

func TestInterruptSendsEscape(t *testing.T) {
	f := newFixture()
	f.store.worker.Activity = domain.ActivityWorking
	if err := f.svc.Interrupt(t.Context(), f.id()); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(f.sessions.stream.writes) != "[\x1b]" || f.store.worker.Activity != domain.ActivityIdle {
		t.Errorf("writes %q, activity %s", f.sessions.stream.writes, f.store.worker.Activity)
	}
}

func TestStartRunsTheAgentInTheWorktreeAndStopEndsIt(t *testing.T) {
	f := newFixture()
	if err := f.svc.Start(t.Context(), f.id()); err != nil {
		t.Fatal(err)
	}
	want := SessionSpec{Slug: "acme", ID: string(f.id()), Agent: domain.AgentClaude, Dir: "site/workers/w1"}
	if len(f.sessions.started) != 1 || f.sessions.started[0] != want {
		t.Errorf("started %+v, want %+v", f.sessions.started, want)
	}
	if len(f.sessions.stream.writes) != 0 {
		t.Error("Start must not type anything: the first question belongs to the person")
	}
	if err := f.svc.Stop(t.Context(), f.id()); err != nil || fmt.Sprint(f.sessions.stopped) != "[acme/"+string(f.id())+"]" {
		t.Errorf("Stop: %v, %v", err, f.sessions.stopped)
	}
}

func TestStartReportsFailuresAndWorkersWithoutAWorktree(t *testing.T) {
	f := newFixture()
	f.sessions.startErr = errors.New("unit exists")
	if err := f.svc.Start(t.Context(), f.id()); err == nil || !strings.Contains(err.Error(), "unit exists") {
		t.Errorf("err = %v", err)
	}
	f = newFixture()
	f.store.worker.Worktree = ""
	if err := f.svc.Start(t.Context(), f.id()); err == nil {
		t.Error("a worker without a worktree must not start")
	}
	if err := f.svc.Start(t.Context(), "nobody"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown worker: err = %v", err)
	}
}

func TestNoteInputClearsAQuestionOnlyWhenAPersonPressedEnter(t *testing.T) {
	f := newFixture()
	f.store.worker.Activity = domain.ActivityNeedsInput
	if err := f.svc.NoteInput(t.Context(), f.id(), []byte("y")); err != nil || f.store.worker.Activity != domain.ActivityNeedsInput {
		t.Fatalf("a key without Enter: %v, activity %s", err, f.store.worker.Activity)
	}
	if err := f.svc.NoteInput(t.Context(), f.id(), []byte("\r")); err != nil || f.store.worker.Activity != domain.ActivityIdle {
		t.Fatalf("Enter: %v, activity %s", err, f.store.worker.Activity)
	}
	f.store.worker.Activity = domain.ActivityWorking
	before := len(f.store.history)
	if err := f.svc.NoteInput(t.Context(), f.id(), []byte("more\r")); err != nil || len(f.store.history) != before {
		t.Error("typing while the agent works must not change the facts")
	}
}

func TestAWorkerWithoutASessionIsReportedAsSuch(t *testing.T) {
	for _, cause := range []error{fs.ErrNotExist, syscall.ECONNREFUSED} {
		f := newFixture()
		f.sessions.attachErr = fmt.Errorf("attach: %w", cause)
		if _, err := f.svc.Open(t.Context(), f.id()); !errors.Is(err, ErrNoSession) {
			t.Errorf("%v: err = %v, want ErrNoSession", cause, err)
		}
	}
	f := newFixture()
	f.sessions.attachErr = errors.New("permission denied")
	if _, err := f.svc.Open(t.Context(), f.id()); errors.Is(err, ErrNoSession) || err == nil {
		t.Errorf("another failure must stay itself: %v", err)
	}
}

type roster struct{ workers []domain.Worker }

func (r roster) Workers(context.Context, string) ([]domain.Worker, error) { return r.workers, nil }

func TestSupervisorWatchesLiveWorkersAndSettlesDeadOnes(t *testing.T) {
	live := newFixture(Signal{Exited: true, ExitCode: 0})
	live.store.worker.Activity = domain.ActivityWorking
	log := slog.New(slog.DiscardHandler)
	sup := NewSupervisor(live.svc, roster{workers: []domain.Worker{live.store.worker, {ID: "no-tree"}}}, log)
	if err := sup.Resume(t.Context()); err != nil {
		t.Fatal(err)
	}
	sup.Track(t.Context(), live.id()) // already tracked or finished: never twice at once
	sup.Wait()
	if live.store.worker.Activity != domain.ActivityIdle {
		t.Errorf("a worker whose agent exited is %s, want idle", live.store.worker.Activity)
	}

	dead := newFixture()
	dead.sessions.attachErr = fs.ErrNotExist
	dead.store.worker.Activity = domain.ActivityWorking
	supDead := NewSupervisor(dead.svc, roster{}, log)
	supDead.Track(t.Context(), dead.id())
	supDead.Wait()
	if dead.store.worker.Activity != domain.ActivityIdle {
		t.Errorf("a worker without a session is %s, want idle", dead.store.worker.Activity)
	}
}

func TestSupervisorKeepsOneWatchPerWorker(t *testing.T) {
	f := newFixture()
	sup := NewSupervisor(f.svc, roster{}, slog.New(slog.DiscardHandler))
	ctx, cancel := context.WithCancel(t.Context())
	sup.Track(ctx, f.id())
	sup.Track(ctx, f.id())
	sup.mu.Lock()
	n := len(sup.running)
	sup.mu.Unlock()
	if n != 1 {
		t.Errorf("%d watches running for one worker", n)
	}
	cancel()
	sup.Wait()
}
