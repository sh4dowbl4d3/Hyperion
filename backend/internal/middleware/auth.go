package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/auth"
	"hyperion/backend/internal/httpx"
)

const (
	ContextUserIDKey = "auth_user_id"
	ContextUserRole  = "auth_user_role"
	bearerPrefix     = "Bearer "
)

func Authenticate(tokens *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, bearerPrefix) {
			httpx.Error(c, http.StatusUnauthorized, "missing_token", "authorization header must contain a bearer token")
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
		if raw == "" {
			httpx.Error(c, http.StatusUnauthorized, "missing_token", "bearer token is empty")
			return
		}

		claims, err := tokens.Verify(raw)
		if errors.Is(err, auth.ErrTokenExpired) {
			httpx.Error(c, http.StatusUnauthorized, "token_expired", "access token has expired")
			return
		}
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, "invalid_token", "access token is invalid")
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserRole, claims.Role)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get(ContextUserRole)
		if !exists {
			httpx.Error(c, http.StatusForbidden, "forbidden", "authentication required for role check")
			return
		}
		if value != role {
			httpx.Forbidden(c, "insufficient permissions")
			return
		}
		c.Next()
	}
}

func UserIDFrom(c *gin.Context) (string, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}
