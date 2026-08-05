package processor

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"house-app/models"

	"github.com/rwcarlsen/goexif/exif"
)

// ImageProcessor handles automatic organization of uploaded images
type ImageProcessor struct {
	uploadDir      string
	storageBaseDir string
	repo           models.ImageRepository
	aiEnabled      bool
}

// NewImageProcessor creates a new image processor
func NewImageProcessor(uploadDir, storageBaseDir string, repo models.ImageRepository) *ImageProcessor {
	return &ImageProcessor{
		uploadDir:      uploadDir,
		storageBaseDir: storageBaseDir,
		repo:           repo,
		aiEnabled:      os.Getenv("ENABLE_AI_CATEGORIZATION") == "true",
	}
}

// ProcessingResult contains the result of processing an image
type ProcessingResult struct {
	OriginalPath string
	NewPath      string
	Category     string
	NewName      string
	Error        error
}

// StartPeriodicProcessing starts a timer that processes images every interval
func (p *ImageProcessor) StartPeriodicProcessing(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Started periodic image processing", "interval", interval, "uploadDir", p.uploadDir, "storageDir", p.storageBaseDir)

	// Run immediately on start
	p.ProcessAllImages()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping periodic image processing")
			return
		case <-ticker.C:
			p.ProcessAllImages()
		}
	}
}

// ProcessAllImages processes all images in the upload directory
func (p *ImageProcessor) ProcessAllImages() {
	slog.Info("Starting image processing cycle")

	files, err := os.ReadDir(p.uploadDir)
	if err != nil {
		slog.Error("Failed to read upload directory", "error", err, "dir", p.uploadDir)
		return
	}

	processed := 0
	failed := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if it's an image file
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if !isImageFile(ext) {
			continue
		}

		sourcePath := filepath.Join(p.uploadDir, file.Name())
		result := p.processImage(sourcePath)

		if result.Error != nil {
			slog.Error("Failed to process image", "file", file.Name(), "error", result.Error)
			failed++
		} else {
			slog.Info("Successfully processed image",
				"originalPath", result.OriginalPath,
				"newPath", result.NewPath,
				"category", result.Category,
				"newName", result.NewName)
			processed++
		}
	}

	slog.Info("Image processing cycle completed", "processed", processed, "failed", failed)
}

// processImage processes a single image
func (p *ImageProcessor) processImage(sourcePath string) ProcessingResult {
	result := ProcessingResult{
		OriginalPath: sourcePath,
	}

	// Extract metadata
	metadata, err := p.extractMetadata(sourcePath)
	if err != nil {
		slog.Warn("Failed to extract metadata, using defaults", "file", sourcePath, "error", err)
		metadata = &ImageMetadata{
			DateTime: time.Now(),
		}
	}

	// Determine category
	var category string
	if p.aiEnabled {
		category, err = p.categorizeWithAI(sourcePath, metadata)
		if err != nil {
			slog.Warn("AI categorization failed, using default", "file", sourcePath, "error", err)
			category = "uncategorized"
		}
	} else {
		category = p.categorizeByMetadata(metadata)
	}

	result.Category = category

	// Generate new filename
	newName := p.generateFileName(sourcePath, metadata, category)
	result.NewName = newName

	// Create destination directory: /e/house_app_storage/YYYY/MM/category/
	year := metadata.DateTime.Format("2006")
	month := metadata.DateTime.Format("01")
	destDir := filepath.Join(p.storageBaseDir, year, month, category)

	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		result.Error = fmt.Errorf("failed to create destination directory: %w", err)
		return result
	}

	destPath := filepath.Join(destDir, newName)

	// Move file. moveFile may rename destPath (appending _1, _2, ...) if a
	// file already exists there, so use the path it actually wrote to.
	finalPath, err := p.moveFile(sourcePath, destPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to move file: %w", err)
		return result
	}
	destPath = finalPath
	result.NewPath = destPath

	// Update database record
	if err := p.updateDatabaseRecord(sourcePath, destPath, category); err != nil {
		slog.Warn("Failed to update database record", "error", err, "file", sourcePath)
	}

	return result
}

// ImageMetadata holds extracted image metadata
type ImageMetadata struct {
	DateTime     time.Time
	Make         string
	Model        string
	Width        int
	Height       int
	Orientation  string
	OriginalName string
}

// extractMetadata extracts EXIF metadata from an image
func (p *ImageProcessor) extractMetadata(imagePath string) (*ImageMetadata, error) {
	metadata := &ImageMetadata{
		DateTime:     time.Now(),
		OriginalName: filepath.Base(imagePath),
	}

	// Open file for EXIF reading
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}

	// Try to extract EXIF data
	x, err := exif.Decode(file)

	// Explicitly close the file immediately after reading EXIF data
	// This is crucial on Windows to release the file handle
	file.Close()

	if err != nil {
		return metadata, nil // Return metadata with defaults if EXIF fails
	}

	// Extract DateTime
	if dt, err := x.DateTime(); err == nil {
		metadata.DateTime = dt
	}

	// Extract camera make
	if make, err := x.Get(exif.Make); err == nil {
		if makeStr, err := make.StringVal(); err == nil {
			metadata.Make = strings.TrimSpace(makeStr)
		}
	}

	// Extract camera model
	if model, err := x.Get(exif.Model); err == nil {
		if modelStr, err := model.StringVal(); err == nil {
			metadata.Model = strings.TrimSpace(modelStr)
		}
	}

	return metadata, nil
}

// categorizeWithAI uses AI to categorize the image
func (p *ImageProcessor) categorizeWithAI(imagePath string, metadata *ImageMetadata) (string, error) {
	// This is a placeholder for AI integration
	// You can integrate with OpenAI Vision API, Google Cloud Vision, etc.

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return p.categorizeByMetadata(metadata), nil
	}

	// Example: Use OpenAI Vision API to analyze the image
	category, err := p.callOpenAIVision(imagePath)
	if err != nil {
		return "", err
	}

	return category, nil
}

// callOpenAIVision calls OpenAI Vision API to categorize the image
func (p *ImageProcessor) callOpenAIVision(imagePath string) (string, error) {
	// Placeholder for OpenAI Vision API integration
	// This would encode the image as base64 and send to OpenAI API
	// For now, return a smart default based on filename patterns

	filename := strings.ToLower(filepath.Base(imagePath))

	// Simple keyword-based categorization as fallback
	keywords := map[string]string{
		"vacation":   "travel",
		"holiday":    "travel",
		"trip":       "travel",
		"beach":      "travel",
		"mountain":   "nature",
		"landscape":  "nature",
		"sunset":     "nature",
		"sunrise":    "nature",
		"portrait":   "people",
		"family":     "family",
		"wedding":    "events",
		"birthday":   "events",
		"party":      "events",
		"food":       "food",
		"recipe":     "food",
		"screenshot": "screenshots",
		"document":   "documents",
	}

	for keyword, category := range keywords {
		if strings.Contains(filename, keyword) {
			return category, nil
		}
	}

	return "general", nil
}

// categorizeByMetadata categorizes based on metadata only
func (p *ImageProcessor) categorizeByMetadata(metadata *ImageMetadata) string {
	// Categorize by camera type
	if metadata.Make != "" {
		make := strings.ToLower(metadata.Make)
		if strings.Contains(make, "canon") || strings.Contains(make, "nikon") || strings.Contains(make, "sony") {
			return "photography"
		}
		if strings.Contains(make, "apple") || strings.Contains(make, "samsung") {
			return "mobile"
		}
	}

	// Default category
	return "general"
}

// generateFileName generates a descriptive filename
func (p *ImageProcessor) generateFileName(sourcePath string, metadata *ImageMetadata, category string) string {
	ext := filepath.Ext(sourcePath)
	baseName := strings.TrimSuffix(filepath.Base(sourcePath), ext)

	// Remove timestamp prefix if present (from our upload naming)
	parts := strings.SplitN(baseName, "_", 2)
	if len(parts) == 2 {
		baseName = parts[1]
	}

	// Create descriptive name: category_datetime_originalname.ext
	timestamp := metadata.DateTime.Format("20060102_150405")

	// Clean the base name
	baseName = strings.ReplaceAll(baseName, " ", "_")
	baseName = strings.ToLower(baseName)

	newName := fmt.Sprintf("%s_%s_%s%s", category, timestamp, baseName, ext)

	return newName
}

// moveFile moves a file from source to destination, returning the path it
// actually wrote to (which may differ from destPath if a file already
// existed there).
func (p *ImageProcessor) moveFile(sourcePath, destPath string) (string, error) {
	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		// File exists, append counter
		ext := filepath.Ext(destPath)
		base := strings.TrimSuffix(destPath, ext)
		counter := 1
		for {
			newDest := fmt.Sprintf("%s_%d%s", base, counter, ext)
			if _, err := os.Stat(newDest); os.IsNotExist(err) {
				destPath = newDest
				break
			}
			counter++
		}
	}

	// Copy file
	source, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}

	destination, err := os.Create(destPath)
	if err != nil {
		source.Close()
		return "", err
	}

	_, err = io.Copy(destination, source)

	// Explicitly close both files before attempting to delete
	// This is crucial on Windows to release file handles
	source.Close()
	destination.Close()

	if err != nil {
		return "", err
	}

	// Add a tiny delay to ensure Windows releases the file handle
	time.Sleep(10 * time.Millisecond)

	// Delete source file
	if err := os.Remove(sourcePath); err != nil {
		return "", err
	}

	return destPath, nil
}

// updateDatabaseRecord updates the image record in the database so its
// storagePath reflects where processImage actually moved the file.
func (p *ImageProcessor) updateDatabaseRecord(oldPath, newPath, category string) error {
	if err := p.repo.UpdateImagePath(oldPath, newPath, category); err != nil {
		return err
	}
	slog.Info("Database record updated", "oldPath", oldPath, "newPath", newPath, "category", category)
	return nil
}

// isImageFile checks if a file extension is an image
func isImageFile(ext string) bool {
	imageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".heic": true,
		".heif": true,
	}
	return imageExts[ext]
}

// AICategorizationRequest is sent to AI service
type AICategorizationRequest struct {
	ImageBase64 string `json:"image"`
	Prompt      string `json:"prompt"`
}

// AICategorizationResponse is received from AI service
type AICategorizationResponse struct {
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}
