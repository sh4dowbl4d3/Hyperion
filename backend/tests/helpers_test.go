//go:build integration

package tests

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func randomSuffix(t *testing.T) string {
	t.Helper()
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return hex.EncodeToString(buf)
}

// createTestUser inserts a synthetic user with a unique email and returns id.
func createTestUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var id string
	email := fmt.Sprintf("race-%s@example.test", randomSuffix(t))
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash, role_id)
		 VALUES ($1, 'x', 1) RETURNING id::text`,
		email,
	).Scan(&id)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE email = $1`, email)
	})
	return id
}
