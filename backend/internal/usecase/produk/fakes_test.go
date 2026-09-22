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

func (f *fakeProducts) Update(_ context.Context, product domainproduk.Product) (domainproduk.Product, error) {
	if f.err != nil {
		return domainproduk.Product{}, f.err
	}

	existing, ok := f.products[product.ID]
	if !ok {
		return domainproduk.Product{}, domainproduk.ErrProductNotFound
	}

	// Active and Sold are not part of an edit: the fake keeps them, exactly as
	// the SQL UPDATE does.
	product.Active = existing.Active
	product.Sold = existing.Sold
	f.products[product.ID] = product

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
