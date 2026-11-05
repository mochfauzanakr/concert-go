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

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.New()

	_ = r.SetTrustedProxies(nil)

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	if cfg.LocalStoragePath != "" {
		r.Static("/uploads", cfg.LocalStoragePath)
	}

	helloRepo := postgres.NewHelloRepository(config.SqlDB)
	userRepo := postgres.NewUserRepository(config.SqlDB)
	roleRepo := postgres.NewRoleRepository(config.SqlDB)
	sessionRepo := postgres.NewSessionRepository(config.SqlDB)
	inviteRepo := postgres.NewInviteRepository(config.SqlDB)
	organizerRepo := postgres.NewOrganizerRepository(config.SqlDB)
	otpRepo := redis.NewOTPRepository(config.RedisClient)
	emailSender := util.NewEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)
	storageService := util.NewStorageService(cfg)

	helloUsecase := usecase.NewHelloUsecase(helloRepo)
	roleUsecase := usecase.NewRoleUsecase(roleRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, otpRepo, roleRepo, emailSender, cfg)
	userUsecase := usecase.NewUserUsecase(userRepo, roleRepo, inviteRepo, sessionRepo, emailSender, storageService, cfg)
	organizerUsecase := usecase.NewOrganizerUsecase(organizerRepo, userRepo, storageService)

	helloHandler := handler.NewHelloHandler(helloUsecase)
	roleHandler := handler.NewRoleHandler(roleUsecase)
	authHandler := handler.NewAuthHandler(authUsecase, cfg)
	userHandler := handler.NewUserHandler(userUsecase)
	organizerHandler := handler.NewOrganizerHandler(organizerUsecase)

	api := r.Group("/api")
	{
		HelloRoute(api, helloHandler)
		AuthRoute(api, authHandler, cfg.JWTSecret)
		RoleRoute(api, roleHandler)
		UserRoute(api, userHandler, cfg.JWTSecret, roleRepo)
		OrganizerRoute(api, organizerHandler, cfg.JWTSecret, roleRepo)

		// TODO: delete it in when it goes to main
		DevRoute(api, cfg)
	}

	return r
}

