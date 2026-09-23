package produk

import (
	"context"
	"sort"
	"strings"

	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// fakeProducts is an in-memory ProductRepository. The use cases are tested
// against this instead of SQLite: the port is the seam (ADR-0007).
type fakeProducts struct {
	products map[int64]domainproduk.Product
	nextID   int64
	// err, when set, is returned by every method — for error propagation.
	err error
}

func newFakeProducts() *fakeProducts {
	return &fakeProducts{products: map[int64]domainproduk.Product{}}
}

// seed inserts a Produk directly, bypassing the use cases, and returns its id.
func (f *fakeProducts) seed(product domainproduk.Product) int64 {
	f.nextID++
	product.ID = f.nextID
	f.products[product.ID] = product

	return product.ID
}

func (f *fakeProducts) Create(_ context.Context, product domainproduk.Product) (domainproduk.Product, error) {
	if f.err != nil {
		return domainproduk.Product{}, f.err
	}

	if product.Code != nil {
		if _, err := f.FindByCode(context.Background(), *product.Code); err == nil {
			return domainproduk.Product{}, domainproduk.ErrCodeTaken
		}
	}

	f.nextID++
	product.ID = f.nextID
	f.products[product.ID] = product

	return product, nil
}

func (f *fakeProducts) Update(_ context.Context, id int64, edit domainproduk.ProductEdit) (domainproduk.Product, error) {
	if f.err != nil {
		return domainproduk.Product{}, f.err
	}

	product, ok := f.products[id]
	if !ok {
		return domainproduk.Product{}, domainproduk.ErrProductNotFound
	}

	// Only the editable fields move. Stok, Active and Sold are not in an edit at
	// all — exactly as the SQL UPDATE leaves their columns alone (ADR-0014).
	product.Name = edit.Name
	product.Code = edit.Code
	product.Price = edit.Price
	product.Category = edit.Category
	f.products[id] = product

	return product, nil
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

func (f *fakeProducts) FindByCode(_ context.Context, code string) (domainproduk.Product, error) {
	if f.err != nil {
		return domainproduk.Product{}, f.err
	}

	for _, product := range f.products {
		if product.Code != nil && *product.Code == code {
			return product, nil
		}
	}

	return domainproduk.Product{}, domainproduk.ErrProductNotFound
}

func (f *fakeProducts) List(_ context.Context, filter domainproduk.Filter) ([]domainproduk.Product, error) {
	if f.err != nil {
		return nil, f.err
	}

	products := []domainproduk.Product{}
	for id := int64(1); id <= f.nextID; id++ {
		product, ok := f.products[id]
		if !ok {
			continue
		}
		if !matches(product, filter) {
			continue
		}
		products = append(products, product)
	}

	return products, nil
}

// AddStock mirrors the SQL `stock = stock + ?`: it adds to whatever is stored
// rather than replacing it, and reports a Produk that is not there.
func (f *fakeProducts) AddStock(_ context.Context, id int64, quantity int64) (domainproduk.Product, error) {
	if f.err != nil {
		return domainproduk.Product{}, f.err
	}

	product, ok := f.products[id]
	if !ok {
		return domainproduk.Product{}, domainproduk.ErrProductNotFound
	}

	product.Stock += quantity
	f.products[id] = product

	return product, nil
}

// ListLowStock mirrors the SQL the real repository builds: Active Produk below
// the threshold, thinnest first.
func (f *fakeProducts) ListLowStock(_ context.Context, threshold int64) ([]domainproduk.Product, error) {
	if f.err != nil {
		return nil, f.err
	}

	low := []domainproduk.Product{}
	for id := int64(1); id <= f.nextID; id++ {
		product, ok := f.products[id]
		if !ok || !product.Active || product.Stock >= threshold {
			continue
		}
		low = append(low, product)
	}

	sort.Slice(low, func(i, j int) bool {
		if low[i].Stock != low[j].Stock {
			return low[i].Stock < low[j].Stock
		}

		return low[i].Name < low[j].Name
	})

	return low, nil
}

func (f *fakeProducts) Categories(context.Context) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}

	seen := map[string]bool{}
	for _, product := range f.products {
		if product.Category != nil {
			seen[*product.Category] = true
		}
	}

	categories := make([]string, 0, len(seen))
	for category := range seen {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	return categories, nil
}

func (f *fakeProducts) SetActive(_ context.Context, id int64, active bool) error {
	if f.err != nil {
		return f.err
	}

	product, ok := f.products[id]
	if !ok {
		return domainproduk.ErrProductNotFound
	}

	product.Active = active
	f.products[id] = product

	return nil
}

func (f *fakeProducts) Delete(_ context.Context, id int64) error {
	if f.err != nil {
		return f.err
	}

	if _, ok := f.products[id]; !ok {
		return domainproduk.ErrProductNotFound
	}

	delete(f.products, id)

	return nil
}

// matches mirrors the SQL WHERE clauses the real repository builds, so the fake
// and SQLite answer a filter the same way.
func matches(product domainproduk.Product, filter domainproduk.Filter) bool {
	if filter.Name != "" && !strings.Contains(strings.ToLower(product.Name), strings.ToLower(filter.Name)) {
		return false
	}
	if filter.Code != "" {
		if product.Code == nil || !strings.Contains(strings.ToLower(*product.Code), strings.ToLower(filter.Code)) {
			return false
		}
	}
	if filter.Category != "" {
		if product.Category == nil || *product.Category != filter.Category {
			return false
		}
	}
	if filter.Active != nil && product.Active != *filter.Active {
		return false
	}

	return true
}

// newTestService wires a Service over fresh fakes.
func newTestService(products *fakeProducts) *Service {
	return NewService(products)
}
