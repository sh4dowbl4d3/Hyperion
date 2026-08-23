package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/auth"
)

const testSecret = "middleware-test-secret-0123456789abcdef"

func newAuthTestRouter(tokens *auth.TokenService, guarded func(c *gin.Context)) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	protected := r.Group("/")
	protected.Use(Authenticate(tokens))
	protected.GET("/protected", guarded)
	r.GET("/admin-only", Authenticate(tokens), RequireRole("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func issueToken(t *testing.T, svc *auth.TokenService, userID, role string) string {
	t.Helper()
	token, _, err := svc.Issue(userID, role)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return token
}

func TestAuthenticateAllowsValidToken(t *testing.T) {
	tokens := auth.NewTokenService(testSecret, time.Hour)
	var gotID, gotRole string
	r := newAuthTestRouter(tokens, func(c *gin.Context) {
		gotID, _ = UserIDFrom(c)
		v, _ := c.Get(ContextUserRole)
		gotRole, _ = v.(string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+issueToken(t, tokens, "u-1", "user"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if gotID != "u-1" || gotRole != "user" {
		t.Errorf("context identity = (%q, %q), want (u-1, user)", gotID, gotRole)
	}
}

func requestWithToken(r *gin.Engine, token string) int {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec.Code
}

func errorBody(t *testing.T, rec *httptest.ResponseRecorder) (int, string) {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid error envelope: %v", err)
	}
	return rec.Code, body.Error.Code
}

func TestAuthenticateRejectionMatrix(t *testing.T) {
	tokens := auth.NewTokenService(testSecret, time.Hour)
	r := newAuthTestRouter(tokens, func(c *gin.Context) { c.Status(http.StatusOK) })

	cases := []struct {
		name      string
		authValue string
		wantCode  int
		wantErr   string
	}{
		{"no header", "", 401, "missing_token"},
		{"wrong scheme", "Basic dXNlcjpwYXNz", 401, "missing_token"},
		{"empty bearer", "Bearer ", 401, "missing_token"},
		{"garbage token", "Bearer not-a-jwt", 401, "invalid_token"},
		{"expired token", "Bearer " + mustExpiredToken(t, tokens), 401, "token_expired"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.authValue != "" {
				req.Header.Set("Authorization", tc.authValue)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			code, errCode := errorBody(t, rec)
			if code != tc.wantCode || errCode != tc.wantErr {
				t.Fatalf("got (%d, %s), want (%d, %s)", code, errCode, tc.wantCode, tc.wantErr)
			}
		})
	}
}

func mustExpiredToken(t *testing.T, valid *auth.TokenService) string {
	t.Helper()
	expired := auth.NewTokenService(testSecret, -time.Minute)
	token, _, err := expired.Issue("u-1", "user")
	if err != nil {
		t.Fatalf("issue expired token: %v", err)
	}
	if _, err := valid.Verify(token); err == nil {
		t.Fatal("test setup: token should not verify as valid")
	}
	return token
}

func TestRequireRoleRejectsWrongRole(t *testing.T) {
	tokens := auth.NewTokenService(testSecret, time.Hour)
	r := newAuthTestRouter(tokens, func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+issueToken(t, tokens, "u-1", "user"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	code, errCode := errorBody(t, rec)
	if code != http.StatusForbidden || errCode != "forbidden" {
		t.Fatalf("got (%d, %s), want (403, forbidden)", code, errCode)
	}
}

func TestRequireRoleAllowsMatchingRole(t *testing.T) {
	tokens := auth.NewTokenService(testSecret, time.Hour)
	r := newAuthTestRouter(tokens, func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+issueToken(t, tokens, "a-1", "admin"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want admin access granted", rec.Code)
	}
}
