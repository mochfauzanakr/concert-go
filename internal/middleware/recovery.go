package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"konserGo/internal/dto/response"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware catches panics and returns a 500 error matching rules
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v\n%s", err, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorResponse{
					Message: "Internal server error",
				})
			}
		}()
		c.Next()
	}
}
