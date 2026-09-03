package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds environment variables
type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	AppPort       string
	EnableMigration bool
}

// LoadConfig reads the .env file and sets up the configuration
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	redisDBStr := getEnv("REDIS_DB")
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		log.Fatalf("environment variable REDIS_DB is not a valid integer")
	}

	enableMigrationStr := getEnv("ENABLE_MIGRATION_DB")
	enableMigration, err := strconv.ParseBool(enableMigrationStr)
	if err != nil {
		log.Fatalf("environment variable ENABLE_MIGRATION_DB is not a valid boolean")
	}

	return &Config{
		DBHost:        getEnv("DB_HOST"),
		DBPort:        getEnv("DB_PORT"),
		DBUser:        getEnv("DB_USER"),
		DBPassword:    getEnv("DB_PASSWORD"),
		DBName:        getEnv("DB_NAME"),
		DBSSLMode:     getEnv("DB_SSLMODE"),
		RedisHost:     getEnv("REDIS_HOST"),
		RedisPort:     getEnv("REDIS_PORT"),
		RedisPassword: getEnv("REDIS_PASSWORD"),
		RedisDB:       redisDB,
		AppPort:       getEnv("APP_PORT"),
		EnableMigration: enableMigration,
	}
}

func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("environment variable %s not set", key)
	return ""
}
