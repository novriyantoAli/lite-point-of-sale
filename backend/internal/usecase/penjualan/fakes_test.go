package penjualan

import (
	"context"
	"errors"

	domainpengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/pengaturan"
	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// errNotUsed is what the fake catalogue answers with if a checkout reaches for a
// method it has no business calling. The checkout only ever reads a Produk by id;
// the rest of the port is there so the type satisfies it.
var errNotUsed = errors.New("the checkout must not call this ProductRepository method")

// fakeProducts is an in-memory ProductRepository. The use cases are tested
// against this instead of SQLite: the port is the seam (ADR-0007).
type fakeProducts struct {
	products map[int64]domainproduk.Product
	// err, when set, is returned by FindByID — for error propagation.
	err error
}

func newFakeProducts(products ...domainproduk.Product) *fakeProducts {
	fake := &fakeProducts{products: map[int64]domainproduk.Product{}}
	for _, product := range products {
		fake.products[product.ID] = product
	}

	return fake
}

func (f *fakeProducts) FindByID(_ context.Context, id int64) (domainproduk.Product, error) {
	if f.err != nil {
		return domainproduk.Product{}, f.err
	}

	product, ok := f.products[id]
	if !ok {
		return domainproduk.Product{}, domainproduk.ErrProductNotFound
	}

	return product, nil
}

func (f *fakeProducts) Create(context.Context, domainproduk.Product) (domainproduk.Product, error) {
	return domainproduk.Product{}, errNotUsed
}

func (f *fakeProducts) Update(context.Context, int64, domainproduk.ProductEdit) (domainproduk.Product, error) {
	return domainproduk.Product{}, errNotUsed
}

func (f *fakeProducts) FindByCode(context.Context, string) (domainproduk.Product, error) {
	return domainproduk.Product{}, errNotUsed
}

func (f *fakeProducts) List(context.Context, domainproduk.Filter) ([]domainproduk.Product, error) {
	return nil, errNotUsed
}

func (f *fakeProducts) AddStock(context.Context, int64, int64) (domainproduk.Product, error) {
	return domainproduk.Product{}, errNotUsed
}

func (f *fakeProducts) ListLowStock(context.Context, int64) ([]domainproduk.Product, error) {
	return nil, errNotUsed
}

func (f *fakeProducts) Categories(context.Context) ([]string, error) { return nil, errNotUsed }

func (f *fakeProducts) SetActive(context.Context, int64, bool) error { return errNotUsed }

func (f *fakeProducts) Delete(context.Context, int64) error { return errNotUsed }

// fakeSales is an in-memory SaleRepository. It records what the checkout handed
// it and assigns the Nomor Struk the way the SQL repository does, so a use case
// test can assert on the sale that would be stored without a database.
type fakeSales struct {
	recorded []domainpenjualan.Sale
	next     int64
	// err, when set, is returned by Create — for error propagation.
	err error
	// reportErr, when set, is returned by ListSales and DailyRevenue — for the
	// path where the report cannot be read.
	reportErr error
}

func newFakeSales() *fakeSales {
	return &fakeSales{}
}

func (f *fakeSales) Create(_ context.Context, sale domainpenjualan.Sale) (domainpenjualan.Sale, error) {
	if f.err != nil {
		return domainpenjualan.Sale{}, f.err
	}

	f.next++
	sale.ReceiptNumber = f.next
	// The real repository reads the timestamp back from the database; the fake
	// states one, so a test can tell the sale that came out of Create from the one
	// that went in.
	sale.CreatedAt = "2026-09-23 10:00:00"

	f.recorded = append(f.recorded, sale)

	return sale, nil
}

func (f *fakeSales) FindByReceiptNumber(_ context.Context, receiptNumber int64) (domainpenjualan.Sale, error) {
	for _, sale := range f.recorded {
		if sale.ReceiptNumber == receiptNumber {
			return sale, nil
		}
	}

	return domainpenjualan.Sale{}, domainpenjualan.ErrSaleNotFound
}

// ListSales answers the sales of the filter's day, newest first — the same order
// the SQL repository answers in.
func (f *fakeSales) ListSales(_ context.Context, filter domainpenjualan.ReportFilter) ([]domainpenjualan.SaleSummary, error) {
	if f.reportErr != nil {
		return nil, f.reportErr
	}

	sales := []domainpenjualan.SaleSummary{}
	for i := len(f.recorded) - 1; i >= 0; i-- {
		sale := f.recorded[i]
		if !f.onDay(sale, filter) {
			continue
		}
		sales = append(sales, domainpenjualan.SaleSummary{
			ReceiptNumber: sale.ReceiptNumber,
			CreatedAt:     sale.CreatedAt,
			CashierID:     sale.CashierID,
			CashierName:   sale.CashierName,
			Total:         sale.Total,
			Method:        sale.Payment.Method,
		})
	}

	return sales, nil
}

// DailyRevenue aggregates the recorded sales the way the SQL repository does:
// only the methods that appear, with the rule that every method is answered left
// to the use case.
func (f *fakeSales) DailyRevenue(_ context.Context, filter domainpenjualan.ReportFilter) (domainpenjualan.DailyRevenue, error) {
	if f.reportErr != nil {
		return domainpenjualan.DailyRevenue{}, f.reportErr
	}

	report := domainpenjualan.DailyRevenue{Date: filter.Date}
	methods := map[domainpenjualan.PaymentMethod]*domainpenjualan.MethodTotal{}
	cashiers := map[int64]*domainpenjualan.CashierTotal{}

	for _, sale := range f.recorded {
		if !f.onDay(sale, filter) {
			continue
		}

		report.Total += sale.Total
		report.Transactions++

		total, ok := methods[sale.Payment.Method]
		if !ok {
			total = &domainpenjualan.MethodTotal{Method: sale.Payment.Method}
			methods[sale.Payment.Method] = total
		}
		total.Total += sale.Total
		total.Transactions++

		cashier, ok := cashiers[sale.CashierID]
		if !ok {
			cashier = &domainpenjualan.CashierTotal{CashierID: sale.CashierID, CashierName: sale.CashierName}
			cashiers[sale.CashierID] = cashier
		}
		cashier.CashierName = sale.CashierName
		cashier.Total += sale.Total
		cashier.Transactions++
	}

	report.ByMethod = []domainpenjualan.MethodTotal{}
	for _, total := range methods {
		report.ByMethod = append(report.ByMethod, *total)
	}

	report.ByCashier = []domainpenjualan.CashierTotal{}
	for _, total := range cashiers {
		report.ByCashier = append(report.ByCashier, *total)
	}

	return report, nil
}

// onDay reports whether a recorded sale falls on the filter's day. The fake
// stamps every sale with `fakeCreatedAt`, so a test picks the day it wants by
// naming that date; an empty filter matches everything, which is what a test of
// the aggregation itself wants.
func (f *fakeSales) onDay(sale domainpenjualan.Sale, filter domainpenjualan.ReportFilter) bool {
	return filter.Date == "" || sale.CreatedAt[:10] == filter.Date
}

// fakeSettings is an in-memory ReceiptSettings: the store's one Pengaturan row,
// which is where the Struk template and paper width are read from. The use cases
// are tested against this instead of SQLite (ADR-0007).
type fakeSettings struct {
	settings domainpengaturan.Settings
	// err, when set, is returned by Get — for the path where the template cannot
	// be read at all.
	err error
}

func newFakeSettings() *fakeSettings {
	return &fakeSettings{settings: domainpengaturan.Settings{ID: 1, PaperWidth: 80}}
}

func (f *fakeSettings) Get(context.Context) (domainpengaturan.Settings, error) {
	if f.err != nil {
		return domainpengaturan.Settings{}, f.err
	}

	return f.settings, nil
}

// fakePrinter records the lines it was asked to print, so a use case test can
// assert on what the Struk said without a device (ADR-0017, keputusan 2).
type fakePrinter struct {
	printed [][]string
	// err, when set, is returned by Print — the printer that is missing or broken.
	err error
}

func (f *fakePrinter) Print(_ context.Context, lines []string) error {
	if f.err != nil {
		return f.err
	}

	f.printed = append(f.printed, lines)

	return nil
}

// newTestService wires a Service over fresh fakes, with a printer that works and
// the seeded Pengaturan. Tests that need to steer the printer or the settings
// build the Service themselves.
func newTestService(products *fakeProducts, sales *fakeSales) *Service {
	return NewService(products, sales, newFakeSettings(), &fakePrinter{})
}
