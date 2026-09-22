package auth

import (
	"context"
	"errors"
	"fmt"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
)

// CreateUserInput is what an Admin fills in to add a Pengguna.
type CreateUserInput struct {
	Username string
	Password string
	Role     domainauth.Role
}

// CreateUser adds a Pengguna with a hashed password. Only an Admin reaches
// this use case; the HTTP adapter is what enforces that role.
func (s *Service) CreateUser(ctx context.Context, input CreateUserInput) (domainauth.PublicUser, error) {
	username := normalizeUsername(input.Username)
	if username == "" {
		return domainauth.PublicUser{}, InputError{Message: "Username wajib diisi."}
	}
	if len(input.Password) < MinPasswordLength {
		return domainauth.PublicUser{}, InputError{
			Message: fmt.Sprintf("Password minimal %d karakter.", MinPasswordLength),
		}
	}
	if len(input.Password) > MaxPasswordLength {
		return domainauth.PublicUser{}, InputError{
			Message: fmt.Sprintf("Password maksimal %d byte.", MaxPasswordLength),
		}
	}
	if !input.Role.Valid() {
		return domainauth.PublicUser{}, InputError{Message: "Peran harus kasir atau admin."}
	}

	// Ask first so the everyday duplicate answers with a clear conflict. The
	// repository still maps the unique constraint on its own, because two
	// Admins can race the same username past this check.
	if _, err := s.users.FindByUsername(ctx, username); err == nil {
		return domainauth.PublicUser{}, domainauth.ErrUsernameTaken
	} else if !errors.Is(err, domainauth.ErrUserNotFound) {
		return domainauth.PublicUser{}, err
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return domainauth.PublicUser{}, fmt.Errorf("hash password: %w", err)
	}

	created, err := s.users.Create(ctx, domainauth.User{
		Username:     username,
		PasswordHash: hash,
		Role:         input.Role,
		Active:       true,
	})
	if err != nil {
		return domainauth.PublicUser{}, err
	}

	return created.Public(), nil
}

// ListUsers returns every Pengguna, passwords left behind (PublicUser).
func (s *Service) ListUsers(ctx context.Context) ([]domainauth.PublicUser, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, err
	}

	public := make([]domainauth.PublicUser, 0, len(users))
	for _, user := range users {
		public = append(public, user.Public())
	}

	return public, nil
}

// SetUserActive activates or deactivates a Pengguna and returns its new state,
// so the caller answers with the updated Pengguna instead of reading it back —
// a second round trip that could also see a different state.
//
// An Admin cannot deactivate their own account: on a store with a single Admin
// that is the one move that locks everybody out, since creating a Pengguna
// already requires an Admin. With a second Admin in place, either one can
// deactivate the other.
func (s *Service) SetUserActive(ctx context.Context, targetID, actorID int64, active bool) (domainauth.PublicUser, error) {
	if !active && targetID == actorID {
		return domainauth.PublicUser{}, domainauth.ErrCannotDeactivateSelf
	}

	target, err := s.users.FindByID(ctx, targetID)
	if err != nil {
		return domainauth.PublicUser{}, err
	}

	if err := s.users.SetActive(ctx, targetID, active); err != nil {
		return domainauth.PublicUser{}, err
	}

	target.Active = active

	return target.Public(), nil
}

// EnsureAdmin creates the initial Admin Pengguna when the store has none, and
// does nothing on every start after that.
//
// Startup is the only place this can happen: the API endpoint that creates a
// Pengguna is itself Admin-only, so a store with no Admin could never get one.
func (s *Service) EnsureAdmin(ctx context.Context, username, password string) (bool, error) {
	hasAdmin, err := s.users.HasAdmin(ctx)
	if err != nil {
		return false, err
	}
	if hasAdmin {
		return false, nil
	}

	_, err = s.CreateUser(ctx, CreateUserInput{
		Username: username,
		Password: password,
		Role:     domainauth.RoleAdmin,
	})
	if err != nil {
		return false, fmt.Errorf("seed admin Pengguna: %w", err)
	}

	return true, nil
}
