package ssrf

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/labs"
	"hyperion/backend/internal/middleware"
)

type fakeFetcher struct {
	status  int
	body    string
	err     error
	lastURL string
}

func (f *fakeFetcher) Fetch(_ context.Context, rawURL string) (int, string, error) {
	f.lastURL = rawURL
	if f.err != nil {
		return 0, "", f.err
	}
	return f.status, f.body, nil
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

func newTestLab(t *testing.T, fetcher Fetcher, completer Completer) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/targets/" + LabSlug)
	group.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, "user-1")
		c.Next()
	})
	if err := NewLab(fetcher, completer, "http://127.0.0.1:9999").RegisterRoutes(group); err != nil {
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

func TestPreviewFetchesArbitraryURL(t *testing.T) {
	fetcher := &fakeFetcher{status: 200, body: `{"ok":true}`}
	r := newTestLab(t, fetcher, nil)

	w := doGet(r, "/targets/ssrf/preview?url=http://127.0.0.1:1/x")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(fetcher.lastURL, "127.0.0.1") {
		t.Errorf("loopback URL not passed to fetcher: %q", fetcher.lastURL)
	}
}

func TestConfigRevealsInternalBaseURL(t *testing.T) {
	r := newTestLab(t, &fakeFetcher{}, nil)

	w := doGet(r, "/targets/ssrf/config")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "http://127.0.0.1:9999") {
		t.Errorf("config missing internal base url: %s", w.Body.String())
	}
}

func TestCompletionAwardedForInternalFlagFetch(t *testing.T) {
	fetcher := &fakeFetcher{status: 200, body: "junk\n" + FlagContent + "\n"}
	completer := &fakeCompleter{}
	r := newTestLab(t, fetcher, completer)

	w := doGet(r, "/targets/ssrf/preview?url=http://127.0.0.1:1/secret/credentials.txt")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if completer.calls != 1 {
		t.Fatalf("completion calls = %d, want 1", completer.calls)
	}
	if completer.slug != LabSlug {
		t.Errorf("slug = %q, want %q", completer.slug, LabSlug)
	}
}

func TestNoCompletionWithoutFlag(t *testing.T) {
	fetcher := &fakeFetcher{status: 200, body: "harmless"}
	completer := &fakeCompleter{}
	r := newTestLab(t, fetcher, completer)

	doGet(r, "/targets/ssrf/preview?url=http://example.test/")
	if completer.calls != 0 {
		t.Errorf("completion calls = %d, want 0", completer.calls)
	}
}

func TestSafePreviewRejectsLoopbackAndPrivate(t *testing.T) {
	fetcher := &fakeFetcher{status: 200, body: FlagContent}
	completer := &fakeCompleter{}
	r := newTestLab(t, fetcher, completer)

	blocked := []string{
		"http://127.0.0.1/secret/credentials.txt",
		"http://localhost/secret",
		"http://[::1]/secret",
		"http://10.0.0.5/",
		"http://192.168.1.10/",
		"http://172.16.0.9/",
		"http://169.254.169.254/latest/meta-data/",
		"file:///etc/passwd",
		"gopher://host/",
		"http://svc.internal/",
		"http://metadata.google.internal/computeMetadata/v1/",
	}
	for _, u := range blocked {
		w := doGet(r, "/targets/ssrf/preview-safe?url="+u)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", u, w.Code)
		}
	}
	// The fetcher must never have been invoked.
	if fetcher.lastURL != "" {
		t.Errorf("safe endpoint fetched a blocked URL: %q", fetcher.lastURL)
	}
	if completer.calls != 0 {
		t.Errorf("completion awarded on safe twin: %d", completer.calls)
	}
}

func TestValidatePublicURLAllowsPublicHosts(t *testing.T) {
	allowed := []string{
		"https://example.com/page",
		"http://203.0.113.7/data.json",
		"https://training.hyperion.test/docs",
	}
	for _, u := range allowed {
		if err := ValidatePublicURL(u); err != nil {
			t.Errorf("%s: unexpectedly blocked: %v", u, err)
		}
	}
}

func TestMissingURLOnBothTwins(t *testing.T) {
	r := newTestLab(t, &fakeFetcher{}, nil)
	for _, path := range []string{"/targets/ssrf/preview", "/targets/ssrf/preview-safe"} {
		w := doGet(r, path)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", path, w.Code)
		}
	}
}

func TestFetchErrorIs502NotLeaky(t *testing.T) {
	fetcher := &fakeFetcher{err: errors.New("dial tcp 10.0.0.1:5432: connection refused")}
	r := newTestLab(t, fetcher, nil)

	w := doGet(r, "/targets/ssrf/preview?url=http://10.0.0.1/")
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", w.Code)
	}
	for _, leak := range []string{"dial", "10.0.0.1", "5432", "refused"} {
		if strings.Contains(w.Body.String(), leak) {
			t.Errorf("response leaked %q: %s", leak, w.Body.String())
		}
	}
}

func TestInternalServiceServesSyntheticData(t *testing.T) {
	ctx := context.Background()
	base, err := StartInternalService(ctx)
	if err != nil {
		t.Fatalf("start internal service: %v", err)
	}
	fetcher := NewHTTPFetcher()
	status, body, err := fetcher.Fetch(ctx, base+SecretPath)
	if err != nil || status != http.StatusOK {
		t.Fatalf("fetch secret: status=%d err=%v", status, err)
	}
	if !IsInternalFlag(body) {
		t.Errorf("flag missing from internal service response: %s", body)
	}
}

func TestRealSSRFRoundTripThroughVulnerableEndpoint(t *testing.T) {
	ctx := context.Background()
	base, err := StartInternalService(ctx)
	if err != nil {
		t.Fatalf("start internal service: %v", err)
	}
	completer := &fakeCompleter{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/targets/" + LabSlug)
	group.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, "user-1")
		c.Next()
	})
	if err := NewLab(NewHTTPFetcher(), completer, base).RegisterRoutes(group); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	// Safe twin refuses the loopback URL even against the real service.
	w := doGet(r, "/targets/ssrf/preview-safe?url="+base+SecretPath)
	if w.Code != http.StatusBadRequest {
		t.Errorf("safe twin fetched loopback URL, status = %d", w.Code)
	}

	// Vulnerable endpoint performs the SSRF and awards completion.
	w = doGet(r, "/targets/ssrf/preview?url="+base+SecretPath)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), FlagContent) {
		t.Errorf("ssrf round trip failed: %d %s", w.Code, w.Body.String())
	}
	if completer.calls != 1 {
		t.Errorf("completion calls = %d, want 1", completer.calls)
	}
}

func TestMetaIsValid(t *testing.T) {
	r := labs.NewRegistry()
	lab := NewLab(&fakeFetcher{}, nil, "http://127.0.0.1:1")
	if err := r.Register(stubMetaLab{lab.Meta()}); err != nil {
		t.Fatalf("meta invalid: %v", err)
	}
}

type stubMetaLab struct{ meta labs.Meta }

func (s stubMetaLab) Meta() labs.Meta                     { return s.meta }
func (stubMetaLab) RegisterRoutes(*gin.RouterGroup) error { return nil }
