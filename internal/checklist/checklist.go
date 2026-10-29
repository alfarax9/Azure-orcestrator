package checklist

import (
	"context"
	"fmt"
	"strings"

	"github.com/alfarax9/astra/internal/domain"
)

// Directory lists workspaces and their projects.
type Directory interface {
	Workspaces(ctx context.Context) ([]domain.Workspace, error)
	Projects(ctx context.Context, workspace domain.WorkspaceID) ([]domain.Project, error)
}

// Logins says which agents are signed in for a workspace, by agent kind ("claude", "codex").
type Logins interface {
	AgentLogin(ctx context.Context, slug string) (map[string]bool, error)
}

// Item is one step of the checklist.
type Item struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Done   bool   `json:"done"`
	// Soon marks a step that cannot be done yet because its feature is not built. It is shown,
	// and counted in neither the percentage nor in whether the checklist is complete.
	Soon bool `json:"soon"`
}

// Checklist is the whole list and how far along it is.
type Checklist struct {
	Items    []Item `json:"items"`
	Percent  int    `json:"percent"`
	Complete bool   `json:"complete"`
}

// Service builds the checklist.
type Service struct {
	dir    Directory
	logins Logins
}

// New returns a Service.
func New(d Directory, l Logins) *Service { return &Service{dir: d, logins: l} }

var agentNames = map[string]string{"claude": "Claude Code", "codex": "Codex"}

// Get reads the facts and returns the checklist.
func (s *Service) Get(ctx context.Context) (Checklist, error) {
	workspaces, err := s.dir.Workspaces(ctx)
	if err != nil {
		return Checklist{}, err
	}
	projects := 0
	for _, w := range workspaces {
		ps, err := s.dir.Projects(ctx, w.ID)
		if err != nil {
			return Checklist{}, err
		}
		projects += len(ps)
	}
	items := []Item{
		{ID: "server", Title: "Connect a server", Detail: "This server runs the daemon you are talking to.", Done: true},
		{ID: "domain", Title: "Connect a domain", Detail: "Previews and production need a domain. This arrives with previews.", Soon: true},
		s.agents(ctx, workspaces),
		{ID: "project", Title: "Add your first project", Detail: "Clone a repository, or start an empty one, then give an agent a task.", Done: projects > 0},
	}
	return summarise(items), nil
}

// agents is the sign-in step: done once any workspace has any agent signed in.
func (s *Service) agents(ctx context.Context, workspaces []domain.Workspace) Item {
	item := Item{ID: "agent", Title: "Sign in to an agent"}
	if len(workspaces) == 0 {
		item.Detail = "Add a project first. Then open its terminal and run claude or codex to sign in with your own account."
		return item
	}
	var signed []string
	for _, w := range workspaces {
		status, err := s.logins.AgentLogin(ctx, w.Slug)
		if err != nil {
			item.Detail = "Couldn't check the agents' sign-in: " + err.Error()
			return item
		}
		for _, kind := range []string{"claude", "codex"} {
			if status[kind] {
				signed = append(signed, fmt.Sprintf("%s in %s", agentNames[kind], w.Slug))
			}
		}
	}
	if len(signed) == 0 {
		item.Detail = "Open a worker's terminal and run claude or codex, then sign in with your own account. Astra never sees the login."
		return item
	}
	item.Done, item.Detail = true, "Signed in: "+strings.Join(signed, ", ")+"."
	return item
}

func summarise(items []Item) Checklist {
	counted, done := 0, 0
	for _, it := range items {
		if it.Soon {
			continue
		}
		counted++
		if it.Done {
			done++
		}
	}
	percent := 100
	if counted > 0 {
		percent = done * 100 / counted
	}
	return Checklist{Items: items, Percent: percent, Complete: done == counted}
}
