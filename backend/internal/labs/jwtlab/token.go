package jwtlab

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// WeakSecret is intentionally short and guessable — it exists ONLY inside
// this lab's isolated flow. The platform auth service is untouched.
const (
	WeakSecret   = "secret"
	IssuerLab    = "moderndvwa-jwt-lab"
	FlagPayload  = "FLAG-JWT-5c2e91: forged admin token accepted"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	Sub   string `json:"sub"`
	Role  string `json:"role"`
	Iss   string `json:"iss"`
	Exp   int64  `json:"exp"`
}

// --- vulnerable implementation: naive split-and-trust decoder ---

// DecodeVulnerable parses a JWT without verifying the signature when the
// header says "none", and verifies with the weak secret otherwise. It also
// trusts claims without checking expiry strictly enough for production use.
// This mirrors real-world bugs where algorithm selection is attacker-controlled.
func DecodeVulnerable(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerJSON, err := b64Decode(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var header struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, ErrInvalidToken
	}

	if strings.EqualFold(header.Alg, "none") {
		// Flaw: trusts unsigned tokens because the client chose the algorithm.
		return parseClaims(parts[1])
	}

	sig, err := b64Decode(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	mac := hmac.New(sha256.New, []byte(WeakSecret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return nil, ErrInvalidToken
	}
	return parseClaims(parts[1])
}

// IssueVulnerable mints a low-privilege user token signed with the weak secret.
func IssueVulnerable(sub string, now time.Time) (string, error) {
	claims := Claims{
		Sub:  sub,
		Role: "user",
		Iss:  IssuerLab,
		Exp:  now.Add(15 * time.Minute).Unix(),
	}
	return signVulnerable(claims)
}

func signVulnerable(claims Claims) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, []byte(WeakSecret))
	mac.Write([]byte(header + "." + payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return header + "." + payload + "." + sig, nil
}

// --- secure reference implementation: fixed algorithm + full claim checks ---

type SecureVerifier struct {
	Secret string
	Issuer string
	Now    func() time.Time
}

var ErrExpired = errors.New("token expired")

func (v *SecureVerifier) Verify(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerJSON, err := b64Decode(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var header struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, ErrInvalidToken
	}
	// Fix 1: algorithm is pinned by the server, never taken from the token.
	if header.Alg != "HS256" {
		return nil, ErrInvalidToken
	}

	sig, err := b64Decode(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	mac := hmac.New(sha256.New, []byte(v.Secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return nil, ErrInvalidToken
	}

	claims, err := parseClaims(parts[1])
	if err != nil {
		return nil, err
	}
	// Fix 2: issuer and expiry are actually enforced.
	if v.Issuer != "" && claims.Iss != v.Issuer {
		return nil, ErrInvalidToken
	}
	now := time.Now()
	if v.Now != nil {
		now = v.Now()
	}
	if claims.Exp <= now.Unix() {
		return nil, ErrExpired
	}
	return claims, nil
}

func parseClaims(rawB64 string) (*Claims, error) {
	raw, err := b64Decode(rawB64)
	if err != nil {
		return nil, ErrInvalidToken
	}
	var c Claims
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, ErrInvalidToken
	}
	return &c, nil
}

func b64Decode(s string) ([]byte, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("base64: %w", err)
	}
	return b, nil
}
