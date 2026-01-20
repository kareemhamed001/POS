package entity

import "gorm.io/gorm"

type Address struct {
	gorm.Model
	UserID      uint    `json:"user_id"`
	User        User    `json:"user" gorm:"foreignKey:UserID"`
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	Governorate string  `json:"governorate"`
	City        string  `json:"city"`
	Address     string  `json:"address"`
	Phone       *string `json:"phone"`
	IsPrimary   bool    `json:"is_primary"`
}
