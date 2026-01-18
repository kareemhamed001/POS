package entity

type DiscountType string

const (
	DiscountFixed   DiscountType = "fixed"
	DiscountPercent DiscountType = "percent"
)

// ValidDiscountTypes returns a slice of all valid discount types
func ValidDiscountTypes() []DiscountType {
	return []DiscountType{
		DiscountFixed,
		DiscountPercent,
	}
}

// IsValid checks if the discount type is valid
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
