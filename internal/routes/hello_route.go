package routes

import (
	"concert-go/internal/handler"

	"github.com/gin-gonic/gin"
)

func HelloRoute(router *gin.RouterGroup, helloHandler *handler.HelloHandler) {
	hello := router.Group("/hello")
	hello.GET("", helloHandler.SayHello)
}
