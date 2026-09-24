package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"

	adaptersqlite "github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/sqlite"
	domainpengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/pengaturan"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/sqlite"
)

// newSettingsRepository opens a migrated SQLite file — which is what seeds the
// single Pengaturan row via migration 0005 — and returns a repository over it.
// This is the adapter integration test of ADR-0007: the repository runs against
// the real database, no fake.
func newSettingsRepository(t *testing.T) *adaptersqlite.SettingsRepository {
	t.Helper()

	db, err := sqlite.Open(context.Background(), filepath.Join(t.TempDir(), "pos.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return adaptersqlite.NewSettingsRepository(db)
}

func TestSettingsRepositoryReadsTheSeededRow(t *testing.T) {
	repository := newSettingsRepository(t)

	settings, err := repository.Get(context.Background())
	if err != nil {
		t.Fatalf("get Pengaturan: %v", err)
	}

	// Migration 0005 seeds the row with today's values (ADR-0017, keputusan 3).
	if settings.ID != 1 {
		t.Errorf("id: got %d, want 1", settings.ID)
	}
	if settings.PaperWidth != 80 {
		t.Errorf("paper width: got %d, want 80", settings.PaperWidth)
	}
	if settings.LowStockThreshold != 5 {
		t.Errorf("low stock threshold: got %d, want 5", settings.LowStockThreshold)
	}
	if settings.Header != "" || settings.Footer != "" {
		t.Errorf("header/footer: got %q/%q, want both empty", settings.Header, settings.Footer)
	}
}

func TestSettingsRepositoryWritesTheRow(t *testing.T) {
	repository := newSettingsRepository(t)
	ctx := context.Background()

	updated, err := repository.Update(ctx, domainpengaturan.Settings{
		ID:                1,
		Header:            "Toko Kopi",
		Footer:            "Terima kasih",
		PaperWidth:        58,
		LowStockThreshold: 3,
	})
	if err != nil {
		t.Fatalf("update Pengaturan: %v", err)
	}

	if updated.PaperWidth != 58 || updated.LowStockThreshold != 3 {
		t.Errorf("update Pengaturan: got %+v, want the written values", updated)
	}

	// The answer is the stored row: a later read agrees with what was written.
	readBack, err := repository.Get(ctx)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if readBack.Header != "Toko Kopi" || readBack.Footer != "Terima kasih" {
		t.Errorf("read back: got %+v, want the written template blocks", readBack)
	}
	if readBack.PaperWidth != 58 || readBack.LowStockThreshold != 3 {
		t.Errorf("read back: got %+v, want the written paper width and threshold", readBack)
	}
}
