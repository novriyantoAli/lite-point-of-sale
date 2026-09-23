// Package produk contains the Produk use cases: an Admin managing the store
// catalogue. Everything it needs arrives as a port, so these cases run without
// SQLite or HTTP (ADR-0004, ADR-0007).
package produk

import (
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

// Service holds the ports the Produk use cases need.
type Service struct {
	products domainproduk.ProductRepository
}

// NewService wires the Produk use cases to their port.
func NewService(products domainproduk.ProductRepository) *Service {
	return &Service{products: products}
}
