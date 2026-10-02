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

	"house-app/backup"
	"house-app/models"

	"github.com/rwcarlsen/goexif/exif"
)

// ImageProcessor handles automatic organization of uploaded images
type ImageProcessor struct {
	uploadDir      string
	storageBaseDir string
	repo           models.ImageRepository
	aiEnabled      bool
	destinations   []backup.Destination
}

// NewImageProcessor creates a new image processor. destinations may be
// empty (no backup configured) — see main.go for how it's built from
// BACKUP_LOCAL_DIRS/BACKUP_S3_* env vars.
func NewImageProcessor(uploadDir, storageBaseDir string, repo models.ImageRepository, destinations []backup.Destination) *ImageProcessor {
	return &ImageProcessor{
		uploadDir:      uploadDir,
		storageBaseDir: storageBaseDir,
		repo:           repo,
		aiEnabled:      os.Getenv("ENABLE_AI_CATEGORIZATION") == "true",
		destinations:   destinations,
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
	p.runCycle()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping periodic image processing")
			return
		case <-ticker.C:
			p.runCycle()
		}
	}
}

// runCycle processes new uploads first, then retries missing backups, so a
// backup backlog never delays organizing new files.
func (p *ImageProcessor) runCycle() {
	p.ProcessAllImages()
	p.RetryMissingBackups()
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

		// Check if it's a supported image or video file
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if !isMediaFile(ext) {
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

	// Videos default to their own "video" bucket rather than falling into
	// the same generic category as uncategorized photos, unless something
	// more specific (AI or keyword matching) already applied.
	if (category == "general" || category == "uncategorized") && isVideoFile(strings.ToLower(filepath.Ext(sourcePath))) {
		category = "video"
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

	p.backupFile(destPath)

	return result
}

// backupCopyTimeout bounds how long a single destination's Copy can take,
// so one slow or unreachable drive/cloud endpoint can't stall the whole
// processing cycle indefinitely.
const backupCopyTimeout = 5 * time.Minute

// backupFile mirrors destPath (its current location under storageBaseDir)
// to every configured backup.Destination, recording in the database which
// ones succeeded (models.Image.BackedUpTo). A destination failing doesn't
// affect the others or fail processing overall — the primary copy in
// STORAGE_DIR already succeeded by the time this runs.
func (p *ImageProcessor) backupFile(destPath string) {
	if len(p.destinations) == 0 {
		return
	}

	var succeeded []string
	for _, dest := range p.destinations {
		if err := p.copyToDestination(dest, destPath); err != nil {
			slog.Warn("Backup destination failed", "destination", dest.Name(), "file", destPath, "error", err)
			continue
		}
		slog.Info("Backed up file", "destination", dest.Name(), "file", destPath)
		succeeded = append(succeeded, dest.Name())
	}

	if err := p.repo.UpdateBackupStatus(destPath, succeeded); err != nil {
		slog.Warn("Failed to update backup status", "error", err, "file", destPath)
	}
}

// copyToDestination copies storagePath (a file under storageBaseDir) to
// dest under the same relative path, bounded by backupCopyTimeout.
func (p *ImageProcessor) copyToDestination(dest backup.Destination, storagePath string) error {
	relativePath, err := filepath.Rel(p.storageBaseDir, storagePath)
	if err != nil || !filepath.IsLocal(relativePath) {
		return fmt.Errorf("file %s is not under storage dir %s", storagePath, p.storageBaseDir)
	}

	ctx, cancel := context.WithTimeout(context.Background(), backupCopyTimeout)
	defer cancel()
	return dest.Copy(ctx, relativePath, storagePath)
}

// backupRetryBatchSize caps how many files RetryMissingBackups tries per
// destination per cycle, so a large backfill (e.g. a newly added drive)
// spreads over several cycles instead of delaying new uploads for hours.
const backupRetryBatchSize = 200

// maxConsecutiveBackupFailures is how many copies in a row may fail before
// RetryMissingBackups gives up on a destination until the next cycle: it's
// probably offline, and each attempt could take up to backupCopyTimeout.
const maxConsecutiveBackupFailures = 3

// RetryMissingBackups copies processed files to every configured
// destination their BackedUpTo doesn't list yet: copies that failed at
// processing time, and files stored before a destination was added.
// Files no longer in STORAGE_DIR are skipped without counting against the
// destination.
func (p *ImageProcessor) RetryMissingBackups() {
	for _, dest := range p.destinations {
		images, err := p.repo.FindImagesMissingBackup(dest.Name(), backupRetryBatchSize)
		if err != nil {
			slog.Error("Failed to find images missing backup", "destination", dest.Name(), "error", err)
			continue
		}

		failures := 0
		for _, img := range images {
			if _, err := os.Stat(img.StoragePath); err != nil {
				slog.Warn("Skipping backup retry for missing file", "file", img.StoragePath, "error", err)
				continue
			}
			if err := p.copyToDestination(dest, img.StoragePath); err != nil {
				slog.Warn("Backup retry failed", "destination", dest.Name(), "file", img.StoragePath, "error", err)
				if failures++; failures >= maxConsecutiveBackupFailures {
					slog.Warn("Pausing backup retries for this cycle", "destination", dest.Name())
					break
				}
				continue
			}
			failures = 0
			if err := p.repo.AddBackupDestination(img.StoragePath, dest.Name()); err != nil {
				slog.Warn("Failed to record backup", "destination", dest.Name(), "file", img.StoragePath, "error", err)
			}
		}
	}
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

	// Video files don't carry EXIF data — skip straight to file-system
	// defaults instead of attempting (and always failing) an EXIF decode.
	if isVideoFile(strings.ToLower(filepath.Ext(imagePath))) {
		return metadata, nil
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

// keywordCategories drives callOpenAIVision's fallback categorization. It's
// an ordered slice, not a map: Go randomizes map iteration order, and a
// filename can match more than one keyword (e.g. "beach_sunset.jpg" matches
// both "beach" and "sunset") — this makes the first match in list order
// win, deterministically, instead of picking a different category each run.
var keywordCategories = []struct {
	keyword  string
	category string
}{
	{"vacation", "travel"},
	{"holiday", "travel"},
	{"trip", "travel"},
	{"beach", "travel"},
	{"mountain", "nature"},
	{"landscape", "nature"},
	{"sunset", "nature"},
	{"sunrise", "nature"},
	{"portrait", "people"},
	{"family", "family"},
	{"wedding", "events"},
	{"birthday", "events"},
	{"party", "events"},
	{"food", "food"},
	{"recipe", "food"},
	{"screenshot", "screenshots"},
	{"document", "documents"},
}

// callOpenAIVision calls OpenAI Vision API to categorize the image
func (p *ImageProcessor) callOpenAIVision(imagePath string) (string, error) {
	// Placeholder for OpenAI Vision API integration
	// This would encode the image as base64 and send to OpenAI API
	// For now, return a smart default based on filename patterns

	filename := strings.ToLower(filepath.Base(imagePath))

	for _, kc := range keywordCategories {
		if strings.Contains(filename, kc.keyword) {
			return kc.category, nil
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

// isMediaFile checks if a file extension is a supported image or video
// type. ext is expected to already be lowercased (see ProcessAllImages).
func isMediaFile(ext string) bool {
	return models.SupportedImageExtensions[ext] || models.SupportedVideoExtensions[ext]
}

// isVideoFile checks if a file extension is a supported video type. ext is
// expected to already be lowercased.
func isVideoFile(ext string) bool {
	return models.SupportedVideoExtensions[ext]
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
