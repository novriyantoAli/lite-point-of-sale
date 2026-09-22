package httpapi

import (
	"log/slog"
	"net/http"
	"strconv"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
	usecaseauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/auth"
)

// createUserRequest is what an Admin posts to add a Pengguna. The Peran
// arrives as a string and is validated by the use case, so an unknown value is
// invalid input rather than a decoding error.
type createUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// setUserActiveRequest is the body of a deactivation or a reactivation. The
// field is a pointer so that a body which forgot it is rejected instead of
// being read as "deactivate".
type setUserActiveRequest struct {
	Active *bool `json:"active"`
}

// createUserHandler adds a Pengguna. It is Admin-only: the router puts the
// role guard in front of it.
func createUserHandler(service AuthService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request createUserRequest
		if !decodeJSON(w, r, &request, logger) {
			return
		}

		created, err := service.CreateUser(r.Context(), usecaseauth.CreateUserInput{
			Username: request.Username,
			Password: request.Password,
			Role:     domainauth.Role(request.Role),
		})
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusCreated, dataResponse{Data: userEnvelope{
			User: newUserResponse(created),
		}}, logger)
	}
}

// listUserHandler answers every Pengguna of the store: the staff list an
// Admin manages.
func listUserHandler(service AuthService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := service.ListUsers(r.Context())
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: newUserResponses(users)}, logger)
	}
}

// setUserActiveHandler activates or deactivates a Pengguna: a resigned
// Kasir keeps their history but can no longer log in (CONTEXT.md, Nonaktif).
//
// The acting Admin is passed along, because deactivating your own account is
// how a single-Admin store locks itself out.
func setUserActiveHandler(service AuthService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetID, ok := userID(w, r, logger)
		if !ok {
			return
		}

		var request setUserActiveRequest
		if !decodeJSON(w, r, &request, logger) {
			return
		}
		if request.Active == nil {
			writeInvalidInput(w, "Field active wajib diisi.", logger)
			return
		}

		actor, ok := requireUser(w, r, logger)
		if !ok {
			return
		}

		updated, err := service.SetUserActive(r.Context(), targetID, actor.ID, *request.Active)
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: userEnvelope{
			User: newUserResponse(updated),
		}}, logger)
	}
}

// userID reads the {id} of the route. A path that is not an id at all is
// invalid input, not a missing Pengguna.
func userID(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeInvalidInput(w, "Id Pengguna tidak valid.", logger)
		return 0, false
	}

	return id, true
}
