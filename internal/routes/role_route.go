package routes

import (
	"concert-go/internal/handler"

	"github.com/gin-gonic/gin"
)

func RoleRoute(router *gin.RouterGroup, roleHandler *handler.RoleHandler) {
	router.GET("/roles", roleHandler.GetRoles)
}
