package sqlite_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	adaptersqlite "github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/sqlite"
	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/sqlite"
)

// newProductRepository opens a migrated SQLite file and returns a repository
// over it with the raw handle, so a test can plant state the API cannot reach
// yet — `sold`, which only a finished Penjualan (#6) will set. This is the
// adapter integration test of ADR-0007: the repository runs against the real
// database, no fake.
func newProductRepository(t *testing.T) (*adaptersqlite.ProductRepository, *sql.DB, context.Context) {
	t.Helper()

	ctx := context.Background()

	db, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "pos.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return adaptersqlite.NewProductRepository(db), db, ctx
}

func productPtr(value string) *string { return &value }

// seedProduct inserts a Produk through the repository, failing the test on error.
func seedProduct(t *testing.T, repository *adaptersqlite.ProductRepository, ctx context.Context, product domainproduk.Product) domainproduk.Product {
	t.Helper()

	created, err := repository.Create(ctx, product)
	if err != nil {
		t.Fatalf("seed Produk %q: %v", product.Name, err)
	}

	return created
}

func TestProductRepositoryStoresAndReadsBackAProduk(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	created := seedProduct(t, repository, ctx, domainproduk.Product{
		Name:     "Kopi Susu",
		Code:     productPtr("KOPI-01"),
		Price:    18000,
		Category: productPtr("Minuman"),
		Stock:    10,
		Active:   true,
	})
	if created.ID == 0 {
		t.Error("create: got id 0, want the id SQLite assigned")
	}

	byID, err := repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	byCode, err := repository.FindByCode(ctx, "KOPI-01")
	if err != nil {
		t.Fatalf("find by code: %v", err)
	}

	for name, product := range map[string]domainproduk.Product{"by id": byID, "by code": byCode} {
		if product.ID != created.ID {
			t.Errorf("%s: got id %d, want %d", name, product.ID, created.ID)
		}
		if product.Name != "Kopi Susu" {
			t.Errorf("%s: got name %q, want %q", name, product.Name, "Kopi Susu")
		}
		if product.Price != 18000 {
			t.Errorf("%s: got price %d, want 18000", name, product.Price)
		}
		if product.Stock != 10 {
			t.Errorf("%s: got stock %d, want 10", name, product.Stock)
		}
		if product.Code == nil || *product.Code != "KOPI-01" {
			t.Errorf("%s: got code %v, want KOPI-01", name, product.Code)
		}
		if product.Category == nil || *product.Category != "Minuman" {
			t.Errorf("%s: got category %v, want Minuman", name, product.Category)
		}
		if !product.Active {
			t.Errorf("%s: got nonaktif, want active", name)
		}
		if product.Sold {
			t.Errorf("%s: got sold, want no sales", name)
		}
	}
}

func TestProductRepositoryKeepsAnAbsentKodeAndKategoriAsNull(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	created := seedProduct(t, repository, ctx, domainproduk.Product{
		Name: "Air Mineral", Price: 3000, Active: true,
	})

	stored, err := repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}

	if stored.Code != nil {
		t.Errorf("code: got %q, want nil", *stored.Code)
	}
	if stored.Category != nil {
		t.Errorf("category: got %q, want nil", *stored.Category)
	}
}

func TestProductRepositoryAllowsManyProdukWithoutAKode(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	// SQLite's UNIQUE allows many NULLs, which is what makes an optional Kode
	// possible at all.
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Air", Price: 3000, Active: true})
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Teh", Price: 5000, Active: true})

	listed, err := repository.List(ctx, domainproduk.Filter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 2 {
		t.Errorf("list: got %d Produk, want 2", len(listed))
	}
}

func TestProductRepositoryKeepsKodeUnique(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	product := domainproduk.Product{Name: "Kopi", Code: productPtr("KOPI-01"), Price: 18000, Active: true}
	seedProduct(t, repository, ctx, product)

	// The use case checks for an existing Kode before inserting; this is the
	// race it cannot cover.
	_, err := repository.Create(ctx, domainproduk.Product{Name: "Kopi Besar", Code: productPtr("KOPI-01"), Price: 20000, Active: true})
	if !errors.Is(err, domainproduk.ErrCodeTaken) {
		t.Fatalf("create duplicate Kode: got error %v, want %v", err, domainproduk.ErrCodeTaken)
	}
}

func TestProductRepositoryUpdateReplacesTheEditableFieldsOnly(t *testing.T) {
	repository, db, ctx := newProductRepository(t)

	created := seedProduct(t, repository, ctx, domainproduk.Product{
		Name: "Kopi", Code: productPtr("KOPI-01"), Price: 18000, Active: false,
	})

	// Plant the state only a finished Penjualan will set (#6).
	if _, err := db.ExecContext(ctx, `UPDATE produk SET sold = 1 WHERE id = ?`, created.ID); err != nil {
		t.Fatalf("mark sold: %v", err)
	}

	updated, err := repository.Update(ctx, domainproduk.Product{
		ID:       created.ID,
		Name:     "Kopi Susu Gula Aren",
		Code:     productPtr("KOPI-02"),
		Price:    22000,
		Category: productPtr("Minuman"),
		Stock:    7,
		// Deliberately the wrong values: an edit must not be able to write them.
		Active: true,
		Sold:   false,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Kopi Susu Gula Aren" || updated.Price != 22000 || updated.Stock != 7 {
		t.Errorf("returned Produk: got %+v, want the new name, price and stock", updated)
	}

	stored, err := repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find after update: %v", err)
	}
	if stored.Active {
		t.Error("active: got active, want the stored Nonaktif state to survive an edit")
	}
	if !stored.Sold {
		t.Error("sold: got false, want the stored sales history to survive an edit")
	}
	if stored.Code == nil || *stored.Code != "KOPI-02" {
		t.Errorf("code: got %v, want KOPI-02", stored.Code)
	}
}

func TestProductRepositoryUpdateKeepsKodeUnique(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Kopi", Code: productPtr("KOPI-01"), Price: 18000, Active: true})
	other := seedProduct(t, repository, ctx, domainproduk.Product{Name: "Teh", Code: productPtr("TEH-01"), Price: 6000, Active: true})

	_, err := repository.Update(ctx, domainproduk.Product{
		ID: other.ID, Name: "Teh Manis", Code: productPtr("KOPI-01"), Price: 6000,
	})
	if !errors.Is(err, domainproduk.ErrCodeTaken) {
		t.Fatalf("update to a taken Kode: got error %v, want %v", err, domainproduk.ErrCodeTaken)
	}
}

func TestProductRepositoryListFilters(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Kopi Susu", Code: productPtr("KOPI-01"), Category: productPtr("Minuman"), Price: 18000, Active: true})
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Teh Manis", Code: productPtr("TEH-01"), Category: productPtr("Minuman"), Price: 6000, Active: false})
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Roti Bakar", Category: productPtr("Makanan"), Price: 15000, Active: true})

	activeOnly := true
	tests := []struct {
		name   string
		filter domainproduk.Filter
		want   []string
	}{
		{name: "no filter", filter: domainproduk.Filter{}, want: []string{"Kopi Susu", "Roti Bakar", "Teh Manis"}},
		{name: "name, case-insensitively", filter: domainproduk.Filter{Name: "kopi"}, want: []string{"Kopi Susu"}},
		{name: "kode", filter: domainproduk.Filter{Code: "TEH"}, want: []string{"Teh Manis"}},
		{name: "kategori", filter: domainproduk.Filter{Category: "Minuman"}, want: []string{"Kopi Susu", "Teh Manis"}},
		{name: "active only", filter: domainproduk.Filter{Active: &activeOnly}, want: []string{"Kopi Susu", "Roti Bakar"}},
		{
			name:   "kategori and active together",
			filter: domainproduk.Filter{Category: "Minuman", Active: &activeOnly},
			want:   []string{"Kopi Susu"},
		},
		{name: "no match", filter: domainproduk.Filter{Name: "Hantu"}, want: []string{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			listed, err := repository.List(ctx, test.filter)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(listed) != len(test.want) {
				t.Fatalf("list: got %d Produk, want %d", len(listed), len(test.want))
			}
			for i, name := range test.want {
				if listed[i].Name != name {
					t.Errorf("list[%d]: got %q, want %q", i, listed[i].Name, name)
				}
			}
		})
	}
}

func TestProductRepositoryListEscapesLikeWildcards(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	// `100%` and `A_B` are literal Kode, not patterns: without escaping, the
	// first would match everything and the second would match `A B`.
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Diskon", Code: productPtr("100%"), Price: 1000, Active: true})
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Lain", Code: productPtr("A B"), Price: 2000, Active: true})

	for _, test := range []struct {
		name string
		code string
		want int
	}{
		{name: "percent is literal", code: "100%", want: 1},
		{name: "underscore is literal", code: "A_B", want: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			listed, err := repository.List(ctx, domainproduk.Filter{Code: test.code})
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(listed) != test.want {
				t.Errorf("list: got %d Produk, want %d", len(listed), test.want)
			}
		})
	}
}

func TestProductRepositoryCategories(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	categories, err := repository.Categories(ctx)
	if err != nil {
		t.Fatalf("categories of an empty catalogue: %v", err)
	}
	if len(categories) != 0 {
		t.Errorf("categories of an empty catalogue: got %v, want none", categories)
	}

	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Kopi", Category: productPtr("Minuman"), Active: true})
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Teh", Category: productPtr("Minuman"), Active: true})
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Roti", Category: productPtr("Makanan"), Active: true})
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Air", Active: true})

	categories, err = repository.Categories(ctx)
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

func TestProductRepositorySetActiveAndDelete(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	created := seedProduct(t, repository, ctx, domainproduk.Product{Name: "Kopi", Price: 18000, Active: true})

	if err := repository.SetActive(ctx, created.ID, false); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	stored, err := repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find after deactivate: %v", err)
	}
	if stored.Active {
		t.Error("deactivate: got active, want nonaktif")
	}

	if err := repository.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repository.FindByID(ctx, created.ID); !errors.Is(err, domainproduk.ErrProductNotFound) {
		t.Errorf("find after delete: got error %v, want %v", err, domainproduk.ErrProductNotFound)
	}
}

func TestProductRepositoryAddStockAddsToTheStoredStok(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	created := seedProduct(t, repository, ctx, domainproduk.Product{Name: "Kopi", Price: 18000, Stock: 10, Active: true})

	updated, err := repository.AddStock(ctx, created.ID, 3)
	if err != nil {
		t.Fatalf("add Stok: %v", err)
	}
	if updated.Stock != 13 {
		t.Errorf("add Stok: got %d, want 13", updated.Stock)
	}

	// It adds rather than replaces: the Stok the row holds is what the next
	// restock — and the next sale — builds on.
	if _, err := repository.AddStock(ctx, created.ID, 2); err != nil {
		t.Fatalf("add Stok again: %v", err)
	}

	stored, err := repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find after restock: %v", err)
	}
	if stored.Stock != 15 {
		t.Errorf("stored stock: got %d, want 15", stored.Stock)
	}

	// The rest of the Produk comes back untouched.
	if stored.Name != "Kopi" || stored.Price != 18000 || !stored.Active {
		t.Errorf("stored Produk: got %+v, want the seeded name, price and status", stored)
	}
}

func TestProductRepositoryListLowStockAnswersActiveOnesThinnestFirst(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	habis := seedProduct(t, repository, ctx, domainproduk.Product{Name: "Habis", Stock: 0, Active: true})
	satu := seedProduct(t, repository, ctx, domainproduk.Product{Name: "Satu", Stock: 1, Active: true})
	menipis := seedProduct(t, repository, ctx, domainproduk.Product{Name: "Menipis", Stock: 2, Active: true})
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Di Ambang", Stock: 5, Active: true})
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Aman", Stock: 6, Active: true})
	seedProduct(t, repository, ctx, domainproduk.Product{Name: "Nonaktif", Stock: 0, Active: false})

	low, err := repository.ListLowStock(ctx, 5)
	if err != nil {
		t.Fatalf("low Stok: %v", err)
	}

	want := []int64{habis.ID, satu.ID, menipis.ID}
	if len(low) != len(want) {
		t.Fatalf("low Stok: got %d Produk, want %d", len(low), len(want))
	}
	for i, id := range want {
		if low[i].ID != id {
			t.Errorf("low Stok[%d]: got id %d (%s, stock %d), want %d", i, low[i].ID, low[i].Name, low[i].Stock, id)
		}
	}

	// The comparison is strict: at a threshold of 2 the Produk that sits exactly
	// on it is not below it, so only the two thinnest answer.
	belowTwo, err := repository.ListLowStock(ctx, 2)
	if err != nil {
		t.Fatalf("low Stok at two: %v", err)
	}
	if len(belowTwo) != 2 || belowTwo[0].ID != habis.ID || belowTwo[1].ID != satu.ID {
		t.Errorf("low Stok at two: got %+v, want the Produk with stock 0 and 1", belowTwo)
	}

	// An empty answer is an empty list, not a nil one: the API answers `[]`, not
	// `null`, and the UI reads a list either way.
	empty, err := repository.ListLowStock(ctx, 0)
	if err != nil {
		t.Fatalf("low Stok below zero: %v", err)
	}
	if empty == nil || len(empty) != 0 {
		t.Errorf("low Stok below zero: got %v, want an empty list", empty)
	}
}

func TestProductRepositoryReportsAMissingProduk(t *testing.T) {
	repository, _, ctx := newProductRepository(t)

	if _, err := repository.FindByID(ctx, 404); !errors.Is(err, domainproduk.ErrProductNotFound) {
		t.Errorf("find by id: got error %v, want %v", err, domainproduk.ErrProductNotFound)
	}
	if _, err := repository.FindByCode(ctx, "HANTU"); !errors.Is(err, domainproduk.ErrProductNotFound) {
		t.Errorf("find by code: got error %v, want %v", err, domainproduk.ErrProductNotFound)
	}
	if _, err := repository.Update(ctx, domainproduk.Product{ID: 404, Name: "Hantu"}); !errors.Is(err, domainproduk.ErrProductNotFound) {
		t.Errorf("update: got error %v, want %v", err, domainproduk.ErrProductNotFound)
	}
	if _, err := repository.AddStock(ctx, 404, 1); !errors.Is(err, domainproduk.ErrProductNotFound) {
		t.Errorf("add Stok: got error %v, want %v", err, domainproduk.ErrProductNotFound)
	}
	if err := repository.SetActive(ctx, 404, false); !errors.Is(err, domainproduk.ErrProductNotFound) {
		t.Errorf("set active: got error %v, want %v", err, domainproduk.ErrProductNotFound)
	}
	if err := repository.Delete(ctx, 404); !errors.Is(err, domainproduk.ErrProductNotFound) {
		t.Errorf("delete: got error %v, want %v", err, domainproduk.ErrProductNotFound)
	}
}
