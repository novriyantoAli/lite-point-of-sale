// Package auth holds the Pengguna domain: an account (username + password) with
// one Peran (Kasir or Admin). It declares the ports the use cases need without
// depending on SQLite, HTTP, bcrypt or any token encoding (ADR-0004).
package auth

import (
	"context"
	"errors"
	"time"
)

// Role is the Peran of a Pengguna (CONTEXT.md). The string values are what the
// API and the frontend schemas use; keep them stable.
type Role string

const (
	// RoleKasir can sell: open a keranjang, checkout, print a Struk.
	RoleKasir Role = "kasir"
	// RoleAdmin can manage Produk, Stok, Pengaturan and Pengguna.
	RoleAdmin Role = "admin"
)

// Valid reports whether r is a role the domain knows about.
func (r Role) Valid() bool {
	return r == RoleKasir || r == RoleAdmin
}

// User is the persisted Pengguna: username plus a password hash and a single
// Peran. PasswordHash must never cross the HTTP boundary — public views use
// PublicUser instead.
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         Role
	Active       bool
}

// PublicUser is the view of a Pengguna that is safe to show outside the
// usecase layer: no password hash.
type PublicUser struct {
	ID       int64
	Username string
	Role     Role
	Active   bool
}

// Public converts a User to its safe view.
func (u User) Public() PublicUser {
	return PublicUser{
		ID:       u.ID,
		Username: u.Username,
		Role:     u.Role,
		Active:   u.Active,
	}
}

// Claims is what the internal token carries: who the holder is and until when
// the token is valid.
type Claims struct {
	Username  string
	Role      Role
	ExpiresAt time.Time
}

// Session is the result of a successful login: the token SvelteKit stores in
// its httpOnly cookie, and the user it belongs to (ADR-0001, ADR-0006).
type Session struct {
	Token string
	User  PublicUser
}

// Sentinel errors the use cases return and the HTTP adapter maps to status
// codes. Keep the set small and meaningful.
var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidToken         = errors.New("invalid token")
	ErrForbidden            = errors.New("forbidden")
	ErrInvalidInput         = errors.New("invalid input")
	ErrUserNotFound         = errors.New("user not found")
	ErrUsernameTaken        = errors.New("username taken")
	ErrCannotDeactivateSelf = errors.New("cannot deactivate self")
)

// PasswordHasher hashes a plaintext password and verifies a password against a
// stored hash. Implemented with bcrypt in the adapter; the use cases only know
// this port.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

// TokenManager issues and verifies the internal token Go signs and validates
// on every call (ADR-0001). The encoding (HMAC, JWT, ...) is an adapter detail.
type TokenManager interface {
	Issue(user PublicUser) (string, error)
	Verify(token string) (Claims, error)
}

// UserRepository is the outbound port for Pengguna persistence. Implemented in
// the SQLite adapter; the use cases never see SQL (ADR-0004).
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (User, error)
	FindByID(ctx context.Context, id int64) (User, error)
	Create(ctx context.Context, user User) (User, error)
	SetActive(ctx context.Context, id int64, active bool) error
	List(ctx context.Context) ([]User, error)
	HasAdmin(ctx context.Context) (bool, error)
}
