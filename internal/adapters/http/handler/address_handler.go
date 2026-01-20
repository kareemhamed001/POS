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
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AddressHandler struct {
	addressUsecase *usecase.AddressUsecase
	validate       *validator.Validate
	tracer         trace.Tracer
}

func NewAddressHandler(addressUsecase *usecase.AddressUsecase, validate *validator.Validate) *AddressHandler {
	return &AddressHandler{
		addressUsecase: addressUsecase,
		validate:       validate,
		tracer:         otel.Tracer("address-handler"),
	}
}

func (h *AddressHandler) AddAddress(ctx *gin.Context) {
	spanCtx, span := h.tracer.Start(ctx.Request.Context(), "AddressHandler.AddAddress")
	defer span.End()

	_, bindSpan := h.tracer.Start(spanCtx, "Bind AddAddress Request Body")

	var req request.CreateAddressRequest
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

	_, validationSpan := h.tracer.Start(spanCtx, "Validate AddAddress Request")
	if err := h.validate.Struct(req); err != nil {
		validationSpan.RecordError(err)
		validationSpan.SetStatus(codes.Error, err.Error())
		validationSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, validation.FormatValidationError(err), http.StatusBadRequest)
		return
	}
	validationSpan.End()

	_, createAddressStructSpan := h.tracer.Start(spanCtx, "Create Address Struct")
	address := &entity.Address{
		UserID:      req.UserID,
		Name:        req.Name,
		Country:     req.Country,
		Governorate: req.Governorate,
		City:        req.City,
		Address:     req.Address,
		Phone:       req.Phone,
		IsPrimary:   req.IsPrimary,
	}
	createAddressStructSpan.End()

	addAddressSpanCtx, addAddressSpan := h.tracer.Start(spanCtx, "Add Address Usecase")
	if err := h.addressUsecase.AddAddress(addAddressSpanCtx, address); err != nil {
		addAddressSpan.RecordError(err)
		addAddressSpan.SetStatus(codes.Error, err.Error())
		addAddressSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}
	addAddressSpan.End()
	span.SetStatus(codes.Ok, "Address added successfully")
	helper.WriteAPIResponse(ctx, gin.H{"address_id": address.ID}, "Address added successfully", http.StatusCreated)
}

func (h *AddressHandler) GetUserAddresses(ctx *gin.Context) {
	spanCtx, span := h.tracer.Start(ctx.Request.Context(), "AddressHandler.GetUserAddresses")
	defer span.End()

	_, bindUserIdSpan := h.tracer.Start(spanCtx, "Bind UserID from URL Param")
	userID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		bindUserIdSpan.RecordError(err)
		bindUserIdSpan.SetStatus(codes.Error, err.Error())
		bindUserIdSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, "invalid user id", http.StatusBadRequest)
		return
	}
	bindUserIdSpan.End()

	getUserAddressesSpanCtx, getUserAddressesSpan := h.tracer.Start(spanCtx, "Get User Addresses Usecase")

	addresses, err := h.addressUsecase.GetUserAddresses(getUserAddressesSpanCtx, uint(userID))
	if err != nil {
		getUserAddressesSpan.RecordError(err)
		getUserAddressesSpan.SetStatus(codes.Error, err.Error())
		getUserAddressesSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}
	getUserAddressesSpan.End()

	_, mapAddressesSpan := h.tracer.Start(spanCtx, "Map Addresses to Response Structs")
	addressResponses := make([]response.AddressResponse, len(addresses))
	for i, addr := range addresses {
		addressResponses[i] = response.AddressResponse{
			ID:          addr.ID,
			UserID:      addr.UserID,
			Name:        addr.Name,
			Country:     addr.Country,
			Governorate: addr.Governorate,
			City:        addr.City,
			Address:     addr.Address,
			Phone:       addr.Phone,
			IsPrimary:   addr.IsPrimary,
		}
	}
	mapAddressesSpan.End()
	span.SetStatus(codes.Ok, "Addresses retrieved successfully")
	helper.WriteAPIResponse(ctx, gin.H{"addresses": addressResponses}, "Addresses retrieved successfully", http.StatusOK)
}
