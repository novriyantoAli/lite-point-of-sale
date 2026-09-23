// Package penjualan holds the Penjualan domain: a finished sale (checkout) with
// its Item snapshots and one Pembayaran. It declares the ports the checkout use
// case needs without depending on SQLite or HTTP (ADR-0004).
//
// The Go identifiers are English (Sale, SaleItem, SaleRepository) while the
// domain term, the route and the frontend types stay Indonesian (Penjualan,
// Item, /api/penjualan), the boundary ADR-0012 records.
package penjualan

import (
	"context"
	"errors"
)

// PaymentMethod is the Pembayaran method of CONTEXT.md: Tunai, QRIS, Debit or
// Transfer. The string values are what the API and the frontend schemas use;
// keep them stable.
//
// All four are declared here even though only Tunai is accepted by checkout
// today: the method of a stored Penjualan has to stay readable after #7 lands
// non-tunai Pembayaran, and the column's CHECK constraint is written once.
type PaymentMethod string

const (
	// PaymentCash is Tunai: physical money. An amount above the total produces
	// a Kembalian (CONTEXT.md, Tunai, Kembalian).
	PaymentCash PaymentMethod = "cash"
	// PaymentQRIS, PaymentDebit and PaymentTransfer are recorded only — no
	// gateway is involved (CONTEXT.md, Pembayaran).
	PaymentQRIS     PaymentMethod = "qris"
	PaymentDebit    PaymentMethod = "debit"
	PaymentTransfer PaymentMethod = "transfer"
)

// Valid reports whether m is a method the domain knows about.
func (m PaymentMethod) Valid() bool {
	switch m {
	case PaymentCash, PaymentQRIS, PaymentDebit, PaymentTransfer:
		return true
	default:
		return false
	}
}

// Item is one line of a Penjualan: one Produk, an integer quantity, and the
// name and price as they stood at checkout (CONTEXT.md, Item).
//
// Name and Price are copies, not a reference to the Produk: a later rename or
// reprice must not rewrite the history of a sale that already happened.
type Item struct {
	ProductID int64
	Name      string
	Price     int64
	Quantity  int64
}

// Subtotal is what this line contributed to the Penjualan total.
func (i Item) Subtotal() int64 {
	return i.Price * i.Quantity
}

// Payment is the one Pembayaran of a Penjualan: its method and its nominal. For
// Tunai, Amount is what the buyer handed over and Change is what was handed
// back; for a non-tunai method Change is zero (CONTEXT.md, Pembayaran).
type Payment struct {
	Method PaymentMethod
	Amount int64
	Change int64
}

// Sale is a finished Penjualan. It is final: nothing in this app voids, refunds
// or reopens one (CONTEXT.md, Penjualan).
//
// CreatedAt is a store-local timestamp (`YYYY-MM-DD HH:MM:SS`), not a canonical
// instant. There is one store and one terminal (ADR-0002), and both the daily
// revenue report (#9) and the printed Struk (#8) are read in store-local time —
// a UTC instant would have to be converted back, and would read as yesterday on
// a Struk printed early in the morning.
type Sale struct {
	// ReceiptNumber is the Nomor Struk: a global, unique, never-reused sequence
	// that does not reset daily (CONTEXT.md, Nomor Struk). It is zero on the way
	// in and assigned by the repository inside the checkout transaction.
	ReceiptNumber int64
	CreatedAt     string
	// CashierID is the Pengguna who rang it up; CashierName is their username as
	// it stood at checkout, because that is the name printed on the Struk and a
	// later rename must not change a Struk that was already handed over.
	CashierID   int64
	CashierName string
	Total       int64
	Items       []Item
	Payment     Payment
}

// Sentinel errors the use cases return and the HTTP adapter maps to status
// codes. Keep the set small and meaningful.
var (
	ErrSaleNotFound      = errors.New("sale not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidInput      = errors.New("invalid input")
)

// SaleRepository is the outbound port for Penjualan persistence. Implemented in
// the SQLite adapter; the use cases never see SQL (ADR-0004).
type SaleRepository interface {
	// Create records a finished Penjualan in one transaction: it assigns the
	// Nomor Struk, stores the Item snapshots and the Pembayaran, decrements each
	// Item's Stok and marks those Produk as sold. Either all of it is stored or
	// none of it is — a half-written sale would leave the Stok and the history
	// disagreeing (CONTEXT.md, Penjualan).
	//
	// The returned Sale is the stored one, with the Nomor Struk the repository
	// assigned. A Stok that ran out between the checkout's own check and this
	// write is answered with ErrInsufficientStock and stores nothing.
	Create(ctx context.Context, sale Sale) (Sale, error)

	// FindByReceiptNumber answers one stored Penjualan, its Items in the order
	// they were rung up. It is how the till reads a sale back, and what a
	// reprint (#8) and the sales list (#9) build on.
	FindByReceiptNumber(ctx context.Context, receiptNumber int64) (Sale, error)
}
