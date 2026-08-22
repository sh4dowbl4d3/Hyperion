package database

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var migrationNamePattern = regexp.MustCompile(`^(\d{4})_([a-z0-9_]+)\.(up|down)\.sql$`)

type Migration struct {
	Version string
	Name    string
	UpSQL   string
	DownSQL string
}

func LoadMigrations(fsys fs.FS) ([]Migration, error) {
	ups := make(map[string]*Migration)
	downs := make(map[string]string)

	entries, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return nil, fmt.Errorf("scan migrations: %w", err)
	}
	for _, entry := range entries {
		m := migrationNamePattern.FindStringSubmatch(path.Base(entry))
		if m == nil {
			return nil, fmt.Errorf("invalid migration file name %q: want NNNN_description.up.sql", path.Base(entry))
		}
		version, name, kind := m[1], m[2], m[3]
		content, err := fs.ReadFile(fsys, entry)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry, err)
		}
		switch kind {
		case "up":
			if _, dup := ups[version]; dup {
				return nil, fmt.Errorf("duplicate up migration for version %s", version)
			}
			ups[version] = &Migration{Version: version, Name: name, UpSQL: string(content)}
		case "down":
			if _, dup := downs[version]; dup {
				return nil, fmt.Errorf("duplicate down migration for version %s", version)
			}
			downs[version] = string(content)
		}
	}

	list := make([]Migration, 0, len(ups))
	for version, mig := range ups {
		down, ok := downs[version]
		if !ok {
			return nil, fmt.Errorf("migration %s (%s) has no matching down migration", version, mig.Name)
		}
		delete(downs, version)
		mig.DownSQL = down
		list = append(list, *mig)
	}
	if len(downs) > 0 {
		orphan := maps_key(downs)
		return nil, fmt.Errorf("down migration %s has no matching up migration", orphan)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Version < list[j].Version })
	return list, nil
}

func MigrateUp(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) ([]string, error) {
	migrations, err := LoadMigrations(fsys)
	if err != nil {
		return nil, err
	}
	if err := ensureSchemaMigrations(ctx, pool); err != nil {
		return nil, err
	}

	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return nil, err
	}

	var newlyApplied []string
	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}
		if err := applyMigration(ctx, pool, m); err != nil {
			return newlyApplied, fmt.Errorf("apply migration %s_%s: %w", m.Version, m.Name, err)
		}
		newlyApplied = append(newlyApplied, m.Version+"_"+m.Name)
	}
	return newlyApplied, nil
}

func StepDown(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) (string, error) {
	migrations, err := LoadMigrations(fsys)
	if err != nil {
		return "", err
	}
	byVersion := make(map[string]Migration, len(migrations))
	for _, m := range migrations {
		byVersion[m.Version] = m
	}

	var latest string
	err = pool.QueryRow(ctx,
		`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`,
	).Scan(&latest)
	if err != nil {
		return "", fmt.Errorf("find latest applied migration: %w", err)
	}

	m, ok := byVersion[latest]
	if !ok {
		return "", fmt.Errorf("applied migration %s not found in embedded migrations", latest)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Conn().PgConn().Exec(ctx, m.DownSQL).ReadAll(); err != nil {
		return "", fmt.Errorf("execute down migration %s_%s: %w", m.Version, m.Name, err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, m.Version); err != nil {
		return "", fmt.Errorf("record rollback of %s: %w", m.Version, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit rollback of %s: %w", m.Version, err)
	}
	return m.Version + "_" + m.Name, nil
}

func ensureSchemaMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version text PRIMARY KEY,
    applied_at timestamptz NOT NULL DEFAULT now()
)`
	_, err := pool.Exec(ctx, ddl)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	return nil
}

func appliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[string]bool, error) {
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, m Migration) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if strings.TrimSpace(m.UpSQL) == "" {
		return fmt.Errorf("up migration is empty")
	}
	if _, err := tx.Conn().PgConn().Exec(ctx, m.UpSQL).ReadAll(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO schema_migrations (version) VALUES ($1)`, m.Version,
	); err != nil {
		return fmt.Errorf("record version: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func maps_key(m map[string]string) string {
	for k := range m {
		return k
	}
	return ""
}
