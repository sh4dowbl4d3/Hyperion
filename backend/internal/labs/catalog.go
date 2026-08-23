package labs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CatalogEntry struct {
	Slug              string
	Name              string
	Description       string
	Objective         string
	Category          string
	Difficulty        Difficulty
	XP                int
	VulnerabilityType string
	Hints             []string
}

type CatalogRepository struct {
	pool *pgxpool.Pool
}

func NewCatalogRepository(pool *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{pool: pool}
}

func (r *CatalogRepository) Upsert(ctx context.Context, meta Meta) error {
	hints, err := json.Marshal(meta.Hints)
	if err != nil {
		return fmt.Errorf("marshal hints for %s: %w", meta.Slug, err)
	}
	const q = `
INSERT INTO labs (slug, name, description, category, difficulty, xp, vulnerability_type, objective, hints)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    difficulty = EXCLUDED.difficulty,
    xp = EXCLUDED.xp,
    vulnerability_type = EXCLUDED.vulnerability_type,
    objective = EXCLUDED.objective,
    hints = EXCLUDED.hints`
	_, err = r.pool.Exec(ctx, q,
		meta.Slug, meta.Name, meta.Description, meta.Category,
		string(meta.Difficulty), meta.XP, meta.VulnerabilityType, meta.Objective, hints,
	)
	if err != nil {
		return fmt.Errorf("upsert lab %s: %w", meta.Slug, err)
	}
	return nil
}

func (r *CatalogRepository) Seed(ctx context.Context, metas []Meta) error {
	for _, meta := range metas {
		if err := r.Upsert(ctx, meta); err != nil {
			return err
		}
	}
	return nil
}

const entryColumns = `slug, name, description, category, difficulty, xp, vulnerability_type, objective, hints`

func scanEntry(row interface{ Scan(dest ...any) error }) (*CatalogEntry, error) {
	var e CatalogEntry
	var difficulty string
	var hintsJSON []byte
	if err := row.Scan(&e.Slug, &e.Name, &e.Description, &e.Category, &difficulty, &e.XP, &e.VulnerabilityType, &e.Objective, &hintsJSON); err != nil {
		return nil, err
	}
	e.Difficulty = Difficulty(difficulty)
	if err := json.Unmarshal(hintsJSON, &e.Hints); err != nil {
		return nil, fmt.Errorf("unmarshal hints for %s: %w", e.Slug, err)
	}
	if e.Hints == nil {
		e.Hints = []string{}
	}
	return &e, nil
}

func (r *CatalogRepository) List(ctx context.Context) ([]CatalogEntry, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+entryColumns+` FROM labs ORDER BY slug`)
	if err != nil {
		return nil, fmt.Errorf("list labs: %w", err)
	}
	defer rows.Close()

	var entries []CatalogEntry
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}
	if entries == nil {
		entries = []CatalogEntry{}
	}
	return entries, rows.Err()
}

func (r *CatalogRepository) Get(ctx context.Context, slug string) (*CatalogEntry, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+entryColumns+` FROM labs WHERE slug = $1`, slug)
	entry, err := scanEntry(row)
	if err != nil {
		return nil, fmt.Errorf("get lab %s: %w", slug, err)
	}
	return entry, nil
}

// PruneMissing removes catalog rows whose slugs are absent from metas, as long
// as no user progress references them. This keeps the listing aligned with the
// compiled-in lab registry even when a database volume outlives a code change
// (e.g. removed or renamed labs) or integration-test residue. Rows with
// existing progress are kept so user history is never destroyed.
func (r *CatalogRepository) PruneMissing(ctx context.Context, metas []Meta) error {
	if len(metas) == 0 {
		return nil
	}
	slugs := make([]string, 0, len(metas))
	for _, m := range metas {
		slugs = append(slugs, m.Slug)
	}
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM labs l WHERE NOT (l.slug = ANY($1))
		 AND NOT EXISTS (SELECT 1 FROM progress p WHERE p.lab_slug = l.slug)`,
		slugs,
	)
	if err != nil {
		return fmt.Errorf("prune missing labs: %w", err)
	}
	_ = tag
	return nil
}
