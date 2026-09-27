package middleware

import (
	"strings"

	"concert-go/internal/domain/model"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

const UserSessionContextKey = "user_session"

// AuthMiddleware validates JWT Bearer token and injects UserSession into context
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			util.RespondUnauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			util.RespondUnauthorized(c, "Invalid Authorization header format")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := util.ValidateJWT(jwtSecret, tokenStr)
		if err != nil {
			util.RespondUnauthorized(c, err.Error())
			c.Abort()
			return
		}

		session := model.UserSession{
			UserID: claims.UserID,
			Email:  claims.Email,
			RoleID: claims.RoleID,
		}

		c.Set(UserSessionContextKey, session)
		c.Next()
	}
}

// GetUserSession extracts UserSession from gin.Context
func GetUserSession(c *gin.Context) (*model.UserSession, bool) {
	val, exists := c.Get(UserSessionContextKey)
	if !exists {
		return nil, false
	}
	session, ok := val.(model.UserSession)
	return &session, ok
}
