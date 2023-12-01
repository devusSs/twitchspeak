package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config for redis
type Config struct {
	Host     string
	Port     uint
	Password string
	DB       uint
}

// Init initializes the redis client
func Init(cfg Config) error {
	redisDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       int(cfg.DB),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := redisDB.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("redis: error connecting to redis: %v", err)
	}

	return nil
}

// GetClient returns the redis client
func GetClient() *redis.Client {
	return redisDB
}

var (
	redisDB *redis.Client
)
