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

type UserHandler struct {
	userUsecase usecase.UserUsecaseInterface
	validate    *validator.Validate
	tracer      trace.Tracer
}

func NewUserHandler(userUsecase usecase.UserUsecaseInterface, validate *validator.Validate) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
		validate:    validate,
		tracer:      otel.Tracer("user-handler"),
	}
}

func (u *UserHandler) ListUsers(ctx *gin.Context) {
	traceCtx, span := u.tracer.Start(ctx.Request.Context(), "UserHandler.ListUsers")
	defer span.End()

	_, parseSpan := u.tracer.Start(traceCtx, "UserHandler.ParseListUsersQuery")
	search := ctx.Query("search")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "10"))
	parseSpan.SetAttributes(
		attribute.String("query.search", search),
		attribute.Int("query.page", page),
		attribute.Int("query.per_page", perPage),
	)
	parseSpan.End()

	usecaseCtx, usecaseSpan := u.tracer.Start(traceCtx, "UserHandler.ListUsersUsecase")
	users, total, err := u.userUsecase.ListUsers(usecaseCtx, search, page, perPage)
	if err != nil {
		usecaseSpan.RecordError(err)
		usecaseSpan.SetStatus(codes.Error, err.Error())
		usecaseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}
	usecaseSpan.End()

	_, mapSpan := u.tracer.Start(traceCtx, "UserHandler.MapUsersResponse")
	usersResponse := make([]response.UserResponse, 0, len(*users))

	for _, user := range *users {
		usersResponse = append(usersResponse, user.ToUserResponse())
	}
	mapSpan.End()
	span.SetStatus(codes.Ok, "Users retrieved successfully")
	helper.WriteAPIResponse(ctx, gin.H{"users": usersResponse, "total": total, "page": page}, "Users retrieved successfully", http.StatusOK)
}

func (u *UserHandler) CreateUser(ctx *gin.Context) {
	traceCtx, span := u.tracer.Start(ctx.Request.Context(), "UserHandler.CreateUser")
	defer span.End()

	_, bindSpan := u.tracer.Start(traceCtx, "UserHandler.BindCreateUserRequest")
	var request request.CreateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		bindSpan.RecordError(err)
		bindSpan.SetStatus(codes.Error, err.Error())
		bindSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, "invalid request body", http.StatusBadRequest)
		return
	}
	bindSpan.End()

	_, validateSpan := u.tracer.Start(traceCtx, "UserHandler.ValidateCreateUserRequest")
	if err := u.validate.Struct(&request); err != nil {
		validateSpan.RecordError(err)
		validateSpan.SetStatus(codes.Error, err.Error())
		validateSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, validation.FormatValidationError(err), http.StatusBadRequest)
		return
	}
	validateSpan.End()

	_, buildSpan := u.tracer.Start(traceCtx, "UserHandler.BuildUserEntity")
	user := &entity.User{
		Name:     request.Name,
		Email:    request.Email,
		Phone:    request.Phone,
		Password: request.Password,
	}
	buildSpan.End()

	usecaseCtx, usecaseSpan := u.tracer.Start(traceCtx, "UserHandler.CreateUserUsecase")
	user, err := u.userUsecase.CreateUser(usecaseCtx, user)
	if err != nil {
		usecaseSpan.RecordError(err)
		usecaseSpan.SetStatus(codes.Error, err.Error())
		usecaseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}
	usecaseSpan.End()

	span.SetStatus(codes.Ok, "Users created successfully")
	helper.WriteAPIResponse(ctx, gin.H{"user": user.ToUserResponse()}, "Users created successfully", http.StatusOK)

}

func (u *UserHandler) GetUserByID(ctx *gin.Context) {
	traceCtx, span := u.tracer.Start(ctx.Request.Context(), "UserHandler.GetUserByID")
	defer span.End()

	_, parseSpan := u.tracer.Start(traceCtx, "UserHandler.ParseUserID")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.SetStatus(codes.Error, err.Error())
		parseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, "invalid user id", http.StatusBadRequest)
		return
	}
	parseSpan.SetAttributes(attribute.Int("user.id", id))
	parseSpan.End()

	usecaseCtx, usecaseSpan := u.tracer.Start(traceCtx, "UserHandler.GetUserByIDUsecase")
	user, err := u.userUsecase.GetUserByID(usecaseCtx, uint(id))
	if err != nil {
		usecaseSpan.RecordError(err)
		usecaseSpan.SetStatus(codes.Error, err.Error())
		usecaseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}
	usecaseSpan.End()

	span.SetStatus(codes.Ok, "User retrieved successfully")
	helper.WriteAPIResponse(ctx, gin.H{"user": user.ToUserResponse()}, "User retrieved successfully", http.StatusOK)
}

func (u *UserHandler) UpdateUser(ctx *gin.Context) {
	traceCtx, span := u.tracer.Start(ctx.Request.Context(), "UserHandler.UpdateUser")
	defer span.End()

	_, parseSpan := u.tracer.Start(traceCtx, "UserHandler.ParseUserID")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.SetStatus(codes.Error, err.Error())
		parseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, "invalid user id", http.StatusBadRequest)
		return
	}
	parseSpan.SetAttributes(attribute.Int("user.id", id))
	parseSpan.End()

	_, bindSpan := u.tracer.Start(traceCtx, "UserHandler.BindUpdateUserRequest")
	var request request.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		bindSpan.RecordError(err)
		bindSpan.SetStatus(codes.Error, err.Error())
		bindSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, "invalid request body", http.StatusBadRequest)
		return
	}
	bindSpan.End()

	_, validateSpan := u.tracer.Start(traceCtx, "UserHandler.ValidateUpdateUserRequest")
	if err := u.validate.Struct(&request); err != nil {
		validateSpan.RecordError(err)
		validateSpan.SetStatus(codes.Error, err.Error())
		validateSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, validation.FormatValidationError(err), http.StatusBadRequest)
		return
	}
	validateSpan.End()

	_, buildSpan := u.tracer.Start(traceCtx, "UserHandler.BuildUserUpdateEntity")
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
	buildSpan.End()

	usecaseCtx, usecaseSpan := u.tracer.Start(traceCtx, "UserHandler.UpdateUserUsecase")
	err = u.userUsecase.UpdateUser(usecaseCtx, uint(id), user)
	if err != nil {
		usecaseSpan.RecordError(err)
		usecaseSpan.SetStatus(codes.Error, err.Error())
		usecaseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}
	usecaseSpan.End()

	span.SetStatus(codes.Ok, "User updated successfully")
	helper.WriteAPIResponse(ctx, nil, "User updated successfully", http.StatusNoContent)
}

func (u *UserHandler) DeleteUser(ctx *gin.Context) {
	traceCtx, span := u.tracer.Start(ctx.Request.Context(), "UserHandler.DeleteUser")
	defer span.End()

	_, parseSpan := u.tracer.Start(traceCtx, "UserHandler.ParseUserID")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.SetStatus(codes.Error, err.Error())
		parseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, "invalid user id", http.StatusBadRequest)
		return
	}
	parseSpan.SetAttributes(attribute.Int("user.id", id))
	parseSpan.End()

	usecaseCtx, usecaseSpan := u.tracer.Start(traceCtx, "UserHandler.DeleteUserUsecase")
	err = u.userUsecase.DeleteUser(usecaseCtx, uint(id))
	if err != nil {
		usecaseSpan.RecordError(err)
		usecaseSpan.SetStatus(codes.Error, err.Error())
		usecaseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
		return
	}
	usecaseSpan.End()

	span.SetStatus(codes.Ok, "User deleted successfully")
	helper.WriteAPIResponse(ctx, nil, "User deleted successfully", http.StatusOK)
}
