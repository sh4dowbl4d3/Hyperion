# Lab: Race Condition — Limited Coupon Drop

- **Slug:** `race-condition`
- **Category:** business-logic
- **Difficulty:** hard
- **XP:** 200

## Objective

Send many concurrent redemptions of `LAUNCH-2026` to
`/api/v1/targets/race-condition/redeem` and push the global counter past its
cap. Confirm the transactional twin at `/redeem-safe` refuses over-redemption.

## Learning goal

Understand time-of-check-to-time-of-use (TOCTOU) flaws in business logic: when
a validation step and the state change it protects are not atomic, concurrent
requests can all pass validation before any of them commits.

## Scenario

A launch coupon may be redeemed exactly five times in total. The vulnerable
endpoint reads remaining capacity and inserts the redemption as two separate
statements inside a transaction with **no row lock**, leaving a window where
every concurrent request observes spare capacity. The safe endpoint reserves
capacity with a single conditional UPDATE whose row lock serialises contenders.

## Walkthrough (spoilers)

1. Check state: `GET /coupon` shows `max_redemptions: 5`.
2. Fire 10+ simultaneous POSTs to `/redeem` with `{"code":"LAUNCH-2026"}`
   (a sequential attacker can never win the gap).
3. Observe nearly all requests succeed and the counter land far above the cap
   (12 parallel attempts produced 13/5 during development testing).
4. Repeat against `/redeem-safe`: exactly one success per user, counter never
   exceeds the cap, excess requests receive 409.

## Impact

Financial loss (coupons, refunds, withdrawals), inventory oversell, duplicate
payouts, and voting/rating manipulation — bugs that pass every functional test
written sequentially.

## Secure implementation

`internal/labs/race/store.go` contains both paths:

- `RedeemVulnerable` — read check + insert + counter bump without locking.
- `RedeemSafe` —

```sql
UPDATE race_coupons SET redeemed = redeemed + 1
WHERE code = $1 AND redeemed < max_redemptions;
```

The row lock taken by the UPDATE serialises contenders; zero rows affected
means exhausted (409). A transactional per-user existence check enforces
one-redemption-per-user without relying on schema constraints alone.

## Mitigations

1. Make guard and mutation atomic — conditional UPDATE, or `SELECT … FOR
   UPDATE` before acting.
2. Enforce idempotency/uniqueness at the database level as backstop.
3. Load-test critical counters with true concurrency, not sequential tests.
4. Keep transactions short; long-held locks trade races for throughput.

## Testing

- Handler tests (`lab_test.go`): routing between twins, conflict/not-found
  mapping, validation, completion awarded only when over-redemption is observed.
- Integration tests (`tests/race_test.go`, build tag `integration`, real
  Postgres): 12 goroutines race a fresh 3-cap coupon — the vulnerable path
  demonstrably over-redeems while the safe path lands at exactly the cap.
