package response

type OrderResponse struct {
	ID            uint                `json:"id"`
	Status        string              `json:"status"`
	UserID        uint                `json:"user_id"`
	AddressID     uint                `json:"address_id"`
	SubTotal      float32             `json:"subtotal"`
	ShippingCost  float32             `json:"shipping_cost"`
	DiscountType  string              `json:"discount_type"`
	DiscountValue float32             `json:"discount_value"`
	Total         float32             `json:"total"`
	Items         []OrderItemResponse `json:"items,omitempty"`
}

type OrderItemResponse struct {
	ID            uint    `json:"id"`
	ProductID     uint    `json:"product_id"`
	Quantity      int     `json:"quantity"`
	Price         float32 `json:"price"`
	SubTotal      float32 `json:"subtotal"`
	DiscountType  string  `json:"discount_type"`
	DiscountValue float32 `json:"discount_value"`
	Total         float32 `json:"total"`
}
