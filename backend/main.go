package main

import (
	"log"

	"gestor-gastos/config"
	"gestor-gastos/controllers"
	"gestor-gastos/database"
	"gestor-gastos/models"
	"gestor-gastos/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer db.Close()

	if err := models.RunMigrations(db); err != nil {
		log.Fatalf("migrations error: %v", err)
	}

	handler := controllers.NewHandler(db, cfg.JWTSecret)
	router := routes.Setup(cfg, handler)

	if err := router.Run(":" + cfg.APIPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
