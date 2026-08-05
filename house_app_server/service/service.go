package service

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"house-app/models"
)

// supportedImageExts are the file extensions accepted by upload endpoints.
var supportedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
}

// ParseTags splits a comma-separated tags string into a trimmed, non-empty
// tag slice (e.g. "vacation, beach ,,2025" -> ["vacation", "beach", "2025"]).
// Returns nil if tagsStr is empty or contains only separators.
func ParseTags(tagsStr string) []string {
	if tagsStr == "" {
		return nil
	}
	var tags []string
	for _, tag := range strings.Split(tagsStr, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

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
	// Strip any directory components the client may have sent, so the
	// filename can't be used to escape uploadDir (e.g. "../../evil.jpg").
	safeName := filepath.Base(file.Filename)

	ext := strings.ToLower(filepath.Ext(safeName))
	if !supportedImageExts[ext] {
		return nil, fmt.Errorf("unsupported file type %q: allowed types are jpg, jpeg, png, gif, bmp, webp", ext)
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), safeName)
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
		Filename:     safeName,
		OriginalPath: safeName,
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

	// Walk through the source folder
	err := filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			slog.Error("Error accessing path", "path", path, "error", err)
			return nil // Continue processing other files
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file has supported extension
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if !supportedImageExts[ext] {
			return nil // Skip non-image files
		}

		response.TotalImages++

		// Copy file to uploads directory
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), info.Name())
		storagePath := filepath.Join(uploadDir, filename)

		src, err := os.Open(path)
		if err != nil {
			slog.Error("Failed to open file", "path", path, "error", err)
			response.FailedCount++
			response.FailedFiles = append(response.FailedFiles, path)
			return nil
		}
		defer src.Close()

		dst, err := os.Create(storagePath)
		if err != nil {
			slog.Error("Failed to create destination file", "storagePath", storagePath, "error", err)
			response.FailedCount++
			response.FailedFiles = append(response.FailedFiles, path)
			return nil
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			slog.Error("Failed to copy file", "path", path, "error", err)
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
			slog.Error("Failed to save image record", "path", path, "error", err)
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
