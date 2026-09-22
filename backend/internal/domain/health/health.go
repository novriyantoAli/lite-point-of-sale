// Package health holds the health of the service and the port the check needs
// to observe its database. It has no dependency outside the standard library
// (ADR-0004).
package health

import "context"

// Status describes the state of the service or one of its dependencies.
type Status string

const (
	// StatusOK means the subject is healthy.
	StatusOK Status = "ok"
	// StatusDegraded means the service is up but a dependency is not usable.
	StatusDegraded Status = "degraded"
	// StatusUnavailable means the subject cannot be reached at all.
	StatusUnavailable Status = "unavailable"
)

// Report is the outcome of a health check.
type Report struct {
	Status   Status
	Database Status
}

// DatabaseChecker is the outbound port the health use case needs: it reports
// whether the database is reachable and readable. Implemented in adapter
// (ADR-0004), so the port stays free of SQLite types.
type DatabaseChecker interface {
	Check(ctx context.Context) error
}
