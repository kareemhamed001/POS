package entity

type DiscountType string

const (
	DiscountFixed   DiscountType = "fixed"
	DiscountPercent DiscountType = "percent"
)

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
