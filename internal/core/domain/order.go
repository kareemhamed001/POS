package entity

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	Status        OrderStatus  `json:"status"`
	UserID        uint         `json:"user_id"`
	User          User         `json:"user" gorm:"foreignKey:UserID"`
	AddressID     uint         `json:"address_id" gorm:"foreignKey:AddressID"`
	Address       Address      `json:"address" gorm:"foreignKey:AddressID"`
	SubTotal      float32      `json:"subtotal"`
	ShippingCost  float32      `json:"shipping_cost"`
	DiscountType  DiscountType `json:"discount_type"`
	DiscountValue float32      `json:"discount_value"`
	Total         float32      `json:"total"`
	Items         []OrderItem  `json:"items" gorm:"foreignKey:OrderID"`
}
