package service

import (
	"testing"

	"house-app/models"

	"go.mongodb.org/mongo-driver/mongo"
)

// MockImageRepository is a mock of ImageRepository
type MockImageRepository struct {
	CreateImageFunc  func(models.Image) (*mongo.InsertOneResult, error)
	GetImagesFunc    func() ([]models.Image, error)
	SearchImagesFunc func(string) ([]models.Image, error)
	CreateImagesFunc func([]models.Image) (*mongo.InsertManyResult, error)
}

func (m *MockImageRepository) CreateImage(img models.Image) (*mongo.InsertOneResult, error) {
	return m.CreateImageFunc(img)
}

func (m *MockImageRepository) GetImages() ([]models.Image, error) {
	return m.GetImagesFunc()
}

func (m *MockImageRepository) SearchImages(q string) ([]models.Image, error) {
	return m.SearchImagesFunc(q)
}

func (m *MockImageRepository) CreateImages(imgs []models.Image) (*mongo.InsertManyResult, error) {
	return m.CreateImagesFunc(imgs)
}

func TestService_CreateImage(t *testing.T) {
	mockRepo := &MockImageRepository{
		CreateImageFunc: func(img models.Image) (*mongo.InsertOneResult, error) {
			return &mongo.InsertOneResult{}, nil
		},
	}

	service := NewService(mockRepo)

	image := models.Image{
		Filename:    "test.jpg",
		StoragePath: "./uploads/test.jpg",
		Size:        1024,
		ContentType: "image/jpeg",
	}

	createdImage, err := service.CreateImage(image)
	if err != nil {
		t.Fatalf("CreateImage failed: %v", err)
	}

	if createdImage.Filename != "test.jpg" {
		t.Errorf("Expected Filename to be 'test.jpg', but got '%s'", createdImage.Filename)
	}
}

func TestService_GetImages(t *testing.T) {
	expectedImages := []models.Image{
		{Filename: "image1.jpg"},
		{Filename: "image2.png"},
	}

	mockRepo := &MockImageRepository{
		GetImagesFunc: func() ([]models.Image, error) {
			return expectedImages, nil
		},
	}

	service := NewService(mockRepo)

	images, err := service.GetImages()
	if err != nil {
		t.Fatalf("GetImages failed: %v", err)
	}

	if len(images) != 2 {
		t.Errorf("Expected 2 images, but got %d", len(images))
	}
}
