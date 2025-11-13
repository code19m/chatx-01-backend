package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TokenType represents the type of token.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims represents JWT claims.
type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

// Generator defines the interface for token generation and validation.
type Generator interface {
	Generate(userID int, role string, tokenType TokenType) (string, error)
	Validate(token string) (*Claims, error)
}

type jwtGenerator struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewGenerator creates a new JWT token generator.
func NewGenerator(secret string, accessTokenTTL, refreshTokenTTL time.Duration) Generator {
	return &jwtGenerator{
		secret:          []byte(secret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// Generate creates a new JWT token.
func (g *jwtGenerator) Generate(userID int, role string, tokenType TokenType) (string, error) {
	now := time.Now()
	var exp time.Time

	if tokenType == TokenTypeAccess {
		exp = now.Add(g.accessTokenTTL)
	} else {
		exp = now.Add(g.refreshTokenTTL)
	}

	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   string(tokenType),
		Iat:    now.Unix(),
		Exp:    exp.Unix(),
	}

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Encode header and claims
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerEncoded + "." + claimsEncoded
	signature := g.sign(message)

	return message + "." + signature, nil
}

// Validate validates a JWT token and returns the claims.
func (g *jwtGenerator) Validate(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signatureEncoded := parts[2]

	// Verify signature
	message := headerEncoded + "." + claimsEncoded
	expectedSignature := g.sign(message)

	if signatureEncoded != expectedSignature {
		return nil, fmt.Errorf("invalid token signature")
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

func (g *jwtGenerator) sign(message string) string {
	h := hmac.New(sha256.New, g.secret)
	h.Write([]byte(message))
	signature := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(signature)
}
