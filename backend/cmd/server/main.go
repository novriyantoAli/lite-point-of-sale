// Command server runs the Go API: HTTP on one side, SQLite on the other.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/app"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/config"
)

const shutdownTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer application.Close()

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           application.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverError := make(chan error, 1)
	go func() {
		logger.Info("api listening", "addr", cfg.HTTPAddr, "database", cfg.DBPath)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError <- err
		}
		close(serverError)
	}()

	select {
	case err := <-serverError:
		return err
	case <-ctx.Done():
		logger.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}
