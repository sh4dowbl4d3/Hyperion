package idor

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Document is a per-user resource. OwnerID identifies the rightful owner;
// the lab's vulnerability lives in whether handlers enforce it.
type Document struct {
	ID             int    `json:"id"`
	OwnerID        string `json:"owner_id"`
	Title          string `json:"title"`
	Classification string `json:"classification"`
	Content        string `json:"content"`
}

type Store interface {
	ListForUser(ctx context.Context, userID string) ([]Document, error)
	GetByID(ctx context.Context, id int) (*Document, error)
}

type PgStore struct {
	pool *pgxpool.Pool
}

func NewPgStore(pool *pgxpool.Pool) *PgStore {
	return &PgStore{pool: pool}
}

const docColumns = `id, owner_id::text, title, classification, content`

func (s *PgStore) ListForUser(ctx context.Context, userID string) ([]Document, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT `+docColumns+` FROM idor_documents WHERE owner_id = $1 ORDER BY id`, userID)
	if err != nil {
		return nil, fmt.Errorf("idor list: %w", err)
	}
	defer rows.Close()

	docs := []Document{}
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.OwnerID, &d.Title, &d.Classification, &d.Content); err != nil {
			return nil, fmt.Errorf("idor scan: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s *PgStore) GetByID(ctx context.Context, id int) (*Document, error) {
	var d Document
	err := s.pool.QueryRow(ctx,
		`SELECT `+docColumns+` FROM idor_documents WHERE id = $1`, id,
	).Scan(&d.ID, &d.OwnerID, &d.Title, &d.Classification, &d.Content)
	if err != nil {
		return nil, fmt.Errorf("idor get %d: %w", id, err)
	}
	return &d, nil
}

// EnsureSeeded gives a first-time visitor their own demo documents plus a
// deterministic "victim" document owned by a synthetic service account, so
// the IDOR target always exists. All rows are idempotent.
func (s *PgStore) EnsureSeeded(ctx context.Context, userID string) error {
	if err := s.ensureVictim(ctx); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO idor_documents (owner_id, title, classification, content)
		 SELECT $1, v.title, v.classification, v.content
		 FROM (VALUES
		     ('Welcome memo', 'internal', 'Thanks for joining the pilot programme.'),
		     ('Team roster', 'internal', 'Five engineers, one support rotation.')
		 ) AS v(title, classification, content)
		 WHERE NOT EXISTS (SELECT 1 FROM idor_documents WHERE owner_id = $1)`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("idor seed for %s: %w", userID, err)
	}
	return nil
}

func (s *PgStore) ensureVictim(ctx context.Context) error {
	// Synthetic service account that owns the confidential target document.
	if _, err := s.pool.Exec(ctx,
		`INSERT INTO users (email, password_hash, role_id)
		 SELECT 'idor-victim@example.test', 'x', 1
		 WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'idor-victim@example.test')`,
	); err != nil {
		return fmt.Errorf("idor seed victim user: %w", err)
	}

	_, err := s.pool.Exec(ctx,
		`INSERT INTO idor_documents (owner_id, title, classification, content)
		 SELECT u.id, v.title, v.classification, v.content
		 FROM users u, (VALUES
		     ('Q3 Access Review', 'confidential',
		      'FLAG-IDOR-3f8a92: quarterly access review'),
		     ('Vendor invoices', 'internal', 'Processing as usual.')
		 ) AS v(title, classification, content)
		 WHERE u.email = 'idor-victim@example.test'
		   AND NOT EXISTS (
		       SELECT 1 FROM idor_documents d WHERE d.owner_id = u.id
		   )`,
	)
	if err != nil {
		return fmt.Errorf("idor seed victim: %w", err)
	}
	return nil
}
