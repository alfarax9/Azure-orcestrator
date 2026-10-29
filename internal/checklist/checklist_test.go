package checklist

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alfarax9/astra/internal/domain"
)

type directory struct {
	workspaces []domain.Workspace
	projects   map[domain.WorkspaceID][]domain.Project
	err        error
}

func (d directory) Workspaces(context.Context) ([]domain.Workspace, error) {
	return d.workspaces, d.err
}
func (d directory) Projects(_ context.Context, id domain.WorkspaceID) ([]domain.Project, error) {
	return d.projects[id], d.err
}

type logins map[string]map[string]bool

func (l logins) AgentLogin(_ context.Context, slug string) (map[string]bool, error) {
	if l == nil {
		return nil, errors.New("astra-priv is not installed")
	}
	return l[slug], nil
}

func item(c Checklist, id string) Item {
	for _, it := range c.Items {
		if it.ID == id {
			return it
		}
	}
	return Item{}
}

func TestANewServerIsAThirdDoneBecauseOnlyTheServerStepIsTrue(t *testing.T) {
	c, err := New(directory{}, logins{}).Get(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !item(c, "server").Done || item(c, "agent").Done || item(c, "project").Done {
		t.Errorf("items = %+v", c.Items)
	}
	if c.Percent != 33 || c.Complete {
		t.Errorf("percent %d, complete %v", c.Percent, c.Complete)
	}
	if !strings.Contains(item(c, "agent").Detail, "Add a project first") {
		t.Errorf("agent detail = %q", item(c, "agent").Detail)
	}
}

func TestTheDomainStepIsShownButNeitherCountedNorRequired(t *testing.T) {
	ws := domain.Workspace{ID: "w", Slug: "warung-sari"}
	d := directory{workspaces: []domain.Workspace{ws}, projects: map[domain.WorkspaceID][]domain.Project{"w": {{ID: "p"}}}}
	c, _ := New(d, logins{"warung-sari": {"claude": true}}).Get(t.Context())
	domainItem := item(c, "domain")
	if !domainItem.Soon || domainItem.Done {
		t.Errorf("domain = %+v", domainItem)
	}
	if c.Percent != 100 || !c.Complete {
		t.Errorf("with everything possible done: %d%%, complete %v", c.Percent, c.Complete)
	}
	if got := item(c, "agent").Detail; got != "Signed in: Claude Code in warung-sari." {
		t.Errorf("agent detail = %q", got)
	}
}

func TestProgressIsCountedFromTheFacts(t *testing.T) {
	ws := domain.Workspace{ID: "w", Slug: "warung-sari"}
	withProject := directory{workspaces: []domain.Workspace{ws}, projects: map[domain.WorkspaceID][]domain.Project{"w": {{ID: "p"}}}}
	c, _ := New(withProject, logins{"warung-sari": {}}).Get(t.Context())
	if c.Percent != 66 || c.Complete || !item(c, "project").Done || item(c, "agent").Done {
		t.Errorf("a project but no login: %d%% %+v", c.Percent, c.Items)
	}
	if !strings.Contains(item(c, "agent").Detail, "run claude or codex") {
		t.Errorf("agent detail = %q", item(c, "agent").Detail)
	}
	both := logins{"warung-sari": {"claude": true, "codex": true}}
	c, _ = New(withProject, both).Get(t.Context())
	if got := item(c, "agent").Detail; got != "Signed in: Claude Code in warung-sari, Codex in warung-sari." {
		t.Errorf("agent detail = %q", got)
	}
}

func TestALoginCheckThatFailsIsSaidAndNotCountedAsDone(t *testing.T) {
	ws := domain.Workspace{ID: "w", Slug: "warung-sari"}
	d := directory{workspaces: []domain.Workspace{ws}, projects: map[domain.WorkspaceID][]domain.Project{"w": {{ID: "p"}}}}
	c, err := New(d, logins(nil)).Get(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if got := item(c, "agent"); got.Done || !strings.Contains(got.Detail, "Couldn't check") {
		t.Errorf("agent = %+v", got)
	}
}

func TestAFailingDirectoryIsAnError(t *testing.T) {
	if _, err := New(directory{err: errors.New("database is locked")}, logins{}).Get(t.Context()); err == nil {
		t.Error("an unreadable directory must be an error")
	}
	bad := directory{workspaces: []domain.Workspace{{ID: "w"}}}
	bad.err = nil
	if _, err := New(bad, logins{}).Get(t.Context()); err != nil {
		t.Errorf("a workspace without projects: %v", err)
	}
}
