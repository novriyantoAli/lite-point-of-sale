package backup_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	adapterbackup "github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/backup"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/sqlite"

	_ "modernc.org/sqlite" // the snapshot is opened back up to prove it is a real database
)

// newRepository opens a migrated SQLite file and a backup folder next to it,
// returning the repository with both the folder and the raw handle — the
// adapter integration test of ADR-0007: real database, real filesystem, no
// fake.
func newRepository(t *testing.T) (*adapterbackup.Repository, string, *sql.DB, context.Context) {
	t.Helper()

	ctx := context.Background()

	db, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "pos.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	dir := filepath.Join(t.TempDir(), "backup")

	return adapterbackup.NewRepository(db, dir), dir, db, ctx
}

func TestRepositorySnapshotsAConsistentDatabase(t *testing.T) {
	repo, dir, db, ctx := newRepository(t)

	// A row written after the store opened proves the snapshot captures the
	// committed data the WAL holds, not just whatever is in the main file.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO app_meta (key, value) VALUES ('backup_test', 'present')`); err != nil {
		t.Fatalf("plant the row: %v", err)
	}

	at := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)
	snapshot, err := repo.Snapshot(ctx, at)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	if !strings.HasPrefix(snapshot.Name, "pos-") || !strings.HasSuffix(snapshot.Name, ".db") {
		t.Errorf("name: got %q, want a pos-*.db file", snapshot.Name)
	}
	if snapshot.Size == 0 {
		t.Error("size: got 0, want a non-empty snapshot")
	}

	path := filepath.Join(dir, snapshot.Name)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat the snapshot file: %v", err)
	}

	// The snapshot is itself a valid SQLite database holding the planted row.
	snapshotDB, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatalf("open the snapshot: %v", err)
	}
	defer snapshotDB.Close()

	var value string
	if err := snapshotDB.QueryRow(`SELECT value FROM app_meta WHERE key = 'backup_test'`).Scan(&value); err != nil {
		t.Fatalf("read the planted row back from the snapshot: %v", err)
	}
	if value != "present" {
		t.Errorf("planted row: got %q, want %q", value, "present")
	}
}

func TestRepositoryListsOldestFirstAndDeletes(t *testing.T) {
	repo, dir, _, ctx := newRepository(t)

	// A folder that does not exist yet is an empty list, not an error.
	listed, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list a missing folder: %v", err)
	}
	if len(listed) != 0 {
		t.Errorf("list a missing folder: got %d backups, want 0", len(listed))
	}

	first, err := repo.Snapshot(ctx, time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("first snapshot: %v", err)
	}

	// Age the first file so the ordering does not depend on two modtimes that
	// might fall in the same instant.
	old := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(filepath.Join(dir, first.Name), old, old); err != nil {
		t.Fatalf("age the first snapshot: %v", err)
	}

	second, err := repo.Snapshot(ctx, time.Date(2026, time.January, 16, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("second snapshot: %v", err)
	}

	listed, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("list: got %d backups, want 2", len(listed))
	}
	if listed[0].Name != first.Name || listed[1].Name != second.Name {
		t.Errorf("list order: got [%s, %s], want [%s, %s]", listed[0].Name, listed[1].Name, first.Name, second.Name)
	}

	if err := repo.Delete(ctx, second.Name); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, second.Name)); !os.IsNotExist(err) {
		t.Errorf("deleted snapshot: still on disk (stat err %v), want it gone", err)
	}

	// Deleting an already-deleted backup is not an error: retention may race a
	// manual export of the same window.
	if err := repo.Delete(ctx, second.Name); err != nil {
		t.Errorf("delete again: got %v, want nil (idempotent)", err)
	}
}

func TestRepositoryRefusesToDeleteOutsideTheBackupFolder(t *testing.T) {
	repo, _, _, ctx := newRepository(t)

	for _, name := range []string{"", "../pos.db", "sub/pos.db", "/etc/passwd"} {
		if err := repo.Delete(ctx, name); err == nil {
			t.Errorf("delete %q: got nil error, want a refusal", name)
		}
	}
}
