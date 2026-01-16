package usecase

import (
	"context"
	"errors"

	"github.com/kareemhamed001/POS/internal/entity"
)

type ProductUsecase struct {
	productRepo ProductRepository
}

func NewProductUsecase(productRepo ProductRepository) *ProductUsecase {
	return &ProductUsecase{
		productRepo: productRepo,
	}
}

func (u *ProductUsecase) CreateProduct(ctx context.Context, product *entity.Product) error {
	return u.productRepo.CreateProduct(ctx, product)
}

func (u *ProductUsecase) GetProductByID(ctx context.Context, id uint) (*entity.Product, error) {
	return u.productRepo.GetProductByID(ctx, id)
}

func (u *ProductUsecase) ListProducts(ctx context.Context) ([]entity.Product, error) {
	return u.productRepo.ListProducts(ctx)
}

func (u *ProductUsecase) UpdateProduct(ctx context.Context, id uint, product *entity.Product) error {
	return u.productRepo.UpdateProduct(ctx, id, product)
}

func (u *ProductUsecase) RestockProduct(ctx context.Context, id uint, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}

	product, err := u.productRepo.GetProductByID(ctx, id)
	if err != nil {
		return err
	}

	product.Quantity += quantity
	return u.productRepo.UpdateProduct(ctx, id, product)
}
