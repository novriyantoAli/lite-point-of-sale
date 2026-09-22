// Package httpapi adapts use cases to HTTP. It decides status codes and
// response shapes; it holds no business rules (ADR-0004).
//
// These routes are the internal API of ADR-0001: SvelteKit calls them and the
// browser never does. Every Pengguna route needs the token the BFF forwards as
// a Bearer header, and the roles it enforces are the Peran of CONTEXT.md
// (Kasir/Admin) — never a per-route permission list.
//
// There is deliberately no logout route. The token carries no server-side
// state, so logging out is SvelteKit dropping its httpOnly cookie; a token that
// has already been handed out stays valid until it expires.
package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/health"
)

// HealthChecker is the use case the health endpoint depends on.
type HealthChecker interface {
	Check(ctx context.Context) health.Report
}

// NewRouter returns the API router with every route the service exposes.
func NewRouter(healthChecker HealthChecker, auth AuthService, products ProductService, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// Public: a liveness and readiness probe, and the one route the UI reads
	// before anybody has logged in.
	mux.Handle("GET /api/health", healthHandler(healthChecker, logger))

	// SvelteKit needs no token to log in — it is asking for one.
	mux.Handle("POST /api/auth/login", loginHandler(auth, logger))

	// Every other route runs behind the token check first, then the role check.
	authenticated := func(next http.HandlerFunc) http.HandlerFunc {
		return withAuthentication(auth, logger, next)
	}
	adminOnly := func(next http.HandlerFunc) http.HandlerFunc {
		return authenticated(requireRole(domainauth.RoleAdmin, logger, next))
	}

	mux.Handle("GET /api/auth/me", authenticated(meHandler(logger)))
	mux.Handle("POST /api/pengguna", adminOnly(createUserHandler(auth, logger)))
	mux.Handle("GET /api/pengguna", adminOnly(listUserHandler(auth, logger)))
	mux.Handle("PATCH /api/pengguna/{id}", adminOnly(setUserActiveHandler(auth, logger)))

	// The catalogue is an Admin's to manage, so every Produk route sits behind
	// the same role guard. The kasir lookup of #6 is a read this API already
	// supports (`GET /api/produk?active=true`) — it will relax the guard on that
	// one route when the till needs it.
	mux.Handle("GET /api/produk", adminOnly(listProductHandler(products, logger)))
	// Registered before the wildcard, and as a literal it wins either way: the
	// Kategori list is not a Produk id.
	mux.Handle("GET /api/produk/kategori", adminOnly(listCategoryHandler(products, logger)))
	mux.Handle("POST /api/produk", adminOnly(createProductHandler(products, logger)))
	mux.Handle("PUT /api/produk/{id}", adminOnly(updateProductHandler(products, logger)))
	mux.Handle("PATCH /api/produk/{id}", adminOnly(setProductActiveHandler(products, logger)))
	mux.Handle("DELETE /api/produk/{id}", adminOnly(deleteProductHandler(products, logger)))

	return mux
}

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func healthHandler(checker HealthChecker, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := checker.Check(r.Context())

		status := http.StatusOK
		if report.Status != health.StatusOK {
			status = http.StatusServiceUnavailable
		}

		writeJSON(w, status, healthResponse{
			Status:   string(report.Status),
			Database: string(report.Database),
		}, logger)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("write response", "error", err)
	}
}
