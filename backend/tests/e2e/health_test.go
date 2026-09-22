// Package e2e exercises the Go API at its HTTP seam: a real server, a real
// SQLite file, and a real client. Nothing below the HTTP boundary is touched
// (ADR-0007, e2e layer).
package e2e

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/app"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/config"
)

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func TestHealthEndpointReflectsDatabaseState(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pos.db")

	server, err := app.New(context.Background(), config.Config{HTTPAddr: ":0", DBPath: dbPath}, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("start app: %v", err)
	}
	// Second call is a no-op; keeps the test safe on early failure.
	defer server.Close()

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected SQLite file at %s after startup: %v", dbPath, err)
	}

	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	status, health := getHealth(t, httpServer.URL+"/api/health")
	if status != http.StatusOK {
		t.Fatalf("healthy database: got status %d, want %d", status, http.StatusOK)
	}
	if health.Status != "ok" || health.Database != "ok" {
		t.Fatalf("healthy database: got %+v, want {Status:ok Database:ok}", health)
	}

	// With the database gone the endpoint must stop claiming to be healthy —
	// proof the check reads SQLite instead of returning a constant.
	if err := server.Close(); err != nil {
		t.Fatalf("close app: %v", err)
	}

	status, health = getHealth(t, httpServer.URL+"/api/health")
	if status != http.StatusServiceUnavailable {
		t.Fatalf("unreachable database: got status %d, want %d", status, http.StatusServiceUnavailable)
	}
	if health.Status != "degraded" || health.Database != "unavailable" {
		t.Fatalf("unreachable database: got %+v, want {Status:degraded Database:unavailable}", health)
	}
}

func getHealth(t *testing.T, url string) (int, healthResponse) {
	t.Helper()

	response, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer response.Body.Close()

	var health healthResponse
	if err := json.NewDecoder(response.Body).Decode(&health); err != nil {
		t.Fatalf("decode %s response: %v", url, err)
	}
	return response.StatusCode, health
}
