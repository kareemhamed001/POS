package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/POS/internal/adapters/http/helper"
	"github.com/kareemhamed001/POS/internal/adapters/http/request"
	"github.com/kareemhamed001/POS/internal/adapters/http/response"
	validation "github.com/kareemhamed001/POS/internal/adapters/http/validation"
	entity "github.com/kareemhamed001/POS/internal/core/domain"
	"github.com/kareemhamed001/POS/internal/core/usecase"
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
	traceCtx, span := h.tracer.Start(ctx.Request.Context(), "OrderHandler.GetOrderByID")
	defer span.End()

	_, parseSpan := h.tracer.Start(traceCtx, "ParseOrderID")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.SetStatus(codes.Error, err.Error())
		parseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, "invalid order id", http.StatusBadRequest)
		return
	}
	parseSpan.SetAttributes(attribute.Int("order.id", id))
	parseSpan.End()

	usecaseCtx, usecaseSpan := h.tracer.Start(traceCtx, "GetOrderUsecase")
	order, err := h.orderUsecase.GetOrder(usecaseCtx, uint(id))
	if err != nil {
		usecaseSpan.RecordError(err)
		usecaseSpan.SetStatus(codes.Error, err.Error())
		usecaseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusNotFound)
		return
	}
	usecaseSpan.End()

	_, mapSpan := h.tracer.Start(traceCtx, "MapOrderResponse")
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
	mapSpan.End()

	span.SetStatus(codes.Ok, "Order retrieved successfully")
	helper.WriteAPIResponse(ctx, gin.H{"order": orderResponse}, "Order retrieved successfully", http.StatusOK)
}

func (h *OrderHandler) UpdateStatus(ctx *gin.Context) {
	traceCtx, span := h.tracer.Start(ctx.Request.Context(), "OrderHandler.UpdateStatus")
	defer span.End()

	_, parseSpan := h.tracer.Start(traceCtx, "ParseOrderID")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.SetStatus(codes.Error, err.Error())
		parseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, "invalid order id", http.StatusBadRequest)
		return
	}
	parseSpan.SetAttributes(attribute.Int("order.id", id))
	parseSpan.End()

	_, bindSpan := h.tracer.Start(traceCtx, "BindUpdateOrderStatusRequest")
	var req request.UpdateOrderStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		bindSpan.RecordError(err)
		bindSpan.SetStatus(codes.Error, err.Error())
		bindSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, "invalid request body", http.StatusBadRequest)
		return
	}
	bindSpan.End()

	_, validateSpan := h.tracer.Start(traceCtx, "ValidateUpdateOrderStatusRequest")
	if err := h.validate.Struct(req); err != nil {
		validateSpan.RecordError(err)
		validateSpan.SetStatus(codes.Error, err.Error())
		validateSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, validation.FormatValidationError(err), http.StatusBadRequest)
		return
	}
	validateSpan.End()

	usecaseCtx, usecaseSpan := h.tracer.Start(traceCtx, "UpdateOrderStatusUsecase")
	if err := h.orderUsecase.UpdateStatus(usecaseCtx, uint(id), req.Status); err != nil {
		usecaseSpan.RecordError(err)
		usecaseSpan.SetStatus(codes.Error, err.Error())
		usecaseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}
	usecaseSpan.End()

	span.SetStatus(codes.Ok, "Order status updated successfully")
	helper.WriteAPIResponse(ctx, nil, "Order status updated successfully", http.StatusOK)
}
