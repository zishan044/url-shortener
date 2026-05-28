package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

func Get[T any](ctx context.Context, key string) (*T, error) {
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var data T
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return &data, nil
}

func Set[T any](ctx context.Context, key string, value T, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := Client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set value in cache: %w", err)
	}

	return nil
}

func Delete(ctx context.Context, key string) error {
	if err := Client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete value from cache: %w", err)
	}
	return nil
}
