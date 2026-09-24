package pengaturan

import (
	"context"

	domainpengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/pengaturan"
)

// fakeSettings is an in-memory SettingsRepository. The use cases are tested
// against this instead of SQLite: the port is the seam (ADR-0007).
type fakeSettings struct {
	stored domainpengaturan.Settings
	err    error
}

// newFakeSettings returns a fake seeded with the values migration 0005 writes,
// so the zero state of a test is the state a fresh store has.
func newFakeSettings() *fakeSettings {
	return &fakeSettings{stored: domainpengaturan.Settings{
		ID:                1,
		Header:            "",
		Footer:            "",
		PaperWidth:        80,
		LowStockThreshold: 5,
	}}
}

func (f *fakeSettings) Get(_ context.Context) (domainpengaturan.Settings, error) {
	if f.err != nil {
		return domainpengaturan.Settings{}, f.err
	}

	return f.stored, nil
}

func (f *fakeSettings) Update(_ context.Context, settings domainpengaturan.Settings) (domainpengaturan.Settings, error) {
	if f.err != nil {
		return domainpengaturan.Settings{}, f.err
	}

	settings.ID = 1
	f.stored = settings

	return settings, nil
}

func newTestService(settings *fakeSettings) *Service {
	return NewService(settings)
}
