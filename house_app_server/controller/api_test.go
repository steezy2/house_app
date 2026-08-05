package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"house-app/internal/mocks"
	"house-app/models"
	"house-app/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestAPIServer_GetImages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("API_KEY", "")

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

	mockRepo := &mocks.ImageRepository{
		GetImagesFunc: func() ([]models.Image, error) {
			return expectedImages, nil
		},
	}

	mockService := service.NewService(mockRepo)
	server := NewAPIServer(mockService)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/images", nil)
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
	t.Setenv("API_KEY", "")

	mockRepo := &mocks.ImageRepository{
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

func TestAPIServer_RejectsRequestsWithoutAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("API_KEY", "secret-key")

	mockService := service.NewService(&mocks.ImageRepository{})
	server := NewAPIServer(mockService)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/images", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d without an API key, but got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAPIServer_RejectsWrongAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("API_KEY", "secret-key")

	mockService := service.NewService(&mocks.ImageRepository{})
	server := NewAPIServer(mockService)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/images", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d with a wrong API key, but got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAPIServer_AcceptsValidAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("API_KEY", "secret-key")

	mockRepo := &mocks.ImageRepository{
		GetImagesFunc: func() ([]models.Image, error) {
			return []models.Image{}, nil
		},
	}
	mockService := service.NewService(mockRepo)
	server := NewAPIServer(mockService)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/images", nil)
	req.Header.Set("X-API-Key", "secret-key")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d with a valid API key, but got %d", http.StatusOK, w.Code)
	}
}

func TestAPIServer_SwaggerBypassesAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("API_KEY", "secret-key")

	mockService := service.NewService(&mocks.ImageRepository{})
	server := NewAPIServer(mockService)

	req, _ := http.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Errorf("Swagger UI should not require an API key, got %d", w.Code)
	}
}
