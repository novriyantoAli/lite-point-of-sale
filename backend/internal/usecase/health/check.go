// Package health contains the health check use case.
package health

import (
	"context"

	domainhealth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/health"
)

// Checker reports whether the service and its dependencies are healthy.
type Checker struct {
	database domainhealth.DatabaseChecker
}

// NewChecker wires the health check to its database port.
func NewChecker(database domainhealth.DatabaseChecker) *Checker {
	return &Checker{database: database}
}

// Check returns a report for the service. A database that cannot be read
// degrades the whole report: the API is useless without its data.
func (c *Checker) Check(ctx context.Context) domainhealth.Report {
	report := domainhealth.Report{
		Status:   domainhealth.StatusOK,
		Database: domainhealth.StatusOK,
	}

	if err := c.database.Check(ctx); err != nil {
		report.Status = domainhealth.StatusDegraded
		report.Database = domainhealth.StatusUnavailable
	}

	return report
}
