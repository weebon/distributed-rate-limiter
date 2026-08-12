package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisSlidingWindow struct {
	client *redis.Client
}

func NewRedisSlidingWindow(addr string) *RedisSlidingWindow {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisSlidingWindow{client: client}
}

const slidingWindowScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local window_start = now - window

redis.call("ZREMRANGEBYSCORE", key, "-inf", window_start)

local count = redis.call("ZCARD", key)

local allowed = 0
if count < limit then
    redis.call("ZADD", key, now, tostring(now) .. "-" .. tostring(math.random()))
    allowed = 1
end

redis.call("EXPIRE", key, math.ceil(window))

return allowed
`

 
func (sw *RedisSlidingWindow) Allow(ctx context.Context, key string, limit int64, windowSecs float64) (bool, error) {
	now := float64(time.Now().UnixNano()) / 1e9

	result, err := sw.client.Eval(ctx, slidingWindowScript,
		[]string{"sliding:" + key},
		limit, windowSecs, now,
	).Result()
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}