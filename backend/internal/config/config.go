package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Config struct {
	Env         string
	Addr        string
	LogLevel    string
	DatabaseDSN string
	JWTSecret   string
	JWTTTL      time.Duration
}

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"

	minJWTSecretBytes = 32
	defaultJWTTTL     = time.Hour

	jwtSecretEnv = "JWT_SECRET"
	jwtTTLEnv    = "JWT_TTL"
)

var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

func Load(lookup func(string) string) (*Config, error) {
	cfg := &Config{
		Env:         valueOrDefault(lookup, "APP_ENV", EnvDevelopment),
		Addr:        valueOrDefault(lookup, "SERVER_ADDR", ":8080"),
		LogLevel:    strings.ToLower(valueOrDefault(lookup, "LOG_LEVEL", "info")),
		DatabaseDSN: valueOrDefault(lookup, "DATABASE_DSN", ""),
		JWTSecret:   strings.TrimSpace(lookup(jwtSecretEnv)),
	}

	if cfg.Env != EnvDevelopment && cfg.Env != EnvProduction {
		return nil, fmt.Errorf("invalid APP_ENV %q: must be %q or %q", cfg.Env, EnvDevelopment, EnvProduction)
	}
	if cfg.Addr == "" {
		return nil, errors.New("SERVER_ADDR must not be empty")
	}
	if !validLogLevels[cfg.LogLevel] {
		return nil, fmt.Errorf("invalid LOG_LEVEL %q: must be debug, info, warn or error", cfg.LogLevel)
	}
	if cfg.DatabaseDSN == "" {
		return nil, errors.New("DATABASE_DSN is required")
	}
	if len(cfg.JWTSecret) < minJWTSecretBytes {
		return nil, fmt.Errorf("%s must be at least %d bytes (generate one with: openssl rand -base64 48)", jwtSecretEnv, minJWTSecretBytes)
	}
	ttl := strings.TrimSpace(lookup(jwtTTLEnv))
	if ttl == "" {
		cfg.JWTTTL = defaultJWTTTL
	} else {
		parsed, err := time.ParseDuration(ttl)
		if err != nil {
			return nil, fmt.Errorf("invalid %s %q: must be a duration like 15m or 1h", jwtTTLEnv, ttl)
		}
		if parsed < time.Minute {
			return nil, fmt.Errorf("invalid %s %q: must be at least 1m", jwtTTLEnv, ttl)
		}
		cfg.JWTTTL = parsed
	}
	return cfg, nil
}

func valueOrDefault(lookup func(string) string, key, fallback string) string {
	if v := strings.TrimSpace(lookup(key)); v != "" {
		return v
	}
	return fallback
}
