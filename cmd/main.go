package main

import (
	"fmt"
	"log"

	"chatx-01-backend/internal/config"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Log configuration (be careful not to log secrets in production)
	log.Printf("Starting chatx-01-backend server")
	log.Printf("Database: %s", cfg.Postgres.Host+":"+fmt.Sprint(cfg.Postgres.Port))

	// Your application initialization code goes here
	fmt.Println("Server is ready to start...")

	// Example: Start HTTP server, initialize database, etc.
}
