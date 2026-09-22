package token

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
)

var secret = []byte("rahasia-uji-yang-panjang")

func newTestManager(t *testing.T) *Manager {
	t.Helper()

	manager, err := NewManager(secret, time.Hour)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	return manager
}

func TestIssueAndVerifyRoundTrip(t *testing.T) {
	manager := newTestManager(t)
	issuedAt := time.Date(2026, time.September, 22, 9, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return issuedAt }

	token, err := manager.Issue(domainauth.PublicUser{ID: 1, Username: "kasir1", Role: domainauth.RoleKasir})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if claims.Username != "kasir1" {
		t.Errorf("username: got %q, want %q", claims.Username, "kasir1")
	}
	if claims.Role != domainauth.RoleKasir {
		t.Errorf("role: got %q, want %q", claims.Role, domainauth.RoleKasir)
	}
	if want := issuedAt.Add(time.Hour); !claims.ExpiresAt.Equal(want) {
		t.Errorf("expiry: got %s, want %s", claims.ExpiresAt, want)
	}
}

func TestVerifyRejectsAnythingItDidNotIssue(t *testing.T) {
	manager := newTestManager(t)

	valid, err := manager.Issue(domainauth.PublicUser{Username: "kasir1", Role: domainauth.RoleKasir})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	payload, _, ok := strings.Cut(valid, ".")
	if !ok {
		t.Fatalf("issued token %q has no signature part", valid)
	}

	otherSigner, err := NewManager([]byte("kunci-yang-berbeda"), time.Hour)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	foreign, err := otherSigner.Issue(domainauth.PublicUser{Username: "kasir1", Role: domainauth.RoleAdmin})
	if err != nil {
		t.Fatalf("issue with another secret: %v", err)
	}

	tests := []struct {
		name  string
		token string
	}{
		{name: "empty"},
		{name: "no signature part", token: payload},
		{name: "extra part", token: valid + ".ekstra"},
		{name: "signature that is not base64", token: payload + ".!!!"},
		{name: "payload that is not base64", token: "!!!." + payload},
		{name: "payload that is not our JSON", token: base64.RawURLEncoding.EncodeToString([]byte("bukan json")) + "." + payload},
		{name: "signed with another secret", token: foreign},
		{
			name: "payload edited after signing",
			token: base64.RawURLEncoding.EncodeToString([]byte(
				`{"sub":"admin","role":"admin","exp":`+futureUnix()+`}`)) + "." + signatureOf(t, manager, valid),
		},
		{
			name:  "claims without a username",
			token: signedPayload(t, manager, `{"sub":"","role":"admin","exp":`+futureUnix()+`}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims, err := manager.Verify(test.token)

			if !errors.Is(err, domainauth.ErrInvalidToken) {
				t.Fatalf("verify: got error %v, want %v", err, domainauth.ErrInvalidToken)
			}
			if claims != (domainauth.Claims{}) {
				t.Errorf("verify: got claims %+v, want none", claims)
			}
		})
	}
}

func TestVerifyRejectsAnExpiredToken(t *testing.T) {
	manager := newTestManager(t)
	issuedAt := time.Date(2026, time.September, 22, 9, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return issuedAt }

	token, err := manager.Issue(domainauth.PublicUser{Username: "kasir1", Role: domainauth.RoleKasir})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	// A terminal left unattended overnight must not still be logged in.
	manager.now = func() time.Time { return issuedAt.Add(2 * time.Hour) }

	if _, err := manager.Verify(token); !errors.Is(err, domainauth.ErrInvalidToken) {
		t.Fatalf("verify expired: got error %v, want %v", err, domainauth.ErrInvalidToken)
	}
}

func TestNewManagerRefusesAnUnusableConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		secret []byte
		ttl    time.Duration
	}{
		{name: "no secret", secret: nil, ttl: time.Hour},
		{name: "empty secret", secret: []byte{}, ttl: time.Hour},
		{name: "zero lifetime", secret: secret, ttl: 0},
		{name: "negative lifetime", secret: secret, ttl: -time.Minute},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewManager(test.secret, test.ttl); err == nil {
				t.Fatal("new manager: got a manager, want an error")
			}
		})
	}
}

// signatureOf returns the signature part of an already issued token.
func signatureOf(t *testing.T, manager *Manager, token string) string {
	t.Helper()

	_, signature, ok := strings.Cut(token, ".")
	if !ok {
		t.Fatalf("issued token %q has no signature part", token)
	}

	return signature
}

// signedPayload builds a token carrying an arbitrary payload, signed the way
// the Manager signs, so a test can probe what Verify accepts.
func signedPayload(t *testing.T, manager *Manager, payload string) string {
	t.Helper()

	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))

	return encoded + "." + base64.RawURLEncoding.EncodeToString(manager.sign(encoded))
}

func futureUnix() string {
	return "4102444800" // 2100-01-01, far enough that only the payload matters
}
