package e2e

import (
	"net/http"
	"testing"
)

// saleSummaryPayload is one row of the sales list (#9): the Nomor Struk and the
// few fields the list shows, without the Item lines a detail read carries.
type saleSummaryPayload struct {
	ReceiptNumber int64  `json:"receipt_number"`
	CreatedAt     string `json:"created_at"`
	CashierID     int64  `json:"cashier_id"`
	CashierName   string `json:"cashier_name"`
	Total         int64  `json:"total"`
	Method        string `json:"method"`
}

// methodTotalPayload is what one Pembayaran method contributed to a day.
type methodTotalPayload struct {
	Method       string `json:"method"`
	Total        int64  `json:"total"`
	Transactions int64  `json:"transactions"`
}

// cashierTotalPayload is what one Kasir rang up in a day.
type cashierTotalPayload struct {
	CashierID    int64  `json:"cashier_id"`
	CashierName  string `json:"cashier_name"`
	Total        int64  `json:"total"`
	Transactions int64  `json:"transactions"`
}

// revenuePayload is the omzet of one store-local day: the total, the number of
// Penjualan, and the breakdown by method and by Kasir.
type revenuePayload struct {
	Date         string                `json:"date"`
	Total        int64                 `json:"total"`
	Transactions int64                 `json:"transactions"`
	ByMethod     []methodTotalPayload  `json:"by_method"`
	ByCashier    []cashierTotalPayload `json:"by_cashier"`
}

// listPenjualan answers the sales list of a day, failing the test when the API
// refused it.
func listPenjualan(t *testing.T, baseURL, token, date string) []saleSummaryPayload {
	t.Helper()

	var listed dataEnvelope[[]saleSummaryPayload]
	status := apiCall(t, http.MethodGet, baseURL+"/api/penjualan?date="+date, token, nil, &listed)

	if status != http.StatusOK {
		t.Fatalf("list Penjualan %s: got status %d, want %d", date, status, http.StatusOK)
	}

	return listed.Data
}

// dailyRevenue answers the omzet report of a day, failing the test when the API
// refused it.
func dailyRevenue(t *testing.T, baseURL, token, date string) revenuePayload {
	t.Helper()

	var report dataEnvelope[revenuePayload]
	status := apiCall(t, http.MethodGet, baseURL+"/api/penjualan/omzet?date="+date, token, nil, &report)

	if status != http.StatusOK {
		t.Fatalf("daily revenue %s: got status %d, want %d", date, status, http.StatusOK)
	}

	return report.Data
}

// qris builds a non-tunai checkout of one line: the method and the nominal are
// recorded, and the nominal is the total of the sale (ADR-0016).
func qris(productID, quantity, amount int64) checkoutPayload {
	return checkoutPayload{
		Items:   []checkoutItemPayload{{ProductID: productID, Quantity: quantity}},
		Payment: checkoutPaymentPayload{Method: "qris", Amount: amount},
	}
}

func TestLaporanRequiresAnAdmin(t *testing.T) {
	baseURL, adminToken := newAdminToken(t)
	createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")
	kasirToken := logIn(t, baseURL, "kasir1", kasirPassword)

	// The reports are the Admin's: the omzet and the sales list are not the
	// till's screens. Both Peran still read one Penjualan and reprint its Struk.
	tests := []struct {
		name string
		path string
	}{
		{name: "sales list", path: "/api/penjualan"},
		{name: "daily revenue", path: "/api/penjualan/omzet"},
	}

	for _, test := range tests {
		t.Run(test.name+" as Kasir", func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodGet, baseURL+test.path, kasirToken, nil, &failure)

			if status != http.StatusForbidden {
				t.Fatalf("got status %d, want %d", status, http.StatusForbidden)
			}
			if failure.Error != "forbidden" {
				t.Errorf("got code %q, want %q", failure.Error, "forbidden")
			}
		})

		t.Run(test.name+" anonymously", func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodGet, baseURL+test.path, "", nil, &failure)

			if status != http.StatusUnauthorized {
				t.Fatalf("got status %d, want %d", status, http.StatusUnauthorized)
			}
			if failure.Error != "invalid_token" {
				t.Errorf("got code %q, want %q", failure.Error, "invalid_token")
			}
		})
	}
}

func TestSalesListAnswersTheDaysPenjualan(t *testing.T) {
	baseURL, token := newAdminToken(t)
	kopi := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 10})

	first := checkout(t, baseURL, token, tunai(kopi.ID, 1, 18000))
	second := checkout(t, baseURL, token, tunai(kopi.ID, 2, 36000))

	listed := listPenjualan(t, baseURL, token, first.CreatedAt[:10])

	if len(listed) != 2 {
		t.Fatalf("sales: got %d, want 2", len(listed))
	}
	// Newest Nomor Struk first, which is the order the screen reads.
	if listed[0].ReceiptNumber != second.ReceiptNumber || listed[1].ReceiptNumber != first.ReceiptNumber {
		t.Errorf("order: got %d then %d, want newest first",
			listed[0].ReceiptNumber, listed[1].ReceiptNumber)
	}
	if listed[0].Total != 36000 || listed[0].Method != "cash" {
		t.Errorf("row: got %+v, want the total and method of the sale", listed[0])
	}
	if listed[0].CashierName != seededAdmin || listed[0].CreatedAt == "" {
		t.Errorf("row: got %+v, want the cashier and timestamp of the sale", listed[0])
	}
}

func TestDailyRevenueAggregatesTheDay(t *testing.T) {
	baseURL, adminToken := newAdminToken(t)
	kopi := createProduk(t, baseURL, adminToken, produkPayload{Name: "Kopi", Price: 18000, Stock: 20})
	createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")
	kasirToken := logIn(t, baseURL, "kasir1", kasirPassword)

	// Three sales: two by the Admin (one Tunai, one QRIS), one by the Kasir.
	first := checkout(t, baseURL, adminToken, tunai(kopi.ID, 1, 18000))
	checkout(t, baseURL, adminToken, qris(kopi.ID, 2, 36000))
	checkout(t, baseURL, kasirToken, tunai(kopi.ID, 1, 18000))

	report := dailyRevenue(t, baseURL, adminToken, first.CreatedAt[:10])

	if report.Total != 72000 || report.Transactions != 3 {
		t.Errorf("totals: got %d/%d, want 72000/3", report.Total, report.Transactions)
	}
	if report.Date != first.CreatedAt[:10] {
		t.Errorf("date: got %q, want %q", report.Date, first.CreatedAt[:10])
	}

	// All four methods, in the order the till offers them — a method nobody used
	// is a zero row, not a missing one.
	want := []methodTotalPayload{
		{Method: "cash", Total: 36000, Transactions: 2},
		{Method: "qris", Total: 36000, Transactions: 1},
		{Method: "debit"},
		{Method: "transfer"},
	}
	if len(report.ByMethod) != len(want) {
		t.Fatalf("by method: got %d rows, want %d", len(report.ByMethod), len(want))
	}
	for i, total := range want {
		if report.ByMethod[i] != total {
			t.Errorf("by method %d: got %+v, want %+v", i, report.ByMethod[i], total)
		}
	}

	// Per Kasir, biggest first.
	if len(report.ByCashier) != 2 {
		t.Fatalf("by Kasir: got %d rows, want 2", len(report.ByCashier))
	}
	if report.ByCashier[0].CashierName != seededAdmin || report.ByCashier[0].Total != 54000 ||
		report.ByCashier[0].Transactions != 2 {
		t.Errorf("by Kasir[0]: got %+v, want the Admin with 54000 in 2 Penjualan", report.ByCashier[0])
	}
	if report.ByCashier[1].CashierName != "kasir1" || report.ByCashier[1].Total != 18000 {
		t.Errorf("by Kasir[1]: got %+v, want kasir1 with 18000", report.ByCashier[1])
	}
}

func TestLaporanAnswersAnEmptyDay(t *testing.T) {
	baseURL, token := newAdminToken(t)
	kopi := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 10})
	checkout(t, baseURL, token, tunai(kopi.ID, 1, 18000))

	const kosong = "2000-01-01"

	listed := listPenjualan(t, baseURL, token, kosong)
	if len(listed) != 0 {
		t.Errorf("sales: got %d, want none", len(listed))
	}

	report := dailyRevenue(t, baseURL, token, kosong)
	if report.Date != kosong || report.Total != 0 || report.Transactions != 0 {
		t.Errorf("report: got %+v, want an empty day", report)
	}
	// Even a day with no sales answers the four methods, so the screen has a
	// complete breakdown to render.
	if len(report.ByMethod) != 4 {
		t.Fatalf("by method: got %d rows, want 4", len(report.ByMethod))
	}
	for _, total := range report.ByMethod {
		if total.Total != 0 || total.Transactions != 0 {
			t.Errorf("by method %q: got %+v, want a zero row", total.Method, total)
		}
	}
	if len(report.ByCashier) != 0 {
		t.Errorf("by Kasir: got %d rows, want none", len(report.ByCashier))
	}
}

func TestLaporanRefusesAMalformedDate(t *testing.T) {
	baseURL, token := newAdminToken(t)

	tests := []struct {
		name string
		path string
	}{
		{name: "sales list", path: "/api/penjualan?date=23-09-2026"},
		{name: "daily revenue", path: "/api/penjualan/omzet?date=kemarin"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodGet, baseURL+test.path, token, nil, &failure)

			if status != http.StatusBadRequest {
				t.Fatalf("got status %d, want %d", status, http.StatusBadRequest)
			}
			if failure.Error != "invalid_input" {
				t.Errorf("got code %q, want %q", failure.Error, "invalid_input")
			}
			if failure.Message == "" {
				t.Error("got no message, want one the screen can show")
			}
		})
	}
}
