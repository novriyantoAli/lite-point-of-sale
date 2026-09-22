// Package config reads the server configuration from the environment.
package config

import "os"

// Config is everything the server needs to start.
type Config struct {
	// HTTPAddr is the listen address of the API, e.g. ":8080".
	HTTPAddr string
	// DBPath is the SQLite database file. One store, one terminal (ADR-0002).
	DBPath string
}

// Load reads the configuration from the environment, falling back to
// development defaults.
func Load() Config {
	return Config{
		HTTPAddr: envOr("POS_HTTP_ADDR", ":8080"),
		DBPath:   envOr("POS_DB_PATH", "./data/pos.db"),
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
