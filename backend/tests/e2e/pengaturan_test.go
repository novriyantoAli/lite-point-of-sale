package e2e

import (
	"net/http"
	"testing"
)

// pengaturanPayload is the body an Admin puts to change the store's Pengaturan:
// the Struk template blocks, the paper width, and the ambang Stok menipis
// (ADR-0017).
type pengaturanPayload struct {
	Header            string `json:"header"`
	Footer            string `json:"footer"`
	PaperWidth        int64  `json:"paper_width"`
	LowStockThreshold int64  `json:"low_stock_threshold"`
}

// pengaturanResponsePayload is one Pengaturan as the API answers it.
type pengaturanResponsePayload struct {
	ID                int64  `json:"id"`
	Header            string `json:"header"`
	Footer            string `json:"footer"`
	PaperWidth        int64  `json:"paper_width"`
	LowStockThreshold int64  `json:"low_stock_threshold"`
}

type pengaturanEnvelope struct {
	Settings pengaturanResponsePayload `json:"settings"`
}

func TestPengaturanRequiresAnAdmin(t *testing.T) {
	baseURL, adminToken := newAdminToken(t)
	createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")
	kasirToken := logIn(t, baseURL, "kasir1", kasirPassword)

	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "read Pengaturan", method: http.MethodGet, path: "/api/pengaturan"},
		{
			name:   "change Pengaturan",
			method: http.MethodPut,
			path:   "/api/pengaturan",
			body:   pengaturanPayload{PaperWidth: 58, LowStockThreshold: 3},
		},
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

		t.Run(test.name+" as Kasir", func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, test.method, baseURL+test.path, kasirToken, test.body, &failure)

			if status != http.StatusForbidden {
				t.Fatalf("got status %d, want %d", status, http.StatusForbidden)
			}
			if failure.Error != "forbidden" {
				t.Errorf("got code %q, want %q", failure.Error, "forbidden")
			}
		})
	}
}

func TestPengaturanSavesAndReadsBack(t *testing.T) {
	baseURL, token := newAdminToken(t)

	// Migration 0005 seeds today's values: empty template, 80 mm, ambang 5.
	var seeded dataEnvelope[pengaturanEnvelope]
	if status := apiCall(t, http.MethodGet, baseURL+"/api/pengaturan", token, nil, &seeded); status != http.StatusOK {
		t.Fatalf("read Pengaturan: got status %d, want %d", status, http.StatusOK)
	}
	if seeded.Data.Settings.PaperWidth != 80 || seeded.Data.Settings.LowStockThreshold != 5 {
		t.Errorf("seeded Pengaturan: got %+v, want paper width 80 and ambang 5", seeded.Data.Settings)
	}

	var updated dataEnvelope[pengaturanEnvelope]
	status := apiCall(t, http.MethodPut, baseURL+"/api/pengaturan", token, pengaturanPayload{
		Header:            "Toko Kopi\nJl. Melati 1",
		Footer:            "Terima kasih",
		PaperWidth:        58,
		LowStockThreshold: 3,
	}, &updated)
	if status != http.StatusOK {
		t.Fatalf("change Pengaturan: got status %d, want %d", status, http.StatusOK)
	}
	if updated.Data.Settings.Header != "Toko Kopi\nJl. Melati 1" ||
		updated.Data.Settings.PaperWidth != 58 ||
		updated.Data.Settings.LowStockThreshold != 3 {
		t.Errorf("changed Pengaturan: got %+v, want the written values", updated.Data.Settings)
	}

	// A later read answers the same stored row.
	var readBack dataEnvelope[pengaturanEnvelope]
	if status := apiCall(t, http.MethodGet, baseURL+"/api/pengaturan", token, nil, &readBack); status != http.StatusOK {
		t.Fatalf("read back Pengaturan: got status %d, want %d", status, http.StatusOK)
	}
	if readBack.Data.Settings.Header != "Toko Kopi\nJl. Melati 1" ||
		readBack.Data.Settings.LowStockThreshold != 3 {
		t.Errorf("read back Pengaturan: got %+v, want the saved values", readBack.Data.Settings)
	}
}

func TestPengaturanValidatesItsInput(t *testing.T) {
	baseURL, token := newAdminToken(t)

	tests := []struct {
		name string
		body pengaturanPayload
	}{
		{name: "paper width other than 58 or 80", body: pengaturanPayload{PaperWidth: 60, LowStockThreshold: 5}},
		{name: "an ambang of zero", body: pengaturanPayload{PaperWidth: 80, LowStockThreshold: 0}},
		{name: "a negative ambang", body: pengaturanPayload{PaperWidth: 80, LowStockThreshold: -5}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodPut, baseURL+"/api/pengaturan", token, test.body, &failure)

			if status != http.StatusBadRequest {
				t.Fatalf("got status %d, want %d", status, http.StatusBadRequest)
			}
			if failure.Error != "invalid_input" {
				t.Errorf("got code %q, want %q", failure.Error, "invalid_input")
			}
			if failure.Message == "" {
				t.Error("got no message, want one the form can show")
			}
		})
	}
}

// The ambang Stok menipis is a stored setting now, so changing it has to change
// what the restock list answers — without changing the list's shape.
func TestTheStoredThresholdDrivesTheRestockList(t *testing.T) {
	baseURL, token := newAdminToken(t)

	// At the seeded ambang of 5, a Produk at 4 is menipis and one at 6 is not.
	menipis := createProduk(t, baseURL, token, produkPayload{Name: "Menipis", Price: 1000, Stock: 4})
	aman := createProduk(t, baseURL, token, produkPayload{Name: "Aman", Price: 1000, Stock: 6})

	low := listLowStock(t, baseURL, token)
	if low.Threshold != 5 {
		t.Fatalf("threshold before: got %d, want the seeded 5", low.Threshold)
	}

	// Raise the ambang above both: the list is re-read from the stored setting,
	// not from a constant the handler remembers.
	var updated dataEnvelope[pengaturanEnvelope]
	if status := apiCall(t, http.MethodPut, baseURL+"/api/pengaturan", token, pengaturanPayload{
		Header:            "",
		Footer:            "",
		PaperWidth:        80,
		LowStockThreshold: 7,
	}, &updated); status != http.StatusOK {
		t.Fatalf("change Pengaturan: got status %d, want %d", status, http.StatusOK)
	}

	low = listLowStock(t, baseURL, token)
	if low.Threshold != 7 {
		t.Fatalf("threshold after: got %d, want 7", low.Threshold)
	}

	// Both Produk are now below the stored ambang.
	got := map[int64]bool{}
	for _, product := range low.Products {
		got[product.ID] = true
	}
	if !got[menipis.ID] || !got[aman.ID] {
		t.Errorf("restock list after raising the ambang: got %+v, want both %d and %d",
			low.Products, menipis.ID, aman.ID)
	}
}
