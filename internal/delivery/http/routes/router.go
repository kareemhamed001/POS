package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kareemhamed001/POS/internal/delivery/http/handler"
	"github.com/kareemhamed001/POS/internal/delivery/http/middleware"
	"github.com/kareemhamed001/POS/pkg/jwt"
)

func SetupAuthRoutes(router *gin.Engine, authHandler *handler.AuthHandler) {
	authRoutes := router.Group("/api/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}
}

func SetupUserRoutes(router *gin.Engine, userHandler *handler.UserHandler, jwtManager *jwt.JWTManager) {
	userRoutes := router.Group("/api/users")
	userRoutes.Use(middleware.Auth(jwtManager))
	{
		userRoutes.GET("", userHandler.ListUsers)
		userRoutes.POST("", middleware.HasRole("admin"), userHandler.CreateUser)
		userRoutes.GET("/:id", userHandler.GetUserByID)
		userRoutes.PUT("/:id", userHandler.UpdateUser)
		userRoutes.DELETE("/:id", middleware.HasRole("admin"), userHandler.DeleteUser)
	}
}

func SetupOrderRoutes(router *gin.Engine, orderHandler *handler.OrderHandler, jwtManager *jwt.JWTManager) {
	orderRoutes := router.Group("/api/orders")
	orderRoutes.Use(middleware.Auth(jwtManager))
	{
		orderRoutes.POST("", orderHandler.CreateOrder)
		orderRoutes.GET("/:id", orderHandler.GetOrderByID)
		orderRoutes.PATCH("/:id/status", middleware.HasRole("admin"), orderHandler.UpdateStatus)
	}
}

func SetupAddressRoutes(router *gin.Engine, addressHandler *handler.AddressHandler, jwtManager *jwt.JWTManager) {
	addressRoutes := router.Group("/api/addresses")
	addressRoutes.Use(middleware.Auth(jwtManager))
	{
		addressRoutes.POST("", addressHandler.AddAddress)
		addressRoutes.GET("/user/:user_id", addressHandler.GetUserAddresses)
	}
}

func SetupProductRoutes(router *gin.Engine, productHandler *handler.ProductHandler, jwtManager *jwt.JWTManager) {
	productRoutes := router.Group("/api/products")
	productRoutes.Use(middleware.Auth(jwtManager))
	{
		productRoutes.POST("", middleware.HasRole("admin"), productHandler.CreateProduct)
		productRoutes.GET("", productHandler.ListProducts)
		productRoutes.GET("/:id", productHandler.GetProduct)
	}
}
