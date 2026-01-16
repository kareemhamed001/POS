package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/POS/internal/delivery/http/helper"
	"github.com/kareemhamed001/POS/internal/delivery/http/request"
	"github.com/kareemhamed001/POS/internal/delivery/http/response"
	validation "github.com/kareemhamed001/POS/internal/delivery/http/validator"
	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/internal/usecase"
)

type AddressHandler struct {
	addressUsecase *usecase.AddressUsecase
	validate       *validator.Validate
}

func NewAddressHandler(addressUsecase *usecase.AddressUsecase, validate *validator.Validate) *AddressHandler {
	return &AddressHandler{
		addressUsecase: addressUsecase,
		validate:       validate,
	}
}

func (h *AddressHandler) AddAddress(ctx *gin.Context) {
	var req request.CreateAddressRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.WriteAPIResponse(ctx, nil, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		helper.WriteAPIResponse(ctx, nil, validation.FormatValidationError(err), http.StatusBadRequest)
		return
	}

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

	if err := h.addressUsecase.AddAddress(ctx.Request.Context(), address); err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	helper.WriteAPIResponse(ctx, gin.H{"address_id": address.ID}, "Address added successfully", http.StatusCreated)
}

func (h *AddressHandler) GetUserAddresses(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("user_id"))
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, "invalid user id", http.StatusBadRequest)
		return
	}

	addresses, err := h.addressUsecase.GetUserAddresses(ctx.Request.Context(), uint(userID))
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

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

	helper.WriteAPIResponse(ctx, gin.H{"addresses": addressResponses}, "Addresses retrieved successfully", http.StatusOK)
}
