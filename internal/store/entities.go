package store

import (
	"context"
	"fmt"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

// CreateWorkspace stores w. A slug that is already taken gives domain.ErrConflict.
func (s *Store) CreateWorkspace(ctx context.Context, w domain.Workspace) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO workspaces (id, slug, name, created_at) VALUES (?, ?, ?, ?)`,
		w.ID, w.Slug, w.Name, w.CreatedAt.UnixMilli())
	if err == nil {
		s.publish(domain.Change{Workspace: &w})
	}
	return wrap("create workspace "+w.Slug, err)
}

// Workspace returns the workspace with the given slug.
func (s *Store) Workspace(ctx context.Context, slug string) (domain.Workspace, error) {
	var w domain.Workspace
	var created int64
	err := s.db.QueryRowContext(ctx, `SELECT id, slug, name, created_at FROM workspaces WHERE slug = ?`, slug).
		Scan(&w.ID, &w.Slug, &w.Name, &created)
	w.CreatedAt = domain.Timestamp{Time: time.UnixMilli(created)}
	return w, wrap("find workspace "+slug, err)
}

// Workspaces lists every workspace, oldest first.
func (s *Store) Workspaces(ctx context.Context) ([]domain.Workspace, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, slug, name, created_at FROM workspaces ORDER BY id`)
	if err != nil {
		return nil, wrap("list workspaces", err)
	}
	defer func() { _ = rows.Close() }()
	var out []domain.Workspace
	for rows.Next() {
		var w domain.Workspace
		var created int64
		if err := rows.Scan(&w.ID, &w.Slug, &w.Name, &created); err != nil {
			return nil, wrap("read workspace", err)
		}
		w.CreatedAt = domain.Timestamp{Time: time.UnixMilli(created)}
		out = append(out, w)
	}
	return out, wrap("list workspaces", rows.Err())
}

// CreateProject stores p. An unknown workspace or a repeated name gives domain.ErrConflict.
func (s *Store) CreateProject(ctx context.Context, p domain.Project) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO projects (`+projectColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.WorkspaceID, p.Name, p.RepoURL, p.DefaultBranch, p.Stack,
		p.Commands.Dev, p.Commands.Build, p.Commands.Test, p.CreatedAt.UnixMilli())
	if err == nil {
		s.publish(domain.Change{Project: &p})
	}
	return wrap("create project "+p.Name, err)
}

const projectColumns = `id, workspace_id, name, repo_url, default_branch, stack, dev_cmd, build_cmd, test_cmd, created_at`

type scanner interface{ Scan(dest ...any) error }

func scanProject(r scanner) (domain.Project, error) {
	var p domain.Project
	var created int64
	err := r.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.RepoURL, &p.DefaultBranch, &p.Stack,
		&p.Commands.Dev, &p.Commands.Build, &p.Commands.Test, &created)
	p.CreatedAt = domain.Timestamp{Time: time.UnixMilli(created)}
	return p, err
}

// SetCommands replaces the dev, build and test commands of a project; the user corrects
// what detection found (F2-03).
func (s *Store) SetCommands(ctx context.Context, id domain.ProjectID, c domain.Commands) error {
	res, err := s.db.ExecContext(ctx, `UPDATE projects SET dev_cmd = ?, build_cmd = ?, test_cmd = ? WHERE id = ?`,
		c.Dev, c.Build, c.Test, id)
	if err != nil {
		return wrap("set commands of project "+string(id), err)
	}
	if n, err := res.RowsAffected(); err != nil || n == 0 {
		return fmt.Errorf("set commands of project %s: %w", id, domain.ErrNotFound)
	}
	if p, err := s.Project(ctx, id); err == nil {
		s.publish(domain.Change{Project: &p})
	}
	return nil
}

// Project returns the project with the given id.
func (s *Store) Project(ctx context.Context, id domain.ProjectID) (domain.Project, error) {
	p, err := scanProject(s.db.QueryRowContext(ctx, `SELECT `+projectColumns+` FROM projects WHERE id = ?`, id))
	return p, wrap("find project "+string(id), err)
}

// Projects lists the projects of a workspace, oldest first.
func (s *Store) Projects(ctx context.Context, workspace domain.WorkspaceID) ([]domain.Project, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+projectColumns+` FROM projects WHERE workspace_id = ? ORDER BY id`, workspace)
	if err != nil {
		return nil, wrap("list projects", err)
	}
	defer func() { _ = rows.Close() }()
	var out []domain.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, wrap("read project", err)
		}
		out = append(out, p)
	}
	return out, wrap("list projects", rows.Err())
}

// WorkspaceOf returns the workspace a project belongs to; the daemon needs it to pick the
// agent user and folders for the project's workers.
func (s *Store) WorkspaceOf(ctx context.Context, project domain.ProjectID) (domain.Workspace, error) {
	var w domain.Workspace
	var created int64
	err := s.db.QueryRowContext(ctx,
		`SELECT w.id, w.slug, w.name, w.created_at FROM workspaces w JOIN projects p ON p.workspace_id = w.id WHERE p.id = ?`,
		project).Scan(&w.ID, &w.Slug, &w.Name, &created)
	w.CreatedAt = domain.Timestamp{Time: time.UnixMilli(created)}
	if err != nil {
		return w, wrap(fmt.Sprintf("find the workspace of project %s", project), err)
	}
	return w, nil
}
