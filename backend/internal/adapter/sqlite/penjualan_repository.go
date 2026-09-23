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
