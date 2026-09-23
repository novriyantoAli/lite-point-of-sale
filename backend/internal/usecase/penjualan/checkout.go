package penjualan

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// Checkout turns a cart into a finished Penjualan: it prices every Item from the
// catalogue, refuses a cart the Stok cannot cover, works out the Kembalian, and
// hands the sale to the repository, which stores the Penjualan, the Item
// snapshots and the Stok decrement in one transaction (CONTEXT.md, Penjualan).
//
// cashier is the Pengguna ringing it up. Their username is copied onto the sale,
// because that is the name the Struk prints — a later rename must not rewrite a
// Struk that was already handed over.
func (s *Service) Checkout(ctx context.Context, cashier domainauth.PublicUser, input CheckoutInput) (domainpenjualan.Sale, error) {
	lines, err := mergeLines(input.Items)
	if err != nil {
		return domainpenjualan.Sale{}, err
	}

	method, err := paymentMethod(input.Payment)
	if err != nil {
		return domainpenjualan.Sale{}, err
	}

	items, total, err := s.price(ctx, lines)
	if err != nil {
		return domainpenjualan.Sale{}, err
	}

	// Tunai below the total is refused rather than stored: Kembalian is
	// `bayar - total` and has to be at least zero (CONTEXT.md, Kembalian). A
	// non-tunai Pembayaran is recorded for its nominal, so the same rule is what
	// stops it recording less than the sale it pays for.
	if input.Payment.Amount < total {
		return domainpenjualan.Sale{}, InputError{Message: "Jumlah bayar kurang dari total Penjualan."}
	}

	sale := domainpenjualan.Sale{
		CashierID:   cashier.ID,
		CashierName: cashier.Username,
		Total:       total,
		Items:       items,
		Payment: domainpenjualan.Payment{
			Method: method,
			Amount: input.Payment.Amount,
			Change: input.Payment.Amount - total,
		},
	}

	return s.sales.Create(ctx, sale)
}

// mergeLines validates the cart and folds repeated Produk into one line.
//
// A Penjualan's Item is one Produk with an integer quantity (CONTEXT.md, Item),
// so a cart that names the same Produk twice is one line of the summed quantity.
// Two lines would also let a cart of 2 + 2 pass a Stok check of 3 one line at a
// time and only fail at the till. The first occurrence keeps its place, so the
// order the Kasir rang things up survives.
func mergeLines(input []ItemInput) ([]ItemInput, error) {
	if len(input) == 0 {
		return nil, InputError{Message: "Keranjang masih kosong."}
	}

	lines := make([]ItemInput, 0, len(input))
	at := make(map[int64]int, len(input))

	for _, item := range input {
		if item.ProductID <= 0 {
			return nil, InputError{Message: "Ada Produk yang tidak dikenal di keranjang."}
		}
		if item.Quantity <= 0 {
			return nil, InputError{Message: "Jumlah Item harus lebih dari nol."}
		}

		if index, seen := at[item.ProductID]; seen {
			lines[index].Quantity += item.Quantity
			continue
		}

		at[item.ProductID] = len(lines)
		lines = append(lines, item)
	}

	return lines, nil
}

// paymentMethod reads the Pembayaran method of this checkout.
//
// Only Tunai is accepted here. The domain already names QRIS, Debit and Transfer
// — a stored Penjualan has to stay readable when #7 lands them — but accepting
// one today would record a Pembayaran this slice cannot answer for.
func paymentMethod(input PaymentInput) (domainpenjualan.PaymentMethod, error) {
	method := domainpenjualan.PaymentMethod(strings.TrimSpace(input.Method))

	if method == domainpenjualan.PaymentCash {
		return method, nil
	}
	if !method.Valid() {
		return "", InputError{Message: "Metode pembayaran tidak dikenal."}
	}

	return "", InputError{Message: "Metode pembayaran ini belum tersedia."}
}

// price reads every Produk in the cart and turns the lines into Item snapshots,
// summing the total.
//
// It is also where a checkout is blocked before anything is written: a Produk
// that is gone, is Nonaktif, or does not have the Stok for the quantity asked
// for is refused here, with a message naming the Produk and what is left
// (CONTEXT.md, Stok, Nonaktif). The Stok check is repeated inside the
// repository's transaction, because the value read here can be stale by the time
// the sale is written.
func (s *Service) price(ctx context.Context, lines []ItemInput) ([]domainpenjualan.Item, int64, error) {
	items := make([]domainpenjualan.Item, 0, len(lines))
	var total int64

	for _, line := range lines {
		product, err := s.products.FindByID(ctx, line.ProductID)
		if err != nil {
			if errors.Is(err, domainproduk.ErrProductNotFound) {
				return nil, 0, InputError{Message: "Ada Produk di keranjang yang sudah tidak ada."}
			}
			return nil, 0, err
		}

		if !product.Active {
			return nil, 0, InputError{
				Message: fmt.Sprintf("Produk %s sudah Nonaktif dan tidak bisa dijual.", product.Name),
			}
		}

		if product.Stock < line.Quantity {
			return nil, 0, StockError{
				Message: fmt.Sprintf("Stok %s tidak cukup: tersisa %d, diminta %d.", product.Name, product.Stock, line.Quantity),
			}
		}

		item := domainpenjualan.Item{
			ProductID: product.ID,
			Name:      product.Name,
			Price:     product.Price,
			Quantity:  line.Quantity,
		}
		items = append(items, item)
		total += item.Subtotal()
	}

	return items, total, nil
}
