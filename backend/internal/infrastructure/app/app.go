// Package app is the composition root: it builds the whole service (database,
// adapters, use cases, HTTP handler) from the configuration. Everything the
// process needs to run is assembled here, so tests can too.
package app

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"sync"

	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/httpapi"
	adaptersqlite "github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/sqlite"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/config"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/sqlite"
	usecasehealth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/health"
)

// App is a running set of adapters and use cases, ready to serve HTTP.
type App struct {
	handler  http.Handler
	database *sql.DB

	closeOnce sync.Once
	closeErr  error
}

// New opens the database, applies migrations and wires the HTTP API.
func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*App, error) {
	db, err := sqlite.Open(ctx, cfg.DBPath)
	if err != nil {
		return nil, err
	}

	healthChecker := usecasehealth.NewChecker(adaptersqlite.NewDatabaseChecker(db))

	return &App{
		handler:  httpapi.NewRouter(healthChecker, logger),
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
