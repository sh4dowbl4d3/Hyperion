package xss

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"hyperion/backend/internal/labs"
)

const LabSlug = "xss"

// FlagMarker must appear inside a submitted comment body for completion.
// It mimics the proof-of-execution beacons used in real XSS training ranges
// and is inert on its own — it is just a string in synthetic data.
const FlagMarker = "<script>FLAG-XSS-77b1e4</script>"

// Comment is returned by BOTH endpoints. Rendered marks whether the body is
// safe to inject into HTML directly (safe endpoint) or must be treated as raw
// markup (vulnerable endpoint).
type Comment struct {
	ID        int        `json:"id"`
	Author    string     `json:"author"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	Rendered  *string    `json:"rendered,omitempty"`
}

type Store interface {
	ListVulnerable(ctx context.Context) ([]Comment, error)
	CreateVulnerable(ctx context.Context, author, body string) (*Comment, error)
	ListSafe(ctx context.Context) ([]Comment, error)
	CreateSafe(ctx context.Context, author, body string) (*Comment, error)
}

type Completer interface {
	CompleteLab(ctx context.Context, userID, labSlug string) (*labs.CompletionResult, error)
}

const listQuery = `SELECT id, author, body, created_at FROM xss_comments ORDER BY created_at DESC LIMIT 50`

type PgStore struct {
	pool *pgxpool.Pool
}

func NewPgStore(pool *pgxpool.Pool) *PgStore {
	return &PgStore{pool: pool}
}

func scanComments(ctx context.Context, s *PgStore, q string, args ...any) ([]Comment, error) {
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("xss list: %w", err)
	}
	defer rows.Close()
	out := []Comment{}
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.Author, &c.Body, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("xss scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *PgStore) ListVulnerable(ctx context.Context) ([]Comment, error) {
	return scanComments(ctx, s, listQuery)
}

func (s *PgStore) CreateVulnerable(ctx context.Context, author, body string) (*Comment, error) {
	var c Comment
	err := s.pool.QueryRow(ctx,
		`INSERT INTO xss_comments (author, body) VALUES ($1, $2)
		 RETURNING id, author, body, created_at`,
		author, body,
	).Scan(&c.ID, &c.Author, &c.Body, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("xss create: %w", err)
	}
	return &c, nil
}

// CreateSafe is the reference mitigation: markup is neutralised before
// storage so no consumer can execute it regardless of rendering path.
func (s *PgStore) CreateSafe(ctx context.Context, author, body string) (*Comment, error) {
	return s.CreateVulnerable(ctx, author, SanitizeBody(body))
}

// ListSafe escapes stored bodies server-side. The reference frontend can then
// insert `rendered` as trusted markup without enabling script execution.
func (s *PgStore) ListSafe(ctx context.Context) ([]Comment, error) {
	comments, err := scanComments(ctx, s, listQuery)
	if err != nil {
		return nil, err
	}
	for i := range comments {
		r := html.EscapeString(comments[i].Body)
		comments[i].Rendered = &r
	}
	return comments, nil
}

var ErrEmptyBody = errorsNew("comment body must not be empty")

func errorsNew(s string) error { return &emptyBodyError{s} }

type emptyBodyError struct{ msg string }

func (e *emptyBodyError) Error() string { return e.msg }

// SanitizeBody is the reference mitigation: neutralise markup before storage
// so no consumer can execute it regardless of rendering path.
func SanitizeBody(body string) string {
	stripped := stripTags(body)
	return strings.TrimSpace(stripped)
}

func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch r {
		case '<':
			inTag = true
		case '>':
			if inTag {
				inTag = false
				continue
			}
			b.WriteRune(r)
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
