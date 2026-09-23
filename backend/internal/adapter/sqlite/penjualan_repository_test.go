package sqlite_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	adaptersqlite "github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/sqlite"
	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/sqlite"
)

// newSaleRepository opens a migrated SQLite file and returns a Penjualan
// repository over it, with the Produk repository and the raw handle a test needs
// to plant state and to read it back. This is the adapter integration test of
// ADR-0007: the repository runs against the real database, no fake.
func newSaleRepository(t *testing.T) (*adaptersqlite.SaleRepository, *adaptersqlite.ProductRepository, *sql.DB, context.Context) {
	t.Helper()

	ctx := context.Background()

	db, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "pos.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// The Pengguna every Penjualan references. Inserted with raw SQL because this
	// test is about the sale store, not about how a Pengguna is created.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO pengguna (username, password_hash, role) VALUES ('admin', 'x', 'admin')`); err != nil {
		t.Fatalf("seed Pengguna: %v", err)
	}

	return adaptersqlite.NewSaleRepository(db), adaptersqlite.NewProductRepository(db), db, ctx
}

// saleOf builds a sale of one Produk, the way the checkout use case hands one to
// the repository: the name and the price are already snapshotted.
func saleOf(product domainproduk.Product, quantity int64) domainpenjualan.Sale {
	item := domainpenjualan.Item{
		ProductID: product.ID,
		Name:      product.Name,
		Price:     product.Price,
		Quantity:  quantity,
	}

	return domainpenjualan.Sale{
		CashierID:   1,
		CashierName: "admin",
		Total:       item.Subtotal(),
		Items:       []domainpenjualan.Item{item},
		Payment: domainpenjualan.Payment{
			Method: domainpenjualan.PaymentCash,
			Amount: item.Subtotal(),
		},
	}
}

// countRows answers how many rows one table holds, so a test can show that a
// rolled-back checkout stored nothing.
func countRows(t *testing.T, db *sql.DB, ctx context.Context, table string) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+table).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}

	return count
}

// storedStock answers a Produk's Stok straight from the table, so a test does not
// depend on the Produk repository to read what the sale repository wrote.
func storedStock(t *testing.T, db *sql.DB, ctx context.Context, id int64) int64 {
	t.Helper()

	var stock int64
	if err := db.QueryRowContext(ctx, `SELECT stock FROM produk WHERE id = ?`, id).Scan(&stock); err != nil {
		t.Fatalf("read Stok of Produk %d: %v", id, err)
	}

	return stock
}

func TestSaleRepositoryStoresAndReadsBackASale(t *testing.T) {
	sales, products, db, ctx := newSaleRepository(t)
	kopi := seedProduct(t, products, ctx, domainproduk.Product{Name: "Kopi Susu", Price: 18000, Stock: 10, Active: true})

	created, err := sales.Create(ctx, saleOf(kopi, 2))
	if err != nil {
		t.Fatalf("create Penjualan: %v", err)
	}

	if created.ReceiptNumber != 1 {
		t.Errorf("Nomor Struk: got %d, want the first one to be 1", created.ReceiptNumber)
	}
	if created.CreatedAt == "" {
		t.Error("created_at: got nothing, want the timestamp that was stored")
	}

	// The Stok left in the same statement that guarded it.
	if stock := storedStock(t, db, ctx, kopi.ID); stock != 8 {
		t.Errorf("Stok: got %d, want 8", stock)
	}

	stored, err := sales.FindByReceiptNumber(ctx, created.ReceiptNumber)
	if err != nil {
		t.Fatalf("find by Nomor Struk: %v", err)
	}

	if stored.Total != 36000 {
		t.Errorf("total: got %d, want 36000", stored.Total)
	}
	if stored.CashierName != "admin" {
		t.Errorf("cashier: got %q, want %q", stored.CashierName, "admin")
	}
	if stored.Payment.Method != domainpenjualan.PaymentCash || stored.Payment.Amount != 36000 {
		t.Errorf("payment: got %+v, want a Tunai Pembayaran of 36000", stored.Payment)
	}
	if len(stored.Items) != 1 || stored.Items[0].Name != "Kopi Susu" || stored.Items[0].Price != 18000 {
		t.Errorf("items: got %+v, want the snapshot that was stored", stored.Items)
	}
}

func TestSaleRepositoryKeepsTheItemOrder(t *testing.T) {
	sales, products, _, ctx := newSaleRepository(t)
	kopi := seedProduct(t, products, ctx, domainproduk.Product{Name: "Kopi", Price: 18000, Stock: 10, Active: true})
	teh := seedProduct(t, products, ctx, domainproduk.Product{Name: "Teh", Price: 6000, Stock: 10, Active: true})

	sale := saleOf(kopi, 1)
	tehItem := domainpenjualan.Item{ProductID: teh.ID, Name: teh.Name, Price: teh.Price, Quantity: 1}
	sale.Items = append(sale.Items, tehItem)
	sale.Total += tehItem.Subtotal()
	sale.Payment.Amount = sale.Total

	created, err := sales.Create(ctx, sale)
	if err != nil {
		t.Fatalf("create Penjualan: %v", err)
	}

	stored, err := sales.FindByReceiptNumber(ctx, created.ReceiptNumber)
	if err != nil {
		t.Fatalf("find by Nomor Struk: %v", err)
	}

	// The order the Kasir rang them up is the order the Struk prints.
	if len(stored.Items) != 2 || stored.Items[0].Name != "Kopi" || stored.Items[1].Name != "Teh" {
		t.Errorf("items: got %+v, want Kopi then Teh", stored.Items)
	}
}

func TestSaleRepositoryAssignsReceiptNumbersInOrder(t *testing.T) {
	sales, products, _, ctx := newSaleRepository(t)
	kopi := seedProduct(t, products, ctx, domainproduk.Product{Name: "Kopi", Price: 18000, Stock: 10, Active: true})

	first, err := sales.Create(ctx, saleOf(kopi, 1))
	if err != nil {
		t.Fatalf("first checkout: %v", err)
	}
	second, err := sales.Create(ctx, saleOf(kopi, 1))
	if err != nil {
		t.Fatalf("second checkout: %v", err)
	}

	if second.ReceiptNumber != first.ReceiptNumber+1 {
		t.Errorf("Nomor Struk: got %d then %d, want a sequence that keeps counting",
			first.ReceiptNumber, second.ReceiptNumber)
	}
}

func TestSaleRepositoryMarksTheProdukAsSold(t *testing.T) {
	sales, products, db, ctx := newSaleRepository(t)
	kopi := seedProduct(t, products, ctx, domainproduk.Product{Name: "Kopi", Price: 18000, Stock: 2, Active: true})

	if _, err := sales.Create(ctx, saleOf(kopi, 1)); err != nil {
		t.Fatalf("create Penjualan: %v", err)
	}

	var sold bool
	if err := db.QueryRowContext(ctx, `SELECT sold FROM produk WHERE id = ?`, kopi.ID).Scan(&sold); err != nil {
		t.Fatalf("read sold: %v", err)
	}
	if !sold {
		t.Error("sold: got false, want a Produk that was part of a Penjualan to be marked sold")
	}
}

// TestSaleRepositoryRollsBackWhenAnItemIsShort is the atomicity check at the
// adapter seam: a checkout whose second Item cannot be covered must leave the
// first Item's Stok alone and store no Penjualan and no Item. The use case
// checks the Stok first, so this is the guard for the race between that check
// and this write — and the only place it can be exercised deterministically.
func TestSaleRepositoryRollsBackWhenAnItemIsShort(t *testing.T) {
	sales, products, db, ctx := newSaleRepository(t)
	enough := seedProduct(t, products, ctx, domainproduk.Product{Name: "Kopi", Price: 18000, Stock: 5, Active: true})
	short := seedProduct(t, products, ctx, domainproduk.Product{Name: "Teh", Price: 6000, Stock: 1, Active: true})

	sale := saleOf(enough, 3)
	shortItem := domainpenjualan.Item{ProductID: short.ID, Name: short.Name, Price: short.Price, Quantity: 2}
	sale.Items = append(sale.Items, shortItem)
	sale.Total += shortItem.Subtotal()
	sale.Payment.Amount = sale.Total

	_, err := sales.Create(ctx, sale)
	if !errors.Is(err, domainpenjualan.ErrInsufficientStock) {
		t.Fatalf("create with a short Item: got %v, want ErrInsufficientStock", err)
	}

	if stock := storedStock(t, db, ctx, enough.ID); stock != 5 {
		t.Errorf("Stok of the Item that was covered: got %d, want 5 — the checkout was not atomic", stock)
	}
	if stock := storedStock(t, db, ctx, short.ID); stock != 1 {
		t.Errorf("Stok of the Item that was short: got %d, want 1", stock)
	}
	if count := countRows(t, db, ctx, "penjualan"); count != 0 {
		t.Errorf("Penjualan rows: got %d, want none", count)
	}
	if count := countRows(t, db, ctx, "penjualan_item"); count != 0 {
		t.Errorf("Item rows: got %d, want none", count)
	}

	// And the Nomor Struk the rolled-back checkout would have taken is still free.
	next, err := sales.Create(ctx, saleOf(enough, 1))
	if err != nil {
		t.Fatalf("checkout after the rollback: %v", err)
	}
	if next.ReceiptNumber != 1 {
		t.Errorf("Nomor Struk after the rollback: got %d, want 1", next.ReceiptNumber)
	}
}

func TestSaleRepositoryRefusesAProdukThatCannotBeSold(t *testing.T) {
	sales, products, db, ctx := newSaleRepository(t)
	nonaktif := seedProduct(t, products, ctx, domainproduk.Product{Name: "Teh", Price: 6000, Stock: 5, Active: false})

	_, err := sales.Create(ctx, saleOf(nonaktif, 1))
	if !errors.Is(err, domainpenjualan.ErrInsufficientStock) {
		t.Fatalf("create with a Nonaktif Produk: got %v, want ErrInsufficientStock", err)
	}

	if stock := storedStock(t, db, ctx, nonaktif.ID); stock != 5 {
		t.Errorf("Stok of the Nonaktif Produk: got %d, want 5", stock)
	}
	if count := countRows(t, db, ctx, "penjualan"); count != 0 {
		t.Errorf("Penjualan rows: got %d, want none", count)
	}
}

func TestSaleRepositoryRefusesAProdukThatIsGone(t *testing.T) {
	sales, _, db, ctx := newSaleRepository(t)

	_, err := sales.Create(ctx, domainpenjualan.Sale{
		CashierID:   1,
		CashierName: "admin",
		Total:       18000,
		Items:       []domainpenjualan.Item{{ProductID: 404, Name: "Hantu", Price: 18000, Quantity: 1}},
		Payment:     domainpenjualan.Payment{Method: domainpenjualan.PaymentCash, Amount: 18000},
	})
	if !errors.Is(err, domainpenjualan.ErrInsufficientStock) {
		t.Fatalf("create with a Produk that is gone: got %v, want ErrInsufficientStock", err)
	}

	if count := countRows(t, db, ctx, "penjualan"); count != 0 {
		t.Errorf("Penjualan rows: got %d, want none", count)
	}
}

func TestFindByReceiptNumberReportsAMissingSale(t *testing.T) {
	sales, _, _, ctx := newSaleRepository(t)

	_, err := sales.FindByReceiptNumber(ctx, 404)
	if !errors.Is(err, domainpenjualan.ErrSaleNotFound) {
		t.Fatalf("find a missing Penjualan: got %v, want ErrSaleNotFound", err)
	}
}
