package main

import (
	"log"
	"net/http"
	"os"

	"amnezia-mikrotik-panel/internal/config"
	"amnezia-mikrotik-panel/internal/database"
	"amnezia-mikrotik-panel/internal/web"
	"amnezia-mikrotik-panel/internal/worker"
)

func main() {
	log.Println("Bootstrapping Amnezia Mikrotik Panel...")

	cfg := config.Load()

	// Ensure DB directory exists
	if cfg.DBPath == "/app/data/panel.db" {
		os.MkdirAll("/app/data", 0755)
	}

	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start background worker for Delta bandwidth accounting
	worker.StartCron()

	// Register HTTP Routes
	web.RegisterRoutes()

	addr := ":" + cfg.WebPort
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
