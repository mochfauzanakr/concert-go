package routes

import (
	"concert-go/internal/config"
	"concert-go/internal/handler"
	"concert-go/internal/middleware"
	"concert-go/internal/repository/postgres"
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.RouterGroup, cfg *config.Config) {
	userRepo := postgres.NewUserRepository(config.SqlDB)
	roleRepo := postgres.NewRoleRepository(config.SqlDB)
	inviteRepo := postgres.NewInviteRepository(config.SqlDB)
	emailSender := util.NewEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)

	userUsecase := usecase.NewUserUsecase(userRepo, roleRepo, inviteRepo, emailSender, cfg)
	userHandler := handler.NewUserHandler(userUsecase)

	users := router.Group("/users")
	{
		// Public routes to validate and accept invitation
		users.GET("/validate-invite", userHandler.ValidateInvite)
		users.POST("/accept-invite", userHandler.AcceptInvite)

		// Protected routes
		protected := users.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			
			protected.GET("", middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN"), userHandler.GetAllUsers)
			protected.PUT("/:id", middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN"), userHandler.AdminUpdateUser)
			
			protected.POST("/invite", middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN", "ORGANIZER"), userHandler.InviteUser)
			
			protected.PATCH("/:id/status", middleware.RequireRole(roleRepo, "SUPER_ADMIN", "ADMIN"), userHandler.UpdateUserStatus)
		}
	}
}
