package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound   = errors.New("user not found")
	ErrEmailTaken = errors.New("email already registered")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const userColumns = `u.id, u.email, u.password_hash, r.name, u.created_at, u.updated_at`

const userJoin = `FROM users u JOIN roles r ON r.id = u.role_id`

func (r *Repository) Create(ctx context.Context, email, passwordHash, roleName string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	const q = `
INSERT INTO users (email, password_hash, role_id)
VALUES ($1, $2, (SELECT id FROM roles WHERE name = $3))
RETURNING id, email, password_hash, $3, created_at, updated_at`

	row := r.pool.QueryRow(ctx, q, email, passwordHash, roleName)
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return &u, nil
}

func (r *Repository) ByEmail(ctx context.Context, email string) (*User, error) {
	const q = `SELECT ` + userColumns + ` ` + userJoin + ` WHERE u.email = lower($1)`
	return scanUser(r.pool.QueryRow(ctx, q, email))
}

func (r *Repository) ByID(ctx context.Context, id string) (*User, error) {
	const q = `SELECT ` + userColumns + ` ` + userJoin + ` WHERE u.id = $1`
	return scanUser(r.pool.QueryRow(ctx, q, id))
}

func scanUser(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}
