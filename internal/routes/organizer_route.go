package routes

import (
	"concert-go/internal/handler"
	"concert-go/internal/middleware"
	"concert-go/internal/repository"

	"github.com/gin-gonic/gin"
)

func OrganizerRoute(router *gin.RouterGroup, organizerHandler *handler.OrganizerHandler, jwtSecret string, roleRepo repository.RoleRepository) {
	organizers := router.Group("/organizers")
	{
		protected := organizers.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		protected.Use(middleware.RequireRole(roleRepo, "ORGANIZER"))
		{
			protected.PUT("/profile", organizerHandler.UpsertProfile)
		}
	}
}
