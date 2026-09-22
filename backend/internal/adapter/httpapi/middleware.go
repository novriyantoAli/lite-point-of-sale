package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
)

// contextKey is private so nothing outside this package can plant a Pengguna in
// a request context and impersonate one.
type contextKey struct{}

var userContextKey contextKey

// Authenticator is the one use case the middleware needs.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (domainauth.PublicUser, error)
}

// withAuthentication turns away a request that carries no valid token, and puts
// the Pengguna it belongs to on the request context. Go validates the token on
// every call, so a deactivated Pengguna loses access immediately (ADR-0001).
func withAuthentication(authenticator Authenticator, logger *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := authenticator.Authenticate(r.Context(), bearerToken(r))
		if err != nil {
			writeError(w, err, logger)
			return
		}

		next(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
	}
}

// requireRole turns away an authenticated Pengguna whose Peran is not the one
// the route calls for. It runs after withAuthentication, so an anonymous
// request is answered 401 and a wrong role 403.
func requireRole(role domainauth.Role, logger *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := currentUser(r.Context())
		if !ok {
			// Reaching here means the route was wired without
			// withAuthentication in front of it.
			writeError(w, domainauth.ErrInvalidToken, logger)
			return
		}

		if user.Role != role {
			writeError(w, domainauth.ErrForbidden, logger)
			return
		}

		next(w, r)
	}
}

// currentUser returns the Pengguna withAuthentication put on the context.
func currentUser(ctx context.Context) (domainauth.PublicUser, bool) {
	user, ok := ctx.Value(userContextKey).(domainauth.PublicUser)
	return user, ok
}

// bearerToken reads the token SvelteKit forwards on behalf of the browser. An
// absent header yields an empty token, which fails verification like any other
// invalid one.
func bearerToken(r *http.Request) string {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return ""
	}

	return strings.TrimSpace(token)
}
