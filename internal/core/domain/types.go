package entity

type DiscountType string

const (
	DiscountFixed   DiscountType = "fixed"
	DiscountPercent DiscountType = "percent"
)

func ValidDiscountTypes() []DiscountType {
	return []DiscountType{
		DiscountFixed,
		DiscountPercent,
	}
}

func (d DiscountType) IsValid() bool {
	for _, valid := range ValidDiscountTypes() {
		if d == valid {
			return true
		}
	}
	return false
}

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusCompleted OrderStatus = "completed"
	StatusCancelled OrderStatus = "cancelled"
)

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleAdmin    UserRole = "admin"
)
