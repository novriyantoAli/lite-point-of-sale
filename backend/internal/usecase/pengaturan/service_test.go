package pengaturan

import (
	"context"
	"errors"
	"testing"

	domainpengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/pengaturan"
)

func TestGetAnswersTheStoredSettings(t *testing.T) {
	settings := newFakeSettings()

	got, err := newTestService(settings).Get(context.Background())
	if err != nil {
		t.Fatalf("get Pengaturan: %v", err)
	}

	if got.PaperWidth != 80 || got.LowStockThreshold != 5 {
		t.Errorf("get Pengaturan: got %+v, want the seeded defaults", got)
	}
}

func TestUpdateReplacesTheSettings(t *testing.T) {
	settings := newFakeSettings()
	service := newTestService(settings)

	updated, err := service.Update(context.Background(), UpdateInput{
		Header:            "Toko Kopi\nJl. Melati 1",
		Footer:            "Terima kasih",
		PaperWidth:        58,
		LowStockThreshold: 3,
	})
	if err != nil {
		t.Fatalf("update Pengaturan: %v", err)
	}

	if updated.Header != "Toko Kopi\nJl. Melati 1" || updated.Footer != "Terima kasih" {
		t.Errorf("update Pengaturan: got %+v, want the new template blocks", updated)
	}
	if updated.PaperWidth != 58 {
		t.Errorf("update Pengaturan: got paper width %d, want 58", updated.PaperWidth)
	}
	if updated.LowStockThreshold != 3 {
		t.Errorf("update Pengaturan: got threshold %d, want 3", updated.LowStockThreshold)
	}

	// The answer is the stored row: a later Get reads the same values back.
	stored, _ := settings.Get(context.Background())
	if stored.LowStockThreshold != 3 {
		t.Errorf("stored Pengaturan: got threshold %d, want 3", stored.LowStockThreshold)
	}
}

func TestUpdateValidatesThePaperWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int64
	}{
		{name: "zero", width: 0},
		{name: "between the two widths", width: 60},
		{name: "above both widths", width: 81},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			settings := newFakeSettings()

			_, err := newTestService(settings).Update(context.Background(), UpdateInput{
				Header:            "x",
				Footer:            "y",
				PaperWidth:        test.width,
				LowStockThreshold: 5,
			})

			if !errors.Is(err, domainpengaturan.ErrInvalidInput) {
				t.Fatalf("update Pengaturan: got error %v, want %v", err, domainpengaturan.ErrInvalidInput)
			}

			var inputErr InputError
			if !errors.As(err, &inputErr) || inputErr.Message == "" {
				t.Errorf("update Pengaturan: got %v, want an InputError with a message", err)
			}

			// Refused means refused: the stored row is where it was.
			if stored, _ := settings.Get(context.Background()); stored.PaperWidth != 80 {
				t.Errorf("stored Pengaturan: got paper width %d, want 80", stored.PaperWidth)
			}
		})
	}
}

func TestUpdateValidatesTheThreshold(t *testing.T) {
	tests := []struct {
		name      string
		threshold int64
	}{
		{name: "zero", threshold: 0},
		{name: "negative", threshold: -5},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			settings := newFakeSettings()

			_, err := newTestService(settings).Update(context.Background(), UpdateInput{
				Header:            "x",
				Footer:            "y",
				PaperWidth:        80,
				LowStockThreshold: test.threshold,
			})

			if !errors.Is(err, domainpengaturan.ErrInvalidInput) {
				t.Fatalf("update Pengaturan: got error %v, want %v", err, domainpengaturan.ErrInvalidInput)
			}

			var inputErr InputError
			if !errors.As(err, &inputErr) || inputErr.Message == "" {
				t.Errorf("update Pengaturan: got %v, want an InputError with a message", err)
			}
		})
	}
}

func TestUpdatePropagatesARepositoryFailure(t *testing.T) {
	settings := newFakeSettings()
	settings.err = errors.New("database is gone")

	_, err := newTestService(settings).Update(context.Background(), UpdateInput{
		Header:            "x",
		Footer:            "y",
		PaperWidth:        80,
		LowStockThreshold: 5,
	})

	if err == nil {
		t.Fatal("update Pengaturan: got no error, want the repository failure")
	}
}

func TestLowStockThresholdAnswersTheStoredValue(t *testing.T) {
	settings := newFakeSettings()
	settings.stored.LowStockThreshold = 7

	threshold, err := newTestService(settings).LowStockThreshold(context.Background())
	if err != nil {
		t.Fatalf("low Stok threshold: %v", err)
	}

	if threshold != 7 {
		t.Errorf("low Stok threshold: got %d, want 7", threshold)
	}
}
