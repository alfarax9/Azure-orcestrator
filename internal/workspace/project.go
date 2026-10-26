package workspace

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/alfarax9/astra/internal/domain"
)

// repoURL accepts the ways a git remote is written: HTTPS, SSH (both forms) and local
// file URLs. Anything else, such as git's ext:: transport that runs a command, is refused.
var repoURL = regexp.MustCompile(`^(https://|ssh://|file:///|[A-Za-z0-9._-]+@[A-Za-z0-9.-]+:)[^\s\x00-\x1f]+$`)

// author signs the one commit Astra makes itself: the first, empty commit of a new repository.
var author = []string{"-c", "user.name=Astra", "-c", "user.email=astra@localhost"}

func repoDir(project string) string { return project + "/repo" }

// AddProject clones url into the workspace, or starts an empty repository when url is
// empty, then detects the stack and its commands (F2-02, F2-03). The clone runs on the
// server as the workspace's agent user.
func (s *Service) AddProject(ctx context.Context, slug, name, url string) (domain.Project, error) {
	if !domain.ValidProjectName(name) {
		return domain.Project{}, fmt.Errorf("%w: project name %q: use lowercase letters, digits, dots, hyphens and underscores", ErrInvalid, name)
	}
	if url != "" && !repoURL.MatchString(url) {
		return domain.Project{}, fmt.Errorf("%w: repository URL must use https://, ssh://, user@host: or file:///", ErrInvalid)
	}
	ws, err := s.workspaces.Workspace(ctx, slug)
	if err != nil {
		return domain.Project{}, err
	}
	p := domain.Project{ID: domain.NewProjectID(), WorkspaceID: ws.ID, Name: name, RepoURL: url,
		CreatedAt: domain.Timestamp{Time: s.now()}}
	if err := s.fetchRepo(ctx, slug, name, url); err != nil {
		return domain.Project{}, err
	}
	if err := s.describe(ctx, slug, &p); err != nil {
		return domain.Project{}, s.undoRepo(ctx, slug, name, err)
	}
	if err := s.projects.CreateProject(ctx, p); err != nil {
		return domain.Project{}, s.undoRepo(ctx, slug, name, err)
	}
	return p, nil
}

// fetchRepo clones or initialises the repository and makes sure it has a commit, since a
// worktree needs one to branch from.
func (s *Service) fetchRepo(ctx context.Context, slug, name, url string) error {
	// mkdir without -p fails when the folder exists: a name that is taken, even by a
	// project the store lost, is never cloned over or removed below.
	if _, err := s.run.Run(ctx, slug, "", "mkdir", "-p", "--", name); err != nil {
		return fmt.Errorf("create project folder: %w", err)
	}
	if _, err := s.run.Run(ctx, slug, "", "mkdir", "--", repoDir(name)); err != nil {
		return fmt.Errorf("project %s: %w", name, domain.ErrConflict)
	}
	cmd := []string{"git", "init", "--quiet", "-b", "main", repoDir(name)}
	if url != "" {
		cmd = []string{"git", "clone", "--quiet", "--", url, repoDir(name)}
	}
	if _, err := s.run.Run(ctx, slug, "", cmd...); err != nil {
		return s.undoRepo(ctx, slug, name, fmt.Errorf("%s %s: %w", cmd[1], url, err))
	}
	if _, err := s.git(ctx, slug, name, "rev-parse", "--quiet", "--verify", "HEAD"); err == nil {
		return nil
	}
	args := append(append([]string{}, author...), "commit", "--quiet", "--allow-empty", "-m", "Start the project")
	if _, err := s.git(ctx, slug, name, args...); err != nil {
		return s.undoRepo(ctx, slug, name, fmt.Errorf("make the first commit: %w", err))
	}
	return nil
}

// describe fills in the default branch and what detection finds in the repository.
func (s *Service) describe(ctx context.Context, slug string, p *domain.Project) error {
	out, err := s.git(ctx, slug, p.Name, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return fmt.Errorf("read the default branch: %w", err)
	}
	p.DefaultBranch = strings.TrimSpace(string(out))
	p.Stack, p.Commands, err = detect(gitTree{s: s, ctx: ctx, slug: slug, dir: repoDir(p.Name)})
	return err
}

// undoRepo removes the repository folder this call created, and returns cause.
func (s *Service) undoRepo(ctx context.Context, slug, name string, cause error) error {
	if _, err := s.run.Run(ctx, slug, "", "rm", "-rf", "--", repoDir(name)); err != nil {
		return errors.Join(cause, fmt.Errorf("remove the partial clone: %w", err))
	}
	return cause
}

// git runs git in the project's clone.
func (s *Service) git(ctx context.Context, slug, project string, args ...string) ([]byte, error) {
	return s.run.Run(ctx, slug, repoDir(project), append([]string{"git"}, args...)...)
}

// gitTree reads the files of a clone's current commit through git, so detection sees what
// is committed and the daemon still never opens the folder.
type gitTree struct {
	s    *Service
	ctx  context.Context //nolint:containedctx // lives only for one detect call
	slug string
	dir  string
}

func (t gitTree) List(dir string) ([]string, error) {
	out, err := t.s.run.Run(t.ctx, t.slug, t.dir, "git", "ls-tree", "--name-only", "HEAD:"+dir)
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(out)), nil
}

func (t gitTree) Read(path string) ([]byte, error) {
	return t.s.run.Run(t.ctx, t.slug, t.dir, "git", "show", "HEAD:"+path)
}
