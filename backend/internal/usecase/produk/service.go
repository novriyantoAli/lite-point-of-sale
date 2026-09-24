// Package produk contains the Produk use cases: an Admin managing the store
// catalogue. Everything it needs arrives as a port, so these cases run without
// SQLite or HTTP (ADR-0004, ADR-0007).
package produk

import (
	"context"

	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// CreateInput is what an Admin fills in to add a Produk.
//
// Stock is the Stok awal the Produk starts life with (CONTEXT.md, Stok): it is
// set here, once. From then on only AddStock raises it and only a Penjualan
// lowers it, which is why UpdateInput below has no such field.
//
// Active is a pointer because only Create reads it. A new Produk is Aktif unless
// the Admin says otherwise, and Update keeps the Produk's existing status —
// deactivating is its own action, not an edit-form decision.
type CreateInput struct {
	Name     string
	Code     string
	Price    int64
	Category string
	Stock    int64
	Active   *bool
}

// UpdateInput is what an Admin fills in to change a Produk: the editable record,
// which is nama, Kode, harga and Kategori.
//
// There is no Stock here, and that is the point (ADR-0014). An edit form holds a
// Stok it read when it opened, and sending it back would overwrite a delivery
// that arrived in between. Stok moves through exactly three paths — the Stok awal
// of a create, AddStock, and a Penjualan — and an edit is not one of them.
type UpdateInput struct {
	Name     string
	Code     string
	Price    int64
	Category string
}

// InputError is a validation failure that carries a message fit for the API
// response. It unwraps to domainproduk.ErrInvalidInput, so the HTTP adapter
// only needs errors.Is to pick the status code and errors.As to read the
// message.
type InputError struct {
	Message string
}

func (e InputError) Error() string { return e.Message }

// InputMessage is what the HTTP adapter reads. Each usecase package declares
// its own InputError; this method is the shape they share.
func (e InputError) InputMessage() string { return e.Message }

// Unwrap makes errors.Is(err, domainproduk.ErrInvalidInput) true.
func (e InputError) Unwrap() error { return domainproduk.ErrInvalidInput }

// LowStockReport is what the restock list answers: the Produk to restock, and
// the ambang that selected them. The threshold travels with the list because it
// is the rule behind it — the UI shows the Admin "Stok di bawah N", and a
// threshold it held separately would be a second copy of that rule.
type LowStockReport struct {
	Threshold int64
	Products  []domainproduk.Product
}

// LowStockSettings is the one setting the restock list needs: the ambang below
// which a Produk counts as menipis. It is a port of this package rather than an
// import of the pengaturan domain, so the dependency still points inward
// (ADR-0004, ADR-0017).
type LowStockSettings interface {
	LowStockThreshold(ctx context.Context) (int64, error)
}

// Service holds the ports the Produk use cases need.
type Service struct {
	products domainproduk.ProductRepository
	// settings answers the ambang Stok menipis, which moved out of a domain
	// constant and into the stored Pengaturan (ADR-0017).
	settings LowStockSettings
}

// NewService wires the Produk use cases to their ports.
func NewService(products domainproduk.ProductRepository, settings LowStockSettings) *Service {
	return &Service{products: products, settings: settings}
}
