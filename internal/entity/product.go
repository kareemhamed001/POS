package entity

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name              string       `json:"name"`
	ShortDescription  *string      `json:"short_description"`
	Description       string       `json:"description"`
	Price             float32      `json:"price"`
	DiscountType      DiscountType `json:"discount_type"`
	DiscountValue     float32      `json:"discount_value"`
	DiscountStartDate *time.Time   `json:"discount_start_date"`
	DiscountEndDate   *time.Time   `json:"discount_end_date"`
	ImageUrl          string       `json:"image_url"`
	Quantity          int          `json:"quantity"`
}
