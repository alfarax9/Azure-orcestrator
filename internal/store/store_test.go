package store

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.Context(), filepath.Join(t.TempDir(), "astra.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func workspace(slug string) domain.Workspace {
	return domain.Workspace{ID: domain.NewWorkspaceID(), Slug: slug, Name: "Client " + slug, CreatedAt: domain.Timestamp{Time: time.UnixMilli(1700000000000)}}
}

func project(ws domain.Workspace, name string) domain.Project {
	return domain.Project{
		ID: domain.NewProjectID(), WorkspaceID: ws.ID, Name: name, RepoURL: "https://example.com/" + name + ".git",
		DefaultBranch: "main", Stack: "nextjs",
		Commands:  domain.Commands{Dev: "npm run dev", Build: "npm run build", Test: "npm test"},
		CreatedAt: domain.Timestamp{Time: time.UnixMilli(1700000001000)},
	}
}

// seed creates a workspace with one project and returns both.
func seed(t *testing.T, s *Store, slug string) (domain.Workspace, domain.Project) {
	t.Helper()
	ws := workspace(slug)
	p := project(ws, "web")
	if err := s.CreateWorkspace(t.Context(), ws); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateProject(t.Context(), p); err != nil {
		t.Fatal(err)
	}
	return ws, p
}

func TestWorkspacesAndProjectsRoundTrip(t *testing.T) {
	ctx := t.Context()
	s := open(t)
	ws, p := seed(t, s, "acme")

	got, err := s.Workspace(ctx, "acme")
	if err != nil || got != ws {
		t.Fatalf("workspace: got %+v, %v", got, err)
	}
	gotP, err := s.Project(ctx, p.ID)
	if err != nil || gotP != p {
		t.Fatalf("project: got %+v, %v", gotP, err)
	}
	owner, err := s.WorkspaceOf(ctx, p.ID)
	if err != nil || owner.Slug != "acme" || owner.AgentUser() != "astra-agent-acme" {
		t.Fatalf("workspace of project: %+v, %v", owner, err)
	}
	list, err := s.Projects(ctx, ws.ID)
	if err != nil || len(list) != 1 || list[0] != p {
		t.Fatalf("projects: %+v, %v", list, err)
	}
	all, err := s.Workspaces(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("workspaces: %+v, %v", all, err)
	}
}

func TestCorrectedCommandsAreKept(t *testing.T) {
	ctx := t.Context()
	s := open(t)
	_, p := seed(t, s, "acme")
	fixed := domain.Commands{Dev: "pnpm dev", Build: "pnpm build"}
	if err := s.SetCommands(ctx, p.ID, fixed); err != nil {
		t.Fatal(err)
	}
	got, err := s.Project(ctx, p.ID)
	if err != nil || got.Commands != fixed || got.Stack != "nextjs" {
		t.Fatalf("project = %+v, %v", got, err)
	}
	if err := s.SetCommands(ctx, domain.NewProjectID(), fixed); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown project: err = %v, want ErrNotFound", err)
	}
}

func TestConstraintsComeBackAsDomainErrors(t *testing.T) {
	ctx := t.Context()
	s := open(t)
	ws, p := seed(t, s, "acme")

	dup := workspace("acme")
	if err := s.CreateWorkspace(ctx, dup); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("a taken slug: %v", err)
	}
	if err := s.CreateProject(ctx, project(ws, "web")); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("a repeated project name in one workspace: %v", err)
	}
	orphan := project(workspace("ghost"), "x")
	if err := s.CreateProject(ctx, orphan); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("a project in an unknown workspace: %v", err)
	}
	stray := domain.NewWorker("unknown", "t", domain.AgentClaude, "b", true, time.Now())
	if err := s.CreateWorker(ctx, stray); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("a worker in an unknown project: %v", err)
	}
	other := workspace("rina")
	if err := s.CreateWorkspace(ctx, other); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateProject(ctx, project(other, "web")); err != nil {
		t.Errorf("the same project name in another workspace is fine: %v", err)
	}
	for name, err := range map[string]error{
		"workspace": func() error { _, e := s.Workspace(ctx, "nope"); return e }(),
		"project":   func() error { _, e := s.Project(ctx, "nope"); return e }(),
		"worker":    func() error { _, e := s.Worker(ctx, "nope"); return e }(),
		"update":    func() error { _, e := s.UpdateWorker(ctx, "nope", func(*domain.Worker) {}); return e }(),
		"owner":     func() error { _, e := s.WorkspaceOf(ctx, "nope"); return e }(),
	} {
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s: %v", name, err)
		}
	}
	_ = p
}

func TestWorkersKeepTheirFactsAndNeverAStatus(t *testing.T) {
	ctx := t.Context()
	s := open(t)
	_, p := seed(t, s, "acme")

	w := domain.NewWorker(p.ID, "Fix checkout total", domain.AgentCodex, "fix/checkout-total", true, time.UnixMilli(1700000002000))
	w.Worktree = "/var/lib/astra/agents/acme/work/x"
	if err := s.CreateWorker(ctx, w); err != nil {
		t.Fatal(err)
	}
	got, err := s.Worker(ctx, w.ID)
	if err != nil || got != w {
		t.Fatalf("round trip: got %+v, %v", got, err)
	}
	updated, err := s.UpdateWorker(ctx, w.ID, func(w *domain.Worker) {
		w.Activity = domain.ActivityIdle
		w.PreviewLive, w.Shared, w.NewFeedback = true, true, 2
		w.Gate = domain.GateFact{State: domain.GateFailed, Cause: "2 tests failing"}
		w.Added, w.Removed = 184, 22
		// Identity cannot be rewritten by an update.
		w.ID, w.Agent, w.ProjectID = "other", domain.AgentClaude, "other"
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != w.ID || updated.Agent != domain.AgentCodex || updated.ProjectID != p.ID {
		t.Errorf("identity changed: %+v", updated)
	}
	stored, _ := s.Worker(ctx, w.ID)
	if stored != updated || stored.Gate.Cause != "2 tests failing" || stored.Status() != domain.StatusGateFailed {
		t.Errorf("stored %+v, status %s", stored, stored.Status())
	}
}

func TestWorkersListNewestFirstPerWorkspace(t *testing.T) {
	ctx := t.Context()
	s := open(t)
	_, acme := seed(t, s, "acme")
	_, rina := seed(t, s, "rina")
	var ids []domain.WorkerID
	for _, p := range []domain.Project{acme, acme, rina} {
		w := domain.NewWorker(p.ID, "task", domain.AgentClaude, "b", true, time.Now())
		if err := s.CreateWorker(ctx, w); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, w.ID)
	}
	acmeWorkers, err := s.Workers(ctx, "acme")
	if err != nil || len(acmeWorkers) != 2 || acmeWorkers[0].ID != ids[1] || acmeWorkers[1].ID != ids[0] {
		t.Fatalf("acme: got %v, %v", acmeWorkers, err)
	}
	if rinaWorkers, _ := s.Workers(ctx, "rina"); len(rinaWorkers) != 1 || rinaWorkers[0].ID != ids[2] {
		t.Errorf("a workspace must not see another's workers: %v", rinaWorkers)
	}
	if everyone, _ := s.Workers(ctx, ""); len(everyone) != 3 {
		t.Errorf("an empty slug lists everything, got %d", len(everyone))
	}
	if none, err := s.Workers(ctx, "nobody"); err != nil || len(none) != 0 {
		t.Errorf("unknown workspace: %v, %v", none, err)
	}
}

func TestConcurrentUpdatesNeverLoseAFact(t *testing.T) {
	ctx := t.Context()
	s := open(t)
	_, p := seed(t, s, "acme")
	w := domain.NewWorker(p.ID, "busy", domain.AgentClaude, "b", true, time.Now())
	if err := s.CreateWorker(ctx, w); err != nil {
		t.Fatal(err)
	}
	const updates = 60
	var wg sync.WaitGroup
	for range updates {
		wg.Go(func() {
			if _, err := s.UpdateWorker(ctx, w.ID, func(w *domain.Worker) { w.Added++ }); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if got, _ := s.Worker(ctx, w.ID); got.Added != updates {
		t.Fatalf("lost updates: %d of %d", got.Added, updates)
	}
}

func TestDataSurvivesReopeningAndMigrationsRunOnce(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "astra.db")
	s, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	ws := workspace("acme")
	if err := s.CreateWorkspace(ctx, ws); err != nil {
		t.Fatal(err)
	}
	var version int
	files, _ := fs.Glob(migrations, "migrations/*.sql")
	if err := s.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil || version != len(files) {
		t.Fatalf("schema version %d, %v, want %d", version, err, len(files))
	}
	_ = s.Close()

	again, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() { _ = again.Close() }()
	if got, err := again.Workspace(ctx, "acme"); err != nil || got.ID != ws.ID {
		t.Fatalf("after reopening: %+v, %v", got, err)
	}
}

func TestOpenFailsOnAnUnusablePathAndAnInMemoryDatabaseWorks(t *testing.T) {
	if _, err := Open(t.Context(), filepath.Join(t.TempDir(), "missing", "dir", "astra.db")); err == nil {
		t.Error("expected an error for a folder that does not exist")
	}
	s, err := Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if err := s.CreateWorkspace(t.Context(), workspace("acme")); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationNamesMustStartWithTheirVersion(t *testing.T) {
	for name, ok := range map[string]bool{"migrations/001_init.sql": true, "migrations/12_x.sql": true, "migrations/init.sql": false, "migrations/000_x.sql": false} {
		if _, err := migrationVersion(name); (err == nil) != ok {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestQualifiedPrefixesEveryColumn(t *testing.T) {
	if got := qualified("w", "id, title,\n\tagent"); got != "w.id, w.title, w.agent" {
		t.Fatalf("got %q", got)
	}
}
