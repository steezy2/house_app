package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Image represents an image entry in the database
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
