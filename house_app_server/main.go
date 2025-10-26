package main

import (
	"log"

	"house-app/clients"
	"house-app/controller"
	"house-app/service"

	"github.com/joho/godotenv"
)

// @title House App API
// @version 1.0
// @description API for house management including image and file storage.

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize repository
	repo, err := clients.NewMongoRepo()
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}

	// Initialize service
	service := service.NewService(repo)

	// Initialize API server
	server := controller.NewAPIServer(service)

	// Run server
	log.Println("Starting server on :8080")
	server.Run(":8080")
}
