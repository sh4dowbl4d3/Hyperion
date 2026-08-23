//go:build integration

package tests

import (
	"context"
	"os"
	"testing"

	"hyperion/backend/internal/database"
	"hyperion/backend/internal/labs"
	"hyperion/backend/migrations"
)

func labTestRepo(t *testing.T) (*labs.CatalogRepository, func()) {
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
	return labs.NewCatalogRepository(pool), pool.Close
}

func sampleMeta(slug string) labs.Meta {
	return labs.Meta{
		Slug:              slug,
		Name:              "Sample " + slug,
		Description:       "Seeded by integration tests.",
		Objective:         "Prove catalog persistence.",
		Category:          "testing",
		Difficulty:        labs.DifficultyMedium,
		XP:                75,
		VulnerabilityType: "sample-type",
		Hints:             []string{"hint one", "hint two"},
	}
}

func TestCatalogUpsertInsertsAndUpdates(t *testing.T) {
	repo, cleanup := labTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	if err := repo.Upsert(ctx, sampleMeta("it-cat-lab")); err != nil {
		t.Fatalf("upsert insert: %v", err)
	}

	got, err := repo.Get(ctx, "it-cat-lab")
	if err != nil {
		t.Fatalf("get after insert: %v", err)
	}
	if got.Name != "Sample it-cat-lab" || got.XP != 75 || len(got.Hints) != 2 {
		t.Fatalf("unexpected entry: %+v", got)
	}

	updated := sampleMeta("it-cat-lab")
	updated.Name = "Renamed Lab"
	updated.XP = 120
	if err := repo.Upsert(ctx, updated); err != nil {
		t.Fatalf("upsert update: %v", err)
	}

	got, err = repo.Get(ctx, "it-cat-lab")
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if got.Name != "Renamed Lab" || got.XP != 120 {
		t.Errorf("update not applied: %+v", got)
	}
}

func TestCatalogListIncludesSeededLabs(t *testing.T) {
	repo, cleanup := labTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	err := repo.Seed(ctx, []labs.Meta{sampleMeta("it-list-a"), sampleMeta("it-list-b")})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	entries, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	slugs := map[string]bool{}
	for _, e := range entries {
		slugs[e.Slug] = true
	}
	if !slugs["it-list-a"] || !slugs["it-list-b"] {
		t.Errorf("seeded labs missing from list: %v", slugs)
	}
}

func TestCatalogGetUnknownLab(t *testing.T) {
	repo, cleanup := labTestRepo(t)
	defer cleanup()

	if _, err := repo.Get(context.Background(), "does-not-exist"); err == nil {
		t.Fatal("expected error for unknown slug")
	}
}
