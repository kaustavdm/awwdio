package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JWTHeader represents the JWT header
type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// JWTClaims represents the JWT payload claims
type JWTClaims struct {
	Sub string `json:"sub"` // Subject (user identifier)
	Iat int64  `json:"iat"` // Issued at
	Exp int64  `json:"exp"` // Expiration time
}

// GenerateJWT creates a new JWT token for the given subject using HS256
func GenerateJWT(subject string, secret string, expiry time.Duration) (string, error) {
	header := JWTHeader{
		Alg: "HS256",
		Typ: "JWT",
	}

	now := time.Now()
	claims := JWTClaims{
		Sub: subject,
		Iat: now.Unix(),
		Exp: now.Add(expiry).Unix(),
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode claims
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	signingInput := headerB64 + "." + claimsB64
	signature := signHS256(signingInput, secret)

	return signingInput + "." + signature, nil
}

// ValidateJWT validates a JWT token and returns the claims if valid
func ValidateJWT(token string, secret string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerB64 := parts[0]
	claimsB64 := parts[1]
	signature := parts[2]

	// Verify signature
	signingInput := headerB64 + "." + claimsB64
	expectedSig := signHS256(signingInput, secret)
	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode and verify header
	headerJSON, err := base64.RawURLEncoding.DecodeString(headerB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	var header JWTHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	if header.Alg != "HS256" {
		return nil, fmt.Errorf("unsupported algorithm: %s", header.Alg)
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// signHS256 creates an HMAC-SHA256 signature
func signHS256(input string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
