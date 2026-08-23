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

local ttl = redis.call("TTL", KEYS[1])

if ttl == -1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
    ttl = tonumber(ARGV[1])
end

if count <= tonumber(ARGV[2]) then
    return {1, ttl}
end

return {0, ttl}
`)

type RateLimitInterface interface {
	RateLimit(ctx context.Context, ip string) (
		status RateStatus,
		retryAfter int64,
		err error,
	)
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

func (r *RateLimitRedisRepository) RateLimit(ctx context.Context, ip string) (RateStatus, int64, error) {
	if err := ctx.Err(); err != nil {
		return RateStatusError, 0, err
	}
	window := 60 * time.Second
	values, err := script.Run(
		ctx,
		r.RedisClient,
		[]string{rateKey + ip},
		int64(window.Seconds()),
		5,
	).Slice()

	if err != nil {
		return RateStatusError, 0, err
	}

	if len(values) != 2 {
		return RateStatusError, 0, fmt.Errorf("unexpected result length: %d", len(values))
	}

	allowed, ok := values[0].(int64)
	if !ok {
		return RateStatusError, 0, fmt.Errorf("unexpected allowed type: %T", values[0])
	}

	ttl, ok := values[1].(int64)
	if !ok {
		return RateStatusError, 0, fmt.Errorf("unexpected TTL type: %T", values[1])
	}

	if allowed == 1 {
		return RateStatusAllowed, 0, nil
	}

	return RateStatusRetryAfter, ttl, nil
}
