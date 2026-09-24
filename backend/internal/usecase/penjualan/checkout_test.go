package penjualan

import (
	"context"
	"errors"
	"strings"
	"testing"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// kasir is the Pengguna ringing the test sales up.
var kasir = domainauth.PublicUser{ID: 7, Username: "kasir1", Role: domainauth.RoleKasir}

// kopi and teh are the two Produk most tests sell.
func kopi() domainproduk.Product {
	return domainproduk.Product{ID: 1, Name: "Kopi Susu", Price: 18000, Stock: 10, Active: true}
}

func teh() domainproduk.Product {
	return domainproduk.Product{ID: 2, Name: "Teh Botol", Price: 6000, Stock: 4, Active: true}
}

// tunai builds a Tunai checkout of one line.
func tunai(productID, quantity, amount int64) CheckoutInput {
	return CheckoutInput{
		Items:   []ItemInput{{ProductID: productID, Quantity: quantity}},
		Payment: PaymentInput{Method: "cash", Amount: amount},
	}
}

// nonTunai builds a checkout of one line paid by a recorded method — QRIS, Debit
// or Transfer. Nothing is processed through a gateway: the method and the
// nominal are all that is written (CONTEXT.md, Pembayaran).
func nonTunai(method string, productID, quantity, amount int64) CheckoutInput {
	return CheckoutInput{
		Items:   []ItemInput{{ProductID: productID, Quantity: quantity}},
		Payment: PaymentInput{Method: method, Amount: amount},
	}
}

// saleOf runs a checkout and answers the Penjualan it stored, failing the test
// when the checkout was refused. What the checkout printed is the subject of
// cetak_test.go.
func saleOf(t *testing.T, service *Service, input CheckoutInput) domainpenjualan.Sale {
	t.Helper()

	result, err := service.Checkout(context.Background(), kasir, input)
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}

	return result.Sale
}

func TestCheckoutPricesTheCartFromTheCatalogue(t *testing.T) {
	products := newFakeProducts(kopi(), teh())
	sales := newFakeSales()

	sale := saleOf(t, newTestService(products, sales), CheckoutInput{
		Items: []ItemInput{
			{ProductID: 1, Quantity: 2},
			{ProductID: 2, Quantity: 3},
		},
		Payment: PaymentInput{Method: "cash", Amount: 60000},
	})

	if sale.Total != 2*18000+3*6000 {
		t.Errorf("total: got %d, want %d", sale.Total, 2*18000+3*6000)
	}
	if sale.Payment.Change != 60000-(2*18000+3*6000) {
		t.Errorf("Kembalian: got %d, want %d", sale.Payment.Change, 60000-(2*18000+3*6000))
	}
	if sale.Payment.Method != domainpenjualan.PaymentCash {
		t.Errorf("method: got %q, want %q", sale.Payment.Method, domainpenjualan.PaymentCash)
	}

	// The Item snapshots are the name and the price the catalogue held, not an
	// id the till would have to look up again later.
	if len(sale.Items) != 2 {
		t.Fatalf("items: got %d, want 2", len(sale.Items))
	}
	if sale.Items[0].Name != "Kopi Susu" || sale.Items[0].Price != 18000 || sale.Items[0].Quantity != 2 {
		t.Errorf("first Item: got %+v, want a Kopi Susu snapshot of 2 at 18000", sale.Items[0])
	}
	if sale.Items[0].Subtotal() != 36000 {
		t.Errorf("first Item subtotal: got %d, want 36000", sale.Items[0].Subtotal())
	}
}

func TestCheckoutCopiesTheCashierAndAnswersTheStoredSale(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()

	sale := saleOf(t, newTestService(products, sales), tunai(1, 1, 18000))

	if sale.CashierID != kasir.ID || sale.CashierName != kasir.Username {
		t.Errorf("cashier: got %d/%q, want %d/%q", sale.CashierID, sale.CashierName, kasir.ID, kasir.Username)
	}

	// What the use case answers is what the repository stored — including the
	// Nomor Struk and the timestamp it assigned, which the use case never invents.
	if sale.ReceiptNumber != 1 || sale.CreatedAt == "" {
		t.Errorf("stored sale: got receipt %d and timestamp %q, want the repository's own",
			sale.ReceiptNumber, sale.CreatedAt)
	}
	if len(sales.recorded) != 1 {
		t.Fatalf("recorded: got %d Penjualan, want 1", len(sales.recorded))
	}
}

func TestCheckoutBlocksACartTheStokCannotCover(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()

	_, err := newTestService(products, sales).Checkout(context.Background(), kasir, tunai(1, 11, 200000))
	if !errors.Is(err, domainpenjualan.ErrInsufficientStock) {
		t.Fatalf("checkout of 11 against a Stok of 10: got %v, want ErrInsufficientStock", err)
	}

	// The message names the Produk and what is left, which is what the Kasir needs
	// to fix the cart.
	var stockErr StockError
	if !errors.As(err, &stockErr) {
		t.Fatalf("error: got %T, want a StockError the adapter can read", err)
	}
	if !strings.Contains(stockErr.Message, "Kopi Susu") {
		t.Errorf("message: got %q, want it to name the Produk", stockErr.Message)
	}

	// Refused means nothing was stored.
	if len(sales.recorded) != 0 {
		t.Errorf("recorded: got %d Penjualan, want none", len(sales.recorded))
	}
}

func TestCheckoutAcceptsACartThatExactlyFitsTheStok(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()

	// The Stok is the number that is still enough: 10 of 10 is a sale, not a
	// refusal.
	if _, err := newTestService(products, sales).Checkout(context.Background(), kasir, tunai(1, 10, 180000)); err != nil {
		t.Fatalf("checkout of the whole Stok: %v", err)
	}
}

func TestCheckoutBlocksAPaymentBelowTheTotal(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()

	_, err := newTestService(products, sales).Checkout(context.Background(), kasir, tunai(1, 1, 17999))
	if !errors.Is(err, domainpenjualan.ErrInvalidInput) {
		t.Fatalf("underpaid checkout: got %v, want ErrInvalidInput", err)
	}
	if len(sales.recorded) != 0 {
		t.Errorf("recorded: got %d Penjualan, want none", len(sales.recorded))
	}
}

func TestCheckoutAcceptsAPaymentOfExactlyTheTotal(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()

	sale := saleOf(t, newTestService(products, sales), tunai(1, 1, 18000))

	if sale.Payment.Change != 0 {
		t.Errorf("Kembalian: got %d, want 0", sale.Payment.Change)
	}
}

func TestCheckoutRefusesAnEmptyCart(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()

	tests := []struct {
		name string
		body CheckoutInput
	}{
		{
			name: "no Items at all",
			body: CheckoutInput{Payment: PaymentInput{Method: "cash", Amount: 10000}},
		},
		{
			name: "a quantity of zero",
			body: tunai(1, 0, 10000),
		},
		{
			name: "a negative quantity",
			body: tunai(1, -3, 10000),
		},
		{
			name: "a Produk id that is not one",
			body: tunai(0, 1, 10000),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := newTestService(products, sales).Checkout(context.Background(), kasir, test.body)
			if !errors.Is(err, domainpenjualan.ErrInvalidInput) {
				t.Fatalf("got %v, want ErrInvalidInput", err)
			}
		})
	}

	if len(sales.recorded) != 0 {
		t.Errorf("recorded: got %d Penjualan, want none", len(sales.recorded))
	}
}

func TestCheckoutMergesRepeatedProdukIntoOneItem(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()

	sale := saleOf(t, newTestService(products, sales), CheckoutInput{
		Items: []ItemInput{
			{ProductID: 1, Quantity: 2},
			{ProductID: 1, Quantity: 3},
		},
		Payment: PaymentInput{Method: "cash", Amount: 90000},
	})

	// One Item of the summed quantity: an Item is one Produk (CONTEXT.md, Item).
	if len(sale.Items) != 1 || sale.Items[0].Quantity != 5 {
		t.Fatalf("items: got %+v, want one Item of 5", sale.Items)
	}
	if sale.Total != 5*18000 {
		t.Errorf("total: got %d, want %d", sale.Total, 5*18000)
	}
}

func TestCheckoutBlocksARepeatedProdukThatOnlyFitsWhenAddedUp(t *testing.T) {
	products := newFakeProducts(domainproduk.Product{ID: 1, Name: "Kopi", Price: 18000, Stock: 3, Active: true})
	sales := newFakeSales()

	// 2 + 2 each fit a Stok of 3 on their own; together they do not, and the cart
	// is refused rather than half sold.
	_, err := newTestService(products, sales).Checkout(context.Background(), kasir, CheckoutInput{
		Items: []ItemInput{
			{ProductID: 1, Quantity: 2},
			{ProductID: 1, Quantity: 2},
		},
		Payment: PaymentInput{Method: "cash", Amount: 72000},
	})
	if !errors.Is(err, domainpenjualan.ErrInsufficientStock) {
		t.Fatalf("got %v, want ErrInsufficientStock", err)
	}
}

func TestCheckoutRefusesAProdukThatIsNotSellable(t *testing.T) {
	tests := []struct {
		name     string
		products *fakeProducts
		id       int64
	}{
		{
			name:     "a Produk that is gone",
			products: newFakeProducts(kopi()),
			id:       404,
		},
		{
			name: "a Produk that is Nonaktif",
			products: newFakeProducts(domainproduk.Product{
				ID: 1, Name: "Teh", Price: 6000, Stock: 5, Active: false,
			}),
			id: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sales := newFakeSales()

			_, err := newTestService(test.products, sales).Checkout(context.Background(), kasir, tunai(test.id, 1, 10000))
			if !errors.Is(err, domainpenjualan.ErrInvalidInput) {
				t.Fatalf("got %v, want ErrInvalidInput", err)
			}
			if len(sales.recorded) != 0 {
				t.Errorf("recorded: got %d Penjualan, want none", len(sales.recorded))
			}
		})
	}
}

func TestCheckoutRecordsANonTunaiPembayaran(t *testing.T) {
	tests := []struct {
		name   string
		method domainpenjualan.PaymentMethod
	}{
		{name: "QRIS", method: domainpenjualan.PaymentQRIS},
		{name: "Debit", method: domainpenjualan.PaymentDebit},
		{name: "Transfer", method: domainpenjualan.PaymentTransfer},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			products := newFakeProducts(kopi())
			sales := newFakeSales()

			// The nominal of a non-tunai Pembayaran is the total of the Penjualan it
			// pays for: one sale, one Pembayaran, no split (CONTEXT.md, Pembayaran).
			sale := saleOf(t, newTestService(products, sales), nonTunai(string(test.method), 1, 2, 36000))

			if sale.Payment.Method != test.method {
				t.Errorf("method: got %q, want %q", sale.Payment.Method, test.method)
			}
			if sale.Payment.Amount != 36000 {
				t.Errorf("nominal: got %d, want 36000", sale.Payment.Amount)
			}
			// No Kembalian: only Tunai can be handed back (CONTEXT.md, Kembalian).
			if sale.Payment.Change != 0 {
				t.Errorf("Kembalian: got %d, want 0", sale.Payment.Change)
			}

			// What is recorded is what the repository stored, method included.
			if len(sales.recorded) != 1 || sales.recorded[0].Payment.Method != test.method {
				t.Errorf("recorded: got %+v, want one Penjualan paid by %q", sales.recorded, test.method)
			}
		})
	}
}

func TestCheckoutRefusesANonTunaiPembayaranThatIsNotTheTotal(t *testing.T) {
	tests := []struct {
		name   string
		amount int64
	}{
		{name: "more than the total", amount: 40000},
		{name: "less than the total", amount: 35000},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			products := newFakeProducts(kopi())
			sales := newFakeSales()

			// There is no Kembalian to absorb an overpayment and no split payment to
			// absorb an underpayment, so a nominal that is not the total is refused
			// rather than recorded (CONTEXT.md, Pembayaran, Kembalian).
			_, err := newTestService(products, sales).Checkout(context.Background(), kasir,
				nonTunai("qris", 1, 2, test.amount))
			if !errors.Is(err, domainpenjualan.ErrInvalidInput) {
				t.Fatalf("got %v, want ErrInvalidInput", err)
			}
			if len(sales.recorded) != 0 {
				t.Errorf("recorded: got %d Penjualan, want none", len(sales.recorded))
			}
		})
	}
}

func TestCheckoutRefusesAMethodItDoesNotKnow(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{name: "not a method at all", method: "bitcoin"},
		{name: "blank", method: "  "},
		{name: "an upper-cased method", method: "QRIS"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			products := newFakeProducts(kopi())
			sales := newFakeSales()

			_, err := newTestService(products, sales).Checkout(context.Background(), kasir,
				nonTunai(test.method, 1, 1, 18000))
			if !errors.Is(err, domainpenjualan.ErrInvalidInput) {
				t.Fatalf("got %v, want ErrInvalidInput", err)
			}
			if len(sales.recorded) != 0 {
				t.Errorf("recorded: got %d Penjualan, want none", len(sales.recorded))
			}
		})
	}
}

func TestCheckoutPropagatesACatalogueFailure(t *testing.T) {
	products := newFakeProducts(kopi())
	products.err = errors.New("catalogue is down")
	sales := newFakeSales()

	_, err := newTestService(products, sales).Checkout(context.Background(), kasir, tunai(1, 1, 18000))
	if !errors.Is(err, products.err) {
		t.Fatalf("got %v, want the catalogue failure", err)
	}
	if len(sales.recorded) != 0 {
		t.Errorf("recorded: got %d Penjualan, want none", len(sales.recorded))
	}
}

func TestCheckoutPropagatesASaleRepositoryFailure(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()
	sales.err = errors.New("disk is full")

	_, err := newTestService(products, sales).Checkout(context.Background(), kasir, tunai(1, 1, 18000))
	if !errors.Is(err, sales.err) {
		t.Fatalf("got %v, want the repository failure", err)
	}
}

func TestFindByReceiptNumberAnswersAStoredSale(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()
	service := newTestService(products, sales)

	created := saleOf(t, service, tunai(1, 1, 18000))

	stored, err := service.FindByReceiptNumber(context.Background(), created.ReceiptNumber)
	if err != nil {
		t.Fatalf("find by Nomor Struk: %v", err)
	}
	if stored.Total != created.Total {
		t.Errorf("total: got %d, want %d", stored.Total, created.Total)
	}
}

func TestFindByReceiptNumberReportsAMissingSale(t *testing.T) {
	service := newTestService(newFakeProducts(), newFakeSales())

	_, err := service.FindByReceiptNumber(context.Background(), 404)
	if !errors.Is(err, domainpenjualan.ErrSaleNotFound) {
		t.Fatalf("got %v, want ErrSaleNotFound", err)
	}
}
