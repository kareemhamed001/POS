package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	validation "github.com/kareemhamed001/POS/internal/delivery/http/validation"
	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/pkg/file"
)

type mockProductUsecase struct {
	createProductFn  func(ctx context.Context, product *entity.Product) error
	getProductByIDFn func(ctx context.Context, id uint) (*entity.Product, error)
	listProductsFn   func(ctx context.Context, page, perPage int) ([]entity.Product, int, error)
	updateProductFn  func(ctx context.Context, id uint, product *entity.Product) error
	restockProductFn func(ctx context.Context, id uint, quantity int) error
	deleteProductFn  func(ctx context.Context, id uint) error
}

func (m *mockProductUsecase) CreateProduct(ctx context.Context, product *entity.Product) error {
	return m.createProductFn(ctx, product)
}
func (m *mockProductUsecase) GetProductByID(ctx context.Context, id uint) (*entity.Product, error) {
	return m.getProductByIDFn(ctx, id)
}
func (m *mockProductUsecase) ListProducts(ctx context.Context, page, perPage int) ([]entity.Product, int, error) {
	return m.listProductsFn(ctx, page, perPage)
}
func (m *mockProductUsecase) UpdateProduct(ctx context.Context, id uint, product *entity.Product) error {
	return m.updateProductFn(ctx, id, product)
}
func (m *mockProductUsecase) RestockProduct(ctx context.Context, id uint, quantity int) error {
	return m.restockProductFn(ctx, id, quantity)
}
func (m *mockProductUsecase) DeleteProduct(ctx context.Context, id uint) error {
	return m.deleteProductFn(ctx, id)
}

func TestProductHandler_CreateProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUsecase := &mockProductUsecase{}
	v := validation.NewValidator()
	mockFileStorage := file.NewLocalStorage(file.StorageConfig{LocalBasePath: "test_uploads", LocalBaseURL: "/uploads"})
	handler := NewProductHandler(mockUsecase, v, mockFileStorage)

	t.Run("success", func(t *testing.T) {
		product := entity.Product{Name: "Test Product"}
		body, _ := json.Marshal(product)

		req, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		mockUsecase.createProductFn = func(ctx context.Context, p *entity.Product) error {
			p.ID = 1
			return nil
		}

		handler.CreateProduct(ctx)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.CreateProduct(ctx)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})
}

func TestProductHandler_GetProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUsecase := &mockProductUsecase{}
	v := validation.NewValidator()
	mockFileStorage := file.NewLocalStorage(file.StorageConfig{LocalBasePath: "test_uploads", LocalBaseURL: "/uploads"})
	handler := NewProductHandler(mockUsecase, v, mockFileStorage)

	t.Run("success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/products/1", nil)
		w := httptest.NewRecorder()
		_, engine := gin.CreateTestContext(w)
		engine.GET("/products/:id", handler.GetProduct)

		mockUsecase.getProductByIDFn = func(ctx context.Context, id uint) (*entity.Product, error) {
			return &entity.Product{Name: "Test Product"}, nil
		}

		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})
}
