package postgres

import (
	"context"
	"errors"

	"github.com/kareemhamed001/POS/internal/entity"
	"gorm.io/gorm"
)

var (
	ErrOrderItemNotFound = errors.New("Order item not found")
)

type OrderItemRepository struct {
	db *gorm.DB
}

func NewOrderItemRepository(db *gorm.DB) *OrderItemRepository {
	return &OrderItemRepository{db: db}
}

func (r *OrderItemRepository) AddOrderItem(ctx context.Context, item *entity.OrderItem) error {
	return gorm.G[entity.OrderItem](r.db).Create(ctx, item)
}

func (r *OrderItemRepository) GetOrderItemsByOrderID(ctx context.Context, orderID uint) ([]entity.OrderItem, error) {
	items, err := gorm.G[entity.OrderItem](r.db).
		Preload("Product", nil).
		Where("order_id = ?", orderID).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *OrderItemRepository) UpdateOrderItem(ctx context.Context, item *entity.OrderItem) error {
	rowsAffected, err := gorm.G[entity.OrderItem](r.db).
		Where("id = ?", item.ID).
		Updates(ctx, *item)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrOrderItemNotFound
	}
	return nil
}

func (r *OrderItemRepository) DeleteOrderItem(ctx context.Context, id uint) error {
	rowsAffected, err := gorm.G[entity.OrderItem](r.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrOrderItemNotFound
	}
	return nil
}
