package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
	usecaseproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/produk"
)

// ProductService is the Produk use cases the HTTP adapter depends on. Declaring
// it here, next to the handlers that use it, keeps this package testable with a
// fake service and keeps the dependency pointing inward (ADR-0004).
type ProductService interface {
	Create(ctx context.Context, input usecaseproduk.ProductInput) (domainproduk.Product, error)
	Update(ctx context.Context, id int64, input usecaseproduk.ProductInput) (domainproduk.Product, error)
	List(ctx context.Context, filter domainproduk.Filter) ([]domainproduk.Product, error)
	AddStock(ctx context.Context, id int64, quantity int64) (domainproduk.Product, error)
	LowStock(ctx context.Context) ([]domainproduk.Product, error)
	SetActive(ctx context.Context, id int64, active bool) (domainproduk.Product, error)
	Delete(ctx context.Context, id int64) error
	Categories(ctx context.Context) ([]string, error)
}

// productRequest is what an Admin posts to add or change a Produk. Code and
// Category are pointers so an absent or `null` value means "no Kode"/"no
// Kategori" instead of being confused with the empty string.
type productRequest struct {
	Name     string  `json:"name"`
	Code     *string `json:"code"`
	Price    int64   `json:"price"`
	Category *string `json:"category"`
	Stock    int64   `json:"stock"`
	// Active is a pointer so an edit that omits it is not read as "deactivate".
	// Only Create reads it, and an absent value means Aktif.
	Active *bool `json:"active"`
}

func (r productRequest) input() usecaseproduk.ProductInput {
	return usecaseproduk.ProductInput{
		Name:     r.Name,
		Code:     derefText(r.Code),
		Price:    r.Price,
		Category: derefText(r.Category),
		Stock:    r.Stock,
		Active:   r.Active,
	}
}

// setProductActiveRequest is the body of a deactivation or a reactivation. The
// field is a pointer so that a body which forgot it is rejected instead of
// being read as "deactivate".
type setProductActiveRequest struct {
	Active *bool `json:"active"`
}

// addStockRequest is the body of a restock: how many units arrived. It is not
// a pointer, unlike the other optional fields: a body that forgot the quantity
// arrives as 0, and 0 is not a restock either way — so the use case's own rule
// answers it instead of this struct keeping a second copy of that rule.
type addStockRequest struct {
	Quantity int64 `json:"quantity"`
}

// lowStockResponse is what the restock list answers. The threshold travels with
// the Produk because it is the rule that decided the list: the UI shows the
// Admin "Stok menipis ≤ 5", and a threshold the frontend held separately would
// be a second copy of the rule that could drift from this one.
type lowStockResponse struct {
	Threshold int64             `json:"threshold"`
	Products  []productResponse `json:"products"`
}

// productResponse is the JSON view of a Produk. `sold` is answered so the UI can
// tell a Produk that may be deleted from one that may only be deactivated.
type productResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Code     *string `json:"code"`
	Price    int64   `json:"price"`
	Category *string `json:"category"`
	Stock    int64   `json:"stock"`
	Active   bool    `json:"active"`
	Sold     bool    `json:"sold"`
}

func newProductResponse(product domainproduk.Product) productResponse {
	return productResponse{
		ID:       product.ID,
		Name:     product.Name,
		Code:     product.Code,
		Price:    product.Price,
		Category: product.Category,
		Stock:    product.Stock,
		Active:   product.Active,
		Sold:     product.Sold,
	}
}

func newProductResponses(products []domainproduk.Product) []productResponse {
	responses := make([]productResponse, 0, len(products))
	for _, product := range products {
		responses = append(responses, newProductResponse(product))
	}

	return responses
}

// productEnvelope wraps a Produk, the same way userEnvelope wraps a Pengguna.
type productEnvelope struct {
	Product productResponse `json:"product"`
}

// createProductHandler adds a Produk. It is Admin-only: the router puts the
// role guard in front of it.
func createProductHandler(service ProductService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, ok := decodeProductRequest(w, r, logger)
		if !ok {
			return
		}

		created, err := service.Create(r.Context(), request.input())
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusCreated, dataResponse{Data: productEnvelope{
			Product: newProductResponse(created),
		}}, logger)
	}
}

// updateProductHandler replaces the editable fields of a Produk. Active and Sold
// are not part of the body: the API is what decides those.
func updateProductHandler(service ProductService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := productID(w, r, logger)
		if !ok {
			return
		}

		request, ok := decodeProductRequest(w, r, logger)
		if !ok {
			return
		}

		updated, err := service.Update(r.Context(), id, request.input())
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: productEnvelope{
			Product: newProductResponse(updated),
		}}, logger)
	}
}

// listProductHandler answers the catalogue, narrowed by the query parameters an
// Admin may pass. No `active` parameter means "every Produk", which is what the
// catalogue screen wants; `active=true` is the kasir lookup of #6.
func listProductHandler(service ProductService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, ok := productFilter(w, r, logger)
		if !ok {
			return
		}

		products, err := service.List(r.Context(), filter)
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: newProductResponses(products)}, logger)
	}
}

// listCategoryHandler answers the Kategori in use, so the filter offers the ones
// that exist instead of a list the frontend would have to maintain.
func listCategoryHandler(service ProductService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categories, err := service.Categories(r.Context())
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: categories}, logger)
	}
}

// setProductActiveHandler activates or deactivates a Produk: a Produk that is
// no longer sold disappears from the kasir lookup while its history stays
// readable (CONTEXT.md, Nonaktif).
func setProductActiveHandler(service ProductService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := productID(w, r, logger)
		if !ok {
			return
		}

		var request setProductActiveRequest
		if !decodeJSON(w, r, &request, logger) {
			return
		}
		if request.Active == nil {
			writeInvalidInput(w, "Field active wajib diisi.", logger)
			return
		}

		updated, err := service.SetActive(r.Context(), id, *request.Active)
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: productEnvelope{
			Product: newProductResponse(updated),
		}}, logger)
	}
}

// addStockHandler records a restock: the units an Admin received for one Produk
// (CONTEXT.md, Stok). It is Admin-only, like the rest of the catalogue — the
// Kasir's Stok only ever moves through a Penjualan (#6).
func addStockHandler(service ProductService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := productID(w, r, logger)
		if !ok {
			return
		}

		var request addStockRequest
		if !decodeJSON(w, r, &request, logger) {
			return
		}

		updated, err := service.AddStock(r.Context(), id, request.Quantity)
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: productEnvelope{
			Product: newProductResponse(updated),
		}}, logger)
	}
}

// listLowStockHandler answers the Produk an Admin has to restock: the ones whose
// Stok is menipis or habis, thinnest first. What counts as menipis is the
// domain's rule, so the threshold is answered alongside the list rather than
// asked for.
func listLowStockHandler(service ProductService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		products, err := service.LowStock(r.Context())
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: lowStockResponse{
			Threshold: domainproduk.LowStockThreshold,
			Products:  newProductResponses(products),
		}}, logger)
	}
}

// deleteProductHandler removes a Produk that has never sold. Whether that is
// allowed at all is the use case's call; a Produk that sold answers 409.
func deleteProductHandler(service ProductService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := productID(w, r, logger)
		if !ok {
			return
		}

		if err := service.Delete(r.Context(), id); err != nil {
			writeError(w, err, logger)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// decodeProductRequest reads the body of a create or an update.
func decodeProductRequest(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (productRequest, bool) {
	var request productRequest
	if !decodeJSON(w, r, &request, logger) {
		return productRequest{}, false
	}

	return request, true
}

// productFilter reads the catalogue filters off the query string. An `active`
// that is neither true nor false is invalid input rather than a silent
// no-filter, because a lookup that quietly widens is worse than one that fails.
func productFilter(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (domainproduk.Filter, bool) {
	query := r.URL.Query()
	filter := domainproduk.Filter{
		Name:     strings.TrimSpace(query.Get("name")),
		Code:     strings.TrimSpace(query.Get("code")),
		Category: strings.TrimSpace(query.Get("category")),
	}

	if raw := query.Get("active"); raw != "" {
		active, err := strconv.ParseBool(raw)
		if err != nil {
			writeInvalidInput(w, "Filter active harus true atau false.", logger)
			return domainproduk.Filter{}, false
		}
		filter.Active = &active
	}

	return filter, true
}

// productID reads the {id} of the route. A path that is not an id at all is
// invalid input, not a missing Produk.
func productID(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeInvalidInput(w, "Id Produk tidak valid.", logger)
		return 0, false
	}

	return id, true
}

// derefText turns an optional JSON string into the plain string the use case
// takes; an absent or null value is the empty string, which the use case reads
// as "none".
func derefText(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
