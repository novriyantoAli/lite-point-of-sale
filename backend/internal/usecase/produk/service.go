// Package produk contains the Produk use cases: an Admin managing the store
// catalogue. Everything it needs arrives as a port, so these cases run without
// SQLite or HTTP (ADR-0004, ADR-0007).
package produk

import (
	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// ProductInput is what an Admin fills in to add or change a Produk. The same
// fields serve both: an edit replaces the whole editable record, and Stok is
// part of it — the dedicated "tambah Stok" flow of #5 arrives on top of this,
// not instead of it.
type ProductInput struct {
	Name     string
	Code     string
	Price    int64
	Category string
	Stock    int64
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
