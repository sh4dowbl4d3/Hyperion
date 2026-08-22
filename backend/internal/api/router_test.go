package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter(db DatabaseChecker) *gin.Engine {
	return NewRouter(Deps{Log: slog.New(slog.NewTextHandler(testWriter{}, nil)), DB: db})
}

type fakeDatabase struct {
	pingErr error
}

func (f *fakeDatabase) Ping(ctx context.Context) error { return f.pingErr }

type testWriter struct{}

func (testWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestHealthEndpoint(t *testing.T) {
	r := newTestRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Status  string `json:"status"`
		Service string `json:"service"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if body.Status != "ok" {
		t.Errorf("status = %q, want ok", body.Status)
	}
	if body.Service != ServiceName {
		t.Errorf("service = %q, want %q", body.Service, ServiceName)
	}
	if body.Version == "" {
		t.Error("version must not be empty")
	}
}

func TestUnknownRouteReturnsErrorEnvelope(t *testing.T) {
	r := newTestRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/does-not-exist", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if body.Error.Code != "not_found" {
		t.Errorf("error.code = %q, want not_found", body.Error.Code)
	}
}

func TestMethodNotAllowedReturnsErrorEnvelope(t *testing.T) {
	r := newTestRouter(nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/healthz", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}

	var body map[string]struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if body["error"].Code != "method_not_allowed" {
		t.Errorf("error.code = %q, want method_not_allowed", body["error"].Code)
	}
}

func TestRequestIDGeneratedAndEchoed(t *testing.T) {
	r := newTestRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-Id")
	if id == "" {
		t.Fatal("X-Request-Id header missing on response")
	}
}

func TestRequestIDPreservedFromClient(t *testing.T) {
	r := newTestRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	req.Header.Set("X-Request-Id", "trace-abc-123")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-Id"); got != "trace-abc-123" {
		t.Errorf("X-Request-Id = %q, want client-supplied value preserved", got)
	}
}

func readyzBody(t *testing.T, rec *httptest.ResponseRecorder) (int, map[string]any) {
	t.Helper()
	var body struct {
		Status string         `json:"status"`
		Checks map[string]any `json:"checks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	return rec.Code, body.Checks
}

func TestReadyzReportsOkWhenDatabaseHealthy(t *testing.T) {
	r := newTestRouter(&fakeDatabase{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/readyz", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	code, checks := readyzBody(t, rec)
	if code != http.StatusOK {
		t.Errorf("status = %d, want %d", code, http.StatusOK)
	}
	if checks["database"] != "ok" {
		t.Errorf("checks.database = %v, want ok", checks["database"])
	}
}

func TestReadyzUnavailableWhenDatabaseFails(t *testing.T) {
	r := newTestRouter(&fakeDatabase{pingErr: errors.New("connection refused")})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/readyz", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	code, checks := readyzBody(t, rec)
	if code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", code, http.StatusServiceUnavailable)
	}
	if checks["database"] != "error" {
		t.Errorf("checks.database = %v, want error", checks["database"])
	}
}

func TestReadyzUnavailableWithoutDatabase(t *testing.T) {
	r := newTestRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/readyz", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	code, checks := readyzBody(t, rec)
	if code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", code, http.StatusServiceUnavailable)
	}
	if checks["database"] != "not_configured" {
		t.Errorf("checks.database = %v, want not_configured", checks["database"])
	}
}
