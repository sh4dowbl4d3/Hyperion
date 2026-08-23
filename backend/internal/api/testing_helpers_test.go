package api

import (
	"log/slog"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/auth"
)

func discardAPILogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(testWriter{}, nil))
}

func newTestTokens() *auth.TokenService {
	return auth.NewTokenService(endpointTestSecret, time.Hour)
}

func issueTestToken(t *testing.T, _ *gin.Engine) string {
	t.Helper()
	tokens := newTestTokens()
	token, _, err := tokens.Issue("user-1", "user")
	if err != nil {
		t.Fatalf("issue test token: %v", err)
	}
	return token
}
