package domain

import "regexp"

// slugPattern keeps a workspace slug usable as part of a Linux user name
// (astra-agent-<slug>, at most 31 characters) and as a folder name (docs/adr/0005).
var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,18}[a-z0-9]$`)

// projectNamePattern keeps a project name usable as a folder name in the workspace.
var projectNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,62}$`)

// Workspace is one client's space: its projects, domains and agent user (F2-01).
type Workspace struct {
	ID        WorkspaceID `json:"id"`
	Slug      string      `json:"slug"`
	Name      string      `json:"name"`
	CreatedAt Timestamp   `json:"createdAt"`
}

// ValidSlug reports whether s may name a workspace.
func ValidSlug(s string) bool { return slugPattern.MatchString(s) }

// AgentUser is the Linux user that runs every agent of the workspace.
func (w Workspace) AgentUser() string { return "astra-agent-" + w.Slug }

// Commands are what a project runs to develop, build and test; detected from the stack and
// corrected by the user (F2-03).
type Commands struct {
	Dev   string `json:"dev,omitzero"`
	Build string `json:"build,omitzero"`
	Test  string `json:"test,omitzero"`
}

// Project is a repository inside a workspace.
type Project struct {
	ID            ProjectID   `json:"id"`
	WorkspaceID   WorkspaceID `json:"workspaceId"`
	Name          string      `json:"name"`
	RepoURL       string      `json:"repoUrl,omitzero"`
	DefaultBranch string      `json:"defaultBranch"`
	Stack         string      `json:"stack,omitzero"` // detected kind of app, such as "nextjs" (F2-03)
	Commands      Commands    `json:"commands"`
	CreatedAt     Timestamp   `json:"createdAt"`
}

// ValidProjectName reports whether s may name a project: lowercase letters, digits, dots,
// hyphens and underscores, starting with a letter or digit.
func ValidProjectName(s string) bool { return projectNamePattern.MatchString(s) }
