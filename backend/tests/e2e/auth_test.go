package e2e

import (
	"net/http"
	"strings"
	"testing"
)

const kasirPassword = "rahasia-kasir"

func TestLoginRejectsBadCredentials(t *testing.T) {
	_, baseURL := startAPI(t, newTestConfig(t))

	tests := []struct {
		name     string
		username string
		password string
	}{
		{name: "wrong password", username: seededAdmin, password: "salah-sekali"},
		{name: "unknown username", username: "hantu", password: testAdminPassword},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodPost, baseURL+"/api/auth/login", "",
				loginPayload{Username: test.username, Password: test.password}, &failure)

			if status != http.StatusUnauthorized {
				t.Fatalf("login: got status %d, want %d", status, http.StatusUnauthorized)
			}
			if failure.Error != "invalid_credentials" || failure.Message == "" {
				t.Errorf("login: got %+v, want code invalid_credentials with a message", failure)
			}
		})
	}
}

func TestAnonymousCallsAreRejected(t *testing.T) {
	_, baseURL := startAPI(t, newTestConfig(t))

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "read own session", method: http.MethodGet, path: "/api/auth/me"},
		{name: "list Pengguna", method: http.MethodGet, path: "/api/pengguna"},
		{name: "create Pengguna", method: http.MethodPost, path: "/api/pengguna"},
		{name: "deactivate Pengguna", method: http.MethodPatch, path: "/api/pengguna/1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, test.method, baseURL+test.path, "", nil, &failure)

			if status != http.StatusUnauthorized {
				t.Fatalf("%s %s: got status %d, want %d", test.method, test.path, status, http.StatusUnauthorized)
			}
			if failure.Error != "invalid_token" {
				t.Errorf("%s %s: got code %q, want %q", test.method, test.path, failure.Error, "invalid_token")
			}
		})
	}
}

func TestSeededAdminStaysLoggedIn(t *testing.T) {
	_, baseURL := startAPI(t, newTestConfig(t))

	var session dataEnvelope[sessionPayload]
	status := apiCall(t, http.MethodPost, baseURL+"/api/auth/login", "",
		loginPayload{Username: seededAdmin, Password: testAdminPassword}, &session)

	if status != http.StatusOK {
		t.Fatalf("login: got status %d, want %d", status, http.StatusOK)
	}
	if session.Data.User.Role != "admin" || !session.Data.User.Active {
		t.Fatalf("login: got Pengguna %+v, want an active admin", session.Data.User)
	}

	// The session has to survive every later call: that is what "bertahan
	// sampai logout" means for a terminal used all day.
	for attempt := 1; attempt <= 2; attempt++ {
		var me dataEnvelope[userEnvelope]
		status := apiCall(t, http.MethodGet, baseURL+"/api/auth/me", session.Data.Token, nil, &me)

		if status != http.StatusOK {
			t.Fatalf("call %d to /api/auth/me: got status %d, want %d", attempt, status, http.StatusOK)
		}
		if me.Data.User.Username != seededAdmin {
			t.Errorf("call %d to /api/auth/me: got %q, want %q", attempt, me.Data.User.Username, seededAdmin)
		}
	}
}

func TestKasirIsForbiddenFromAdminRoutes(t *testing.T) {
	_, baseURL := startAPI(t, newTestConfig(t))
	adminToken := logIn(t, baseURL, seededAdmin, testAdminPassword)
	createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")
	kasirToken := logIn(t, baseURL, "kasir1", kasirPassword)

	// Authenticated, so the Kasir's own session works…
	var me dataEnvelope[userEnvelope]
	if status := apiCall(t, http.MethodGet, baseURL+"/api/auth/me", kasirToken, nil, &me); status != http.StatusOK {
		t.Fatalf("/api/auth/me as Kasir: got status %d, want %d", status, http.StatusOK)
	}
	if me.Data.User.Role != "kasir" {
		t.Fatalf("/api/auth/me as Kasir: got role %q, want %q", me.Data.User.Role, "kasir")
	}

	// …but the Peran is enforced on every Admin route.
	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "list Pengguna", method: http.MethodGet, path: "/api/pengguna"},
		{
			name:   "create Pengguna",
			method: http.MethodPost,
			path:   "/api/pengguna",
			body:   loginPayload{Username: "kasir2", Password: kasirPassword},
		},
		{name: "deactivate Pengguna", method: http.MethodPatch, path: "/api/pengguna/1", body: activePayload{Active: false}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, test.method, baseURL+test.path, kasirToken, test.body, &failure)

			if status != http.StatusForbidden {
				t.Fatalf("%s %s as Kasir: got status %d, want %d", test.method, test.path, status, http.StatusForbidden)
			}
			if failure.Error != "forbidden" {
				t.Errorf("%s %s as Kasir: got code %q, want %q", test.method, test.path, failure.Error, "forbidden")
			}
		})
	}
}

func TestDeactivatingAKasirEndsTheirSession(t *testing.T) {
	_, baseURL := startAPI(t, newTestConfig(t))
	adminToken := logIn(t, baseURL, seededAdmin, testAdminPassword)
	kasir := createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")
	kasirToken := logIn(t, baseURL, "kasir1", kasirPassword)

	var updated dataEnvelope[userEnvelope]
	status := apiCall(t, http.MethodPatch, baseURL+"/api/pengguna/"+itoa(kasir.ID), adminToken,
		activePayload{Active: false}, &updated)

	if status != http.StatusOK {
		t.Fatalf("deactivate: got status %d, want %d", status, http.StatusOK)
	}
	if updated.Data.User.Active {
		t.Fatalf("deactivate: got %+v, want an inactive Pengguna", updated.Data.User)
	}

	// The account cannot log in again…
	var failure errorPayload
	if status := apiCall(t, http.MethodPost, baseURL+"/api/auth/login", "",
		loginPayload{Username: "kasir1", Password: kasirPassword}, &failure); status != http.StatusUnauthorized {
		t.Errorf("login after deactivation: got status %d, want %d", status, http.StatusUnauthorized)
	}

	// …and the session it already had is over on the next call, instead of
	// living on until the token expires.
	if status := apiCall(t, http.MethodGet, baseURL+"/api/auth/me", kasirToken, nil, &failure); status != http.StatusUnauthorized {
		t.Errorf("old token after deactivation: got status %d, want %d", status, http.StatusUnauthorized)
	}
}

func TestAdminCannotDeactivateTheirOwnAccount(t *testing.T) {
	_, baseURL := startAPI(t, newTestConfig(t))
	adminToken := logIn(t, baseURL, seededAdmin, testAdminPassword)

	var me dataEnvelope[userEnvelope]
	if status := apiCall(t, http.MethodGet, baseURL+"/api/auth/me", adminToken, nil, &me); status != http.StatusOK {
		t.Fatalf("/api/auth/me: got status %d, want %d", status, http.StatusOK)
	}

	var failure errorPayload
	status := apiCall(t, http.MethodPatch, baseURL+"/api/pengguna/"+itoa(me.Data.User.ID), adminToken,
		activePayload{Active: false}, &failure)

	if status != http.StatusBadRequest {
		t.Fatalf("deactivate self: got status %d, want %d", status, http.StatusBadRequest)
	}
	if failure.Error != "cannot_deactivate_self" {
		t.Errorf("deactivate self: got code %q, want %q", failure.Error, "cannot_deactivate_self")
	}
}

func TestCreatePenggunaValidatesInput(t *testing.T) {
	_, baseURL := startAPI(t, newTestConfig(t))
	adminToken := logIn(t, baseURL, seededAdmin, testAdminPassword)
	createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")

	tests := []struct {
		name          string
		body          any
		wantStatus    int
		wantErrorCode string
	}{
		{
			name:          "empty username",
			body:          loginPayload{Username: "  ", Password: kasirPassword},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_input",
		},
		{
			name:          "password below the minimum",
			body:          loginPayload{Username: "kasir2", Password: "pendek"},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_input",
		},
		{
			name:          "Peran the domain does not know",
			body:          map[string]any{"username": "kasir2", "password": kasirPassword, "role": "pemilik"},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_input",
		},
		{
			name:          "username already taken, whatever the case",
			body:          map[string]any{"username": "KASIR1", "password": kasirPassword, "role": "kasir"},
			wantStatus:    http.StatusConflict,
			wantErrorCode: "username_taken",
		},
		{
			name:          "body that is not JSON at all",
			body:          "bukan json",
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_input",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodPost, baseURL+"/api/pengguna", adminToken, test.body, &failure)

			if status != test.wantStatus {
				t.Fatalf("create: got status %d, want %d", status, test.wantStatus)
			}
			if failure.Error != test.wantErrorCode {
				t.Errorf("create: got code %q, want %q", failure.Error, test.wantErrorCode)
			}
			if failure.Message == "" {
				t.Error("create: got no message, want one the form can show")
			}
		})
	}
}

func TestAPenggunaIsNeverAnsweredWithItsPassword(t *testing.T) {
	_, baseURL := startAPI(t, newTestConfig(t))
	adminToken := logIn(t, baseURL, seededAdmin, testAdminPassword)
	createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")

	status, body := apiCallRaw(t, http.MethodGet, baseURL+"/api/pengguna", adminToken)
	if status != http.StatusOK {
		t.Fatalf("list: got status %d, want %d", status, http.StatusOK)
	}

	for _, secret := range []string{kasirPassword, "password_hash", "\"password\""} {
		if strings.Contains(body, secret) {
			t.Errorf("list response contains %q: %s", secret, body)
		}
	}
}

func TestSetPenggunaActiveRequiresTheField(t *testing.T) {
	_, baseURL := startAPI(t, newTestConfig(t))
	adminToken := logIn(t, baseURL, seededAdmin, testAdminPassword)

	// A body without `active` must not be read as "deactivate".
	var failure errorPayload
	status := apiCall(t, http.MethodPatch, baseURL+"/api/pengguna/1", adminToken, map[string]any{}, &failure)

	if status != http.StatusBadRequest {
		t.Fatalf("patch without active: got status %d, want %d", status, http.StatusBadRequest)
	}
	if failure.Error != "invalid_input" {
		t.Errorf("patch without active: got code %q, want %q", failure.Error, "invalid_input")
	}
}

// createPengguna adds a Pengguna as an Admin and returns it, failing the test
// when the Admin could not create it.
func createPengguna(t *testing.T, baseURL, adminToken, username, password, role string) userPayload {
	t.Helper()

	var created dataEnvelope[userEnvelope]
	status := apiCall(t, http.MethodPost, baseURL+"/api/pengguna", adminToken, map[string]any{
		"username": username,
		"password": password,
		"role":     role,
	}, &created)

	if status != http.StatusCreated {
		t.Fatalf("create Pengguna %s: got status %d, want %d", username, status, http.StatusCreated)
	}
	if created.Data.User.Username != username || created.Data.User.Role != role || !created.Data.User.Active {
		t.Fatalf("create Pengguna %s: got %+v", username, created.Data.User)
	}

	return created.Data.User
}
