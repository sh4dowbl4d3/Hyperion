package redirect

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"moderndvwa/backend/internal/labs"
	"moderndvwa/backend/internal/middleware"
)

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

func newTestLab(t *testing.T, completer Completer) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/targets/" + LabSlug)
	group.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, "user-1")
		c.Next()
	})
	if err := NewLab(completer).RegisterRoutes(group); err != nil {
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

func TestVulnerableAcceptsExternalHost(t *testing.T) {
	completer := &fakeCompleter{}
	r := newTestLab(t, completer)

	w := doGet(r, "/targets/open-redirect/redirect?to=https://evil.example.com/phish")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (vulnerability)", w.Code)
	}
	if completer.calls != 0 {
		t.Errorf("completion awarded for non-flag target")
	}
}

func TestCompletionAwardedForAdminConsoleTarget(t *testing.T) {
	completer := &fakeCompleter{}
	r := newTestLab(t, completer)

	w := doGet(r, "/targets/open-redirect/redirect?to=http://10.0.0.9"+FlagTargetPath)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if completer.calls != 1 || completer.slug != LabSlug {
		t.Errorf("completion calls=%d slug=%q", completer.calls, completer.slug)
	}
}

func TestSafeTwinBlocksAbsoluteURLs(t *testing.T) {
	blocked := []string{
		"https://evil.example.com",
		"http://10.0.0.9/admin/secret-token",
		"//evil.example.com",
		"javascript:alert(1)",
	}
	for _, to := range blocked {
		r := newTestLab(t, &fakeCompleter{})
		w := doGet(r, "/targets/open-redirect/redirect-safe?to="+to)
		if w.Code != http.StatusBadRequest {
			t.Errorf("to=%q: status = %d, want 400", to, w.Code)
		}
	}
}

func TestSafeTwinAllowsInternalSlugs(t *testing.T) {
	r := newTestLab(t, &fakeCompleter{})
	for _, to := range []string{"/promo", "/newsletter"} {
		w := doGet(r, "/targets/open-redirect/redirect-safe?to="+to)
		if w.Code != http.StatusOK {
			t.Errorf("to=%q: status = %d, want 200", to, w.Code)
		}
	}
}

func TestMissingToRejectedOnBothTwins(t *testing.T) {
	for _, path := range []string{"/targets/open-redirect/redirect", "/targets/open-redirect/redirect-safe"} {
		r := newTestLab(t, &fakeCompleter{})
		w := doGet(r, path)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", path, w.Code)
		}
	}
}

func TestCampaignListingWorks(t *testing.T) {
	r := newTestLab(t, &fakeCompleter{})
	w := doGet(r, "/targets/open-redirect/campaigns")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "promo") {
		t.Errorf("campaigns listing failed: %d %s", w.Code, w.Body.String())
	}
}

func TestMetaIsValid(t *testing.T) {
	r := labs.NewRegistry()
	lab := NewLab(&fakeCompleter{})
	if err := r.Register(stubMetaLab{lab.Meta()}); err != nil {
		t.Fatalf("meta invalid: %v", err)
	}
}

type stubMetaLab struct{ meta labs.Meta }

func (s stubMetaLab) Meta() labs.Meta                     { return s.meta }
func (stubMetaLab) RegisterRoutes(*gin.RouterGroup) error { return nil }

var _ = errors.New // keep imports stable
