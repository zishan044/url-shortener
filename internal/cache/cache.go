package cache

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Connect(redisURL string) {

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v\n", err)
	}

	Client = redis.NewClient(opts)

	if err := Client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v\n", err)
	}

	log.Println("Connected to Redis!")
}
