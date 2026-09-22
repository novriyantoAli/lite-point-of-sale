package password

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
	usecaseauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/auth"
)

// newTestHasher uses bcrypt's minimum cost: the work factor is the only thing
// being traded off, and paying the production cost per test run buys nothing.
func newTestHasher() *Hasher {
	return NewHasherWithCost(bcrypt.MinCost)
}

func TestHashAndVerifyRoundTrip(t *testing.T) {
	hasher := newTestHasher()

	hash, err := hasher.Hash("rahasia123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	if strings.Contains(hash, "rahasia123") {
		t.Errorf("hash %q contains the plaintext password", hash)
	}
	if err := hasher.Verify("rahasia123", hash); err != nil {
		t.Errorf("verify: got %v, want nil", err)
	}
}

func TestHashSaltsEveryPassword(t *testing.T) {
	hasher := newTestHasher()

	first, err := hasher.Hash("rahasia123")
	if err != nil {
		t.Fatalf("hash first: %v", err)
	}
	second, err := hasher.Hash("rahasia123")
	if err != nil {
		t.Fatalf("hash second: %v", err)
	}

	if first == second {
		t.Error("hashing the same password twice gave the same hash, want a fresh salt each time")
	}
}

func TestVerifyRejectsWhatItCannotMatch(t *testing.T) {
	hasher := newTestHasher()

	hash, err := hasher.Hash("rahasia123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
	}{
		{name: "wrong password", password: "salah123", hash: hash},
		{name: "empty password", password: "", hash: hash},
		{name: "hash that is not bcrypt", password: "rahasia123", hash: "bukan-hash"},
		{name: "empty hash", password: "rahasia123", hash: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// The use case only distinguishes "matches" from "does not"; a
			// corrupt hash must not become a panic or a bcrypt error.
			if err := hasher.Verify(test.password, test.hash); !errors.Is(err, domainauth.ErrInvalidCredentials) {
				t.Fatalf("verify: got error %v, want %v", err, domainauth.ErrInvalidCredentials)
			}
		})
	}
}

func TestHashRejectsMoreThanBcryptAccepts(t *testing.T) {
	hasher := newTestHasher()

	if _, err := hasher.Hash(strings.Repeat("a", MaxLength+1)); err == nil {
		t.Fatal("hash: got a hash for an over-long password, want an error")
	}
}

// The use case rejects over-long passwords before they reach this package, so
// its limit and bcrypt's ceiling have to be the same number — otherwise a
// password between the two answers 500 instead of 400.
func TestMaxLengthMatchesTheUseCaseLimit(t *testing.T) {
	if MaxLength != usecaseauth.MaxPasswordLength {
		t.Fatalf("password.MaxLength (%d) and usecaseauth.MaxPasswordLength (%d) have drifted apart",
			MaxLength, usecaseauth.MaxPasswordLength)
	}
}
