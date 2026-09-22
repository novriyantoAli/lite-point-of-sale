package auth

import (
	"context"
	"errors"
	"testing"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		name     string
		seed     *domainauth.User
		username string
		password string
		wantErr  error
	}{
		{
			name:     "correct password issues a token for the Pengguna",
			seed:     &domainauth.User{Username: "kasir1", Role: domainauth.RoleKasir, Active: true},
			username: "kasir1",
			password: "kasir1",
		},
		{
			name:     "username is matched regardless of case and padding",
			seed:     &domainauth.User{Username: "kasir1", Role: domainauth.RoleKasir, Active: true},
			username: "  Kasir1 ",
			password: "kasir1",
		},
		{
			name:     "wrong password",
			seed:     &domainauth.User{Username: "kasir1", Role: domainauth.RoleKasir, Active: true},
			username: "kasir1",
			password: "salah",
			wantErr:  domainauth.ErrInvalidCredentials,
		},
		{
			name:     "unknown username answers like a wrong password",
			username: "tidak-ada",
			password: "apa-saja",
			wantErr:  domainauth.ErrInvalidCredentials,
		},
		{
			name:     "nonaktif Pengguna cannot log in",
			seed:     &domainauth.User{Username: "kasir1", Role: domainauth.RoleKasir, Active: false},
			username: "kasir1",
			password: "kasir1",
			wantErr:  domainauth.ErrInvalidCredentials,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			users := newFakeUsers()
			if test.seed != nil {
				users.seed(*test.seed)
			}

			session, err := newTestService(users, nil, nil).Login(context.Background(), test.username, test.password)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("login: got error %v, want %v", err, test.wantErr)
			}
			if test.wantErr != nil {
				return
			}

			if want := "token:" + test.seed.Username; session.Token != want {
				t.Errorf("token: got %q, want %q", session.Token, want)
			}
			if session.User.Username != test.seed.Username {
				t.Errorf("session user: got %q, want %q", session.User.Username, test.seed.Username)
			}
			if session.User.Role != test.seed.Role {
				t.Errorf("session role: got %q, want %q", session.User.Role, test.seed.Role)
			}
		})
	}
}

func TestLoginPropagatesRepositoryFailure(t *testing.T) {
	users := newFakeUsers()
	users.err = errors.New("database is closed")

	if _, err := newTestService(users, nil, nil).Login(context.Background(), "kasir1", "kasir1"); err == nil {
		t.Fatal("login: got a session, want the repository error")
	}
}

func TestLoginPropagatesTokenFailure(t *testing.T) {
	users := newFakeUsers()
	users.seed(domainauth.User{Username: "kasir1", Role: domainauth.RoleKasir, Active: true})

	_, err := newTestService(users, nil, fakeTokens{err: errors.New("no secret")}).
		Login(context.Background(), "kasir1", "kasir1")

	if err == nil || errors.Is(err, domainauth.ErrInvalidCredentials) {
		t.Fatalf("login: got %v, want the token manager error", err)
	}
}

func TestAuthenticate(t *testing.T) {
	tests := []struct {
		name    string
		seed    *domainauth.User
		token   string
		wantErr error
	}{
		{
			name:  "a token for an active Pengguna authenticates",
			seed:  &domainauth.User{Username: "kasir1", Role: domainauth.RoleKasir, Active: true},
			token: "token:kasir1",
		},
		{
			name:  "the stored Peran wins over the one the token carried",
			seed:  &domainauth.User{Username: "kasir1", Role: domainauth.RoleAdmin, Active: true},
			token: "token:kasir1",
		},
		{
			name:    "a token with no Pengguna behind it is rejected",
			token:   "token:hantu",
			wantErr: domainauth.ErrInvalidToken,
		},
		{
			name:    "a token that was not issued here is rejected",
			seed:    &domainauth.User{Username: "kasir1", Role: domainauth.RoleKasir, Active: true},
			token:   "dibuat-oleh-orang-lain",
			wantErr: domainauth.ErrInvalidToken,
		},
		{
			name:    "deactivating a Pengguna ends their session on the next call",
			seed:    &domainauth.User{Username: "kasir1", Role: domainauth.RoleKasir, Active: false},
			token:   "token:kasir1",
			wantErr: domainauth.ErrInvalidToken,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			users := newFakeUsers()
			if test.seed != nil {
				users.seed(*test.seed)
			}

			user, err := newTestService(users, nil, nil).Authenticate(context.Background(), test.token)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("authenticate: got error %v, want %v", err, test.wantErr)
			}
			if test.wantErr != nil {
				return
			}

			if user.Role != test.seed.Role {
				t.Errorf("role: got %q, want %q", user.Role, test.seed.Role)
			}
		})
	}
}
