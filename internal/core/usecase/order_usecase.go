package usecase

import (
	"context"
	"errors"
	"sync"
	"time"

	entity "github.com/kareemhamed001/POS/internal/core/domain"
	"github.com/kareemhamed001/POS/internal/core/ports"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type OrderUsecase struct {
	orderRepo   ports.OrderRepository
	productRepo ports.ProductRepository
	tracer      trace.Tracer
}

func NewOrderUsecase(orderRepo ports.OrderRepository, productRepo ports.ProductRepository) *OrderUsecase {
	return &OrderUsecase{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		tracer:      otel.Tracer("order-usecase"),
	}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, order *entity.Order) error {
	ctx, span := u.tracer.Start(ctx, "OrderUsecase.CreateOrder")
	defer span.End()

	span.SetAttributes(
		attribute.Int("order.items.count", len(order.Items)),
	)

	fetchCtx, fetchSpan := u.tracer.Start(ctx, "FetchAllProducts")
	productIDs := make([]uint, len(order.Items))
	for i := range order.Items {
		productIDs[i] = order.Items[i].ProductID
	}

	products, err := u.productRepo.GetProductsByIDs(fetchCtx, productIDs)
	fetchSpan.End()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	productMap := make(map[uint]*entity.Product)
	for i := range products {
		productMap[products[i].ID] = &products[i]
	}

	quantityDeductions := make(map[uint]int)

	itemCalcCtx, itemCalcSpan := u.tracer.Start(ctx, "ProcessOrderItemsParallel")
	processErr := u.processOrderItemsParallel(itemCalcCtx, order, productMap, quantityDeductions)
	itemCalcSpan.End()
	if processErr != nil {
		span.RecordError(processErr)
		span.SetStatus(codes.Error, processErr.Error())
		return processErr
	}

	for productID, deduction := range quantityDeductions {
		if product, exists := productMap[productID]; exists {
			product.Quantity -= deduction
		}
	}

	var subTotal float32
	for _, item := range order.Items {
		subTotal += item.Total
	}

	_, orderCalcSpan := u.tracer.Start(ctx, "CalculateOrderTotals")
	order.SubTotal = subTotal

	finalTotal := subTotal + order.ShippingCost
	if order.DiscountType == entity.DiscountFixed {
		orderCalcSpan.SetAttributes(attribute.String("order.discount.type", "fixed"))
		finalTotal -= order.DiscountValue
	} else if order.DiscountType == entity.DiscountPercent {
		orderCalcSpan.SetAttributes(attribute.String("order.discount.type", "percent"))
		finalTotal -= finalTotal * (order.DiscountValue / 100)
	}

	if finalTotal < 0 {
		orderCalcSpan.SetAttributes(attribute.String("order.total.adjustment", "set to zero from negative"))
		finalTotal = 0
	}

	order.Total = finalTotal
	order.Status = entity.StatusPending
	orderCalcSpan.End()

	updateCtx, updateSpan := u.tracer.Start(ctx, "UpdateProductStock")
	updatedProducts := make([]*entity.Product, 0, len(productMap))
	for _, product := range productMap {
		updatedProducts = append(updatedProducts, product)
	}

	for _, product := range updatedProducts {
		if err := u.productRepo.UpdateProduct(updateCtx, product.ID, product); err != nil {
			updateSpan.RecordError(err)
			updateSpan.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	}
	updateSpan.End()

	err = u.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (u *OrderUsecase) processOrderItemsParallel(ctx context.Context, order *entity.Order, productMap map[uint]*entity.Product, quantityDeductions map[uint]int) error {
	now := time.Now()
	var mu sync.Mutex
	var processErr error
	var wg sync.WaitGroup

	for i := range order.Items {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			_, itemSpan := u.tracer.Start(ctx, "ProcessOrderItem")
			defer itemSpan.End()

			itemSpan.SetAttributes(
				attribute.Int("item.index", index),
				attribute.Int("item.product_id", int(order.Items[index].ProductID)),
				attribute.Int("item.quantity", order.Items[index].Quantity),
			)

			item := &order.Items[index]

			mu.Lock()
			product, exists := productMap[item.ProductID]
			totalDeductions := quantityDeductions[item.ProductID]
			mu.Unlock()

			if !exists {
				err := errors.New("product not found")
				mu.Lock()
				processErr = err
				mu.Unlock()
				itemSpan.RecordError(err)
				itemSpan.SetStatus(codes.Error, err.Error())
				return
			}

			availableStock := product.Quantity - totalDeductions
			if availableStock < item.Quantity {
				err := errors.New("insufficient stock for product: " + product.Name)
				mu.Lock()
				processErr = err
				mu.Unlock()
				itemSpan.RecordError(err)
				itemSpan.SetStatus(codes.Error, err.Error())
				return
			}

			item.Price = product.Price
			item.SubTotal = item.Price * float32(item.Quantity)

			if item.DiscountValue == 0 && product.DiscountValue > 0 {
				discountActive := true
				if product.DiscountStartDate != nil && now.Before(*product.DiscountStartDate) {
					itemSpan.SetAttributes(attribute.String("discount.status", "not started"))
					discountActive = false
				}
				if product.DiscountEndDate != nil && now.After(*product.DiscountEndDate) {
					itemSpan.SetAttributes(attribute.String("discount.status", "expired"))
					discountActive = false
				}

				if discountActive {
					item.DiscountType = product.DiscountType
					item.DiscountValue = product.DiscountValue
				}
			}

			itemTotal := item.SubTotal
			if item.DiscountType == entity.DiscountFixed {
				itemTotal -= item.DiscountValue
			} else if item.DiscountType == entity.DiscountPercent {
				itemTotal -= item.SubTotal * (item.DiscountValue / 100)
			}

			if itemTotal < 0 {
				itemSpan.SetAttributes(attribute.String("item.total.adjustment", "set to zero from negative"))
				itemTotal = 0
			}

			item.Total = itemTotal

			mu.Lock()
			quantityDeductions[item.ProductID] += item.Quantity
			mu.Unlock()

			itemSpan.SetStatus(codes.Ok, "item processed successfully")
		}(i)
	}

	wg.Wait()
	return processErr
}

func (u *OrderUsecase) GetOrder(ctx context.Context, id uint) (*entity.Order, error) {
	ctx, span := u.tracer.Start(ctx, "OrderUsecase.GetOrder")
	defer span.End()

	span.SetAttributes(attribute.Int("order.id", int(id)))

	order, err := u.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "order retrieved")
	return order, nil
}

func (u *OrderUsecase) UpdateStatus(ctx context.Context, id uint, status entity.OrderStatus) error {
	ctx, span := u.tracer.Start(ctx, "OrderUsecase.UpdateStatus")
	defer span.End()

	span.SetAttributes(
		attribute.Int("order.id", int(id)),
		attribute.String("order.status", string(status)),
	)

	order, err := u.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if status == entity.StatusCancelled && order.Status != entity.StatusCancelled {
		_, revertSpan := u.tracer.Start(ctx, "OrderUsecase.RevertStock")
		for _, item := range order.Items {
			product, err := u.productRepo.GetProductByID(ctx, item.ProductID)
			if err != nil {
				continue
			}
			product.Quantity += item.Quantity
			if err := u.productRepo.UpdateProduct(ctx, product.ID, product); err != nil {
			}
		}
		revertSpan.End()
	}

	if err := u.orderRepo.UpdateOrderStatus(ctx, id, status); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "order status updated")
	return nil
}

func (u *OrderUsecase) DeleteOrder(ctx context.Context, id uint) error {
	ctx, span := u.tracer.Start(ctx, "OrderUsecase.DeleteOrder")
	defer span.End()

	span.SetAttributes(attribute.Int("order.id", int(id)))

	if err := u.orderRepo.DeleteOrder(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "order deleted")
	return nil
}
