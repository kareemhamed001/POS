package response

type AddressResponse struct {
	ID          uint    `json:"id"`
	UserID      uint    `json:"user_id"`
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	Governorate string  `json:"governorate"`
	City        string  `json:"city"`
	Address     string  `json:"address"`
	Phone       *string `json:"phone"`
	IsPrimary   bool    `json:"is_primary"`
}
