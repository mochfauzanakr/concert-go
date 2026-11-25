package routes

import (
	"concert-go/internal/handler"
	"concert-go/internal/middleware"
	"concert-go/internal/repository"

	"github.com/gin-gonic/gin"
)

func EventRoute(router *gin.RouterGroup, eventHandler *handler.EventHandler, jwtSecret string, roleRepo repository.RoleRepository) {
	events := router.Group("/events")
	{
		events.GET("", eventHandler.GetPublicEvents)
		events.GET("/:id", eventHandler.GetPublicEventDetail)

		organizerEvents := events.Group("")
		organizerEvents.Use(middleware.AuthMiddleware(jwtSecret))
		organizerEvents.Use(middleware.RequireRole(roleRepo, "ORGANIZER"))
		{
			organizerEvents.POST("", eventHandler.CreateEvent)
			organizerEvents.PUT("/:id", eventHandler.UpdateEvent)
			organizerEvents.PATCH("/:id/status", eventHandler.UpdateEventStatus)
		}
	}

	organizers := router.Group("/organizers")
	organizers.Use(middleware.AuthMiddleware(jwtSecret))
	organizers.Use(middleware.RequireRole(roleRepo, "ORGANIZER"))
	{
		organizers.GET("/events", eventHandler.GetOrganizerEvents)
	}
}
