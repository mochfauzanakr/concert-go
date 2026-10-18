package middleware

import (
	"strings"

	"concert-go/internal/repository"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

// RequireRole ensures the authenticated user has one of the permitted role names fetched dynamically from DB
func RequireRole(roleRepo repository.RoleRepository, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, ok := GetUserSession(c)
		if !ok || session == nil {
			util.RespondUnauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		if session.RoleID == nil {
			util.RespondForbidden(c, "Access denied: user has no assigned role")
			c.Abort()
			return
		}

		role, err := roleRepo.FindByID(c.Request.Context(), *session.RoleID)
		if err != nil || role == nil {
			util.RespondForbidden(c, "Access denied: role not found")
			c.Abort()
			return
		}

		for _, allowed := range allowedRoles {
			if strings.EqualFold(role.Name, allowed) {
				c.Next()
				return
			}
		}

		util.RespondForbidden(c, "Access denied: insufficient permissions")
		c.Abort()
	}
}
