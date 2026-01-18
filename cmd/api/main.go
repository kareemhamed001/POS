package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kareemhamed001/POS/internal/cache"
	redisCache "github.com/kareemhamed001/POS/internal/cache/redis"
	"github.com/kareemhamed001/POS/internal/config"
	"github.com/kareemhamed001/POS/internal/db"
	"github.com/kareemhamed001/POS/internal/db/seeder"
	"github.com/kareemhamed001/POS/internal/delivery/http/handler"
	"github.com/kareemhamed001/POS/internal/delivery/http/middleware"
	"github.com/kareemhamed001/POS/internal/delivery/http/routes"
	validation "github.com/kareemhamed001/POS/internal/delivery/http/validation"
	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/internal/repository/postgres"
	"github.com/kareemhamed001/POS/internal/usecase"
	"github.com/kareemhamed001/POS/pkg/file"
	"github.com/kareemhamed001/POS/pkg/jwt"
	"github.com/kareemhamed001/POS/pkg/logger"
	redisClient "github.com/kareemhamed001/POS/pkg/redis"
	"github.com/kareemhamed001/POS/pkg/tracer"
)

func main() {
	config := config.NewConfig()

	// Initialize global logger
	logger.InitGlobal(config.AppEnv)
	defer logger.Sync()

	// Initialize OpenTelemetry tracer
	jaegerEndpoint := getEnv("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces")
	tp, err := tracer.InitTracer("pos-api", jaegerEndpoint)
	if err != nil {
		logger.Warnf("Failed to initialize tracer: %v. Continuing without tracing.", err)
	} else {
		defer func() {
			if err := tracer.Shutdown(context.Background(), tp); err != nil {
				logger.Errorf("Failed to shutdown tracer: %v", err)
			}
		}()
		logger.Info("OpenTelemetry tracer initialized successfully")
	}

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

	// Initialize Redis
	rdb, err := redisClient.NewClient(config)
	if err != nil {
		logger.Warnf("Redis initialization failed: %v. Continuing without cache.", err)
		rdb = &redisClient.Client{} // Empty client for graceful degradation
	}
	defer rdb.Close()

	// Initialize caches
	var productCache cache.ProductCache = redisCache.NewProductCache(rdb)
	var tokenCache cache.TokenCache = redisCache.NewTokenCache(rdb)
	var rateLimitCache cache.RateLimitCache = redisCache.NewRateLimitCache(rdb)

	router := gin.Default()

	if config.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Add tracing middleware (should be early in the chain)
	router.Use(middleware.TracingMiddleware("pos-api"))
	router.Use(middleware.LoggerMiddleware(logger.Get()))

	// Serve uploaded files for local storage
	if config.StorageType == "local" {
		router.Static(config.StorageLocalURL, config.StorageLocalPath)
	}

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "OK",
		})
	})
	userRepository := postgres.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepository)
	validate := validation.NewValidator()
	userHandler := handler.NewUserHandler(logger.Get(), userUsecase, validate)

	// Seed default admin user (idempotent)
	if err := seeder.SeedAdmin(context.Background(), userRepository, config.AdminName, config.AdminEmail, config.AdminPhone, config.AdminPassword); err != nil {
		logger.Errorf("failed to seed admin user: %v", err)
	}

	// Initialize File Storage
	fileStorage, err := file.NewFileStorage(file.StorageConfig{
		Type:          config.StorageType,
		LocalBasePath: config.StorageLocalPath,
		LocalBaseURL:  config.StorageLocalURL,
		S3Bucket:      config.S3Bucket,
		S3Region:      config.S3Region,
		S3AccessKey:   config.S3AccessKey,
		S3SecretKey:   config.S3SecretKey,
		S3Endpoint:    config.S3Endpoint,
		S3BaseURL:     config.S3BaseURL,
		S3ACL:         config.S3ACL,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to initialize file storage: %v", err))
	}
	logger.Infof("File storage initialized: %s", config.StorageType)

	productRepository := postgres.NewProductRepository(db)

	orderRepository := postgres.NewOrderRepository(db)
	orderUsecase := usecase.NewOrderUsecase(orderRepository, productRepository)
	orderHandler := handler.NewOrderHandler(orderUsecase, validate)

	addressRepository := postgres.NewAddressRepository(db)
	addressUsecase := usecase.NewAddressUsecase(addressRepository)
	addressHandler := handler.NewAddressHandler(addressUsecase, validate)

	productUsecase := usecase.NewProductUsecase(productRepository, productCache)
	productHandler := handler.NewProductHandler(productUsecase, validate, fileStorage)

	// Initialize JWT Manager
	jwtManager := jwt.NewJWTManager(config.JWTPrivateKey, config.JWTTokenDuration)

	// Initialize Auth Handler
	authHandler := handler.NewAuthHandler(logger.Get(), userUsecase, validate, jwtManager, tokenCache)

	// Setup Routes
	routes.SetupAuthRoutes(router, authHandler, rateLimitCache, jwtManager, tokenCache)
	routes.SetupUserRoutes(router, userHandler, jwtManager, tokenCache)
	routes.SetupOrderRoutes(router, orderHandler, jwtManager, tokenCache)
	routes.SetupAddressRoutes(router, addressHandler, jwtManager, tokenCache)
	routes.SetupProductRoutes(router, productHandler, jwtManager, tokenCache)

	logger.Info("Starting Server on port " + strconv.Itoa(config.AppPort))
	router.Run(":" + strconv.Itoa(config.AppPort))
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
