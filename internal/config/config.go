package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	RedisURL string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://zishan044:password@localhost:5432/url_shortener"),
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}