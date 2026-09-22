// Package token issues and verifies the internal token Go hands to SvelteKit
// and expects back on every call (ADR-0001). It is an adapter: the use cases
// only know domain/auth.TokenManager, so how the token is encoded stays here.
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
)

// Manager issues HMAC-SHA256 tokens: a base64url JSON payload, a dot, and the
// base64url signature of that payload. No dependency, and no secret in the
// payload — a holder can read it, but only the signer can change it.
//
// A token is a pure function of the Pengguna and the second it was issued in:
// there is no nonce, so two logins inside the same second produce the same
// token. Nothing depends on tokens being distinct — the use case re-reads the
// Pengguna on every call — and leaving randomness out keeps the encoding
// reproducible.
type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// claims is the wire form of domain/auth.Claims.
type claims struct {
	Username  string          `json:"sub"`
	Role      domainauth.Role `json:"role"`
	ExpiresAt int64           `json:"exp"`
}

// NewManager returns a Manager signing with secret and issuing tokens that
// live for ttl. An empty secret is refused: it would sign with a key anybody
// knows, which is the same as signing with none.
func NewManager(secret []byte, ttl time.Duration) (*Manager, error) {
	if len(secret) == 0 {
		return nil, errors.New("token secret must not be empty")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("token lifetime must be positive, got %s", ttl)
	}

	return &Manager{secret: secret, ttl: ttl, now: time.Now}, nil
}

// Issue returns a signed token for the Pengguna.
func (m *Manager) Issue(user domainauth.PublicUser) (string, error) {
	payload, err := json.Marshal(claims{
		Username:  user.Username,
		Role:      user.Role,
		ExpiresAt: m.now().Add(m.ttl).Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("encode claims: %w", err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)

	return encoded + "." + base64.RawURLEncoding.EncodeToString(m.sign(encoded)), nil
}

// Verify returns the claims of a token this Manager issued and that has not
// expired. Anything else answers domainauth.ErrInvalidToken — the caller never
// learns whether the payload or the signature was the problem.
func (m *Manager) Verify(token string) (domainauth.Claims, error) {
	encoded, signature, ok := strings.Cut(token, ".")
	if !ok {
		return domainauth.Claims{}, domainauth.ErrInvalidToken
	}

	given, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return domainauth.Claims{}, domainauth.ErrInvalidToken
	}
	// Verified before the payload is parsed: nothing an attacker wrote should
	// reach the JSON decoder.
	if !hmac.Equal(given, m.sign(encoded)) {
		return domainauth.Claims{}, domainauth.ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return domainauth.Claims{}, domainauth.ErrInvalidToken
	}

	var decoded claims
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return domainauth.Claims{}, domainauth.ErrInvalidToken
	}
	if decoded.Username == "" {
		return domainauth.Claims{}, domainauth.ErrInvalidToken
	}

	expiresAt := time.Unix(decoded.ExpiresAt, 0)
	if !m.now().Before(expiresAt) {
		return domainauth.Claims{}, domainauth.ErrInvalidToken
	}

	return domainauth.Claims{
		Username:  decoded.Username,
		Role:      decoded.Role,
		ExpiresAt: expiresAt,
	}, nil
}

func (m *Manager) sign(payload string) []byte {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))

	return mac.Sum(nil)
}
