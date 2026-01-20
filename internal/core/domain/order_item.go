package entity

import "gorm.io/gorm"

type OrderItem struct {
	gorm.Model

	ProductID     uint         `json:"product_id"`
	Product       Product      `json:"product" gorm:"foreignKey:ProductID"`
	OrderID       uint         `json:"order_id"`
	Order         Order        `json:"order" gorm:"foreignKey:OrderID"`
	Quantity      int          `json:"quantity"`
	Price         float32      `json:"price"`
	SubTotal      float32      `json:"subtotal"`
	DiscountType  DiscountType `json:"discount_type"`
	DiscountValue float32      `json:"discount_value"`
	Total         float32      `json:"total"`
}
