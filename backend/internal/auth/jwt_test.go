package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueAndVerifyRoundTrip(t *testing.T) {
	svc := NewTokenService(testJWTSecret, 15*time.Minute)

	signed, expiresAt, err := svc.Issue("user-123", "admin")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if !expiresAt.After(time.Now()) {
		t.Errorf("expiresAt = %v, want in the future", expiresAt)
	}

	claims, err := svc.Verify(signed)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want user-123", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %q, want admin", claims.Role)
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	expired := NewTokenService(testJWTSecret, -time.Minute)

	signed, _, err := expired.Issue("user-123", "user")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	validSvc := NewTokenService(testJWTSecret, time.Hour)
	_, err = validSvc.Verify(signed)
	if err == nil {
		t.Fatal("Verify accepted an expired token")
	}
	if err != ErrTokenExpired {
		t.Fatalf("error = %v, want ErrTokenExpired (callers must distinguish expiry)", err)
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	issuer := NewTokenService("secret-for-signing-only-0123456789abcdef", time.Hour)
	verifier := NewTokenService("a-completely-different-secret-0123456789", time.Hour)

	signed, _, err := issuer.Issue("user-123", "user")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := verifier.Verify(signed); err != ErrTokenInvalid {
		t.Fatalf("error = %v, want ErrTokenInvalid", err)
	}
}

func TestVerifyRejectsTamperedPayload(t *testing.T) {
	svc := NewTokenService(testJWTSecret, time.Hour)

	signed, _, err := svc.Issue("user-123", "user")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	parts := strings.Split(signed, ".")
	forged, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, registeredClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			Issuer:    tokenIssuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Role: "admin",
	}).SignedString([]byte("attacker-controlled-secret-not-the-servers"))
	if forged == "" || len(parts) != 3 {
		t.Fatalf("test setup broken: forged=%q parts=%d", forged, len(parts))
	}
	if _, err := svc.Verify(forged); err != ErrTokenInvalid {
		t.Fatalf("error = %v, want ErrTokenInvalid for token signed with attacker secret", err)
	}
}

func TestVerifyRejectsUnexpectedAlgorithm(t *testing.T) {
	svc := NewTokenService(testJWTSecret, time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, registeredClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			Issuer:    tokenIssuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Role: "user",
	})
	hs512Token, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign HS512: %v", err)
	}

	_, err = svc.Verify(hs512Token)
	if err == nil {
		t.Fatal("Verify accepted HS512; algorithm allowlist is not enforced")
	}
	if err != ErrTokenInvalid {
		t.Fatalf("error = %v, want ErrTokenInvalid", err)
	}
}

func TestVerifyRejectsMissingExpiration(t *testing.T) {
	svc := NewTokenService(testJWTSecret, time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, registeredClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  "user-123",
			Issuer:   tokenIssuer,
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
		Role: "user",
	})
	noExp, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := svc.Verify(noExp); err != ErrTokenInvalid {
		t.Fatalf("error = %v, want ErrTokenInvalid for token without exp claim", err)
	}
}

func TestVerifyRejectsMissingRoleOrSubject(t *testing.T) {
	svc := NewTokenService(testJWTSecret, time.Hour)

	make := func(sub, role string) string {
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, registeredClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   sub,
				Issuer:    tokenIssuer,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
			Role: role,
		}).SignedString([]byte(testJWTSecret))
		if err != nil {
			t.Fatalf("sign: %v", err)
		}
		return token
	}

	for name, tok := range map[string]string{
		"empty subject": make("", "user"),
		"empty role":    make("user-123", ""),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := svc.Verify(tok); err != ErrTokenInvalid {
				t.Fatalf("error = %v, want ErrTokenInvalid", err)
			}
		})
	}
}

func TestVerifyRejectsGarbageInput(t *testing.T) {
	svc := NewTokenService(testJWTSecret, time.Hour)

	for name, input := range map[string]string{
		"empty":            "",
		"not a jwt":        "garbage",
		"two segments":     "aaa.bbb",
		"broken signature": "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ4In0.not-base64!!",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := svc.Verify(input); err != ErrTokenInvalid && err != ErrTokenExpired {
				t.Fatalf("error = %v, want ErrTokenInvalid", err)
			}
		})
	}
}
