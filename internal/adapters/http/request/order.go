package request

import entity "github.com/kareemhamed001/POS/internal/core/domain"

type CreateOrderRequest struct {
	UserID        uint                 `json:"user_id" validate:"required"`
	AddressID     uint                 `json:"address_id" validate:"required"`
	Items         []CreateOrderItemReq `json:"items" validate:"required,dive"`
	ShippingCost  float32              `json:"shipping_cost"`
	DiscountType  entity.DiscountType  `json:"discount_type"`
	DiscountValue float32              `json:"discount_value"`
}

type CreateOrderItemReq struct {
	ProductID     uint                `json:"product_id" validate:"required"`
	Quantity      int                 `json:"quantity" validate:"required,gt=0"`
	Price         float32             `json:"price" validate:"required"`
	DiscountType  entity.DiscountType `json:"discount_type"`
	DiscountValue float32             `json:"discount_value"`
}

type UpdateOrderStatusRequest struct {
	Status entity.OrderStatus `json:"status" validate:"required"`
}
