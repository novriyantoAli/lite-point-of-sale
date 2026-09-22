package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/sqlite"
)

// produkPayload is the body an Admin posts to add or change a Produk. Code and
// Category are pointers so a test can tell "no Kode" from "Kode of spaces".
type produkPayload struct {
	Name     string  `json:"name"`
	Code     *string `json:"code"`
	Price    int64   `json:"price"`
	Category *string `json:"category"`
	Stock    int64   `json:"stock"`
	// Active is only read on create; leaving it out means Aktif.
	Active *bool `json:"active"`
}

// productPayload is one Produk as the API answers it.
type productPayload struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Code     *string `json:"code"`
	Price    int64   `json:"price"`
	Category *string `json:"category"`
	Stock    int64   `json:"stock"`
	Active   bool    `json:"active"`
	Sold     bool    `json:"sold"`
}

type productEnvelope struct {
	Product productPayload `json:"product"`
}

// newAdminToken boots the API and logs in as the seeded Admin, which is who
// every Produk route requires.
func newAdminToken(t *testing.T) (baseURL, token string) {
	t.Helper()

	_, baseURL = startAPI(t, newTestConfig(t))

	return baseURL, logIn(t, baseURL, seededAdmin, testAdminPassword)
}

func strPtr(value string) *string { return &value }

func TestProdukRequiresAnAdmin(t *testing.T) {
	baseURL, adminToken := newAdminToken(t)
	createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")
	kasirToken := logIn(t, baseURL, "kasir1", kasirPassword)

	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "list the catalogue", method: http.MethodGet, path: "/api/produk"},
		{name: "list Kategori", method: http.MethodGet, path: "/api/produk/kategori"},
		{name: "create a Produk", method: http.MethodPost, path: "/api/produk", body: produkPayload{Name: "Kopi", Price: 18000}},
		{name: "change a Produk", method: http.MethodPut, path: "/api/produk/1", body: produkPayload{Name: "Kopi", Price: 19000}},
		{name: "deactivate a Produk", method: http.MethodPatch, path: "/api/produk/1", body: activePayload{Active: false}},
		{name: "delete a Produk", method: http.MethodDelete, path: "/api/produk/1"},
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

func TestCreateProdukWithEveryField(t *testing.T) {
	baseURL, token := newAdminToken(t)

	created := createProduk(t, baseURL, token, produkPayload{
		Name:     "  Kopi Susu ",
		Code:     strPtr("KOPI-01"),
		Price:    18000,
		Category: strPtr("Minuman"),
		Stock:    12,
	})

	if created.Name != "Kopi Susu" {
		t.Errorf("name: got %q, want the trimmed %q", created.Name, "Kopi Susu")
	}
	if created.Code == nil || *created.Code != "KOPI-01" {
		t.Errorf("code: got %v, want KOPI-01", created.Code)
	}
	if created.Category == nil || *created.Category != "Minuman" {
		t.Errorf("category: got %v, want Minuman", created.Category)
	}
	if created.Price != 18000 || created.Stock != 12 {
		t.Errorf("price/stock: got %d/%d, want 18000/12", created.Price, created.Stock)
	}
	if !created.Active {
		t.Error("active: got false, want a new Produk with no Status asked for to be active")
	}
	if created.Sold {
		t.Error("sold: got true, want a new Produk to have no sales")
	}
}

func TestCreateProdukCanArriveNonaktif(t *testing.T) {
	baseURL, token := newAdminToken(t)

	// An Admin entering something not yet for sale does not have to add it and
	// then deactivate it.
	inactive := false
	created := createProduk(t, baseURL, token, produkPayload{
		Name:   "Belum Dijual",
		Price:  1000,
		Stock:  1,
		Active: &inactive,
	})

	if created.Active {
		t.Error("active: got true, want the Nonaktif the Admin asked for")
	}

	// The catalogue screen asks for everything and still sees it…
	if all := listProduk(t, baseURL, token, ""); len(all) != 1 {
		t.Errorf("list all: got %d Produk, want 1", len(all))
	}

	// …while `active=true` — the kasir lookup of #6 — does not.
	if active := listProduk(t, baseURL, token, "active=true"); len(active) != 0 {
		t.Errorf("list active: got %d Produk, want none", len(active))
	}
}

func TestCreateProdukAllowsAnAbsentKode(t *testing.T) {
	baseURL, token := newAdminToken(t)

	// Two Produk without a Kode: an optional unique Kode has to allow this, or
	// the catalogue could hold only one of them.
	first := createProduk(t, baseURL, token, produkPayload{Name: "Air Mineral", Price: 3000, Stock: 5})
	second := createProduk(t, baseURL, token, produkPayload{Name: "Teh Botol", Price: 5000, Stock: 5})

	for name, product := range map[string]productPayload{"first": first, "second": second} {
		if product.Code != nil {
			t.Errorf("%s: got code %q, want none", name, *product.Code)
		}
		if product.Category != nil {
			t.Errorf("%s: got category %q, want none", name, *product.Category)
		}
	}
}

func TestCreateProdukRejectsADuplicateKode(t *testing.T) {
	baseURL, token := newAdminToken(t)
	createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Code: strPtr("KOPI-01"), Price: 18000})

	var failure errorPayload
	status := apiCall(t, http.MethodPost, baseURL+"/api/produk", token,
		produkPayload{Name: "Kopi Besar", Code: strPtr("KOPI-01"), Price: 20000}, &failure)

	if status != http.StatusConflict {
		t.Fatalf("create duplicate Kode: got status %d, want %d", status, http.StatusConflict)
	}
	if failure.Error != "code_taken" {
		t.Errorf("create duplicate Kode: got code %q, want %q", failure.Error, "code_taken")
	}
	if failure.Message == "" {
		t.Error("create duplicate Kode: got no message, want one the form can show")
	}
}

func TestCreateProdukValidatesInput(t *testing.T) {
	baseURL, token := newAdminToken(t)

	tests := []struct {
		name string
		body any
	}{
		{name: "empty name", body: produkPayload{Name: "   ", Price: 1000}},
		{name: "negative price", body: produkPayload{Name: "Kopi", Price: -1}},
		{name: "negative stock", body: produkPayload{Name: "Kopi", Price: 1000, Stock: -1}},
		{name: "body that is not JSON at all", body: "bukan json"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodPost, baseURL+"/api/produk", token, test.body, &failure)

			if status != http.StatusBadRequest {
				t.Fatalf("create: got status %d, want %d", status, http.StatusBadRequest)
			}
			if failure.Error != "invalid_input" {
				t.Errorf("create: got code %q, want %q", failure.Error, "invalid_input")
			}
			if failure.Message == "" {
				t.Error("create: got no message, want one the form can show")
			}
		})
	}
}

func TestUpdateProdukReplacesTheEditableFields(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Code: strPtr("KOPI-01"), Price: 18000, Stock: 10})

	var updated dataEnvelope[productEnvelope]
	status := apiCall(t, http.MethodPut, baseURL+"/api/produk/"+itoa(created.ID), token, produkPayload{
		Name:     "Kopi Susu Gula Aren",
		Code:     strPtr("KOPI-02"),
		Price:    22000,
		Category: strPtr("Minuman"),
		Stock:    7,
	}, &updated)

	if status != http.StatusOK {
		t.Fatalf("update: got status %d, want %d", status, http.StatusOK)
	}
	product := updated.Data.Product
	if product.Name != "Kopi Susu Gula Aren" || product.Price != 22000 || product.Stock != 7 {
		t.Errorf("update: got %+v, want the new name, price and stock", product)
	}
	if product.Code == nil || *product.Code != "KOPI-02" {
		t.Errorf("update: got code %v, want KOPI-02", product.Code)
	}
	if !product.Active {
		t.Error("update: got nonaktif, want the Produk to stay active")
	}
}

func TestUpdateProdukRejectsAKodeAnotherProdukUses(t *testing.T) {
	baseURL, token := newAdminToken(t)
	createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Code: strPtr("KOPI-01"), Price: 18000})
	other := createProduk(t, baseURL, token, produkPayload{Name: "Teh", Code: strPtr("TEH-01"), Price: 6000})

	var failure errorPayload
	status := apiCall(t, http.MethodPut, baseURL+"/api/produk/"+itoa(other.ID), token,
		produkPayload{Name: "Teh Manis", Code: strPtr("KOPI-01"), Price: 6000}, &failure)

	if status != http.StatusConflict {
		t.Fatalf("update to a taken Kode: got status %d, want %d", status, http.StatusConflict)
	}
	if failure.Error != "code_taken" {
		t.Errorf("update to a taken Kode: got code %q, want %q", failure.Error, "code_taken")
	}
}

func TestDeactivateProdukHidesItFromAnActiveLookup(t *testing.T) {
	baseURL, token := newAdminToken(t)
	kept := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000})
	retired := createProduk(t, baseURL, token, produkPayload{Name: "Teh", Price: 6000})

	var updated dataEnvelope[productEnvelope]
	status := apiCall(t, http.MethodPatch, baseURL+"/api/produk/"+itoa(retired.ID), token,
		activePayload{Active: false}, &updated)

	if status != http.StatusOK {
		t.Fatalf("deactivate: got status %d, want %d", status, http.StatusOK)
	}
	if updated.Data.Product.Active {
		t.Fatal("deactivate: got an active Produk, want Nonaktif")
	}

	// The catalogue screen asks for everything and still sees both…
	all := listProduk(t, baseURL, token, "")
	if len(all) != 2 {
		t.Errorf("list all: got %d Produk, want 2", len(all))
	}

	// …while `active=true` — the kasir lookup of #6 — no longer offers the
	// Nonaktif one.
	active := listProduk(t, baseURL, token, "active=true")
	if len(active) != 1 {
		t.Fatalf("list active: got %d Produk, want 1", len(active))
	}
	if active[0].ID != kept.ID {
		t.Errorf("list active: got id %d, want the active Produk %d", active[0].ID, kept.ID)
	}
}

func TestDeleteProdukRemovesOneThatNeverSold(t *testing.T) {
	baseURL, token := newAdminToken(t)
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000})

	var failure errorPayload
	// A 204 has no body to decode, so `out` stays nil for the successful delete.
	status := apiCall(t, http.MethodDelete, baseURL+"/api/produk/"+itoa(created.ID), token, nil, nil)
	if status != http.StatusNoContent {
		t.Fatalf("delete: got status %d, want %d", status, http.StatusNoContent)
	}

	if listed := listProduk(t, baseURL, token, ""); len(listed) != 0 {
		t.Errorf("list after delete: got %d Produk, want none", len(listed))
	}

	// A Produk that is already gone is a 404, not a silent success.
	if status := apiCall(t, http.MethodDelete, baseURL+"/api/produk/"+itoa(created.ID), token, nil, &failure); status != http.StatusNotFound {
		t.Errorf("delete twice: got status %d, want %d", status, http.StatusNotFound)
	}
}

func TestDeleteProdukIsRefusedOnceItSold(t *testing.T) {
	cfg := newTestConfig(t)
	_, baseURL := startAPI(t, cfg)
	token := logIn(t, baseURL, seededAdmin, testAdminPassword)

	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000})
	markSold(t, cfg.DBPath, created.ID)

	var failure errorPayload
	status := apiCall(t, http.MethodDelete, baseURL+"/api/produk/"+itoa(created.ID), token, nil, &failure)

	if status != http.StatusConflict {
		t.Fatalf("delete sold Produk: got status %d, want %d", status, http.StatusConflict)
	}
	if failure.Error != "product_has_sales" {
		t.Errorf("delete sold Produk: got code %q, want %q", failure.Error, "product_has_sales")
	}

	// Refused means refused: the Produk is still there to be deactivated instead.
	if listed := listProduk(t, baseURL, token, ""); len(listed) != 1 {
		t.Errorf("list after refused delete: got %d Produk, want the Produk to survive", len(listed))
	}
}

func TestListProdukSearchesByNameKodeAndKategori(t *testing.T) {
	baseURL, token := newAdminToken(t)
	createProduk(t, baseURL, token, produkPayload{Name: "Kopi Susu", Code: strPtr("KOPI-01"), Category: strPtr("Minuman"), Price: 18000})
	createProduk(t, baseURL, token, produkPayload{Name: "Teh Manis", Code: strPtr("TEH-01"), Category: strPtr("Minuman"), Price: 6000})
	createProduk(t, baseURL, token, produkPayload{Name: "Roti Bakar", Category: strPtr("Makanan"), Price: 15000})

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{name: "no filter", query: "", want: []string{"Kopi Susu", "Roti Bakar", "Teh Manis"}},
		{name: "by name, whatever the case", query: "name=kopi", want: []string{"Kopi Susu"}},
		{name: "by Kode", query: "code=TEH", want: []string{"Teh Manis"}},
		{name: "by Kategori", query: "category=Minuman", want: []string{"Kopi Susu", "Teh Manis"}},
		{name: "by Kategori and name", query: "category=Minuman&name=teh", want: []string{"Teh Manis"}},
		{name: "no match", query: "name=hantu", want: []string{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			listed := listProduk(t, baseURL, token, test.query)

			if len(listed) != len(test.want) {
				t.Fatalf("list: got %d Produk, want %d", len(listed), len(test.want))
			}
			for i, name := range test.want {
				if listed[i].Name != name {
					t.Errorf("list[%d]: got %q, want %q", i, listed[i].Name, name)
				}
			}
		})
	}
}

func TestListProdukRejectsAnUnusableActiveFilter(t *testing.T) {
	baseURL, token := newAdminToken(t)

	var failure errorPayload
	status := apiCall(t, http.MethodGet, baseURL+"/api/produk?active=maybe", token, nil, &failure)

	if status != http.StatusBadRequest {
		t.Fatalf("list with a bad active filter: got status %d, want %d", status, http.StatusBadRequest)
	}
	if failure.Error != "invalid_input" {
		t.Errorf("list with a bad active filter: got code %q, want %q", failure.Error, "invalid_input")
	}
}

func TestListKategoriAnswersTheOnesInUse(t *testing.T) {
	baseURL, token := newAdminToken(t)

	createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Category: strPtr("Minuman"), Price: 18000})
	createProduk(t, baseURL, token, produkPayload{Name: "Teh", Category: strPtr("Minuman"), Price: 6000})
	createProduk(t, baseURL, token, produkPayload{Name: "Roti", Category: strPtr("Makanan"), Price: 15000})
	createProduk(t, baseURL, token, produkPayload{Name: "Air", Price: 3000})

	var listed dataEnvelope[[]string]
	status := apiCall(t, http.MethodGet, baseURL+"/api/produk/kategori", token, nil, &listed)

	if status != http.StatusOK {
		t.Fatalf("list Kategori: got status %d, want %d", status, http.StatusOK)
	}

	want := []string{"Makanan", "Minuman"}
	if len(listed.Data) != len(want) {
		t.Fatalf("list Kategori: got %v, want %v", listed.Data, want)
	}
	for i := range want {
		if listed.Data[i] != want[i] {
			t.Errorf("list Kategori[%d]: got %q, want %q", i, listed.Data[i], want[i])
		}
	}
}

func TestProdukRoutesReportAMissingOrUnusableId(t *testing.T) {
	baseURL, token := newAdminToken(t)

	tests := []struct {
		name          string
		method        string
		path          string
		body          any
		wantStatus    int
		wantErrorCode string
	}{
		{
			name: "a Produk that is not there", method: http.MethodPut, path: "/api/produk/404",
			body: produkPayload{Name: "Kopi", Price: 1000}, wantStatus: http.StatusNotFound, wantErrorCode: "product_not_found",
		},
		{
			name: "a Produk that is not there", method: http.MethodPatch, path: "/api/produk/404",
			body: activePayload{Active: false}, wantStatus: http.StatusNotFound, wantErrorCode: "product_not_found",
		},
		{
			name: "an id that is not a number", method: http.MethodDelete, path: "/api/produk/abc",
			wantStatus: http.StatusBadRequest, wantErrorCode: "invalid_input",
		},
		{
			name: "a PATCH without active", method: http.MethodPatch, path: "/api/produk/1",
			body: map[string]any{}, wantStatus: http.StatusBadRequest, wantErrorCode: "invalid_input",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, test.method, baseURL+test.path, token, test.body, &failure)

			if status != test.wantStatus {
				t.Fatalf("got status %d, want %d", status, test.wantStatus)
			}
			if failure.Error != test.wantErrorCode {
				t.Errorf("got code %q, want %q", failure.Error, test.wantErrorCode)
			}
		})
	}
}

// createProduk adds a Produk as an Admin and returns it, failing the test when
// the Admin could not create it.
func createProduk(t *testing.T, baseURL, token string, payload produkPayload) productPayload {
	t.Helper()

	var created dataEnvelope[productEnvelope]
	status := apiCall(t, http.MethodPost, baseURL+"/api/produk", token, payload, &created)

	if status != http.StatusCreated {
		t.Fatalf("create Produk %q: got status %d, want %d", payload.Name, status, http.StatusCreated)
	}

	return created.Data.Product
}

// listProduk answers the catalogue, with the query string passed through as the
// test wrote it.
func listProduk(t *testing.T, baseURL, token, query string) []productPayload {
	t.Helper()

	path := "/api/produk"
	if query != "" {
		path += "?" + query
	}

	var listed dataEnvelope[[]productPayload]
	status := apiCall(t, http.MethodGet, baseURL+path, token, nil, &listed)

	if status != http.StatusOK {
		t.Fatalf("list Produk %q: got status %d, want %d", query, status, http.StatusOK)
	}

	return listed.Data
}

// markSold plants the state only a finished Penjualan will set (#6): the API has
// no route that sells a Produk yet, and the refusal it must answer with is worth
// exercising at the HTTP seam today rather than when the till lands.
func markSold(t *testing.T, dbPath string, id int64) {
	t.Helper()

	ctx := context.Background()

	db, err := sqlite.Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("open database to mark a Produk sold: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, `UPDATE produk SET sold = 1 WHERE id = ?`, id); err != nil {
		t.Fatalf("mark Produk %d sold: %v", id, err)
	}
}
