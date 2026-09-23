// Package penjualan contains the Penjualan use cases: the Kasir's checkout. It
// composes two domain ports — the catalogue a cart is priced from and the sale
// store it is written to — so the checkout runs without SQLite or HTTP
// (ADR-0004, ADR-0007).
package penjualan

import (
	"context"

	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// ItemInput is one line of a cart: which Produk, and how many units.
type ItemInput struct {
	ProductID int64
	Quantity  int64
}

// PaymentInput is the Pembayaran the Kasir recorded. Amount is what the buyer
// handed over for Tunai; for a non-tunai method it is the nominal being recorded
// and no gateway is involved (CONTEXT.md, Pembayaran).
type PaymentInput struct {
	Method string
	Amount int64
}

// CheckoutInput is the cart a Kasir is checking out.
type CheckoutInput struct {
	Items   []ItemInput
	Payment PaymentInput
}

// InputError is a validation failure that carries a message fit for the API
// response. It unwraps to domainpenjualan.ErrInvalidInput, so the HTTP adapter
// answers 400 while keeping the message the till shows.
type InputError struct {
	Message string
}

func (e InputError) Error() string { return e.Message }

// InputMessage is what the HTTP adapter reads. Each usecase package declares its
// own InputError; this method is the shape they share.
func (e InputError) InputMessage() string { return e.Message }

// Unwrap makes errors.Is(err, domainpenjualan.ErrInvalidInput) true.
func (e InputError) Unwrap() error { return domainpenjualan.ErrInvalidInput }

// StockError is a checkout refused because a Produk's Stok cannot cover the
// cart. It unwraps to domainpenjualan.ErrInsufficientStock, so the adapter
// answers 409, and still carries a message naming the Produk that is short —
// which is more use to the Kasir than a status code alone.
type StockError struct {
	Message string
}

func (e StockError) Error() string { return e.Message }

// InputMessage is what the HTTP adapter reads.
func (e StockError) InputMessage() string { return e.Message }

// Unwrap makes errors.Is(err, domainpenjualan.ErrInsufficientStock) true.
func (e StockError) Unwrap() error { return domainpenjualan.ErrInsufficientStock }

// Service holds the ports the Penjualan use cases need.
type Service struct {
	// products is the catalogue the cart is priced from. The checkout needs only
	// to read a Produk by id, but it takes the whole Produk port rather than a
	// one-method port of its own: it is the same SQLite repository, and a second
	// interface over the same table would be a second thing to keep in step.
	products domainproduk.ProductRepository
	sales    domainpenjualan.SaleRepository
}

// NewService wires the Penjualan use cases to their ports.
func NewService(products domainproduk.ProductRepository, sales domainpenjualan.SaleRepository) *Service {
	return &Service{products: products, sales: sales}
}

// FindByReceiptNumber answers one stored Penjualan by its Nomor Struk. It is the
// read half of the slice: what the till shows after a checkout, and what a
// reprint (#8) and the sales list (#9) build on.
func (s *Service) FindByReceiptNumber(ctx context.Context, receiptNumber int64) (domainpenjualan.Sale, error) {
	return s.sales.FindByReceiptNumber(ctx, receiptNumber)
}
