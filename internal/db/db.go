package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitializeDB(DB_DRIVER string, DB_HOST string, DB_PORT string, DB_USER string, DB_PASSWORD string, DB_NAME string) (*gorm.DB, error) {

	switch DB_DRIVER {
	case "postgres":
		connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME)
		return gorm.Open(postgres.Open(connectionString), &gorm.Config{})

	default:
		panic("Unsupported DB_DRIVER: " + DB_DRIVER)
	}

}
