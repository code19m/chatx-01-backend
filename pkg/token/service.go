package token

import (
	"context"
	"fmt"
	"time"
)

// TokenStore defines the interface for token storage operations.
type TokenStore interface {
	StoreToken(ctx context.Context, tokenID string, userID int, tokenType string, ttl time.Duration) error
	TokenExists(ctx context.Context, tokenID string, tokenType string) (bool, error)
	GetUserIDByToken(ctx context.Context, tokenID string, tokenType string) (int, error)
	RevokeToken(ctx context.Context, tokenID string, tokenType string, userID int) error
	RevokeAllUserTokens(ctx context.Context, userID int) error
}

// Service wraps the token generator with Redis-based token storage.
type Service struct {
	generator       Generator
	tokenStore      TokenStore
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewService creates a new token service.
func NewService(generator Generator, tokenStore TokenStore, accessTokenTTL, refreshTokenTTL time.Duration) *Service {
	return &Service{
		generator:       generator,
		tokenStore:      tokenStore,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// GenerateAndStore generates a JWT token and stores it in Redis.
func (s *Service) GenerateAndStore(ctx context.Context, userID int, role string, tokenType TokenType) (string, error) {
	// Generate JWT
	tokenString, err := s.generator.Generate(userID, role, tokenType)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Parse to get JTI
	claims, err := s.generator.Validate(tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to validate generated token: %w", err)
	}

	// Determine TTL
	var ttl time.Duration
	if tokenType == TokenTypeAccess {
		ttl = s.accessTokenTTL
	} else {
		ttl = s.refreshTokenTTL
	}

	// Store in Redis
	err = s.tokenStore.StoreToken(ctx, claims.JTI, userID, string(tokenType), ttl)
	if err != nil {
		return "", fmt.Errorf("failed to store token: %w", err)
	}

	return tokenString, nil
}

// ValidateAndCheck validates the JWT and checks if it exists in Redis.
func (s *Service) ValidateAndCheck(ctx context.Context, tokenString string) (*Claims, error) {
	// Validate JWT signature and expiration
	claims, err := s.generator.Validate(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if token exists in Redis (not revoked)
	exists, err := s.tokenStore.TokenExists(ctx, claims.JTI, claims.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to check token status: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("token has been revoked")
	}

	return claims, nil
}

// Revoke revokes a specific token.
func (s *Service) Revoke(ctx context.Context, tokenString string) error {
	// Parse token to get JTI and user ID
	claims, err := s.generator.Validate(tokenString)
	if err != nil {
		// Token is invalid, consider it already revoked
		return nil
	}

	// Revoke from Redis
	err = s.tokenStore.RevokeToken(ctx, claims.JTI, claims.Type, claims.UserID)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens revokes all tokens for a specific user.
func (s *Service) RevokeAllUserTokens(ctx context.Context, userID int) error {
	err := s.tokenStore.RevokeAllUserTokens(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}

	return nil
}

// GetClaims parses and returns claims without Redis validation (for logout).
func (s *Service) GetClaims(tokenString string) (*Claims, error) {
	return s.generator.Validate(tokenString)
}
