package auth

import (
	"log/slog"
)

const testJWTSecret = "test-jwt-secret-0123456789abcdef-0123456789"

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&discardWriter{}, nil))
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
