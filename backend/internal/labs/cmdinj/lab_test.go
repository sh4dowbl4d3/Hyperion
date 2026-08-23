package cmdinj

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/labs"
	"hyperion/backend/internal/middleware"
)

type fakeCompleter struct {
	calls int
	slug  string
}

func (f *fakeCompleter) CompleteLab(_ context.Context, _, labSlug string) (*labs.CompletionResult, error) {
	f.calls++
	f.slug = labSlug
	return &labs.CompletionResult{NewlyCompleted: true}, nil
}

func newTestLab(t *testing.T, completer Completer) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/targets/" + LabSlug)
	group.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, "user-1")
		c.Next()
	})
	if err := NewLab(NewSimulator(), completer).RegisterRoutes(group); err != nil {
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

func TestVulnerableInjectionRunsSecondCommand(t *testing.T) {
	completer := &fakeCompleter{}
	r := newTestLab(t, completer)

	w := doGet(r, "/targets/command-injection/lookup?host="+url.QueryEscape("example.org; cat /etc/passwd"))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), FlagOutput) {
		t.Errorf("flag output missing from response: %s", w.Body.String())
	}
	if completer.calls != 1 || completer.slug != LabSlug {
		t.Errorf("completion calls=%d slug=%q", completer.calls, completer.slug)
	}
}

func TestVulnerableNormalLookupStillWorks(t *testing.T) {
	r := newTestLab(t, &fakeCompleter{})

	w := doGet(r, "/targets/command-injection/lookup?host=example.org")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if strings.Contains(w.Body.String(), FlagOutput) {
		t.Error("flag leaked without injection")
	}
}

func TestSafeTwinRejectsMetacharacters(t *testing.T) {
	blocked := []string{
		"example.org; cat /etc/passwd",
		"example.org | id",
		"example.org && whoami",
		"`whoami`",
		"$(id)",
		"example.org\nreboot",
	}
	for _, host := range blocked {
		r := newTestLab(t, &fakeCompleter{})
		w := doGet(r, "/targets/command-injection/lookup-safe?host="+url.QueryEscape(host))
		if w.Code != http.StatusBadRequest {
			t.Errorf("host=%q: status = %d, want 400", host, w.Code)
		}
	}
}

func TestSafeTwinAllowsPlainHostnames(t *testing.T) {
	r := newTestLab(t, &fakeCompleter{})
	for _, host := range []string{"example.org", "a.b.c.example-site.co"} {
		w := doGet(r, "/targets/command-injection/lookup-safe?host="+url.QueryEscape(host))
		if w.Code != http.StatusOK {
			t.Errorf("host=%q: status = %d, want 200", host, w.Code)
		}
	}
}

func TestMissingHostRejected(t *testing.T) {
	for _, path := range []string{"/targets/command-injection/lookup", "/targets/command-injection/lookup-safe"} {
		r := newTestLab(t, &fakeCompleter{})
		w := doGet(r, path)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", path, w.Code)
		}
	}
}

func TestMetaIsValid(t *testing.T) {
	r := labs.NewRegistry()
	lab := NewLab(NewSimulator(), &fakeCompleter{})
	if err := r.Register(stubMetaLab{lab.Meta()}); err != nil {
		t.Fatalf("meta invalid: %v", err)
	}
}

type stubMetaLab struct{ meta labs.Meta }

func (s stubMetaLab) Meta() labs.Meta                     { return s.meta }
func (stubMetaLab) RegisterRoutes(*gin.RouterGroup) error { return nil }
