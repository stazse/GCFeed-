package infracache

import (
	"context"

	infraconfig "GCFeed/internal/infra/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient 创建 Redis 客户端。
func NewRedisClient(cfg infraconfig.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return client
}

// Ping 测试 Redis 是否连通。
func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}