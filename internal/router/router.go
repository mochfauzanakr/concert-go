package router

import (
	"konserGo/internal/config"
	"konserGo/internal/handler"
	"konserGo/internal/middleware"
	"konserGo/internal/repository"
	"konserGo/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes the Gin engine and registers routes
func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	// Initialization of repository -> service -> handler
	helloRepo := repository.NewHelloRepository(config.SqlDB)
	helloService := service.NewHelloService(helloRepo)
	helloHandler := handler.NewHelloHandler(helloService)

	// Grouping routes
	api := r.Group("/api/v1")
	{
		api.GET("/hello", helloHandler.SayHello)
	}

	return r
}
