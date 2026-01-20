package postgres

import (
	"context"
	"errors"

	entity "github.com/kareemhamed001/POS/internal/core/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrOrderItemNotFound = errors.New("Order item not found")
)

type OrderItemRepository struct {
	db     *gorm.DB
	tracer trace.Tracer
}

func NewOrderItemRepository(db *gorm.DB) *OrderItemRepository {
	return &OrderItemRepository{
		db:     db,
		tracer: otel.Tracer("order-item-repo"),
	}
}

func (r *OrderItemRepository) AddOrderItem(ctx context.Context, item *entity.OrderItem) error {
	ctx, span := r.tracer.Start(ctx, "OrderItemRepository.AddOrderItem")
	defer span.End()

	span.SetAttributes(
		attribute.Int("order_item.product_id", int(item.ProductID)),
		attribute.Int("order_item.quantity", item.Quantity),
	)

	if err := gorm.G[entity.OrderItem](r.db).Create(ctx, item); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "order item created")
	return nil
}

func (r *OrderItemRepository) GetOrderItemsByOrderID(ctx context.Context, orderID uint) ([]entity.OrderItem, error) {
	ctx, span := r.tracer.Start(ctx, "OrderItemRepository.GetOrderItemsByOrderID")
	defer span.End()

	span.SetAttributes(attribute.Int("order.id", int(orderID)))

	items, err := gorm.G[entity.OrderItem](r.db).
		Preload("Product", nil).
		Where("order_id = ?", orderID).
		Find(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("order_item.count", len(items)))
	span.SetStatus(codes.Ok, "order items retrieved")
	return items, nil
}

func (r *OrderItemRepository) UpdateOrderItem(ctx context.Context, item *entity.OrderItem) error {
	ctx, span := r.tracer.Start(ctx, "OrderItemRepository.UpdateOrderItem")
	defer span.End()

	span.SetAttributes(attribute.Int("order_item.id", int(item.ID)))

	rowsAffected, err := gorm.G[entity.OrderItem](r.db).
		Where("id = ?", item.ID).
		Updates(ctx, *item)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if rowsAffected == 0 {
		span.SetStatus(codes.Error, ErrOrderItemNotFound.Error())
		return ErrOrderItemNotFound
	}

	span.SetStatus(codes.Ok, "order item updated")
	return nil
}

func (r *OrderItemRepository) DeleteOrderItem(ctx context.Context, id uint) error {
	ctx, span := r.tracer.Start(ctx, "OrderItemRepository.DeleteOrderItem")
	defer span.End()

	span.SetAttributes(attribute.Int("order_item.id", int(id)))

	rowsAffected, err := gorm.G[entity.OrderItem](r.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if rowsAffected == 0 {
		span.SetStatus(codes.Error, ErrOrderItemNotFound.Error())
		return ErrOrderItemNotFound
	}

	span.SetStatus(codes.Ok, "order item deleted")
	return nil
}
