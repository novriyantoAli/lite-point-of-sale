package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
	usecaseauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/auth"
)

// dataResponse is the envelope of every successful answer of this API: the
// payload sits under `data`, so a list can grow siblings later (total_count,
// page, page_size) without breaking clients. The health endpoint predates this
// rule and keeps its flat shape.
type dataResponse struct {
	Data any `json:"data"`
}

// errorResponse is the only error shape of this API. The frontend normalizes it
// into AppError {message, status, code} — lib/api/errors.ts reads `message` and
// `error` — so no client ever turns a raw HTTP status into prose.
type errorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// failure pairs the HTTP status with the machine-readable code and the message
// shown when an error carries none of its own.
type failure struct {
	status  int
	code    string
	message string
}

// invalidInputFailure is also the answer to a body that will not parse.
var invalidInputFailure = failure{
	status:  http.StatusBadRequest,
	code:    "invalid_input",
	message: "Permintaan tidak valid.",
}

// failures maps the domain's sentinel errors to HTTP. It is an ordered list,
// not a map: the answer to a given error should not depend on iteration order.
var failures = []struct {
	cause error
	as    failure
}{
	{
		cause: domainauth.ErrInvalidCredentials,
		as: failure{
			status:  http.StatusUnauthorized,
			code:    "invalid_credentials",
			message: "Username atau password salah.",
		},
	},
	{
		cause: domainauth.ErrInvalidToken,
		as: failure{
			status:  http.StatusUnauthorized,
			code:    "invalid_token",
			message: "Sesi tidak valid atau sudah berakhir. Silakan login kembali.",
		},
	},
	{
		cause: domainauth.ErrForbidden,
		as: failure{
			status:  http.StatusForbidden,
			code:    "forbidden",
			message: "Anda tidak berhak melakukan tindakan ini.",
		},
	},
	{
		cause: domainauth.ErrUserNotFound,
		as: failure{
			status:  http.StatusNotFound,
			code:    "user_not_found",
			message: "Pengguna tidak ditemukan.",
		},
	},
	{
		cause: domainauth.ErrUsernameTaken,
		as: failure{
			status:  http.StatusConflict,
			code:    "username_taken",
			message: "Username sudah dipakai.",
		},
	},
	{
		cause: domainauth.ErrCannotDeactivateSelf,
		as: failure{
			status:  http.StatusBadRequest,
			code:    "cannot_deactivate_self",
			message: "Anda tidak bisa menonaktifkan akun sendiri.",
		},
	},
	{
		cause: domainauth.ErrInvalidInput,
		as:    invalidInputFailure,
	},
}

// writeError answers a use case failure. A validation error keeps its own
// message, because it is written for the person filling in the form; anything
// unrecognized is a 500 whose detail stays in the log.
func writeError(w http.ResponseWriter, err error, logger *slog.Logger) {
	var inputErr usecaseauth.InputError
	if errors.As(err, &inputErr) && inputErr.Message != "" {
		writeJSON(w, invalidInputFailure.status, errorResponse{
			Message: inputErr.Message,
			Error:   invalidInputFailure.code,
		}, logger)
		return
	}

	for _, mapping := range failures {
		if errors.Is(err, mapping.cause) {
			writeJSON(w, mapping.as.status, errorResponse{
				Message: mapping.as.message,
				Error:   mapping.as.code,
			}, logger)
			return
		}
	}

	logger.Error("unhandled api error", "error", err)
	writeJSON(w, http.StatusInternalServerError, errorResponse{
		Message: "Terjadi kesalahan pada server.",
		Error:   "internal_error",
	}, logger)
}

// decodeJSON reads the request body into target, answering 400 instead of
// letting a malformed body reach a use case.
func decodeJSON(w http.ResponseWriter, r *http.Request, target any, logger *slog.Logger) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeJSON(w, invalidInputFailure.status, errorResponse{
			Message: invalidInputFailure.message,
			Error:   invalidInputFailure.code,
		}, logger)
		return false
	}

	return true
}
