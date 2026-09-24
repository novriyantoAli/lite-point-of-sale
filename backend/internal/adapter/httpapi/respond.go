package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
	domainpengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/pengaturan"
	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
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

// inputError is how a use case reports a failure whose message is written for
// the person using the form. The status code still comes from the sentinel the
// error unwraps to: an InputError unwraps to its domain's ErrInvalidInput and
// answers 400, while a checkout refused for Stok unwraps to
// ErrInsufficientStock and answers 409 with the same kind of message.
//
// Every usecase package declares its own InputError; this is the shape they
// share, so this adapter does not have to know each of them.
type inputError interface {
	error
	InputMessage() string
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
	{
		cause: domainproduk.ErrProductNotFound,
		as: failure{
			status:  http.StatusNotFound,
			code:    "product_not_found",
			message: "Produk tidak ditemukan.",
		},
	},
	{
		cause: domainproduk.ErrCodeTaken,
		as: failure{
			status:  http.StatusConflict,
			code:    "code_taken",
			message: "Kode sudah dipakai Produk lain.",
		},
	},
	{
		cause: domainproduk.ErrProductHasSales,
		as: failure{
			status:  http.StatusConflict,
			code:    "product_has_sales",
			message: "Produk yang sudah pernah terjual hanya bisa dinonaktifkan.",
		},
	},
	{
		cause: domainproduk.ErrInvalidInput,
		as:    invalidInputFailure,
	},
	{
		cause: domainpenjualan.ErrSaleNotFound,
		as: failure{
			status:  http.StatusNotFound,
			code:    "sale_not_found",
			message: "Penjualan tidak ditemukan.",
		},
	},
	{
		cause: domainpenjualan.ErrInsufficientStock,
		as: failure{
			status:  http.StatusConflict,
			code:    "insufficient_stock",
			message: "Stok tidak cukup untuk salah satu Item.",
		},
	},
	{
		cause: domainpenjualan.ErrInvalidInput,
		as:    invalidInputFailure,
	},
	{
		cause: domainpengaturan.ErrInvalidInput,
		as:    invalidInputFailure,
	},
}

// writeError answers a use case failure. A use case that wrote a message for the
// person using the form keeps it — that message is more use than a status code —
// while the status code and the machine-readable code are still the sentinel's.
// Anything unrecognized is a 500 whose detail stays in the log.
func writeError(w http.ResponseWriter, err error, logger *slog.Logger) {
	message := ""
	var inputErr inputError
	if errors.As(err, &inputErr) {
		message = inputErr.InputMessage()
	}

	for _, mapping := range failures {
		if errors.Is(err, mapping.cause) {
			if message == "" {
				message = mapping.as.message
			}
			writeJSON(w, mapping.as.status, errorResponse{
				Message: message,
				Error:   mapping.as.code,
			}, logger)
			return
		}
	}

	// An error that wrote a message but unwraps to no sentinel this API knows is
	// still the caller's mistake rather than ours: 400 with the message it wrote,
	// instead of a 500 that hides a form error.
	if message != "" {
		writeInvalidInput(w, message, logger)
		return
	}

	logger.Error("unhandled api error", "error", err)
	writeJSON(w, http.StatusInternalServerError, errorResponse{
		Message: "Terjadi kesalahan pada server.",
		Error:   "internal_error",
	}, logger)
}

// writeInvalidInput answers a 400 whose message is written for the person using
// the form. It is the path for input that never reaches a use case — a body
// that will not parse, a path that is not an id, a query parameter that is not
// a boolean — and for the InputError a use case returns.
func writeInvalidInput(w http.ResponseWriter, message string, logger *slog.Logger) {
	writeJSON(w, invalidInputFailure.status, errorResponse{
		Message: message,
		Error:   invalidInputFailure.code,
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
