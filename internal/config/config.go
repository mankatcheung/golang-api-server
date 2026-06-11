// Package config provides configuration management for the application,
// loading settings from environment variables and .env files.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv           string
	ServerAddress    string
	ServerTimeout    time.Duration
	DatabaseURL      string
	DatabaseReadURL  string
	DBMaxConns       int32
	DBMinConns       int32
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	AllowedOrigins   string
	KafkaBrokers     []string
	KafkaTopic       string
	KafkaGroupID     string
	GRPCAddress      string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	CacheTTL         time.Duration

	LogLevel        string
	LogFormat       string
	LogKafkaTopic   string
	LogKafkaGroupID string

	// Rate limits in requests per minute. Burst equals the per-minute value.
	AuthRateLimit int
	APIRateLimit  int
}

const defaultJWTSecret = "change-me-in-production"

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		ServerAddress:   getEnv("SERVER_ADDRESS", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/users_db?sslmode=disable"),
		DatabaseReadURL: getEnv("DATABASE_READ_URL", "postgres://postgres:postgres@localhost:5432/users_db?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", defaultJWTSecret),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
		GRPCAddress:    getEnv("GRPC_ADDRESS", ":9090"),
	}

	if secs, e := strconv.Atoi(getEnv("SERVER_TIMEOUT", "30")); e != nil {
		return nil, fmt.Errorf("parse SERVER_TIMEOUT: %w", e)
	} else {
		cfg.ServerTimeout = time.Duration(secs) * time.Second
	}

	if maxConns, e := strconv.Atoi(getEnv("DB_MAX_CONNS", "25")); e != nil {
		return nil, fmt.Errorf("parse DB_MAX_CONNS: %w", e)
	} else {
		cfg.DBMaxConns = int32(maxConns)
	}

	if minConns, e := strconv.Atoi(getEnv("DB_MIN_CONNS", "5")); e != nil {
		return nil, fmt.Errorf("parse DB_MIN_CONNS: %w", e)
	} else {
		cfg.DBMinConns = int32(minConns)
	}

	if accessMins, e := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRY", "15")); e != nil {
		return nil, fmt.Errorf("parse JWT_ACCESS_EXPIRY: %w", e)
	} else {
		cfg.JWTAccessExpiry = time.Duration(accessMins) * time.Minute
	}

	if refreshDays, e := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY", "7")); e != nil {
		return nil, fmt.Errorf("parse JWT_REFRESH_EXPIRY: %w", e)
	} else {
		cfg.JWTRefreshExpiry = time.Duration(refreshDays) * 24 * time.Hour
	}

	insecureSecret := cfg.JWTSecret == "" || cfg.JWTSecret == defaultJWTSecret
	if insecureSecret && cfg.AppEnv != "development" && cfg.AppEnv != "test" {
		return nil, errors.New("JWT_SECRET must be set to a strong secret in non-development environments")
	}
	if insecureSecret {
		fmt.Println("WARNING: Using default JWT_SECRET. Set a strong secret in production.")
	}

	brokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	cfg.KafkaBrokers = splitCSV(brokers)
	cfg.KafkaTopic = getEnv("KAFKA_TOPIC", "conversions")
	cfg.KafkaGroupID = getEnv("KAFKA_GROUP_ID", "api-server")

	cfg.RedisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.RedisPassword = getEnv("REDIS_PASSWORD", "")

	if redisDB, e := strconv.Atoi(getEnv("REDIS_DB", "0")); e != nil {
		return nil, fmt.Errorf("parse REDIS_DB: %w", e)
	} else {
		cfg.RedisDB = redisDB
	}

	if cacheTTLsecs, e := strconv.Atoi(getEnv("CACHE_TTL", "600")); e != nil {
		return nil, fmt.Errorf("parse CACHE_TTL: %w", e)
	} else {
		cfg.CacheTTL = time.Duration(cacheTTLsecs) * time.Second
	}

	cfg.LogLevel = getEnv("LOG_LEVEL", "info")
	cfg.LogFormat = getEnv("LOG_FORMAT", "json")
	cfg.LogKafkaTopic = getEnv("LOG_KAFKA_TOPIC", "app-logs")
	cfg.LogKafkaGroupID = getEnv("LOG_KAFKA_GROUP_ID", "log-consumer")

	if authRL, e := strconv.Atoi(getEnv("AUTH_RATE_LIMIT", "10")); e != nil {
		return nil, fmt.Errorf("parse AUTH_RATE_LIMIT: %w", e)
	} else {
		cfg.AuthRateLimit = authRL
	}

	if apiRL, e := strconv.Atoi(getEnv("API_RATE_LIMIT", "60")); e != nil {
		return nil, fmt.Errorf("parse API_RATE_LIMIT: %w", e)
	} else {
		cfg.APIRateLimit = apiRL
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	var result []string
	for _, v := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
