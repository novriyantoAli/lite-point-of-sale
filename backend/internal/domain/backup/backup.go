// Package backup holds the Backup domain: a snapshot of the store's SQLite
// database, kept as a file in a backup folder. It declares the port the use
// cases need without depending on SQLite, the filesystem or HTTP (ADR-0004).
//
// The Go identifiers are English (Backup, BackupRepository) while the domain
// term and the route stay Indonesian (Backup, /api/backup) — the boundary
// ADR-0012 records.
package backup

import (
	"context"
	"time"
)

// Backup is one snapshot file of the store database. Name is the file name
// within the backup folder; Size is its length in bytes; CreatedAt is when it
// was taken.
type Backup struct {
	Name      string
	Size      int64
	CreatedAt time.Time
}

// Keep reports whether a backup taken at `at` is still within the retention
// window of `retentionDays` days ending at `now`. The boundary is inclusive in
// the sense that a backup exactly on the cutoff is kept; anything strictly
// older is pruned.
//
// A retention window of zero or less keeps nothing: there is no "keep forever"
// in this domain, only "keep the last N days" (the issue #10 requirement).
func Keep(at, now time.Time, retentionDays int) bool {
	if retentionDays <= 0 {
		return false
	}

	cutoff := now.Add(-time.Duration(retentionDays) * 24 * time.Hour)

	return !at.Before(cutoff)
}

// BackupRepository is the outbound port for backup storage: snapshot the live
// database into the backup folder, list what is there, and remove one file.
// Implemented in the adapter; the use cases never see filesystem or SQL
// (ADR-0004).
type BackupRepository interface {
	// Snapshot copies the live database into the backup folder and answers the
	// backup file it created. `at` is when the snapshot is taken, and the
	// implementation is what turns it into a file name.
	Snapshot(ctx context.Context, at time.Time) (Backup, error)
	// List answers the backup files on disk, oldest first.
	List(ctx context.Context) ([]Backup, error)
	// Delete removes one backup file by its name within the backup folder.
	Delete(ctx context.Context, name string) error
}
