package cache

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"kahootclone/internal/config"
)

// RedisClient wraps the go-redis client.
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient creates a new Redis client from the application config.
func NewRedisClient(ctx context.Context, cfg *config.Config) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Verify connectivity
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Warn("Redis ping failed — leaderboard features may be unavailable", "error", err.Error(), "addr", cfg.RedisAddr)
		// Don't return error; allow the app to start without Redis for basic operations
	} else {
		slog.Info("Redis client connected", "addr", cfg.RedisAddr)
	}

	return &RedisClient{Client: rdb}, nil
}

// Close closes the Redis connection.
func (r *RedisClient) Close() error {
	return r.Client.Close()
}
