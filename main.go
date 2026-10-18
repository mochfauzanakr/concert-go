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
		// Seed default roles if they don't exist
		var roleCount int64
		config.GormDB.Model(&entity.Role{}).Count(&roleCount)
		if roleCount == 0 {
			defaultRoles := []entity.Role{
				{ID: 1, Name: "SUPER_ADMIN"},
				{ID: 2, Name: "ADMIN"},
				{ID: 3, Name: "ORGANIZER"},
				{ID: 4, Name: "USER"},
				{ID: 5, Name: "STAFF"},
			}
			if err := config.GormDB.Create(&defaultRoles).Error; err != nil {
				log.Printf("Role seed warning: %v", err)
			} else {
				log.Println("Default roles seeded successfully")
			}
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
