// Package pengaturan contains the Pengaturan use cases: an Admin reading and
// changing the store's one row of settings. Everything it needs arrives as a
// port, so these cases run without SQLite or HTTP (ADR-0004, ADR-0007).
package pengaturan

import (
	"context"

	domainpengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/pengaturan"
)

// UpdateInput is what the Pengaturan form fills in: the Struk template blocks,
// the paper width, and the ambang Stok menipis.
type UpdateInput struct {
	Header            string
	Footer            string
	PaperWidth        int64
	LowStockThreshold int64
}

// InputError is a validation failure that carries a message fit for the API
// response. It unwraps to domainpengaturan.ErrInvalidInput, so the HTTP adapter
// only needs errors.Is to pick the status code and errors.As to read the
// message.
type InputError struct {
	Message string
}

func (e InputError) Error() string { return e.Message }

// InputMessage is what the HTTP adapter reads. Each usecase package declares
// its own InputError; this method is the shape they share.
func (e InputError) InputMessage() string { return e.Message }

// Unwrap makes errors.Is(err, domainpengaturan.ErrInvalidInput) true.
func (e InputError) Unwrap() error { return domainpengaturan.ErrInvalidInput }

// Service holds the ports the Pengaturan use cases need.
type Service struct {
	settings domainpengaturan.SettingsRepository
}

// NewService wires the Pengaturan use cases to their port.
func NewService(settings domainpengaturan.SettingsRepository) *Service {
	return &Service{settings: settings}
}

// Get answers the store's single row of Pengaturan as it is stored.
func (s *Service) Get(ctx context.Context) (domainpengaturan.Settings, error) {
	return s.settings.Get(ctx)
}

// Update replaces the store's Pengaturan after the rules have accepted it, and
// answers the stored row — the same one a later Get reads.
func (s *Service) Update(ctx context.Context, input UpdateInput) (domainpengaturan.Settings, error) {
	if !domainpengaturan.ValidPaperWidth(input.PaperWidth) {
		return domainpengaturan.Settings{}, InputError{Message: "Lebar kertas harus 58 atau 80 mm."}
	}
	if !domainpengaturan.ValidLowStockThreshold(input.LowStockThreshold) {
		return domainpengaturan.Settings{}, InputError{Message: "Ambang Stok menipis harus lebih dari nol."}
	}

	return s.settings.Update(ctx, domainpengaturan.Settings{
		Header:            input.Header,
		Footer:            input.Footer,
		PaperWidth:        input.PaperWidth,
		LowStockThreshold: input.LowStockThreshold,
	})
}

// LowStockThreshold answers the ambang Stok menipis as it is stored. It is the
// one-method port `usecase/produk.LowStockSettings` needs, satisfied here so
// the Produk use case reads the setting without importing this domain
// (ADR-0017, keputusan 3).
func (s *Service) LowStockThreshold(ctx context.Context) (int64, error) {
	settings, err := s.settings.Get(ctx)
	if err != nil {
		return 0, err
	}

	return settings.LowStockThreshold, nil
}
