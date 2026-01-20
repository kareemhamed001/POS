package entity

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name      string    `json:"name"`
	Email     string    `json:"email" gorm:"unique"`
	Phone     string    `json:"phone" gorm:"unique"`
	Password  string    `json:"-"`
	Orders    []Order   `json:"orders"`
	Addresses []Address `json:"addresses"`
	Role      UserRole  `json:"role" gorm:"default:'customer'"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Role == "" {
		u.Role = "customer"
	}
	return
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	if u.Role == "" {
		u.Role = "customer"
	}
	return
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

func (u *User) ToString() string {
	return u.Name + " <" + u.Email + ">"
}

func (u *User) MaskedEmail() string {
	if len(u.Email) < 3 {
		return "****"
	}
	return u.Email[:2] + "****" + u.Email[len(u.Email)-1:]
}

func (u *User) MaskedPhone() string {
	if len(u.Phone) < 4 {
		return "****"
	}
	return u.Phone[:2] + "****" + u.Phone[len(u.Phone)-2:]
}
