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
	"github.com/kareemhamed001/POS/pkg/logger"

	"github.com/kareemhamed001/POS/internal/usecase"
)

type UserHandler struct {
	userUsecase usecase.UserUsecaseInterface
	validate    *validator.Validate
	logger      *logger.Logger
}

func NewUserHandler(logger *logger.Logger, userUsecase usecase.UserUsecaseInterface, validate *validator.Validate) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
		validate:    validate,
		logger:      logger,
	}
}

func (u *UserHandler) ListUsers(ctx *gin.Context) {
	search := ctx.Query("search")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "10"))

	users, total, err := u.userUsecase.ListUsers(ctx.Request.Context(), search, page, perPage)
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	usersResponse := make([]response.UserResponse, 0, len(*users))

	for _, user := range *users {
		usersResponse = append(usersResponse, user.ToUserResponse())
	}
	helper.WriteAPIResponse(ctx, gin.H{"users": usersResponse, "total": total, "page": page}, "Users retrieved successfully", http.StatusOK)
}

func (u *UserHandler) CreateUser(ctx *gin.Context) {

	var request request.CreateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		helper.WriteAPIResponse(ctx, nil, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := u.validate.Struct(&request); err != nil {
		helper.WriteAPIResponse(ctx, nil, validation.FormatValidationError(err), http.StatusBadRequest)
		return
	}

	user := &entity.User{
		Name:     request.Name,
		Email:    request.Email,
		Phone:    request.Phone,
		Password: request.Password,
	}

	user, err := u.userUsecase.CreateUser(ctx.Request.Context(), user)
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	helper.WriteAPIResponse(ctx, gin.H{"user": user.ToUserResponse()}, "Users created successfully", http.StatusOK)

}

func (u *UserHandler) GetUserByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, "invalid user id", http.StatusBadRequest)
		return
	}
	user, err := u.userUsecase.GetUserByID(ctx, uint(id))
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	helper.WriteAPIResponse(ctx, gin.H{"user": user.ToUserResponse()}, "User retrieved successfully", http.StatusOK)
}

func (u *UserHandler) UpdateUser(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, "invalid user id", http.StatusBadRequest)
		return
	}

	var request request.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		helper.WriteAPIResponse(ctx, nil, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := u.validate.Struct(&request); err != nil {
		helper.WriteAPIResponse(ctx, nil, validation.FormatValidationError(err), http.StatusBadRequest)
		return
	}

	user := &entity.User{}
	if request.Name != nil {
		user.Name = *request.Name
	}
	if request.Email != nil {
		user.Email = *request.Email
	}
	if request.Phone != nil {
		user.Phone = *request.Phone
	}
	if request.Password != nil {
		user.Password = *request.Password
	}

	err = u.userUsecase.UpdateUser(ctx.Request.Context(), uint(id), user)
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	helper.WriteAPIResponse(ctx, nil, "User updated successfully", http.StatusNoContent)
}

func (u *UserHandler) DeleteUser(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, "invalid user id", http.StatusBadRequest)
		return
	}

	err = u.userUsecase.DeleteUser(ctx.Request.Context(), uint(id))
	if err != nil {
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	helper.WriteAPIResponse(ctx, nil, "User deleted successfully", http.StatusOK)
}
