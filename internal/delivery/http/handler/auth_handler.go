package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kareemhamed001/POS/internal/cache"
	"github.com/kareemhamed001/POS/internal/delivery/http/helper"
	"github.com/kareemhamed001/POS/internal/delivery/http/request"
	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/internal/usecase"
	"github.com/kareemhamed001/POS/pkg/jwt"
	"github.com/kareemhamed001/POS/pkg/logger"
)

type AuthHandler struct {
	logger      *logger.Logger
	userUsecase *usecase.UserUsecase
	validate    *validator.Validate
	jwtManager  *jwt.JWTManager
	tokenCache  cache.TokenCache
}

func NewAuthHandler(
	log *logger.Logger,
	userUsecase *usecase.UserUsecase,
	validate *validator.Validate,
	jwtManager *jwt.JWTManager,
	tokenCache cache.TokenCache,
) *AuthHandler {
	return &AuthHandler{
		logger:      log,
		userUsecase: userUsecase,
		validate:    validate,
		jwtManager:  jwtManager,
		tokenCache:  tokenCache,
	}
}

// Register creates a new user account
func (a *AuthHandler) Register(ctx *gin.Context) {
	var req request.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Failed to bind request body: " + err.Error())
		helper.WriteError(ctx, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request")
		return
	}

	if err := a.validate.Struct(&req); err != nil {
		a.logger.Error("Validation error: " + err.Error())
		helper.WriteError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		Role:     entity.RoleCustomer,
	}

	createdUser, err := a.userUsecase.Register(ctx, user)
	if err != nil {
		a.logger.Error("Failed to register user: " + err.Error())
		helper.WriteError(ctx, http.StatusInternalServerError, "REGISTRATION_FAILED", err.Error())
		return
	}

	// Generate token
	token, err := a.jwtManager.Generate(createdUser.ID, createdUser.Email, string(createdUser.Role))
	if err != nil {
		a.logger.Error("Failed to generate token: " + err.Error())
		helper.WriteError(ctx, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", err.Error())
		return
	}

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

// Login authenticates user and returns JWT token
func (a *AuthHandler) Login(ctx *gin.Context) {
	var req request.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Failed to bind request body: " + err.Error())
		helper.WriteError(ctx, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request")
		return
	}

	if err := a.validate.Struct(&req); err != nil {
		a.logger.Error("Validation error: " + err.Error())
		helper.WriteError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Get user by email
	user, err := a.userUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		a.logger.Error("Failed to login: " + err.Error())
		helper.WriteError(ctx, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		return
	}

	// Generate token
	token, err := a.jwtManager.Generate(user.ID, user.Email, string(user.Role))
	if err != nil {
		a.logger.Error("Failed to generate token: " + err.Error())
		helper.WriteError(ctx, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", err.Error())
		return
	}

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

// Logout invalidates the user's JWT token
func (a *AuthHandler) Logout(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		helper.WriteError(ctx, http.StatusBadRequest, "MISSING_TOKEN", "authorization header required")
		return
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		helper.WriteError(ctx, http.StatusBadRequest, "INVALID_TOKEN_FORMAT", "invalid authorization header format")
		return
	}

	token := parts[1]

	// Verify token to get expiration
	claims, err := a.jwtManager.Verify(token)
	if err != nil {
		helper.WriteError(ctx, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired token")
		return
	}

	// Calculate remaining TTL
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl < 0 {
		helper.WriteError(ctx, http.StatusUnauthorized, "TOKEN_EXPIRED", "token already expired")
		return
	}

	// Blacklist the token
	if err := a.tokenCache.BlacklistToken(ctx, token, ttl); err != nil {
		a.logger.Errorf("Failed to blacklist token: %v", err)
		helper.WriteError(ctx, http.StatusInternalServerError, "LOGOUT_FAILED", "failed to logout")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}
