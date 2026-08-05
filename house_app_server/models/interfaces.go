package models

import "go.mongodb.org/mongo-driver/mongo"

// ImageRepository defines the interface for image persistence.
type ImageRepository interface {
	CreateImage(Image) (*mongo.InsertOneResult, error)
	GetImages() ([]Image, error)
	SearchImages(string) ([]Image, error)
	CreateImages([]Image) (*mongo.InsertManyResult, error)
	// UpdateImagePath updates the storagePath, category, and processedAt of
	// the image record whose storagePath currently equals oldStoragePath.
	UpdateImagePath(oldStoragePath, newStoragePath, category string) error
}
