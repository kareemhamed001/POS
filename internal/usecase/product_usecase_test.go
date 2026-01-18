package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/kareemhamed001/POS/internal/entity"
)

type mockProductRepository struct {
	createProductFn  func(ctx context.Context, product *entity.Product) error
	getProductByIDFn func(ctx context.Context, id uint) (*entity.Product, error)
	listProductsFn   func(ctx context.Context, page, perPage int) ([]entity.Product, int, error)
	updateProductFn  func(ctx context.Context, id uint, product *entity.Product) error
	deleteProductFn  func(ctx context.Context, id uint) error
}

func (m *mockProductRepository) CreateProduct(ctx context.Context, product *entity.Product) error {
	return m.createProductFn(ctx, product)
}

func (m *mockProductRepository) GetProductByID(ctx context.Context, id uint) (*entity.Product, error) {
	return m.getProductByIDFn(ctx, id)
}

func (m *mockProductRepository) GetProductsByIDs(ctx context.Context, ids []uint) ([]entity.Product, error) {
	return nil, nil // Not needed for product usecase tests
}

func (m *mockProductRepository) ListProducts(ctx context.Context, page, perPage int) ([]entity.Product, int, error) {
	return m.listProductsFn(ctx, page, perPage)
}

func (m *mockProductRepository) UpdateProduct(ctx context.Context, id uint, product *entity.Product) error {
	return m.updateProductFn(ctx, id, product)
}

func (m *mockProductRepository) DeleteProduct(ctx context.Context, id uint) error {
	return m.deleteProductFn(ctx, id)
}

func TestProductUsecase_CreateProduct(t *testing.T) {
	mockRepo := &mockProductRepository{}
	usecase := NewProductUsecase(mockRepo, nil) // nil cache for tests
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		product := &entity.Product{Name: "Test Product"}
		mockRepo.createProductFn = func(ctx context.Context, p *entity.Product) error {
			p.ID = 1
			return nil
		}

		err := usecase.CreateProduct(ctx, product)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if product.ID != 1 {
			t.Fatalf("expected product ID 1, got %d", product.ID)
		}
	})

	t.Run("failure", func(t *testing.T) {
		product := &entity.Product{Name: "Test Product"}
		expectedErr := errors.New("db error")
		mockRepo.createProductFn = func(ctx context.Context, p *entity.Product) error {
			return expectedErr
		}

		err := usecase.CreateProduct(ctx, product)
		if err != expectedErr {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestProductUsecase_GetProductByID(t *testing.T) {
	mockRepo := &mockProductRepository{}
	usecase := NewProductUsecase(mockRepo, nil) // nil cache for tests
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedProduct := &entity.Product{Name: "Test Product"}
		mockRepo.getProductByIDFn = func(ctx context.Context, id uint) (*entity.Product, error) {
			return expectedProduct, nil
		}

		product, err := usecase.GetProductByID(ctx, 1)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if product != expectedProduct {
			t.Fatalf("expected product %v, got %v", expectedProduct, product)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.getProductByIDFn = func(ctx context.Context, id uint) (*entity.Product, error) {
			return nil, errors.New("not found")
		}

		product, err := usecase.GetProductByID(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if product != nil {
			t.Fatalf("expected nil product, got %v", product)
		}
	})
}

func TestProductUsecase_ListProducts(t *testing.T) {
	mockRepo := &mockProductRepository{}
	usecase := NewProductUsecase(mockRepo, nil) // nil cache for tests
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedProducts := []entity.Product{{Name: "P1"}, {Name: "P2"}}
		mockRepo.listProductsFn = func(ctx context.Context, page, perPage int) ([]entity.Product, int, error) {
			return expectedProducts, 2, nil
		}
		page := 1
		perPage := 10

		products, total, err := usecase.ListProducts(ctx, page, perPage)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if total != 2 {
			t.Fatalf("expected 2 products, got %d", total)
		}
		if len(products) != len(expectedProducts) {
			t.Fatalf("expected products %v, got %v", expectedProducts, products)
		}
	})
}

func TestProductUsecase_UpdateProduct(t *testing.T) {
	mockRepo := &mockProductRepository{}
	usecase := NewProductUsecase(mockRepo, nil) // nil cache for tests
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		product := &entity.Product{Name: "Updated Name"}
		mockRepo.updateProductFn = func(ctx context.Context, id uint, p *entity.Product) error {
			return nil
		}

		err := usecase.UpdateProduct(ctx, 1, product)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}

func TestProductUsecase_RestockProduct(t *testing.T) {
	mockRepo := &mockProductRepository{}
	usecase := NewProductUsecase(mockRepo, nil) // nil cache for tests
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		product := &entity.Product{Quantity: 10}
		mockRepo.getProductByIDFn = func(ctx context.Context, id uint) (*entity.Product, error) {
			return product, nil
		}
		mockRepo.updateProductFn = func(ctx context.Context, id uint, p *entity.Product) error {
			return nil
		}

		err := usecase.RestockProduct(ctx, 1, 5)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if product.Quantity != 15 {
			t.Fatalf("expected quantity 15, got %d", product.Quantity)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		err := usecase.RestockProduct(ctx, 1, 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
