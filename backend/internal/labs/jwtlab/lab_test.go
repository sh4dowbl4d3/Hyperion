package jwtlab

import (
	"context"
	"errors"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func newTestLab(t *testing.T, completer Completer, now func() time.Time) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/targets/" + LabSlug)
	group.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, "user-1")
		c.Next()
	})
	lab := NewLab(completer)
	if now != nil {
		lab.now = now
	}
	if err := lab.RegisterRoutes(group); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	return r
}

func doGet(r *gin.Engine, path, badge string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if badge != "" {
		req.Header.Set("X-Lab-Badge", "Bearer "+badge)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// craftToken builds a token from explicit header/claims JSON and a signature
// produced with the given key (nil = no signature).
func craftToken(t *testing.T, headerJSON, claimsJSON string, key []byte) string {
	t.Helper()
	h := base64RawURLEncode([]byte(headerJSON))
	p := base64RawURLEncode([]byte(claimsJSON))
	sig := ""
	if key != nil {
		sig = base64RawURLEncode(hmacSHA256([]byte(h+"."+p), key))
	} else {
		sig = base64RawURLEncode([]byte("irrelevant"))
	}
	return h + "." + p + "." + sig
}

func adminClaimsJSON(t *testing.T, exp int64) string {
	t.Helper()
	b, _ := json.Marshal(Claims{Sub: "badge-user-1", Role: "admin", Iss: IssuerLab, Exp: exp})
	return string(b)
}

const future = int64(4102444800) // 2100-01-01

// --- vulnerable decoder behaviour ---

func TestVulnerableDecoderAcceptsAlgNone(t *testing.T) {
	token := craftToken(t,
		`{"alg":"none","typ":"JWT"}`,
		adminClaimsJSON(t, future),
		nil,
	)
	claims, err := DecodeVulnerable(token)
	if err != nil {
		t.Fatalf("DecodeVulnerable rejected alg:none token: %v", err)
	}
	if claims.Role != "admin" {
		t.Errorf("role = %q, want admin", claims.Role)
	}
}

func TestVulnerableDecoderAcceptsWeakSecretSignature(t *testing.T) {
	token := craftToken(t,
		`{"alg":"HS256","typ":"JWT"}`,
		adminClaimsJSON(t, future),
		[]byte(WeakSecret),
	)
	claims, err := DecodeVulnerable(token)
	if err != nil {
		t.Fatalf("DecodeVulnerable rejected weak-secret forgery: %v", err)
	}
	if claims.Role != "admin" {
		t.Errorf("role = %q, want admin", claims.Role)
	}
}

func TestVulnerableDecoderRejectsWrongSignature(t *testing.T) {
	token := craftToken(t,
		`{"alg":"HS256","typ":"JWT"}`,
		adminClaimsJSON(t, future),
		[]byte("not-the-secret"),
	)
	if _, err := DecodeVulnerable(token); err == nil {
		t.Error("decoder accepted a token signed with the wrong secret")
	}
}

// --- secure verifier behaviour ---

func TestSecureVerifierRejectsAlgNone(t *testing.T) {
	token := craftToken(t,
		`{"alg":"none","typ":"JWT"}`,
		adminClaimsJSON(t, future),
		nil,
	)
	v := &SecureVerifier{Secret: StrongSecret, Issuer: IssuerLab}
	if _, err := v.Verify(token); err == nil {
		t.Error("secure verifier accepted alg:none token")
	}
}

func TestSecureVerifierRejectsWeakSecretForgery(t *testing.T) {
	token := craftToken(t,
		`{"alg":"HS256","typ":"JWT"}`,
		adminClaimsJSON(t, future),
		[]byte(WeakSecret),
	)
	v := &SecureVerifier{Secret: StrongSecret, Issuer: IssuerLab}
	if _, err := v.Verify(token); err == nil {
		t.Error("secure verifier accepted token signed with the lab's weak secret")
	}
}

func TestSecureVerifierAcceptsOnlyGenuineStrongTokens(t *testing.T) {
	// A correctly signed strong-secret admin token would pass — but no such
	// token is ever issued to learners; verify issuance path stays 'user'.
	guest, err := IssueVulnerable("badge-x", time.Unix(future-1000, 0))
	if err != nil {
		t.Fatalf("issue guest: %v", err)
	}
	if !strings.Contains(guest, ".") {
		t.Error("guest token malformed")
	}
	v := &SecureVerifier{Secret: StrongSecret, Issuer: IssuerLab}
	if _, err := v.Verify(guest); err == nil {
		t.Error("weak-signed guest token must not satisfy the secure verifier")
	}
}

func TestSecureVerifierEnforcesIssuerAndExpiry(t *testing.T) {
	wrongIss, _ := json.Marshal(Claims{Sub: "x", Role: "admin", Iss: "other", Exp: future})
	tokenIss := signWithStrong(t, string(wrongIss))
	v := &SecureVerifier{Secret: StrongSecret, Issuer: IssuerLab}
	if _, err := v.Verify(tokenIss); err == nil {
		t.Error("issuer claim not enforced")
	}

	expired, _ := json.Marshal(Claims{Sub: "x", Role: "admin", Iss: IssuerLab, Exp: 100})
	tokenExp := signWithStrong(t, string(expired))
	if _, err := v.Verify(tokenExp); !errors.Is(err, ErrExpired) {
		t.Errorf("expiry not enforced, got %v", err)
	}
}

func signWithStrong(t *testing.T, claimsJSON string) string {
	t.Helper()
	return craftToken(t, `{"alg":"HS256","typ":"JWT"}`, claimsJSON, []byte(StrongSecret))
}

// --- HTTP flow ---

func TestGuestTokenIssuesUserBadge(t *testing.T) {
	r := newTestLab(t, nil, nil)

	w := doGet(r, "/targets/jwt/guest-token", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	claims, err := DecodeVulnerable(body.Token)
	if err != nil {
		t.Fatalf("issued token undecodable: %v", err)
	}
	if claims.Role != "user" {
		t.Errorf("issued role = %q, want user", claims.Role)
	}
}

func TestAdminPanelRequiresRole(t *testing.T) {
	r := newTestLab(t, nil, nil)

	gw := doGet(r, "/targets/jwt/guest-token", "")
	var body struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(gw.Body.Bytes(), &body)

	w := doGet(r, "/targets/jwt/admin-panel", body.Token)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 for guest role", w.Code)
	}
}

func TestForgedAdminBadgeUnlocksPanelAndCompletesLab(t *testing.T) {
	completer := &fakeCompleter{}
	r := newTestLab(t, completer, nil)

	forged := craftToken(t,
		`{"alg":"none","typ":"JWT"}`,
		adminClaimsJSON(t, future),
		nil,
	)
	w := doGet(r, "/targets/jwt/admin-panel", forged)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for forged admin badge: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), FlagPayload) {
		t.Errorf("flag missing from panel response: %s", w.Body.String())
	}
	if completer.calls != 1 {
		t.Fatalf("completion calls = %d, want 1", completer.calls)
	}
	if completer.slug != LabSlug {
		t.Errorf("slug = %q, want %q", completer.slug, LabSlug)
	}
}

func TestSameForgeryRejectedBySafeTwinWithoutCompletion(t *testing.T) {
	completer := &fakeCompleter{}
	r := newTestLab(t, completer, nil)

	forged := craftToken(t,
		`{"alg":"HS256","typ":"JWT"}`,
		adminClaimsJSON(t, future),
		[]byte(WeakSecret),
	)
	w := doGet(r, "/targets/jwt/admin-panel-safe", forged)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 from safe panel", w.Code)
	}
	if completer.calls != 0 {
		t.Errorf("safe twin awarded completion, calls = %d", completer.calls)
	}
}

func TestMissingBadgeRejected(t *testing.T) {
	r := newTestLab(t, nil, nil)
	w := doGet(r, "/targets/jwt/admin-panel", "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestPlatformIdentityIsolatedFromLabTokens(t *testing.T) {
	// Lab subjects live in their own namespace.
	sub := labSubject("03de45e8-5533")
	if !strings.HasPrefix(sub, "badge-") {
		t.Errorf("lab subject = %q, want badge- prefix", sub)
	}
}

func TestMetaIsValid(t *testing.T) {
	r := labs.NewRegistry()
	if err := r.Register(stubMetaLab{NewLab(nil).Meta()}); err != nil {
		t.Fatalf("meta invalid: %v", err)
	}
}

type stubMetaLab struct{ meta labs.Meta }

func (s stubMetaLab) Meta() labs.Meta                     { return s.meta }
func (stubMetaLab) RegisterRoutes(*gin.RouterGroup) error { return nil }
