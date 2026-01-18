package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/kareemhamed001/POS/internal/config"
	"github.com/kareemhamed001/POS/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client
type Client struct {
	*redis.Client
	enabled bool
}

// NewClient creates a new Redis client with connection pooling
func NewClient(cfg *config.Config) (*Client, error) {
	if !cfg.RedisEnabled {
		logger.Info("Redis is disabled, cache operations will be skipped")
		return &Client{enabled: false}, nil
	}

	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Errorf("Failed to connect to Redis at %s: %v", addr, err)
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	logger.Infof("Redis connected successfully at %s", addr)
	return &Client{Client: rdb, enabled: true}, nil
}

// IsEnabled returns whether Redis is enabled
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if c.enabled && c.Client != nil {
		return c.Client.Close()
	}
	return nil
}
