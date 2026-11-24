package routes

import (
	"concert-go/internal/handler"
	"concert-go/internal/middleware"
	"concert-go/internal/repository"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.RouterGroup, userHandler *handler.UserHandler, jwtSecret string, roleRepo repository.RoleRepository) {
	users := router.Group("/users")
	{
		users.GET("/validate-invite", userHandler.ValidateInvite)
		users.POST("/accept-invite", userHandler.AcceptInvite)

		protected := users.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			protected.GET("", middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN"), userHandler.GetAllUsers)
			protected.GET("/:id", middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN"), userHandler.GetUserDetail)

			protected.PUT("/me", userHandler.UpdateProfile)
			protected.PUT("/:id", middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN"), userHandler.AdminUpdateUser)

			protected.POST("/invite", middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN", "ORGANIZER"), userHandler.InviteUser)

			protected.PATCH("/:id/status", middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN"), userHandler.UpdateUserStatus)
		}
	}
}
