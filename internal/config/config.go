package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	RedisURL         string
	RabbitMQURL      string
	JWTSecret        string
	JWTExpiry        time.Duration
	RateLimitRequests int
	RateLimitWindow   time.Duration
	AllowedOrigins   []string
	RequestTimeout   time.Duration
	Port             string
	Environment      string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	allowedOrigins := []string{
		getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080"),
	}

	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://zishan044:password@postgres:5432/url_shortener?sslmode=disable"),
		RedisURL:          getEnv("REDIS_URL", "redis://redis:6379/0"),
		RabbitMQURL:       getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		JWTSecret:         getEnv("JWT_SECRET", "supersecret"),
		JWTExpiry:         parseDuration(getEnv("JWT_EXPIRY", "72h"), 72*time.Hour),
		RateLimitRequests: parseInt(getEnv("RATE_LIMIT_REQUESTS", "100"), 100),
		RateLimitWindow:   parseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"), time.Minute),
		AllowedOrigins:    allowedOrigins,
		RequestTimeout:    parseDuration(getEnv("REQUEST_TIMEOUT", "30s"), 30*time.Second),
		Port:              getEnv("PORT", "8080"),
		Environment:       getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func parseInt(val string, fallback int) int {
	if v, err := strconv.Atoi(val); err == nil {
		return v
	}
	return fallback
}

func parseDuration(val string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(val); err == nil {
		return d
	}
	return fallback
}