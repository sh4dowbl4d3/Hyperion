package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost       = 12
	maxPasswordBytes = 72
)

func HashPassword(password string) (string, error) {
	if len(password) > maxPasswordBytes {
		return "", fmt.Errorf("password exceeds %d bytes", maxPasswordBytes)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
