package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps Redis client with token-specific operations.
type Client struct {
	rdb *redis.Client
}

// Config holds Redis configuration.
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewClient creates a new Redis client.
func NewClient(cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// StoreToken stores a token in Redis with the given TTL.
func (c *Client) StoreToken(
	ctx context.Context,
	tokenID string,
	userID int,
	tokenType string,
	ttl time.Duration,
) error {
	key := fmt.Sprintf("token:%s:%s", tokenType, tokenID)
	userIndexKey := fmt.Sprintf("user:tokens:%d", userID)

	// Use pipeline for atomic operations
	pipe := c.rdb.Pipeline()

	// Store token with user ID as value
	pipe.Set(ctx, key, userID, ttl)

	// Add token ID to user's token set
	pipe.SAdd(ctx, userIndexKey, tokenID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	return nil
}

// TokenExists checks if a token exists in Redis.
func (c *Client) TokenExists(ctx context.Context, tokenID string, tokenType string) (bool, error) {
	key := fmt.Sprintf("token:%s:%s", tokenType, tokenID)

	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token existence: %w", err)
	}

	return exists > 0, nil
}

// GetUserIDByToken retrieves the user ID associated with a token.
func (c *Client) GetUserIDByToken(ctx context.Context, tokenID string, tokenType string) (int, error) {
	key := fmt.Sprintf("token:%s:%s", tokenType, tokenID)

	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("token not found")
		}
		return 0, fmt.Errorf("failed to get user ID: %w", err)
	}

	userID, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %w", err)
	}

	return userID, nil
}

// RevokeToken revokes a specific token by removing it from Redis.
func (c *Client) RevokeToken(ctx context.Context, tokenID string, tokenType string, userID int) error {
	key := fmt.Sprintf("token:%s:%s", tokenType, tokenID)
	userIndexKey := fmt.Sprintf("user:tokens:%d", userID)

	// Use pipeline for atomic operations
	pipe := c.rdb.Pipeline()

	// Delete token
	pipe.Del(ctx, key)

	// Remove token ID from user's token set
	pipe.SRem(ctx, userIndexKey, tokenID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens revokes all tokens for a specific user.
func (c *Client) RevokeAllUserTokens(ctx context.Context, userID int) error {
	userIndexKey := fmt.Sprintf("user:tokens:%d", userID)

	// Get all token IDs for the user
	tokenIDs, err := c.rdb.SMembers(ctx, userIndexKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	if len(tokenIDs) == 0 {
		return nil // No tokens to revoke
	}

	// Build keys for all tokens (both access and refresh)
	keys := make([]string, 0, len(tokenIDs)*2+1)
	for _, tokenID := range tokenIDs {
		keys = append(keys, fmt.Sprintf("token:access:%s", tokenID))
		keys = append(keys, fmt.Sprintf("token:refresh:%s", tokenID))
	}
	keys = append(keys, userIndexKey)

	// Delete all keys
	_, err = c.rdb.Del(ctx, keys...).Result()
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	return nil
}

// GetTokenTTL returns the remaining TTL for a token.
func (c *Client) GetTokenTTL(ctx context.Context, tokenID string, tokenType string) (time.Duration, error) {
	key := fmt.Sprintf("token:%s:%s", tokenType, tokenID)

	ttl, err := c.rdb.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get token TTL: %w", err)
	}

	return ttl, nil
}
