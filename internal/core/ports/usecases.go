package ports

import (
	"context"

	entity "github.com/kareemhamed001/POS/internal/core/domain"
)

type ProductUsecase interface {
	CreateProduct(ctx context.Context, product *entity.Product) error
	GetProductByID(ctx context.Context, id uint) (*entity.Product, error)
	ListProducts(ctx context.Context, page, perPage int) ([]entity.Product, int, error)
	UpdateProduct(ctx context.Context, id uint, product *entity.Product) error
	DeleteProduct(ctx context.Context, id uint) error
	RestockProduct(ctx context.Context, id uint, quantity int) error
}

type UserUsecase interface {
	ListUsers(ctx context.Context, search string, page, perPage int) (*[]entity.User, int, error)
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUserByID(ctx context.Context, id uint) (*entity.User, error)
	UpdateUser(ctx context.Context, id uint, user *entity.User) error
	DeleteUser(ctx context.Context, id uint) error
}
