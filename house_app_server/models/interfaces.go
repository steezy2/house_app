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
	// UpdateBackupStatus records which backup.Destination names have
	// confirmed a copy of the image record whose storagePath currently
	// equals storagePath (see backup.Destination and
	// processor.processImage).
	UpdateBackupStatus(storagePath string, backedUpTo []string) error
	// FindImagesMissingBackup returns up to limit processed image records
	// whose BackedUpTo doesn't include destination.
	FindImagesMissingBackup(destination string, limit int64) ([]Image, error)
	// AddBackupDestination adds destination to BackedUpTo of the image
	// record whose storagePath currently equals storagePath.
	AddBackupDestination(storagePath, destination string) error
}
