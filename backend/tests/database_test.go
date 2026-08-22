//go:build integration

package tests

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"moderndvwa/backend/internal/database"
	"moderndvwa/backend/migrations"
)

func testPool(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set; skipping database integration test")
	}
	ctx := context.Background()
	pool, err := database.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return pool, pool.Close
}

func TestMigrationsApplyIdempotently(t *testing.T) {
	pool, cleanup := testPool(t)
	defer cleanup()

	ctx := context.Background()
	fsys := migrations.FS()

	all, err := database.LoadMigrations(fsys)
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}

	applied, err := database.MigrateUp(ctx, pool, fsys)
	if err != nil {
		t.Fatalf("first MigrateUp: %v", err)
	}
	if len(applied) != len(all) {
		t.Fatalf("first run applied %d migrations, want %d", len(applied), len(all))
	}

	reapplied, err := database.MigrateUp(ctx, pool, fsys)
	if err != nil {
		t.Fatalf("second MigrateUp: %v", err)
	}
	if len(reapplied) != 0 {
		t.Fatalf("second run applied %d migrations, want 0 (migrations must be idempotent)", len(reapplied))
	}
}

func tableExists(t *testing.T, pool *pgxpool.Pool, table string) bool {
	t.Helper()
	var exists bool
	err := pool.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)`,
		table,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("check table %s: %v", table, err)
	}
	return exists
}

func TestMigrationsCreateExpectedTables(t *testing.T) {
	pool, cleanup := testPool(t)
	defer cleanup()

	ctx := context.Background()
	if _, err := database.MigrateUp(ctx, pool, migrations.FS()); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	for _, table := range []string{"roles", "users", "labs", "progress"} {
		if !tableExists(t, pool, table) {
			t.Errorf("expected table %q to exist after migrations", table)
		}
	}
}

func TestRolesSeeded(t *testing.T) {
	pool, cleanup := testPool(t)
	defer cleanup()

	ctx := context.Background()
	if _, err := database.MigrateUp(ctx, pool, migrations.FS()); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	rows, err := pool.Query(ctx, `SELECT name FROM roles ORDER BY id`)
	if err != nil {
		t.Fatalf("query roles: %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan role: %v", err)
		}
		got = append(got, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate roles: %v", err)
	}
	want := []string{"user", "admin"}
	if len(got) != len(want) {
		t.Fatalf("seeded roles = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("roles[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestStepDownRevertsLatestMigration(t *testing.T) {
	pool, cleanup := testPool(t)
	defer cleanup()

	ctx := context.Background()
	fsys := migrations.FS()
	if _, err := database.MigrateUp(ctx, pool, fsys); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	all, err := database.LoadMigrations(fsys)
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}
	last := all[len(all)-1]

	reverted, err := database.StepDown(ctx, pool, fsys)
	if err != nil {
		t.Fatalf("StepDown: %v", err)
	}
	if reverted != last.Version+"_"+last.Name {
		t.Errorf("reverted %q, want %q", reverted, last.Version+"_"+last.Name)
	}
	if tableExists(t, pool, "progress") {
		t.Error("progress table should be dropped by reverting the latest migration")
	}

	if _, err := database.MigrateUp(ctx, pool, fsys); err != nil {
		t.Fatalf("reapply after StepDown: %v", err)
	}
	if !tableExists(t, pool, "progress") {
		t.Error("progress table should exist again after reapply")
	}
}
