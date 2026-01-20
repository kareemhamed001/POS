package db

import (
	"fmt"
	"time"

	"github.com/kareemhamed001/POS/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func InitializeDB(DB_DRIVER string, DB_HOST string, DB_PORT string, DB_USER string, DB_PASSWORD string, DB_NAME string) (*gorm.DB, error) {

	// Create custom GORM logger using zap
	gormLogger := logger.NewGormLoggerFromGlobal().
		SetLogLevel(gormlogger.Info).
		SetSlowThreshold(200 * time.Millisecond).
		SetIgnoreRecordNotFoundError(true)

	switch DB_DRIVER {
	case "postgres":
		connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME)
		return gorm.Open(postgres.Open(connectionString), &gorm.Config{
			Logger: gormLogger,
		})

	default:
		panic("Unsupported DB_DRIVER: " + DB_DRIVER)
	}

}
