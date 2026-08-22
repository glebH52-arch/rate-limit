package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var script = redis.NewScript(`
    local count = redis.call("INCR", KEYS[1])

    if count == 1 then
        redis.call("EXPIRE", KEYS[1], ARGV[1])
    end

    if count > 5 then
        return redis.call("TTL", KEYS[1])
    end

    return 1
`)

type RateLimitInterface interface {
	RateLimit(ctx context.Context, ip string) (RateStatus, error)
}

type RateLimitRedisRepository struct {
	RedisClient *redis.Client
}

func NewRateLimitRedisRepository(redisClient *redis.Client) *RateLimitRedisRepository {
	return &RateLimitRedisRepository{
		RedisClient: redisClient,
	}
}

type RateStatus string

const rateKey = "login_rate:"
const (
	RateStatusAllowed    RateStatus = "allowed"
	RateStatusRetryAfter RateStatus = "retryAfter"
	RateStatusError      RateStatus = "error"
)

func (r *RateLimitRedisRepository) RateLimit(ctx context.Context, ip string) (RateStatus, error) {
	if err := ctx.Err(); err != nil {
		return RateStatusError, err
	}
	window := 60 * time.Second
	value, err := script.Run(ctx, r.RedisClient, []string{rateKey + ip}, int64(window.Seconds())).Result()
	if err != nil {
		return RateStatusError, err
	}

	result, ok := value.(int64)
	if !ok {
		return RateStatusError, fmt.Errorf("unexpected redis script result type: %T", value)
	}
	if result == 1 {
		return RateStatusAllowed, nil
	}

	retryAfter := time.Duration(result) * time.Second
	return RateStatusRetryAfter, fmt.Errorf("Попробуйте снова через: %s", retryAfter)
}
