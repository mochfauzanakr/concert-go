package middleware

import (
	"fmt"
	"net/http"
	"time"

	"concert-go/internal/domain/payload/response"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

type KeyFunc func(c *gin.Context) string

// RateLimit limits requests by client IP for a given scope
func RateLimit(store util.RateLimitStore, scope string, limit int, window time.Duration) gin.HandlerFunc {
	return RateLimitWithKey(store, limit, window, func(c *gin.Context) string {
		return util.KeyByIP(c, scope)
	})
}

// RateLimitWithKey limits requests using a custom key generator
func RateLimitWithKey(store util.RateLimitStore, limit int, window time.Duration, keyFunc KeyFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if store == nil {
			c.Next()
			return
		}

		key := keyFunc(c)
		state, err := store.Hit(c.Request.Context(), key, window)
		if err != nil {
			// Fail-open per dev-rules: allow request if Redis is unreachable
			c.Next()
			return
		}

		if state.Count > limit {
			retrySeconds := int(state.RetryAfter.Seconds())
			if retrySeconds <= 0 {
				retrySeconds = 1
			}
			c.Header("Retry-After", fmt.Sprintf("%d", retrySeconds))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, response.ErrorResponse{
				Message: "Too many requests. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
