package e2e

import (
	"net/http"
	"testing"
)

// seededLowStockThreshold is the ambang Stok menipis migration 0005 seeds. The
// test states its own number because the constant that once lived in the domain
// has moved into the stored Pengaturan (ADR-0017).
const seededLowStockThreshold int64 = 5

// addStockPayload is the body an Admin posts to record a restock: how many
// units arrived, never the new total — the Stok that is already there is the
// API's to know (see `usecase/produk.AddStock`).
type addStockPayload struct {
	Quantity int64 `json:"quantity"`
}

// lowStockPayload is what Go answers to the restock list. The threshold travels
// with the Produk because the UI shows the Admin the rule it is reading
// ("Stok menipis ≤ 5"), and a rule the frontend hard-codes is a second copy of
// it that can drift.
type lowStockPayload struct {
	Threshold int64            `json:"threshold"`
	Products  []productPayload `json:"products"`
}

func TestStokRequiresAnAdmin(t *testing.T) {
	baseURL, adminToken := newAdminToken(t)
	createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")
	kasirToken := logIn(t, baseURL, "kasir1", kasirPassword)

	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "add Stok", method: http.MethodPost, path: "/api/produk/1/stok", body: addStockPayload{Quantity: 1}},
		{name: "list Stok menipis", method: http.MethodGet, path: "/api/produk/stok-menipis"},
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

func TestAddStokIncreasesStok(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 10})

	added := addStock(t, baseURL, token, created.ID, 3)
	if added.Stock != 13 {
		t.Errorf("add Stok: got stock %d, want 13", added.Stock)
	}

	// The answer is the Stok that is now stored, not a number the handler
	// computed: the next read has to agree with it.
	listed := listProduk(t, baseURL, token, "")
	if len(listed) != 1 || listed[0].Stock != 13 {
		t.Errorf("list after restock: got %+v, want one Produk with stock 13", listed)
	}
}

func TestAddStokAddsToWhatIsAlreadyThere(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 4})

	added := addStock(t, baseURL, token, created.ID, 2)
	if added.Stock != 6 {
		t.Fatalf("add Stok: got stock %d, want 6", added.Stock)
	}

	// A second delivery adds to the first, rather than replacing it.
	added = addStock(t, baseURL, token, created.ID, 5)
	if added.Stock != 11 {
		t.Errorf("second add Stok: got stock %d, want 11", added.Stock)
	}
}

func TestAddStokLeavesTheRestOfTheProdukAlone(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{
		Name:     "Kopi Susu",
		Code:     strPtr("KOPI-01"),
		Price:    18000,
		Category: strPtr("Minuman"),
		Stock:    10,
	})

	added := addStock(t, baseURL, token, created.ID, 1)

	if added.Name != "Kopi Susu" || added.Price != 18000 {
		t.Errorf("add Stok: got %+v, want the name and Harga to stay put", added)
	}
	if added.Code == nil || *added.Code != "KOPI-01" {
		t.Errorf("add Stok: got code %v, want KOPI-01", added.Code)
	}
	if added.Category == nil || *added.Category != "Minuman" {
		t.Errorf("add Stok: got category %v, want Minuman", added.Category)
	}
	if !added.Active {
		t.Error("add Stok: got nonaktif, want the Produk to stay active")
	}
}

func TestAddStokValidatesTheQuantity(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 10})

	tests := []struct {
		name string
		body any
	}{
		{name: "zero", body: addStockPayload{Quantity: 0}},
		{name: "negative", body: addStockPayload{Quantity: -5}},
		{name: "a body without a quantity", body: map[string]any{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodPost, baseURL+"/api/produk/"+itoa(created.ID)+"/stok", token, test.body, &failure)

			if status != http.StatusBadRequest {
				t.Fatalf("add Stok: got status %d, want %d", status, http.StatusBadRequest)
			}
			if failure.Error != "invalid_input" {
				t.Errorf("add Stok: got code %q, want %q", failure.Error, "invalid_input")
			}
			if failure.Message == "" {
				t.Error("add Stok: got no message, want one the form can show")
			}
		})
	}

	// Refused means refused: the Stok is where it was.
	if listed := listProduk(t, baseURL, token, ""); listed[0].Stock != 10 {
		t.Errorf("list after refused restocks: got stock %d, want 10", listed[0].Stock)
	}
}

func TestAddStokReportsAMissingOrUnusableProduk(t *testing.T) {
	baseURL, token := newAdminToken(t)

	tests := []struct {
		name          string
		path          string
		wantStatus    int
		wantErrorCode string
	}{
		{
			name: "a Produk that is not there", path: "/api/produk/404/stok",
			wantStatus: http.StatusNotFound, wantErrorCode: "product_not_found",
		},
		{
			name: "an id that is not a number", path: "/api/produk/abc/stok",
			wantStatus: http.StatusBadRequest, wantErrorCode: "invalid_input",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodPost, baseURL+test.path, token, addStockPayload{Quantity: 1}, &failure)

			if status != test.wantStatus {
				t.Fatalf("got status %d, want %d", status, test.wantStatus)
			}
			if failure.Error != test.wantErrorCode {
				t.Errorf("got code %q, want %q", failure.Error, test.wantErrorCode)
			}
		})
	}
}

func TestAddStokWorksOnANonaktifProduk(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Teh", Price: 6000, Stock: 0})

	var updated dataEnvelope[productEnvelope]
	status := apiCall(t, http.MethodPatch, baseURL+"/api/produk/"+itoa(created.ID), token,
		activePayload{Active: false}, &updated)
	if status != http.StatusOK {
		t.Fatalf("deactivate: got status %d, want %d", status, http.StatusOK)
	}

	// A Produk that is not for sale yet is exactly the one an Admin restocks
	// before reactivating it (CONTEXT.md, Nonaktif).
	added := addStock(t, baseURL, token, created.ID, 6)
	if added.Stock != 6 {
		t.Errorf("add Stok to a Nonaktif Produk: got stock %d, want 6", added.Stock)
	}
	if added.Active {
		t.Error("add Stok to a Nonaktif Produk: got active, want it to stay Nonaktif")
	}
}

func TestStokMenipisAnswersTheActiveProdukToRestock(t *testing.T) {
	baseURL, token := newAdminToken(t)

	habis := createProduk(t, baseURL, token, produkPayload{Name: "Habis", Price: 1000, Stock: 0})
	menipis := createProduk(t, baseURL, token, produkPayload{Name: "Menipis", Price: 1000, Stock: 2})
	aman := createProduk(t, baseURL, token, produkPayload{Name: "Aman", Price: 1000, Stock: 40})

	// The ambang itself is the first Stok that is still enough, so the Produk
	// sitting exactly on it is not menipis (issue #5: "di bawah ambang atau nol").
	diAmbang := createProduk(t, baseURL, token, produkPayload{
		Name:  "Di Ambang",
		Price: 1000,
		Stock: seededLowStockThreshold,
	})

	// A Produk that is not for sale is not one that runs out, so it stays off
	// the restock list even at Stok 0.
	retired := createProduk(t, baseURL, token, produkPayload{Name: "Nonaktif", Price: 1000, Stock: 0})
	var updated dataEnvelope[productEnvelope]
	if status := apiCall(t, http.MethodPatch, baseURL+"/api/produk/"+itoa(retired.ID), token,
		activePayload{Active: false}, &updated); status != http.StatusOK {
		t.Fatalf("deactivate: got status %d, want %d", status, http.StatusOK)
	}

	low := listLowStock(t, baseURL, token)

	// The answered threshold is the stored setting's, not a number the adapter
	// invented: it is what the UI shows the Admin as the rule behind the list.
	if low.Threshold != seededLowStockThreshold {
		t.Errorf("threshold: got %d, want %d", low.Threshold, seededLowStockThreshold)
	}

	want := []int64{habis.ID, menipis.ID}
	if len(low.Products) != len(want) {
		t.Fatalf("Stok menipis: got %d Produk, want %d (%+v)", len(low.Products), len(want), low.Products)
	}
	for i, id := range want {
		if low.Products[i].ID != id {
			t.Errorf("Stok menipis[%d]: got id %d (%s), want %d (thinnest first)",
				i, low.Products[i].ID, low.Products[i].Name, id)
		}
	}

	// Every Produk in the list is below the answered threshold, and the ones that
	// are not — including the one sitting exactly on it — were left out.
	for _, product := range low.Products {
		if product.Stock >= low.Threshold {
			t.Errorf("Stok menipis: got %s with stock %d, not below the answered threshold %d",
				product.Name, product.Stock, low.Threshold)
		}
	}
	for name, product := range map[string]productPayload{"Aman": aman, "Di Ambang": diAmbang} {
		if product.Stock < low.Threshold {
			t.Errorf("test setup: %s has stock %d, want it not below the threshold %d",
				name, product.Stock, low.Threshold)
		}
	}
}

func TestStokMenipisEmptiesWhenEverythingIsRestocked(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 0})

	if low := listLowStock(t, baseURL, token); len(low.Products) != 1 {
		t.Fatalf("Stok menipis before restock: got %d Produk, want 1", len(low.Products))
	}

	addStock(t, baseURL, token, created.ID, 20)

	if low := listLowStock(t, baseURL, token); len(low.Products) != 0 {
		t.Errorf("Stok menipis after restock: got %d Produk, want none", len(low.Products))
	}
}

// addStock records a restock as an Admin and returns the Produk as the API
// answers it, failing the test when the restock did not go through.
func addStock(t *testing.T, baseURL, token string, id int64, quantity int64) productPayload {
	t.Helper()

	var added dataEnvelope[productEnvelope]
	status := apiCall(t, http.MethodPost, baseURL+"/api/produk/"+itoa(id)+"/stok", token,
		addStockPayload{Quantity: quantity}, &added)

	if status != http.StatusOK {
		t.Fatalf("add Stok %d to Produk %d: got status %d, want %d", quantity, id, status, http.StatusOK)
	}

	return added.Data.Product
}

// listLowStock answers the restock list as an Admin, failing the test when
// the API refused it.
func listLowStock(t *testing.T, baseURL, token string) lowStockPayload {
	t.Helper()

	var low dataEnvelope[lowStockPayload]
	status := apiCall(t, http.MethodGet, baseURL+"/api/produk/stok-menipis", token, nil, &low)

	if status != http.StatusOK {
		t.Fatalf("list Stok menipis: got status %d, want %d", status, http.StatusOK)
	}

	return low.Data
}
