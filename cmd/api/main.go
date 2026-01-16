package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kareemhamed001/POS/internal/config"
	"github.com/kareemhamed001/POS/internal/db"
	"github.com/kareemhamed001/POS/internal/delivery/http/handler"
	"github.com/kareemhamed001/POS/internal/delivery/http/middleware"
	"github.com/kareemhamed001/POS/internal/delivery/http/routes"
	validation "github.com/kareemhamed001/POS/internal/delivery/http/validator"
	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/internal/repository/postgres"
	"github.com/kareemhamed001/POS/internal/usecase"
	"github.com/kareemhamed001/POS/pkg/jwt"
	"github.com/kareemhamed001/POS/pkg/logger"
)

func main() {
	config := config.NewConfig()

	// Initialize global logger
	logger.InitGlobal(config.AppEnv)
	defer logger.Sync()

	db, err := db.InitializeDB(config.DBDriver, config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
	if err != nil {
		panic(err)
	}
	logger.Info("DB started successfully")

	//migrate
	db.AutoMigrate(&entity.User{}, &entity.Address{}, &entity.Order{}, &entity.Product{}, &entity.OrderItem{})
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			panic(err)
		}
		sqlDB.Close()
	}()

	router := gin.Default()

	if config.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router.Use(middleware.LoggerMiddleware(logger.Get()))

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "OK",
		})
	})
	userRepository := postgres.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository)
	validate := validation.NewValidator()
	userHandler := handler.NewUserHandler(logger.Get(), userUsecase, validate)

	productRepository := postgres.NewProductRepository(db)

	orderRepository := postgres.NewOrderRepository(db)
	orderUsecase := usecase.NewOrderUsecase(orderRepository, productRepository)
	orderHandler := handler.NewOrderHandler(orderUsecase, validate)

	addressRepository := postgres.NewAddressRepository(db)
	addressUsecase := usecase.NewAddressUsecase(addressRepository)
	addressHandler := handler.NewAddressHandler(addressUsecase, validate)

	productUsecase := usecase.NewProductUsecase(productRepository)
	productHandler := handler.NewProductHandler(productUsecase, validate)

	// Initialize JWT Manager
	jwtManager := jwt.NewJWTManager(config.JWTPrivateKey, config.JWTTokenDuration)

	// Initialize Auth Handler
	authHandler := handler.NewAuthHandler(logger.Get(), userUsecase, validate, jwtManager)

	// Setup Routes
	routes.SetupAuthRoutes(router, authHandler)
	routes.SetupUserRoutes(router, userHandler, jwtManager)
	routes.SetupOrderRoutes(router, orderHandler, jwtManager)
	routes.SetupAddressRoutes(router, addressHandler, jwtManager)
	routes.SetupProductRoutes(router, productHandler, jwtManager)

	logger.Info("Starting Server on port " + strconv.Itoa(config.AppPort))
	router.Run(":" + strconv.Itoa(config.AppPort))
}
