package race

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/labs"
	"hyperion/backend/internal/middleware"
)

type fakeRedeemer struct {
	vulnErr  error
	safeErr  error
	status   *Status
	statusErr error
	vulnCalls int64
	safeCalls int64
}

func (f *fakeRedeemer) RedeemVulnerable(context.Context, string, string) error {
	atomic.AddInt64(&f.vulnCalls, 1)
	return f.vulnErr
}

func (f *fakeRedeemer) RedeemSafe(context.Context, string, string) error {
	atomic.AddInt64(&f.safeCalls, 1)
	return f.safeErr
}

func (f *fakeRedeemer) StatusForUser(context.Context, string, string) (*Status, error) {
	if f.statusErr != nil {
		return nil, f.statusErr
	}
	return f.status, nil
}

type fakeCompleter struct {
	calls int32
	slug  string
	err   error
}

func (f *fakeCompleter) CompleteLab(_ context.Context, _, labSlug string) (*labs.CompletionResult, error) {
	atomic.AddInt32(&f.calls, 1)
	f.slug = labSlug
	if f.err != nil {
		return nil, f.err
	}
	return &labs.CompletionResult{NewlyCompleted: true}, nil
}

func newTestLab(t *testing.T, store Redeemer, completer Completer) *gin.Engine {
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

func doPost(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRedeemHappyPath(t *testing.T) {
	store := &fakeRedeemer{status: &Status{Code: CouponCode, MaxRedemptions: 5, Redeemed: 6, Mine: true}}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	w := doPost(r, "/targets/race-condition/redeem", `{"code":"LAUNCH-2026"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
}

func TestExhaustedCouponConflicts(t *testing.T) {
	store := &fakeRedeemer{vulnErr: ErrExhausted}
	r := newTestLab(t, store, nil)

	w := doPost(r, "/targets/race-condition/redeem", `{"code":"LAUNCH-2026"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
}

func TestUnknownCouponNotFound(t *testing.T) {
	store := &fakeRedeemer{vulnErr: ErrCouponNotFound}
	r := newTestLab(t, store, nil)

	w := doPost(r, "/targets/race-condition/redeem", `{"code":"NOPE"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestMissingCodeRejected(t *testing.T) {
	r := newTestLab(t, &fakeRedeemer{}, nil)

	for _, body := range []string{`{}`, `{"code":""}`, `not-json`} {
		w := doPost(r, "/targets/race-condition/redeem", body)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", body, w.Code)
		}
	}
}

func TestSafeTwinRoutesToTransactionalPath(t *testing.T) {
	store := &fakeRedeemer{}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	doPost(r, "/targets/race-condition/redeem-safe", `{"code":"LAUNCH-2026"}`)
	if store.safeCalls != 1 || store.vulnCalls != 0 {
		t.Errorf("safe endpoint calls safe=%d vuln=%d", store.safeCalls, store.vulnCalls)
	}
}

// overRedeemedStatus is the state after attackers won the race.
func overRedeemedStatus() *Status {
	return &Status{Code: CouponCode, MaxRedemptions: 5, Redeemed: 9, Mine: true}
}

func TestCompletionAwardedWhenOverRedeemed(t *testing.T) {
	store := &fakeRedeemer{status: overRedeemedStatus()}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	doPost(r, "/targets/race-condition/redeem", `{"code":"LAUNCH-2026"}`)
	if completer.calls != 1 {
		t.Fatalf("completion calls = %d, want 1", completer.calls)
	}
	if completer.slug != LabSlug {
		t.Errorf("slug = %q, want %q", completer.slug, LabSlug)
	}
}

func TestCompletionNotAwardedWithoutOverRedemption(t *testing.T) {
	store := &fakeRedeemer{status: &Status{Code: CouponCode, MaxRedemptions: 5, Redeemed: 3, Mine: true}}
	completer := &fakeCompleter{}
	r := newTestLab(t, store, completer)

	doPost(r, "/targets/race-condition/redeem", `{"code":"LAUNCH-2026"}`)
	if completer.calls != 0 {
		t.Errorf("completion awarded at normal load, calls = %d", completer.calls)
	}
}

func TestCompletionErrorDoesNotFailRequest(t *testing.T) {
	store := &fakeRedeemer{status: overRedeemedStatus()}
	completer := &fakeCompleter{err: errors.New("db down")}
	r := newTestLab(t, store, completer)

	w := doPost(r, "/targets/race-condition/redeem", `{"code":"LAUNCH-2026"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 despite completion error", w.Code)
	}
}

func TestOverRedeamedDetector(t *testing.T) {
	if OverRedeemed(&Status{MaxRedemptions: 5, Redeemed: 5}) {
		t.Error("at-cap is not over-redemption")
	}
	if !OverRedeemed(&Status{MaxRedemptions: 5, Redeemed: 6}) {
		t.Error("beyond-cap must count as over-redemption")
	}
}

func TestMetaIsValid(t *testing.T) {
	r := labs.NewRegistry()
	lab := NewLab(&fakeRedeemer{}, nil)
	if err := r.Register(stubMetaLab{lab.Meta()}); err != nil {
		t.Fatalf("meta invalid: %v", err)
	}
}

type stubMetaLab struct{ meta labs.Meta }

func (s stubMetaLab) Meta() labs.Meta                     { return s.meta }
func (stubMetaLab) RegisterRoutes(*gin.RouterGroup) error { return nil }
