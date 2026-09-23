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

// uploadDir is the landing zone every upload is written to. The background
// processor (see processor/image_processor.go) later moves files out of
// here into STORAGE_DIR.
const uploadDir = "./uploads"

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

	if err := checkSupportedExt(safeName); err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Trust the client's declared Content-Type unless it's missing or the
	// generic fallback many upload libraries send for a type they don't
	// recognize — in that case, derive it from the extension the same way
	// BulkUploadImages already does, rather than storing a useless value.
	contentType := file.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = contentTypeForExt(filepath.Ext(safeName))
	}

	return s.storeImage(src, safeName, safeName, contentType, file.Size, tags)
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

	// Walk through the source folder
	err := filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			slog.Error("Error accessing path", "path", path, "error", err)
			return nil // Continue processing other files
		}

		// Skip directories and anything that isn't a supported image
		if info.IsDir() || checkSupportedExt(info.Name()) != nil {
			return nil
		}

		response.TotalImages++

		src, err := os.Open(path)
		if err != nil {
			slog.Error("Failed to open file", "path", path, "error", err)
			response.FailedCount++
			response.FailedFiles = append(response.FailedFiles, path)
			return nil
		}
		defer src.Close()

		contentType := contentTypeForExt(filepath.Ext(info.Name()))
		if _, err := s.storeImage(src, info.Name(), path, contentType, info.Size(), tags); err != nil {
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

// checkSupportedExt returns an error unless filename's extension is one of
// models.SupportedImageExtensions or models.SupportedVideoExtensions.
func checkSupportedExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !models.SupportedImageExtensions[ext] && !models.SupportedVideoExtensions[ext] {
		return fmt.Errorf("unsupported file type %q: allowed types are jpg, jpeg, png, gif, bmp, webp, heic, heif, mp4, mov, m4v, 3gp, webm, avi", ext)
	}
	return nil
}

// contentTypeForExt returns the MIME type for a supported image or video
// extension (case-insensitive), falling back to a generic image default
// for anything else (checkSupportedExt is expected to have already
// rejected truly unsupported extensions before this is called).
func contentTypeForExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	case ".webp":
		return "image/webp"
	case ".heic":
		return "image/heic"
	case ".heif":
		return "image/heif"
	case ".mov":
		return "video/quicktime"
	case ".m4v":
		return "video/x-m4v"
	case ".3gp":
		return "video/3gpp"
	case ".webm":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	case ".mp4":
		return "video/mp4"
	default:
		return "image/jpeg"
	}
}

// storeImage copies src into uploadDir under a unique, timestamp-prefixed
// name and saves the resulting Image record. Shared by UploadImage (a
// single multipart file) and BulkUploadImages (files read from disk) —
// those two only differ in where the bytes and metadata come from.
func (s *Service) storeImage(src io.Reader, filename, originalPath, contentType string, size int64, tags []string) (*models.Image, error) {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	storagePath := filepath.Join(uploadDir, fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename))

	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	image := models.Image{
		Filename:     filename,
		OriginalPath: originalPath,
		StoragePath:  storagePath,
		Size:         size,
		ContentType:  contentType,
		Tags:         tags,
		CreatedAt:    time.Now(),
	}

	if _, err := s.repo.CreateImage(image); err != nil {
		return nil, fmt.Errorf("failed to save image record: %w", err)
	}

	return &image, nil
}
