package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	client *redis.Client
}

func NewRedisLimiter(addr string) *RedisLimiter {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisLimiter{client: client}
}

const tokenBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local bucket = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = tonumber(bucket[1])
local last_refill = tonumber(bucket[2])

if tokens == nil then
    tokens = capacity
    last_refill = now
end

local elapsed = now - last_refill
tokens = math.min(capacity, tokens + elapsed * refill_rate)

local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
redis.call("EXPIRE", key, 3600)

return allowed
`

// Allow checks whether a request for key is allowed, given a specific capacity and refill rate.
func (rl *RedisLimiter) Allow(ctx context.Context, key string, capacity, refillRate float64) (bool, error) {
	now := float64(time.Now().UnixNano()) / 1e9

	result, err := rl.client.Eval(ctx, tokenBucketScript,
		[]string{"ratelimit:" + key},
		capacity, refillRate, now,
	).Result()
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}