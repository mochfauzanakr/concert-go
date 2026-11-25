package routes

import (
	"concert-go/internal/handler"

	"github.com/gin-gonic/gin"
)

func EventRoute(router *gin.RouterGroup, eventHandler *handler.EventHandler) {
	events := router.Group("/events")
	{
		events.GET("", eventHandler.GetPublicEvents)
		events.GET("/:id", eventHandler.GetPublicEventDetail)
	}
}
