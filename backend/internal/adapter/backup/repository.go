// Package backup adapts the Backup domain to the filesystem and SQLite: a
// backup is a consistent copy of the live database, written into a backup
// folder as a file. It satisfies domain/backup.BackupRepository, so no
// filesystem or SQL detail leaves this package (ADR-0004).
package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	domainbackup "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/backup"
)

// Repository stores backup snapshots as files in a backup folder. The folder
// is one directory for the whole store (one store, one terminal, ADR-0002).
type Repository struct {
	db  *sql.DB
	dir string
}

// NewRepository returns a repository that snapshots the database behind `db`
// into `dir`.
func NewRepository(db *sql.DB, dir string) *Repository {
	return &Repository{db: db, dir: dir}
}

const (
	backupPrefix = "pos-"
	backupSuffix = ".db"
)

// Snapshot satisfies domain/backup.BackupRepository. It writes a consistent
// copy of the live database with SQLite's `VACUUM INTO`: the database runs in
// WAL mode, so copying the main file alone would miss whatever committed data
// the WAL has not yet checkpointed. `VACUUM INTO` folds that into one fresh,
// self-contained file, in one statement.
func (r *Repository) Snapshot(ctx context.Context, at time.Time) (domainbackup.Backup, error) {
	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return domainbackup.Backup{}, fmt.Errorf("create backup directory %s: %w", r.dir, err)
	}

	name := backupName(at)
	path := filepath.Join(r.dir, name)

	if _, err := r.db.ExecContext(ctx, `VACUUM INTO ?`, path); err != nil {
		return domainbackup.Backup{}, fmt.Errorf("snapshot database to %s: %w", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return domainbackup.Backup{}, fmt.Errorf("stat backup %s: %w", path, err)
	}

	return domainbackup.Backup{
		Name:      name,
		Size:      info.Size(),
		CreatedAt: info.ModTime(),
	}, nil
}

// List satisfies domain/backup.BackupRepository: the backup files on disk,
// oldest first. A folder that does not exist yet is an empty list, not an error
// — the store simply has no backups yet.
func (r *Repository) List(_ context.Context) ([]domainbackup.Backup, error) {
	entries, err := os.ReadDir(r.dir)
	if errors.Is(err, os.ErrNotExist) {
		return []domainbackup.Backup{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list backup directory %s: %w", r.dir, err)
	}

	backups := []domainbackup.Backup{}
	for _, entry := range entries {
		if entry.IsDir() || !isBackupFile(entry.Name()) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat backup %s: %w", entry.Name(), err)
		}

		backups = append(backups, domainbackup.Backup{
			Name:      entry.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	// CreatedAt is the file's modification time, and a backup file is never
	// modified after it is written — so this ordering is the order the snapshots
	// were taken, and the one the retention rule reads.
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.Before(backups[j].CreatedAt)
	})

	return backups, nil
}

// Delete satisfies domain/backup.BackupRepository. Only a plain file name is
// accepted — a name that escapes the backup folder is refused rather than
// deleted — and deleting a file that is already gone is not an error.
func (r *Repository) Delete(_ context.Context, name string) error {
	if name == "" || filepath.Base(name) != name {
		return fmt.Errorf("refuse to delete backup %q: not a plain file name", name)
	}

	path := filepath.Join(r.dir, name)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete backup %s: %w", name, err)
	}

	return nil
}

// backupName builds the file name of a snapshot taken at `at`. The UTC clock
// and the nanosecond field make the name unique and sortable, so a manual
// export and the daily backup in the same second do not collide.
func backupName(at time.Time) string {
	return fmt.Sprintf("%s%s-%09d%s", backupPrefix, at.UTC().Format("20060102-150405"), at.Nanosecond(), backupSuffix)
}

func isBackupFile(name string) bool {
	return strings.HasPrefix(name, backupPrefix) && strings.HasSuffix(name, backupSuffix)
}
