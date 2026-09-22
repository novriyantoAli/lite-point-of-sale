package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
	usecaseauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/auth"
)

// AuthService is the Pengguna use cases the HTTP adapter depends on. Declaring
// it here, next to the handlers that use it, keeps this package testable with a
// fake service and keeps the dependency pointing inward (ADR-0004).
type AuthService interface {
	Login(ctx context.Context, username, password string) (domainauth.Session, error)
	Authenticate(ctx context.Context, token string) (domainauth.PublicUser, error)
	CreateUser(ctx context.Context, input usecaseauth.CreateUserInput) (domainauth.PublicUser, error)
	SetUserActive(ctx context.Context, targetID, actorID int64, active bool) (domainauth.PublicUser, error)
	ListUsers(ctx context.Context) ([]domainauth.PublicUser, error)
}

// userResponse is the JSON view of a Pengguna. It is built from
// domainauth.PublicUser, so there is no password hash here to leak by accident.
type userResponse struct {
	ID       int64           `json:"id"`
	Username string          `json:"username"`
	Role     domainauth.Role `json:"role"`
	Active   bool            `json:"active"`
}

func newUserResponse(user domainauth.PublicUser) userResponse {
	return userResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		Active:   user.Active,
	}
}

func newUserResponses(users []domainauth.PublicUser) []userResponse {
	responses := make([]userResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, newUserResponse(user))
	}

	return responses
}

// userEnvelope wraps a Pengguna, so `/auth/me` answers the same shape as
// `/auth/login` does for its Pengguna.
type userEnvelope struct {
	User userResponse `json:"user"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type sessionResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

// loginHandler verifies a username and password and answers the token.
//
// The browser never calls this: the SvelteKit BFF does, and keeps the token in
// its httpOnly cookie instead of handing it to the browser (ADR-0001).
func loginHandler(service AuthService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request loginRequest
		if !decodeJSON(w, r, &request, logger) {
			return
		}

		session, err := service.Login(r.Context(), request.Username, request.Password)
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: sessionResponse{
			Token: session.Token,
			User:  newUserResponse(session.User),
		}}, logger)
	}
}

// meHandler answers the Pengguna behind the token on the request. The BFF calls
// it to rebuild its session view on load, so the cookie only ever holds the
// token and never a copy of the Pengguna.
func meHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := currentUser(r.Context())
		if !ok {
			// Reaching here means the route was wired without
			// withAuthentication in front of it.
			writeError(w, domainauth.ErrInvalidToken, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: userEnvelope{
			User: newUserResponse(user),
		}}, logger)
	}
}
