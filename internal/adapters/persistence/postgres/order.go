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
	ErrOrderNotFound = errors.New("Order not found")
)

type OrderRepository struct {
	db     *gorm.DB
	tracer trace.Tracer
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{
		db:     db,
		tracer: otel.Tracer("order-repo"),
	}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *entity.Order) error {
	ctx, span := r.tracer.Start(ctx, "OrderRepository.CreateOrder")
	defer span.End()

	span.SetAttributes(attribute.Int("order.items.count", len(order.Items)))

	if err := gorm.G[entity.Order](r.db).Create(ctx, order); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(attribute.Int("order.id", int(order.ID)))
	span.SetStatus(codes.Ok, "order created")
	return nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id uint) (*entity.Order, error) {
	ctx, span := r.tracer.Start(ctx, "OrderRepository.GetOrderByID")
	defer span.End()

	span.SetAttributes(attribute.Int("order.id", int(id)))

	order, err := gorm.G[entity.Order](r.db).
		Preload("User", nil).
		Preload("Address", nil).
		Preload("Items.Product", nil).
		Where("id = ?", id).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, ErrOrderNotFound.Error())
			return nil, ErrOrderNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "order retrieved")
	return &order, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id uint, status entity.OrderStatus) error {
	ctx, span := r.tracer.Start(ctx, "OrderRepository.UpdateOrderStatus")
	defer span.End()

	span.SetAttributes(
		attribute.Int("order.id", int(id)),
		attribute.String("order.status", string(status)),
	)

	rowsAffected, err := gorm.G[entity.Order](r.db).
		Where("id = ?", id).
		Updates(ctx, entity.Order{Status: status})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if rowsAffected == 0 {
		span.SetStatus(codes.Error, ErrOrderNotFound.Error())
		return ErrOrderNotFound
	}

	span.SetStatus(codes.Ok, "order status updated")
	return nil
}

func (r *OrderRepository) DeleteOrder(ctx context.Context, id uint) error {
	ctx, span := r.tracer.Start(ctx, "OrderRepository.DeleteOrder")
	defer span.End()

	span.SetAttributes(attribute.Int("order.id", int(id)))

	rowsAffected, err := gorm.G[entity.Order](r.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if rowsAffected == 0 {
		span.SetStatus(codes.Error, ErrOrderNotFound.Error())
		return ErrOrderNotFound
	}

	span.SetStatus(codes.Ok, "order deleted")
	return nil
}
