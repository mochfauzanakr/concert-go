package routes

import (
	"time"

	"concert-go/internal/handler"
	"concert-go/internal/middleware"
	"concert-go/internal/repository"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

func OrganizerRoute(router *gin.RouterGroup, organizerHandler *handler.OrganizerHandler, jwtSecret string, roleRepo repository.RoleRepository, rateLimiter util.RateLimitStore) {
	organizers := router.Group("/organizers")
	{
		organizers.POST("/apply", middleware.RateLimit(rateLimiter, "organizer_apply", 5, 15*time.Minute), organizerHandler.ApplyForPartnership)

		admin := organizers.Group("")
		admin.Use(middleware.AuthMiddleware(jwtSecret))
		admin.Use(middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN"))
		{
			admin.GET("/applications", organizerHandler.GetApplications)
			admin.PATCH("/applications/:id/review", organizerHandler.ReviewApplication)
		}

		protected := organizers.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		protected.Use(middleware.RequireRole(roleRepo, "ORGANIZER"))
		{
			protected.GET("/profile", organizerHandler.GetProfile)
			protected.PUT("/profile", organizerHandler.UpsertProfile)
			protected.GET("/staff", organizerHandler.GetStaffList)
			protected.PATCH("/staff/:id/status", organizerHandler.UpdateStaffStatus)
		}

		organizers.GET("/:id", organizerHandler.GetPublicProfile)
	}
}
