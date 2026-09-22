// Package sqlite implements the outbound ports (repositories) declared by the
// domain on top of SQLite. Keeping this the only layer that knows SQL means a
// move to another database is a rewrite of this package alone (ADR-0004).
package sqlite

import (
	"context"
	"database/sql"
)

// DatabaseChecker reports whether the SQLite database is reachable and has
// been migrated. It reads a row written by the initial migration, so an
// unmigrated database is reported as unhealthy rather than silently OK.
type DatabaseChecker struct {
	db *sql.DB
}

// NewDatabaseChecker returns a checker backed by the given database handle.
func NewDatabaseChecker(db *sql.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

// Check satisfies domain/health.DatabaseChecker.
func (c *DatabaseChecker) Check(ctx context.Context) error {
	var schemaVersion string
	return c.db.QueryRowContext(ctx, `SELECT value FROM app_meta WHERE key = 'schema_version'`).Scan(&schemaVersion)
}
