package xss

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/labs"
	"hyperion/backend/internal/middleware"
)

type fakeStore struct {
	comments []Comment
	lastBody string
	listErr  error
	createErr error
}

func (f *fakeStore) ListVulnerable(context.Context) ([]Comment, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.comments, nil
}

func (f *fakeStore) CreateVulnerable(_ context.Context, author, body string) (*Comment, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.lastBody = body
	c := Comment{ID: len(f.comments) + 1, Author: author, Body: body, CreatedAt: time.Now()}
	f.comments = append([]Comment{c}, f.comments...)
	return &c, nil
}

func (f *fakeStore) ListSafe(ctx context.Context) ([]Comment, error) {
	out, err := f.ListVulnerable(ctx)
	if err != nil {
		return nil, err
	}
	for i := range out {
		r := escaped(out[i].Body)
		out[i].Rendered = &r
	}
	return out, nil
}

func (f *fakeStore) CreateSafe(ctx context.Context, author, body string) (*Comment, error) {
	return f.CreateVulnerable(ctx, author, SanitizeBody(body))
}

func escaped(s string) string {
	return escapeHTML(s)
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
	return &labs.CompletionResult{NewlyCompleted: true}, nil
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

func do(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

const payload = "<script>FLAG-XSS-77b1e4</script>"

func TestCreateStoresRawPayloadInVulnerablePath(t *testing.T) {
	store := &fakeStore{}
	r := newTestLab(t, store, nil)

	w := do(r, http.MethodPost, "/targets/xss/comments", `{"author":"attacker","body":"`+payload+`"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", w.Code, w.Body.String())
	}
	if store.lastBody != payload {
		t.Errorf("vulnerable path must store raw payload, got %q", store.lastBody)
	}
}

func TestFeedServesRawStoredMarkup(t *testing.T) {
	store := &fakeStore{comments: []Comment{{ID: 7, Author: "attacker", Body: payload}}}
	r := newTestLab(t, store, nil)

	w := do(r, http.MethodGet, "/targets/xss/comments/feed", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content type = %q, want text/html", ct)
	}
	if !strings.Contains(w.Body.String(), payload) {
		t.Errorf("vulnerable feed must emit raw stored markup, got: %s", w.Body.String())
	}
}

func TestSanitizeTwinStripsTagsBeforeStorage(t *testing.T) {
	store := &fakeStore{}
	r := newTestLab(t, store, nil)

	w := do(r, http.MethodPost, "/targets/xss/comments-safe", `{"author":"learner","body":"`+payload+` hi"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	if HasRawScript(store.lastBody) {
		t.Errorf("safe twin stored executable markup: %q", store.lastBody)
	}
	if !strings.Contains(store.lastBody, "FLAG-XSS-77b1e4") {
		t.Errorf("sanitizer should keep inner text, got %q", store.lastBody)
	}
}

func TestSafeFeedEscapesStoredBodies(t *testing.T) {
	// Even a pre-existing raw payload in storage is escaped on the safe feed.
	store := &fakeStore{comments: []Comment{{ID: 7, Author: "old", Body: payload}}}
	r := newTestLab(t, store, nil)

	w := do(r, http.MethodGet, "/targets/xss/comments-safe/feed", "")
	body := w.Body.String()
	if strings.Contains(body, payload) {
		t.Error("safe feed emitted raw script tag")
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Errorf("safe feed should escape markup, got: %s", body)
	}
}

func TestCompletionAwardedWhenMarkerStored(t *testing.T) {
	store := &fakeStore{}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	do(r, http.MethodPost, "/targets/xss/comments", `{"author":"a","body":"`+payload+`"}`)
	if completer.calls != 1 {
		t.Fatalf("completion calls = %d, want 1", completer.calls)
	}
	if completer.slug != LabSlug {
		t.Errorf("slug = %q, want %q", completer.slug, LabSlug)
	}
}

func TestCompletionNotAwardedForBenignComment(t *testing.T) {
	store := &fakeStore{}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	do(r, http.MethodPost, "/targets/xss/comments", `{"author":"a","body":"nice product"}`)
	if completer.calls != 0 {
		t.Errorf("benign comment awarded completion, calls = %d", completer.calls)
	}
}

func TestCompletionErrorDoesNotFailRequest(t *testing.T) {
	store := &fakeStore{}
	completer := &fakeCompleter{err: errors.New("db down")}
	r := newTestLab(t, store, completer)

	w := do(r, http.MethodPost, "/targets/xss/comments", `{"author":"a","body":"`+payload+`"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 despite completion error", w.Code)
	}
}

func TestValidationRejectsEmptyAndOversized(t *testing.T) {
	store := &fakeStore{}
	r := newTestLab(t, store, nil)

	cases := []struct {
		name string
		path string
		body string
	}{
		{"empty body", "/targets/xss/comments", `{"author":"a","body":""}`},
		{"empty author", "/targets/xss/comments", `{"author":"","body":"x"}`},
		{"missing fields", "/targets/xss/comments", `{"nope":1}`},
	}
	for _, tc := range cases {
		w := do(r, http.MethodPost, tc.path, tc.body)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", tc.name, w.Code)
		}
	}

	long := strings.Repeat("a", 2001)
	w := do(r, http.MethodPost, "/targets/xss/comments", `{"author":"a","body":"`+long+`"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("oversized body: status = %d, want 400", w.Code)
	}
}

func TestListEndpointReturnsJSON(t *testing.T) {
	store := &fakeStore{comments: []Comment{{ID: 1, Author: "m", Body: "hi"}}}
	r := newTestLab(t, store, nil)

	w := do(r, http.MethodGet, "/targets/xss/comments", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var parsed struct {
		Comments []Comment `json:"comments"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(parsed.Comments) != 1 || parsed.Comments[0].Author != "m" {
		t.Errorf("unexpected comments: %+v", parsed.Comments)
	}
}

func TestMetaIsValid(t *testing.T) {
	lab := NewLab(&fakeStore{}, nil)
	meta := lab.Meta()
	r := labs.NewRegistry()
	if err := r.Register(stubMetaLab{meta}); err != nil {
		t.Fatalf("meta invalid: %v", err)
	}
}

type stubMetaLab struct{ meta labs.Meta }

func (s stubMetaLab) Meta() labs.Meta                        { return s.meta }
func (stubMetaLab) RegisterRoutes(*gin.RouterGroup) error    { return nil }
