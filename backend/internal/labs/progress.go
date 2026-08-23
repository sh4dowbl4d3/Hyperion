package labs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUnknownLab = errors.New("lab not found in catalog")

type ProgressStatus string

const (
	StatusNotStarted ProgressStatus = "not_started"
	StatusInProgress ProgressStatus = "in_progress"
	StatusCompleted  ProgressStatus = "completed"
)

type Record struct {
	LabSlug     string         `json:"lab_slug"`
	Status      ProgressStatus `json:"status"`
	XPAwarded   int            `json:"xp_awarded"`
	CompletedAt *time.Time     `json:"completed_at"`
}

type CompletionResult struct {
	Record         Record
	NewlyCompleted bool
	XPAwarded      int
}

type Summary struct {
	TotalLabs     int `json:"total_labs"`
	CompletedLabs int `json:"completed_labs"`
	XPAvailable   int `json:"xp_available"`
	XPEarned      int `json:"xp_earned"`
}

type ProgressStore struct {
	pool *pgxpool.Pool
}

func NewProgressStore(pool *pgxpool.Pool) *ProgressStore {
	return &ProgressStore{pool: pool}
}

func (s *ProgressStore) ForUser(ctx context.Context, userID string) ([]Record, error) {
	const q = `
SELECT lab_slug, status, xp_awarded, completed_at
FROM progress WHERE user_id = $1 ORDER BY lab_slug`

	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list progress: %w", err)
	}
	defer rows.Close()

	records := []Record{}
	for rows.Next() {
		var rec Record
		if err := rows.Scan(&rec.LabSlug, &rec.Status, &rec.XPAwarded, &rec.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan progress: %w", err)
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (s *ProgressStore) CompleteLab(ctx context.Context, userID, labSlug string) (*CompletionResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin completion transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var labXP int
	err = tx.QueryRow(ctx, `SELECT xp FROM labs WHERE slug = $1`, labSlug).Scan(&labXP)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUnknownLab
	}
	if err != nil {
		return nil, fmt.Errorf("load lab %s: %w", labSlug, err)
	}

	inserted, err := tx.Exec(ctx,
		`INSERT INTO progress (user_id, lab_slug, status, xp_awarded, completed_at)
		 VALUES ($1, $2, $3, $4, now())
		 ON CONFLICT (user_id, lab_slug) DO NOTHING`,
		userID, labSlug, string(StatusCompleted), labXP,
	)
	if err != nil {
		return nil, fmt.Errorf("insert completion: %w", err)
	}

	if inserted.RowsAffected() == 1 {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit completion: %w", err)
		}
		return &CompletionResult{
			Record:         Record{LabSlug: labSlug, Status: StatusCompleted, XPAwarded: labXP},
			NewlyCompleted: true,
			XPAwarded:      labXP,
		}, nil
	}

	var status ProgressStatus
	var existingXP int
	var completedAt *time.Time
	err = tx.QueryRow(ctx,
		`SELECT status, xp_awarded, completed_at FROM progress WHERE user_id = $1 AND lab_slug = $2 FOR UPDATE`,
		userID, labSlug,
	).Scan(&status, &existingXP, &completedAt)
	if err != nil {
		return nil, fmt.Errorf("lock progress row: %w", err)
	}

	if status == StatusCompleted {
		return &CompletionResult{
			Record:         Record{LabSlug: labSlug, Status: status, XPAwarded: existingXP, CompletedAt: completedAt},
			NewlyCompleted: false,
			XPAwarded:      0,
		}, nil
	}

	if _, err := tx.Exec(ctx,
		`UPDATE progress SET status = $3, xp_awarded = $4, completed_at = now(), updated_at = now()
		 WHERE user_id = $1 AND lab_slug = $2`,
		userID, labSlug, string(StatusCompleted), labXP,
	); err != nil {
		return nil, fmt.Errorf("update completion: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit completion: %w", err)
	}
	return &CompletionResult{
		Record:         Record{LabSlug: labSlug, Status: StatusCompleted, XPAwarded: labXP},
		NewlyCompleted: true,
		XPAwarded:      labXP,
	}, nil
}

func (s *ProgressStore) SummaryForUser(ctx context.Context, userID string) (*Summary, error) {
	const q = `
SELECT
    (SELECT COUNT(*) FROM labs),
    (SELECT COALESCE(SUM(xp), 0) FROM labs),
    (SELECT COUNT(*) FROM progress WHERE user_id = $1 AND status = 'completed'),
    (SELECT COALESCE(SUM(xp_awarded), 0) FROM progress WHERE user_id = $1 AND status = 'completed')`

	var summary Summary
	err := s.pool.QueryRow(ctx, q, userID).Scan(
		&summary.TotalLabs, &summary.XPAvailable, &summary.CompletedLabs, &summary.XPEarned,
	)
	if err != nil {
		return nil, fmt.Errorf("progress summary: %w", err)
	}
	return &summary, nil
}
