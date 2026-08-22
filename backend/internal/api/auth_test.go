package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"moderndvwa/backend/internal/auth"
	"moderndvwa/backend/internal/users"
)

const endpointTestSecret = "endpoint-test-secret-0123456789abcdef-0123"

type fakeAuthStore struct {
	users map[string]*users.User
}

func newFakeAuthStore() *fakeAuthStore { return &fakeAuthStore{users: map[string]*users.User{}} }

func (f *fakeAuthStore) Create(_ context.Context, email, passwordHash, role string) (*users.User, error) {
	if _, exists := f.users[email]; exists {
		return nil, users.ErrEmailTaken
	}
	u := &users.User{ID: "id-" + email, Email: email, PasswordHash: passwordHash, Role: role}
	f.users[email] = u
	return u, nil
}

func (f *fakeAuthStore) ByEmail(_ context.Context, email string) (*users.User, error) {
	if u, ok := f.users[email]; ok {
		return u, nil
	}
	return nil, users.ErrNotFound
}

func (f *fakeAuthStore) ByID(_ context.Context, id string) (*users.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, users.ErrNotFound
}

type authTestFixture struct {
	router *gin.Engine
	tokens *auth.TokenService
}

func newAuthFixture() *authTestFixture {
	gin.SetMode(gin.ReleaseMode)
	store := newFakeAuthStore()
	log := slog.New(slog.NewTextHandler(&discardAPIWriter{}, nil))
	svc := auth.NewService(store, log)
	tokens := auth.NewTokenService(endpointTestSecret, time.Hour)

	router := NewRouter(Deps{
		Log:    log,
		DB:     nil,
		Auth:   svc,
		Tokens: tokens,
		Users:  store,
	})
	return &authTestFixture{router: router, tokens: tokens}
}

type discardAPIWriter struct{}

func (discardAPIWriter) Write(p []byte) (int, error) { return len(p), nil }

func doJSON(r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = strings.NewReader(string(raw))
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON response %q: %v", rec.Body.String(), err)
	}
	return m
}

func TestRegisterEndpointCreatesAccount(t *testing.T) {
	fx := newAuthFixture()

	rec := doJSON(fx.router, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"email":    "student@example.com",
		"password": "long-enough-password",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	body := decode(t, rec)
	user, _ := body["user"].(map[string]any)
	if user == nil {
		t.Fatal("response missing user object")
	}
	if user["id"] == "" || user["email"] != "student@example.com" || user["role"] != "user" {
		t.Errorf("unexpected user payload: %v", user)
	}
	if _, leaks := user["password_hash"]; leaks {
		t.Error("user payload must never include password_hash")
	}
}

func TestRegisterEndpointValidationErrors(t *testing.T) {
	fx := newAuthFixture()

	cases := []struct {
		name        string
		payload     map[string]string
		wantStatus  int
		wantErrCode string
	}{
		{"invalid email", map[string]string{"email": "nope", "password": "long-enough-password"}, 400, "bad_request"},
		{"weak password", map[string]string{"email": "a@example.com", "password": "tiny"}, 400, "bad_request"},
		{"missing fields", map[string]string{}, 400, "bad_request"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(fx.router, http.MethodPost, "/api/v1/auth/register", "", tc.payload)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			body := decode(t, rec)
			errObj, _ := body["error"].(map[string]any)
			if errObj == nil || errObj["code"] != tc.wantErrCode {
				t.Errorf("expected error.code %q in %v", tc.wantErrCode, body)
			}
		})
	}
}

func TestRegisterEndpointDuplicateEmailConflicts(t *testing.T) {
	fx := newAuthFixture()

	first := doJSON(fx.router, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"email":    "dup@example.com",
		"password": "long-enough-password",
	})
	if first.Code != http.StatusCreated {
		t.Fatalf("first register status = %d, want 201", first.Code)
	}

	dup := doJSON(fx.router, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"email":    "DUP@example.com",
		"password": "another-long-password",
	})
	if dup.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d, want 409", dup.Code)
	}
	body := decode(t, dup)
	errObj, _ := body["error"].(map[string]any)
	if errObj == nil || errObj["code"] != "email_taken" {
		t.Errorf("expected email_taken code, got %v", body)
	}
}

func TestLoginEndpointIssuesWorkingToken(t *testing.T) {
	fx := newAuthFixture()

	doJSON(fx.router, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"email":    "login-flow@example.com",
		"password": "long-enough-password",
	})

	rec := doJSON(fx.router, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"email":    "login-flow@example.com",
		"password": "long-enough-password",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	body := decode(t, rec)
	token, _ := body["token"].(string)
	if token == "" {
		t.Fatal("login response missing token")
	}
	claims, err := fx.tokens.Verify(token)
	if err != nil {
		t.Fatalf("issued token does not verify: %v", err)
	}
	if claims.UserID != "id-login-flow@example.com" || claims.Role != "user" {
		t.Errorf("token claims = %+v, want user id and role user", claims)
	}

	me := doJSON(fx.router, http.MethodGet, "/api/v1/auth/me", token, nil)
	if me.Code != http.StatusOK {
		t.Fatalf("/me status = %d, want 200", me.Code)
	}
	meBody := decode(t, me)
	user, _ := meBody["user"].(map[string]any)
	if user == nil || user["email"] != "login-flow@example.com" {
		t.Errorf("/me returned unexpected payload: %v", meBody)
	}
}

func TestLoginEndpointGenericFailure(t *testing.T) {
	fx := newAuthFixture()
	doJSON(fx.router, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"email":    "exists@example.com",
		"password": "long-enough-password",
	})

	for name, creds := range map[string][2]string{
		"unknown user": {"ghost@example.com", "whatever-long"},
		"wrong pw":     {"exists@example.com", "not-the-right-one"},
	} {
		t.Run(name, func(t *testing.T) {
			rec := doJSON(fx.router, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
				"email": creds[0], "password": creds[1],
			})
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
			body := decode(t, rec)
			msg, _ := body["error"].(map[string]any)["message"].(string)
			if msg != "invalid email or password" {
				t.Errorf("message = %q, want identical generic text for both failure kinds", msg)
			}
		})
	}
}

func TestMeEndpointRequiresAuthentication(t *testing.T) {
	fx := newAuthFixture()

	rec := doJSON(fx.router, http.MethodGet, "/api/v1/auth/me", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestAdminPingEnforcesRole(t *testing.T) {
	fx := newAuthFixture()

	adminToken, _, err := fx.tokens.Issue("admin-id", "admin")
	if err != nil {
		t.Fatalf("issue admin token: %v", err)
	}
	userToken, _, err := fx.tokens.Issue("user-id", "user")
	if err != nil {
		t.Fatalf("issue user token: %v", err)
	}

	ok := doJSON(fx.router, http.MethodGet, "/api/v1/admin/ping", adminToken, nil)
	if ok.Code != http.StatusOK {
		t.Fatalf("admin ping status = %d, want 200", ok.Code)
	}

	denied := doJSON(fx.router, http.MethodGet, "/api/v1/admin/ping", userToken, nil)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("user ping status = %d, want 403", denied.Code)
	}

	anon := doJSON(fx.router, http.MethodGet, "/api/v1/admin/ping", "", nil)
	if anon.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous ping status = %d, want 401", anon.Code)
	}
}
