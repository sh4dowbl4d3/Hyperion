package config

import (
	"errors"
	"fmt"
	"strings"
)

type Config struct {
	Env      string
	Addr     string
	LogLevel string
}

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

func Load(lookup func(string) string) (*Config, error) {
	cfg := &Config{
		Env:      valueOrDefault(lookup, "APP_ENV", EnvDevelopment),
		Addr:     valueOrDefault(lookup, "SERVER_ADDR", ":8080"),
		LogLevel: strings.ToLower(valueOrDefault(lookup, "LOG_LEVEL", "info")),
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
	return cfg, nil
}

func valueOrDefault(lookup func(string) string, key, fallback string) string {
	if v := strings.TrimSpace(lookup(key)); v != "" {
		return v
	}
	return fallback
}
