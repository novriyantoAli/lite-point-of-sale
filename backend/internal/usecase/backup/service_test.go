package backup

import (
	"context"
	"errors"
	"testing"
	"time"

	domainbackup "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/backup"
)

// baseTime is the fixed "now" the tests pin the service to, so the retention
// window is a number on the page rather than a moving target.
var baseTime = time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)

func TestCreateSnapshotsTheDatabase(t *testing.T) {
	backups := newFakeBackups()
	service := newTestService(backups, 7)
	service.now = func() time.Time { return baseTime }

	created, err := service.Create(context.Background())
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if !created.CreatedAt.Equal(baseTime) {
		t.Errorf("created at: got %v, want %v", created.CreatedAt, baseTime)
	}

	stored, err := backups.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("stored backups: got %d, want 1", len(stored))
	}
}

func TestCreatePrunesBackupsOutsideTheWindow(t *testing.T) {
	backups := newFakeBackups()
	backups.backups["old.db"] = domainbackup.Backup{
		Name:      "old.db",
		Size:      10,
		CreatedAt: baseTime.Add(-8 * 24 * time.Hour),
	}
	backups.backups["recent.db"] = domainbackup.Backup{
		Name:      "recent.db",
		Size:      10,
		CreatedAt: baseTime.Add(-2 * 24 * time.Hour),
	}

	service := newTestService(backups, 7)
	service.now = func() time.Time { return baseTime }

	if _, err := service.Create(context.Background()); err != nil {
		t.Fatalf("create: %v", err)
	}

	stored, err := backups.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	names := map[string]bool{}
	for _, candidate := range stored {
		names[candidate.Name] = true
	}

	if names["old.db"] {
		t.Error("old.db: still stored, want it pruned as outside the window")
	}
	if !names["recent.db"] {
		t.Error("recent.db: missing, want it kept as inside the window")
	}
	// The snapshot Create just took is there too, alongside the kept one.
	if len(stored) != 2 {
		t.Errorf("stored backups: got %d, want 2 (recent + the new snapshot)", len(stored))
	}
}

func TestCreateKeepsEverythingWhenNothingIsOld(t *testing.T) {
	backups := newFakeBackups()
	backups.backups["yesterday.db"] = domainbackup.Backup{
		Name:      "yesterday.db",
		Size:      10,
		CreatedAt: baseTime.Add(-24 * time.Hour),
	}

	service := newTestService(backups, 7)
	service.now = func() time.Time { return baseTime }

	if _, err := service.Create(context.Background()); err != nil {
		t.Fatalf("create: %v", err)
	}

	stored, _ := backups.List(context.Background())
	if len(stored) != 2 {
		t.Errorf("stored backups: got %d, want 2", len(stored))
	}
}

func TestCreatePropagatesAPruneFailure(t *testing.T) {
	backups := newFakeBackups()
	backups.backups["stuck.db"] = domainbackup.Backup{
		Name:      "stuck.db",
		Size:      10,
		CreatedAt: baseTime.Add(-30 * 24 * time.Hour),
	}
	backups.deleteErr = errors.New("the file would not delete")

	service := newTestService(backups, 7)
	service.now = func() time.Time { return baseTime }

	if _, err := service.Create(context.Background()); err == nil {
		t.Fatal("create: got nil error, want the prune failure")
	}
}

func TestCreatePropagatesASnapshotFailure(t *testing.T) {
	backups := newFakeBackups()
	backups.snapshotErr = errors.New("disk full")

	service := newTestService(backups, 7)
	service.now = func() time.Time { return baseTime }

	if _, err := service.Create(context.Background()); err == nil {
		t.Fatal("create: got nil error, want the snapshot failure")
	}
}

func TestCreatePropagatesAListFailure(t *testing.T) {
	backups := newFakeBackups()
	backups.listErr = errors.New("cannot read the folder")

	service := newTestService(backups, 7)
	service.now = func() time.Time { return baseTime }

	if _, err := service.Create(context.Background()); err == nil {
		t.Fatal("create: got nil error, want the list failure")
	}
}

func TestListAnswersTheStoredBackupsOldestFirst(t *testing.T) {
	backups := newFakeBackups()
	backups.backups["new.db"] = domainbackup.Backup{Name: "new.db", CreatedAt: baseTime}
	backups.backups["old.db"] = domainbackup.Backup{Name: "old.db", CreatedAt: baseTime.Add(-48 * time.Hour)}
	backups.backups["older.db"] = domainbackup.Backup{Name: "older.db", CreatedAt: baseTime.Add(-96 * time.Hour)}

	service := newTestService(backups, 7)

	listed, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	want := []string{"older.db", "old.db", "new.db"}
	if len(listed) != len(want) {
		t.Fatalf("list: got %d backups, want %d", len(listed), len(want))
	}
	for i := range want {
		if listed[i].Name != want[i] {
			t.Errorf("list[%d]: got %q, want %q", i, listed[i].Name, want[i])
		}
	}
}
