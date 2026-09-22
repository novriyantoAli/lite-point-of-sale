// Package httpapi adapts use cases to HTTP. It decides status codes and
// response shapes; it holds no business rules (ADR-0004).
package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/health"
)

// HealthChecker is the use case the health endpoint depends on.
type HealthChecker interface {
	Check(ctx context.Context) health.Report
}

// NewRouter returns the API router with every route the service exposes.
func NewRouter(healthChecker HealthChecker, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /api/health", healthHandler(healthChecker, logger))

	return mux
}

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func healthHandler(checker HealthChecker, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := checker.Check(r.Context())

		status := http.StatusOK
		if report.Status != health.StatusOK {
			status = http.StatusServiceUnavailable
		}

		writeJSON(w, status, healthResponse{
			Status:   string(report.Status),
			Database: string(report.Database),
		}, logger)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("write response", "error", err)
	}
}
