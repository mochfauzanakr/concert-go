package routes

import (
	"concert-go/internal/config"
	"concert-go/internal/handler"

	"github.com/gin-gonic/gin"
)

// DevRoute sets up testing utility routes.
// TODO: delete it in when it goes to main
func DevRoute(router *gin.RouterGroup, cfg *config.Config) {
	devHandler := handler.NewDevHandler(cfg)

	dev := router.Group("/dev")
	{
		dev.POST("/hash-password", devHandler.HashPassword)
		dev.POST("/aes-encrypt", devHandler.AESEncrypt)
		dev.POST("/aes-decrypt", devHandler.AESDecrypt)
	}
}