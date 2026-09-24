package penjualan

import (
	"context"
	"errors"
	"testing"
	"time"

	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
)

// fakeSaleDay is the store-local day the fake repository stamps every sale with,
// so a report test can ask for that day and get exactly the sales it made.
const fakeSaleDay = "2026-09-23"

// dayFilter is the report filter every test of a made-up day uses.
func dayFilter() domainpenjualan.ReportFilter {
	return domainpenjualan.ReportFilter{Date: fakeSaleDay}
}

func TestDailyRevenueFillsEveryMethodInTheTillsOrder(t *testing.T) {
	service := newTestService(newFakeProducts(kopi()), newFakeSales())

	saleOf(t, service, tunai(1, 1, 18000))
	saleOf(t, service, nonTunai("qris", 1, 1, 18000))

	report, err := service.DailyRevenue(context.Background(), dayFilter())
	if err != nil {
		t.Fatalf("daily revenue: %v", err)
	}

	if report.Total != 36000 || report.Transactions != 2 {
		t.Errorf("totals: got %d/%d, want 36000/2", report.Total, report.Transactions)
	}
	if report.Date != fakeSaleDay {
		t.Errorf("date: got %q, want %q", report.Date, fakeSaleDay)
	}

	// All four methods, in the order the till offers them — a method nobody used
	// is a zero row rather than a missing one, so the screen never invents it.
	want := []domainpenjualan.MethodTotal{
		{Method: domainpenjualan.PaymentCash, Total: 18000, Transactions: 1},
		{Method: domainpenjualan.PaymentQRIS, Total: 18000, Transactions: 1},
		{Method: domainpenjualan.PaymentDebit},
		{Method: domainpenjualan.PaymentTransfer},
	}
	if len(report.ByMethod) != len(want) {
		t.Fatalf("by method: got %d rows, want %d", len(report.ByMethod), len(want))
	}
	for i, total := range want {
		if report.ByMethod[i] != total {
			t.Errorf("by method %d: got %+v, want %+v", i, report.ByMethod[i], total)
		}
	}
}

func TestDailyRevenueAnswersTodayWhenNoDateIsGiven(t *testing.T) {
	service := newTestService(newFakeProducts(kopi()), newFakeSales())

	report, err := service.DailyRevenue(context.Background(), domainpenjualan.ReportFilter{})
	if err != nil {
		t.Fatalf("daily revenue: %v", err)
	}

	// Store-local time, the same clock `created_at` is written with (ADR-0015).
	if want := time.Now().Format(time.DateOnly); report.Date != want {
		t.Errorf("date: got %q, want today %q", report.Date, want)
	}
}

func TestReportsRefuseAMalformedDate(t *testing.T) {
	service := newTestService(newFakeProducts(kopi()), newFakeSales())

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "daily revenue",
			run: func() error {
				_, err := service.DailyRevenue(context.Background(), domainpenjualan.ReportFilter{Date: "23-09-2026"})
				return err
			},
		},
		{
			name: "sales list",
			run: func() error {
				_, err := service.ListSales(context.Background(), domainpenjualan.ReportFilter{Date: "kemarin"})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			if !errors.Is(err, domainpenjualan.ErrInvalidInput) {
				t.Fatalf("got %v, want an error that unwraps to ErrInvalidInput", err)
			}

			var input InputError
			if !errors.As(err, &input) || input.Message == "" {
				t.Errorf("got %v, want a message the screen can show", err)
			}
		})
	}
}

func TestListSalesAnswersTheDaysSalesNewestFirst(t *testing.T) {
	service := newTestService(newFakeProducts(kopi()), newFakeSales())
	first := saleOf(t, service, tunai(1, 1, 18000))
	second := saleOf(t, service, tunai(1, 1, 18000))

	sales, err := service.ListSales(context.Background(), dayFilter())
	if err != nil {
		t.Fatalf("sales list: %v", err)
	}

	if len(sales) != 2 {
		t.Fatalf("sales: got %d, want 2", len(sales))
	}
	if sales[0].ReceiptNumber != second.ReceiptNumber || sales[1].ReceiptNumber != first.ReceiptNumber {
		t.Errorf("order: got %d then %d, want newest first",
			sales[0].ReceiptNumber, sales[1].ReceiptNumber)
	}
	if sales[0].Method != domainpenjualan.PaymentCash || sales[0].CashierName != kasir.Username {
		t.Errorf("row: got %+v, want the cashier and method of the sale", sales[0])
	}
}

func TestReportsAnswerAnEmptyDay(t *testing.T) {
	service := newTestService(newFakeProducts(kopi()), newFakeSales())

	sales, err := service.ListSales(context.Background(), dayFilter())
	if err != nil {
		t.Fatalf("sales list: %v", err)
	}
	if len(sales) != 0 {
		t.Errorf("sales: got %d, want none", len(sales))
	}

	report, err := service.DailyRevenue(context.Background(), dayFilter())
	if err != nil {
		t.Fatalf("daily revenue: %v", err)
	}
	if report.Total != 0 || report.Transactions != 0 {
		t.Errorf("totals: got %d/%d, want 0/0", report.Total, report.Transactions)
	}
	// Even a day with no sales answers the four methods, so the screen has a
	// complete breakdown to render.
	if len(report.ByMethod) != len(methodOrder) {
		t.Errorf("by method: got %d rows, want %d", len(report.ByMethod), len(methodOrder))
	}
}

func TestReportsSurfaceARepositoryFailure(t *testing.T) {
	sales := newFakeSales()
	sales.reportErr = errors.New("database is gone")
	service := newTestService(newFakeProducts(kopi()), sales)

	if _, err := service.DailyRevenue(context.Background(), dayFilter()); err == nil {
		t.Error("daily revenue: got no error, want the repository's failure")
	}
	if _, err := service.ListSales(context.Background(), dayFilter()); err == nil {
		t.Error("sales list: got no error, want the repository's failure")
	}
}
