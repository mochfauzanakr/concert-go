package routes

import (
	"concert-go/internal/handler"
	"concert-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

func AuthRoute(router *gin.RouterGroup, authHandler *handler.AuthHandler, jwtSecret string) {
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
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	protected.GET("/me", authHandler.Me)
	protected.POST("/change-password", authHandler.ChangePassword)
}
