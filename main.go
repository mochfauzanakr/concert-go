package main

import (
	"log"

	"konserGo/internal/config"
	"konserGo/internal/entity"
	"konserGo/internal/router"
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
		)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}

	// Setup Gin Router
	r := router.SetupRouter()

	// Start server
	log.Printf("Starting server on port %s...", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
