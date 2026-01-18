package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/POS/internal/delivery/http/helper"
	"github.com/kareemhamed001/POS/internal/delivery/http/request"
	"github.com/kareemhamed001/POS/internal/delivery/http/validation"
	"github.com/kareemhamed001/POS/internal/usecase"
	"github.com/kareemhamed001/POS/pkg/file"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ProductHandler struct {
	productUsecase usecase.ProductUsecaseInterface
	validate       *validator.Validate
	fileStorage    file.FileStorage
	tracer         trace.Tracer
}

func NewProductHandler(productUsecase usecase.ProductUsecaseInterface, validate *validator.Validate, fileStorage file.FileStorage) *ProductHandler {
	return &ProductHandler{
		productUsecase: productUsecase,
		validate:       validate,
		fileStorage:    fileStorage,
		tracer:         otel.Tracer("product-handler"),
	}
}

func (h *ProductHandler) ListProducts(ctx *gin.Context) {
	reqCtx, span := h.tracer.Start(ctx.Request.Context(), "ProductHandler.ListProducts")
	defer span.End()

	products, err := h.productUsecase.ListProducts(reqCtx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("products.count", len(products)))
	span.SetStatus(codes.Ok, "Products retrieved successfully")
	helper.WriteAPIResponse(ctx, products, "Products retrieved successfully", http.StatusOK)
}

func (h *ProductHandler) GetProduct(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	reqCtx, span := h.tracer.Start(ctx.Request.Context(), "ProductHandler.GetProduct")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", id))

	product, err := h.productUsecase.GetProductByID(reqCtx, uint(id))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusNotFound)
		return
	}

	span.SetAttributes(
		attribute.String("product.name", product.Name),
		attribute.Float64("product.price", float64(product.Price)),
	)
	span.SetStatus(codes.Ok, "Product retrieved successfully")
	helper.WriteAPIResponse(ctx, product, "Product retrieved successfully", http.StatusOK)
}

func (h *ProductHandler) CreateProduct(ctx *gin.Context) {
	reqCtx, span := h.tracer.Start(ctx.Request.Context(), "ProductHandler.CreateProduct")
	defer span.End()

	var productRequest request.CreateProductRequest
	// Support multipart/form-data for file upload and regular form fields
	if err := ctx.ShouldBind(&productRequest); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "binding failed")
		helper.WriteAPIResponse(ctx, nil, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validation span
	_, validationSpan := h.tracer.Start(reqCtx, "ProductHandler.ValidateProduct")
	if err := h.validate.Struct(&productRequest); err != nil {
		validationSpan.RecordError(err)
		validationSpan.SetStatus(codes.Error, "validation failed")
		validationSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		validationErrMsg := validation.FormatValidationError(err)
		helper.WriteAPIResponse(ctx, nil, validationErrMsg, http.StatusBadRequest)
		return
	}
	validationSpan.End()

	// File processing span
	_, fileSpan := h.tracer.Start(reqCtx, "ProductHandler.ProcessProductData")
	product, err := productRequest.ToProduct(reqCtx, h.fileStorage)
	if err != nil {
		fileSpan.RecordError(err)
		fileSpan.SetStatus(codes.Error, "file processing failed")
		fileSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, "file processing failed")
		helper.WriteAPIResponse(ctx, nil, "failed to process product data: "+err.Error(), http.StatusBadRequest)
		return
	}
	fileSpan.End()

	span.SetAttributes(
		attribute.String("product.name", product.Name),
		attribute.Float64("product.price", float64(product.Price)),
		attribute.String("product.discount_type", string(product.DiscountType)),
	)

	if err := h.productUsecase.CreateProduct(reqCtx, product); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("product.id", int(product.ID)))
	span.SetStatus(codes.Ok, "Product created successfully")
	helper.WriteAPIResponse(ctx, product, "Product created successfully", http.StatusCreated)
}

func (h *ProductHandler) UpdateProduct(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	reqCtx, span := h.tracer.Start(ctx.Request.Context(), "ProductHandler.UpdateProduct")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", id))

	var productRequest request.UpdateProductRequest
	// Support multipart/form-data for file upload and regular form fields
	if err := ctx.ShouldBind(&productRequest); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "binding failed")
		helper.WriteAPIResponse(ctx, nil, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validation span
	_, validationSpan := h.tracer.Start(reqCtx, "ProductHandler.ValidateUpdateProduct")
	if err := h.validate.Struct(&productRequest); err != nil {
		validationSpan.RecordError(err)
		validationSpan.SetStatus(codes.Error, "validation failed")
		validationSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		validationErrMsg := validation.FormatValidationError(err)
		helper.WriteAPIResponse(ctx, nil, validationErrMsg, http.StatusBadRequest)
		return
	}
	validationSpan.End()

	// Get existing product
	existingProduct, err := h.productUsecase.GetProductByID(reqCtx, uint(id))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "product not found")
		helper.WriteAPIResponse(ctx, nil, "product not found", http.StatusNotFound)
		return
	}

	// File processing span (if image is being updated)
	_, fileSpan := h.tracer.Start(reqCtx, "ProductHandler.ProcessUpdateProductData")
	product, err := productRequest.ToProduct(existingProduct, reqCtx, h.fileStorage)
	if err != nil {
		fileSpan.RecordError(err)
		fileSpan.SetStatus(codes.Error, "file processing failed")
		fileSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, "file processing failed")
		helper.WriteAPIResponse(ctx, nil, "failed to process product data: "+err.Error(), http.StatusBadRequest)
		return
	}
	fileSpan.SetAttributes(attribute.Bool("image.updated", productRequest.Image != nil))
	fileSpan.End()

	span.SetAttributes(
		attribute.String("product.name", product.Name),
		attribute.Float64("product.price", float64(product.Price)),
	)

	if err := h.productUsecase.UpdateProduct(reqCtx, uint(id), product); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "Product updated successfully")
	helper.WriteAPIResponse(ctx, product, "Product updated successfully", http.StatusOK)
}

func (h *ProductHandler) DeleteProduct(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	reqCtx, span := h.tracer.Start(ctx.Request.Context(), "ProductHandler.DeleteProduct")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", id))

	if err := h.productUsecase.DeleteProduct(reqCtx, uint(id)); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "Product deleted successfully")
	helper.WriteAPIResponse(ctx, nil, "Product deleted successfully", http.StatusOK)
}
