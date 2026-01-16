package postgres

import (
	"context"
	"errors"

	"github.com/kareemhamed001/POS/internal/entity"
	"gorm.io/gorm"
)

var (
	ErrOrderNotFound = errors.New("Order not found")
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *entity.Order) error {
	return gorm.G[entity.Order](r.db).Create(ctx, order)
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id uint) (*entity.Order, error) {
	order, err := gorm.G[entity.Order](r.db).
		Preload("User", nil).
		Preload("Address", nil).
		Preload("Items.Product", nil).
		Where("id = ?", id).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id uint, status entity.OrderStatus) error {
	rowsAffected, err := gorm.G[entity.Order](r.db).
		Where("id = ?", id).
		Updates(ctx, entity.Order{Status: status})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) DeleteOrder(ctx context.Context, id uint) error {
	rowsAffected, err := gorm.G[entity.Order](r.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrOrderNotFound
	}
	return nil
}
