package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	connectTimeout  = 5 * time.Second
	defaultMaxConns = 10
	defaultMinConns = 2
	maxConnLifetime = time.Hour
	maxConnIdleTime = 15 * time.Minute
)

func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database dsn: %w", err)
	}
	cfg.MaxConns = defaultMaxConns
	cfg.MinConns = defaultMinConns
	cfg.MaxConnLifetime = maxConnLifetime
	cfg.MaxConnIdleTime = maxConnIdleTime

	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}
