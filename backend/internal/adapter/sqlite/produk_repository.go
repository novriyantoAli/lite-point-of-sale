package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	domainproduk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/produk"
)

// ProductRepository stores Produk in SQLite. It satisfies
// domain/produk.ProductRepository, so no SQL leaves this package (ADR-0004).
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository returns a repository backed by the given database handle.
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

const productColumns = `id, name, code, price, category, stock, active, sold`

// Create satisfies domain/produk.ProductRepository. The id is read back from
// the insert, so callers get the stored Produk rather than the one they sent.
func (r *ProductRepository) Create(ctx context.Context, product domainproduk.Product) (domainproduk.Product, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO produk (name, code, price, category, stock, active, sold)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		product.Name, product.Code, product.Price, product.Category, product.Stock, product.Active, product.Sold)
	if err != nil {
		// The use case checks for an existing Kode first; this catches the race
		// between that check and this insert.
		if isUniqueViolation(err) {
			return domainproduk.Product{}, domainproduk.ErrCodeTaken
		}
		return domainproduk.Product{}, fmt.Errorf("insert Produk %q: %w", product.Name, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domainproduk.Product{}, fmt.Errorf("read id of Produk %q: %w", product.Name, err)
	}

	product.ID = id

	return product, nil
}

// Update satisfies domain/produk.ProductRepository. Only the editable fields
// are written: active and sold have their own paths, so an edit can neither
// reactivate a Nonaktif Produk nor erase that it has sold.
func (r *ProductRepository) Update(ctx context.Context, product domainproduk.Product) (domainproduk.Product, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE produk SET name = ?, code = ?, price = ?, category = ?, stock = ? WHERE id = ?`,
		product.Name, product.Code, product.Price, product.Category, product.Stock, product.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return domainproduk.Product{}, domainproduk.ErrCodeTaken
		}
		return domainproduk.Product{}, fmt.Errorf("update Produk %d: %w", product.ID, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return domainproduk.Product{}, fmt.Errorf("count updated Produk %d: %w", product.ID, err)
	}
	if affected == 0 {
		return domainproduk.Product{}, domainproduk.ErrProductNotFound
	}

	return product, nil
}

// FindByID satisfies domain/produk.ProductRepository.
func (r *ProductRepository) FindByID(ctx context.Context, id int64) (domainproduk.Product, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+productColumns+` FROM produk WHERE id = ?`, id)

	return scanProduct(row)
}

// FindByCode satisfies domain/produk.ProductRepository. A Produk without a
// Kode is not stored as an empty string but as NULL, so a search for one never
// matches them.
func (r *ProductRepository) FindByCode(ctx context.Context, code string) (domainproduk.Product, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+productColumns+` FROM produk WHERE code = ?`, code)

	return scanProduct(row)
}

// List satisfies domain/produk.ProductRepository, ordered by name so the API
// answers in a stable order.
func (r *ProductRepository) List(ctx context.Context, filter domainproduk.Filter) ([]domainproduk.Product, error) {
	query := `SELECT ` + productColumns + ` FROM produk`
	conditions, args := productConditions(filter)
	if len(conditions) > 0 {
		query += ` WHERE ` + strings.Join(conditions, ` AND `)
	}
	query += ` ORDER BY name, id`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list Produk: %w", err)
	}
	defer rows.Close()

	products := []domainproduk.Product{}
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read Produk rows: %w", err)
	}

	return products, nil
}

// Categories satisfies domain/produk.ProductRepository: the distinct Kategori
// in use, sorted. A Produk without one contributes nothing.
func (r *ProductRepository) Categories(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT category FROM produk WHERE category IS NOT NULL ORDER BY category`)
	if err != nil {
		return nil, fmt.Errorf("list Kategori: %w", err)
	}
	defer rows.Close()

	categories := []string{}
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, fmt.Errorf("read Kategori: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read Kategori rows: %w", err)
	}

	return categories, nil
}

// SetActive satisfies domain/produk.ProductRepository. A Produk that is not
// there is reported as such instead of passing silently.
func (r *ProductRepository) SetActive(ctx context.Context, id int64, active bool) error {
	result, err := r.db.ExecContext(ctx, `UPDATE produk SET active = ? WHERE id = ?`, active, id)
	if err != nil {
		return fmt.Errorf("set active of Produk %d: %w", id, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count updated Produk %d: %w", id, err)
	}
	if affected == 0 {
		return domainproduk.ErrProductNotFound
	}

	return nil
}

// Delete satisfies domain/produk.ProductRepository. Whether a Produk may be
// deleted at all is a domain rule the use case decides; this only removes the
// row, and reports a Produk that is not there.
func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM produk WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete Produk %d: %w", id, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count deleted Produk %d: %w", id, err)
	}
	if affected == 0 {
		return domainproduk.ErrProductNotFound
	}

	return nil
}

// productConditions builds the WHERE clauses of a catalogue listing. It is a
// list of conditions rather than one query per filter combination, so adding a
// filter never multiplies the paths through the code.
func productConditions(filter domainproduk.Filter) ([]string, []any) {
	conditions := []string{}
	args := []any{}

	if filter.Name != "" {
		conditions = append(conditions, `name LIKE ? ESCAPE '\'`)
		args = append(args, likePattern(filter.Name))
	}
	if filter.Code != "" {
		conditions = append(conditions, `code LIKE ? ESCAPE '\'`)
		args = append(args, likePattern(filter.Code))
	}
	if filter.Category != "" {
		conditions = append(conditions, `category = ?`)
		args = append(args, filter.Category)
	}
	if filter.Active != nil {
		conditions = append(conditions, `active = ?`)
		args = append(args, *filter.Active)
	}

	return conditions, args
}

// likePattern wraps text in a contains-match, escaping the LIKE wildcards so a
// Kode that happens to contain `%` or `_` searches for itself instead of
// matching every Produk.
func likePattern(value string) string {
	escaper := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

	return `%` + escaper.Replace(value) + `%`
}

func scanProduct(source row) (domainproduk.Product, error) {
	var (
		product  domainproduk.Product
		code     sql.NullString
		category sql.NullString
	)

	err := source.Scan(
		&product.ID, &product.Name, &code, &product.Price, &category,
		&product.Stock, &product.Active, &product.Sold,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domainproduk.Product{}, domainproduk.ErrProductNotFound
	}
	if err != nil {
		return domainproduk.Product{}, fmt.Errorf("read Produk: %w", err)
	}

	if code.Valid {
		product.Code = &code.String
	}
	if category.Valid {
		product.Category = &category.String
	}

	return product, nil
}
