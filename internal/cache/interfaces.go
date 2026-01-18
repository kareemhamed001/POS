package cache

import (
	"context"
	"time"

	"github.com/kareemhamed001/POS/internal/entity"
)

// ProductCache defines caching operations for products
type ProductCache interface {
	GetProduct(ctx context.Context, id uint) (*entity.Product, error)
	SetProduct(ctx context.Context, product *entity.Product, ttl time.Duration) error
	DeleteProduct(ctx context.Context, id uint) error
	GetProductList(ctx context.Context, key string) ([]entity.Product, error)
	SetProductList(ctx context.Context, key string, products []entity.Product, ttl time.Duration) error
	InvalidateProductList(ctx context.Context) error
}

// TokenCache defines caching operations for JWT tokens (blacklist)
type TokenCache interface {
	BlacklistToken(ctx context.Context, token string, expiration time.Duration) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

// RateLimitCache defines rate limiting operations
type RateLimitCache interface {
	IncrementCounter(ctx context.Context, key string, window time.Duration) (int64, error)
	GetCounter(ctx context.Context, key string) (int64, error)
}
