//go:build integration

package tests

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"hyperion/backend/internal/database"
	"hyperion/backend/internal/labs"
	"hyperion/backend/internal/users"
	"hyperion/backend/migrations"
)

type progressFixture struct {
	pool   *pgxpool.Pool
	store  *labs.ProgressStore
	userID string
}

func progressTestEnv(t *testing.T) *progressFixture {
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
	t.Cleanup(pool.Close)

	if _, err := database.MigrateUp(ctx, pool, migrations.FS()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	catalog := labs.NewCatalogRepository(pool)
	if err := catalog.Upsert(ctx, sampleMeta("it-progress-lab")); err != nil {
		t.Fatalf("seed lab: %v", err)
	}

	userRepo := users.NewRepository(pool)
	email := uniqueEmail(t)
	user, err := userRepo.Create(ctx, email, "hash-not-used", "user")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	return &progressFixture{
		pool:   pool,
		store:  labs.NewProgressStore(pool),
		userID: user.ID,
	}
}

func TestCompleteLabAwardsXPOnce(t *testing.T) {
	fx := progressTestEnv(t)
	ctx := context.Background()

	first, err := fx.store.CompleteLab(ctx, fx.userID, "it-progress-lab")
	if err != nil {
		t.Fatalf("first completion: %v", err)
	}
	if !first.NewlyCompleted || first.XPAwarded != 75 {
		t.Fatalf("first = %+v, want newly completed with 75 xp", first)
	}

	second, err := fx.store.CompleteLab(ctx, fx.userID, "it-progress-lab")
	if err != nil {
		t.Fatalf("repeat completion: %v", err)
	}
	if second.NewlyCompleted {
		t.Error("repeat completion must not report newly completed")
	}
	if second.XPAwarded != 0 {
		t.Errorf("repeat completion awarded %d extra xp, want 0", second.XPAwarded)
	}

	records, err := fx.store.ForUser(ctx, fx.userID)
	if err != nil {
		t.Fatalf("for user: %v", err)
	}
	if len(records) != 1 || records[0].XPAwarded != 75 {
		t.Fatalf("records = %+v, want single record with 75 xp", records)
	}
}

func TestCompleteLabConcurrentCallsAwardExactlyOnce(t *testing.T) {
	fx := progressTestEnv(t)
	ctx := context.Background()

	const concurrency = 8
	var wg sync.WaitGroup
	newlyCompleted := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := fx.store.CompleteLab(ctx, fx.userID, "it-progress-lab")
			if err != nil {
				t.Errorf("concurrent completion: %v", err)
				return
			}
			newlyCompleted <- result.NewlyCompleted
		}()
	}
	wg.Wait()
	close(newlyCompleted)

	count := 0
	for newly := range newlyCompleted {
		if newly {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("%d concurrent calls reported newly completed, want exactly 1", count)
	}

	summary, err := fx.store.SummaryForUser(ctx, fx.userID)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.XPEarned != 75 || summary.CompletedLabs != 1 {
		t.Fatalf("summary = %+v, want 75 xp earned and 1 completed", summary)
	}
}

func TestSummaryReflectsCatalogTotals(t *testing.T) {
	fx := progressTestEnv(t)
	ctx := context.Background()

	other := sampleMeta("it-summary-lab-2")
	other.XP = 25
	catalog := labs.NewCatalogRepository(fx.pool)
	if err := catalog.Upsert(ctx, other); err != nil {
		t.Fatalf("seed second lab: %v", err)
	}

	if _, err := fx.store.CompleteLab(ctx, fx.userID, "it-progress-lab"); err != nil {
		t.Fatalf("complete: %v", err)
	}

	summary, err := fx.store.SummaryForUser(ctx, fx.userID)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.CompletedLabs != 1 || summary.XPEarned != 75 {
		t.Errorf("completed/xp = %d/%d, want 1/75", summary.CompletedLabs, summary.XPEarned)
	}
	if summary.TotalLabs < 2 {
		t.Errorf("total labs = %d, want at least the 2 seeded labs", summary.TotalLabs)
	}
}

func TestCompleteLabUnknownSlug(t *testing.T) {
	fx := progressTestEnv(t)
	if _, err := fx.store.CompleteLab(context.Background(), fx.userID, "no-such-lab"); err != labs.ErrUnknownLab {
		t.Fatalf("error = %v, want ErrUnknownLab", err)
	}
}
