package config

import (
	"strings"
	"testing"
)

func envLookup(env map[string]string) func(string) string {
	return func(key string) string { return env[key] }
}

func TestLoadDefaults(t *testing.T) {
	env := map[string]string{
		"DATABASE_DSN": "postgres://moderndvwa:devpassword@localhost:5432/moderndvwa?sslmode=disable",
	}
	cfg, err := Load(envLookup(env))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Env != EnvDevelopment {
		t.Errorf("Env = %q, want %q", cfg.Env, EnvDevelopment)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":8080")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.DatabaseDSN != env["DATABASE_DSN"] {
		t.Errorf("DatabaseDSN = %q, want %q", cfg.DatabaseDSN, env["DATABASE_DSN"])
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	env := map[string]string{
		"APP_ENV":      "production",
		"SERVER_ADDR":  "127.0.0.1:9090",
		"LOG_LEVEL":    " DEBUG ",
		"DATABASE_DSN": "postgres://u:p@db:5432/x",
	}
	cfg, err := Load(envLookup(env))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Env != EnvProduction {
		t.Errorf("Env = %q, want %q", cfg.Env, EnvProduction)
	}
	if cfg.Addr != "127.0.0.1:9090" {
		t.Errorf("Addr = %q, want 127.0.0.1:9090", cfg.Addr)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug (trimmed and lowercased)", cfg.LogLevel)
	}
}

func TestLoadInvalidValues(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"missing dsn", map[string]string{}, "DATABASE_DSN"},
		{"invalid env", map[string]string{"APP_ENV": "staging"}, "APP_ENV"},
		{"invalid log level", map[string]string{"LOG_LEVEL": "verbose"}, "LOG_LEVEL"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(envLookup(tc.env))
			if err == nil {
				t.Fatal("Load() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}
