package backup

import (
	"context"
	"fmt"
	"sort"
	"time"

	domainbackup "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/backup"
)

// fakeBackups is an in-memory BackupRepository. The use cases are tested
// against this instead of the filesystem: the port is the seam (ADR-0007).
type fakeBackups struct {
	backups map[string]domainbackup.Backup
	nextID  int64

	// Each method has its own error so a test can fail one step without failing
	// the others — e.g. prove a deletion that fails does not lose the snapshot.
	snapshotErr error
	listErr     error
	deleteErr   error
}

func newFakeBackups() *fakeBackups {
	return &fakeBackups{backups: map[string]domainbackup.Backup{}}
}

func (f *fakeBackups) Snapshot(_ context.Context, at time.Time) (domainbackup.Backup, error) {
	if f.snapshotErr != nil {
		return domainbackup.Backup{}, f.snapshotErr
	}

	f.nextID++
	snapshot := domainbackup.Backup{
		Name:      fmt.Sprintf("pos-%03d.db", f.nextID),
		Size:      10,
		CreatedAt: at,
	}
	f.backups[snapshot.Name] = snapshot

	return snapshot, nil
}

// List mirrors the adapter's ordering: oldest first.
func (f *fakeBackups) List(context.Context) ([]domainbackup.Backup, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}

	backups := make([]domainbackup.Backup, 0, len(f.backups))
	for _, candidate := range f.backups {
		backups = append(backups, candidate)
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.Before(backups[j].CreatedAt)
	})

	return backups, nil
}

func (f *fakeBackups) Delete(_ context.Context, name string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}

	delete(f.backups, name)

	return nil
}

// newTestService wires a Service over a fresh fake. Tests pin `service.now` to
// a fixed time so retention does not depend on the clock.
func newTestService(backups *fakeBackups, retentionDays int) *Service {
	return NewService(backups, retentionDays)
}
