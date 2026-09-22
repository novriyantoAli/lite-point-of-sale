package produk

import (
	"context"
	"errors"
	"strings"

	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// Create adds a Produk to the catalogue. Only an Admin reaches this use case;
// the HTTP adapter is what enforces that role.
func (s *Service) Create(ctx context.Context, input ProductInput) (domainproduk.Product, error) {
	product, err := validate(input)
	if err != nil {
		return domainproduk.Product{}, err
	}

	if err := s.ensureCodeFree(ctx, product.Code, 0); err != nil {
		return domainproduk.Product{}, err
	}

	created, err := s.products.Create(ctx, product)
	if err != nil {
		return domainproduk.Product{}, err
	}

	return created, nil
}

// Update replaces the editable fields of a Produk. A Nonaktif Produk stays
// Nonaktif and one that has sold stays sold: neither is an edit-form decision.
func (s *Service) Update(ctx context.Context, id int64, input ProductInput) (domainproduk.Product, error) {
	existing, err := s.products.FindByID(ctx, id)
	if err != nil {
		return domainproduk.Product{}, err
	}

	product, err := validate(input)
	if err != nil {
		return domainproduk.Product{}, err
	}

	product.ID = id
	product.Active = existing.Active
	product.Sold = existing.Sold

	if err := s.ensureCodeFree(ctx, product.Code, id); err != nil {
		return domainproduk.Product{}, err
	}

	return s.products.Update(ctx, product)
}

// List answers the catalogue, narrowed by whatever the filter asks for. An
// Admin's catalogue screen passes no Active filter and sees Nonaktif Produk
// too; the kasir lookup of #6 asks for Active=true only.
func (s *Service) List(ctx context.Context, filter domainproduk.Filter) ([]domainproduk.Product, error) {
	return s.products.List(ctx, filter)
}

// SetActive activates or deactivates a Produk and returns its new state, so the
// caller answers with the updated Produk instead of reading it back — a second
// round trip that could also see a different state.
func (s *Service) SetActive(ctx context.Context, id int64, active bool) (domainproduk.Product, error) {
	product, err := s.products.FindByID(ctx, id)
	if err != nil {
		return domainproduk.Product{}, err
	}

	if err := s.products.SetActive(ctx, id, active); err != nil {
		return domainproduk.Product{}, err
	}

	product.Active = active

	return product, nil
}

// Delete removes a Produk that has never been sold. A Produk that was part of a
// finished Penjualan is refused: the sale copies the name and price it recorded,
// but the history it belongs to is not something a catalogue edit may erase
// (CONTEXT.md, Nonaktif).
func (s *Service) Delete(ctx context.Context, id int64) error {
	product, err := s.products.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if product.Sold {
		return domainproduk.ErrProductHasSales
	}

	return s.products.Delete(ctx, id)
}

// Categories answers every Kategori in use, so the UI can offer the ones that
// exist rather than a hand-maintained list.
func (s *Service) Categories(ctx context.Context) ([]string, error) {
	return s.products.Categories(ctx)
}

// validate turns form input into a Product, rejecting what the catalogue cannot
// store. Blank text is not an error for Kode or Kategori — it is how the form
// says "none" — but it is for a name, which every Produk needs.
func validate(input ProductInput) (domainproduk.Product, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domainproduk.Product{}, InputError{Message: "Nama Produk wajib diisi."}
	}
	if input.Price < 0 {
		return domainproduk.Product{}, InputError{Message: "Harga tidak boleh negatif."}
	}
	if input.Stock < 0 {
		return domainproduk.Product{}, InputError{Message: "Stok tidak boleh negatif."}
	}

	return domainproduk.Product{
		Name:     name,
		Code:     optionalText(input.Code),
		Price:    input.Price,
		Category: optionalText(input.Category),
		Stock:    input.Stock,
		Active:   activeOrTrue(input.Active),
	}, nil
}

// activeOrTrue reads the Status a create asked for. An absent one means Aktif:
// adding a Produk is how an Admin puts something on sale, so the default has to
// be the status the kasir lookup of #6 shows.
func activeOrTrue(active *bool) bool {
	if active == nil {
		return true
	}

	return *active
}

// optionalText turns blank or whitespace-only text into nil, so "no Kode" is
// stored as NULL and many Produk can go without one.
func optionalText(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

// ensureCodeFree rejects a Kode another Produk already has. selfID is the
// Produk being edited (0 when creating), so keeping your own Kode is not a
// conflict with yourself.
//
// Ask first so the everyday duplicate answers with a clear conflict. The
// repository still maps the unique constraint on its own, because two Admins
// can race the same Kode past this check.
func (s *Service) ensureCodeFree(ctx context.Context, code *string, selfID int64) error {
	if code == nil {
		return nil
	}

	existing, err := s.products.FindByCode(ctx, *code)
	if errors.Is(err, domainproduk.ErrProductNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	if existing.ID != selfID {
		return domainproduk.ErrCodeTaken
	}

	return nil
}
