package processor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"house-app/internal/mocks"
)

func TestIsImageFile(t *testing.T) {
	cases := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".heic": true,
		".heif": true,
		".txt":  false,
		".JPG":  false, // callers are expected to lowercase before calling
		"":      false,
	}
	for ext, want := range cases {
		if got := isImageFile(ext); got != want {
			t.Errorf("isImageFile(%q) = %v, want %v", ext, got, want)
		}
	}
}

func TestCategorizeByMetadata(t *testing.T) {
	cases := []struct {
		name string
		make string
		want string
	}{
		{"canon camera", "Canon", "photography"},
		{"nikon camera", "NIKON CORPORATION", "photography"},
		{"sony camera", "SONY", "photography"},
		{"apple phone", "Apple", "mobile"},
		{"samsung phone", "samsung", "mobile"},
		{"unknown make", "Fujifilm", "general"},
		{"no make", "", "general"},
	}
	p := &ImageProcessor{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := p.categorizeByMetadata(&ImageMetadata{Make: tc.make})
			if got != tc.want {
				t.Errorf("categorizeByMetadata(Make=%q) = %q, want %q", tc.make, got, tc.want)
			}
		})
	}
}

func TestCallOpenAIVisionKeywordFallback(t *testing.T) {
	cases := map[string]string{
		"beach_sunset.jpg":     "travel",
		"mountain_hike.jpg":    "nature",
		"family_reunion.jpg":   "family",
		"wedding_photos.jpg":   "events",
		"recipe_pasta.jpg":     "food",
		"screenshot_2025.png":  "screenshots",
		"scanned_document.png": "documents",
		"random_name123.jpg":   "general",
	}
	p := &ImageProcessor{}
	for filename, want := range cases {
		got, err := p.callOpenAIVision(filepath.Join("uploads", filename))
		if err != nil {
			t.Fatalf("callOpenAIVision(%q) returned error: %v", filename, err)
		}
		if got != want {
			t.Errorf("callOpenAIVision(%q) = %q, want %q", filename, got, want)
		}
	}
}

func TestCategorizeWithAI_FallsBackWithoutAPIKey(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	p := &ImageProcessor{aiEnabled: true}
	metadata := &ImageMetadata{Make: "Canon"}

	got, err := p.categorizeWithAI("uploads/photo.jpg", metadata)
	if err != nil {
		t.Fatalf("categorizeWithAI returned error: %v", err)
	}
	if want := p.categorizeByMetadata(metadata); got != want {
		t.Errorf("categorizeWithAI without API key = %q, want metadata-based %q", got, want)
	}
}

func TestGenerateFileName(t *testing.T) {
	p := &ImageProcessor{}
	dt := time.Date(2025, 10, 26, 14, 30, 5, 0, time.UTC)
	metadata := &ImageMetadata{DateTime: dt}

	got := p.generateFileName("uploads/1730000000_Beach Sunset.JPG", metadata, "travel")
	want := "travel_20251026_143005_beach_sunset.JPG"
	if got != want {
		t.Errorf("generateFileName() = %q, want %q", got, want)
	}
}

func TestMoveFile(t *testing.T) {
	tmp := t.TempDir()
	p := &ImageProcessor{}

	src := filepath.Join(tmp, "source.jpg")
	if err := os.WriteFile(src, []byte("first"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}
	dest := filepath.Join(tmp, "dest.jpg")

	finalPath, err := p.moveFile(src, dest)
	if err != nil {
		t.Fatalf("moveFile returned error: %v", err)
	}
	if finalPath != dest {
		t.Errorf("finalPath = %q, want %q", finalPath, dest)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("expected source file to be removed, stat err = %v", err)
	}
	if data, err := os.ReadFile(dest); err != nil || string(data) != "first" {
		t.Errorf("dest file contents = %q, err = %v", data, err)
	}

	// A second move to the same destination must not clobber the existing
	// file — moveFile should rename to dest_1.jpg and report that path.
	src2 := filepath.Join(tmp, "source2.jpg")
	if err := os.WriteFile(src2, []byte("second"), 0o644); err != nil {
		t.Fatalf("failed to write second source file: %v", err)
	}
	finalPath2, err := p.moveFile(src2, dest)
	if err != nil {
		t.Fatalf("moveFile (collision) returned error: %v", err)
	}
	wantCollisionPath := filepath.Join(tmp, "dest_1.jpg")
	if finalPath2 != wantCollisionPath {
		t.Errorf("finalPath2 = %q, want %q", finalPath2, wantCollisionPath)
	}
	if data, err := os.ReadFile(dest); err != nil || string(data) != "first" {
		t.Errorf("original dest file should be untouched, got %q, err = %v", data, err)
	}
	if data, err := os.ReadFile(wantCollisionPath); err != nil || string(data) != "second" {
		t.Errorf("collision file contents = %q, err = %v", data, err)
	}
}

func TestUpdateDatabaseRecord(t *testing.T) {
	var gotOld, gotNew, gotCategory string
	repo := &mocks.ImageRepository{
		UpdateImagePathFunc: func(oldPath, newPath, category string) error {
			gotOld, gotNew, gotCategory = oldPath, newPath, category
			return nil
		},
	}
	p := &ImageProcessor{repo: repo}

	if err := p.updateDatabaseRecord("uploads/a.jpg", "storage/2025/10/travel/a.jpg", "travel"); err != nil {
		t.Fatalf("updateDatabaseRecord returned error: %v", err)
	}
	if gotOld != "uploads/a.jpg" || gotNew != "storage/2025/10/travel/a.jpg" || gotCategory != "travel" {
		t.Errorf("repo.UpdateImagePath called with (%q, %q, %q)", gotOld, gotNew, gotCategory)
	}
}

func TestUpdateDatabaseRecord_PropagatesError(t *testing.T) {
	repo := &mocks.ImageRepository{
		UpdateImagePathFunc: func(oldPath, newPath, category string) error {
			return os.ErrInvalid
		},
	}
	p := &ImageProcessor{repo: repo}

	if err := p.updateDatabaseRecord("old", "new", "cat"); err == nil {
		t.Error("expected error to propagate from repo.UpdateImagePath, got nil")
	}
}
