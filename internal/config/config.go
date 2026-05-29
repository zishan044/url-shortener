package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	RedisURL string
	RabbitMQURL string
	JWTSecret string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://zishan044:password@postgres:5432/url_shortener?sslmode=disable"),
		RedisURL: getEnv("REDIS_URL", "redis://redis:6379/0"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		JWTSecret: getEnv("JWT_SECRET", "supersecret"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}