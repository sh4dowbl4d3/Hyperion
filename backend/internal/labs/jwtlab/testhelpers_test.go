package jwtlab

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func base64RawURLEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func hmacSHA256(data, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
