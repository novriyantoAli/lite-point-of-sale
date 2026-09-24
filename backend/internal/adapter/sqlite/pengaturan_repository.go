package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainpengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/pengaturan"
)

// SettingsRepository stores Pengaturan in SQLite. It satisfies
// domain/pengaturan.SettingsRepository, so no SQL leaves this package
// (ADR-0004).
type SettingsRepository struct {
	db *sql.DB
}

// NewSettingsRepository returns a repository backed by the given database
// handle.
func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

const settingsColumns = `id, header, footer, paper_width, low_stock_threshold`

// Get satisfies domain/pengaturan.SettingsRepository: the one row migration
// 0005 seeded. There is no "not found" state by construction (ADR-0017,
// keputusan 3), so a missing row is a corruption error rather than a value a
// caller has to handle.
func (r *SettingsRepository) Get(ctx context.Context) (domainpengaturan.Settings, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+settingsColumns+` FROM pengaturan WHERE id = 1`)

	settings, err := scanSettings(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domainpengaturan.Settings{}, fmt.Errorf("read Pengaturan: row is missing, migration 0005 did not seed it")
	}

	return settings, err
}

// Update satisfies domain/pengaturan.SettingsRepository. The statement names
// only the settable columns and answers the row as it now stands from the same
// statement that wrote it, so the caller does not need a second query to report
// the stored Pengaturan.
func (r *SettingsRepository) Update(ctx context.Context, settings domainpengaturan.Settings) (domainpengaturan.Settings, error) {
	row := r.db.QueryRowContext(ctx,
		`UPDATE pengaturan SET header = ?, footer = ?, paper_width = ?, low_stock_threshold = ? WHERE id = 1 RETURNING `+settingsColumns,
		settings.Header, settings.Footer, settings.PaperWidth, settings.LowStockThreshold)

	updated, err := scanSettings(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domainpengaturan.Settings{}, fmt.Errorf("update Pengaturan: row is missing, migration 0005 did not seed it")
	}

	return updated, err
}

func scanSettings(source row) (domainpengaturan.Settings, error) {
	var settings domainpengaturan.Settings

	err := source.Scan(
		&settings.ID, &settings.Header, &settings.Footer,
		&settings.PaperWidth, &settings.LowStockThreshold,
	)
	if err != nil {
		return domainpengaturan.Settings{}, fmt.Errorf("decode Pengaturan row: %w", err)
	}

	return settings, nil
}
