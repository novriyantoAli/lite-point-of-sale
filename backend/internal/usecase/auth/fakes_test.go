package auth

import (
	"context"
	"strings"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
)

// fakeUsers is an in-memory UserRepository. The use cases are tested against
// this instead of SQLite: the port is the seam (ADR-0007).
type fakeUsers struct {
	users  map[int64]domainauth.User
	nextID int64
	// err, when set, is returned by every method — for error propagation.
	err error
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{users: map[int64]domainauth.User{}}
}

// seed inserts a Pengguna directly, bypassing the use cases, and returns its id.
func (f *fakeUsers) seed(user domainauth.User) int64 {
	f.nextID++
	user.ID = f.nextID
	if user.PasswordHash == "" {
		user.PasswordHash = "hash:" + user.Username
	}
	f.users[user.ID] = user

	return user.ID
}

func (f *fakeUsers) FindByUsername(_ context.Context, username string) (domainauth.User, error) {
	if f.err != nil {
		return domainauth.User{}, f.err
	}

	for _, user := range f.users {
		if user.Username == username {
			return user, nil
		}
	}

	return domainauth.User{}, domainauth.ErrUserNotFound
}

func (f *fakeUsers) FindByID(_ context.Context, id int64) (domainauth.User, error) {
	if f.err != nil {
		return domainauth.User{}, f.err
	}

	user, ok := f.users[id]
	if !ok {
		return domainauth.User{}, domainauth.ErrUserNotFound
	}

	return user, nil
}

func (f *fakeUsers) Create(_ context.Context, user domainauth.User) (domainauth.User, error) {
	if f.err != nil {
		return domainauth.User{}, f.err
	}

	if _, err := f.FindByUsername(context.Background(), user.Username); err == nil {
		return domainauth.User{}, domainauth.ErrUsernameTaken
	}

	f.nextID++
	user.ID = f.nextID
	f.users[user.ID] = user

	return user, nil
}

func (f *fakeUsers) SetActive(_ context.Context, id int64, active bool) error {
	if f.err != nil {
		return f.err
	}

	user, ok := f.users[id]
	if !ok {
		return domainauth.ErrUserNotFound
	}

	user.Active = active
	f.users[id] = user

	return nil
}

func (f *fakeUsers) List(context.Context) ([]domainauth.User, error) {
	if f.err != nil {
		return nil, f.err
	}

	users := make([]domainauth.User, 0, len(f.users))
	for id := int64(1); id <= f.nextID; id++ {
		if user, ok := f.users[id]; ok {
			users = append(users, user)
		}
	}

	return users, nil
}

func (f *fakeUsers) HasAdmin(context.Context) (bool, error) {
	if f.err != nil {
		return false, f.err
	}

	for _, user := range f.users {
		if user.Role == domainauth.RoleAdmin {
			return true, nil
		}
	}

	return false, nil
}

// fakeHasher stands in for bcrypt: a hash is the password with a prefix, so a
// test can read what the use case stored without paying the bcrypt cost.
type fakeHasher struct{ err error }

func (f fakeHasher) Hash(password string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return "hash:" + password, nil
}

func (f fakeHasher) Verify(password, hash string) error {
	if hash != "hash:"+password {
		return domainauth.ErrInvalidCredentials
	}
	return nil
}

// fakeTokens stands in for the HMAC adapter: a token is the username with a
// prefix, so a test can forge one for a user it never logged in.
type fakeTokens struct{ err error }

func (f fakeTokens) Issue(user domainauth.PublicUser) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return "token:" + user.Username, nil
}

func (f fakeTokens) Verify(token string) (domainauth.Claims, error) {
	if f.err != nil {
		return domainauth.Claims{}, f.err
	}

	username, ok := strings.CutPrefix(token, "token:")
	if !ok || username == "" {
		return domainauth.Claims{}, domainauth.ErrInvalidToken
	}

	return domainauth.Claims{Username: username, Role: domainauth.RoleKasir}, nil
}

// newTestService wires a Service over fresh fakes.
func newTestService(users *fakeUsers, hasher domainauth.PasswordHasher, tokens domainauth.TokenManager) *Service {
	if hasher == nil {
		hasher = fakeHasher{}
	}
	if tokens == nil {
		tokens = fakeTokens{}
	}
	return NewService(users, hasher, tokens)
}
