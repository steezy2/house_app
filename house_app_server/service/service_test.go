package service

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"house-app/internal/mocks"
	"house-app/models"

	"go.mongodb.org/mongo-driver/mongo"
)

// newTestFileHeader builds a *multipart.FileHeader for the given filename
// and content, the way Gin would produce one from an incoming request.
func newTestFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("failed to write form file content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/", body)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err := req.ParseMultipartForm(32 << 20); err != nil {
		t.Fatalf("failed to parse multipart form: %v", err)
	}
	return req.MultipartForm.File["file"][0]
}

func TestService_CreateImage(t *testing.T) {
	mockRepo := &mocks.ImageRepository{
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

	mockRepo := &mocks.ImageRepository{
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

func TestParseTags(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"single tag", "vacation", []string{"vacation"}},
		{"comma separated", "vacation,beach,2025", []string{"vacation", "beach", "2025"}},
		{"trims whitespace", "vacation, beach , 2025", []string{"vacation", "beach", "2025"}},
		{"drops empty segments", "vacation,,beach,", []string{"vacation", "beach"}},
		{"only separators", ",, ,", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseTags(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseTags(%q) = %#v, want %#v", tc.input, got, tc.want)
			}
		})
	}
}

func TestService_UploadImage_RejectsUnsupportedExtension(t *testing.T) {
	t.Cleanup(func() { os.RemoveAll("./uploads") })

	svc := NewService(&mocks.ImageRepository{})
	file := newTestFileHeader(t, "malware.exe", []byte("not an image"))

	_, err := svc.UploadImage(file, nil)
	if err == nil {
		t.Fatal("expected an error for an unsupported file extension, got nil")
	}
}

func TestService_UploadImage_SanitizesFilename(t *testing.T) {
	t.Cleanup(func() { os.RemoveAll("./uploads") })

	var created models.Image
	mockRepo := &mocks.ImageRepository{
		CreateImageFunc: func(img models.Image) (*mongo.InsertOneResult, error) {
			created = img
			return &mongo.InsertOneResult{}, nil
		},
	}
	svc := NewService(mockRepo)

	// A filename with directory traversal components should be reduced to
	// its base name before being used to build the storage path.
	file := newTestFileHeader(t, "../../evil.jpg", []byte("image bytes"))

	image, err := svc.UploadImage(file, []string{"test"})
	if err != nil {
		t.Fatalf("UploadImage failed: %v", err)
	}

	if image.Filename != "evil.jpg" {
		t.Errorf("Filename = %q, want %q", image.Filename, "evil.jpg")
	}
	if created.Filename != "evil.jpg" {
		t.Errorf("stored record Filename = %q, want %q", created.Filename, "evil.jpg")
	}
	if filepath.Dir(image.StoragePath) != filepath.Clean("./uploads") {
		t.Errorf("StoragePath escaped uploads dir: %q", image.StoragePath)
	}
}
