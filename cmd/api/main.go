package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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
	"gorm.io/gorm"
)

func main() {
	cfg := config.NewConfig()

	logger.InitGlobal(cfg.AppEnv)
	defer logger.Sync()

	setGinMode(cfg.AppEnv)

	shutdownTracing := initTracing(cfg)
	defer shutdownTracing()

	metricsHandler, shutdownMetrics := initMetrics()
	defer shutdownMetrics()

	metricsCollector := tracer.NewGenericMetricsCollector("http-handler")

	dbConn, closeDB, err := initDatabase(cfg)
	if err != nil {
		logger.Errorf("Failed to initialize DB: %v", err)
		return
	}
	defer closeDB()

	if err := dbConn.AutoMigrate(&entity.User{}, &entity.Address{}, &entity.Order{}, &entity.Product{}, &entity.OrderItem{}); err != nil {
		logger.Errorf("AutoMigrate failed: %v", err)
	}

	rdb := initRedis(cfg)
	defer rdb.Close()

	var productCache cache.ProductCache = redisCache.NewProductCache(rdb)
	var tokenCache cache.TokenCache = redisCache.NewTokenCache(rdb)
	var rateLimitCache cache.RateLimitCache = redisCache.NewRateLimitCache(rdb)

	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Warnf("Failed to set trusted proxies: %v", err)
	}

	router.Use(
		middleware.RecoveryMiddleware(),
		middleware.TracingMiddleware("pos-api"),
		middleware.MetricsMiddleware(metricsCollector),
		middleware.LoggerMiddleware(logger.Get()),
	)

	if cfg.StorageType == "local" {
		router.Static(cfg.StorageLocalURL, cfg.StorageLocalPath)
	}

	if metricsHandler != nil {
		router.GET("/metrics", gin.WrapH(metricsHandler))
	}

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "OK"})
	})

	userRepository := postgres.NewUserRepository(dbConn)
	userUsecase := usecase.NewUserUsecase(userRepository)
	validate := validation.NewValidator()
	userHandler := handler.NewUserHandler(userUsecase, validate)

	if err := seeder.SeedAdmin(context.Background(), userRepository, cfg.AdminName, cfg.AdminEmail, cfg.AdminPhone, cfg.AdminPassword); err != nil {
		logger.Errorf("failed to seed admin user: %v", err)
	}

	fileStorage, err := file.NewFileStorage(file.StorageConfig{
		Type:          cfg.StorageType,
		LocalBasePath: cfg.StorageLocalPath,
		LocalBaseURL:  cfg.StorageLocalURL,
		S3Bucket:      cfg.S3Bucket,
		S3Region:      cfg.S3Region,
		S3AccessKey:   cfg.S3AccessKey,
		S3SecretKey:   cfg.S3SecretKey,
		S3Endpoint:    cfg.S3Endpoint,
		S3BaseURL:     cfg.S3BaseURL,
		S3ACL:         cfg.S3ACL,
	})
	if err != nil {
		logger.Errorf("failed to initialize file storage: %v", err)
		return
	}
	logger.Infof("File storage initialized: %s", cfg.StorageType)

	productRepository := postgres.NewProductRepository(dbConn)
	orderRepository := postgres.NewOrderRepository(dbConn)
	orderUsecase := usecase.NewOrderUsecase(orderRepository, productRepository)
	orderHandler := handler.NewOrderHandler(orderUsecase, validate)

	addressRepository := postgres.NewAddressRepository(dbConn)
	addressUsecase := usecase.NewAddressUsecase(addressRepository)
	addressHandler := handler.NewAddressHandler(addressUsecase, validate)

	productUsecase := usecase.NewProductUsecase(productRepository, productCache)
	productHandler := handler.NewProductHandler(productUsecase, validate, fileStorage)

	jwtManager := jwt.NewJWTManager(cfg.JWTPrivateKey, cfg.JWTTokenDuration)
	authHandler := handler.NewAuthHandler(userUsecase, validate, jwtManager, tokenCache)

	routes.SetupAuthRoutes(router, authHandler, rateLimitCache, jwtManager, tokenCache)
	routes.SetupUserRoutes(router, userHandler, jwtManager, tokenCache)
	routes.SetupOrderRoutes(router, orderHandler, jwtManager, tokenCache)
	routes.SetupAddressRoutes(router, addressHandler, jwtManager, tokenCache)
	routes.SetupProductRoutes(router, productHandler, jwtManager, tokenCache)

	startServer(router, cfg.AppPort)
}

func setGinMode(appEnv string) {
	if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
}

func initTracing(cfg *config.Config) func() {
	jaegerEndpoint := getEnv("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces")
	tp, err := tracer.InitTracer("pos-api", jaegerEndpoint)
	if err != nil {
		logger.Warnf("Failed to initialize tracer: %v. Continuing without tracing.", err)
		return func() {}
	}

	logger.Info("OpenTelemetry tracer initialized successfully")
	return func() {
		if err := tracer.Shutdown(context.Background(), tp); err != nil {
			logger.Errorf("Failed to shutdown tracer: %v", err)
		}
	}
}

func initMetrics() (http.Handler, func()) {
	mp, metricsHandler, err := tracer.InitPrometheusMeterProvider("pos")
	if err != nil {
		logger.Warnf("Failed to initialize prometheus metrics: %v. Continuing without metrics.", err)
		return nil, func() {}
	}

	logger.Info("Prometheus metrics exporter initialized successfully")
	return metricsHandler, func() {
		if err := mp.Shutdown(context.Background()); err != nil {
			logger.Errorf("Failed to shutdown meter provider: %v", err)
		}
	}
}

func initDatabase(cfg *config.Config) (*gorm.DB, func(), error) {
	dbConn, err := db.InitializeDB(cfg.DBDriver, cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		return nil, func() {}, err
	}
	logger.Info("DB started successfully")

	closeFn := func() {
		sqlDB, err := dbConn.DB()
		if err != nil {
			logger.Errorf("Failed to get sql DB: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			logger.Errorf("Failed to close DB: %v", err)
		}
	}

	return dbConn, closeFn, nil
}

func initRedis(cfg *config.Config) *redisClient.Client {
	rdb, err := redisClient.NewClient(cfg)
	if err != nil {
		logger.Warnf("Redis initialization failed: %v. Continuing without cache.", err)
		return &redisClient.Client{}
	}
	return rdb
}

func startServer(router *gin.Engine, port int) {
	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		logger.Infof("Starting server on port %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Infof("Shutdown signal received: %s", sig.String())
	case err := <-errChan:
		logger.Errorf("Server error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown failed: %v", err)
	} else {
		logger.Info("Server stopped gracefully")
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
