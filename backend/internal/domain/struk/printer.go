package struk

import (
	"context"
	"errors"
)

// Printer is the outbound port for the thermal printer. It takes the lines a
// Struk is made of, not ESC/POS bytes: the domain says *what* is printed and the
// adapter translates that to the bytes the device understands (ADR-0017,
// keputusan 2). Implemented in the escpos adapter; the use cases never see a
// device path.
type Printer interface {
	// Print sends the lines to the printer, one Struk per call. A printer that
	// cannot be reached is an error the caller reports to the Kasir — never a
	// panic and never a silent success.
	Print(ctx context.Context, lines []string) error
}

// ErrPrinterNotConfigured is what the null printer answers: no printer device is
// configured, so there is nothing to print to.
//
// It is a print *failure* — the Kasir sees "Printer belum diatur." and can try
// again — rather than a crash or a success nobody can tell apart from a real
// print (ADR-0017, keputusan 2).
var ErrPrinterNotConfigured = errors.New("printer belum diatur")
