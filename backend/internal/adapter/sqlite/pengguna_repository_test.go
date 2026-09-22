package sqlite_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	adaptersqlite "github.com/novriyantoAli/lite-point-of-sale/backend/internal/adapter/sqlite"
	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
	"github.com/novriyantoAli/lite-point-of-sale/backend/internal/infrastructure/sqlite"
)

// newUserRepository opens a migrated SQLite file and returns a repository over
// it. This is the adapter integration test of ADR-0007: the repository runs
// against the real database, no fake.
func newUserRepository(t *testing.T) (*adaptersqlite.UserRepository, context.Context) {
	t.Helper()

	ctx := context.Background()

	db, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "pos.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return adaptersqlite.NewUserRepository(db), ctx
}

func TestUserRepositoryStoresAndReadsBackAPengguna(t *testing.T) {
	repository, ctx := newUserRepository(t)

	created, err := repository.Create(ctx, domainauth.User{
		Username:     "kasir1",
		PasswordHash: "hash-rahasia",
		Role:         domainauth.RoleKasir,
		Active:       true,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Error("create: got id 0, want the id SQLite assigned")
	}

	byName, err := repository.FindByUsername(ctx, "kasir1")
	if err != nil {
		t.Fatalf("find by username: %v", err)
	}
	byID, err := repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}

	for name, user := range map[string]domainauth.User{"by username": byName, "by id": byID} {
		if user.ID != created.ID {
			t.Errorf("%s: got id %d, want %d", name, user.ID, created.ID)
		}
		if user.Username != "kasir1" {
			t.Errorf("%s: got username %q, want %q", name, user.Username, "kasir1")
		}
		if user.PasswordHash != "hash-rahasia" {
			t.Errorf("%s: got password hash %q, want %q", name, user.PasswordHash, "hash-rahasia")
		}
		if user.Role != domainauth.RoleKasir {
			t.Errorf("%s: got role %q, want %q", name, user.Role, domainauth.RoleKasir)
		}
		if !user.Active {
			t.Errorf("%s: got inactive, want active", name)
		}
	}
}

func TestUserRepositoryKeepsUsernamesUnique(t *testing.T) {
	repository, ctx := newUserRepository(t)

	user := domainauth.User{Username: "kasir1", PasswordHash: "hash", Role: domainauth.RoleKasir, Active: true}
	if _, err := repository.Create(ctx, user); err != nil {
		t.Fatalf("create first: %v", err)
	}

	// The use case checks for an existing username before inserting; this is
	// the race it cannot cover.
	if _, err := repository.Create(ctx, user); !errors.Is(err, domainauth.ErrUsernameTaken) {
		t.Fatalf("create duplicate: got error %v, want %v", err, domainauth.ErrUsernameTaken)
	}
}

func TestUserRepositoryReportsAMissingPengguna(t *testing.T) {
	repository, ctx := newUserRepository(t)

	if _, err := repository.FindByUsername(ctx, "hantu"); !errors.Is(err, domainauth.ErrUserNotFound) {
		t.Errorf("find by username: got error %v, want %v", err, domainauth.ErrUserNotFound)
	}
	if _, err := repository.FindByID(ctx, 404); !errors.Is(err, domainauth.ErrUserNotFound) {
		t.Errorf("find by id: got error %v, want %v", err, domainauth.ErrUserNotFound)
	}
	if err := repository.SetActive(ctx, 404, false); !errors.Is(err, domainauth.ErrUserNotFound) {
		t.Errorf("set active: got error %v, want %v", err, domainauth.ErrUserNotFound)
	}
}

func TestUserRepositorySetActive(t *testing.T) {
	repository, ctx := newUserRepository(t)

	created, err := repository.Create(ctx, domainauth.User{
		Username: "kasir1", PasswordHash: "hash", Role: domainauth.RoleKasir, Active: true,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repository.SetActive(ctx, created.ID, false); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	stored, err := repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find after deactivate: %v", err)
	}
	if stored.Active {
		t.Error("deactivate: got active, want nonaktif")
	}

	if err := repository.SetActive(ctx, created.ID, true); err != nil {
		t.Fatalf("reactivate: %v", err)
	}

	stored, err = repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find after reactivate: %v", err)
	}
	if !stored.Active {
		t.Error("reactivate: got nonaktif, want active")
	}
}

func TestUserRepositoryListsEveryPenggunaByUsername(t *testing.T) {
	repository, ctx := newUserRepository(t)

	for _, user := range []domainauth.User{
		{Username: "kasir2", PasswordHash: "hash", Role: domainauth.RoleKasir, Active: true},
		{Username: "admin1", PasswordHash: "hash", Role: domainauth.RoleAdmin, Active: true},
		{Username: "kasir1", PasswordHash: "hash", Role: domainauth.RoleKasir, Active: false},
	} {
		if _, err := repository.Create(ctx, user); err != nil {
			t.Fatalf("create %s: %v", user.Username, err)
		}
	}

	listed, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	want := []string{"admin1", "kasir1", "kasir2"}
	if len(listed) != len(want) {
		t.Fatalf("list: got %d Pengguna, want %d", len(listed), len(want))
	}
	for i, username := range want {
		if listed[i].Username != username {
			t.Errorf("list[%d]: got %q, want %q", i, listed[i].Username, username)
		}
	}

	if listed[1].Active {
		t.Error("list: lost the nonaktif state of kasir1")
	}
	if listed[0].Role != domainauth.RoleAdmin {
		t.Errorf("list: got role %q for admin1, want %q", listed[0].Role, domainauth.RoleAdmin)
	}
}

func TestUserRepositoryHasAdmin(t *testing.T) {
	repository, ctx := newUserRepository(t)

	hasAdmin, err := repository.HasAdmin(ctx)
	if err != nil {
		t.Fatalf("has admin on an empty store: %v", err)
	}
	if hasAdmin {
		t.Error("has admin on an empty store: got true, want false")
	}

	// A Kasir is not an Admin: seeding must not stop just because staff exist.
	if _, err := repository.Create(ctx, domainauth.User{
		Username: "kasir1", PasswordHash: "hash", Role: domainauth.RoleKasir, Active: true,
	}); err != nil {
		t.Fatalf("create Kasir: %v", err)
	}

	hasAdmin, err = repository.HasAdmin(ctx)
	if err != nil {
		t.Fatalf("has admin with only a Kasir: %v", err)
	}
	if hasAdmin {
		t.Error("has admin with only a Kasir: got true, want false")
	}

	if _, err := repository.Create(ctx, domainauth.User{
		Username: "admin1", PasswordHash: "hash", Role: domainauth.RoleAdmin, Active: true,
	}); err != nil {
		t.Fatalf("create Admin: %v", err)
	}

	hasAdmin, err = repository.HasAdmin(ctx)
	if err != nil {
		t.Fatalf("has admin with an Admin: %v", err)
	}
	if !hasAdmin {
		t.Error("has admin with an Admin: got false, want true")
	}
}
