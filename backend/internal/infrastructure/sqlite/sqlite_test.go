package sqlite_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/sqlite"
)

func TestOpenMigratesDatabase(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "pos.db")

	first, err := sqlite.Open(ctx, path)
	if err != nil {
		t.Fatalf("open fresh database: %v", err)
	}

	// A fresh database must come up migrated: at least one migration applied
	// and a numeric schema_version in app_meta. The exact count and value are
	// deliberately not asserted — every new migration would break the test.
	appliedFresh := countMigrations(t, ctx, first)
	if appliedFresh == 0 {
		t.Error("applied migrations after first open: got 0, want at least 1")
	}
	versionFresh := schemaVersion(t, ctx, first)
	if _, err := strconv.Atoi(versionFresh); err != nil {
		t.Errorf("schema_version after first open: got %q, want a number", versionFresh)
	}
	first.Close()

	// Re-opening an existing database must not re-apply migrations.
	db, err := sqlite.Open(ctx, path)
	if err != nil {
		t.Fatalf("reopen migrated database: %v", err)
	}
	defer db.Close()

	if applied := countMigrations(t, ctx, db); applied != appliedFresh {
		t.Errorf("applied migrations after reopen: got %d, want %d (unchanged)", applied, appliedFresh)
	}

	if version := schemaVersion(t, ctx, db); version != versionFresh {
		t.Errorf("schema_version after reopen: got %q, want %q (unchanged)", version, versionFresh)
	}
}

func countMigrations(t *testing.T, ctx context.Context, db *sql.DB) int {
	t.Helper()

	var applied int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	return applied
}

func schemaVersion(t *testing.T, ctx context.Context, db *sql.DB) string {
	t.Helper()

	var version string
	if err := db.QueryRowContext(ctx, `SELECT value FROM app_meta WHERE key = 'schema_version'`).Scan(&version); err != nil {
		t.Fatalf("read schema_version: %v", err)
	}
	return version
}
