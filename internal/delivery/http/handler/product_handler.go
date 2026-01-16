package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/POS/internal/delivery/http/helper"
	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/internal/usecase"
)

type ProductHandler struct {
	productUsecase usecase.ProductUsecaseInterface
	validate       *validator.Validate
}

func NewProductHandler(productUsecase usecase.ProductUsecaseInterface, validate *validator.Validate) *ProductHandler {
	return &ProductHandler{
		productUsecase: productUsecase,
		validate:       validate,
	}
}

func (h *ProductHandler) CreateProduct(ctx *gin.Context) {
	var product entity.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		helper.WriteAPIResponse(ctx, nil, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.productUsecase.CreateProduct(ctx.Request.Context(), &product); err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	helper.WriteAPIResponse(ctx, product, "Product created successfully", http.StatusCreated)
}

func (h *ProductHandler) ListProducts(ctx *gin.Context) {
	products, err := h.productUsecase.ListProducts(ctx.Request.Context())
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}
	helper.WriteAPIResponse(ctx, products, "Products retrieved successfully", http.StatusOK)
}

func (h *ProductHandler) GetProduct(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	product, err := h.productUsecase.GetProductByID(ctx.Request.Context(), uint(id))
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusNotFound)
		return
	}
	helper.WriteAPIResponse(ctx, product, "Product retrieved successfully", http.StatusOK)
}
