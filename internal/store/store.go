package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/alfarax9/astra/internal/domain"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Store is the daemon's database.
type Store struct {
	db   *sql.DB
	now  func() time.Time
	subs listeners
}

// Open opens (creating if needed) the database at path and brings its schema up to date.
// Use ":memory:" for a throwaway database in tests.
func Open(ctx context.Context, path string) (*Store, error) {
	// Foreign keys are off by default in SQLite; WAL lets readers and the writer overlap.
	q := url.Values{}
	for _, p := range []string{"foreign_keys(1)", "busy_timeout(5000)", "journal_mode(WAL)", "synchronous(NORMAL)"} {
		q.Add("_pragma", p)
	}
	dsn := "file:" + path + "?" + q.Encode()
	if path == ":memory:" {
		dsn = "file::memory:?" + q.Encode()
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database %s: %w", path, err)
	}
	// One connection: every write is serialised and an in-memory database stays one database.
	// The queries are small, so nothing waits long.
	db.SetMaxOpenConns(1)
	s := &Store{db: db, now: time.Now}
	if err := s.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// migrate applies every migration newer than the database's user_version, each in its own
// transaction, so a failed migration leaves the previous version intact.
func (s *Store) migrate(ctx context.Context) error {
	entries, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(entries)
	var current int
	if err := s.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	for _, name := range entries {
		version, err := migrationVersion(name)
		if err != nil {
			return err
		}
		if version <= current {
			continue
		}
		body, err := migrations.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if err := s.apply(ctx, version, string(body)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
	}
	return nil
}

func (s *Store) apply(ctx context.Context, version int, body string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, body); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	// PRAGMA does not take parameters; version comes from an embedded file name.
	if _, err := tx.ExecContext(ctx, "PRAGMA user_version = "+strconv.Itoa(version)); err != nil {
		return fmt.Errorf("record version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// migrationVersion reads the number at the start of "migrations/001_init.sql".
func migrationVersion(name string) (int, error) {
	base := name[strings.LastIndex(name, "/")+1:]
	digits, _, _ := strings.Cut(base, "_")
	v, err := strconv.Atoi(digits)
	if err != nil || v < 1 {
		return 0, fmt.Errorf("migration %s must start with its version number", name)
	}
	return v, nil
}

// wrap turns SQLite's constraint and no-row errors into the domain's errors, so callers
// never see driver details.
func wrap(op string, err error) error {
	var se *sqlite.Error
	switch {
	case err == nil:
		return nil
	case errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("%s: %w", op, domain.ErrNotFound)
	case errors.As(err, &se) && isConstraint(se.Code()):
		return fmt.Errorf("%s: %w", op, domain.ErrConflict)
	}
	return fmt.Errorf("%s: %w", op, err)
}

func isConstraint(code int) bool {
	switch code {
	case sqlite3.SQLITE_CONSTRAINT, sqlite3.SQLITE_CONSTRAINT_UNIQUE, sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY,
		sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY, sqlite3.SQLITE_CONSTRAINT_NOTNULL, sqlite3.SQLITE_CONSTRAINT_CHECK:
		return true
	}
	return false
}
