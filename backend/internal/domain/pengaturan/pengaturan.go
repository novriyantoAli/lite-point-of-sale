// Package pengaturan holds the Pengaturan domain: the one row of store settings
// an Admin may change (CONTEXT.md, Pengaturan). It declares the port the use
// cases need without depending on SQLite or HTTP (ADR-0004).
//
// The Go identifiers are English (Settings, SettingsRepository) while the
// domain term and the route stay Indonesian (Pengaturan, /api/pengaturan), the
// boundary ADR-0012 records.
package pengaturan

import (
	"context"
	"errors"
)

// Settings is the store's single row of Pengaturan. One store, one terminal
// (ADR-0002) is what makes a fixed row honest rather than a key-value bag: the
// columns are typed, so the read path stays typed end to end.
type Settings struct {
	// ID is always 1: there is exactly one store, so there is exactly one row.
	ID int64
	// Header and Footer are the Struk template blocks: free text, newline-separated,
	// printed verbatim around the sale lines (CONTEXT.md, Struk).
	Header string
	Footer string
	// PaperWidth is the thermal roll width in millimetres — 58 or 80. The number
	// of printable columns is derived from it by the encoder, never stored as a
	// second value that could drift (ADR-0017, keputusan 4).
	PaperWidth int64
	// LowStockThreshold is the ambang below which an Active Produk counts as
	// Stok menipis. It moved here from domainproduk.LowStockThreshold, so there
	// is one stored source of truth instead of a constant (ADR-0014, ADR-0017).
	LowStockThreshold int64
}

// ValidPaperWidth reports whether width is one of the two thermal roll widths
// the encoder can derive a column count from (ADR-0017, keputusan 4).
func ValidPaperWidth(width int64) bool {
	return width == 58 || width == 80
}

// ValidLowStockThreshold reports whether threshold is an ambang the domain can
// mean: it has to be above zero, because an ambang of zero would call every
// Active Produk menipis and a negative one is not a Stok at all.
func ValidLowStockThreshold(threshold int64) bool {
	return threshold > 0
}

// Sentinel errors the use cases return and the HTTP adapter maps to status
// codes. Keep the set small and meaningful.
var (
	// ErrInvalidInput is the parent of an usecase InputError: a Pengaturan the
	// rules refuse (a paper width that is not 58/80, an ambang of zero or less).
	ErrInvalidInput = errors.New("invalid input")
)

// SettingsRepository is the outbound port for Pengaturan persistence.
// Implemented in the SQLite adapter; the use cases never see SQL (ADR-0004).
//
// The row is seeded by migration 0005, so there is deliberately no "not found"
// state a reader must handle: Get always answers the one stored row.
type SettingsRepository interface {
	// Get answers the store's single row of Pengaturan.
	Get(ctx context.Context) (Settings, error)
	// Update replaces the row and answers it as it now stands.
	Update(ctx context.Context, settings Settings) (Settings, error)
}
