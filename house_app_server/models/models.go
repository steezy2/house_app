package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SupportedImageExtensions are the image file extensions accepted for
// upload and eligible for background processing — the single source of
// truth shared by service.UploadImage/BulkUploadImages (which reject
// anything not in this set or SupportedVideoExtensions) and
// processor.isMediaFile (which skips anything else).
var SupportedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
	".heic": true,
	".heif": true,
}

// SupportedVideoExtensions are the video file extensions accepted
// alongside SupportedImageExtensions, covering the common formats phone
// cameras record in (iOS: .mov/.mp4, Android: .mp4/.3gp/.webm).
var SupportedVideoExtensions = map[string]bool{
	".mp4":  true,
	".mov":  true,
	".m4v":  true,
	".3gp":  true,
	".webm": true,
	".avi":  true,
}

// Image represents an image or video entry in the database. The name
// predates video support (see MIGRATION_SUMMARY.md/CLAUDE.md) — kept as-is
// to avoid an invasive rename of the collection, routes, and Swagger docs
// for no functional benefit.
type Image struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Filename     string             `bson:"filename" json:"filename"`
	OriginalPath string             `bson:"originalPath" json:"originalPath"`
	StoragePath  string             `bson:"storagePath" json:"storagePath"`
	Size         int64              `bson:"size" json:"size"`
	ContentType  string             `bson:"contentType" json:"contentType"`
	Tags         []string           `bson:"tags,omitempty" json:"tags,omitempty"`
	Category     string             `bson:"category,omitempty" json:"category,omitempty"`
	Metadata     interface{}        `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	ProcessedAt  time.Time          `bson:"processedAt,omitempty" json:"processedAt,omitempty"`
	// BackedUpTo lists the names of configured backup.Destinations that
	// have confirmed a copy of this file (see processor.processImage).
	// Partial credit is expected: if 2 of 3 configured destinations
	// succeeded, this has 2 entries. A future delete-from-phone feature
	// should check this covers every configured destination before
	// treating a file as safe to remove from the source device.
	BackedUpTo []string `bson:"backedUpTo,omitempty" json:"backedUpTo,omitempty"`
}

// BulkUploadRequest represents the request body for bulk image upload
type BulkUploadRequest struct {
	SourceFolder string   `json:"sourceFolder"`
	Tags         []string `json:"tags,omitempty"`
}

// BulkUploadResponse represents the response for bulk image upload
type BulkUploadResponse struct {
	TotalImages     int      `json:"totalImages"`
	SuccessfulCount int      `json:"successfulCount"`
	FailedCount     int      `json:"failedCount"`
	FailedFiles     []string `json:"failedFiles,omitempty"`
}
