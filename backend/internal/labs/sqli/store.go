package sqli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"hyperion/backend/internal/labs"
)

const LabSlug = "sqli"

// FlagNote is the value stored on the hidden archive contact. A learner only
// sees it by bypassing the intended WHERE clause through injection.
const FlagNote = "FLAG-SQLI-9d41c2"

type Contact struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Note       string `json:"note"`
}

// Store abstracts the database so handlers can be tested without Postgres.
type Store interface {
	SearchVulnerable(ctx context.Context, q string) ([]Contact, error)
	SearchSafe(ctx context.Context, q string) ([]Contact, error)
}

// Completer is the slice of the progress store labs need to award credit.
type Completer interface {
	CompleteLab(ctx context.Context, userID, labSlug string) (*labs.CompletionResult, error)
}

// BuildVulnerableQuery demonstrates the flaw: untrusted input is concatenated
// straight into SQL text. It exists so tests can assert exact behaviour and is
// used verbatim by the vulnerable store path.
func BuildVulnerableQuery(q string) string {
	return fmt.Sprintf(
		`SELECT name, email, department, note FROM sqli_contacts `+
			`WHERE hidden = false AND name ILIKE '%%%s%%' ORDER BY id`,
		q,
	)
}

// BuildSafeQuery shows the reference fix: the statement is fixed and only the
// parameter value varies.
const safeQuery = `SELECT name, email, department, note FROM sqli_contacts
WHERE hidden = false AND (name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
ORDER BY id`

type PgStore struct {
	pool *pgxpool.Pool
}

func NewPgStore(pool *pgxpool.Pool) *PgStore {
	return &PgStore{pool: pool}
}

func (s *PgStore) scanContacts(ctx context.Context, sqlText string, args ...any) ([]Contact, error) {
	rows, err := s.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("sqli search: %w", err)
	}
	defer rows.Close()

	contacts := []Contact{}
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.Name, &c.Email, &c.Department, &c.Note); err != nil {
			return nil, fmt.Errorf("sqli scan: %w", err)
		}
		contacts = append(contacts, c)
	}
	return contacts, rows.Err()
}

// SearchVulnerable executes the attacker-influenced SQL. Any SQL error from
// malformed input surfaces as a generic failure, mirroring real-world leaks.
func (s *PgStore) SearchVulnerable(ctx context.Context, q string) ([]Contact, error) {
	return s.scanContacts(ctx, BuildVulnerableQuery(q))
}

func (s *PgStore) SearchSafe(ctx context.Context, q string) ([]Contact, error) {
	return s.scanContacts(ctx, safeQuery, q)
}

var ErrEmptyQuery = errors.New("query must not be empty")

// ContainsFlag reports whether a result set leaked the hidden archive record.
func ContainsFlag(contacts []Contact) bool {
	for _, c := range contacts {
		if strings.Contains(c.Note, FlagNote) || strings.EqualFold(c.Name, "Archive Vault") {
			return true
		}
	}
	return false
}
