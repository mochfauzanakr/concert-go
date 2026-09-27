package routes

import (
	"concert-go/internal/config"
	"concert-go/internal/handler"
	"concert-go/internal/middleware"
	"concert-go/internal/repository/postgres"
	"concert-go/internal/usecase"

	"github.com/gin-gonic/gin"
)

func AuthRoute(router *gin.RouterGroup, cfg *config.Config) {
	userRepo := postgres.NewUserRepository(config.SqlDB)
	sessionRepo := postgres.NewSessionRepository(config.SqlDB)
	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, cfg)
	authHandler := handler.NewAuthHandler(authUsecase)

	auth := router.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.RefreshToken)
	auth.POST("/logout", authHandler.Logout)

	protected := auth.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protected.GET("/me", authHandler.Me)
}
