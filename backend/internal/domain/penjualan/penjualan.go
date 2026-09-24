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
// All four are declared here — and named once in the column's CHECK constraint —
// because a stored Penjualan has to stay readable: the set is fixed rather than
// widened by a later migration. Checkout accepts all four; only Tunai produces a
// Kembalian.
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

// SaleSummary is one row of the sales list (#9): the Nomor Struk, when and by
// whom it was rung up, its total and how it was paid. It is deliberately not a
// Sale — the list shows no Item lines, and reading them per row would be a query
// per sale.
type SaleSummary struct {
	ReceiptNumber int64
	CreatedAt     string
	CashierID     int64
	CashierName   string
	Total         int64
	Method        PaymentMethod
}

// MethodTotal is what one Pembayaran method contributed to a store-local day:
// how many Penjualan and how much.
type MethodTotal struct {
	Method       PaymentMethod
	Total        int64
	Transactions int64
}

// CashierTotal is what one Kasir rang up in a store-local day. Attribution is by
// Pengguna id, so one Kasir is one row; CashierName is the name copied onto their
// sales, because that is the name the Struk printed. (A Pengguna cannot be
// renamed in this app, so there is only ever one such name to answer with.)
type CashierTotal struct {
	CashierID    int64
	CashierName  string
	Total        int64
	Transactions int64
}

// DailyRevenue is the omzet of one store-local day (#9): what came in, how many
// Penjualan made it, and the two breakdowns an Admin reads it by.
//
// Date is a store-local calendar date (`YYYY-MM-DD`), because `created_at` is
// store-local time and a UTC instant would read as yesterday for a sale rung up
// early in the morning (ADR-0015).
type DailyRevenue struct {
	Date         string
	Total        int64
	Transactions int64
	// ByMethod is the day's split by Pembayaran method. The repository answers only
	// the methods that were used; the use case fills in a zero row for the rest, so
	// a caller of the use case always sees every method in the till's order.
	ByMethod []MethodTotal
	// ByCashier has one entry per Kasir who rang something up that day, biggest
	// first. A day with no sales has none.
	ByCashier []CashierTotal
}

// ReportFilter narrows the sales list and the daily report to one store-local
// day (`YYYY-MM-DD`). An empty Date means the store's today, which the use case
// resolves: the domain says what a report is over, not what day it is.
type ReportFilter struct {
	Date string
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

	// ListSales answers the Penjualan of one store-local day, newest Nomor Struk
	// first: the sales list of #9. Each row is a SaleSummary rather than a full
	// Sale, so the list costs one query instead of one per sale.
	ListSales(ctx context.Context, filter ReportFilter) ([]SaleSummary, error)

	// DailyRevenue answers the omzet of one store-local day: the total, the number
	// of Penjualan, and the breakdown by Pembayaran method and by Kasir (#9).
	DailyRevenue(ctx context.Context, filter ReportFilter) (DailyRevenue, error)
}
