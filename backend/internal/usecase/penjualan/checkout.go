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
// Once the Penjualan is stored it prints the Struk, and answers the sale together
// with the outcome of that print. Printing runs *after* the write and never
// inside it: if the printer is missing or fails, the sale stays recorded and the
// result says so — the money has already moved, and a Struk is a document that can
// be reissued (ADR-0017, keputusan 1).
//
// cashier is the Pengguna ringing it up. Their username is copied onto the sale,
// because that is the name the Struk prints — a later rename must not rewrite a
// Struk that was already handed over.
func (s *Service) Checkout(ctx context.Context, cashier domainauth.PublicUser, input CheckoutInput) (CheckoutResult, error) {
	lines, err := mergeLines(input.Items)
	if err != nil {
		return CheckoutResult{}, err
	}

	method, err := paymentMethod(input.Payment)
	if err != nil {
		return CheckoutResult{}, err
	}

	items, total, err := s.price(ctx, lines)
	if err != nil {
		return CheckoutResult{}, err
	}

	// The Pembayaran is recorded as it was taken, and what each method accepts
	// differs (CONTEXT.md, Pembayaran, Kembalian):
	//
	//   - Tunai is what the buyer handed over, and it has to cover the total:
	//     Kembalian is `bayar - total` and must be at least zero.
	//   - QRIS, Debit and Transfer are recorded for the total of the sale. There is
	//     no Kembalian to absorb a difference, and one Penjualan has one Pembayaran,
	//     so there is no split to absorb an underpayment either.
	//
	// Change stays zero for a recorded method: there is nothing to hand back.
	var change int64
	if method == domainpenjualan.PaymentCash {
		if input.Payment.Amount < total {
			return CheckoutResult{}, InputError{Message: "Jumlah bayar kurang dari total Penjualan."}
		}
		change = input.Payment.Amount - total
	} else if input.Payment.Amount != total {
		return CheckoutResult{}, InputError{
			Message: fmt.Sprintf("Pembayaran non-tunai harus sebesar total Penjualan: %d.", total),
		}
	}

	sale := domainpenjualan.Sale{
		CashierID:   cashier.ID,
		CashierName: cashier.Username,
		Total:       total,
		Items:       items,
		Payment: domainpenjualan.Payment{
			Method: method,
			Amount: input.Payment.Amount,
			Change: change,
		},
	}

	stored, err := s.sales.Create(ctx, sale)
	if err != nil {
		return CheckoutResult{}, err
	}

	printed, err := s.printStruk(ctx, stored)
	if err != nil {
		// The template could not be read, so there was nothing to print. The
		// Penjualan is stored either way: a print that could not even be composed is
		// a failed print, never a failed sale (ADR-0017, keputusan 1).
		printed = PrintResult{Printed: false, Message: composeFailureMessage}
	}

	return CheckoutResult{Sale: stored, Print: printed}, nil
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
// All four methods of CONTEXT.md are accepted: Tunai takes money, and QRIS, Debit
// and Transfer are recorded — no gateway is involved (CONTEXT.md, Pembayaran). An
// unknown method is refused rather than stored, because the method column names
// exactly these four and a Pembayaran the app cannot name is not one it can show
// or report on.
func paymentMethod(input PaymentInput) (domainpenjualan.PaymentMethod, error) {
	method := domainpenjualan.PaymentMethod(strings.TrimSpace(input.Method))

	if !method.Valid() {
		return "", InputError{Message: "Metode pembayaran tidak dikenal."}
	}

	return method, nil
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
