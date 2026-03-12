package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type ICacheClient interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
}

type redisClient struct {
	ctx    context.Context
	client *redis.Client
}

func NewRedisClient(ctx context.Context, client *redis.Client) (ICacheClient, error) {
	return &redisClient{ctx: ctx, client: client}, nil
}

func (r *redisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

func (r *redisClient) Set(key string, value string, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}
