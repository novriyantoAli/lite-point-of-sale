// Package auth contains the Pengguna use cases: login, the per-call
// authentication SvelteKit relies on, and Pengguna management for Admin.
// Everything it needs arrives as a port, so these cases run without SQLite,
// bcrypt or a token encoding (ADR-0004, ADR-0007).
package auth

import (
	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
)

// MinPasswordLength is the shortest password an Admin may set for a Pengguna.
const MinPasswordLength = 8

// MaxPasswordLength is bcrypt's ceiling: it hashes at most 72 bytes, so a
// longer password is rejected here as invalid input rather than surfacing as a
// hashing failure. A test in the password adapter fails if the two drift apart.
const MaxPasswordLength = 72

// InputError is a validation failure that carries a message fit for the API
// response. It unwraps to domainauth.ErrInvalidInput, so the HTTP adapter only
// needs errors.Is to pick the status code and errors.As to read the message.
type InputError struct {
	Message string
}

func (e InputError) Error() string { return e.Message }

// Unwrap makes errors.Is(err, domainauth.ErrInvalidInput) true.
func (e InputError) Unwrap() error { return domainauth.ErrInvalidInput }

// Service holds the ports the Pengguna use cases need.
type Service struct {
	users  domainauth.UserRepository
	hasher domainauth.PasswordHasher
	tokens domainauth.TokenManager
}

// NewService wires the Pengguna use cases to their ports.
func NewService(users domainauth.UserRepository, hasher domainauth.PasswordHasher, tokens domainauth.TokenManager) *Service {
	return &Service{users: users, hasher: hasher, tokens: tokens}
}
