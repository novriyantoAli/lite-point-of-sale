package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
)

// SaleRepository stores Penjualan in SQLite. It satisfies
// domain/penjualan.SaleRepository, so no SQL leaves this package (ADR-0004).
type SaleRepository struct {
	db *sql.DB
}

// NewSaleRepository returns a repository backed by the given database handle.
func NewSaleRepository(db *sql.DB) *SaleRepository {
	return &SaleRepository{db: db}
}

// Create satisfies domain/penjualan.SaleRepository. Everything a checkout has
// to change — the Nomor Struk, the Penjualan, its Item snapshots, the Stok of
// each Produk and the fact that those Produk have now sold — happens in one
// transaction, so a refusal anywhere in it leaves the store exactly as it was
// (CONTEXT.md, Penjualan).
func (r *SaleRepository) Create(ctx context.Context, sale domainpenjualan.Sale) (domainpenjualan.Sale, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domainpenjualan.Sale{}, fmt.Errorf("begin checkout: %w", err)
	}
	// Rollback after a successful commit is a no-op, so this one defer covers
	// every early return below.
	defer tx.Rollback()

	// The Nomor Struk is assigned here, in the same transaction that stores the
	// sale: one store, one terminal, one write connection (ADR-0002), so this
	// read and the insert below cannot interleave with another checkout. The
	// UNIQUE constraint on the column is the backstop.
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(receipt_number), 0) + 1 FROM penjualan`).Scan(&sale.ReceiptNumber); err != nil {
		return domainpenjualan.Sale{}, fmt.Errorf("assign Nomor Struk: %w", err)
	}

	result, err := tx.ExecContext(ctx,
		`INSERT INTO penjualan (receipt_number, cashier_id, cashier_name, total, method, amount, change_amount)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sale.ReceiptNumber, sale.CashierID, sale.CashierName, sale.Total,
		sale.Payment.Method, sale.Payment.Amount, sale.Payment.Change)
	if err != nil {
		return domainpenjualan.Sale{}, fmt.Errorf("insert Penjualan %d: %w", sale.ReceiptNumber, err)
	}

	saleID, err := result.LastInsertId()
	if err != nil {
		return domainpenjualan.Sale{}, fmt.Errorf("read id of Penjualan %d: %w", sale.ReceiptNumber, err)
	}

	for _, item := range sale.Items {
		// The Stok leaves in the same statement that guards it: `stock >= ?`
		// means a Produk whose Stok ran out between the checkout's own check and
		// this write affects no row, and `active = 1` means a Produk deactivated
		// in between is not sold either. Either way the transaction is rolled
		// back, so the Stok never goes negative and no sale is recorded.
		result, err := tx.ExecContext(ctx,
			`UPDATE produk SET stock = stock - ?, sold = 1 WHERE id = ? AND active = 1 AND stock >= ?`,
			item.Quantity, item.ProductID, item.Quantity)
		if err != nil {
			return domainpenjualan.Sale{}, fmt.Errorf("reduce Stok of Produk %d: %w", item.ProductID, err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return domainpenjualan.Sale{}, fmt.Errorf("count reduced Produk %d: %w", item.ProductID, err)
		}
		if affected == 0 {
			return domainpenjualan.Sale{}, domainpenjualan.ErrInsufficientStock
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO penjualan_item (penjualan_id, product_id, name, price, quantity)
			 VALUES (?, ?, ?, ?, ?)`,
			saleID, item.ProductID, item.Name, item.Price, item.Quantity); err != nil {
			return domainpenjualan.Sale{}, fmt.Errorf("insert Item of Penjualan %d: %w", sale.ReceiptNumber, err)
		}
	}

	// created_at is the database's own clock, so it is read back rather than
	// guessed in Go: the answer has to be the timestamp that was stored.
	if err := tx.QueryRowContext(ctx,
		`SELECT created_at FROM penjualan WHERE id = ?`, saleID).Scan(&sale.CreatedAt); err != nil {
		return domainpenjualan.Sale{}, fmt.Errorf("read created_at of Penjualan %d: %w", sale.ReceiptNumber, err)
	}

	if err := tx.Commit(); err != nil {
		return domainpenjualan.Sale{}, fmt.Errorf("commit Penjualan %d: %w", sale.ReceiptNumber, err)
	}

	return sale, nil
}

// FindByReceiptNumber satisfies domain/penjualan.SaleRepository. The Items come
// back in the order they were rung up, which is the order the Struk prints them.
func (r *SaleRepository) FindByReceiptNumber(ctx context.Context, receiptNumber int64) (domainpenjualan.Sale, error) {
	var (
		sale   domainpenjualan.Sale
		saleID int64
		method string
	)

	err := r.db.QueryRowContext(ctx,
		`SELECT id, receipt_number, created_at, cashier_id, cashier_name, total, method, amount, change_amount
		 FROM penjualan WHERE receipt_number = ?`, receiptNumber).
		Scan(&saleID, &sale.ReceiptNumber, &sale.CreatedAt, &sale.CashierID, &sale.CashierName,
			&sale.Total, &method, &sale.Payment.Amount, &sale.Payment.Change)
	if errors.Is(err, sql.ErrNoRows) {
		return domainpenjualan.Sale{}, domainpenjualan.ErrSaleNotFound
	}
	if err != nil {
		return domainpenjualan.Sale{}, fmt.Errorf("read Penjualan %d: %w", receiptNumber, err)
	}

	sale.Payment.Method = domainpenjualan.PaymentMethod(method)

	rows, err := r.db.QueryContext(ctx,
		`SELECT product_id, name, price, quantity FROM penjualan_item WHERE penjualan_id = ? ORDER BY id`,
		saleID)
	if err != nil {
		return domainpenjualan.Sale{}, fmt.Errorf("list Item of Penjualan %d: %w", receiptNumber, err)
	}
	defer rows.Close()

	sale.Items = []domainpenjualan.Item{}
	for rows.Next() {
		var item domainpenjualan.Item
		if err := rows.Scan(&item.ProductID, &item.Name, &item.Price, &item.Quantity); err != nil {
			return domainpenjualan.Sale{}, fmt.Errorf("read Item of Penjualan %d: %w", receiptNumber, err)
		}
		sale.Items = append(sale.Items, item)
	}
	if err := rows.Err(); err != nil {
		return domainpenjualan.Sale{}, fmt.Errorf("read Item rows of Penjualan %d: %w", receiptNumber, err)
	}

	return sale, nil
}

// ListSales satisfies domain/penjualan.SaleRepository. The rows are the list
// view of a day, newest Nomor Struk first, and they carry no Item lines: a list
// that read the Items of every sale would be a query per row for a screen that
// never shows them.
func (r *SaleRepository) ListSales(ctx context.Context, filter domainpenjualan.ReportFilter) ([]domainpenjualan.SaleSummary, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT receipt_number, created_at, cashier_id, cashier_name, total, method
		 FROM penjualan WHERE date(created_at) = ? ORDER BY receipt_number DESC`, filter.Date)
	if err != nil {
		return nil, fmt.Errorf("list Penjualan of %s: %w", filter.Date, err)
	}
	defer rows.Close()

	sales := []domainpenjualan.SaleSummary{}
	for rows.Next() {
		var (
			summary domainpenjualan.SaleSummary
			method  string
		)
		if err := rows.Scan(&summary.ReceiptNumber, &summary.CreatedAt, &summary.CashierID,
			&summary.CashierName, &summary.Total, &method); err != nil {
			return nil, fmt.Errorf("read Penjualan row of %s: %w", filter.Date, err)
		}
		summary.Method = domainpenjualan.PaymentMethod(method)
		sales = append(sales, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read Penjualan rows of %s: %w", filter.Date, err)
	}

	return sales, nil
}

// DailyRevenue satisfies domain/penjualan.SaleRepository. It runs three
// aggregates over the day's Penjualan — the totals, the split by method and the
// split by Kasir — and leaves the rule that every method is answered (a zero row
// for one nobody used) to the use case, where it is testable without a database.
//
// The three run inside one read transaction. A checkout landing between them
// would otherwise leave the total disagreeing with the breakdowns it is supposed
// to be the sum of, and the Admin reading the report is exactly who the Kasir is
// selling alongside. SQLite reads in WAL mode, so the transaction holds a
// consistent snapshot without blocking the till's write.
//
// The cashier breakdown groups by Pengguna id rather than by the copied name, so
// one Kasir is one row. The name is the one snapshotted on their sales; a
// Pengguna cannot be renamed in this app, so the MAX is a formality that keeps
// the query valid rather than a choice between names.
func (r *SaleRepository) DailyRevenue(ctx context.Context, filter domainpenjualan.ReportFilter) (domainpenjualan.DailyRevenue, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domainpenjualan.DailyRevenue{}, fmt.Errorf("begin omzet report of %s: %w", filter.Date, err)
	}
	// Rollback after a successful commit is a no-op, so this one defer covers
	// every early return below.
	defer tx.Rollback()

	report := domainpenjualan.DailyRevenue{Date: filter.Date}

	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(total), 0) FROM penjualan WHERE date(created_at) = ?`, filter.Date).
		Scan(&report.Transactions, &report.Total); err != nil {
		return domainpenjualan.DailyRevenue{}, fmt.Errorf("total omzet of %s: %w", filter.Date, err)
	}

	methods, err := tx.QueryContext(ctx,
		`SELECT method, COUNT(*), COALESCE(SUM(total), 0)
		 FROM penjualan WHERE date(created_at) = ? GROUP BY method`, filter.Date)
	if err != nil {
		return domainpenjualan.DailyRevenue{}, fmt.Errorf("omzet per method of %s: %w", filter.Date, err)
	}
	defer methods.Close()

	report.ByMethod = []domainpenjualan.MethodTotal{}
	for methods.Next() {
		var (
			total  domainpenjualan.MethodTotal
			method string
		)
		if err := methods.Scan(&method, &total.Transactions, &total.Total); err != nil {
			return domainpenjualan.DailyRevenue{}, fmt.Errorf("read omzet per method of %s: %w", filter.Date, err)
		}
		total.Method = domainpenjualan.PaymentMethod(method)
		report.ByMethod = append(report.ByMethod, total)
	}
	if err := methods.Err(); err != nil {
		return domainpenjualan.DailyRevenue{}, fmt.Errorf("read omzet per method rows of %s: %w", filter.Date, err)
	}

	cashiers, err := tx.QueryContext(ctx,
		`SELECT cashier_id, MAX(cashier_name) AS name, COUNT(*) AS sales, COALESCE(SUM(total), 0) AS revenue
		 FROM penjualan WHERE date(created_at) = ? GROUP BY cashier_id ORDER BY revenue DESC, name`, filter.Date)
	if err != nil {
		return domainpenjualan.DailyRevenue{}, fmt.Errorf("omzet per Kasir of %s: %w", filter.Date, err)
	}
	defer cashiers.Close()

	report.ByCashier = []domainpenjualan.CashierTotal{}
	for cashiers.Next() {
		var total domainpenjualan.CashierTotal
		if err := cashiers.Scan(&total.CashierID, &total.CashierName, &total.Transactions, &total.Total); err != nil {
			return domainpenjualan.DailyRevenue{}, fmt.Errorf("read omzet per Kasir of %s: %w", filter.Date, err)
		}
		report.ByCashier = append(report.ByCashier, total)
	}
	if err := cashiers.Err(); err != nil {
		return domainpenjualan.DailyRevenue{}, fmt.Errorf("read omzet per Kasir rows of %s: %w", filter.Date, err)
	}

	if err := tx.Commit(); err != nil {
		return domainpenjualan.DailyRevenue{}, fmt.Errorf("commit omzet report of %s: %w", filter.Date, err)
	}

	return report, nil
}
