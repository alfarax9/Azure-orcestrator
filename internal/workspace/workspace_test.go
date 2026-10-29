package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

// localRunner runs commands as the test's own user under root/<slug>/work, the way the
// real runner does in local mode.
type localRunner struct {
	root  string
	calls int
}

func (r *localRunner) Run(ctx context.Context, slug, dir string, argv ...string) ([]byte, error) {
	r.calls++
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = filepath.Join(r.root, slug, "work", dir)
	cmd.Env = append(os.Environ(), "HOME="+r.root, "GIT_CONFIG_NOSYSTEM=1")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("%v: %w: %s", argv, err, stderr.String())
	}
	return out, nil
}

func (r *localRunner) AddWorkspace(_ context.Context, slug string) error {
	return os.MkdirAll(filepath.Join(r.root, slug, "work"), 0o700)
}

// memStore keeps everything in maps and can be told to fail its next write.
type memStore struct {
	workspaces map[string]domain.Workspace
	projects   map[domain.ProjectID]domain.Project
	workers    map[domain.WorkerID]domain.Worker
	failWrite  error
}

func (m *memStore) CreateWorkspace(_ context.Context, w domain.Workspace) error {
	m.workspaces[w.Slug] = w
	return nil
}

func (m *memStore) Workspace(_ context.Context, slug string) (domain.Workspace, error) {
	if w, ok := m.workspaces[slug]; ok {
		return w, nil
	}
	return domain.Workspace{}, domain.ErrNotFound
}

func (m *memStore) WorkspaceOf(_ context.Context, id domain.ProjectID) (domain.Workspace, error) {
	for _, w := range m.workspaces {
		if w.ID == m.projects[id].WorkspaceID {
			return w, nil
		}
	}
	return domain.Workspace{}, domain.ErrNotFound
}

func (m *memStore) CreateProject(_ context.Context, p domain.Project) error {
	if m.failWrite != nil {
		return m.failWrite
	}
	m.projects[p.ID] = p
	return nil
}

func (m *memStore) Project(_ context.Context, id domain.ProjectID) (domain.Project, error) {
	if p, ok := m.projects[id]; ok {
		return p, nil
	}
	return domain.Project{}, domain.ErrNotFound
}

func (m *memStore) SetCommands(_ context.Context, id domain.ProjectID, c domain.Commands) error {
	p, ok := m.projects[id]
	if !ok {
		return domain.ErrNotFound
	}
	p.Commands = c
	m.projects[id] = p
	return nil
}

func (m *memStore) Worker(_ context.Context, id domain.WorkerID) (domain.Worker, error) {
	if w, ok := m.workers[id]; ok {
		return w, nil
	}
	return domain.Worker{}, domain.ErrNotFound
}

func (m *memStore) UpdateWorker(_ context.Context, id domain.WorkerID, change func(*domain.Worker)) (domain.Worker, error) {
	w, ok := m.workers[id]
	if !ok {
		return w, domain.ErrNotFound
	}
	change(&w)
	m.workers[id] = w
	return w, nil
}

func (m *memStore) CreateWorker(_ context.Context, w domain.Worker) error {
	if m.failWrite != nil {
		return m.failWrite
	}
	m.workers[w.ID] = w
	return nil
}

type fixture struct {
	svc   *Service
	run   *localRunner
	store *memStore
	root  string
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	root := t.TempDir()
	run := &localRunner{root: root}
	st := &memStore{workspaces: map[string]domain.Workspace{}, projects: map[domain.ProjectID]domain.Project{},
		workers: map[domain.WorkerID]domain.Worker{}}
	now := func() time.Time { return time.UnixMilli(1700000000000) }
	f := fixture{svc: New(run, run, st, st, st, now), run: run, store: st, root: root}
	if _, err := f.svc.AddWorkspace(t.Context(), "acme", "Acme"); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f fixture) work(parts ...string) string {
	return filepath.Join(append([]string{f.root, "acme", "work"}, parts...)...)
}

// git runs git in dir for test setup, with an author.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// remote makes a repository with a Next.js package.json on branch trunk and returns its URL.
func remote(t *testing.T) (dir, url string) {
	t.Helper()
	dir = t.TempDir()
	git(t, dir, "init", "--quiet", "-b", "trunk")
	pkg := `{"dependencies": {"next": "15"}, "scripts": {"dev": "next dev", "build": "next build"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", ".")
	git(t, dir, "commit", "--quiet", "-m", "first")
	return dir, "file://" + dir
}

func TestAddWorkspaceRefusesBadOrTakenSlugs(t *testing.T) {
	f := newFixture(t)
	if _, err := f.svc.AddWorkspace(t.Context(), "acme", "Again"); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("taken slug: err = %v", err)
	}
	for _, c := range [][2]string{{"Bad", "x"}, {"ok-slug", " "}, {"../x", "x"}} {
		if _, err := f.svc.AddWorkspace(t.Context(), c[0], c[1]); !errors.Is(err, ErrInvalid) {
			t.Errorf("%q: err = %v, want ErrInvalid", c, err)
		}
	}
	if w := f.store.workspaces["acme"]; w.Name != "Acme" || w.CreatedAt.UnixMilli() != 1700000000000 {
		t.Errorf("stored %+v", w)
	}
}

func TestAddProjectClonesAndDetects(t *testing.T) {
	f := newFixture(t)
	_, url := remote(t)
	p, err := f.svc.AddProject(t.Context(), "acme", "site", url)
	if err != nil {
		t.Fatal(err)
	}
	if p.DefaultBranch != "trunk" || p.Stack != "nextjs" || p.Commands.Dev != "npm run dev" || p.RepoURL != url {
		t.Errorf("project = %+v", p)
	}
	if _, err := os.Stat(f.work("site", "repo", "package.json")); err != nil {
		t.Errorf("clone missing: %v", err)
	}
	if f.store.projects[p.ID].Name != "site" {
		t.Error("project not stored")
	}
	fixed := domain.Commands{Dev: "pnpm dev"}
	if err := f.svc.SetCommands(t.Context(), p.ID, fixed); err != nil || f.store.projects[p.ID].Commands != fixed {
		t.Errorf("SetCommands: %v, %+v", err, f.store.projects[p.ID].Commands)
	}
}

func TestAddProjectStartsAnEmptyRepository(t *testing.T) {
	f := newFixture(t)
	p, err := f.svc.AddProject(t.Context(), "acme", "fresh", "")
	if err != nil {
		t.Fatal(err)
	}
	if p.DefaultBranch != "main" || p.Stack != "" {
		t.Errorf("project = %+v", p)
	}
	if msg := git(t, f.work("fresh", "repo"), "log", "--format=%s %an"); msg != "Start the project Astra" {
		t.Errorf("first commit = %q", msg)
	}
}

func TestAddProjectRefusesBadInputBeforeRunningAnything(t *testing.T) {
	f := newFixture(t)
	f.run.calls = 0
	for _, c := range [][2]string{
		{"Site", ""}, {"../x", ""}, {"-x", ""}, {"site", "ext::sh -c id"}, {"site", "-uhttps://x"},
		{"site", "http://example.com/x.git"}, {"site", "https://example.com/x y.git"}, {"site", "/tmp/repo"},
	} {
		if _, err := f.svc.AddProject(t.Context(), "acme", c[0], c[1]); !errors.Is(err, ErrInvalid) {
			t.Errorf("%q: err = %v, want ErrInvalid", c, err)
		}
	}
	if f.run.calls != 0 {
		t.Errorf("%d commands ran for refused input", f.run.calls)
	}
	if _, err := f.svc.AddProject(t.Context(), "nobody", "site", ""); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown workspace: err = %v", err)
	}
}

func TestAddProjectNeverTouchesATakenName(t *testing.T) {
	f := newFixture(t)
	_, url := remote(t)
	first, err := f.svc.AddProject(t.Context(), "acme", "site", url)
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range []string{url, ""} {
		if _, err := f.svc.AddProject(t.Context(), "acme", "site", u); !errors.Is(err, domain.ErrConflict) {
			t.Errorf("url %q: err = %v, want ErrConflict", u, err)
		}
	}
	if _, err := os.Stat(f.work("site", "repo", "package.json")); err != nil || len(f.store.projects) != 1 {
		t.Errorf("the first project %s was damaged: %v", first.ID, err)
	}
}

func TestAddProjectCleansUpWhatFailed(t *testing.T) {
	f := newFixture(t)
	if _, err := f.svc.AddProject(t.Context(), "acme", "gone", "file:///no/such/repo"); err == nil {
		t.Fatal("a clone of a missing repository must fail")
	}
	_, url := remote(t)
	f.store.failWrite = errors.New("disk full")
	if _, err := f.svc.AddProject(t.Context(), "acme", "site", url); err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("err = %v", err)
	}
	for _, dir := range []string{f.work("gone", "repo"), f.work("site", "repo")} {
		if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s was left behind", dir)
		}
	}
}

func TestWorkersGetTheirOwnWorktreeAndBranch(t *testing.T) {
	f := newFixture(t)
	origin, url := remote(t)
	p, err := f.svc.AddProject(t.Context(), "acme", "site", url)
	if err != nil {
		t.Fatal(err)
	}
	// A change on the default branch after the clone: workers start from it.
	if err := os.WriteFile(filepath.Join(origin, "later.txt"), []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, origin, "add", ".")
	git(t, origin, "commit", "--quiet", "-m", "later")

	a, err := f.svc.CreateWorker(t.Context(), p.ID, "Fix the header!", domain.AgentClaude, true)
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.svc.CreateWorker(t.Context(), p.ID, "Fix the header!", domain.AgentCodex, false)
	if err != nil {
		t.Fatal(err)
	}
	if a.Branch == b.Branch || a.Worktree == b.Worktree || !strings.HasPrefix(a.Branch, "astra/fix-the-header-") {
		t.Errorf("branches %q %q, worktrees %q %q", a.Branch, b.Branch, a.Worktree, b.Worktree)
	}
	dirA, dirB := f.work(a.Worktree), f.work(b.Worktree)
	if _, err := os.Stat(filepath.Join(dirA, "later.txt")); err != nil {
		t.Error("the worker did not start from the latest default branch")
	}
	if err := os.WriteFile(filepath.Join(dirA, "header.txt"), []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dirB, "header.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Error("a file written by one worker showed up in the other's worktree")
	}
	if got := git(t, dirA, "branch", "--show-current"); got != a.Branch {
		t.Errorf("worktree is on %q, want %q", got, a.Branch)
	}
	if out, err := exec.CommandContext(t.Context(), "git", "-C", dirA, "rev-parse", "--abbrev-ref", "@{u}").CombinedOutput(); err == nil {
		t.Errorf("the worker branch tracks %s; it must not push to the default branch", out)
	}
	if len(f.store.workers) != 2 || f.store.workers[a.ID].Activity != domain.ActivityWorking {
		t.Errorf("stored workers: %+v", f.store.workers)
	}
}

func TestCreateWorkerFromAnEmptyRepository(t *testing.T) {
	f := newFixture(t)
	p, err := f.svc.AddProject(t.Context(), "acme", "fresh", "")
	if err != nil {
		t.Fatal(err)
	}
	w, err := f.svc.CreateWorker(t.Context(), p.ID, "日本語", domain.AgentClaude, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(w.Branch, "astra/task-") {
		t.Errorf("branch = %q", w.Branch)
	}
}

func TestCreateWorkerUndoesTheWorktreeWhenStoringFails(t *testing.T) {
	f := newFixture(t)
	_, url := remote(t)
	p, err := f.svc.AddProject(t.Context(), "acme", "site", url)
	if err != nil {
		t.Fatal(err)
	}
	f.store.failWrite = errors.New("disk full")
	if _, err := f.svc.CreateWorker(t.Context(), p.ID, "task", domain.AgentClaude, true); err == nil {
		t.Fatal("want the store's error")
	}
	if entries, _ := os.ReadDir(f.work("site", "workers")); len(entries) != 0 {
		t.Errorf("worktrees left: %v", entries)
	}
	if branches := git(t, f.work("site", "repo"), "branch", "--list", "astra/*"); branches != "" {
		t.Errorf("branches left: %q", branches)
	}
}

func TestCreateWorkerRefusesBadInput(t *testing.T) {
	f := newFixture(t)
	p, err := f.svc.AddProject(t.Context(), "acme", "fresh", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		title string
		agent domain.AgentKind
	}{{"", domain.AgentClaude}, {strings.Repeat("x", 201), domain.AgentClaude}, {"ok", "gpt"}} {
		if _, err := f.svc.CreateWorker(t.Context(), p.ID, c.title, c.agent, true); !errors.Is(err, ErrInvalid) {
			t.Errorf("%+v: err = %v", c, err)
		}
	}
	if _, err := f.svc.CreateWorker(t.Context(), domain.NewProjectID(), "ok", domain.AgentClaude, true); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown project: err = %v", err)
	}
}

func TestBranchNames(t *testing.T) {
	id := domain.WorkerID("0198a3f2-7c1d-7000-8000-00000abcdef1")
	for title, want := range map[string]string{
		"Fix the header!":         "astra/fix-the-header-bcdef1",
		"  --Add  /login page-- ": "astra/add-login-page-bcdef1",
		"???":                     "astra/task-bcdef1",
		strings.Repeat("ab ", 30): "astra/ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-bcdef1",
	} {
		if got := branchName(title, id); got != want {
			t.Errorf("branchName(%q) = %q, want %q", title, got, want)
		}
	}
}

func TestRefreshCountsEverythingAWorkerChanged(t *testing.T) {
	f := newFixture(t)
	_, url := remote(t)
	p, err := f.svc.AddProject(t.Context(), "acme", "site", url)
	if err != nil {
		t.Fatal(err)
	}
	w, err := f.svc.CreateWorker(t.Context(), p.ID, "Count", domain.AgentClaude, true)
	if err != nil {
		t.Fatal(err)
	}
	dir := f.work(w.Worktree)
	if err := f.svc.Refresh(t.Context(), w.ID); err != nil || f.store.workers[w.ID].Added != 0 {
		t.Fatalf("a fresh worktree has no changes: %v, %+v", err, f.store.workers[w.ID])
	}
	// One new file of 3 lines, and a committed change of one line to package.json plus a
	// second, uncommitted one.
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("new.txt", "a\nb\nc\n")
	write("package.json", "{}\n")
	git(t, dir, "add", "package.json")
	git(t, dir, "commit", "--quiet", "-m", "shrink")
	write("package.json", "{}\n// more\n")
	if err := f.svc.Refresh(t.Context(), w.ID); err != nil {
		t.Fatal(err)
	}
	got := f.store.workers[w.ID]
	if got.Added != 5 || got.Removed != 1 {
		t.Errorf("counted +%d -%d, want +5 -1", got.Added, got.Removed)
	}
	if status := git(t, dir, "status", "--porcelain"); !strings.Contains(status, "?? new.txt") {
		t.Errorf("the worker's own staging was touched: %q", status)
	}
	if err := f.svc.Refresh(t.Context(), "nobody"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown worker: err = %v", err)
	}
}

func TestNumstatSkipsBinaryFilesAndNoise(t *testing.T) {
	if a, r := sumNumstat("3\t1\ta.go\n-\t-\timg.png\n\nnot a line\n2\t0\tb.go\n"); a != 5 || r != 1 {
		t.Errorf("+%d -%d, want +5 -1", a, r)
	}
}
