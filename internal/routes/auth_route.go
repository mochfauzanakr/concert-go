package routes

import (
	"concert-go/internal/config"
	"concert-go/internal/handler"
	"concert-go/internal/middleware"
	"concert-go/internal/repository/postgres"
	"concert-go/internal/repository/redis"
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

func AuthRoute(router *gin.RouterGroup, cfg *config.Config) {
	userRepo := postgres.NewUserRepository(config.SqlDB)
	sessionRepo := postgres.NewSessionRepository(config.SqlDB)
	otpRepo := redis.NewOTPRepository(config.RedisClient)
	emailSender := util.NewEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)

	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, otpRepo, emailSender, cfg)
	authHandler := handler.NewAuthHandler(authUsecase, cfg)

	auth := router.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/verify-otp", authHandler.VerifyOTP)
	auth.POST("/login", authHandler.Login)
	auth.POST("/forgot-password", authHandler.ForgotPassword)
	auth.POST("/reset-password", authHandler.ResetPassword)

	auth.GET("/google/login", authHandler.GoogleLogin)
	auth.GET("/google/callback", authHandler.GoogleCallback)

	auth.POST("/refresh", authHandler.RefreshToken)
	auth.POST("/logout", authHandler.Logout)

	protected := auth.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protected.GET("/me", authHandler.Me)
}
