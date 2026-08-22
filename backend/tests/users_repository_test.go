//go:build integration

package tests

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"moderndvwa/backend/internal/database"
	"moderndvwa/backend/internal/users"
	"moderndvwa/backend/migrations"
)

func uniqueEmail(t *testing.T) string {
	t.Helper()
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("generate email suffix: %v", err)
	}
	return fmt.Sprintf("it-%s@example.test", hex.EncodeToString(buf))
}

func userTestRepo(t *testing.T) (*users.Repository, func()) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set; skipping database integration test")
	}
	ctx := context.Background()
	pool, err := database.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if _, err := database.MigrateUp(ctx, pool, migrations.FS()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return users.NewRepository(pool), pool.Close
}

func TestUserRepositoryCreateAndLookup(t *testing.T) {
	repo, cleanup := userTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	email := uniqueEmail(t)

	created, err := repo.Create(ctx, email, "bcrypt-hash-placeholder", "user")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("created user has empty id")
	}
	if created.Email != email {
		t.Errorf("email = %q, want %q", created.Email, email)
	}
	if created.Role != "user" {
		t.Errorf("role = %q, want user", created.Role)
	}

	got, err := repo.ByEmail(ctx, email)
	if err != nil {
		t.Fatalf("by email: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ByEmail id = %q, want %q", got.ID, created.ID)
	}

	byID, err := repo.ByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("by id: %v", err)
	}
	if byID.Email != email {
		t.Errorf("ByID email = %q, want %q", byID.Email, email)
	}
}

func TestUserRepositoryByEmailIsCaseInsensitive(t *testing.T) {
	repo, cleanup := userTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	local := "IT-" + uniqueEmail(t)
	email := local + "@Example.test"

	if _, err := repo.Create(ctx, email, "hash", "user"); err != nil {
		t.Fatalf("create: %v", err)
	}

	lowered := "it-" + local[len("IT-"):] + "@example.test"
	if _, err := repo.ByEmail(ctx, email); err != nil {
		t.Fatalf("lookup with original mixed case failed: %v", err)
	}
	got, err := repo.ByEmail(ctx, lowered)
	if err != nil {
		t.Fatalf("lookup lowercased failed: %v", err)
	}
	if got.Email != lowered {
		t.Errorf("stored email = %q, want normalized lowercase %q", got.Email, lowered)
	}
}

func TestUserRepositoryDuplicateEmail(t *testing.T) {
	repo, cleanup := userTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	email := uniqueEmail(t)

	if _, err := repo.Create(ctx, email, "hash", "user"); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := repo.Create(ctx, email, "other-hash", "user")
	if err == nil {
		t.Fatal("expected duplicate create to fail")
	}
	if err != users.ErrEmailTaken {
		t.Fatalf("error = %v, want ErrEmailTaken", err)
	}
}

func TestUserRepositoryUnknownUser(t *testing.T) {
	repo, cleanup := userTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	if _, err := repo.ByEmail(ctx, "nobody-here@example.test"); err != users.ErrNotFound {
		t.Fatalf("ByEmail error = %v, want ErrNotFound", err)
	}
	if _, err := repo.ByID(ctx, "00000000-0000-0000-0000-000000000000"); err != users.ErrNotFound {
		t.Fatalf("ByID error = %v, want ErrNotFound", err)
	}
}
