package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/kareemhamed001/POS/internal/cache"
	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	productCacheTTL     = 30 * time.Minute
	productListCacheTTL = 1 * time.Hour
)

type ProductUsecase struct {
	productRepo  ProductRepository
	productCache cache.ProductCache
	tracer       trace.Tracer
}

func NewProductUsecase(productRepo ProductRepository, productCache cache.ProductCache) *ProductUsecase {
	return &ProductUsecase{
		productRepo:  productRepo,
		productCache: productCache,
		tracer:       otel.Tracer("product-usecase"),
	}
}

func (u *ProductUsecase) CreateProduct(ctx context.Context, product *entity.Product) error {
	ctx, span := u.tracer.Start(ctx, "ProductUsecase.CreateProduct")
	defer span.End()

	span.SetAttributes(
		attribute.String("product.name", product.Name),
		attribute.Float64("product.price", float64(product.Price)),
		attribute.Int("product.quantity", product.Quantity),
	)

	// Database insert span
	_, dbSpan := u.tracer.Start(ctx, "Database.CreateProduct")
	if err := u.productRepo.CreateProduct(ctx, product); err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	dbSpan.SetAttributes(attribute.Int("product.id", int(product.ID)))
	dbSpan.End()

	// Cache invalidation span
	_, cacheSpan := u.tracer.Start(ctx, "Cache.InvalidateProductList")
	if err := u.productCache.InvalidateProductList(ctx); err != nil {
		cacheSpan.RecordError(err)
		logger.Warnf("Failed to invalidate product list cache: %v", err)
	}
	cacheSpan.End()

	span.SetStatus(codes.Ok, "Product created successfully")
	return nil
}

func (u *ProductUsecase) GetProductByID(ctx context.Context, id uint) (*entity.Product, error) {
	ctx, span := u.tracer.Start(ctx, "ProductUsecase.GetProductByID")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", int(id)))

	// Try cache first
	_, cacheSpan := u.tracer.Start(ctx, "Cache.GetProduct")
	product, err := u.productCache.GetProduct(ctx, id)
	if err == nil {
		cacheSpan.SetAttributes(attribute.Bool("cache.hit", true))
		cacheSpan.End()
		logger.Debug("Product cache hit")
		span.SetAttributes(
			attribute.Bool("cache.hit", true),
			attribute.String("product.name", product.Name),
		)
		span.SetStatus(codes.Ok, "Product found in cache")
		return product, nil
	}
	cacheSpan.SetAttributes(attribute.Bool("cache.hit", false))
	cacheSpan.End()

	// Cache miss - fetch from DB
	logger.Debug("Product cache miss, fetching from DB")
	_, dbSpan := u.tracer.Start(ctx, "Database.GetProductByID")
	product, err = u.productRepo.GetProductByID(ctx, id)
	if err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	dbSpan.End()

	// Store in cache (fire and forget)
	_, setCacheSpan := u.tracer.Start(ctx, "Cache.SetProduct")
	if err := u.productCache.SetProduct(ctx, product, productCacheTTL); err != nil {
		setCacheSpan.RecordError(err)
		logger.Warnf("Failed to cache product: %v", err)
	}
	setCacheSpan.End()

	span.SetAttributes(
		attribute.Bool("cache.hit", false),
		attribute.String("product.name", product.Name),
	)
	span.SetStatus(codes.Ok, "Product retrieved from database")
	return product, nil
}

func (u *ProductUsecase) ListProducts(ctx context.Context) ([]entity.Product, error) {
	ctx, span := u.tracer.Start(ctx, "ProductUsecase.ListProducts")
	defer span.End()

	// Try cache first
	_, cacheSpan := u.tracer.Start(ctx, "Cache.GetProductList")
	products, err := u.productCache.GetProductList(ctx, "all")
	if err == nil {
		cacheSpan.SetAttributes(
			attribute.Bool("cache.hit", true),
			attribute.Int("products.count", len(products)),
		)
		cacheSpan.End()
		logger.Debug("Product list cache hit")
		span.SetAttributes(
			attribute.Bool("cache.hit", true),
			attribute.Int("products.count", len(products)),
		)
		span.SetStatus(codes.Ok, "Products found in cache")
		return products, nil
	}
	cacheSpan.SetAttributes(attribute.Bool("cache.hit", false))
	cacheSpan.End()

	// Cache miss - fetch from DB
	logger.Debug("Product list cache miss, fetching from DB")
	_, dbSpan := u.tracer.Start(ctx, "Database.ListProducts")
	products, err = u.productRepo.ListProducts(ctx)
	if err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	dbSpan.SetAttributes(attribute.Int("products.count", len(products)))
	dbSpan.End()

	// Store in cache (fire and forget)
	_, setCacheSpan := u.tracer.Start(ctx, "Cache.SetProductList")
	if err := u.productCache.SetProductList(ctx, "all", products, productListCacheTTL); err != nil {
		setCacheSpan.RecordError(err)
		logger.Warnf("Failed to cache product list: %v", err)
	}
	setCacheSpan.End()

	span.SetAttributes(
		attribute.Bool("cache.hit", false),
		attribute.Int("products.count", len(products)),
	)
	span.SetStatus(codes.Ok, "Products retrieved from database")
	return products, nil
}

func (u *ProductUsecase) UpdateProduct(ctx context.Context, id uint, product *entity.Product) error {
	ctx, span := u.tracer.Start(ctx, "ProductUsecase.UpdateProduct")
	defer span.End()

	span.SetAttributes(
		attribute.Int("product.id", int(id)),
		attribute.String("product.name", product.Name),
		attribute.Float64("product.price", float64(product.Price)),
	)

	// Database update span
	_, dbSpan := u.tracer.Start(ctx, "Database.UpdateProduct")
	if err := u.productRepo.UpdateProduct(ctx, id, product); err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	dbSpan.End()

	// Cache invalidation spans
	_, deleteSpan := u.tracer.Start(ctx, "Cache.DeleteProduct")
	if err := u.productCache.DeleteProduct(ctx, id); err != nil {
		deleteSpan.RecordError(err)
		logger.Warnf("Failed to delete product from cache: %v", err)
	}
	deleteSpan.End()

	_, invalidateSpan := u.tracer.Start(ctx, "Cache.InvalidateProductList")
	if err := u.productCache.InvalidateProductList(ctx); err != nil {
		invalidateSpan.RecordError(err)
		logger.Warnf("Failed to invalidate product list cache: %v", err)
	}
	invalidateSpan.End()

	span.SetStatus(codes.Ok, "Product updated successfully")
	return nil
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

func (u *ProductUsecase) DeleteProduct(ctx context.Context, id uint) error {
	ctx, span := u.tracer.Start(ctx, "ProductUsecase.DeleteProduct")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", int(id)))

	// Database delete span
	_, dbSpan := u.tracer.Start(ctx, "Database.DeleteProduct")
	if err := u.productRepo.DeleteProduct(ctx, id); err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	dbSpan.End()

	// Cache invalidation spans
	_, deleteSpan := u.tracer.Start(ctx, "Cache.DeleteProduct")
	if err := u.productCache.DeleteProduct(ctx, id); err != nil {
		deleteSpan.RecordError(err)
		logger.Warnf("Failed to delete product from cache: %v", err)
	}
	deleteSpan.End()

	_, invalidateSpan := u.tracer.Start(ctx, "Cache.InvalidateProductList")
	if err := u.productCache.InvalidateProductList(ctx); err != nil {
		invalidateSpan.RecordError(err)
		logger.Warnf("Failed to invalidate product list cache: %v", err)
	}
	invalidateSpan.End()

	span.SetStatus(codes.Ok, "Product deleted successfully")
	return nil
}
