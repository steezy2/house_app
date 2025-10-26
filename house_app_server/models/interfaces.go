package models

import "go.mongodb.org/mongo-driver/mongo"

// ImageRepository defines the interface for image persistence.
type ImageRepository interface {
	CreateImage(Image) (*mongo.InsertOneResult, error)
	GetImages() ([]Image, error)
	SearchImages(string) ([]Image, error)
	CreateImages([]Image) (*mongo.InsertManyResult, error)
}
