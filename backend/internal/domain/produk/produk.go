// Package produk holds the Produk domain: one catalogue entry (nama, Kode,
// harga, Kategori, Stok) that is either active or Nonaktif. It declares the
// port the use cases need without depending on SQLite or HTTP (ADR-0004).
//
// The Go identifiers are English (Product, ProductRepository) while the domain
// term, the route and the frontend types stay Indonesian (Produk, /api/produk),
// the same boundary ADR-0011 draws for Pengguna and ADR-0012 makes the rule.
package produk

import (
	"context"
	"errors"
)

// Product is one catalogue entry.
//
// Code and Category are pointers because "no Kode" and "no Kategori" are
// different from an empty string: an absent Kode is NULL in the database and
// `null` in JSON, and a Product with no Kode is found by name instead.
type Product struct {
	ID       int64
	Name     string
	Code     *string
	Price    int64
	Category *string
	Stock    int64
	Active   bool
	// Sold reports that this Product was part of a finished Penjualan. A Sold
	// Product can be deactivated but never deleted (CONTEXT.md, Nonaktif).
	Sold bool
}

// LowStockThreshold is the Stok below which a Produk counts as menipis: the
// Admin is told to restock it (CONTEXT.md, Stok). "Below", not "at or below" —
// the ambang is the first Stok that is still enough. It is a constant rather
// than a stored setting because there is nowhere to store one yet — the
// Pengaturan screen arrives with the Struk template in #8, and this moves there
// when it does.
const LowStockThreshold int64 = 5

// Filter narrows a catalogue listing. Every field is optional: an empty string
// (or a nil Active) means "do not filter on this".
//
// Active is a pointer for the same reason Code is: an Admin managing the
// catalogue sees Nonaktif Produk too, while the kasir lookup asks for
// Active=true only — and "no filter" is neither of those.
type Filter struct {
	Name     string
	Code     string
	Category string
	Active   *bool
}

// Sentinel errors the use cases return and the HTTP adapter maps to status
// codes. Keep the set small and meaningful.
var (
	ErrProductNotFound = errors.New("product not found")
	ErrCodeTaken       = errors.New("product code taken")
	ErrProductHasSales = errors.New("product has sales")
	ErrInvalidInput    = errors.New("invalid input")
)

// ProductRepository is the outbound port for Produk persistence. Implemented in
// the SQLite adapter; the use cases never see SQL (ADR-0004).
type ProductRepository interface {
	Create(ctx context.Context, product Product) (Product, error)
	// Update replaces the mutable fields of a Produk (nama, Kode, harga,
	// Kategori, Stok). Active and Sold are changed by their own methods,
	// because neither is something an edit form may rewrite.
	Update(ctx context.Context, product Product) (Product, error)
	FindByID(ctx context.Context, id int64) (Product, error)
	FindByCode(ctx context.Context, code string) (Product, error)
	List(ctx context.Context, filter Filter) ([]Product, error)
	// AddStock adds quantity units to a Produk's Stok and answers the Produk as
	// it now stands. It adds rather than sets, so two restocks arriving at once
	// cannot lose one of the two — the read-modify-write that would is exactly
	// what this method exists to avoid.
	AddStock(ctx context.Context, id int64, quantity int64) (Product, error)
	// ListLowStock answers the Active Produk whose Stok is below threshold,
	// thinnest first, so the Admin reads what to restock before what can wait.
	// A Nonaktif Produk is left out: it is not for sale, so its Stok cannot run
	// out in a way that matters (CONTEXT.md, Nonaktif).
	ListLowStock(ctx context.Context, threshold int64) ([]Product, error)
	// Categories returns every distinct Kategori in use, sorted, so the UI can
	// offer the ones that exist instead of a hand-maintained list.
	Categories(ctx context.Context) ([]string, error)
	SetActive(ctx context.Context, id int64, active bool) error
	Delete(ctx context.Context, id int64) error
}
