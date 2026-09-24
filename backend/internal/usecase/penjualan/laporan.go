package penjualan

import (
	"context"
	"time"

	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
)

// dateOnly is the shape a report day is written in: a store-local calendar date,
// the way `created_at` is stored (ADR-0015). It is named here because every
// report date in this slice is one, and a second format would be a second answer
// to "what day is this".
const dateOnly = time.DateOnly

// methodOrder is the Pembayaran methods in the order the till offers them
// (CONTEXT.md): Tunai first, then the three that are recorded. A daily report
// always answers all four, in this order, so the screen never invents a row.
var methodOrder = []domainpenjualan.PaymentMethod{
	domainpenjualan.PaymentCash,
	domainpenjualan.PaymentQRIS,
	domainpenjualan.PaymentDebit,
	domainpenjualan.PaymentTransfer,
}

// DailyRevenue answers the omzet of one store-local day (#9): the total, how many
// Penjualan made it, and the breakdown by Pembayaran method and by Kasir. An
// empty date means today, in store-local time — the same clock `created_at` is
// written with (ADR-0015), so "today" is the day the till is living in.
//
// The methods nobody used are filled in as zero rows rather than left out: the
// breakdown is a domain rule ("one entry per method, in the till's order"), and
// leaving it to the screen would be a second place for it to drift.
func (s *Service) DailyRevenue(ctx context.Context, filter domainpenjualan.ReportFilter) (domainpenjualan.DailyRevenue, error) {
	filter, err := reportFilter(filter)
	if err != nil {
		return domainpenjualan.DailyRevenue{}, err
	}

	report, err := s.sales.DailyRevenue(ctx, filter)
	if err != nil {
		return domainpenjualan.DailyRevenue{}, err
	}

	report.Date = filter.Date
	report.ByMethod = withEveryMethod(report.ByMethod)

	return report, nil
}

// ListSales answers the Penjualan of one store-local day, newest Nomor Struk
// first: the sales list of #9. An empty date means today, as in DailyRevenue.
func (s *Service) ListSales(ctx context.Context, filter domainpenjualan.ReportFilter) ([]domainpenjualan.SaleSummary, error) {
	filter, err := reportFilter(filter)
	if err != nil {
		return nil, err
	}

	return s.sales.ListSales(ctx, filter)
}

// reportFilter validates the day of a report and fills in today when it was left
// empty. A date that is not `YYYY-MM-DD` is refused rather than guessed at: the
// comparison against `created_at` is textual, and a malformed day would silently
// match nothing instead of saying so.
func reportFilter(filter domainpenjualan.ReportFilter) (domainpenjualan.ReportFilter, error) {
	if filter.Date == "" {
		filter.Date = time.Now().Format(dateOnly)
		return filter, nil
	}

	date, err := time.Parse(dateOnly, filter.Date)
	if err != nil || date.Format(dateOnly) != filter.Date {
		return domainpenjualan.ReportFilter{}, InputError{Message: "Tanggal laporan tidak valid."}
	}

	return filter, nil
}

// withEveryMethod fills in the methods a day had no sales for, so the breakdown
// is always the four methods of CONTEXT.md in the till's order.
func withEveryMethod(totals []domainpenjualan.MethodTotal) []domainpenjualan.MethodTotal {
	byMethod := make(map[domainpenjualan.PaymentMethod]domainpenjualan.MethodTotal, len(totals))
	for _, total := range totals {
		byMethod[total.Method] = total
	}

	ordered := make([]domainpenjualan.MethodTotal, 0, len(methodOrder))
	for _, method := range methodOrder {
		total, ok := byMethod[method]
		if !ok {
			total = domainpenjualan.MethodTotal{Method: method}
		}
		ordered = append(ordered, total)
	}

	return ordered
}
