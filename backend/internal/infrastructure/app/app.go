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
	"time"

	adapterbackup "github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/backup"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/escpos"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/httpapi"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/password"
	adaptersqlite "github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/sqlite"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/token"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/config"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/scheduler"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/sqlite"
	usecaseauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/auth"
	usecasebackup "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/backup"
	usecasehealth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/health"
	usecasepengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/pengaturan"
	usecasepenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/penjualan"
	usecaseproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/produk"
)

// App is a running set of adapters and use cases, ready to serve HTTP.
type App struct {
	handler  http.Handler
	database *sql.DB

	// backupCancel stops the daily backup goroutine; backupDone reports it has
	// exited, so Close can release the database without racing a snapshot.
	backupCancel context.CancelFunc
	backupDone   chan struct{}

	closeOnce sync.Once
	closeErr  error
}

// backupInterval is how often the automatic backup runs: once a day. The
// scheduler runs it once at startup too, so a fresh store already has a
// snapshot before its first sale (issue #10).
const backupInterval = 24 * time.Hour

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

	// One Produk repository instance serves both slices: the catalogue's own use
	// cases and the checkout that reads a Produk to price a cart.
	productRepository := adaptersqlite.NewProductRepository(db)

	// Pengaturan is its own domain (Struk template, ambang Stok menipis), and its
	// service also satisfies usecaseproduk.LowStockSettings — the Produk use case
	// reads the ambang through that port without importing the pengaturan domain
	// (ADR-0017, keputusan 3).
	settingsService := usecasepengaturan.NewService(adaptersqlite.NewSettingsRepository(db))
	productService := usecaseproduk.NewService(productRepository, settingsService)

	// The printer of this store: a device path when POS_PRINTER_DEVICE is set, and
	// the honest null printer when it is not (ADR-0017, keputusan 2). The same
	// Pengaturan service is read by the checkout for the Struk template.
	printer := escpos.New(cfg.PrinterDevice)

	saleService := usecasepenjualan.NewService(
		productRepository,
		adaptersqlite.NewSaleRepository(db),
		settingsService,
		printer,
	)

	// Backup is the store's safety net: a manual export through the API and a
	// daily snapshot driven by the scheduler. Both take the same snapshot, so a
	// backup is a backup regardless of who asked for it (issue #10).
	backupService := usecasebackup.NewService(
		adapterbackup.NewRepository(db, cfg.BackupDir),
		cfg.BackupRetentionDays,
	)

	backupCtx, backupCancel := context.WithCancel(context.Background())
	backupDone := make(chan struct{})
	go func() {
		defer close(backupDone)
		scheduler.Run(backupCtx, func(ctx context.Context) error {
			_, err := backupService.Create(ctx)
			return err
		}, backupInterval, logger)
	}()

	return &App{
		handler:      httpapi.NewRouter(healthChecker, authService, productService, saleService, settingsService, backupService, logger),
		database:     db,
		backupCancel: backupCancel,
		backupDone:   backupDone,
	}, nil
}

// Handler returns the HTTP handler of the API.
func (a *App) Handler() http.Handler {
	return a.handler
}

// Close releases the database. It is safe to call more than once. The daily
// backup goroutine is stopped first and joined, so no snapshot races the close.
func (a *App) Close() error {
	a.closeOnce.Do(func() {
		a.backupCancel()
		<-a.backupDone
		a.closeErr = a.database.Close()
	})
	return a.closeErr
}
