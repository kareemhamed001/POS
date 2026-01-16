package usecase

import (
	"context"

	"github.com/kareemhamed001/POS/internal/entity"
)

type OrderItemUsecase struct {
	orderItemRepo OrderItemRepository
}

func NewOrderItemUsecase(orderItemRepo OrderItemRepository) *OrderItemUsecase {
	return &OrderItemUsecase{
		orderItemRepo: orderItemRepo,
	}
}

func (u *OrderItemUsecase) AddOrderItem(ctx context.Context, item *entity.OrderItem) error {
	return u.orderItemRepo.AddOrderItem(ctx, item)
}

func (u *OrderItemUsecase) GetOrderItems(ctx context.Context, orderID uint) ([]entity.OrderItem, error) {
	return u.orderItemRepo.GetOrderItemsByOrderID(ctx, orderID)
}

func (u *OrderItemUsecase) UpdateOrderItem(ctx context.Context, item *entity.OrderItem) error {
	return u.orderItemRepo.UpdateOrderItem(ctx, item)
}

func (u *OrderItemUsecase) DeleteOrderItem(ctx context.Context, id uint) error {
	return u.orderItemRepo.DeleteOrderItem(ctx, id)
}
