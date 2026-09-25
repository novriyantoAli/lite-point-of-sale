package e2e

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// backupPayload is one snapshot as the API answers it. The filesystem path is
// never sent: the browser learns a backup exists and how big it is, never where
// on the server it lives.
type backupPayload struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
}

type backupEnvelope struct {
	Backup backupPayload `json:"backup"`
}

func TestBackupRequiresAnAdmin(t *testing.T) {
	baseURL, adminToken := newAdminToken(t)
	createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")
	kasirToken := logIn(t, baseURL, "kasir1", kasirPassword)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "list backups", method: http.MethodGet, path: "/api/backup"},
		{name: "export a backup", method: http.MethodPost, path: "/api/backup"},
	}

	for _, test := range tests {
		t.Run(test.name+" anonymously", func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, test.method, baseURL+test.path, "", nil, &failure)

			if status != http.StatusUnauthorized {
				t.Fatalf("got status %d, want %d", status, http.StatusUnauthorized)
			}
			if failure.Error != "invalid_token" {
				t.Errorf("got code %q, want %q", failure.Error, "invalid_token")
			}
		})

		t.Run(test.name+" as a Kasir", func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, test.method, baseURL+test.path, kasirToken, nil, &failure)

			if status != http.StatusForbidden {
				t.Fatalf("got status %d, want %d", status, http.StatusForbidden)
			}
			if failure.Error != "forbidden" {
				t.Errorf("got code %q, want %q", failure.Error, "forbidden")
			}
		})
	}
}

func TestBackupExportCreatesAFile(t *testing.T) {
	cfg := newTestConfig(t)
	_, baseURL := startAPI(t, cfg)
	adminToken := logIn(t, baseURL, seededAdmin, testAdminPassword)

	var created dataEnvelope[backupEnvelope]
	status := apiCall(t, http.MethodPost, baseURL+"/api/backup", adminToken, nil, &created)

	if status != http.StatusCreated {
		t.Fatalf("export: got status %d, want %d", status, http.StatusCreated)
	}
	if created.Data.Backup.Name == "" || created.Data.Backup.Size == 0 {
		t.Fatalf("export: got %+v, want a snapshot with a name and a size", created.Data.Backup)
	}

	// The acceptance criterion of #10: the file really exists on disk, and it is
	// a valid SQLite database — a consistent snapshot, not an empty placeholder.
	path := filepath.Join(cfg.BackupDir, created.Data.Backup.Name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read the snapshot file: %v", err)
	}
	if len(content) != int(created.Data.Backup.Size) {
		t.Errorf("snapshot size: the file has %d bytes, the API said %d", len(content), created.Data.Backup.Size)
	}
	if !strings.HasPrefix(string(content), "SQLite format 3\x00") {
		t.Error("snapshot: not a SQLite database (wrong magic header)")
	}

	// The same snapshot shows up in the list, so the Admin can see it was kept.
	var listed dataEnvelope[[]backupPayload]
	if status := apiCall(t, http.MethodGet, baseURL+"/api/backup", adminToken, nil, &listed); status != http.StatusOK {
		t.Fatalf("list: got status %d, want %d", status, http.StatusOK)
	}

	found := false
	for _, backup := range listed.Data {
		if backup.Name == created.Data.Backup.Name {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("list: got %+v, want the exported snapshot %s in it", listed.Data, created.Data.Backup.Name)
	}
}
