package redis

import (
	"context"
	"fmt"
	"time"

	redisClient "github.com/kareemhamed001/POS/pkg/redis"
)

const (
	rateLimitPrefix = "ratelimit:"
)

type RateLimitCache struct {
	client *redisClient.Client
}

func NewRateLimitCache(client *redisClient.Client) *RateLimitCache {
	return &RateLimitCache{client: client}
}

// IncrementCounter increments the counter for a key and sets expiration
func (c *RateLimitCache) IncrementCounter(ctx context.Context, key string, window time.Duration) (int64, error) {
	if !c.client.IsEnabled() {
		return 0, nil // Disable rate limiting if Redis is off
	}

	fullKey := fmt.Sprintf("%s%s", rateLimitPrefix, key)

	pipe := c.client.Pipeline()
	incr := pipe.Incr(ctx, fullKey)
	pipe.Expire(ctx, fullKey, window)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

// GetCounter retrieves the current counter value
func (c *RateLimitCache) GetCounter(ctx context.Context, key string) (int64, error) {
	if !c.client.IsEnabled() {
		return 0, nil
	}

	fullKey := fmt.Sprintf("%s%s", rateLimitPrefix, key)
	return c.client.Get(ctx, fullKey).Int64()
}
