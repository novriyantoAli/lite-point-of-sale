// Package app is the composition root: it builds the whole service (database,
// adapters, use cases, HTTP handler) from the configuration. Everything the
// process needs to run is assembled here, so tests can too.
package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/httpapi"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/password"
	adaptersqlite "github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/sqlite"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/token"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/config"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/sqlite"
	usecaseauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/auth"
	usecasehealth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/health"
)

// App is a running set of adapters and use cases, ready to serve HTTP.
type App struct {
	handler  http.Handler
	database *sql.DB

	closeOnce sync.Once
	closeErr  error
}

// New opens the database, applies migrations, seeds the first Admin Pengguna if
// the store has none, and wires the HTTP API.
func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*App, error) {
	db, err := sqlite.Open(ctx, cfg.DBPath)
	if err != nil {
		return nil, err
	}

	// Every failure below has to release the database handle: the caller gets
	// no App, so there is nothing left for it to Close.
	fail := func(err error) (*App, error) {
		db.Close()
		return nil, err
	}

	tokenManager, err := token.NewManager([]byte(cfg.TokenSecret), cfg.SessionTTL)
	if err != nil {
		return fail(fmt.Errorf("build token manager: %w", err))
	}
	if cfg.TokenSecret == config.DevTokenSecret {
		logger.Warn("POS_TOKEN_SECRET is not set: signing sessions with the development secret")
	}

	authService := usecaseauth.NewService(
		adaptersqlite.NewUserRepository(db),
		password.NewHasher(),
		tokenManager,
	)

	// A store with no Admin can never create one through the API, so startup
	// seeds it. On every later start this is a lookup that finds an Admin.
	seeded, err := authService.EnsureAdmin(ctx, cfg.AdminUsername, cfg.AdminPassword)
	if err != nil {
		return fail(err)
	}
	if seeded {
		logger.Info("seeded the initial Admin Pengguna", "username", cfg.AdminUsername)
		if cfg.AdminPassword == config.DevAdminPassword {
			logger.Warn("the seeded Admin uses the development password: set POS_ADMIN_PASSWORD")
		}
	}

	healthChecker := usecasehealth.NewChecker(adaptersqlite.NewDatabaseChecker(db))

	return &App{
		handler:  httpapi.NewRouter(healthChecker, authService, logger),
		database: db,
	}, nil
}

// Handler returns the HTTP handler of the API.
func (a *App) Handler() http.Handler {
	return a.handler
}

// Close releases the database. It is safe to call more than once.
func (a *App) Close() error {
	a.closeOnce.Do(func() {
		a.closeErr = a.database.Close()
	})
	return a.closeErr
}
