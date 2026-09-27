package routes

import (
	"concert-go/internal/config"
	"concert-go/internal/handler"
	"concert-go/internal/repository/postgres"
	"concert-go/internal/usecase"

	"github.com/gin-gonic/gin"
)

func HelloRoute(router *gin.RouterGroup) {
	helloRepo := postgres.NewHelloRepository(config.SqlDB)
	helloUsecase := usecase.NewHelloUsecase(helloRepo)
	helloHandler := handler.NewHelloHandler(helloUsecase)

	hello := router.Group("/hello")
	hello.GET("", helloHandler.SayHello)
}
