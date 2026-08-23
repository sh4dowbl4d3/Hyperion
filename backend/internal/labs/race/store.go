package race

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"moderndvwa/backend/internal/labs"
)

const LabSlug = "race-condition"

// CouponCode is the seeded limited-use coupon.
const CouponCode = "LAUNCH-2026"

var (
	ErrCouponNotFound = errors.New("coupon not found")
	ErrExhausted      = errors.New("coupon redemption limit reached")
	ErrAlreadyUsed    = errors.New("user already redeemed this coupon")
)

type Completer interface {
	CompleteLab(ctx context.Context, userID, labSlug string) (*labs.CompletionResult, error)
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// RedeemVulnerable implements the classic TOCTOU flaw: it reads the remaining
// count and inserts the redemption in SEPARATE statements with no lock, so
// concurrent requests can all observe capacity and all succeed.
func (s *Store) RedeemVulnerable(ctx context.Context, userID, code string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var maxRedemptions, redeemed int
	err = tx.QueryRow(ctx,
		`SELECT max_redemptions, redeemed FROM race_coupons WHERE code = $1`, code,
	).Scan(&maxRedemptions, &redeemed)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrCouponNotFound
	}
	if err != nil {
		return fmt.Errorf("load coupon: %w", err)
	}
	if redeemed >= maxRedemptions {
		return ErrExhausted
	}
	// Deliberate race window: another transaction can pass this same check
	// before ours commits. No FOR UPDATE, no atomic guard.

	if _, err := tx.Exec(ctx,
		`INSERT INTO race_redemptions (code, user_id) VALUES ($1, $2)`,
		code, userID,
	); err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23505" {
			return ErrAlreadyUsed
		}
		return fmt.Errorf("insert redemption: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE race_coupons SET redeemed = redeemed + 1 WHERE code = $1`, code,
	); err != nil {
		return fmt.Errorf("bump counter: %w", err)
	}

	return tx.Commit(ctx)
}

// RedeemSafe is the reference fix: a single conditional INSERT guarded by the
// UNIQUE(code, user_id) constraint and an atomic capacity check inside the
// same statement, executed under row lock via the counter update ordering.
func (s *Store) RedeemSafe(ctx context.Context, userID, code string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// Fix 1: serialise on the coupon row first.
	tag, err := tx.Exec(ctx,
		`UPDATE race_coupons SET redeemed = redeemed + 1
		 WHERE code = $1 AND redeemed < max_redemptions`, code)
	if err != nil {
		return fmt.Errorf("reserve capacity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		err := tx.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM race_coupons WHERE code = $1)`, code).Scan(&exists)
		if err != nil {
			return fmt.Errorf("load coupon: %w", err)
		}
		if !exists {
			return ErrCouponNotFound
		}
		return ErrExhausted
	}

	// Fix 2: per-user uniqueness is enforced transactionally — the row lock
	// on the coupon serialises contenders, and an explicit check rejects
	// users who already redeemed.
	var already bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM race_redemptions WHERE code = $1 AND user_id = $2 FOR UPDATE)`,
		code, userID,
	).Scan(&already); err != nil {
		return fmt.Errorf("redemption check: %w", err)
	}
	if already {
		return ErrAlreadyUsed
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO race_redemptions (code, user_id) VALUES ($1, $2)`,
		code, userID,
	); err != nil {
		return fmt.Errorf("insert redemption: %w", err)
	}

	return tx.Commit(ctx)
}

type Status struct {
	Code           string `json:"code"`
	MaxRedemptions int    `json:"max_redemptions"`
	Redeemed       int    `json:"redeemed"`
	Mine           bool   `json:"mine"`
}

func (s *Store) StatusForUser(ctx context.Context, userID, code string) (*Status, error) {
	st := &Status{Code: code}
	err := s.pool.QueryRow(ctx,
		`SELECT max_redemptions, redeemed FROM race_coupons WHERE code = $1`, code,
	).Scan(&st.MaxRedemptions, &st.Redeemed)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCouponNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("coupon status: %w", err)
	}
	err = s.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM race_redemptions WHERE code = $1 AND user_id = $2)`,
		code, userID,
	).Scan(&st.Mine)
	if err != nil {
		return nil, fmt.Errorf("redemption check: %w", err)
	}
	return st, nil
}

// OverRedeemed reports whether the counter exceeded the allowed maximum —
// the observable proof that the race was won by attackers.
func OverRedeemed(st *Status) bool {
	return st.Redeemed > st.MaxRedemptions
}

var _ = labs.CompletionResult{} // keep labs import for completer implementations in lab.go
