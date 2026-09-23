package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"house-app/backup"
	"house-app/clients"
	"house-app/controller"
	"house-app/docs"
	"house-app/processor"
	"house-app/service"

	"github.com/joho/godotenv"
)

// @title House App API
// @version 1.0
// @description API for house management including image and file storage.

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Initialize structured logger with text format for better readability
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		slog.Warn("No .env file found, using environment variables", "error", err)
	}

	// Set Swagger host dynamically based on environment variable
	swaggerHost := getEnv("SWAGGER_HOST", "localhost:8080")
	docs.SwaggerInfo.Host = swaggerHost
	slog.Info("Swagger configured", "host", swaggerHost)

	// Initialize repository
	repo, err := clients.NewMongoRepo()
	if err != nil {
		slog.Error("Failed to create repository", "error", err)
		os.Exit(1)
	}

	// Initialize service
	service := service.NewService(repo)

	// Initialize image processor
	uploadDir := getEnv("UPLOAD_DIR", "./uploads")
	storageDir := getEnv("STORAGE_DIR", "E:/house_app_storage")
	processingInterval := getEnvDuration("PROCESSING_INTERVAL", 5*time.Minute)

	backupDestinations := buildBackupDestinations()
	imageProcessor := processor.NewImageProcessor(uploadDir, storageDir, repo, backupDestinations)

	// Start image processor in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go imageProcessor.StartPeriodicProcessing(ctx, processingInterval)

	// Initialize API server
	server := controller.NewAPIServer(service)

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		slog.Info("Shutting down gracefully...")
		cancel()
		os.Exit(0)
	}()

	// Run server
	slog.Info("Starting server", "port", ":8080")
	server.Run(":8080")
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvDuration gets a duration from environment variable
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// buildBackupDestinations assembles the configured backup.Destinations
// from BACKUP_LOCAL_DIRS (a comma-separated list of local/network paths —
// 0, 1, 2, 3, or more) and the optional BACKUP_S3_* cloud settings. Either,
// both, or neither can be configured; see README.md's Environment
// Variables table.
func buildBackupDestinations() []backup.Destination {
	var destinations []backup.Destination

	if localDirs := os.Getenv("BACKUP_LOCAL_DIRS"); localDirs != "" {
		for _, dir := range strings.Split(localDirs, ",") {
			dir = strings.TrimSpace(dir)
			if dir == "" {
				continue
			}
			destinations = append(destinations, backup.NewLocalDestination(dir))
		}
	}

	s3Bucket := os.Getenv("BACKUP_S3_BUCKET")
	s3AccessKey := os.Getenv("BACKUP_S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("BACKUP_S3_SECRET_KEY")
	if s3Bucket != "" && s3AccessKey != "" && s3SecretKey != "" {
		s3Dest, err := backup.NewS3Destination(context.Background(), backup.S3Config{
			Endpoint:  os.Getenv("BACKUP_S3_ENDPOINT"),
			Bucket:    s3Bucket,
			AccessKey: s3AccessKey,
			SecretKey: s3SecretKey,
			Region:    getEnv("BACKUP_S3_REGION", "us-east-1"),
		})
		if err != nil {
			slog.Error("Failed to configure S3 backup destination, skipping it", "error", err)
		} else {
			destinations = append(destinations, s3Dest)
		}
	}

	if len(destinations) == 0 {
		slog.Warn("No backup destinations configured (BACKUP_LOCAL_DIRS/BACKUP_S3_*) — uploaded files exist only in STORAGE_DIR, with no redundancy")
		return destinations
	}

	names := make([]string, len(destinations))
	for i, d := range destinations {
		names[i] = d.Name()
	}
	slog.Info("Backup destinations configured", "destinations", names)

	return destinations
}
