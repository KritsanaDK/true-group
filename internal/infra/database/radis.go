package database

import (
	"context"
	"tdg/internal/infra/logger"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ConfigRedis struct {
	RedisEndPoint string
	Password      string
}

func ConnectRedis(
	ctx context.Context,
	cfg ConfigRedis,
) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisEndPoint,
		Password:     cfg.Password, // Add password if needed
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		logger.ErrorWithFields("failed to connect to Redis",
			zap.String("endpoint", cfg.RedisEndPoint),
			zap.String("error", err.Error()),
		)
		return nil, err
	} else {
		logger.InfoWithFields("connected to Redis successfully", zap.String("endpoint", cfg.RedisEndPoint))

	}

	return client, nil

}
