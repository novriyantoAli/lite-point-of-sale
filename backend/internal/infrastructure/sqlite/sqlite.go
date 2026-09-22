// Package sqlite owns the SQLite connection and the embedded schema
// migrations. It is infrastructure: the only place that knows how the
// database file is opened (ADR-0004, ADR-0005).
package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite" // pure-Go driver, no cgo (ADR-0004)
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

const migrationsDir = "migrations"

// Open creates (if needed) and opens the SQLite database at path, then brings
// its schema up to date.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create database directory %s: %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite", dataSourceName(path))
	if err != nil {
		return nil, fmt.Errorf("open database %s: %w", path, err)
	}

	// One store, one terminal (ADR-0002): the API is the single writer, so a
	// single connection removes SQLITE_BUSY entirely.
	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database %s: %w", path, err)
	}

	if err := Migrate(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate database %s: %w", path, err)
	}

	return db, nil
}

// Migrate applies every embedded migration that has not been applied yet, in
// file-name order, each in its own transaction.
func Migrate(ctx context.Context, db *sql.DB) error {
	const createMigrationsTable = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`

	if _, err := db.ExecContext(ctx, createMigrationsTable); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	names, err := migrationNames()
	if err != nil {
		return err
	}

	for _, name := range names {
		applied, err := isApplied(ctx, db, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		statements, err := migrationFiles.ReadFile(migrationsDir + "/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}

		if _, err := tx.ExecContext(ctx, string(statements)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, name); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	return nil
}

func migrationNames() ([]string, error) {
	entries, err := fs.ReadDir(migrationFiles, migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	return names, nil
}

func isApplied(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, name).Scan(&count); err != nil {
		return false, fmt.Errorf("look up migration %s: %w", name, err)
	}
	return count > 0, nil
}

// dataSourceName adds the pragmas the app relies on: wait instead of failing
// when the write lock is held, enforce foreign keys, and use WAL so reads do
// not block the writer.
func dataSourceName(path string) string {
	pragmas := url.Values{}
	pragmas.Add("_pragma", "busy_timeout(5000)")
	pragmas.Add("_pragma", "journal_mode(WAL)")
	pragmas.Add("_pragma", "foreign_keys(1)")

	return "file:" + path + "?" + pragmas.Encode()
}
