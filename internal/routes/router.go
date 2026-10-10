package routes

import (
	"concert-go/internal/config"
	"concert-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.New()

	_ = r.SetTrustedProxies(nil)

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	api := r.Group("/api")
	{
		HelloRoute(api)
		AuthRoute(api, cfg)

		// TODO: delete it in when it goes to main
		DevRoute(api, cfg)
	}

	return r
}

