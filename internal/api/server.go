package api

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alfarax9/astra/internal/agent"
	"github.com/alfarax9/astra/internal/checklist"
	"github.com/alfarax9/astra/internal/domain"
)

// Auth pairs devices and recognises them.
type Auth interface {
	Pair(ctx context.Context, pairing, name string) (domain.Device, domain.Secret, error)
	Authenticate(ctx context.Context, token string) (domain.Device, error)
}

// Workspaces creates workspaces, projects and workers.
type Workspaces interface {
	AddWorkspace(ctx context.Context, slug, name string) (domain.Workspace, error)
	AddProject(ctx context.Context, slug, name, url string) (domain.Project, error)
	SetCommands(ctx context.Context, id domain.ProjectID, c domain.Commands) error
	CreateWorker(ctx context.Context, project domain.ProjectID, title string, a domain.AgentKind, clientReview bool) (domain.Worker, error)
}

// Agents runs workers' agents.
type Agents interface {
	Start(ctx context.Context, id domain.WorkerID) error
	Stop(ctx context.Context, id domain.WorkerID) error
	Interrupt(ctx context.Context, id domain.WorkerID) error
	Send(ctx context.Context, id domain.WorkerID, text string) error
	Open(ctx context.Context, id domain.WorkerID) (agent.Stream, error)
	NoteInput(ctx context.Context, id domain.WorkerID, input []byte) error
}

// Reads reads the facts and reports their changes.
type Reads interface {
	Workspaces(ctx context.Context) ([]domain.Workspace, error)
	Workspace(ctx context.Context, slug string) (domain.Workspace, error)
	Projects(ctx context.Context, workspace domain.WorkspaceID) ([]domain.Project, error)
	Project(ctx context.Context, id domain.ProjectID) (domain.Project, error)
	Workers(ctx context.Context, workspaceSlug string) ([]domain.Worker, error)
	Worker(ctx context.Context, id domain.WorkerID) (domain.Worker, error)
	UsageByDay(ctx context.Context, since time.Time) ([]domain.UsageDay, error)
	Subscribe(fn func(domain.Change)) (cancel func())
}

// Checklist works out the first-run checklist.
type Checklist interface {
	Get(ctx context.Context) (checklist.Checklist, error)
}

// Tracker keeps a started worker's facts true for as long as its agent runs. The daemon's own
// lifetime, not a request's, decides when that ends.
type Tracker interface {
	Track(id domain.WorkerID)
}

// Config is what the server needs besides its collaborators.
type Config struct {
	Version string
	// Origins are the web origins, such as http://localhost:5173, that may call the API from a
	// browser. Requests with no Origin (desktop app, CLI) are always fine.
	Origins []string
	Log     *slog.Logger
	Now     func() time.Time
	// Static is the built web UI, served from the same origin as the API so one tunnel or one
	// proxy reaches both. Nil serves only the API.
	Static fs.FS
}

// Server is the API.
type Server struct {
	cfg     Config
	auth    Auth
	ws      Workspaces
	agents  Agents
	reads   Reads
	tracker Tracker
	list    Checklist
}

// New returns a Server.
func New(cfg Config, a Auth, ws Workspaces, ag Agents, r Reads, t Tracker, c Checklist) *Server {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Server{cfg: cfg, auth: a, ws: ws, agents: ag, reads: r, tracker: t, list: c}
}

// apiVersion changes only when a client of the previous one would break (N-16).
const apiVersion = 1

// Handler returns the routes, wrapped in the protections every request gets.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	open := func(pattern string, h http.HandlerFunc) { mux.Handle(pattern, h) }
	authed := func(pattern string, h http.HandlerFunc) { mux.Handle(pattern, s.requireDevice(h, false)) }
	socket := func(pattern string, h http.HandlerFunc) { mux.Handle(pattern, s.requireDevice(h, true)) }

	open("GET /api/v1/version", s.version)
	open("POST /api/v1/pairing", s.pair)
	authed("GET /api/v1/workspaces", s.listWorkspaces)
	authed("POST /api/v1/workspaces", s.addWorkspace)
	authed("GET /api/v1/projects", s.listProjects)
	authed("POST /api/v1/projects", s.addProject)
	authed("PUT /api/v1/projects/{id}/commands", s.setCommands)
	authed("GET /api/v1/checklist", s.getChecklist)
	authed("GET /api/v1/workers", s.listWorkers)
	authed("GET /api/v1/workers/{id}", s.getWorker)
	authed("GET /api/v1/usage", s.getUsage)
	authed("POST /api/v1/projects/{id}/workers", s.newWorker)
	authed("POST /api/v1/workers/{id}/start", s.startWorker)
	authed("POST /api/v1/workers/{id}/stop", s.stopWorker)
	authed("POST /api/v1/workers/{id}/interrupt", s.interruptWorker)
	authed("POST /api/v1/workers/{id}/messages", s.sendMessage)
	socket("GET /api/v1/events", s.events)
	socket("GET /api/v1/sessions/{id}/pty", s.terminal)
	// Everything else: an unknown API route is a JSON 404, and any other path is the web UI when
	// there is one. One handler decides, since a second pattern for "/" would conflict with "/api/".
	files := s.static()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || s.cfg.Static == nil || (r.Method != http.MethodGet && r.Method != http.MethodHead) {
			s.fail(w, r, errNoRoute)
			return
		}
		files.ServeHTTP(w, r)
	})
	return s.protect(mux)
}

// originHosts are the hosts of the allowed origins, for the WebSocket origin check.
func (s *Server) originHosts() []string {
	var hosts []string
	for _, o := range s.cfg.Origins {
		if u, err := url.Parse(o); err == nil && u.Host != "" {
			hosts = append(hosts, u.Host)
		}
	}
	return hosts
}
