package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds environment variables
type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	RedisHost          string
	RedisPort          string
	RedisPassword      string
	RedisDB            int
	AppPort            string
	EnableMigration    bool
	JWTSecret          string
	AESSecret          string
	JWTAccessExpMinutes int
	JWTRefreshExpDays   int
	SMTPHost           string
	SMTPPort           string
	SMTPUsername       string
	SMTPPassword       string
	SMTPFrom           string
}

// LoadConfig reads the .env file and sets up the configuration
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	redisDBStr := getEnvOrDefault("REDIS_DB", "0")
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		log.Fatalf("environment variable REDIS_DB is not a valid integer")
	}

	enableMigrationStr := getEnvOrDefault("ENABLE_MIGRATION_DB", "false")
	enableMigration, err := strconv.ParseBool(enableMigrationStr)
	if err != nil {
		log.Fatalf("environment variable ENABLE_MIGRATION_DB is not a valid boolean")
	}

	accessExpStr := getEnvOrDefault("JWT_ACCESS_EXP_MINUTES", "15")
	accessExp, err := strconv.Atoi(accessExpStr)
	if err != nil {
		accessExp = 15
	}

	refreshExpStr := getEnvOrDefault("JWT_REFRESH_EXP_DAYS", "7")
	refreshExp, err := strconv.Atoi(refreshExpStr)
	if err != nil {
		refreshExp = 7
	}

	return &Config{
		DBHost:              getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:              getEnvOrDefault("DB_PORT", "5432"),
		DBUser:              getEnvOrDefault("DB_USER", "postgres"),
		DBPassword:          getEnvOrDefault("DB_PASSWORD", ""),
		DBName:              getEnvOrDefault("DB_NAME", "concert_go"),
		DBSSLMode:           getEnvOrDefault("DB_SSLMODE", "disable"),
		RedisHost:           getEnvOrDefault("REDIS_HOST", "localhost"),
		RedisPort:           getEnvOrDefault("REDIS_PORT", "6379"),
		RedisPassword:       getEnvOrDefault("REDIS_PASSWORD", ""),
		RedisDB:             redisDB,
		AppPort:             getEnvOrDefault("APP_PORT", "8080"),
		EnableMigration:     enableMigration,
		JWTSecret:           getEnvOrDefault("JWT_SECRET", "super-secret-key-change-in-production"),
		AESSecret:           getEnvOrDefault("AES_SECRET", "super-secret-aes-key-change-in-production"),
		JWTAccessExpMinutes: accessExp,
		JWTRefreshExpDays:   refreshExp,
		SMTPHost:            getEnvOrDefault("SMTP_HOST", "localhost"),
		SMTPPort:            getEnvOrDefault("SMTP_PORT", "587"),
		SMTPUsername:        getEnvOrDefault("SMTP_USERNAME", ""),
		SMTPPassword:        getEnvOrDefault("SMTP_PASSWORD", ""),
		SMTPFrom:            getEnvOrDefault("SMTP_FROM", "no-reply@concertgo.com"),
	}
}

func getEnvOrDefault(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
