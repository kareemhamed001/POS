package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kareemhamed001/POS/internal/core/ports"
	"github.com/kareemhamed001/POS/pkg/jwt"
)

const (
	UserContextKey = "user_claims"
	UserIDKey      = "user_id"
	UserRoleKey    = "user_role"
	UserEmailKey   = "user_email"
)

// Auth middleware verifies JWT token from Authorization header
func Auth(jwtManager *jwt.JWTManager, tokenCache ports.TokenCache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "authorization header required",
			})
			ctx.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid authorization header format",
			})
			ctx.Abort()
			return
		}

		token := parts[1]

		// Check if token is blacklisted
		if blacklisted, err := tokenCache.IsTokenBlacklisted(ctx, token); err == nil && blacklisted {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "token has been revoked",
			})
			ctx.Abort()
			return
		}

		// Verify token
		claims, err := jwtManager.Verify(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid or expired token",
			})
			ctx.Abort()
			return
		}

		// Store claims in context
		ctx.Set(UserContextKey, claims)
		ctx.Set(UserIDKey, claims.UserID)
		ctx.Set(UserRoleKey, claims.Role)
		ctx.Set(UserEmailKey, claims.Email)

		ctx.Next()
	}
}

// HasRole middleware checks if user has one of the required roles
func HasRole(requiredRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, exists := ctx.Get(UserContextKey)
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "user claims not found",
			})
			ctx.Abort()
			return
		}

		userClaims, ok := claims.(*jwt.UserClaims)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid user claims",
			})
			ctx.Abort()
			return
		}

		// Check if user role is in required roles
		hasRole := false
		for _, role := range requiredRoles {
			if userClaims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			ctx.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "insufficient permissions",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
