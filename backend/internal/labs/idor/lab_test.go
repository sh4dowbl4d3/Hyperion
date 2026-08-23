package idor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"moderndvwa/backend/internal/labs"
	"moderndvwa/backend/internal/middleware"
)

const (
	userA = "11111111-1111-1111-1111-111111111111"
	userB = "22222222-2222-2222-2222-222222222222"
)

type fakeStore struct {
	docs  []Document
	getErr error
}

func (f *fakeStore) ListForUser(_ context.Context, userID string) ([]Document, error) {
	out := []Document{}
	for _, d := range f.docs {
		if d.OwnerID == userID {
			out = append(out, d)
		}
	}
	return out, nil
}

func (f *fakeStore) GetByID(_ context.Context, id int) (*Document, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	for i := range f.docs {
		if f.docs[i].ID == id {
			return &f.docs[i], nil
		}
	}
	return nil, fmt.Errorf("no doc %d", id)
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

func seedDocs() []Document {
	return []Document{
		{ID: 1, OwnerID: userA, Title: "My notes", Classification: "internal", Content: "grocery list"},
		{ID: 2, OwnerID: userA, Title: "Trip plan", Classification: "internal", Content: "tickets"},
		{ID: 3, OwnerID: userB, Title: "Q3 Access Review", Classification: "confidential", Content: FlagContent},
		{ID: 4, OwnerID: userB, Title: "B's notes", Classification: "internal", Content: "standup notes"},
	}
}

func newTestLab(t *testing.T, store DocumentStore, completer Completer) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/targets/" + LabSlug)
	group.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, userA)
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

func TestListReturnsOnlyOwnDocuments(t *testing.T) {
	r := newTestLab(t, &fakeStore{docs: seedDocs()}, nil)

	w := doGet(r, "/targets/idor/documents")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body struct {
		Documents []Document `json:"documents"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Documents) != 2 {
		t.Fatalf("got %d docs, want 2 (only user A's)", len(body.Documents))
	}
	for _, d := range body.Documents {
		if d.OwnerID != userA {
			t.Errorf("leaked doc %d owned by %s", d.ID, d.OwnerID)
		}
	}
}

func TestVulnerableGetReturnsForeignDocument(t *testing.T) {
	r := newTestLab(t, &fakeStore{docs: seedDocs()}, nil)

	w := doGet(r, "/targets/idor/documents/3") // user B's confidential doc
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (vulnerability)", w.Code)
	}
	if !strings.Contains(w.Body.String(), FlagContent) {
		t.Errorf("flag content not returned: %s", w.Body.String())
	}
}

func TestSafeGetRejectsForeignDocument(t *testing.T) {
	r := newTestLab(t, &fakeStore{docs: seedDocs()}, nil)

	w := doGet(r, "/targets/idor/documents-safe/3")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 for cross-tenant read", w.Code)
	}
	if strings.Contains(w.Body.String(), FlagContent) {
		t.Error("safe endpoint leaked foreign content")
	}
}

func TestSafeGetAllowsOwnDocument(t *testing.T) {
	r := newTestLab(t, &fakeStore{docs: seedDocs()}, nil)

	w := doGet(r, "/targets/idor/documents-safe/1")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for own document", w.Code)
	}
	if !strings.Contains(w.Body.String(), "grocery list") {
		t.Errorf("own document content missing: %s", w.Body.String())
	}
}

func TestUnknownDocumentIs404OnBothTwins(t *testing.T) {
	r := newTestLab(t, &fakeStore{docs: seedDocs()}, nil)

	for _, path := range []string{"/targets/idor/documents/999", "/targets/idor/documents-safe/999"} {
		w := doGet(r, path)
		if w.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 404", path, w.Code)
		}
	}
}

func TestStoreErrorIs404Not500(t *testing.T) {
	r := newTestLab(t, &fakeStore{getErr: errors.New("db exploded")}, nil)

	w := doGet(r, "/targets/idor/documents/1")
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (no existence oracle)", w.Code)
	}
}

func TestInvalidIDsRejected(t *testing.T) {
	r := newTestLab(t, &fakeStore{docs: seedDocs()}, nil)

	for _, path := range []string{
		"/targets/idor/documents/abc",
		"/targets/idor/documents/0",
		"/targets/idor/documents/-5",
	} {
		w := doGet(r, path)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", path, w.Code)
		}
	}
}

func TestCompletionAwardedForForeignConfidentialRead(t *testing.T) {
	store := &fakeStore{docs: seedDocs()}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	doGet(r, "/targets/idor/documents/3")
	if completer.calls != 1 {
		t.Fatalf("completion calls = %d, want 1", completer.calls)
	}
	if completer.slug != LabSlug {
		t.Errorf("slug = %q, want %q", completer.slug, LabSlug)
	}
}

func TestCompletionNotAwardedOnSafeTwin(t *testing.T) {
	store := &fakeStore{docs: seedDocs()}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	// Even if ownership were bypassed, the safe twin must never award.
	w := doGet(r, "/targets/idor/documents-safe/3")
	if w.Code == http.StatusOK && completer.calls != 0 {
		t.Error("safe twin awarded completion")
	}
	if completer.calls != 0 {
		t.Errorf("completion calls = %d, want 0", completer.calls)
	}
}

func TestCompletionNotAwardedForOwnOrNonConfidential(t *testing.T) {
	store := &fakeStore{docs: seedDocs()}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	doGet(r, "/targets/idor/documents/1") // own doc
	doGet(r, "/targets/idor/documents/4") // foreign but internal
	if completer.calls != 0 {
		t.Errorf("completion calls = %d, want 0", completer.calls)
	}
}

func TestCompletionErrorDoesNotFailRequest(t *testing.T) {
	store := &fakeStore{docs: seedDocs()}
	completer := &fakeCompleter{err: errors.New("db down")}
	r := newTestLab(t, store, completer)

	w := doGet(r, "/targets/idor/documents/3")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 despite completion error", w.Code)
	}
}

func TestMetaIsValid(t *testing.T) {
	lab := NewLab(&fakeStore{}, nil)
	r := labs.NewRegistry()
	if err := r.Register(stubMetaLab{lab.Meta()}); err != nil {
		t.Fatalf("meta invalid: %v", err)
	}
}

type stubMetaLab struct{ meta labs.Meta }

func (s stubMetaLab) Meta() labs.Meta                     { return s.meta }
func (stubMetaLab) RegisterRoutes(*gin.RouterGroup) error { return nil }
