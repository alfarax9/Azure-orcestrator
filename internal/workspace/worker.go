package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/alfarax9/astra/internal/domain"
)

const maxTitle = 200

// CreateWorker starts a worker on its own branch in its own worktree, made from the latest
// default branch, so two workers never write to the same files (F3-01).
func (s *Service) CreateWorker(ctx context.Context, project domain.ProjectID, title string, agent domain.AgentKind, clientReview bool) (domain.Worker, error) {
	title = strings.TrimSpace(title)
	if title == "" || len(title) > maxTitle {
		return domain.Worker{}, fmt.Errorf("%w: a task needs a title of at most %d characters", ErrInvalid, maxTitle)
	}
	if agent != domain.AgentClaude && agent != domain.AgentCodex {
		return domain.Worker{}, fmt.Errorf("%w: unknown agent %q", ErrInvalid, agent)
	}
	p, err := s.projects.Project(ctx, project)
	if err != nil {
		return domain.Worker{}, err
	}
	ws, err := s.workspaces.WorkspaceOf(ctx, project)
	if err != nil {
		return domain.Worker{}, err
	}
	w := domain.NewWorker(p.ID, title, agent, "", clientReview, s.now())
	w.Branch = branchName(title, w.ID)
	w.Worktree = p.Name + "/workers/" + string(w.ID)

	base := s.baseRef(p)
	if p.RepoURL != "" {
		if _, err := s.git(ctx, ws.Slug, p.Name, "fetch", "--quiet", "origin", p.DefaultBranch); err != nil {
			return domain.Worker{}, fmt.Errorf("update %s from origin: %w", p.DefaultBranch, err)
		}
	}
	// --no-track: the worker's branch must never push to the default branch by accident.
	if _, err := s.git(ctx, ws.Slug, p.Name, "worktree", "add", "--quiet", "--no-track", "-b", w.Branch,
		"../workers/"+string(w.ID), base); err != nil {
		return domain.Worker{}, fmt.Errorf("create the worktree: %w", err)
	}
	if err := s.workers.CreateWorker(ctx, w); err != nil {
		return domain.Worker{}, s.undoWorktree(ctx, ws.Slug, p.Name, w, err)
	}
	return w, nil
}

func (s *Service) undoWorktree(ctx context.Context, slug, project string, w domain.Worker, cause error) error {
	_, rmErr := s.git(ctx, slug, project, "worktree", "remove", "--force", "../workers/"+string(w.ID))
	_, brErr := s.git(ctx, slug, project, "branch", "--quiet", "-D", w.Branch)
	return errors.Join(cause, rmErr, brErr)
}

// branchName is astra/<a few words of the title>-<the end of the worker id>, which keeps
// branches readable in git tools and unique per worker.
func branchName(title string, id domain.WorkerID) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(title) {
		if r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
		if b.Len() >= 40 {
			break
		}
	}
	words := b.String()
	if b.Len() >= 40 {
		if i := strings.LastIndexByte(words, '-'); i > 0 {
			words = words[:i] // never end in half a word
		}
	}
	words = strings.Trim(words, "-")
	if words == "" {
		words = "task"
	}
	return "astra/" + words + "-" + string(id[len(id)-6:])
}
