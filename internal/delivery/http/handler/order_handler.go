package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/POS/internal/delivery/http/helper"
	"github.com/kareemhamed001/POS/internal/delivery/http/request"
	"github.com/kareemhamed001/POS/internal/delivery/http/response"
	validation "github.com/kareemhamed001/POS/internal/delivery/http/validation"
	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type OrderHandler struct {
	orderUsecase *usecase.OrderUsecase
	validate     *validator.Validate
	tracer       trace.Tracer
}

func NewOrderHandler(orderUsecase *usecase.OrderUsecase, validate *validator.Validate) *OrderHandler {
	return &OrderHandler{
		orderUsecase: orderUsecase,
		validate:     validate,
		tracer:       otel.Tracer("order-handler"),
	}
}

func (h *OrderHandler) CreateOrder(ctx *gin.Context) {
	traceCtx, span := h.tracer.Start(ctx.Request.Context(), "OrderHandler.CreateOrder")
	defer span.End()

	// Parse and validate request
	_, parseSpan := h.tracer.Start(traceCtx, "ParseRequest")
	var req request.CreateOrderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		parseSpan.RecordError(err)
		parseSpan.SetStatus(codes.Error, err.Error())
		parseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		helper.WriteAPIResponse(ctx, nil, "invalid request body", http.StatusBadRequest)
		return
	}
	parseSpan.End()

	// Validate request
	_, validateSpan := h.tracer.Start(traceCtx, "ValidateRequest")
	if err := h.validate.Struct(req); err != nil {
		validateSpan.RecordError(err)
		validateSpan.SetStatus(codes.Error, err.Error())
		validateSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, validation.FormatValidationError(err), http.StatusBadRequest)
		return
	}
	validateSpan.SetAttributes(attribute.Int("order.items.count", len(req.Items)))
	validateSpan.End()

	// Build order entity
	_, buildSpan := h.tracer.Start(traceCtx, "BuildOrderEntity")
	orderItems := make([]entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		orderItems[i] = entity.OrderItem{
			ProductID:     item.ProductID,
			Quantity:      item.Quantity,
			DiscountType:  item.DiscountType,
			DiscountValue: item.DiscountValue,
		}
	}

	order := &entity.Order{
		UserID:        req.UserID,
		AddressID:     req.AddressID,
		ShippingCost:  req.ShippingCost,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		Items:         orderItems,
	}
	buildSpan.End()

	// Create order
	if err := h.orderUsecase.CreateOrder(traceCtx, order); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "Order created successfully")
	helper.WriteAPIResponse(ctx, gin.H{"order_id": order.ID}, "Order created successfully", http.StatusCreated)
}

func (h *OrderHandler) GetOrderByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, "invalid order id", http.StatusBadRequest)
		return
	}

	order, err := h.orderUsecase.GetOrder(ctx.Request.Context(), uint(id))
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusNotFound)
		return
	}

	// Manual conversion to response since we haven't added ToOrderResponse yet
	itemsResponse := make([]response.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		itemsResponse[i] = response.OrderItemResponse{
			ID:            item.ID,
			ProductID:     item.ProductID,
			Quantity:      item.Quantity,
			Price:         item.Price,
			SubTotal:      item.SubTotal,
			DiscountType:  string(item.DiscountType),
			DiscountValue: item.DiscountValue,
			Total:         item.Total,
		}
	}

	orderResponse := response.OrderResponse{
		ID:            order.ID,
		Status:        string(order.Status),
		UserID:        order.UserID,
		AddressID:     order.AddressID,
		SubTotal:      order.SubTotal,
		ShippingCost:  order.ShippingCost,
		DiscountType:  string(order.DiscountType),
		DiscountValue: order.DiscountValue,
		Total:         order.Total,
		Items:         itemsResponse,
	}

	helper.WriteAPIResponse(ctx, gin.H{"order": orderResponse}, "Order retrieved successfully", http.StatusOK)
}

func (h *OrderHandler) UpdateStatus(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, "invalid order id", http.StatusBadRequest)
		return
	}

	var req request.UpdateOrderStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.WriteAPIResponse(ctx, nil, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		helper.WriteAPIResponse(ctx, nil, validation.FormatValidationError(err), http.StatusBadRequest)
		return
	}

	if err := h.orderUsecase.UpdateStatus(ctx.Request.Context(), uint(id), req.Status); err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	helper.WriteAPIResponse(ctx, nil, "Order status updated successfully", http.StatusOK)
}
