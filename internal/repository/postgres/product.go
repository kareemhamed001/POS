package postgres

import (
	"context"
	"errors"

	"github.com/kareemhamed001/POS/internal/entity"
	"gorm.io/gorm"
)

var (
	ErrProductNotFound = errors.New("Product not found")
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product *entity.Product) error {
	return gorm.G[entity.Product](r.db).Create(ctx, product)
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id uint) (*entity.Product, error) {
	product, err := gorm.G[entity.Product](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}
func (r *ProductRepository) GetProductsByIDs(ctx context.Context, ids []uint) ([]entity.Product, error) {
	products, err := gorm.G[entity.Product](r.db).Where("id IN ?", ids).Find(ctx)
	if err != nil {
		return nil, err
	}
	return products, nil
}
func (r *ProductRepository) UpdateProduct(ctx context.Context, id uint, product *entity.Product) error {
	rowsAffected, err := gorm.G[entity.Product](r.db).Where("id = ?", id).Updates(ctx, *product)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *ProductRepository) ListProducts(ctx context.Context, page, perPage int) ([]entity.Product, int, error) {
	products, err := gorm.G[entity.Product](r.db).Offset((page - 1) * perPage).Limit(perPage).Find(ctx)
	if err != nil {
		return nil, 0, err
	}

	totalCount, err := gorm.G[entity.Product](r.db).Count(ctx, "*")
	if err != nil {
		return nil, 0, err
	}
	return products, int(totalCount), nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id uint) error {
	rowsAffected, err := gorm.G[entity.Product](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}
