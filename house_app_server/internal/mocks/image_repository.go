// Package mocks provides shared, nil-safe test doubles for house-app's
// repository interfaces, so controller/service/processor tests don't each
// hand-roll their own copy of the same mock.
package mocks

import (
	"house-app/models"

	"go.mongodb.org/mongo-driver/mongo"
)

// ImageRepository is a configurable models.ImageRepository test double. Any
// Func field left unset returns a zero value instead of panicking, so a
// test only needs to configure the methods it actually exercises.
type ImageRepository struct {
	CreateImageFunc     func(models.Image) (*mongo.InsertOneResult, error)
	GetImagesFunc       func() ([]models.Image, error)
	SearchImagesFunc    func(string) ([]models.Image, error)
	CreateImagesFunc    func([]models.Image) (*mongo.InsertManyResult, error)
	UpdateImagePathFunc func(oldStoragePath, newStoragePath, category string) error
}

func (m *ImageRepository) CreateImage(img models.Image) (*mongo.InsertOneResult, error) {
	if m.CreateImageFunc == nil {
		return &mongo.InsertOneResult{}, nil
	}
	return m.CreateImageFunc(img)
}

func (m *ImageRepository) GetImages() ([]models.Image, error) {
	if m.GetImagesFunc == nil {
		return nil, nil
	}
	return m.GetImagesFunc()
}

func (m *ImageRepository) SearchImages(query string) ([]models.Image, error) {
	if m.SearchImagesFunc == nil {
		return nil, nil
	}
	return m.SearchImagesFunc(query)
}

func (m *ImageRepository) CreateImages(imgs []models.Image) (*mongo.InsertManyResult, error) {
	if m.CreateImagesFunc == nil {
		return nil, nil
	}
	return m.CreateImagesFunc(imgs)
}

func (m *ImageRepository) UpdateImagePath(oldStoragePath, newStoragePath, category string) error {
	if m.UpdateImagePathFunc == nil {
		return nil
	}
	return m.UpdateImagePathFunc(oldStoragePath, newStoragePath, category)
}

var _ models.ImageRepository = (*ImageRepository)(nil)
