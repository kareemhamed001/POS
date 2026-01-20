package ports

import (
	"context"
	"time"

	entity "github.com/kareemhamed001/POS/internal/core/domain"
)

type ProductCache interface {
	GetProduct(ctx context.Context, id uint) (*entity.Product, error)
	SetProduct(ctx context.Context, product *entity.Product, ttl time.Duration) error
	DeleteProduct(ctx context.Context, id uint) error
}

type TokenCache interface {
	BlacklistToken(ctx context.Context, token string, expiration time.Duration) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

type RateLimitCache interface {
	IncrementCounter(ctx context.Context, key string, window time.Duration) (int64, error)
	GetCounter(ctx context.Context, key string) (int64, error)
}
