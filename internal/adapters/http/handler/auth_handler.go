package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	entity "github.com/kareemhamed001/POS/internal/core/domain"
	"github.com/kareemhamed001/POS/internal/core/ports"
	"github.com/kareemhamed001/POS/internal/core/usecase"
	"github.com/kareemhamed001/POS/internal/adapters/http/helper"
	"github.com/kareemhamed001/POS/internal/adapters/http/request"
	"github.com/kareemhamed001/POS/pkg/jwt"
	"github.com/kareemhamed001/POS/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AuthHandler struct {
	userUsecase *usecase.UserUsecase
	validate    *validator.Validate
	jwtManager  *jwt.JWTManager
	tokenCache  ports.TokenCache
	tracer      trace.Tracer
}

func NewAuthHandler(
	userUsecase *usecase.UserUsecase,
	validate *validator.Validate,
	jwtManager *jwt.JWTManager,
	tokenCache ports.TokenCache,
) *AuthHandler {
	return &AuthHandler{
		userUsecase: userUsecase,
		validate:    validate,
		jwtManager:  jwtManager,
		tokenCache:  tokenCache,
		tracer:      otel.Tracer("auth-handler"),
	}
}

func (a *AuthHandler) Register(ctx *gin.Context) {
	traceCtx, span := a.tracer.Start(ctx.Request.Context(), "AuthHandler.Register")
	defer span.End()

	_, bindSpan := a.tracer.Start(traceCtx, "AuthHandler.BindRegisterRequest")
	var req request.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		bindSpan.RecordError(err)
		bindSpan.SetStatus(codes.Error, err.Error())
		bindSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("Failed to bind request body: " + err.Error())
		helper.WriteError(ctx, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request")
		return
	}
	bindSpan.End()

	_, validateSpan := a.tracer.Start(traceCtx, "AuthHandler.ValidateRegisterRequest")
	if err := a.validate.Struct(&req); err != nil {
		validateSpan.RecordError(err)
		validateSpan.SetStatus(codes.Error, err.Error())
		validateSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("Validation error: " + err.Error())
		helper.WriteError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	validateSpan.End()

	_, buildSpan := a.tracer.Start(traceCtx, "AuthHandler.BuildUserEntity")
	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		Role:     entity.RoleCustomer,
	}
	buildSpan.End()

	usecaseCtx, usecaseSpan := a.tracer.Start(traceCtx, "AuthHandler.RegisterUsecase")
	createdUser, err := a.userUsecase.Register(usecaseCtx, user)
	if err != nil {
		usecaseSpan.RecordError(err)
		usecaseSpan.SetStatus(codes.Error, err.Error())
		usecaseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("Failed to register user: " + err.Error())
		helper.WriteError(ctx, http.StatusInternalServerError, "REGISTRATION_FAILED", err.Error())
		return
	}
	usecaseSpan.End()

	_, tokenSpan := a.tracer.Start(traceCtx, "AuthHandler.GenerateToken")
	token, err := a.jwtManager.Generate(createdUser.ID, createdUser.Email, string(createdUser.Role))
	if err != nil {
		tokenSpan.RecordError(err)
		tokenSpan.SetStatus(codes.Error, err.Error())
		tokenSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("Failed to generate token: " + err.Error())
		helper.WriteError(ctx, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", err.Error())
		return
	}
	tokenSpan.End()

	span.SetStatus(codes.Ok, "User registered successfully")
	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
		"data": gin.H{
			"user": gin.H{
				"id":    createdUser.ID,
				"name":  createdUser.Name,
				"email": createdUser.Email,
				"phone": createdUser.Phone,
				"role":  createdUser.Role,
			},
			"token": token,
		},
	})
}

func (a *AuthHandler) Login(ctx *gin.Context) {
	traceCtx, span := a.tracer.Start(ctx.Request.Context(), "AuthHandler.Login")
	defer span.End()

	_, bindSpan := a.tracer.Start(traceCtx, "AuthHandler.BindLoginRequest")
	var req request.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		bindSpan.RecordError(err)
		bindSpan.SetStatus(codes.Error, err.Error())
		bindSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("Failed to bind request body: " + err.Error())
		helper.WriteError(ctx, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request")
		return
	}
	bindSpan.End()

	_, validateSpan := a.tracer.Start(traceCtx, "AuthHandler.ValidateLoginRequest")
	if err := a.validate.Struct(&req); err != nil {
		validateSpan.RecordError(err)
		validateSpan.SetStatus(codes.Error, err.Error())
		validateSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("Validation error: " + err.Error())
		helper.WriteError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	validateSpan.End()

	usecaseCtx, usecaseSpan := a.tracer.Start(traceCtx, "AuthHandler.LoginUsecase")
	user, err := a.userUsecase.Login(usecaseCtx, req.Email, req.Password)
	if err != nil {
		usecaseSpan.RecordError(err)
		usecaseSpan.SetStatus(codes.Error, err.Error())
		usecaseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("Failed to login: " + err.Error())
		helper.WriteError(ctx, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		return
	}
	usecaseSpan.End()

	_, tokenSpan := a.tracer.Start(traceCtx, "AuthHandler.GenerateToken")
	token, err := a.jwtManager.Generate(user.ID, user.Email, string(user.Role))
	if err != nil {
		tokenSpan.RecordError(err)
		tokenSpan.SetStatus(codes.Error, err.Error())
		tokenSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("Failed to generate token: " + err.Error())
		helper.WriteError(ctx, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", err.Error())
		return
	}
	tokenSpan.End()

	span.SetStatus(codes.Ok, "User logged in successfully")
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User logged in successfully",
		"data": gin.H{
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"phone": user.Phone,
				"role":  user.Role,
			},
			"token": token,
		},
	})
}

func (a *AuthHandler) Logout(ctx *gin.Context) {
	traceCtx, span := a.tracer.Start(ctx.Request.Context(), "AuthHandler.Logout")
	defer span.End()

	_, authHeaderSpan := a.tracer.Start(traceCtx, "AuthHandler.ReadAuthHeader")
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		authHeaderSpan.SetStatus(codes.Error, "missing authorization header")
		authHeaderSpan.End()
		span.SetStatus(codes.Error, "missing authorization header")
		helper.WriteError(ctx, http.StatusBadRequest, "MISSING_TOKEN", "authorization header required")
		return
	}
	authHeaderSpan.End()

	_, parseTokenSpan := a.tracer.Start(traceCtx, "AuthHandler.ParseToken")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		parseTokenSpan.SetStatus(codes.Error, "invalid authorization header format")
		parseTokenSpan.End()
		span.SetStatus(codes.Error, "invalid authorization header format")
		helper.WriteError(ctx, http.StatusBadRequest, "INVALID_TOKEN_FORMAT", "invalid authorization header format")
		return
	}

	token := parts[1]
	parseTokenSpan.End()

	_, verifySpan := a.tracer.Start(traceCtx, "AuthHandler.VerifyToken")
	claims, err := a.jwtManager.Verify(token)
	if err != nil {
		verifySpan.RecordError(err)
		verifySpan.SetStatus(codes.Error, err.Error())
		verifySpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		helper.WriteError(ctx, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired token")
		return
	}
	verifySpan.End()

	_, ttlSpan := a.tracer.Start(traceCtx, "AuthHandler.CalculateTTL")
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl < 0 {
		ttlSpan.SetStatus(codes.Error, "token already expired")
		ttlSpan.End()
		span.SetStatus(codes.Error, "token already expired")
		helper.WriteError(ctx, http.StatusUnauthorized, "TOKEN_EXPIRED", "token already expired")
		return
	}
	ttlSpan.End()

	cacheCtx, blacklistSpan := a.tracer.Start(traceCtx, "AuthHandler.BlacklistToken")
	if err := a.tokenCache.BlacklistToken(cacheCtx, token, ttl); err != nil {
		blacklistSpan.RecordError(err)
		blacklistSpan.SetStatus(codes.Error, err.Error())
		blacklistSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Errorf("Failed to blacklist token: %v", err)
		helper.WriteError(ctx, http.StatusInternalServerError, "LOGOUT_FAILED", "failed to logout")
		return
	}
	blacklistSpan.End()

	span.SetStatus(codes.Ok, "Logged out successfully")
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}
