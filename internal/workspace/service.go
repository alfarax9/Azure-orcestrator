package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

// Runner runs a program as the agent user of workspace slug, in dir below its work folder,
// and returns its standard output.
type Runner interface {
	Run(ctx context.Context, slug, dir string, argv ...string) ([]byte, error)
}

// Provisioner prepares the machine for a new workspace: its agent user and folders.
type Provisioner interface {
	AddWorkspace(ctx context.Context, slug string) error
}

// Workspaces keeps workspaces.
type Workspaces interface {
	CreateWorkspace(ctx context.Context, w domain.Workspace) error
	Workspace(ctx context.Context, slug string) (domain.Workspace, error)
	WorkspaceOf(ctx context.Context, project domain.ProjectID) (domain.Workspace, error)
}

// Projects keeps projects.
type Projects interface {
	CreateProject(ctx context.Context, p domain.Project) error
	Project(ctx context.Context, id domain.ProjectID) (domain.Project, error)
	SetCommands(ctx context.Context, id domain.ProjectID, c domain.Commands) error
}

// Workers keeps workers.
type Workers interface {
	CreateWorker(ctx context.Context, w domain.Worker) error
	Worker(ctx context.Context, id domain.WorkerID) (domain.Worker, error)
	UpdateWorker(ctx context.Context, id domain.WorkerID, change func(*domain.Worker)) (domain.Worker, error)
}

var (
	// ErrInvalid means a name, URL or title was refused before anything was changed.
	ErrInvalid = errors.New("invalid")
	// ErrClone means git could not fetch the repository: it does not exist, or the agent user
	// has no access to it. The wrapped text holds git's own explanation.
	ErrClone = errors.New("could not clone the repository")
)

// Service creates workspaces, projects and workers.
type Service struct {
	run        Runner
	provision  Provisioner
	workspaces Workspaces
	projects   Projects
	workers    Workers
	now        func() time.Time
}

// New returns a Service. now is the clock; tests pass a fixed one.
func New(r Runner, p Provisioner, ws Workspaces, ps Projects, wk Workers, now func() time.Time) *Service {
	return &Service{run: r, provision: p, workspaces: ws, projects: ps, workers: wk, now: now}
}

// AddWorkspace creates a client workspace and its agent user (F2-01).
func (s *Service) AddWorkspace(ctx context.Context, slug, name string) (domain.Workspace, error) {
	name = strings.TrimSpace(name)
	if !domain.ValidSlug(slug) || name == "" {
		return domain.Workspace{}, fmt.Errorf("%w: a workspace needs a slug of 2 to 20 lowercase letters, digits and hyphens, and a name", ErrInvalid)
	}
	switch _, err := s.workspaces.Workspace(ctx, slug); {
	case err == nil:
		return domain.Workspace{}, fmt.Errorf("workspace %s: %w", slug, domain.ErrConflict)
	case !errors.Is(err, domain.ErrNotFound):
		return domain.Workspace{}, err
	}
	// The machine first: preparing it again is harmless, while a stored workspace without
	// an agent user could never run anything.
	if err := s.provision.AddWorkspace(ctx, slug); err != nil {
		return domain.Workspace{}, err
	}
	w := domain.Workspace{ID: domain.NewWorkspaceID(), Slug: slug, Name: name, CreatedAt: domain.Timestamp{Time: s.now()}}
	return w, s.workspaces.CreateWorkspace(ctx, w)
}

// SetCommands replaces what detection found with the user's own commands (F2-03).
func (s *Service) SetCommands(ctx context.Context, project domain.ProjectID, c domain.Commands) error {
	return s.projects.SetCommands(ctx, project, c)
}
