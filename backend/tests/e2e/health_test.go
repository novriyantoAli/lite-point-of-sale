package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func TestHealthEndpointReflectsDatabaseState(t *testing.T) {
	cfg := newTestConfig(t)
	server, baseURL := startAPI(t, cfg)

	if _, err := os.Stat(cfg.DBPath); err != nil {
		t.Fatalf("expected SQLite file at %s after startup: %v", cfg.DBPath, err)
	}

	status, health := getHealth(t, baseURL+"/api/health")
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

	status, health = getHealth(t, baseURL+"/api/health")
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
