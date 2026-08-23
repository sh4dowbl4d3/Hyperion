package sqli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"

	"moderndvwa/backend/internal/labs"
	"moderndvwa/backend/internal/middleware"
)

type fakeStore struct {
	vulnerable []Contact
	vulnErr    error
	safe       []Contact
	safeErr    error
	lastQuery  string
}

func (f *fakeStore) SearchVulnerable(_ context.Context, q string) ([]Contact, error) {
	f.lastQuery = q
	if f.vulnErr != nil {
		return nil, f.vulnErr
	}
	return f.vulnerable, nil
}

func (f *fakeStore) SearchSafe(_ context.Context, q string) ([]Contact, error) {
	f.lastQuery = q
	if f.safeErr != nil {
		return nil, f.safeErr
	}
	return f.safe, nil
}

type fakeCompleter struct {
	calls int
	slug  string
	err   error
}

func (f *fakeCompleter) CompleteLab(_ context.Context, _, labSlug string) (*labs.CompletionResult, error) {
	f.calls++
	f.slug = labSlug
	if f.err != nil {
		return nil, f.err
	}
	return &labs.CompletionResult{NewlyCompleted: true, XPAwarded: 100}, nil
}

func newTestLab(t *testing.T, store Store, completer Completer) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/targets/" + LabSlug)
	group.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, "user-1")
		c.Next()
	})
	if err := NewLab(store, completer).RegisterRoutes(group); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	return r
}

func doGet(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestVulnerableEndpointReturnsContacts(t *testing.T) {
	store := &fakeStore{vulnerable: []Contact{{Name: "Alice", Note: "note"}}}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	w := doGet(r, "/targets/sqli/search?q=ali")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body struct {
		Contacts []Contact `json:"contacts"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body.Contacts) != 1 || body.Contacts[0].Name != "Alice" {
		t.Errorf("unexpected contacts: %v", body.Contacts)
	}
	if store.lastQuery != "ali" {
		t.Errorf("query passed through = %q, want ali", store.lastQuery)
	}
	if completer.calls != 0 {
		t.Errorf("completion awarded without flag leak, calls = %d", completer.calls)
	}
}

func TestSafeEndpointIsParameterizedTwin(t *testing.T) {
	store := &fakeStore{safe: []Contact{{Name: "Carla"}}}
	r := newTestLab(t, store, nil)

	w := doGet(r, "/targets/sqli/search-safe?q=carla")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if store.lastQuery != "carla" {
		t.Errorf("safe endpoint must pass q as parameter, got %q", store.lastQuery)
	}
}

func TestEmptyQueryRejected(t *testing.T) {
	for _, path := range []string{"/targets/sqli/search?q=", "/targets/sqli/search"} {
		r := newTestLab(t, &fakeStore{}, nil)
		w := doGet(r, path)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", path, w.Code)
		}
	}
}

func TestOversizedQueryRejected(t *testing.T) {
	long := make([]byte, 201)
	for i := range long {
		long[i] = 'a'
	}
	r := newTestLab(t, &fakeStore{}, nil)
	w := doGet(r, "/targets/sqli/search?q="+string(long))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestSQLErrorNeverLeaksDriverDetail(t *testing.T) {
	store := &fakeStore{vulnErr: errors.New(`pq: syntax error at or near "'"`)}
	r := newTestLab(t, store, nil)

	w := doGet(r, "/targets/sqli/search?q="+url.QueryEscape(`' OR true --`))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	body := w.Body.String()
	for _, secret := range []string{"pq:", "syntax error"} {
		if contains(body, secret) {
			t.Errorf("response leaked driver detail %q: %s", secret, body)
		}
	}
}

func TestCompletionAwardedWhenFlagLeaked(t *testing.T) {
	store := &fakeStore{vulnerable: []Contact{
		{Name: "Archive Vault", Note: FlagNote},
	}}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	w := doGet(r, "/targets/sqli/search?q="+url.QueryEscape("x' OR true --"))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if completer.calls != 1 {
		t.Fatalf("completion calls = %d, want 1", completer.calls)
	}
	if completer.slug != LabSlug {
		t.Errorf("completed slug = %q, want %q", completer.slug, LabSlug)
	}
}

func TestCompletionNotTriggeredBySafeEndpoint(t *testing.T) {
	store := &fakeStore{safe: []Contact{{Name: "Archive Vault", Note: FlagNote}}}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	doGet(r, "/targets/sqli/search-safe?q=archive")
	if completer.calls != 0 {
		t.Errorf("safe twin must not award completion, calls = %d", completer.calls)
	}
}

func TestCompletionErrorDoesNotFailRequest(t *testing.T) {
	store := &fakeStore{vulnerable: []Contact{{Name: "Archive Vault", Note: FlagNote}}}
	completer := &fakeCompleter{err: errors.New("db down")}
	r := newTestLab(t, store, completer)

	w := doGet(r, "/targets/sqli/search?q=x")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 even when completion fails", w.Code)
	}
}

func TestBuildVulnerableQueryConcatenatesInput(t *testing.T) {
	got := BuildVulnerableQuery(`x' OR true --`)
	if !contains(got, `'%%`+`x' OR true --%%`) && !contains(got, "x' OR true --") {
		t.Errorf("query should embed raw input, got %q", got)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && stringsContains(haystack, needle)
}

func stringsContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
