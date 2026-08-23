package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenInvalid = errors.New("token invalid")
	ErrTokenExpired = errors.New("token expired")
)

const tokenIssuer = "hyperion"

type Claims struct {
	UserID string
	Role   string
}

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string, ttl time.Duration) *TokenService {
	return &TokenService{secret: []byte(secret), ttl: ttl}
}

type registeredClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

func (t *TokenService) TTL() time.Duration { return t.ttl }

func (t *TokenService) Issue(userID, role string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(t.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, registeredClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Role: role,
	})
	signed, err := token.SignedString(t.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

func (t *TokenService) Verify(tokenString string) (Claims, error) {
	var parsed registeredClaims
	token, err := jwt.ParseWithClaims(tokenString, &parsed, func(*jwt.Token) (any, error) {
		return t.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(tokenIssuer),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Claims{}, ErrTokenExpired
		}
		return Claims{}, ErrTokenInvalid
	}
	if !token.Valid || parsed.Subject == "" || parsed.Role == "" {
		return Claims{}, ErrTokenInvalid
	}
	return Claims{UserID: parsed.Subject, Role: parsed.Role}, nil
}
