package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

const workerColumns = `id, project_id, title, agent, branch, worktree, client_review, created_at,
	activity, preview_live, shared, new_feedback, gate_state, gate_cause, deploy_state, deploy_reason, added, removed`

func scanWorker(r scanner) (domain.Worker, error) {
	var w domain.Worker
	var created int64
	err := r.Scan(&w.ID, &w.ProjectID, &w.Title, &w.Agent, &w.Branch, &w.Worktree, &w.ClientReview, &created,
		&w.Activity, &w.PreviewLive, &w.Shared, &w.NewFeedback, &w.Gate.State, &w.Gate.Cause,
		&w.Deployment.State, &w.Deployment.Reason, &w.Added, &w.Removed)
	w.CreatedAt = domain.Timestamp{Time: time.UnixMilli(created)}
	return w, err
}

// CreateWorker stores w. An unknown project gives domain.ErrConflict.
func (s *Store) CreateWorker(ctx context.Context, w domain.Worker) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO workers (`+workerColumns+`, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		w.ID, w.ProjectID, w.Title, w.Agent, w.Branch, w.Worktree, w.ClientReview, w.CreatedAt.UnixMilli(),
		w.Activity, w.PreviewLive, w.Shared, w.NewFeedback, w.Gate.State, w.Gate.Cause,
		w.Deployment.State, w.Deployment.Reason, w.Added, w.Removed, s.now().UnixMilli())
	return wrap("create worker "+string(w.ID), err)
}

// Worker returns the worker with the given id.
func (s *Store) Worker(ctx context.Context, id domain.WorkerID) (domain.Worker, error) {
	w, err := scanWorker(s.db.QueryRowContext(ctx, `SELECT `+workerColumns+` FROM workers WHERE id = ?`, id))
	return w, wrap("find worker "+string(id), err)
}

// Workers lists the workers of a workspace, newest first. Pass an empty slug to list every
// worker on the server.
func (s *Store) Workers(ctx context.Context, workspaceSlug string) ([]domain.Worker, error) {
	query := `SELECT ` + qualified("w", workerColumns) + ` FROM workers w
		JOIN projects p ON p.id = w.project_id JOIN workspaces ws ON ws.id = p.workspace_id`
	var args []any
	if workspaceSlug != "" {
		query += ` WHERE ws.slug = ?`
		args = append(args, workspaceSlug)
	}
	rows, err := s.db.QueryContext(ctx, query+` ORDER BY w.id DESC`, args...)
	if err != nil {
		return nil, wrap("list workers", err)
	}
	defer func() { _ = rows.Close() }()
	var out []domain.Worker
	for rows.Next() {
		w, err := scanWorker(rows)
		if err != nil {
			return nil, wrap("read worker", err)
		}
		out = append(out, w)
	}
	return out, wrap("list workers", rows.Err())
}

// UpdateWorker reads the worker, lets change adjust its facts, and writes it back in one
// transaction, so two updates never overwrite each other's facts. Identity fields (id,
// project, agent, creation time) cannot be changed. It returns the stored result.
func (s *Store) UpdateWorker(ctx context.Context, id domain.WorkerID, change func(*domain.Worker)) (domain.Worker, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return domain.Worker{}, wrap("begin update", err)
	}
	defer func() { _ = tx.Rollback() }()

	w, err := scanWorker(tx.QueryRowContext(ctx, `SELECT `+workerColumns+` FROM workers WHERE id = ?`, id))
	if err != nil {
		return w, wrap("find worker "+string(id), err)
	}
	original := w
	change(&w)
	w.ID, w.ProjectID, w.Agent, w.CreatedAt = original.ID, original.ProjectID, original.Agent, original.CreatedAt

	_, err = tx.ExecContext(ctx, `UPDATE workers SET title = ?, branch = ?, worktree = ?, client_review = ?,
		activity = ?, preview_live = ?, shared = ?, new_feedback = ?, gate_state = ?, gate_cause = ?,
		deploy_state = ?, deploy_reason = ?, added = ?, removed = ?, updated_at = ? WHERE id = ?`,
		w.Title, w.Branch, w.Worktree, w.ClientReview, w.Activity, w.PreviewLive, w.Shared, w.NewFeedback,
		w.Gate.State, w.Gate.Cause, w.Deployment.State, w.Deployment.Reason, w.Added, w.Removed,
		s.now().UnixMilli(), id)
	if err != nil {
		return w, wrap("update worker "+string(id), err)
	}
	if err := tx.Commit(); err != nil {
		return w, fmt.Errorf("commit update of worker %s: %w", id, err)
	}
	return w, nil
}

// qualified prefixes every column in a comma separated list with a table alias.
func qualified(alias, columns string) string {
	parts := strings.Split(columns, ",")
	for i, c := range parts {
		parts[i] = alias + "." + strings.TrimSpace(c)
	}
	return strings.Join(parts, ", ")
}
