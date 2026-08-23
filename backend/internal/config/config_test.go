package config

import (
	"strings"
	"testing"
	"time"
)

func envLookup(env map[string]string) func(string) string {
	return func(key string) string { return env[key] }
}

const testSecret = "test-secret-0123456789abcdef0123456789abcdef"

func TestLoadDefaults(t *testing.T) {
	env := map[string]string{
		"DATABASE_DSN": "postgres://hyperion:devpassword@localhost:5432/hyperion?sslmode=disable",
		"JWT_SECRET":   testSecret,
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
	if cfg.JWTSecret != testSecret {
		t.Errorf("JWTSecret = %q, want the configured value", cfg.JWTSecret)
	}
	if cfg.JWTTTL != time.Hour {
		t.Errorf("JWTTTL = %v, want default 1h", cfg.JWTTTL)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	env := map[string]string{
		"APP_ENV":      "production",
		"SERVER_ADDR":  "127.0.0.1:9090",
		"LOG_LEVEL":    " DEBUG ",
		"DATABASE_DSN": "postgres://u:p@db:5432/x",
		"JWT_SECRET":   testSecret,
		"JWT_TTL":      "15m",
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
	if cfg.JWTTTL != 15*time.Minute {
		t.Errorf("JWTTTL = %v, want 15m", cfg.JWTTTL)
	}
}

func TestLoadInvalidValues(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"missing dsn", map[string]string{"JWT_SECRET": testSecret}, "DATABASE_DSN"},
		{"invalid env", map[string]string{"APP_ENV": "staging", "JWT_SECRET": testSecret}, "APP_ENV"},
		{"invalid log level", map[string]string{"LOG_LEVEL": "verbose", "JWT_SECRET": testSecret}, "LOG_LEVEL"},
		{"missing jwt secret", map[string]string{"DATABASE_DSN": "postgres://u:p@db/x"}, "JWT_SECRET"},
		{"short jwt secret", map[string]string{
			"DATABASE_DSN": "postgres://u:p@db/x",
			"JWT_SECRET":   "too-short",
		}, "JWT_SECRET"},
		{"bad jwt ttl", map[string]string{
			"DATABASE_DSN": "postgres://u:p@db/x",
			"JWT_SECRET":   testSecret,
			"JWT_TTL":      "soon",
		}, "JWT_TTL"},
		{"zero jwt ttl", map[string]string{
			"DATABASE_DSN": "postgres://u:p@db/x",
			"JWT_SECRET":   testSecret,
			"JWT_TTL":      "0s",
		}, "JWT_TTL"},
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
