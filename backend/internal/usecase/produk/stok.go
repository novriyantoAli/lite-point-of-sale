package produk

import (
	"context"

	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// AddStock adds quantity units to a Produk's Stok: the restock an Admin records
// when goods arrive (CONTEXT.md, Stok).
//
// It adds rather than replaces, and the addition happens in the repository's own
// statement — so a Stok is never read, changed in Go and written back, which is
// the shape that loses a restock when two arrive at once.
//
// Only a positive quantity is accepted. A "restock" of zero says nothing, and a
// negative one is not a restock at all: Stok leaves the catalogue through a
// Penjualan, which is the transaction that has to keep Stok non-negative (#6),
// not through an Admin's form.
func (s *Service) AddStock(ctx context.Context, id int64, quantity int64) (domainproduk.Product, error) {
	if quantity <= 0 {
		return domainproduk.Product{}, InputError{Message: "Jumlah Stok harus lebih dari nol."}
	}

	return s.products.AddStock(ctx, id, quantity)
}

// LowStock answers the Produk an Admin has to restock: the Active ones whose
// Stok is below the ambang Stok menipis, thinnest first — together with the
// ambang itself, because it is the rule that decided the list. Whether a Produk
// counts as menipis is a domain rule, and the threshold is now a stored setting
// (ADR-0017), so it is read here from the LowStockSettings port rather than a
// constant or a value the caller passes in.
//
// "Below" and not "at or below": the ambang is the first Stok that is still
// enough, so a Produk sitting exactly on it has not run low yet.
func (s *Service) LowStock(ctx context.Context) (LowStockReport, error) {
	threshold, err := s.settings.LowStockThreshold(ctx)
	if err != nil {
		return LowStockReport{}, err
	}

	products, err := s.products.ListLowStock(ctx, threshold)
	if err != nil {
		return LowStockReport{}, err
	}

	return LowStockReport{Threshold: threshold, Products: products}, nil
}
