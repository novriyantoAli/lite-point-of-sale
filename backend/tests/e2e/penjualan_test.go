package e2e

import (
	"net/http"
	"testing"
)

// checkoutItemPayload is one line of the cart a Kasir posts: which Produk, and
// how many units. It never carries the name or the price — those are the
// catalogue's to know, and a client that sent them could write its own history.
type checkoutItemPayload struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

// checkoutPaymentPayload is the Pembayaran being recorded. The Kembalian is the
// API's to work out from the total, never the client's to declare.
type checkoutPaymentPayload struct {
	Method string `json:"method"`
	Amount int64  `json:"amount"`
}

type checkoutPayload struct {
	Items   []checkoutItemPayload  `json:"items"`
	Payment checkoutPaymentPayload `json:"payment"`
}

// saleItemPayload is one Item as the API answers it: the Produk it came from,
// plus the name and price copied at checkout.
type saleItemPayload struct {
	ProductID int64  `json:"product_id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
	Quantity  int64  `json:"quantity"`
	Subtotal  int64  `json:"subtotal"`
}

type paymentPayload struct {
	Method string `json:"method"`
	Amount int64  `json:"amount"`
	Change int64  `json:"change"`
}

type salePayload struct {
	ReceiptNumber int64             `json:"receipt_number"`
	CreatedAt     string            `json:"created_at"`
	CashierID     int64             `json:"cashier_id"`
	CashierName   string            `json:"cashier_name"`
	Total         int64             `json:"total"`
	Items         []saleItemPayload `json:"items"`
	Payment       paymentPayload    `json:"payment"`
}

type saleEnvelope struct {
	Sale salePayload `json:"sale"`
}

// tunai builds a Tunai checkout of one line.
func tunai(productID, quantity, amount int64) checkoutPayload {
	return checkoutPayload{
		Items:   []checkoutItemPayload{{ProductID: productID, Quantity: quantity}},
		Payment: checkoutPaymentPayload{Method: "cash", Amount: amount},
	}
}

// checkout posts a cart and answers the stored Penjualan, failing the test when
// the API refused it.
func checkout(t *testing.T, baseURL, token string, payload checkoutPayload) salePayload {
	t.Helper()

	var created dataEnvelope[saleEnvelope]
	status := apiCall(t, http.MethodPost, baseURL+"/api/penjualan", token, payload, &created)

	if status != http.StatusCreated {
		t.Fatalf("checkout: got status %d, want %d", status, http.StatusCreated)
	}

	return created.Data.Sale
}

// readSale answers one Penjualan by its Nomor Struk, failing the test when the
// API refused it.
func readSale(t *testing.T, baseURL, token string, receiptNumber int64) salePayload {
	t.Helper()

	var sale dataEnvelope[saleEnvelope]
	status := apiCall(t, http.MethodGet, baseURL+"/api/penjualan/"+itoa(receiptNumber), token, nil, &sale)

	if status != http.StatusOK {
		t.Fatalf("read Penjualan %d: got status %d, want %d", receiptNumber, status, http.StatusOK)
	}

	return sale.Data.Sale
}

// stockOf answers one Produk's Stok as the catalogue now reports it.
func stockOf(t *testing.T, baseURL, token string, id int64) int64 {
	t.Helper()

	for _, product := range listProduk(t, baseURL, token, "") {
		if product.ID == id {
			return product.Stock
		}
	}

	t.Fatalf("Produk %d is not in the catalogue", id)

	return 0
}

func TestCheckoutRequiresASession(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})

	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "checkout", method: http.MethodPost, path: "/api/penjualan", body: tunai(created.ID, 1, 18000)},
		{name: "read a Penjualan", method: http.MethodGet, path: "/api/penjualan/1"},
	}

	for _, test := range tests {
		t.Run(test.name+" anonymously", func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, test.method, baseURL+test.path, "", test.body, &failure)

			if status != http.StatusUnauthorized {
				t.Fatalf("got status %d, want %d", status, http.StatusUnauthorized)
			}
			if failure.Error != "invalid_token" {
				t.Errorf("got code %q, want %q", failure.Error, "invalid_token")
			}
		})
	}
}

func TestCheckoutRecordsAPenjualanAndTakesTheStok(t *testing.T) {
	baseURL, token := newAdminToken(t)
	kopi := createProduk(t, baseURL, token, produkPayload{Name: "Kopi Susu", Code: strPtr("KOPI-01"), Price: 18000, Stock: 10})
	teh := createProduk(t, baseURL, token, produkPayload{Name: "Teh Botol", Price: 6000, Stock: 4})

	sale := checkout(t, baseURL, token, checkoutPayload{
		Items: []checkoutItemPayload{
			{ProductID: kopi.ID, Quantity: 2},
			{ProductID: teh.ID, Quantity: 3},
		},
		Payment: checkoutPaymentPayload{Method: "cash", Amount: 60000},
	})

	if sale.ReceiptNumber != 1 {
		t.Errorf("Nomor Struk: got %d, want the first one to be 1", sale.ReceiptNumber)
	}
	if sale.Total != 2*18000+3*6000 {
		t.Errorf("total: got %d, want %d", sale.Total, 2*18000+3*6000)
	}
	if sale.CashierName != seededAdmin {
		t.Errorf("cashier: got %q, want %q", sale.CashierName, seededAdmin)
	}
	if sale.CreatedAt == "" {
		t.Error("created_at: got nothing, want the time the sale was stored")
	}
	if len(sale.Items) != 2 {
		t.Fatalf("items: got %d, want 2", len(sale.Items))
	}

	// The Stok left by exactly what was sold — no more, no less.
	if stock := stockOf(t, baseURL, token, kopi.ID); stock != 8 {
		t.Errorf("Kopi Susu Stok: got %d, want 8", stock)
	}
	if stock := stockOf(t, baseURL, token, teh.ID); stock != 1 {
		t.Errorf("Teh Botol Stok: got %d, want 1", stock)
	}

	// The sale is stored, not just echoed: reading it back by Nomor Struk answers
	// the same thing.
	stored := readSale(t, baseURL, token, sale.ReceiptNumber)
	if stored.Total != sale.Total || len(stored.Items) != 2 {
		t.Errorf("read back: got %+v, want the Penjualan that was just stored", stored)
	}
	if stored.Items[0].Name != "Kopi Susu" || stored.Items[0].Price != 18000 {
		t.Errorf("read back Item: got %+v, want the name and price copied at checkout", stored.Items[0])
	}
	if stored.Items[0].Subtotal != 36000 {
		t.Errorf("read back subtotal: got %d, want 36000", stored.Items[0].Subtotal)
	}
}

func TestCheckoutWorksAsKasir(t *testing.T) {
	baseURL, adminToken := newAdminToken(t)
	created := createProduk(t, baseURL, adminToken, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})
	createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")
	kasirToken := logIn(t, baseURL, "kasir1", kasirPassword)

	sale := checkout(t, baseURL, kasirToken, tunai(created.ID, 1, 20000))

	// Both Peran sell at a one-terminal store, and the Struk names whoever rang
	// it up.
	if sale.CashierName != "kasir1" {
		t.Errorf("cashier: got %q, want %q", sale.CashierName, "kasir1")
	}
}

func TestCheckoutSnapshotsTheNameAndPrice(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi Susu", Price: 18000, Stock: 5})

	sale := checkout(t, baseURL, token, tunai(created.ID, 1, 18000))

	// The catalogue is repriced and renamed afterwards…
	var updated dataEnvelope[productEnvelope]
	status := apiCall(t, http.MethodPut, baseURL+"/api/produk/"+itoa(created.ID), token,
		produkEditPayload{Name: "Kopi Susu Gula Aren", Price: 25000}, &updated)
	if status != http.StatusOK {
		t.Fatalf("reprice: got status %d, want %d", status, http.StatusOK)
	}

	// …and the Penjualan that already happened is untouched.
	stored := readSale(t, baseURL, token, sale.ReceiptNumber)
	if stored.Items[0].Name != "Kopi Susu" || stored.Items[0].Price != 18000 {
		t.Errorf("Item after reprice: got %+v, want the name and price of the sale", stored.Items[0])
	}
	if stored.Total != 18000 {
		t.Errorf("total after reprice: got %d, want 18000", stored.Total)
	}
}

func TestCheckoutComputesTheKembalian(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})

	sale := checkout(t, baseURL, token, tunai(created.ID, 2, 50000))

	if sale.Total != 36000 {
		t.Fatalf("total: got %d, want 36000", sale.Total)
	}
	if sale.Payment.Change != 14000 {
		t.Errorf("Kembalian: got %d, want 50000-36000", sale.Payment.Change)
	}

	// Paying exactly the total leaves no Kembalian, and is not refused.
	exact := checkout(t, baseURL, token, tunai(created.ID, 1, 18000))
	if exact.Payment.Change != 0 {
		t.Errorf("Kembalian when paying the exact total: got %d, want 0", exact.Payment.Change)
	}
}

func TestCheckoutRefusesAPaymentBelowTheTotal(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})

	var failure errorPayload
	status := apiCall(t, http.MethodPost, baseURL+"/api/penjualan", token, tunai(created.ID, 1, 17000), &failure)

	if status != http.StatusBadRequest {
		t.Fatalf("underpaid checkout: got status %d, want %d", status, http.StatusBadRequest)
	}
	if failure.Error != "invalid_input" {
		t.Errorf("underpaid checkout: got code %q, want %q", failure.Error, "invalid_input")
	}
	if failure.Message == "" {
		t.Error("underpaid checkout: got no message, want one the till can show")
	}

	// Refused means refused: nothing was sold and no Stok left.
	if stock := stockOf(t, baseURL, token, created.ID); stock != 5 {
		t.Errorf("Stok after refused checkout: got %d, want 5", stock)
	}
}

func TestCheckoutIsBlockedWhenTheStokIsShort(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 2})

	var failure errorPayload
	status := apiCall(t, http.MethodPost, baseURL+"/api/penjualan", token, tunai(created.ID, 3, 54000), &failure)

	if status != http.StatusConflict {
		t.Fatalf("short Stok checkout: got status %d, want %d", status, http.StatusConflict)
	}
	if failure.Error != "insufficient_stock" {
		t.Errorf("short Stok checkout: got code %q, want %q", failure.Error, "insufficient_stock")
	}

	// The Stok is where it was, and the refused checkout did not consume a Nomor
	// Struk: the next sale is still the first one.
	if stock := stockOf(t, baseURL, token, created.ID); stock != 2 {
		t.Errorf("Stok after refused checkout: got %d, want 2", stock)
	}
	if sale := checkout(t, baseURL, token, tunai(created.ID, 1, 18000)); sale.ReceiptNumber != 1 {
		t.Errorf("Nomor Struk after a refused checkout: got %d, want 1", sale.ReceiptNumber)
	}
}

// TestCheckoutIsAtomicWhenOneItemIsShort is the atomicity check of the issue: a
// cart whose second Item cannot be covered must leave the first Item's Stok
// exactly as it was, and record no Penjualan at all.
func TestCheckoutIsAtomicWhenOneItemIsShort(t *testing.T) {
	baseURL, token := newAdminToken(t)
	enough := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})
	short := createProduk(t, baseURL, token, produkPayload{Name: "Teh", Price: 6000, Stock: 1})

	var failure errorPayload
	status := apiCall(t, http.MethodPost, baseURL+"/api/penjualan", token, checkoutPayload{
		Items: []checkoutItemPayload{
			{ProductID: enough.ID, Quantity: 3},
			{ProductID: short.ID, Quantity: 2},
		},
		Payment: checkoutPaymentPayload{Method: "cash", Amount: 100000},
	}, &failure)

	if status != http.StatusConflict {
		t.Fatalf("partly-short checkout: got status %d, want %d", status, http.StatusConflict)
	}

	// The first Item's Stok was not taken, even though that Item alone was fine.
	if stock := stockOf(t, baseURL, token, enough.ID); stock != 5 {
		t.Errorf("Stok of the Item that was covered: got %d, want 5", stock)
	}
	if stock := stockOf(t, baseURL, token, short.ID); stock != 1 {
		t.Errorf("Stok of the Item that was short: got %d, want 1", stock)
	}

	// Nothing was stored, and no Nomor Struk was burned.
	var notFound errorPayload
	if status := apiCall(t, http.MethodGet, baseURL+"/api/penjualan/1", token, nil, &notFound); status != http.StatusNotFound {
		t.Errorf("read the refused Penjualan: got status %d, want %d", status, http.StatusNotFound)
	}
	if sale := checkout(t, baseURL, token, tunai(enough.ID, 1, 18000)); sale.ReceiptNumber != 1 {
		t.Errorf("Nomor Struk after the refused checkout: got %d, want 1", sale.ReceiptNumber)
	}
}

func TestCheckoutRefusesAnEmptyCart(t *testing.T) {
	baseURL, token := newAdminToken(t)

	tests := []struct {
		name string
		body checkoutPayload
	}{
		{name: "no items at all", body: checkoutPayload{Payment: checkoutPaymentPayload{Method: "cash", Amount: 10000}}},
		{name: "an empty list", body: checkoutPayload{Items: []checkoutItemPayload{}, Payment: checkoutPaymentPayload{Method: "cash", Amount: 10000}}},
		{name: "a quantity of zero", body: tunai(1, 0, 10000)},
		{name: "a negative quantity", body: tunai(1, -1, 10000)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodPost, baseURL+"/api/penjualan", token, test.body, &failure)

			if status != http.StatusBadRequest {
				t.Fatalf("got status %d, want %d", status, http.StatusBadRequest)
			}
			if failure.Error != "invalid_input" {
				t.Errorf("got code %q, want %q", failure.Error, "invalid_input")
			}
		})
	}
}

func TestCheckoutMergesRepeatedProdukIntoOneItem(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})

	// The same Produk twice is one Item of the summed quantity — and, just as
	// important, the two lines cannot each pass a Stok check of 3 on their own.
	sale := checkout(t, baseURL, token, checkoutPayload{
		Items: []checkoutItemPayload{
			{ProductID: created.ID, Quantity: 2},
			{ProductID: created.ID, Quantity: 2},
		},
		Payment: checkoutPaymentPayload{Method: "cash", Amount: 72000},
	})

	if len(sale.Items) != 1 || sale.Items[0].Quantity != 4 {
		t.Fatalf("items: got %+v, want one Item of 4", sale.Items)
	}
	if sale.Total != 72000 {
		t.Errorf("total: got %d, want 72000", sale.Total)
	}
	if stock := stockOf(t, baseURL, token, created.ID); stock != 1 {
		t.Errorf("Stok: got %d, want 1", stock)
	}

	// A cart that only fits when the lines are added up is still refused.
	var failure errorPayload
	status := apiCall(t, http.MethodPost, baseURL+"/api/penjualan", token, checkoutPayload{
		Items: []checkoutItemPayload{
			{ProductID: created.ID, Quantity: 1},
			{ProductID: created.ID, Quantity: 1},
		},
		Payment: checkoutPaymentPayload{Method: "cash", Amount: 36000},
	}, &failure)
	if status != http.StatusConflict {
		t.Errorf("2+2 against a Stok of 1: got status %d, want %d", status, http.StatusConflict)
	}
}

func TestCheckoutRefusesAProdukThatIsNotSellable(t *testing.T) {
	baseURL, token := newAdminToken(t)
	nonaktif := createProduk(t, baseURL, token, produkPayload{Name: "Teh", Price: 6000, Stock: 5})

	var deactivated dataEnvelope[productEnvelope]
	if status := apiCall(t, http.MethodPatch, baseURL+"/api/produk/"+itoa(nonaktif.ID), token, activePayload{Active: false}, &deactivated); status != http.StatusOK {
		t.Fatalf("deactivate: got status %d, want %d", status, http.StatusOK)
	}

	tests := []struct {
		name string
		body checkoutPayload
	}{
		{name: "a Produk that is gone", body: tunai(404, 1, 10000)},
		{name: "a Produk that is Nonaktif", body: tunai(nonaktif.ID, 1, 10000)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodPost, baseURL+"/api/penjualan", token, test.body, &failure)

			if status != http.StatusBadRequest {
				t.Fatalf("got status %d, want %d", status, http.StatusBadRequest)
			}
			if failure.Error != "invalid_input" {
				t.Errorf("got code %q, want %q", failure.Error, "invalid_input")
			}
		})
	}

	// Refused means refused: the Nonaktif Produk's Stok is untouched.
	if stock := stockOf(t, baseURL, token, nonaktif.ID); stock != 5 {
		t.Errorf("Stok of the Nonaktif Produk: got %d, want 5", stock)
	}
}

func TestCheckoutRecordsANonTunaiPembayaran(t *testing.T) {
	baseURL, token := newAdminToken(t)

	tests := []struct {
		name   string
		method string
	}{
		{name: "QRIS", method: "qris"},
		{name: "Debit", method: "debit"},
		{name: "Transfer", method: "transfer"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Each subtest sells a Produk of its own: the suite shares one store for
			// the whole run, and one Stok would make the assertions depend on the
			// order the subtests ran in.
			created := createProduk(t, baseURL, token, produkPayload{
				Name: "Kopi " + test.name, Price: 18000, Stock: 10,
			})

			// A non-tunai Pembayaran is recorded for the total of the sale: there is
			// no gateway and no Kembalian, just the method and the nominal
			// (CONTEXT.md, Pembayaran).
			sale := checkout(t, baseURL, token, checkoutPayload{
				Items:   []checkoutItemPayload{{ProductID: created.ID, Quantity: 2}},
				Payment: checkoutPaymentPayload{Method: test.method, Amount: 36000},
			})

			if sale.Payment.Method != test.method {
				t.Errorf("method: got %q, want %q", sale.Payment.Method, test.method)
			}
			if sale.Payment.Amount != 36000 {
				t.Errorf("nominal: got %d, want 36000", sale.Payment.Amount)
			}
			if sale.Payment.Change != 0 {
				t.Errorf("Kembalian: got %d, want 0", sale.Payment.Change)
			}

			// The Pembayaran is stored, not just echoed: reading the Penjualan back by
			// its Nomor Struk answers the same method and the same zero Kembalian.
			stored := readSale(t, baseURL, token, sale.ReceiptNumber)
			if stored.Payment != sale.Payment {
				t.Errorf("read back Pembayaran: got %+v, want %+v", stored.Payment, sale.Payment)
			}

			// The sale still takes the Stok out, whatever the method was.
			if stock := stockOf(t, baseURL, token, created.ID); stock != 8 {
				t.Errorf("Stok after a %s sale: got %d, want 8", test.method, stock)
			}
		})
	}
}

func TestCheckoutRefusesANonTunaiPembayaranThatIsNotTheTotal(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})

	tests := []struct {
		name   string
		amount int64
	}{
		{name: "more than the total", amount: 40000},
		{name: "less than the total", amount: 35000},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodPost, baseURL+"/api/penjualan", token, checkoutPayload{
				Items:   []checkoutItemPayload{{ProductID: created.ID, Quantity: 2}},
				Payment: checkoutPaymentPayload{Method: "qris", Amount: test.amount},
			}, &failure)

			// There is no Kembalian to absorb an overpayment and no split payment to
			// absorb an underpayment, so a nominal that is not the total is refused
			// rather than recorded.
			if status != http.StatusBadRequest {
				t.Fatalf("got status %d, want %d", status, http.StatusBadRequest)
			}
			if failure.Error != "invalid_input" {
				t.Errorf("got code %q, want %q", failure.Error, "invalid_input")
			}
			if failure.Message == "" {
				t.Error("got no message, want one the till can show")
			}

			// Refused means nothing was sold and no Stok left.
			if stock := stockOf(t, baseURL, token, created.ID); stock != 5 {
				t.Errorf("Stok after a refused checkout: got %d, want 5", stock)
			}
		})
	}
}

func TestCheckoutRefusesAMethodItDoesNotKnow(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})

	tests := []struct {
		name   string
		method string
	}{
		{name: "something that is not a method at all", method: "bitcoin"},
		{name: "blank", method: "  "},
		{name: "an upper-cased method", method: "QRIS"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodPost, baseURL+"/api/penjualan", token, checkoutPayload{
				Items:   []checkoutItemPayload{{ProductID: created.ID, Quantity: 1}},
				Payment: checkoutPaymentPayload{Method: test.method, Amount: 18000},
			}, &failure)

			if status != http.StatusBadRequest {
				t.Fatalf("got status %d, want %d", status, http.StatusBadRequest)
			}
			if failure.Error != "invalid_input" {
				t.Errorf("got code %q, want %q", failure.Error, "invalid_input")
			}
		})
	}
}

func TestReceiptNumbersKeepCountingUp(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 10})

	first := checkout(t, baseURL, token, tunai(created.ID, 1, 18000))
	second := checkout(t, baseURL, token, tunai(created.ID, 1, 18000))
	third := checkout(t, baseURL, token, tunai(created.ID, 1, 20000))

	if second.ReceiptNumber != first.ReceiptNumber+1 || third.ReceiptNumber != second.ReceiptNumber+1 {
		t.Errorf("Nomor Struk: got %d, %d, %d, want a sequence that keeps counting",
			first.ReceiptNumber, second.ReceiptNumber, third.ReceiptNumber)
	}

	// Each sale keeps its own Items: a later checkout does not rewrite an earlier
	// one.
	if stored := readSale(t, baseURL, token, first.ReceiptNumber); stored.ReceiptNumber != first.ReceiptNumber {
		t.Errorf("read the first Penjualan: got %d, want %d", stored.ReceiptNumber, first.ReceiptNumber)
	}
}

func TestCheckoutMarksTheProdukAsSold(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 1})

	checkout(t, baseURL, token, tunai(created.ID, 1, 18000))

	// Only a checkout makes a Produk "pernah terjual", and a Produk that sold can
	// be deactivated but never deleted (#4 left this gap for #6).
	var failure errorPayload
	status := apiCall(t, http.MethodDelete, baseURL+"/api/produk/"+itoa(created.ID), token, nil, &failure)

	if status != http.StatusConflict {
		t.Fatalf("delete a Produk that sold: got status %d, want %d", status, http.StatusConflict)
	}
	if failure.Error != "product_has_sales" {
		t.Errorf("delete a Produk that sold: got code %q, want %q", failure.Error, "product_has_sales")
	}
}

func TestReadPenjualanReportsAMissingOrUnusableNomorStruk(t *testing.T) {
	baseURL, token := newAdminToken(t)

	tests := []struct {
		name          string
		path          string
		wantStatus    int
		wantErrorCode string
	}{
		{
			name: "a Penjualan that is not there", path: "/api/penjualan/404",
			wantStatus: http.StatusNotFound, wantErrorCode: "sale_not_found",
		},
		{
			name: "a Nomor Struk that is not a number", path: "/api/penjualan/abc",
			wantStatus: http.StatusBadRequest, wantErrorCode: "invalid_input",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodGet, baseURL+test.path, token, nil, &failure)

			if status != test.wantStatus {
				t.Fatalf("got status %d, want %d", status, test.wantStatus)
			}
			if failure.Error != test.wantErrorCode {
				t.Errorf("got code %q, want %q", failure.Error, test.wantErrorCode)
			}
		})
	}
}
