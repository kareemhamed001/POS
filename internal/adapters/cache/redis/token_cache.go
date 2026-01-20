package redis

import (
	"context"
	"fmt"
	"time"

	redisClient "github.com/kareemhamed001/POS/pkg/redis"
)

const (
	tokenBlacklistPrefix = "blacklist:token:"
)

type TokenCache struct {
	client *redisClient.Client
}

func NewTokenCache(client *redisClient.Client) *TokenCache {
	return &TokenCache{client: client}
}

// BlacklistToken adds a token to the blacklist
func (c *TokenCache) BlacklistToken(ctx context.Context, token string, expiration time.Duration) error {
	if !c.client.IsEnabled() {
		return nil
	}

	key := fmt.Sprintf("%s%s", tokenBlacklistPrefix, token)
	return c.client.Set(ctx, key, "1", expiration).Err()
}

// IsTokenBlacklisted checks if a token is blacklisted
func (c *TokenCache) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	if !c.client.IsEnabled() {
		return false, nil
	}

	key := fmt.Sprintf("%s%s", tokenBlacklistPrefix, token)
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}
