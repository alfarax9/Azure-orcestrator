package domain

import (
	"errors"
	"regexp"
	"uuid"
)

// Identifiers are time-ordered UUIDs (version 7): they sort by creation, so listings come
// out in a stable order without an extra column.
type (
	// WorkspaceID identifies a client workspace.
	WorkspaceID string
	// ProjectID identifies a project inside a workspace.
	ProjectID string
	// WorkerID identifies a worker: one task, one agent, one worktree.
	WorkerID string
)

// NewWorkspaceID returns a new workspace identifier.
func NewWorkspaceID() WorkspaceID { return WorkspaceID(uuid.NewV7().String()) }

// NewProjectID returns a new project identifier.
func NewProjectID() ProjectID { return ProjectID(uuid.NewV7().String()) }

// NewWorkerID returns a new worker identifier. It is also the name of the worker's agent
// session, so it must stay lowercase letters, digits and hyphens.
func NewWorkerID() WorkerID { return WorkerID(uuid.NewV7().String()) }

var sessionIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)

// ValidSessionID reports whether s may name an agent session. It becomes part of a socket
// path and a systemd unit name, so it stays lowercase letters, digits and hyphens.
func ValidSessionID(s string) bool { return sessionIDPattern.MatchString(s) }

// Errors callers branch on: the API answers 404 and 409 for them.
var (
	// ErrNotFound means no such workspace, project or worker.
	ErrNotFound = errors.New("not found")
	// ErrConflict means the name or slug is already taken, or a parent does not exist.
	ErrConflict = errors.New("conflict")
)
