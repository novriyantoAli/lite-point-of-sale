package produk

import (
	"context"
	"errors"
	"testing"

	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// ptr is a small helper for the optional fields of a Produk.
func ptr(value string) *string { return &value }

func validInput() ProductInput {
	return ProductInput{
		Name:     "Kopi Susu",
		Code:     "KOPI-01",
		Price:    18000,
		Category: "Minuman",
		Stock:    10,
	}
}

func TestCreateValidatesInput(t *testing.T) {
	tests := []struct {
		name  string
		input ProductInput
	}{
		{name: "empty name", input: ProductInput{Name: "", Price: 1000}},
		{name: "name of spaces only", input: ProductInput{Name: "   ", Price: 1000}},
		{name: "negative price", input: ProductInput{Name: "Kopi", Price: -1}},
		{name: "negative stock", input: ProductInput{Name: "Kopi", Price: 1000, Stock: -1}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			products := newFakeProducts()

			_, err := newTestService(products).Create(context.Background(), test.input)

			if !errors.Is(err, domainproduk.ErrInvalidInput) {
				t.Fatalf("create: got error %v, want %v", err, domainproduk.ErrInvalidInput)
			}

			// The HTTP adapter reads the message straight off this error, so it
			// must carry one.
			var inputErr InputError
			if !errors.As(err, &inputErr) || inputErr.Message == "" {
				t.Errorf("create: got %v, want an InputError with a message", err)
			}

			if stored, _ := products.List(context.Background(), domainproduk.Filter{}); len(stored) != 0 {
				t.Errorf("create: stored %d Produk, want none", len(stored))
			}
		})
	}
}

func TestCreateTrimsTextAndKeepsAnAbsentKodeAbsent(t *testing.T) {
	products := newFakeProducts()

	created, err := newTestService(products).Create(context.Background(), ProductInput{
		Name:     "  Kopi Susu ",
		Code:     "   ",
		Price:    18000,
		Category: "  ",
		Stock:    0,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if created.Name != "Kopi Susu" {
		t.Errorf("name: got %q, want %q", created.Name, "Kopi Susu")
	}
	if created.Code != nil {
		t.Errorf("code: got %q, want nil for a Produk without a Kode", *created.Code)
	}
	if created.Category != nil {
		t.Errorf("category: got %q, want nil for a Produk without a Kategori", *created.Category)
	}
	if !created.Active {
		t.Error("active: got false, want a new Produk to be active")
	}
	if created.Sold {
		t.Error("sold: got true, want a new Produk to have no sales")
	}
	if created.Stock != 0 {
		t.Errorf("stock: got %d, want 0", created.Stock)
	}
}

func TestCreateRejectsADuplicateKode(t *testing.T) {
	products := newFakeProducts()
	products.seed(domainproduk.Product{Name: "Kopi Susu", Code: ptr("KOPI-01"), Active: true})

	_, err := newTestService(products).Create(context.Background(), ProductInput{
		Name:  "Kopi Susu Besar",
		Code:  "KOPI-01",
		Price: 20000,
	})

	if !errors.Is(err, domainproduk.ErrCodeTaken) {
		t.Fatalf("create: got error %v, want %v", err, domainproduk.ErrCodeTaken)
	}
}

func TestCreateAllowsManyProdukWithoutAKode(t *testing.T) {
	products := newFakeProducts()
	service := newTestService(products)

	for _, name := range []string{"Teh", "Air"} {
		if _, err := service.Create(context.Background(), ProductInput{Name: name, Price: 5000}); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	if stored, _ := products.List(context.Background(), domainproduk.Filter{}); len(stored) != 2 {
		t.Errorf("stored Produk: got %d, want 2", len(stored))
	}
}

func TestUpdateReplacesTheEditableFields(t *testing.T) {
	products := newFakeProducts()
	id := products.seed(domainproduk.Product{
		Name: "Kopi Susu", Code: ptr("KOPI-01"), Price: 18000, Active: true, Stock: 10,
	})

	updated, err := newTestService(products).Update(context.Background(), id, ProductInput{
		Name:     "Kopi Susu Gula Aren",
		Code:     "KOPI-02",
		Price:    22000,
		Category: "Minuman",
		Stock:    7,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if updated.Name != "Kopi Susu Gula Aren" || updated.Price != 22000 || updated.Stock != 7 {
		t.Errorf("updated: got %+v, want the new name, price and stock", updated)
	}
	if updated.Code == nil || *updated.Code != "KOPI-02" {
		t.Errorf("code: got %v, want KOPI-02", updated.Code)
	}
	if updated.Category == nil || *updated.Category != "Minuman" {
		t.Errorf("category: got %v, want Minuman", updated.Category)
	}
}

func TestUpdateKeepsTheKodeAProdukAlreadyHas(t *testing.T) {
	products := newFakeProducts()
	id := products.seed(domainproduk.Product{Name: "Kopi Susu", Code: ptr("KOPI-01"), Active: true})

	updated, err := newTestService(products).Update(context.Background(), id, ProductInput{
		Name:  "Kopi Susu",
		Code:  "KOPI-01",
		Price: 19000,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if updated.Code == nil || *updated.Code != "KOPI-01" {
		t.Errorf("code: got %v, want KOPI-01", updated.Code)
	}
}

func TestUpdateRejectsAKodeAnotherProdukUses(t *testing.T) {
	products := newFakeProducts()
	products.seed(domainproduk.Product{Name: "Kopi Susu", Code: ptr("KOPI-01"), Active: true})
	otherID := products.seed(domainproduk.Product{Name: "Teh", Code: ptr("TEH-01"), Active: true})

	_, err := newTestService(products).Update(context.Background(), otherID, ProductInput{
		Name:  "Teh Manis",
		Code:  "KOPI-01",
		Price: 6000,
	})

	if !errors.Is(err, domainproduk.ErrCodeTaken) {
		t.Fatalf("update: got error %v, want %v", err, domainproduk.ErrCodeTaken)
	}
}

func TestUpdateRejectsAnUnknownProduk(t *testing.T) {
	products := newFakeProducts()

	_, err := newTestService(products).Update(context.Background(), 99, validInput())

	if !errors.Is(err, domainproduk.ErrProductNotFound) {
		t.Fatalf("update: got error %v, want %v", err, domainproduk.ErrProductNotFound)
	}
}

func TestUpdateNeverTouchesActiveOrSold(t *testing.T) {
	products := newFakeProducts()
	id := products.seed(domainproduk.Product{Name: "Kopi", Active: false, Sold: true, Price: 1000})

	updated, err := newTestService(products).Update(context.Background(), id, ProductInput{
		Name:  "Kopi Susu",
		Price: 2000,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if updated.Active {
		t.Error("active: got true, want the stored Nonaktif state to survive an edit")
	}
	if !updated.Sold {
		t.Error("sold: got false, want the stored sales history to survive an edit")
	}
}

func TestSetActiveTogglesTheProduk(t *testing.T) {
	products := newFakeProducts()
	id := products.seed(domainproduk.Product{Name: "Kopi", Active: true})
	service := newTestService(products)

	deactivated, err := service.SetActive(context.Background(), id, false)
	if err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if deactivated.Active {
		t.Error("deactivate: got active, want Nonaktif")
	}

	reactivated, err := service.SetActive(context.Background(), id, true)
	if err != nil {
		t.Fatalf("reactivate: %v", err)
	}
	if !reactivated.Active {
		t.Error("reactivate: got Nonaktif, want active")
	}
}

func TestSetActiveRejectsAnUnknownProduk(t *testing.T) {
	products := newFakeProducts()

	_, err := newTestService(products).SetActive(context.Background(), 99, false)

	if !errors.Is(err, domainproduk.ErrProductNotFound) {
		t.Fatalf("set active: got error %v, want %v", err, domainproduk.ErrProductNotFound)
	}
}

func TestDeleteRemovesAProdukThatNeverSold(t *testing.T) {
	products := newFakeProducts()
	id := products.seed(domainproduk.Product{Name: "Kopi", Active: true})

	if err := newTestService(products).Delete(context.Background(), id); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if stored, _ := products.List(context.Background(), domainproduk.Filter{}); len(stored) != 0 {
		t.Errorf("stored Produk: got %d, want none", len(stored))
	}
}

func TestDeleteRefusesAProdukThatSold(t *testing.T) {
	products := newFakeProducts()
	id := products.seed(domainproduk.Product{Name: "Kopi", Active: true, Sold: true})

	err := newTestService(products).Delete(context.Background(), id)

	if !errors.Is(err, domainproduk.ErrProductHasSales) {
		t.Fatalf("delete: got error %v, want %v", err, domainproduk.ErrProductHasSales)
	}
	if stored, _ := products.List(context.Background(), domainproduk.Filter{}); len(stored) != 1 {
		t.Errorf("stored Produk: got %d, want the Produk to survive", len(stored))
	}
}

func TestDeleteRejectsAnUnknownProduk(t *testing.T) {
	products := newFakeProducts()

	err := newTestService(products).Delete(context.Background(), 99)

	if !errors.Is(err, domainproduk.ErrProductNotFound) {
		t.Fatalf("delete: got error %v, want %v", err, domainproduk.ErrProductNotFound)
	}
}

func TestListPassesTheFilterToTheRepository(t *testing.T) {
	products := newFakeProducts()
	products.seed(domainproduk.Product{Name: "Kopi Susu", Code: ptr("KOPI-01"), Category: ptr("Minuman"), Active: true})
	products.seed(domainproduk.Product{Name: "Teh Manis", Code: ptr("TEH-01"), Category: ptr("Minuman"), Active: false})
	products.seed(domainproduk.Product{Name: "Roti Bakar", Category: ptr("Makanan"), Active: true})
	service := newTestService(products)

	activeOnly := true
	tests := []struct {
		name   string
		filter domainproduk.Filter
		want   int
	}{
		{name: "no filter", filter: domainproduk.Filter{}, want: 3},
		{name: "by name", filter: domainproduk.Filter{Name: "kopi"}, want: 1},
		{name: "by kode", filter: domainproduk.Filter{Code: "TEH"}, want: 1},
		{name: "by kategori", filter: domainproduk.Filter{Category: "Minuman"}, want: 2},
		{name: "active only", filter: domainproduk.Filter{Active: &activeOnly}, want: 2},
		{
			name:   "kategori and active together",
			filter: domainproduk.Filter{Category: "Minuman", Active: &activeOnly},
			want:   1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			listed, err := service.List(context.Background(), test.filter)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(listed) != test.want {
				t.Errorf("list: got %d Produk, want %d", len(listed), test.want)
			}
		})
	}
}

func TestCategoriesReturnsTheOnesInUse(t *testing.T) {
	products := newFakeProducts()
	products.seed(domainproduk.Product{Name: "Kopi", Category: ptr("Minuman")})
	products.seed(domainproduk.Product{Name: "Teh", Category: ptr("Minuman")})
	products.seed(domainproduk.Product{Name: "Roti", Category: ptr("Makanan")})
	products.seed(domainproduk.Product{Name: "Air"})

	categories, err := newTestService(products).Categories(context.Background())
	if err != nil {
		t.Fatalf("categories: %v", err)
	}

	want := []string{"Makanan", "Minuman"}
	if len(categories) != len(want) {
		t.Fatalf("categories: got %v, want %v", categories, want)
	}
	for i := range want {
		if categories[i] != want[i] {
			t.Errorf("categories[%d]: got %q, want %q", i, categories[i], want[i])
		}
	}
}
