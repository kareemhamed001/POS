package request

type CreateAddressRequest struct {
	UserID      uint    `json:"user_id" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Country     string  `json:"country" validate:"required"`
	Governorate string  `json:"governorate" validate:"required"`
	City        string  `json:"city" validate:"required"`
	Address     string  `json:"address" validate:"required"`
	Phone       *string `json:"phone"`
	IsPrimary   bool    `json:"is_primary"`
}
