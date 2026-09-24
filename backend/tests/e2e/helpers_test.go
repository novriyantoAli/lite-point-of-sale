package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/app"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/config"
)

// seededAdmin is the username of the Admin Pengguna the service creates on
// first start, and testAdminPassword its password. The password is not the
// development default on purpose: the seed path that matters here is the real
// one, not the one that warns.
const (
	seededAdmin       = "admin"
	testAdminPassword = "rahasia-admin"
)

// testSecret is the token secret of the test run. Explicit rather than the
// development default, so a broken POS_TOKEN_SECRET path cannot pass unnoticed.
const testSecret = "rahasia-uji-e2e"

// newTestConfig is config.Default() with what a test has to decide itself: an
// ephemeral database, an ephemeral port and credentials of its own.
//
// It also points POS_PRINTER_DEVICE at a file of its own. That file stands in for
// the thermal printer, so the e2e suite exercises the *success* path of a print —
// the bytes really reach a device — instead of only ever testing the failure
// (ADR-0017). A test that wants no printer clears PrinterDevice.
func newTestConfig(t *testing.T) config.Config {
	t.Helper()

	dir := t.TempDir()

	cfg := config.Default()
	cfg.HTTPAddr = ":0"
	cfg.DBPath = filepath.Join(dir, "pos.db")
	cfg.TokenSecret = testSecret
	cfg.SessionTTL = time.Hour
	cfg.AdminUsername = seededAdmin
	cfg.AdminPassword = testAdminPassword

	printer := filepath.Join(dir, "printer.bin")
	if err := os.WriteFile(printer, nil, 0o600); err != nil {
		t.Fatalf("create the printer file: %v", err)
	}
	cfg.PrinterDevice = printer

	return cfg
}

// startAPI boots the real service — SQLite file, bcrypt, HMAC tokens, the HTTP
// router — and returns it with its base URL. Nothing is faked: this is the HTTP
// seam of ADR-0007.
func startAPI(t *testing.T, cfg config.Config) (*app.App, string) {
	t.Helper()

	service, err := app.New(context.Background(), cfg, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("start app: %v", err)
	}
	t.Cleanup(func() { service.Close() })

	httpServer := httptest.NewServer(service.Handler())
	t.Cleanup(httpServer.Close)

	return service, httpServer.URL
}

// apiCall performs one API call and decodes the JSON body into out when out is
// not nil. An empty token leaves the Authorization header off, which is how the
// anonymous path is exercised.
func apiCall(t *testing.T, method, url, token string, payload, out any) int {
	t.Helper()

	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("encode %s %s body: %v", method, url, err)
		}
		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("build %s %s: %v", method, url, err)
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer response.Body.Close()

	if out != nil {
		if err := json.NewDecoder(response.Body).Decode(out); err != nil {
			t.Fatalf("decode %s %s response: %v", method, url, err)
		}
	}

	return response.StatusCode
}

// logIn returns the token of a Pengguna that can log in, failing the test
// otherwise. It avoids a second, unrelated failure cascading from one broken
// login.
func logIn(t *testing.T, baseURL, username, password string) string {
	t.Helper()

	var session dataEnvelope[sessionPayload]
	status := apiCall(t, http.MethodPost, baseURL+"/api/auth/login", "",
		loginPayload{Username: username, Password: password}, &session)

	if status != http.StatusOK {
		t.Fatalf("login as %s: got status %d, want %d", username, status, http.StatusOK)
	}
	if session.Data.Token == "" {
		t.Fatalf("login as %s: got no token", username)
	}

	return session.Data.Token
}

// apiCallRaw is apiCall for the rare test that has to read the response as it
// arrived — checking that something is *absent* needs the raw bytes.
func apiCallRaw(t *testing.T, method, url, token string) (int, string) {
	t.Helper()

	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("build %s %s: %v", method, url, err)
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read %s %s response: %v", method, url, err)
	}

	return response.StatusCode, string(body)
}

// itoa renders an id for a URL path.
func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}

// dataEnvelope mirrors the `data` envelope of every successful answer.
type dataEnvelope[T any] struct {
	Data T `json:"data"`
}

type loginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type sessionPayload struct {
	Token string      `json:"token"`
	User  userPayload `json:"user"`
}

type userPayload struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

type userEnvelope struct {
	User userPayload `json:"user"`
}

type activePayload struct {
	Active bool `json:"active"`
}

// errorPayload is the app's one error shape.
type errorPayload struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}
