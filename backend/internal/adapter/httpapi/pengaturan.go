package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	domainpengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/pengaturan"
	usecasepengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/pengaturan"
)

// SettingsService is the Pengaturan use cases the HTTP adapter depends on.
// Declaring it here, next to the handlers that use it, keeps this package
// testable with a fake service and keeps the dependency pointing inward
// (ADR-0004).
type SettingsService interface {
	Get(ctx context.Context) (domainpengaturan.Settings, error)
	Update(ctx context.Context, input usecasepengaturan.UpdateInput) (domainpengaturan.Settings, error)
}

// updateSettingsRequest is what an Admin puts to change the store's one row of
// Pengaturan. The fields are the three the form edits — the Struk template
// blocks, the paper width, and the ambang Stok menipis (ADR-0017).
type updateSettingsRequest struct {
	Header            string `json:"header"`
	Footer            string `json:"footer"`
	PaperWidth        int64  `json:"paper_width"`
	LowStockThreshold int64  `json:"low_stock_threshold"`
}

func (r updateSettingsRequest) input() usecasepengaturan.UpdateInput {
	return usecasepengaturan.UpdateInput{
		Header:            r.Header,
		Footer:            r.Footer,
		PaperWidth:        r.PaperWidth,
		LowStockThreshold: r.LowStockThreshold,
	}
}

// settingsResponse is the JSON view of the store's Pengaturan.
type settingsResponse struct {
	ID                int64  `json:"id"`
	Header            string `json:"header"`
	Footer            string `json:"footer"`
	PaperWidth        int64  `json:"paper_width"`
	LowStockThreshold int64  `json:"low_stock_threshold"`
}

func newSettingsResponse(settings domainpengaturan.Settings) settingsResponse {
	return settingsResponse{
		ID:                settings.ID,
		Header:            settings.Header,
		Footer:            settings.Footer,
		PaperWidth:        settings.PaperWidth,
		LowStockThreshold: settings.LowStockThreshold,
	}
}

// settingsEnvelope wraps the Pengaturan, the same way productEnvelope wraps a
// Produk.
type settingsEnvelope struct {
	Settings settingsResponse `json:"settings"`
}

// storeNameResponse is the public view of the store's name: the one value the
// login screen may read before a session exists. It is deliberately not the
// full settingsResponse — paper width and ambang stay Admin-only.
type storeNameResponse struct {
	StoreName string `json:"store_name"`
}

// storeNameHandler answers the store's name for screens shown before login. It
// is public by design: the router registers it without a guard, and it returns
// only the name derived from the Struk header (ADR-0019).
func storeNameHandler(service SettingsService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		settings, err := service.Get(r.Context())
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: storeNameResponse{
			StoreName: settings.StoreName(),
		}}, logger)
	}
}

// getSettingsHandler answers the store's Pengaturan. It is Admin-only: the
// router puts the role guard in front of it.
func getSettingsHandler(service SettingsService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		settings, err := service.Get(r.Context())
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: settingsEnvelope{
			Settings: newSettingsResponse(settings),
		}}, logger)
	}
}

// updateSettingsHandler replaces the store's Pengaturan. The rules live in the
// use case — a paper width that is not 58/80 and an ambang of zero or less are
// answered 400 with a message the form shows.
func updateSettingsHandler(service SettingsService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request updateSettingsRequest
		if !decodeJSON(w, r, &request, logger) {
			return
		}

		updated, err := service.Update(r.Context(), request.input())
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: settingsEnvelope{
			Settings: newSettingsResponse(updated),
		}}, logger)
	}
}
