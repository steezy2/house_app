package service

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"house-app/models"
)

// Service handles the business logic for the house app.
type Service struct {
	repo models.ImageRepository
}

// NewService creates a new service with the given repository.
func NewService(repo models.ImageRepository) *Service {
	return &Service{repo: repo}
}

// CreateImage creates a new image record.
func (s *Service) CreateImage(image models.Image) (*models.Image, error) {
	_, err := s.repo.CreateImage(image)
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImages retrieves all images.
func (s *Service) GetImages() ([]models.Image, error) {
	return s.repo.GetImages()
}

// SearchImages searches for images.
func (s *Service) SearchImages(query string) ([]models.Image, error) {
	return s.repo.SearchImages(query)
}

// UploadImage handles single image upload
func (s *Service) UploadImage(file *multipart.FileHeader, tags []string) (*models.Image, error) {
	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	storagePath := filepath.Join(uploadDir, filename)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create image record
	image := models.Image{
		Filename:     file.Filename,
		OriginalPath: file.Filename,
		StoragePath:  storagePath,
		Size:         file.Size,
		ContentType:  file.Header.Get("Content-Type"),
		Tags:         tags,
		CreatedAt:    time.Now(),
	}

	_, err = s.repo.CreateImage(image)
	if err != nil {
		return nil, fmt.Errorf("failed to save image record: %w", err)
	}

	return &image, nil
}

// BulkUploadImages handles bulk image upload from a source folder
func (s *Service) BulkUploadImages(sourceFolder string, tags []string) (*models.BulkUploadResponse, error) {
	response := &models.BulkUploadResponse{
		FailedFiles: make([]string, 0),
	}

	// Check if source folder exists
	if _, err := os.Stat(sourceFolder); os.IsNotExist(err) {
		return nil, fmt.Errorf("source folder does not exist: %s", sourceFolder)
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Supported image extensions
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
	}

	// Walk through the source folder
	err := filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v", path, err)
			return nil // Continue processing other files
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file has supported extension
		ext := filepath.Ext(info.Name())
		if !supportedExts[ext] {
			return nil // Skip non-image files
		}

		response.TotalImages++

		// Copy file to uploads directory
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), info.Name())
		storagePath := filepath.Join(uploadDir, filename)

		src, err := os.Open(path)
		if err != nil {
			log.Printf("Failed to open file %s: %v", path, err)
			response.FailedCount++
			response.FailedFiles = append(response.FailedFiles, path)
			return nil
		}
		defer src.Close()

		dst, err := os.Create(storagePath)
		if err != nil {
			log.Printf("Failed to create destination file %s: %v", storagePath, err)
			response.FailedCount++
			response.FailedFiles = append(response.FailedFiles, path)
			return nil
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			log.Printf("Failed to copy file %s: %v", path, err)
			response.FailedCount++
			response.FailedFiles = append(response.FailedFiles, path)
			return nil
		}

		// Determine content type
		contentType := "image/jpeg"
		switch ext {
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".bmp":
			contentType = "image/bmp"
		case ".webp":
			contentType = "image/webp"
		}

		// Create image record
		image := models.Image{
			Filename:     info.Name(),
			OriginalPath: path,
			StoragePath:  storagePath,
			Size:         info.Size(),
			ContentType:  contentType,
			Tags:         tags,
			CreatedAt:    time.Now(),
		}

		_, err = s.repo.CreateImage(image)
		if err != nil {
			log.Printf("Failed to save image record for %s: %v", path, err)
			response.FailedCount++
			response.FailedFiles = append(response.FailedFiles, path)
			return nil
		}

		response.SuccessfulCount++
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through source folder: %w", err)
	}

	return response, nil
}
