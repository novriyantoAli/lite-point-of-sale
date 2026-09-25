package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	domainbackup "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/backup"
)

// BackupService is the Backup use cases the HTTP adapter depends on. Declaring
// it here, next to the handlers that use it, keeps this package testable with a
// fake service and keeps the dependency pointing inward (ADR-0004).
type BackupService interface {
	Create(ctx context.Context) (domainbackup.Backup, error)
	List(ctx context.Context) ([]domainbackup.Backup, error)
}

// backupResponse is the JSON view of one snapshot. The filesystem path is
// deliberately absent: the browser needs to know a backup exists and how big it
// is, never where on the server the file lives.
type backupResponse struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
}

func newBackupResponse(backup domainbackup.Backup) backupResponse {
	return backupResponse{
		Name:      backup.Name,
		Size:      backup.Size,
		CreatedAt: backup.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func newBackupResponses(backups []domainbackup.Backup) []backupResponse {
	responses := make([]backupResponse, 0, len(backups))
	for _, backup := range backups {
		responses = append(responses, newBackupResponse(backup))
	}

	return responses
}

// backupEnvelope wraps one snapshot: the answer to a manual export.
type backupEnvelope struct {
	Backup backupResponse `json:"backup"`
}

// createBackupHandler takes a manual snapshot of the database now. It is
// Admin-only: the router puts the role guard in front of it.
func createBackupHandler(service BackupService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		created, err := service.Create(r.Context())
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusCreated, dataResponse{Data: backupEnvelope{
			Backup: newBackupResponse(created),
		}}, logger)
	}
}

// listBackupHandler answers the backup files on disk, oldest first, so the
// Admin sees what the store has kept. It is Admin-only, like the export.
func listBackupHandler(service BackupService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backups, err := service.List(r.Context())
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: newBackupResponses(backups)}, logger)
	}
}
