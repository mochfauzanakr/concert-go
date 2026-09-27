package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// JWTClaims holds standard and custom token claims
type JWTClaims struct {
	UserID    uuid.UUID `json:"userId"`
	Email     string    `json:"email"`
	RoleID    *int      `json:"roleId,omitempty"`
	ExpiresAt int64     `json:"exp"`
	IssuedAt  int64     `json:"iat"`
}

// GenerateJWT creates a signed HMAC-SHA256 JWT
func GenerateJWT(secret string, claims JWTClaims) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	dataToSign := encodedHeader + "." + encodedPayload

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(dataToSign))
	signature := h.Sum(nil)
	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	return dataToSign + "." + encodedSignature, nil
}

// ValidateJWT verifies signature and expiration, then returns claims
func ValidateJWT(secret string, tokenStr string) (*JWTClaims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	dataToSign := parts[0] + "." + parts[1]
	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid signature encoding")
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(dataToSign))
	expectedSignature := h.Sum(nil)

	if !hmac.Equal(providedSignature, expectedSignature) {
		return nil, errors.New("invalid token signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid payload encoding")
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errors.New("invalid token claims")
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("token has expired")
	}

	return &claims, nil
}

// HashToken hashes a refresh token string using SHA256 for safe DB storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
