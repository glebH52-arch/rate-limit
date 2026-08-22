package redis_client

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(redisAddr string, redisPassword string, redisDB int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}
