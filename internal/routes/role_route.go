package routes

import (
	"concert-go/internal/config"
	"concert-go/internal/handler"
	"concert-go/internal/repository/postgres"
	"concert-go/internal/usecase"

	"github.com/gin-gonic/gin"
)

func RoleRoute(router *gin.RouterGroup) {
	roleRepo := postgres.NewRoleRepository(config.SqlDB)
	roleUsecase := usecase.NewRoleUsecase(roleRepo)
	roleHandler := handler.NewRoleHandler(roleUsecase)

	router.GET("/roles", roleHandler.GetRoles)
}
