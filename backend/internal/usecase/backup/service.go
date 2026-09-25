// Package backup contains the Backup use cases: an Admin exporting a manual
// snapshot and the daily backup the scheduler drives. Everything it needs
// arrives as a port, so these cases run without SQLite, the filesystem or
// HTTP (ADR-0004, ADR-0007).
package backup

import (
	"context"
	"time"

	domainbackup "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/backup"
)

// Service holds the ports the Backup use cases need.
type Service struct {
	backups       domainbackup.BackupRepository
	retentionDays int
	now           func() time.Time
}

// NewService wires the Backup use cases to their port. `retentionDays` is how
// many days of backups to keep before the oldest are pruned (issue #10).
func NewService(backups domainbackup.BackupRepository, retentionDays int) *Service {
	return &Service{
		backups:       backups,
		retentionDays: retentionDays,
		now:           time.Now,
	}
}

// Create takes a snapshot of the database now, prunes the backups that fell out
// of the retention window, and answers the file it just created.
//
// It is the single entry point for both the manual export and the daily
// scheduler: the two differ only in who calls them, not in what a backup is.
func (s *Service) Create(ctx context.Context) (domainbackup.Backup, error) {
	snapshot, err := s.backups.Snapshot(ctx, s.now())
	if err != nil {
		return domainbackup.Backup{}, err
	}

	if err := s.prune(ctx); err != nil {
		return domainbackup.Backup{}, err
	}

	return snapshot, nil
}

// List answers the backup files on disk, oldest first, so an Admin can see
// what the store has kept.
func (s *Service) List(ctx context.Context) ([]domainbackup.Backup, error) {
	return s.backups.List(ctx)
}

// prune removes every backup older than the retention window. A failure is
// returned rather than swallowed: a snapshot that exists but cannot be cleaned
// up is still worth surfacing, so tomorrow's run (or the next manual export)
// can try again.
func (s *Service) prune(ctx context.Context) error {
	backups, err := s.backups.List(ctx)
	if err != nil {
		return err
	}

	now := s.now()
	for _, candidate := range backups {
		if domainbackup.Keep(candidate.CreatedAt, now, s.retentionDays) {
			continue
		}

		if err := s.backups.Delete(ctx, candidate.Name); err != nil {
			return err
		}
	}

	return nil
}
