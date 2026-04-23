package database

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	"mmorpg/server/internal/config"
)

// ConnectRedis creates and tests a Redis client connection.
func ConnectRedis(cfg config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: "",
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Println("Connected to Redis")
	return client, nil
}
