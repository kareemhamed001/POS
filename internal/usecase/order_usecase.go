package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/kareemhamed001/POS/internal/entity"
)

type OrderUsecase struct {
	orderRepo   OrderRepository
	productRepo ProductRepository
}

func NewOrderUsecase(orderRepo OrderRepository, productRepo ProductRepository) *OrderUsecase {
	return &OrderUsecase{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, order *entity.Order) error {
	var subTotal float32
	now := time.Now()

	for i := range order.Items {
		item := &order.Items[i]
		product, err := u.productRepo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return err
		}

		if product.Quantity < item.Quantity {
			return errors.New("insufficient stock for product: " + product.Name)
		}

		// Set base price from product
		item.Price = product.Price
		item.SubTotal = item.Price * float32(item.Quantity)

		// Apply product-level discount if active and not overridden by item-level discount
		// (Or we could decide to always use product discount if available)
		if item.DiscountValue == 0 && product.DiscountValue > 0 {
			discountActive := true
			if product.DiscountStartDate != nil && now.Before(*product.DiscountStartDate) {
				discountActive = false
			}
			if product.DiscountEndDate != nil && now.After(*product.DiscountEndDate) {
				discountActive = false
			}

			if discountActive {
				item.DiscountType = product.DiscountType
				item.DiscountValue = product.DiscountValue
			}
		}

		// Calculate item total after discount
		itemTotal := item.SubTotal
		if item.DiscountType == entity.DiscountFixed {
			itemTotal -= item.DiscountValue
		} else if item.DiscountType == entity.DiscountPercent {
			itemTotal -= item.SubTotal * (item.DiscountValue / 100)
		}

		// Ensure total doesn't go negative
		if itemTotal < 0 {
			itemTotal = 0
		}

		item.Total = itemTotal
		subTotal += itemTotal

		// Deduct stock
		product.Quantity -= item.Quantity
		if err := u.productRepo.UpdateProduct(ctx, product.ID, product); err != nil {
			return err
		}
	}

	order.SubTotal = subTotal

	// Order level discount
	finalTotal := subTotal + order.ShippingCost
	if order.DiscountType == entity.DiscountFixed {
		finalTotal -= order.DiscountValue
	} else if order.DiscountType == entity.DiscountPercent {
		finalTotal -= finalTotal * (order.DiscountValue / 100)
	}

	if finalTotal < 0 {
		finalTotal = 0
	}

	order.Total = finalTotal
	order.Status = entity.StatusPending

	return u.orderRepo.CreateOrder(ctx, order)
}

func (u *OrderUsecase) GetOrder(ctx context.Context, id uint) (*entity.Order, error) {
	return u.orderRepo.GetOrderByID(ctx, id)
}

func (u *OrderUsecase) UpdateStatus(ctx context.Context, id uint, status entity.OrderStatus) error {
	order, err := u.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}

	// Logic to return stock if order is cancelled
	if status == entity.StatusCancelled && order.Status != entity.StatusCancelled {
		for _, item := range order.Items {
			product, err := u.productRepo.GetProductByID(ctx, item.ProductID)
			if err != nil {
				// We might want to log this but continue if product is missing
				continue
			}
			product.Quantity += item.Quantity
			u.productRepo.UpdateProduct(ctx, product.ID, product)
		}
	}

	return u.orderRepo.UpdateOrderStatus(ctx, id, status)
}

func (u *OrderUsecase) DeleteOrder(ctx context.Context, id uint) error {
	return u.orderRepo.DeleteOrder(ctx, id)
}
