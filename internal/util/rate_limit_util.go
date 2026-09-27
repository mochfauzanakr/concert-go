package util

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimitState struct {
	Count      int
	RetryAfter time.Duration
}

type RateLimitStore interface {
	Hit(ctx context.Context, key string, window time.Duration) (RateLimitState, error)
}

type redisRateLimitStore struct {
	client *redis.Client
	script *redis.Script
}

// NewRedisRateLimitStore creates a new Redis-backed rate limit store using an atomic Lua script
func NewRedisRateLimitStore(client *redis.Client) RateLimitStore {
	// Lua script: INCR and EXPIRE atomically. Returns {count, ttl_in_seconds}
	scriptSource := `
		local count = redis.call("INCR", KEYS[1])
		if count == 1 then
			redis.call("EXPIRE", KEYS[1], ARGV[1])
		end
		return {count, redis.call("TTL", KEYS[1])}
	`
	return &redisRateLimitStore{
		client: client,
		script: redis.NewScript(scriptSource),
	}
}

func (s *redisRateLimitStore) Hit(ctx context.Context, key string, window time.Duration) (RateLimitState, error) {
	windowSeconds := int(window.Seconds())
	if windowSeconds <= 0 {
		windowSeconds = 1
	}

	result, err := s.script.Run(ctx, s.client, []string{key}, windowSeconds).Result()
	if err != nil {
		return RateLimitState{}, err
	}

	// The script returns an array: [count, ttl_seconds]
	values, ok := result.([]interface{})
	if !ok || len(values) != 2 {
		return RateLimitState{Count: 1, RetryAfter: window}, nil
	}

	count, _ := values[0].(int64)
	ttlSecs, _ := values[1].(int64)

	return RateLimitState{
		Count:      int(count),
		RetryAfter: time.Duration(ttlSecs) * time.Second,
	}, nil
}

// --- Key Builders ---

func KeyByIP(c *gin.Context, scope string) string {
	return "rl:v1:" + scope + ":ip:" + c.ClientIP()
}

// KeyByEmail normalizes the email to prevent bypassing limits via casing (e.g., User@ vs user@)
func KeyByEmail(scope, email string) string {
	return "rl:v1:" + scope + ":email:" + strings.ToLower(strings.TrimSpace(email))
}

func KeyByParam(scope, param string) string {
	return "rl:v1:" + scope + ":param:" + param
}
