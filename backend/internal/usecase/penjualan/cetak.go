package penjualan

import (
	"context"
	"errors"

	domainpengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/pengaturan"
	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
	domainstruk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/struk"
)

// ReceiptSettings is the one setting a Struk is printed from: the store's
// Pengaturan row, which holds the template blocks and the paper width. It is a
// port of this package rather than an import of the pengaturan *service*, so the
// dependency still points inward (ADR-0004, ADR-0017).
type ReceiptSettings interface {
	Get(ctx context.Context) (domainpengaturan.Settings, error)
}

// PrintResult is the outcome of printing one Struk: whether the paper came out,
// and, when it did not, the reason a Kasir can read. Both print paths answer this
// same shape (ADR-0017, keputusan 5).
type PrintResult struct {
	Printed bool
	// Message is empty on success, and on failure is written for the person at the
	// till rather than for a log.
	Message string
}

// CheckoutResult is what a checkout answers: the stored Penjualan, and the result
// of the Struk print that followed it.
type CheckoutResult struct {
	Sale  domainpenjualan.Sale
	Print PrintResult
}

// CetakStruk prints the Struk of one stored Penjualan. It is both the reprint
// path and what the panel after a checkout retries with, so the two produce the
// same result shape.
//
// The sale is read back by its Nomor Struk, and the template is the store's
// *current* Pengaturan — not a copy from when the sale happened. A Struk is the
// store's document, and a shop that renamed itself prints its new name on a
// receipt it reissues (ADR-0017, keputusan 5).
func (s *Service) CetakStruk(ctx context.Context, receiptNumber int64) (PrintResult, error) {
	sale, err := s.sales.FindByReceiptNumber(ctx, receiptNumber)
	if err != nil {
		return PrintResult{}, err
	}

	return s.printStruk(ctx, sale)
}

// printStruk composes the Struk of a sale and hands it to the printer.
//
// A printer that fails is answered as a PrintResult with Printed false, never as
// an error: the Penjualan is already stored and the money has moved, so a Struk
// that did not come out must never look like a sale that did not happen
// (ADR-0017, keputusan 1). An error is reserved for a failure that stopped the
// Struk being composed at all — a Pengaturan that could not be read.
func (s *Service) printStruk(ctx context.Context, sale domainpenjualan.Sale) (PrintResult, error) {
	settings, err := s.settings.Get(ctx)
	if err != nil {
		return PrintResult{}, err
	}

	lines := domainstruk.Compose(sale, domainstruk.Template{
		Header:     settings.Header,
		Footer:     settings.Footer,
		PaperWidth: settings.PaperWidth,
	})

	if err := s.printer.Print(ctx, lines); err != nil {
		return PrintResult{Printed: false, Message: printFailureMessage(err)}, nil
	}

	return PrintResult{Printed: true}, nil
}

// printFailureMessage is what the Kasir reads when no Struk came out. A printer
// nobody configured says so in those words, because that is the fix; any other
// failure answers with what the Kasir can act on and keeps the device detail out
// of the screen.
func printFailureMessage(err error) string {
	if errors.Is(err, domainstruk.ErrPrinterNotConfigured) {
		return "Printer belum diatur."
	}

	return "Struk gagal dicetak. Periksa printer lalu coba lagi."
}

// composeFailureMessage is what the Kasir reads when the Struk could not even be
// composed — the store's Pengaturan could not be read. It deliberately does not
// blame the printer: nothing was ever sent to one.
const composeFailureMessage = "Struk gagal dicetak. Coba lagi."
