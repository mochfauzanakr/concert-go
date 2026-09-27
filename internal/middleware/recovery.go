package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"concert-go/internal/domain/payload/response"

	"github.com/gin-gonic/gin"
)

func handlePanic(c *gin.Context, err any) {
	log.Printf("panic recovered: %v\n%s", err, debug.Stack())
	c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorResponse{
		Message: "Internal server error",
	})
}

// RecoveryMiddleware catches panics and returns a 500 error matching rules
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(handlePanic)
}
