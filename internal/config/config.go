package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort          int
	AppEnv           string
	DBDriver         string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	MongoDBUri       string
	JWTPrivateKey    string
	JWTTokenDuration time.Duration
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	return &Config{
		AppPort:          getEnvInt("APP_PORT", 8080),
		AppEnv:           getEnvString("APP_ENV", "local"),
		DBDriver:         getEnvString("DB_DRIVER", "postgres"),
		DBHost:           getEnvString("DB_HOST", "localhost"),
		DBPort:           getEnvString("DB_PORT", "5432"),
		DBUser:           getEnvString("DB_USER", "admin"),
		DBPassword:       getEnvString("DB_PASSWORD", "admin"),
		DBName:           getEnvString("DB_NAME", "faq_db"),
		MongoDBUri:       getEnvString("MONGO_DB_URI", "mongodb://admin:admin@localhost:27017/faq_db"),
		JWTPrivateKey:    getEnvString("JWT_PRIVATE_KEY", "your_jwt_private_key"),
		JWTTokenDuration: time.Duration(getEnvInt("JWT_TOKEN_DURATION", 24)) * time.Hour,
	}
}

func getEnvString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}
func getEnvInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return intVal
}

func getEnvBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return boolVal
}
