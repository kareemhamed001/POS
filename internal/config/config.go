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
	AdminName        string
	AdminEmail       string
	AdminPhone       string
	AdminPassword    string
	RedisHost        string
	RedisPort        string
	RedisPassword    string
	RedisDB          int
	RedisEnabled     bool
	// File Storage
	StorageType      string
	StorageLocalPath string
	StorageLocalURL  string
	S3Bucket         string
	S3Region         string
	S3AccessKey      string
	S3SecretKey      string
	S3Endpoint       string
	S3BaseURL        string
	S3ACL            string
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env  file not found, using system environment variables")
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
		AdminName:        getEnvString("ADMIN_NAME", "Admin"),
		AdminEmail:       getEnvString("ADMIN_EMAIL", "admin@example.com"),
		AdminPhone:       getEnvString("ADMIN_PHONE", "+201000000000"),
		AdminPassword:    getEnvString("ADMIN_PASSWORD", "Admin123!"),
		RedisHost:        getEnvString("REDIS_HOST", "localhost"),
		RedisPort:        getEnvString("REDIS_PORT", "6379"),
		RedisPassword:    getEnvString("REDIS_PASSWORD", ""),
		RedisDB:          getEnvInt("REDIS_DB", 0),
		RedisEnabled:     getEnvBool("REDIS_ENABLED", true),
		// File Storage
		StorageType:      getEnvString("STORAGE_TYPE", "local"),
		StorageLocalPath: getEnvString("STORAGE_LOCAL_PATH", "uploads"),
		StorageLocalURL:  getEnvString("STORAGE_LOCAL_URL", "/uploads"),
		S3Bucket:         getEnvString("S3_BUCKET", ""),
		S3Region:         getEnvString("S3_REGION", "us-east-1"),
		S3AccessKey:      getEnvString("S3_ACCESS_KEY", ""),
		S3SecretKey:      getEnvString("S3_SECRET_KEY", ""),
		S3Endpoint:       getEnvString("S3_ENDPOINT", ""),
		S3BaseURL:        getEnvString("S3_BASE_URL", ""),
		S3ACL:            getEnvString("S3_ACL", "public-read"),
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
