//go:build integration

package tests

import (
	"context"
	"sync"
	"testing"

	"moderndvwa/backend/internal/labs/race"
)

// TestRaceVulnerablePathOverRedeems proves the TOCTOU flaw: many concurrent
// redemptions of the same coupon push the counter past its cap.
func TestRaceVulnerablePathOverRedeems(t *testing.T) {
	pool, cleanup := testPool(t)
	defer cleanup()
	ctx := context.Background()

	store := race.NewStore(pool)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}

	// Fresh coupon for this run so prior state cannot mask the race.
	code := "RACE-TEST-" + randomSuffix(t)
	if _, err := pool.Exec(ctx,
		`INSERT INTO race_coupons (code, max_redemptions, redeemed) VALUES ($1, 3, 0)`, code,
	); err != nil {
		t.Fatalf("seed coupon: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM race_coupons WHERE code = $1`, code)
	})

	// Distinct synthetic users are required: UNIQUE(code,user_id) would
	// otherwise reject repeats from one user.
	userIDs := make([]string, 0, 12)
	for i := 0; i < 12; i++ {
		id := createTestUser(t, pool)
		userIDs = append(userIDs, id)
	}

	const workers = 12
	start := make(chan struct{})
	var wg sync.WaitGroup
	results := make([]error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			results[idx] = store.RedeemVulnerable(ctx, userIDs[idx], code)
		}(i)
	}
	close(start)
	wg.Wait()

	successes := 0
	for _, err := range results {
		if err == nil {
			successes++
		}
	}

	st, err := store.StatusForUser(ctx, userIDs[0], code)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if successes <= 3 {
		t.Fatalf("race not exercised: only %d/12 concurrent redemptions succeeded; increase contention or retry", successes)
	}
	if !race.OverRedeemed(st) {
		t.Errorf("vulnerable path did not over-redeem: redeemed=%d max=%d successes=%d", st.Redeemed, st.MaxRedemptions, successes)
	}
}

// TestRaceSafePathNeverOverRedeems proves the transactional fix under the
// identical concurrency profile.
func TestRaceSafePathNeverOverRedeems(t *testing.T) {
	pool, cleanup := testPool(t)
	defer cleanup()
	ctx := context.Background()

	store := race.NewStore(pool)

	code := "SAFE-TEST-" + randomSuffix(t)
	if _, err := pool.Exec(ctx,
		`INSERT INTO race_coupons (code, max_redemptions, redeemed) VALUES ($1, 3, 0)`, code,
	); err != nil {
		t.Fatalf("seed coupon: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM race_coupons WHERE code = $1`, code)
	})

	userIDs := make([]string, 0, 12)
	for i := 0; i < 12; i++ {
		userIDs = append(userIDs, createTestUser(t, pool))
	}

	const workers = 12
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			_ = store.RedeemSafe(ctx, userIDs[idx], code) // errors expected beyond cap
		}(i)
	}
	close(start)
	wg.Wait()

	st, err := store.StatusForUser(ctx, userIDs[0], code)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if st.Redeemed != st.MaxRedemptions {
		t.Errorf("safe path redeemed=%d, want exactly %d", st.Redeemed, st.MaxRedemptions)
	}
	if race.OverRedeemed(st) {
		t.Error("safe path over-redeemed")
	}
}
