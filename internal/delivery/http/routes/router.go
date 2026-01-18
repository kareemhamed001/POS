package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kareemhamed001/POS/internal/cache"
	"github.com/kareemhamed001/POS/internal/delivery/http/handler"
	"github.com/kareemhamed001/POS/internal/delivery/http/middleware"
	"github.com/kareemhamed001/POS/pkg/jwt"
)

func SetupAuthRoutes(router *gin.Engine, authHandler *handler.AuthHandler, rateLimitCache cache.RateLimitCache, jwtManager *jwt.JWTManager, tokenCache cache.TokenCache) {
	authRoutes := router.Group("/api/auth")
	{
		// Rate limit: 10 requests per minute for login/register
		authRoutes.POST("/register", middleware.RateLimit(rateLimitCache, 10, time.Minute), authHandler.Register)
		authRoutes.POST("/login", middleware.RateLimit(rateLimitCache, 10, time.Minute), authHandler.Login)
		authRoutes.POST("/logout", middleware.Auth(jwtManager, tokenCache), authHandler.Logout)
	}
}

func SetupUserRoutes(router *gin.Engine, userHandler *handler.UserHandler, jwtManager *jwt.JWTManager, tokenCache cache.TokenCache) {
	userRoutes := router.Group("/api/users")
	userRoutes.Use(middleware.Auth(jwtManager, tokenCache))
	{
		userRoutes.GET("", userHandler.ListUsers)
		userRoutes.POST("", middleware.HasRole("admin"), userHandler.CreateUser)
		userRoutes.GET("/:id", userHandler.GetUserByID)
		userRoutes.PUT("/:id", userHandler.UpdateUser)
		userRoutes.DELETE("/:id", middleware.HasRole("admin"), userHandler.DeleteUser)
	}
}

func SetupOrderRoutes(router *gin.Engine, orderHandler *handler.OrderHandler, jwtManager *jwt.JWTManager, tokenCache cache.TokenCache) {
	orderRoutes := router.Group("/api/orders")
	orderRoutes.Use(middleware.Auth(jwtManager, tokenCache))
	{
		orderRoutes.POST("", orderHandler.CreateOrder)
		orderRoutes.GET("/:id", orderHandler.GetOrderByID)
		orderRoutes.PATCH("/:id/status", middleware.HasRole("admin"), orderHandler.UpdateStatus)
	}
}

func SetupAddressRoutes(router *gin.Engine, addressHandler *handler.AddressHandler, jwtManager *jwt.JWTManager, tokenCache cache.TokenCache) {
	addressRoutes := router.Group("/api/addresses")
	addressRoutes.Use(middleware.Auth(jwtManager, tokenCache))
	{
		addressRoutes.POST("", addressHandler.AddAddress)
		addressRoutes.GET("/user/:user_id", addressHandler.GetUserAddresses)
	}
}

func SetupProductRoutes(router *gin.Engine, productHandler *handler.ProductHandler, jwtManager *jwt.JWTManager, tokenCache cache.TokenCache) {
	productRoutes := router.Group("/api/products")
	productRoutes.Use(middleware.Auth(jwtManager, tokenCache))
	{
		productRoutes.POST("", middleware.HasRole("admin"), productHandler.CreateProduct)
		productRoutes.GET("", productHandler.ListProducts)
		productRoutes.GET("/:id", productHandler.GetProduct)
		productRoutes.PUT("/:id", middleware.HasRole("admin"), productHandler.UpdateProduct)
		productRoutes.DELETE("/:id", middleware.HasRole("admin"), productHandler.DeleteProduct)
	}
}
