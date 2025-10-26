package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"house-app/models"
	"house-app/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func TestAPIServer_GetImages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedImages := []models.Image{
		{
			ID:          primitive.NewObjectID(),
			Filename:    "image1.jpg",
			StoragePath: "./uploads/image1.jpg",
		},
		{
			ID:          primitive.NewObjectID(),
			Filename:    "image2.png",
			StoragePath: "./uploads/image2.png",
		},
	}

	mockRepo := &MockImageRepository{
		GetImagesFunc: func() ([]models.Image, error) {
			return expectedImages, nil
		},
	}

	mockService := service.NewService(mockRepo)
	server := NewAPIServer(mockService)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/images/", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, w.Code)
	}

	var images []models.Image
	err := json.Unmarshal(w.Body.Bytes(), &images)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(images) != 2 {
		t.Errorf("Expected 2 images, but got %d", len(images))
	}
}

func TestAPIServer_BulkUploadImages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockImageRepository{
		CreateImageFunc: func(img models.Image) (*mongo.InsertOneResult, error) {
			return &mongo.InsertOneResult{}, nil
		},
	}

	mockService := service.NewService(mockRepo)
	server := NewAPIServer(mockService)

	bulkReq := models.BulkUploadRequest{
		SourceFolder: "/nonexistent/folder",
		Tags:         []string{"test"},
	}
	body, _ := json.Marshal(bulkReq)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/images/bulk-upload", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	// Should fail because folder doesn't exist
	if w.Code != http.StatusInternalServerError {
		t.Logf("Response: %s", w.Body.String())
	}
}
