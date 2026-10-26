-- Facts only: the daemon stores what it observed, never a status (PRD section 7).
CREATE TABLE workspaces (
    id         TEXT PRIMARY KEY,
    slug       TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL,
    created_at INTEGER NOT NULL
) STRICT;

CREATE TABLE projects (
    id             TEXT PRIMARY KEY,
    workspace_id   TEXT NOT NULL REFERENCES workspaces(id),
    name           TEXT NOT NULL,
    repo_url       TEXT NOT NULL DEFAULT '',
    default_branch TEXT NOT NULL DEFAULT 'main',
    dev_cmd        TEXT NOT NULL DEFAULT '',
    build_cmd      TEXT NOT NULL DEFAULT '',
    test_cmd       TEXT NOT NULL DEFAULT '',
    created_at     INTEGER NOT NULL,
    UNIQUE (workspace_id, name)
) STRICT;

CREATE TABLE workers (
    id            TEXT PRIMARY KEY,
    project_id    TEXT NOT NULL REFERENCES projects(id),
    title         TEXT NOT NULL,
    agent         TEXT NOT NULL,
    branch        TEXT NOT NULL,
    worktree      TEXT NOT NULL DEFAULT '',
    client_review INTEGER NOT NULL DEFAULT 1,
    created_at    INTEGER NOT NULL,
    updated_at    INTEGER NOT NULL,
    activity      TEXT NOT NULL DEFAULT 'idle',
    preview_live  INTEGER NOT NULL DEFAULT 0,
    shared        INTEGER NOT NULL DEFAULT 0,
    new_feedback  INTEGER NOT NULL DEFAULT 0,
    gate_state    TEXT NOT NULL DEFAULT 'none',
    gate_cause    TEXT NOT NULL DEFAULT '',
    deploy_state  TEXT NOT NULL DEFAULT 'none',
    deploy_reason TEXT NOT NULL DEFAULT '',
    added         INTEGER NOT NULL DEFAULT 0,
    removed       INTEGER NOT NULL DEFAULT 0
) STRICT;

CREATE INDEX workers_by_project ON workers (project_id, created_at DESC);
