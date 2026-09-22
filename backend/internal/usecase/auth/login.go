package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
)

// Login verifies a username and password and issues the token SvelteKit keeps
// in its httpOnly cookie (ADR-0001). The browser never sees this token.
//
// A wrong password, an unknown username and a nonaktif account all answer the
// same error: the API must not tell a caller which usernames exist.
func (s *Service) Login(ctx context.Context, username, password string) (domainauth.Session, error) {
	user, err := s.users.FindByUsername(ctx, normalizeUsername(username))
	if err != nil {
		if errors.Is(err, domainauth.ErrUserNotFound) {
			return domainauth.Session{}, domainauth.ErrInvalidCredentials
		}
		return domainauth.Session{}, err
	}

	if !user.Active {
		return domainauth.Session{}, domainauth.ErrInvalidCredentials
	}

	if err := s.hasher.Verify(password, user.PasswordHash); err != nil {
		return domainauth.Session{}, domainauth.ErrInvalidCredentials
	}

	token, err := s.tokens.Issue(user.Public())
	if err != nil {
		return domainauth.Session{}, fmt.Errorf("issue token: %w", err)
	}

	return domainauth.Session{Token: token, User: user.Public()}, nil
}

// Authenticate validates the token Go receives on every call from SvelteKit
// and returns the Pengguna it belongs to.
//
// The account is re-read on every call instead of trusting the token's own
// claims: deactivating a Pengguna ends their session on the next request, and
// a role change applies immediately rather than at token expiry.
func (s *Service) Authenticate(ctx context.Context, token string) (domainauth.PublicUser, error) {
	claims, err := s.tokens.Verify(token)
	if err != nil {
		return domainauth.PublicUser{}, domainauth.ErrInvalidToken
	}

	user, err := s.users.FindByUsername(ctx, claims.Username)
	if err != nil {
		if errors.Is(err, domainauth.ErrUserNotFound) {
			return domainauth.PublicUser{}, domainauth.ErrInvalidToken
		}
		return domainauth.PublicUser{}, err
	}

	if !user.Active {
		return domainauth.PublicUser{}, domainauth.ErrInvalidToken
	}

	return user.Public(), nil
}

// normalizeUsername keeps one account per name regardless of how it was typed,
// so "Admin" cannot become a second Pengguna beside "admin".
func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}
