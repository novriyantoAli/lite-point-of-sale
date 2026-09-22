package sqlite_test

import (
	"context"
	"path/filepath"
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
	first.Close()

	// Re-opening an existing database must not re-apply migrations.
	db, err := sqlite.Open(ctx, path)
	if err != nil {
		t.Fatalf("reopen migrated database: %v", err)
	}
	defer db.Close()

	var applied int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if applied != 1 {
		t.Errorf("applied migrations: got %d, want 1", applied)
	}

	var schemaVersion string
	if err := db.QueryRowContext(ctx, `SELECT value FROM app_meta WHERE key = 'schema_version'`).Scan(&schemaVersion); err != nil {
		t.Fatalf("read schema_version: %v", err)
	}
	if schemaVersion != "1" {
		t.Errorf("schema_version: got %q, want %q", schemaVersion, "1")
	}
}
