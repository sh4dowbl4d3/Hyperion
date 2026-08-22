package database

import (
	"testing"
	"testing/fstest"
)

func migrationFS(files map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for name, content := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(content)}
	}
	return fsys
}

func TestLoadMigrationsSortedAndPaired(t *testing.T) {
	fsys := migrationFS(map[string]string{
		"0002_labs.up.sql":    "CREATE TABLE labs ();",
		"0001_users.up.sql":   "CREATE TABLE users ();",
		"0001_users.down.sql": "DROP TABLE users;",
		"0002_labs.down.sql":  "DROP TABLE labs;",
	})

	migs, err := LoadMigrations(fsys)
	if err != nil {
		t.Fatalf("LoadMigrations() error = %v", err)
	}
	if len(migs) != 2 {
		t.Fatalf("len(migrations) = %d, want 2", len(migs))
	}
	if migs[0].Version != "0001" || migs[0].Name != "users" {
		t.Errorf("first = %s_%s, want 0001_users", migs[0].Version, migs[0].Name)
	}
	if migs[1].Version != "0002" || migs[1].Name != "labs" {
		t.Errorf("second = %s_%s, want 0002_labs", migs[1].Version, migs[1].Name)
	}
	if migs[0].DownSQL == "" || migs[1].DownSQL == "" {
		t.Error("down SQL must be attached to each migration")
	}
}

func TestLoadMigrationsRejectsUnpairedUp(t *testing.T) {
	fsys := migrationFS(map[string]string{
		"0001_users.up.sql": "CREATE TABLE users ();",
	})
	_, err := LoadMigrations(fsys)
	if err == nil {
		t.Fatal("expected error for missing down migration, got nil")
	}
}

func TestLoadMigrationsRejectsOrphanDown(t *testing.T) {
	fsys := migrationFS(map[string]string{
		"0001_users.down.sql": "DROP TABLE users;",
	})
	_, err := LoadMigrations(fsys)
	if err == nil {
		t.Fatal("expected error for orphan down migration, got nil")
	}
}

func TestLoadMigrationsRejectsBadFileName(t *testing.T) {
	fsys := migrationFS(map[string]string{
		"users.sql": "CREATE TABLE users ();",
	})
	_, err := LoadMigrations(fsys)
	if err == nil {
		t.Fatal("expected error for invalid file name, got nil")
	}
}
