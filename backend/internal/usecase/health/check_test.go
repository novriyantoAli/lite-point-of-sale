package health

import (
	"context"
	"errors"
	"testing"

	domainhealth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/health"
)

type fakeDatabase struct {
	err error
}

func (f fakeDatabase) Check(context.Context) error {
	return f.err
}

func TestCheckerMapsDatabaseStateToReport(t *testing.T) {
	tests := []struct {
		name         string
		databaseErr  error
		wantStatus   domainhealth.Status
		wantDatabase domainhealth.Status
	}{
		{
			name:         "readable database reports ok",
			wantStatus:   domainhealth.StatusOK,
			wantDatabase: domainhealth.StatusOK,
		},
		{
			name:         "unreachable database degrades the service",
			databaseErr:  errors.New("database is closed"),
			wantStatus:   domainhealth.StatusDegraded,
			wantDatabase: domainhealth.StatusUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checker := NewChecker(fakeDatabase{err: test.databaseErr})

			report := checker.Check(context.Background())

			if report.Status != test.wantStatus {
				t.Errorf("status: got %q, want %q", report.Status, test.wantStatus)
			}
			if report.Database != test.wantDatabase {
				t.Errorf("database: got %q, want %q", report.Database, test.wantDatabase)
			}
		})
	}
}
