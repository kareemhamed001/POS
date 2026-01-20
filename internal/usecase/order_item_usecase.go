package usecase

import (
	"context"

	"github.com/kareemhamed001/POS/internal/entity"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type OrderItemUsecase struct {
	orderItemRepo OrderItemRepository
	tracer        trace.Tracer
}

func NewOrderItemUsecase(orderItemRepo OrderItemRepository) *OrderItemUsecase {
	return &OrderItemUsecase{
		orderItemRepo: orderItemRepo,
		tracer:        otel.Tracer("order-item-usecase"),
	}
}

func (u *OrderItemUsecase) AddOrderItem(ctx context.Context, item *entity.OrderItem) error {
	ctx, span := u.tracer.Start(ctx, "OrderItemUsecase.AddOrderItem")
	defer span.End()

	span.SetAttributes(
		attribute.Int("order_item.product_id", int(item.ProductID)),
		attribute.Int("order_item.quantity", item.Quantity),
	)

	if err := u.orderItemRepo.AddOrderItem(ctx, item); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "order item added")
	return nil
}

func (u *OrderItemUsecase) GetOrderItems(ctx context.Context, orderID uint) ([]entity.OrderItem, error) {
	ctx, span := u.tracer.Start(ctx, "OrderItemUsecase.GetOrderItems")
	defer span.End()

	span.SetAttributes(attribute.Int("order.id", int(orderID)))

	items, err := u.orderItemRepo.GetOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("order_item.count", len(items)))
	span.SetStatus(codes.Ok, "order items retrieved")
	return items, nil
}

func (u *OrderItemUsecase) UpdateOrderItem(ctx context.Context, item *entity.OrderItem) error {
	ctx, span := u.tracer.Start(ctx, "OrderItemUsecase.UpdateOrderItem")
	defer span.End()

	span.SetAttributes(
		attribute.Int("order_item.id", int(item.ID)),
		attribute.Int("order_item.quantity", item.Quantity),
	)

	if err := u.orderItemRepo.UpdateOrderItem(ctx, item); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "order item updated")
	return nil
}

func (u *OrderItemUsecase) DeleteOrderItem(ctx context.Context, id uint) error {
	ctx, span := u.tracer.Start(ctx, "OrderItemUsecase.DeleteOrderItem")
	defer span.End()

	span.SetAttributes(attribute.Int("order_item.id", int(id)))

	if err := u.orderItemRepo.DeleteOrderItem(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "order item deleted")
	return nil
}
