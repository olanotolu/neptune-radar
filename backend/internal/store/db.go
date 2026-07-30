// Package store is the Postgres access layer. It owns schema migration and
// exposes narrow, entity-scoped read/write helpers used by the pipeline stages.
package store

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Store struct {
	DB *sql.DB
	// sourceClassCache is a per-tick snapshot of handle→class for active
	// watched_sources, populated by RefreshSourceClassCache. nil = not
	// populated (tests, API paths) → SourceClassForHandle falls back to DB.
	// ponytail: ceiling — a source added mid-tick won't appear until the next
	// RefreshSourceClassCache call; acceptable since the worker loads sources
	// at the start of each tick anyway.
	sourceClassCache map[string]string
	mu               sync.RWMutex
}

// Open connects to Postgres (dsn e.g. "postgres://user:pass@host:5432/neptune")
// and applies any pending migrations.
func Open(dsn string) (*Store, error) {
	if dsn == "" {
		return nil, errors.New("store: empty database DSN (set DATABASE_URL)")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{DB: db}, nil
}

func (s *Store) Close() error { return s.DB.Close() }

// migrate applies embedded migrations in filename order, recording each in
// schema_migrations. Migrations are plain PostgreSQL executed as one batch;
// they must be idempotent-safe (CREATE ... IF NOT EXISTS) only in the sense
// that already-applied versions are skipped — never edit an applied file.
func migrate(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return err
	}
	var files []string
	err := fs.WalkDir(migrationsFS, "migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".sql") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, f := range files {
		version := strings.TrimPrefix(f, "migrations/")
		var applied int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = $1`, version).Scan(&applied); err != nil {
			return err
		}
		if applied > 0 {
			continue
		}
		body, err := migrationsFS.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(body)); err != nil {
			return fmt.Errorf("apply %s: %w", version, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			return err
		}
	}
	return nil
}

// isUniqueViolation reports whether err is a Postgres unique-constraint
// violation (SQLSTATE 23505) — the idempotency signal for duplicate event
// delivery from the provider.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
