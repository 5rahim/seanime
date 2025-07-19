package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type TokenClaims struct {
	Endpoint  string `json:"endpoint"` // The endpoint this token is valid for
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type HMACAuth struct {
	secret []byte
	ttl    time.Duration
}

// base64URLEncode encodes to base64url without padding (to match frontend)
func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

// base64URLDecode decodes from base64url with or without padding
func base64URLDecode(data string) ([]byte, error) {
	// Add padding if needed
	if m := len(data) % 4; m != 0 {
		data += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(data)
}

// NewHMACAuth creates a new HMAC authentication instance
func NewHMACAuth(secret string, ttl time.Duration) *HMACAuth {
	return &HMACAuth{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// GenerateToken generates an HMAC-signed token for the given endpoint
func (h *HMACAuth) GenerateToken(endpoint string) (string, error) {
	now := time.Now().Unix()
	claims := TokenClaims{
		Endpoint:  endpoint,
		IssuedAt:  now,
		ExpiresAt: now + int64(h.ttl.Seconds()),
	}

	// Serialize claims to JSON
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Encode claims as base64url without padding
	claimsB64 := base64URLEncode(claimsJSON)

	// Generate HMAC signature
	mac := hmac.New(sha256.New, h.secret)
	mac.Write([]byte(claimsB64))
	signature := base64URLEncode(mac.Sum(nil))

	// Return token in format: claims.signature
	return claimsB64 + "." + signature, nil
}

// ValidateToken validates an HMAC token and returns the claims if valid
func (h *HMACAuth) ValidateToken(token string, endpoint string) (*TokenClaims, error) {
	// Split token into claims and signature
	parts := splitToken(token)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format - expected 2 parts, got %d", len(parts))
	}

	claimsB64, signature := parts[0], parts[1]

	// Verify signature
	mac := hmac.New(sha256.New, h.secret)
	mac.Write([]byte(claimsB64))
	expectedSignature := base64URLEncode(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, fmt.Errorf("invalid token signature, the password hashes may not match")
	}

	// Decode claims
	claimsJSON, err := base64URLDecode(claimsB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims TokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Validate expiration
	now := time.Now().Unix()
	if claims.ExpiresAt < now {
		return nil, fmt.Errorf("token expired - expires at %d, current time %d", claims.ExpiresAt, now)
	}

	// Validate endpoint (optional, can be wildcard *)
	if endpoint != "" && claims.Endpoint != "*" && claims.Endpoint != endpoint {
		return nil, fmt.Errorf("token not valid for endpoint %s - claim endpoint: %s", endpoint, claims.Endpoint)
	}

	return &claims, nil
}

// GenerateQueryParam generates a query parameter string with the HMAC token
func (h *HMACAuth) GenerateQueryParam(endpoint string, symbol string) (string, error) {
	token, err := h.GenerateToken(endpoint)
	if err != nil {
		return "", err
	}

	if symbol == "" {
		symbol = "?"
	}

	return fmt.Sprintf("%stoken=%s", symbol, token), nil
}

// ValidateQueryParam extracts and validates token from query parameter
func (h *HMACAuth) ValidateQueryParam(tokenParam string, endpoint string) (*TokenClaims, error) {
	if tokenParam == "" {
		return nil, fmt.Errorf("no token provided")
	}

	return h.ValidateToken(tokenParam, endpoint)
}

// splitToken splits a token string by the last dot separator
func splitToken(token string) []string {
	// Find the last dot to split claims from signature
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			return []string{token[:i], token[i+1:]}
		}
	}
	return []string{token}
}

func (h *HMACAuth) GetTokenExpiry(token string) (time.Time, error) {
	parts := splitToken(token)
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid token format")
	}

	claimsJSON, err := base64URLDecode(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims TokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return time.Time{}, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	return time.Unix(claims.ExpiresAt, 0), nil
}

func (h *HMACAuth) IsTokenExpired(token string) bool {
	expiry, err := h.GetTokenExpiry(token)
	if err != nil {
		return true
	}
	return time.Now().After(expiry)
}
