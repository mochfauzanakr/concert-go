package util

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ParseQueryInt extracts a positive integer query parameter with a default fallback
func ParseQueryInt(c *gin.Context, key string, defaultVal int) int {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return defaultVal
	}
	if v, err := strconv.Atoi(raw); err == nil && v > 0 {
		return v
	}
	return defaultVal
}
