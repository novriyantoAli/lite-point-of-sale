package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
)

// UserRepository stores Pengguna in SQLite. It satisfies
// domain/auth.UserRepository, so no SQL leaves this package (ADR-0004).
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository returns a repository backed by the given database handle.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const userColumns = `id, username, password_hash, role, active`

// FindByUsername satisfies domain/auth.UserRepository. Lookups are
// case-sensitive by design: the use case stores usernames lowercased, so the
// single comparison is enough (see usecase/auth.normalizeUsername).
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (domainauth.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM pengguna WHERE username = ?`, username)

	return scanUser(row)
}

// FindByID satisfies domain/auth.UserRepository.
func (r *UserRepository) FindByID(ctx context.Context, id int64) (domainauth.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM pengguna WHERE id = ?`, id)

	return scanUser(row)
}

// Create satisfies domain/auth.UserRepository. The id is read back from the
// insert, so callers get the stored Pengguna rather than the one they sent.
func (r *UserRepository) Create(ctx context.Context, user domainauth.User) (domainauth.User, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO pengguna (username, password_hash, role, active) VALUES (?, ?, ?, ?)`,
		user.Username, user.PasswordHash, string(user.Role), user.Active)
	if err != nil {
		// The use case checks for an existing username first; this catches the
		// race between that check and this insert.
		if isUniqueViolation(err) {
			return domainauth.User{}, domainauth.ErrUsernameTaken
		}
		return domainauth.User{}, fmt.Errorf("insert Pengguna %q: %w", user.Username, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domainauth.User{}, fmt.Errorf("read id of Pengguna %q: %w", user.Username, err)
	}

	user.ID = id

	return user, nil
}

// SetActive satisfies domain/auth.UserRepository. A Pengguna that is not there
// is reported as such instead of passing silently.
func (r *UserRepository) SetActive(ctx context.Context, id int64, active bool) error {
	result, err := r.db.ExecContext(ctx, `UPDATE pengguna SET active = ? WHERE id = ?`, active, id)
	if err != nil {
		return fmt.Errorf("set active of Pengguna %d: %w", id, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count updated Pengguna %d: %w", id, err)
	}
	if affected == 0 {
		return domainauth.ErrUserNotFound
	}

	return nil
}

// List satisfies domain/auth.UserRepository, ordered by username so the API
// answers in a stable order.
func (r *UserRepository) List(ctx context.Context) ([]domainauth.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+userColumns+` FROM pengguna ORDER BY username`)
	if err != nil {
		return nil, fmt.Errorf("list Pengguna: %w", err)
	}
	defer rows.Close()

	users := []domainauth.User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read Pengguna rows: %w", err)
	}

	return users, nil
}

// HasAdmin satisfies domain/auth.UserRepository: it reports whether any Admin
// exists, which is what startup needs before seeding the first one.
//
// Any Admin row means a usable Admin: an Admin cannot deactivate their own
// account, so the API never leaves a store with only nonaktif Admins.
func (r *UserRepository) HasAdmin(ctx context.Context) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM pengguna WHERE role = ?)`, string(domainauth.RoleAdmin)).
		Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("look for an Admin Pengguna: %w", err)
	}

	return exists, nil
}

// row is what both *sql.Row and *sql.Rows can do here.
type row interface {
	Scan(dest ...any) error
}

func scanUser(source row) (domainauth.User, error) {
	var (
		user domainauth.User
		role string
	)

	err := source.Scan(&user.ID, &user.Username, &user.PasswordHash, &role, &user.Active)
	if errors.Is(err, sql.ErrNoRows) {
		return domainauth.User{}, domainauth.ErrUserNotFound
	}
	if err != nil {
		return domainauth.User{}, fmt.Errorf("read Pengguna: %w", err)
	}

	user.Role = domainauth.Role(role)

	return user, nil
}

// isUniqueViolation reports whether err is SQLite's UNIQUE constraint failure.
// The driver exposes no typed error for it, so the constraint name in the
// message is what is left to match on (0002_pengguna.sql: username UNIQUE).
func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
