package usecase

import (
	"context"

	"github.com/kareemhamed001/POS/internal/entity"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *entity.Order) error
	GetOrderByID(ctx context.Context, id uint) (*entity.Order, error)
	UpdateOrderStatus(ctx context.Context, id uint, status entity.OrderStatus) error
	DeleteOrder(ctx context.Context, id uint) error
}

type UserRepository interface {
	ListUsers(ctx context.Context, search string, page, perPage int) (*[]entity.User, int, error) //int for total users for pagination
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByID(ctx context.Context, id uint) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateUser(ctx context.Context, id uint, user *entity.User) error
	DeleteUser(ctx context.Context, id uint) error
}

type AddressRepository interface {
	AddAddress(ctx context.Context, address *entity.Address) error
	GetAddressesByUserID(ctx context.Context, userID uint) ([]entity.Address, error)
	UpdateAddress(ctx context.Context, address *entity.Address) error
	DeleteAddress(ctx context.Context, id uint) error
}

type OrderItemRepository interface {
	AddOrderItem(ctx context.Context, item *entity.OrderItem) error
	GetOrderItemsByOrderID(ctx context.Context, orderID uint) ([]entity.OrderItem, error)
	UpdateOrderItem(ctx context.Context, item *entity.OrderItem) error
	DeleteOrderItem(ctx context.Context, id uint) error
}

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *entity.Product) error
	GetProductByID(ctx context.Context, id uint) (*entity.Product, error)
	GetProductsByIDs(ctx context.Context, ids []uint) ([]entity.Product, error)
	UpdateProduct(ctx context.Context, id uint, product *entity.Product) error
	ListProducts(ctx context.Context, page, perPage int) ([]entity.Product, int, error)
	DeleteProduct(ctx context.Context, id uint) error
}

type ProductUsecaseInterface interface {
	CreateProduct(ctx context.Context, product *entity.Product) error
	GetProductByID(ctx context.Context, id uint) (*entity.Product, error)
	ListProducts(ctx context.Context, page, perPage int) ([]entity.Product, int, error)
	UpdateProduct(ctx context.Context, id uint, product *entity.Product) error
	DeleteProduct(ctx context.Context, id uint) error
	RestockProduct(ctx context.Context, id uint, quantity int) error
}

type UserUsecaseInterface interface {
	ListUsers(ctx context.Context, search string, page, perPage int) (*[]entity.User, int, error)
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUserByID(ctx context.Context, id uint) (*entity.User, error)
	UpdateUser(ctx context.Context, id uint, user *entity.User) error
	DeleteUser(ctx context.Context, id uint) error
}
