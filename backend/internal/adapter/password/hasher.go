// Package password hashes and verifies Pengguna passwords with bcrypt. It is
// an adapter: the use cases only know domain/auth.PasswordHasher, so swapping
// the algorithm (argon2, scrypt) never reaches them.
package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
)

// MaxLength is the longest password bcrypt will accept: it hashes at most 72
// bytes, so the use case rejects anything longer instead of answering 500.
const MaxLength = 72

// Hasher hashes passwords at a fixed bcrypt cost.
type Hasher struct {
	cost int
}

// NewHasher returns a Hasher at bcrypt's default cost.
func NewHasher() *Hasher {
	return &Hasher{cost: bcrypt.DefaultCost}
}

// NewHasherWithCost returns a Hasher at the given bcrypt cost. Tests use the
// minimum cost: the work factor is the only thing being traded off, and paying
// the production cost on every test run buys nothing.
func NewHasherWithCost(cost int) *Hasher {
	return &Hasher{cost: cost}
}

// Hash returns the bcrypt hash of plain.
func (h *Hasher) Hash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hash), nil
}

// Verify reports nil when plain produced hash. A mismatch — and a hash too
// corrupt to compare — both answer domainauth.ErrInvalidCredentials, so the
// login use case never has to know bcrypt's error values.
func (h *Hasher) Verify(plain, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return domainauth.ErrInvalidCredentials
	}

	return nil
}
