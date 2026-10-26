// Package store keeps the facts of workspaces, projects and workers in SQLite, through a
// pure Go driver so the daemon stays one static binary. It stores facts and never a status;
// status is derived from the facts when read (internal/domain).
//
// Schema changes are numbered files in migrations/, applied in order at Open; the database
// remembers how far it got in PRAGMA user_version.
package store
