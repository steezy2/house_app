package models

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestImage(t *testing.T) {
	image := Image{
		ID:          primitive.NewObjectID(),
		Filename:    "test.jpg",
		StoragePath: "./uploads/test.jpg",
		Size:        1024,
		ContentType: "image/jpeg",
		CreatedAt:   time.Now(),
	}

	if image.Filename != "test.jpg" {
		t.Errorf("Expected Filename to be 'test.jpg', but got %s", image.Filename)
	}

	if image.ContentType != "image/jpeg" {
		t.Errorf("Expected ContentType to be 'image/jpeg', but got %s", image.ContentType)
	}
}

func TestBulkUploadRequest(t *testing.T) {
	req := BulkUploadRequest{
		SourceFolder: "/path/to/images",
		Tags:         []string{"test", "upload"},
	}

	if req.SourceFolder != "/path/to/images" {
		t.Errorf("Expected SourceFolder to be '/path/to/images', but got %s", req.SourceFolder)
	}

	if len(req.Tags) != 2 {
		t.Errorf("Expected 2 tags, but got %d", len(req.Tags))
	}
}
