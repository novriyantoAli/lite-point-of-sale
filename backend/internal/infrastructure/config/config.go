// Package config reads the server configuration from the environment.
package config

import (
	"os"
	"strconv"
	"time"
)

// DevTokenSecret is what signs sessions when POS_TOKEN_SECRET is not set. It is
// in the repository on purpose — a development machine has to start without
// one — which is exactly why the server logs a warning while it is in use.
const DevTokenSecret = "pos-development-token-secret"

// DevAdminPassword is the password of the seeded Admin Pengguna when
// POS_ADMIN_PASSWORD is not set. Also warned about at startup.
const DevAdminPassword = "admin123"

// Config is everything the server needs to start.
type Config struct {
	// HTTPAddr is the listen address of the API, e.g. ":8080".
	HTTPAddr string
	// DBPath is the SQLite database file. One store, one terminal (ADR-0002).
	DBPath string
	// TokenSecret signs the internal tokens Go issues to SvelteKit (ADR-0001).
	TokenSecret string
	// SessionTTL is how long an issued token stays valid.
	SessionTTL time.Duration
	// AdminUsername and AdminPassword seed the first Admin Pengguna of a store
	// that has none. Startup is the only place that can do it: the route that
	// creates a Pengguna is itself Admin-only.
	AdminUsername string
	AdminPassword string
	// PrinterDevice is the path of the thermal printer a Struk is written to,
	// e.g. /dev/usb/lp0. There is deliberately no default: /dev/usb/lp0 is only
	// right on Linux, and a wrong guess is worse than an honest "not configured"
	// (ADR-0017, keputusan 2). Empty means the null printer, which reports every
	// print as a failure rather than crashing or succeeding silently.
	PrinterDevice string
	// BackupDir is the folder snapshots of the database are written to (issue
	// #10). One store, one terminal (ADR-0002), so one folder holds all of them.
	BackupDir string
	// BackupRetentionDays is how many days of backups to keep before the oldest
	// are pruned. It has to be at least one: a window of zero would keep nothing.
	BackupRetentionDays int
}

// Default returns the configuration a development machine runs with. Tests
// build on it too, so there is one place where a default is written down.
func Default() Config {
	return Config{
		HTTPAddr:      ":8080",
		DBPath:        "./data/pos.db",
		TokenSecret:   DevTokenSecret,
		SessionTTL:    12 * time.Hour,
		AdminUsername: "admin",
		AdminPassword: DevAdminPassword,
		// No printer is configured by default: see Config.PrinterDevice.
		PrinterDevice:       "",
		BackupDir:           "./data/backup",
		BackupRetentionDays: 7,
	}
}

// Load reads the configuration from the environment, falling back to Default.
func Load() Config {
	cfg := Default()

	cfg.HTTPAddr = envOr("POS_HTTP_ADDR", cfg.HTTPAddr)
	cfg.DBPath = envOr("POS_DB_PATH", cfg.DBPath)
	cfg.TokenSecret = envOr("POS_TOKEN_SECRET", cfg.TokenSecret)
	cfg.SessionTTL = durationOr("POS_SESSION_TTL", cfg.SessionTTL)
	cfg.AdminUsername = envOr("POS_ADMIN_USERNAME", cfg.AdminUsername)
	cfg.AdminPassword = envOr("POS_ADMIN_PASSWORD", cfg.AdminPassword)
	cfg.PrinterDevice = envOr("POS_PRINTER_DEVICE", cfg.PrinterDevice)
	cfg.BackupDir = envOr("POS_BACKUP_DIR", cfg.BackupDir)
	cfg.BackupRetentionDays = intOr("POS_BACKUP_RETENTION_DAYS", cfg.BackupRetentionDays)

	return cfg
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// durationOr reads a duration such as "12h" or "30m". A value that does not
// parse falls back rather than stopping the process: a store that cannot open
// because of a typo in one env var is worse than one running on the default
// session lifetime, and the token manager refuses a non-positive value anyway.
func durationOr(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

// intOr reads a positive integer such as "7". A value that does not parse, or
// is not positive, falls back rather than stopping the process — a backup window
// of zero or less would silently keep no backups at all, which is a worse
// failure than running the default.
func intOr(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
