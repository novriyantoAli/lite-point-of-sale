package produk

import (
	"context"
	"errors"
	"testing"

	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// seedStock plants a Produk with a Stok and a Status, bypassing the use cases,
// so a test can start from the state it wants to reason about.
func seedStock(products *fakeProducts, name string, stock int64, active bool) int64 {
	return products.seed(domainproduk.Product{
		Name:   name,
		Price:  1000,
		Stock:  stock,
		Active: active,
	})
}

func TestAddStockIncreasesTheStok(t *testing.T) {
	products := newFakeProducts()
	id := seedStock(products, "Kopi", 10, true)

	updated, err := newTestService(products).AddStock(context.Background(), id, 3)
	if err != nil {
		t.Fatalf("add Stok: %v", err)
	}

	if updated.Stock != 13 {
		t.Errorf("stock: got %d, want 13", updated.Stock)
	}

	// The answer is not just a number the use case computed: it is the Stok that
	// is now stored, which is what the next request will read.
	stored, err := products.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if stored.Stock != 13 {
		t.Errorf("stored stock: got %d, want 13", stored.Stock)
	}
}

func TestAddStockAddsToWhatIsAlreadyThere(t *testing.T) {
	products := newFakeProducts()
	id := seedStock(products, "Kopi", 4, true)

	service := newTestService(products)
	for i := 1; i <= 3; i++ {
		if _, err := service.AddStock(context.Background(), id, 2); err != nil {
			t.Fatalf("add Stok %d: %v", i, err)
		}
	}

	stored, _ := products.FindByID(context.Background(), id)
	if stored.Stock != 10 {
		t.Errorf("stock: got %d, want 10 after three additions of 2 onto 4", stored.Stock)
	}
}

func TestAddStockValidatesTheQuantity(t *testing.T) {
	tests := []struct {
		name     string
		quantity int64
	}{
		{name: "zero", quantity: 0},
		{name: "negative", quantity: -5},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			products := newFakeProducts()
			id := seedStock(products, "Kopi", 10, true)

			_, err := newTestService(products).AddStock(context.Background(), id, test.quantity)

			if !errors.Is(err, domainproduk.ErrInvalidInput) {
				t.Fatalf("add Stok %d: got error %v, want %v", test.quantity, err, domainproduk.ErrInvalidInput)
			}

			// The HTTP adapter reads the message straight off this error, so it
			// must carry one.
			var inputErr InputError
			if !errors.As(err, &inputErr) || inputErr.Message == "" {
				t.Errorf("add Stok: got %v, want an InputError with a message", err)
			}

			stored, _ := products.FindByID(context.Background(), id)
			if stored.Stock != 10 {
				t.Errorf("stock: got %d, want the refused addition to leave 10", stored.Stock)
			}
		})
	}
}

func TestAddStockReportsAMissingProduk(t *testing.T) {
	products := newFakeProducts()

	_, err := newTestService(products).AddStock(context.Background(), 404, 5)

	if !errors.Is(err, domainproduk.ErrProductNotFound) {
		t.Fatalf("add Stok to a missing Produk: got error %v, want %v", err, domainproduk.ErrProductNotFound)
	}
}

// A Produk that is not for sale yet is exactly the one an Admin restocks before
// reactivating it, so the Stok of a Nonaktif Produk is still the Admin's to add
// to (CONTEXT.md, Nonaktif).
func TestAddStockIsAllowedOnANonaktifProduk(t *testing.T) {
	products := newFakeProducts()
	id := seedStock(products, "Teh", 0, false)

	updated, err := newTestService(products).AddStock(context.Background(), id, 6)
	if err != nil {
		t.Fatalf("add Stok to a Nonaktif Produk: %v", err)
	}

	if updated.Stock != 6 {
		t.Errorf("stock: got %d, want 6", updated.Stock)
	}
	if updated.Active {
		t.Error("active: got true, want adding Stok to leave the Produk Nonaktif")
	}
}

func TestAddStockPropagatesARepositoryFailure(t *testing.T) {
	products := newFakeProducts()
	id := seedStock(products, "Kopi", 1, true)
	products.err = errors.New("database is gone")

	_, err := newTestService(products).AddStock(context.Background(), id, 1)

	if err == nil {
		t.Fatal("add Stok: got no error, want the repository failure")
	}
}

func TestLowStockAnswersTheThinnestActiveProdukFirst(t *testing.T) {
	products := newFakeProducts()
	habis := seedStock(products, "Habis", 0, true)
	menipis := seedStock(products, "Menipis", 2, true)
	// The ambang itself is the first Stok that is still enough, so the Produk
	// sitting exactly on it is not menipis (issue #5: "di bawah ambang atau nol").
	seedStock(products, "Di Ambang", domainproduk.LowStockThreshold, true)
	seedStock(products, "Aman", domainproduk.LowStockThreshold+1, true)
	seedStock(products, "Nonaktif", 0, false)

	low, err := newTestService(products).LowStock(context.Background())
	if err != nil {
		t.Fatalf("low Stok: %v", err)
	}

	want := []int64{habis, menipis}
	if len(low) != len(want) {
		t.Fatalf("low Stok: got %d Produk, want %d", len(low), len(want))
	}
	for i, id := range want {
		if low[i].ID != id {
			t.Errorf("low Stok[%d]: got id %d, want %d (thinnest first)", i, low[i].ID, id)
		}
	}
}

// The threshold is what the UI shows the Admin ("Stok menipis ≤ 5"), so the one
// that decides the list is the one that must be answered.
func TestLowStockUsesTheThresholdItAnswers(t *testing.T) {
	products := newFakeProducts()
	seedStock(products, "Kopi", 1, true)

	low, err := newTestService(products).LowStock(context.Background())
	if err != nil {
		t.Fatalf("low Stok: %v", err)
	}
	if len(low) != 1 {
		t.Fatalf("low Stok: got %d Produk, want 1", len(low))
	}

	if domainproduk.LowStockThreshold <= 0 {
		t.Errorf("LowStockThreshold: got %d, want a positive threshold", domainproduk.LowStockThreshold)
	}
}

func TestLowStockPropagatesARepositoryFailure(t *testing.T) {
	products := newFakeProducts()
	products.err = errors.New("database is gone")

	_, err := newTestService(products).LowStock(context.Background())

	if err == nil {
		t.Fatal("low Stok: got no error, want the repository failure")
	}
}
