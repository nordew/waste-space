package main

import (
	"log"
	"waste-space/internal/app"
)

// @title Waste Space API
// @version 1.0
// @description REST API for waste management services
// @host localhost:8080
// @BasePath /
func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
