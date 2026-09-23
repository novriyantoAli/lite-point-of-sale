package penjualan

import (
	"context"
	"errors"

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

// newTestService wires a Service over fresh fakes.
func newTestService(products *fakeProducts, sales *fakeSales) *Service {
	return NewService(products, sales)
}
