package workspace

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/alfarax9/astra/internal/domain"
)

// diffScript prints "added removed path" for everything a worker changed since its branch left
// the default branch, committed or not, new files included. It counts in a copy of the index,
// so the worker's own staging is never touched. The default branch is the first argument.
const diffScript = `t=$(mktemp) || exit 1
trap 'rm -f "$t"' EXIT
cp "$(git rev-parse --git-path index)" "$t" || exit 1
export GIT_INDEX_FILE=$t
git add -A >/dev/null || exit 1
git diff --cached --numstat "$(git merge-base HEAD "$1")"`

// Refresh recounts the lines a worker has added and removed and stores them (the +x −y on its
// card, F4-01).
func (s *Service) Refresh(ctx context.Context, id domain.WorkerID) error {
	w, err := s.workers.Worker(ctx, id)
	if err != nil {
		return err
	}
	p, err := s.projects.Project(ctx, w.ProjectID)
	if err != nil {
		return err
	}
	ws, err := s.workspaces.WorkspaceOf(ctx, w.ProjectID)
	if err != nil {
		return err
	}
	out, err := s.run.Run(ctx, ws.Slug, w.Worktree, "sh", "-c", diffScript, "sh", s.baseRef(p))
	if err != nil {
		return fmt.Errorf("count the changes of worker %s: %w", id, err)
	}
	added, removed := sumNumstat(string(out))
	if added == w.Added && removed == w.Removed {
		return nil
	}
	_, err = s.workers.UpdateWorker(ctx, id, func(w *domain.Worker) { w.Added, w.Removed = added, removed })
	return err
}

// baseRef is what a worker's branch was made from: the remote's default branch for a cloned
// project, the local one for a repository started here.
func (s *Service) baseRef(p domain.Project) string {
	if p.RepoURL != "" {
		return "origin/" + p.DefaultBranch
	}
	return p.DefaultBranch
}

// sumNumstat adds up git's "added<TAB>removed<TAB>path" lines. Binary files show "-" and
// count as nothing.
func sumNumstat(out string) (added, removed int) {
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, "\t")
		if len(f) < 3 {
			continue
		}
		a, errA := strconv.Atoi(f[0])
		r, errR := strconv.Atoi(f[1])
		if errA == nil && errR == nil {
			added, removed = added+a, removed+r
		}
	}
	return added, removed
}
