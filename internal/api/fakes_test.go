package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/alfarax9/astra/internal/agent"
	"github.com/alfarax9/astra/internal/auth"
	"github.com/alfarax9/astra/internal/checklist"
	"github.com/alfarax9/astra/internal/domain"
)

const goodToken = "good-device-token"

// fakes is every collaborator of the server, in memory.
type fakes struct {
	mu         sync.Mutex
	workspaces []domain.Workspace
	projects   []domain.Project
	workers    []domain.Worker
	usage      []domain.UsageDay
	usageSince time.Time // the start of the last usage question
	listeners  []func(domain.Change)
	tracked    []domain.WorkerID
	calls      []string
	stream     *fakeStream
	openErr    error
	failWith   error // returned by the next capability call
	logs       bytes.Buffer
	static     fs.FS
}

func (f *fakes) call(s string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, s)
	err := f.failWith
	f.failWith = nil
	return err
}

func (f *fakes) lastCall() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.calls) == 0 {
		return ""
	}
	return f.calls[len(f.calls)-1]
}

// Auth
func (f *fakes) Pair(_ context.Context, pairing, name string) (domain.Device, domain.Secret, error) {
	if pairing != "pairing-token" {
		return domain.Device{}, domain.Secret{}, auth.ErrUnauthorized
	}
	if name == "" {
		return domain.Device{}, domain.Secret{}, fmt.Errorf("%w: a device name is required", auth.ErrInvalid)
	}
	return domain.Device{ID: "d1", Name: name}, domain.NewSecret(goodToken), nil
}

func (f *fakes) Authenticate(_ context.Context, token string) (domain.Device, error) {
	if token != goodToken {
		return domain.Device{}, auth.ErrUnauthorized
	}
	return domain.Device{ID: "d1", Name: "Laptop"}, nil
}

// Workspaces
func (f *fakes) AddWorkspace(_ context.Context, slug, name string) (domain.Workspace, error) {
	if err := f.call("AddWorkspace " + slug); err != nil {
		return domain.Workspace{}, err
	}
	ws := domain.Workspace{ID: domain.WorkspaceID("ws-" + slug), Slug: slug, Name: name}
	f.mu.Lock()
	f.workspaces = append(f.workspaces, ws)
	f.mu.Unlock()
	f.publish(domain.Change{Workspace: &ws})
	return ws, nil
}

func (f *fakes) AddProject(_ context.Context, slug, name, url string) (domain.Project, error) {
	if err := f.call("AddProject " + slug + " " + name + " " + url); err != nil {
		return domain.Project{}, err
	}
	p := domain.Project{ID: domain.ProjectID("p-" + name), WorkspaceID: domain.WorkspaceID("ws-" + slug), Name: name, RepoURL: url}
	f.mu.Lock()
	f.projects = append(f.projects, p)
	f.mu.Unlock()
	return p, nil
}

func (f *fakes) SetCommands(_ context.Context, id domain.ProjectID, c domain.Commands) error {
	if err := f.call("SetCommands " + string(id) + " " + c.Dev); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.projects {
		if f.projects[i].ID == id {
			f.projects[i].Commands = c
			return nil
		}
	}
	return domain.ErrNotFound
}

func (f *fakes) CreateWorker(_ context.Context, project domain.ProjectID, title string, a domain.AgentKind, review bool) (domain.Worker, error) {
	if err := f.call(fmt.Sprintf("CreateWorker %s %q %s review=%t", project, title, a, review)); err != nil {
		return domain.Worker{}, err
	}
	w := domain.Worker{ID: "w1", ProjectID: project, Title: title, Agent: a, ClientReview: review}
	f.mu.Lock()
	f.workers = append(f.workers, w)
	f.mu.Unlock()
	return w, nil
}

// Agents
func (f *fakes) Start(_ context.Context, id domain.WorkerID) error {
	return f.call("Start " + string(id))
}
func (f *fakes) Stop(_ context.Context, id domain.WorkerID) error {
	return f.call("Stop " + string(id))
}
func (f *fakes) Interrupt(_ context.Context, id domain.WorkerID) error {
	return f.call("Interrupt " + string(id))
}

func (f *fakes) Send(_ context.Context, id domain.WorkerID, text string) error {
	return f.call("Send " + string(id) + " " + text)
}

func (f *fakes) Open(context.Context, domain.WorkerID) (agent.Stream, error) {
	if f.openErr != nil {
		return nil, f.openErr
	}
	return f.stream, nil
}

func (f *fakes) NoteInput(_ context.Context, id domain.WorkerID, input []byte) error {
	return f.call("NoteInput " + string(id) + " " + string(input))
}

// Reads
func (f *fakes) Workspaces(context.Context) ([]domain.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]domain.Workspace(nil), f.workspaces...), nil
}

func (f *fakes) Workspace(_ context.Context, slug string) (domain.Workspace, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, w := range f.workspaces {
		if w.Slug == slug {
			return w, nil
		}
	}
	return domain.Workspace{}, domain.ErrNotFound
}

func (f *fakes) Projects(_ context.Context, ws domain.WorkspaceID) ([]domain.Project, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.Project
	for _, p := range f.projects {
		if p.WorkspaceID == ws {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakes) Project(_ context.Context, id domain.ProjectID) (domain.Project, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, p := range f.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return domain.Project{}, domain.ErrNotFound
}

func (f *fakes) Workers(context.Context, string) ([]domain.Worker, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]domain.Worker(nil), f.workers...), nil
}

func (f *fakes) UsageByDay(_ context.Context, since time.Time) ([]domain.UsageDay, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.usageSince = since
	return f.usage, nil
}

func (f *fakes) Worker(_ context.Context, id domain.WorkerID) (domain.Worker, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, w := range f.workers {
		if w.ID == id {
			return w, nil
		}
	}
	return domain.Worker{}, domain.ErrNotFound
}

func (f *fakes) Subscribe(fn func(domain.Change)) func() {
	f.mu.Lock()
	f.listeners = append(f.listeners, fn)
	f.mu.Unlock()
	return func() {}
}

func (f *fakes) publish(c domain.Change) {
	f.mu.Lock()
	ls := slices.Clone(f.listeners)
	f.mu.Unlock()
	for _, fn := range ls {
		fn(c)
	}
}

func (f *fakes) listenerCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.listeners)
}

// Checklist
func (f *fakes) Get(context.Context) (checklist.Checklist, error) {
	if err := f.call("Checklist"); err != nil {
		return checklist.Checklist{}, err
	}
	return checklist.Checklist{Items: []checklist.Item{{ID: "server", Title: "Connect a server", Done: true}}, Percent: 100, Complete: true}, nil
}

// Tracker
func (f *fakes) Track(id domain.WorkerID) {
	f.mu.Lock()
	f.tracked = append(f.tracked, id)
	f.mu.Unlock()
}

// fakeStream is a session: signals come out of Next, writes are recorded.
type fakeStream struct {
	signals chan agent.Signal
	mu      sync.Mutex
	writes  [][]byte
	resizes [][2]uint16
	once    sync.Once
}

func newFakeStream() *fakeStream { return &fakeStream{signals: make(chan agent.Signal, 16)} }

func (s *fakeStream) Next() (agent.Signal, error) {
	sig, ok := <-s.signals
	if !ok {
		return agent.Signal{}, io.EOF
	}
	return sig, nil
}

func (s *fakeStream) Write(p []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writes = append(s.writes, append([]byte(nil), p...))
	return nil
}

func (s *fakeStream) Resize(rows, cols uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resizes = append(s.resizes, [2]uint16{rows, cols})
	return nil
}

func (s *fakeStream) Close() error {
	s.once.Do(func() { close(s.signals) })
	return nil
}

func (s *fakeStream) written() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []string
	for _, w := range s.writes {
		out = append(out, string(w))
	}
	return out
}

// newServer starts the API over the fakes. The log goes to f.logs.
func newServer(t *testing.T, origins ...string) (*httptest.Server, *fakes) {
	t.Helper()
	f := &fakes{stream: newFakeStream()}
	log := slog.New(slog.NewTextHandler(&lockedWriter{f: f}, nil))
	srv := New(Config{Version: "test", Origins: origins, Log: log, Now: time.Now, Static: f.static}, f, f, f, f, f, f)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, f
}

type lockedWriter struct{ f *fakes }

func (w *lockedWriter) Write(p []byte) (int, error) {
	w.f.mu.Lock()
	defer w.f.mu.Unlock()
	return w.f.logs.Write(p)
}

func (f *fakes) logged() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.logs.String()
}
