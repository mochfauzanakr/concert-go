package main

import (
	"log"

	"concert-go/internal/config"
	"concert-go/internal/domain/entity"
	"concert-go/internal/routes"
)

func main() {
	// Load Configuration
	cfg := config.LoadConfig()

	// Initialize Database (GORM, pure SQL, and Redis)
	config.InitDatabase(cfg)
	defer config.SqlDB.Close()
	defer config.RedisClient.Close()

	if cfg.EnableMigration {
		log.Println("Running database migrations...")
		err := config.GormDB.AutoMigrate(
			&entity.User{},
			&entity.Session{},
			&entity.Role{},
		)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}

	// Setup Gin Router
	r := routes.SetupRouter(cfg)

	// Start server
	log.Printf("Starting server on port %s...", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
