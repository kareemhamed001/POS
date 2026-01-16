package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kareemhamed001/POS/internal/delivery/http/handler"
)

func SetupUserRoutes(router *gin.Engine, userHandler *handler.UserHandler) {
	userRoutes := router.Group("/api/users")
	{
		userRoutes.GET("", userHandler.ListUsers)
		userRoutes.POST("", userHandler.CreateUser)
		userRoutes.GET("/:id", userHandler.GetUserByID)
		userRoutes.PUT("/:id", userHandler.UpdateUser)
		userRoutes.DELETE("/:id", userHandler.DeleteUser)
	}
}

func SetupOrderRoutes(router *gin.Engine, orderHandler *handler.OrderHandler) {
	orderRoutes := router.Group("/api/orders")
	{
		orderRoutes.POST("", orderHandler.CreateOrder)
		orderRoutes.GET("/:id", orderHandler.GetOrderByID)
		orderRoutes.PATCH("/:id/status", orderHandler.UpdateStatus)
	}
}

func SetupAddressRoutes(router *gin.Engine, addressHandler *handler.AddressHandler) {
	addressRoutes := router.Group("/api/addresses")
	{
		addressRoutes.POST("", addressHandler.AddAddress)
		addressRoutes.GET("/user/:user_id", addressHandler.GetUserAddresses)
	}
}

func SetupProductRoutes(router *gin.Engine, productHandler *handler.ProductHandler) {
	productRoutes := router.Group("/api/products")
	{
		productRoutes.POST("", productHandler.CreateProduct)
		productRoutes.GET("", productHandler.ListProducts)
		productRoutes.GET("/:id", productHandler.GetProduct)
	}
}
