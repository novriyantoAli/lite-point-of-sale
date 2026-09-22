package auth

import (
	"context"
	"errors"
	"testing"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
)

func TestCreateUserValidatesInput(t *testing.T) {
	tests := []struct {
		name  string
		input CreateUserInput
	}{
		{
			name:  "empty username",
			input: CreateUserInput{Username: "", Password: "rahasia123", Role: domainauth.RoleKasir},
		},
		{
			name:  "username of spaces only",
			input: CreateUserInput{Username: "   ", Password: "rahasia123", Role: domainauth.RoleKasir},
		},
		{
			name:  "password shorter than the minimum",
			input: CreateUserInput{Username: "kasir2", Password: "pendek", Role: domainauth.RoleKasir},
		},
		{
			name:  "unknown role",
			input: CreateUserInput{Username: "kasir2", Password: "rahasia123", Role: domainauth.Role("pemilik")},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			users := newFakeUsers()

			_, err := newTestService(users, nil, nil).CreateUser(context.Background(), test.input)

			if !errors.Is(err, domainauth.ErrInvalidInput) {
				t.Fatalf("create: got error %v, want %v", err, domainauth.ErrInvalidInput)
			}

			// The HTTP adapter reads the message straight off this error, so it
			// must carry one.
			var inputErr InputError
			if !errors.As(err, &inputErr) || inputErr.Message == "" {
				t.Errorf("create: got %v, want an InputError with a message", err)
			}

			if stored, _ := users.List(context.Background()); len(stored) != 0 {
				t.Errorf("create: stored %d Pengguna, want none", len(stored))
			}
		})
	}
}

func TestCreateUserStoresHashedPasswordAndNormalizedUsername(t *testing.T) {
	users := newFakeUsers()

	created, err := newTestService(users, nil, nil).CreateUser(context.Background(), CreateUserInput{
		Username: "  Kasir Baru ",
		Password: "rahasia123",
		Role:     domainauth.RoleKasir,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if created.Username != "kasir baru" {
		t.Errorf("username: got %q, want %q", created.Username, "kasir baru")
	}
	if !created.Active {
		t.Error("active: got false, want a new Pengguna to be active")
	}
	if created.Role != domainauth.RoleKasir {
		t.Errorf("role: got %q, want %q", created.Role, domainauth.RoleKasir)
	}

	stored, err := users.FindByUsername(context.Background(), "kasir baru")
	if err != nil {
		t.Fatalf("find created Pengguna: %v", err)
	}
	if stored.PasswordHash == "rahasia123" {
		t.Error("password: stored in plaintext, want the hasher output")
	}
	if want := "hash:rahasia123"; stored.PasswordHash != want {
		t.Errorf("password hash: got %q, want %q", stored.PasswordHash, want)
	}
}

func TestCreateUserRejectsADuplicateUsername(t *testing.T) {
	users := newFakeUsers()
	users.seed(domainauth.User{Username: "kasir1", Role: domainauth.RoleKasir, Active: true})

	// Same Pengguna, typed differently: normalization must still catch it.
	_, err := newTestService(users, nil, nil).CreateUser(context.Background(), CreateUserInput{
		Username: "KASIR1",
		Password: "rahasia123",
		Role:     domainauth.RoleKasir,
	})

	if !errors.Is(err, domainauth.ErrUsernameTaken) {
		t.Fatalf("create: got error %v, want %v", err, domainauth.ErrUsernameTaken)
	}
}

func TestListUsersNeverReturnsPasswordHashes(t *testing.T) {
	users := newFakeUsers()
	users.seed(domainauth.User{Username: "admin1", Role: domainauth.RoleAdmin, Active: true})
	users.seed(domainauth.User{Username: "kasir1", Role: domainauth.RoleKasir, Active: false})

	listed, err := newTestService(users, nil, nil).ListUsers(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(listed) != 2 {
		t.Fatalf("list: got %d Pengguna, want 2", len(listed))
	}
	if want := "admin1"; listed[0].Username != want {
		t.Errorf("first Pengguna: got %q, want %q", listed[0].Username, want)
	}
	if listed[1].Active {
		t.Error("second Pengguna: got active, want the stored nonaktif state")
	}
}

func TestSetUserActive(t *testing.T) {
	tests := []struct {
		name       string
		active     bool
		targetIsID bool
		wantErr    error
	}{
		{name: "deactivating another Pengguna", targetIsID: false},
		{name: "reactivating another Pengguna", active: true},
		{name: "an Admin cannot deactivate their own account", targetIsID: true, wantErr: domainauth.ErrCannotDeactivateSelf},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			users := newFakeUsers()
			actorID := users.seed(domainauth.User{Username: "admin1", Role: domainauth.RoleAdmin, Active: true})
			// The target starts on the opposite side of the change, so either
			// direction is actually exercised.
			otherID := users.seed(domainauth.User{Username: "kasir1", Role: domainauth.RoleKasir, Active: !test.active})

			targetID := otherID
			if test.targetIsID {
				targetID = actorID
			}

			updated, err := newTestService(users, nil, nil).
				SetUserActive(context.Background(), targetID, actorID, test.active)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("set active: got error %v, want %v", err, test.wantErr)
			}
			if test.wantErr != nil {
				return
			}

			if updated.Active != test.active {
				t.Errorf("returned Pengguna: got active %v, want %v", updated.Active, test.active)
			}

			target, err := users.FindByID(context.Background(), targetID)
			if err != nil {
				t.Fatalf("find target: %v", err)
			}
			if target.Active != test.active {
				t.Errorf("stored active: got %v, want %v", target.Active, test.active)
			}
		})
	}
}

func TestSetUserActiveRejectsAnUnknownPengguna(t *testing.T) {
	users := newFakeUsers()
	actorID := users.seed(domainauth.User{Username: "admin1", Role: domainauth.RoleAdmin, Active: true})

	_, err := newTestService(users, nil, nil).SetUserActive(context.Background(), actorID+99, actorID, false)

	if !errors.Is(err, domainauth.ErrUserNotFound) {
		t.Fatalf("set active: got error %v, want %v", err, domainauth.ErrUserNotFound)
	}
}

func TestEnsureAdminSeedsOnce(t *testing.T) {
	users := newFakeUsers()
	service := newTestService(users, nil, nil)

	created, err := service.EnsureAdmin(context.Background(), "admin", "rahasia123")
	if err != nil {
		t.Fatalf("first start: %v", err)
	}
	if !created {
		t.Error("first start: got no Admin created, want one")
	}

	seeded, err := users.FindByUsername(context.Background(), "admin")
	if err != nil {
		t.Fatalf("find seeded Admin: %v", err)
	}
	if seeded.Role != domainauth.RoleAdmin {
		t.Errorf("seeded role: got %q, want %q", seeded.Role, domainauth.RoleAdmin)
	}

	created, err = service.EnsureAdmin(context.Background(), "admin", "rahasia123")
	if err != nil {
		t.Fatalf("second start: %v", err)
	}
	if created {
		t.Error("second start: seeded a second Admin, want none")
	}

	if stored, _ := users.List(context.Background()); len(stored) != 1 {
		t.Errorf("stored Pengguna: got %d, want 1", len(stored))
	}
}

func TestEnsureAdminFailsOnAnUnusableInitialPassword(t *testing.T) {
	users := newFakeUsers()
	// An unusable initial password is a configuration mistake; startup must
	// fail loudly instead of running a store nobody can log in to.
	_, err := newTestService(users, nil, nil).EnsureAdmin(context.Background(), "admin", "pendek")

	if !errors.Is(err, domainauth.ErrInvalidInput) {
		t.Fatalf("seed: got error %v, want %v", err, domainauth.ErrInvalidInput)
	}
}
