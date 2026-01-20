package postgres

import (
	"context"
	"errors"

	"github.com/kareemhamed001/POS/internal/entity"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrProductNotFound = errors.New("Product not found")
)

type ProductRepository struct {
	db     *gorm.DB
	tracer trace.Tracer
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		db:     db,
		tracer: otel.Tracer("product-repo"),
	}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product *entity.Product) error {
	ctx, span := r.tracer.Start(ctx, "ProductRepository.CreateProduct")
	defer span.End()

	span.SetAttributes(
		attribute.String("product.name", product.Name),
		attribute.Float64("product.price", float64(product.Price)),
	)

	if err := gorm.G[entity.Product](r.db).Create(ctx, product); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(attribute.Int("product.id", int(product.ID)))
	span.SetStatus(codes.Ok, "product created")
	return nil
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id uint) (*entity.Product, error) {
	ctx, span := r.tracer.Start(ctx, "ProductRepository.GetProductByID")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", int(id)))

	product, err := gorm.G[entity.Product](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, ErrProductNotFound.Error())
			return nil, ErrProductNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.String("product.name", product.Name))
	span.SetStatus(codes.Ok, "product retrieved")
	return &product, nil
}
func (r *ProductRepository) GetProductsByIDs(ctx context.Context, ids []uint) ([]entity.Product, error) {
	ctx, span := r.tracer.Start(ctx, "ProductRepository.GetProductsByIDs")
	defer span.End()

	span.SetAttributes(attribute.Int("product.ids.count", len(ids)))

	products, err := gorm.G[entity.Product](r.db).Where("id IN ?", ids).Find(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("products.count", len(products)))
	span.SetStatus(codes.Ok, "products retrieved")
	return products, nil
}
func (r *ProductRepository) UpdateProduct(ctx context.Context, id uint, product *entity.Product) error {
	ctx, span := r.tracer.Start(ctx, "ProductRepository.UpdateProduct")
	defer span.End()

	span.SetAttributes(
		attribute.Int("product.id", int(id)),
		attribute.String("product.name", product.Name),
	)

	rowsAffected, err := gorm.G[entity.Product](r.db).Where("id = ?", id).Updates(ctx, *product)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if rowsAffected == 0 {
		span.SetStatus(codes.Error, ErrProductNotFound.Error())
		return ErrProductNotFound
	}

	span.SetStatus(codes.Ok, "product updated")
	return nil
}

func (r *ProductRepository) ListProducts(ctx context.Context, page, perPage int) ([]entity.Product, int, error) {
	ctx, span := r.tracer.Start(ctx, "ProductRepository.ListProducts")
	defer span.End()

	span.SetAttributes(
		attribute.Int("query.page", page),
		attribute.Int("query.per_page", perPage),
	)

	products, err := gorm.G[entity.Product](r.db).Offset((page - 1) * perPage).Limit(perPage).Find(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}

	totalCount, err := gorm.G[entity.Product](r.db).Count(ctx, "*")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}

	span.SetAttributes(attribute.Int("products.count", len(products)))
	span.SetStatus(codes.Ok, "products listed")
	return products, int(totalCount), nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id uint) error {
	ctx, span := r.tracer.Start(ctx, "ProductRepository.DeleteProduct")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", int(id)))

	rowsAffected, err := gorm.G[entity.Product](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if rowsAffected == 0 {
		span.SetStatus(codes.Error, ErrProductNotFound.Error())
		return ErrProductNotFound
	}

	span.SetStatus(codes.Ok, "product deleted")
	return nil
}
