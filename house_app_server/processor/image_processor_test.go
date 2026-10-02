package processor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"house-app/backup"
	"house-app/internal/mocks"
	"house-app/models"
)

// fakeDestination is a minimal backup.Destination test double that records
// every Copy call it receives and can be made to fail.
type fakeDestination struct {
	name    string
	failErr error
	calls   []fakeDestinationCall
}

type fakeDestinationCall struct {
	relativePath string
	localPath    string
}

func (d *fakeDestination) Name() string { return d.name }

func (d *fakeDestination) Copy(ctx context.Context, relativePath, localPath string) error {
	d.calls = append(d.calls, fakeDestinationCall{relativePath: relativePath, localPath: localPath})
	return d.failErr
}

var _ backup.Destination = (*fakeDestination)(nil)

func TestIsMediaFile(t *testing.T) {
	cases := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".heic": true,
		".heif": true,
		".mp4":  true,
		".mov":  true,
		".m4v":  true,
		".3gp":  true,
		".webm": true,
		".avi":  true,
		".txt":  false,
		".JPG":  false, // callers are expected to lowercase before calling
		"":      false,
	}
	for ext, want := range cases {
		if got := isMediaFile(ext); got != want {
			t.Errorf("isMediaFile(%q) = %v, want %v", ext, got, want)
		}
	}
}

func TestIsVideoFile(t *testing.T) {
	cases := map[string]bool{
		".mp4":  true,
		".mov":  true,
		".jpg":  false,
		".heic": false,
		"":      false,
	}
	for ext, want := range cases {
		if got := isVideoFile(ext); got != want {
			t.Errorf("isVideoFile(%q) = %v, want %v", ext, got, want)
		}
	}
}

func TestProcessImage_VideoDefaultsToVideoCategory(t *testing.T) {
	tmp := t.TempDir()
	uploadDir := filepath.Join(tmp, "uploads")
	storageDir := filepath.Join(tmp, "storage")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}

	src := filepath.Join(uploadDir, "1700000000_clip.mp4")
	if err := os.WriteFile(src, []byte("not a real video, just bytes"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	p := &ImageProcessor{
		uploadDir:      uploadDir,
		storageBaseDir: storageDir,
		repo:           &mocks.ImageRepository{},
	}

	result := p.processImage(src)
	if result.Error != nil {
		t.Fatalf("processImage returned error: %v", result.Error)
	}
	if result.Category != "video" {
		t.Errorf("Category = %q, want %q", result.Category, "video")
	}
	if _, err := os.Stat(result.NewPath); err != nil {
		t.Errorf("expected file at %q, stat err = %v", result.NewPath, err)
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

func TestBackupFile_RecordsOnlySucceededDestinations(t *testing.T) {
	storageDir := t.TempDir()
	fileDir := filepath.Join(storageDir, "2026", "08", "travel")
	if err := os.MkdirAll(fileDir, 0o755); err != nil {
		t.Fatalf("failed to create storage subdirectory: %v", err)
	}
	filePath := filepath.Join(fileDir, "photo.jpg")
	if err := os.WriteFile(filePath, []byte("bytes"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	okDest := &fakeDestination{name: "D:/backup1"}
	failDest := &fakeDestination{name: "s3:bucket", failErr: errors.New("network error")}

	var gotPath string
	var gotBackedUpTo []string
	repo := &mocks.ImageRepository{
		UpdateBackupStatusFunc: func(storagePath string, backedUpTo []string) error {
			gotPath = storagePath
			gotBackedUpTo = backedUpTo
			return nil
		},
	}

	p := &ImageProcessor{
		storageBaseDir: storageDir,
		repo:           repo,
		destinations:   []backup.Destination{okDest, failDest},
	}

	p.backupFile(filePath)

	if gotPath != filePath {
		t.Errorf("UpdateBackupStatus storagePath = %q, want %q", gotPath, filePath)
	}
	if !reflect.DeepEqual(gotBackedUpTo, []string{"D:/backup1"}) {
		t.Errorf("backedUpTo = %v, want [D:/backup1]", gotBackedUpTo)
	}

	wantRelative := filepath.Join("2026", "08", "travel", "photo.jpg")
	if len(okDest.calls) != 1 || okDest.calls[0].relativePath != wantRelative {
		t.Errorf("okDest.calls = %+v, want one call with relativePath %q", okDest.calls, wantRelative)
	}
	if len(failDest.calls) != 1 {
		t.Errorf("failDest.calls = %+v, want exactly one call despite failing", failDest.calls)
	}
}

func TestBackupFile_NoDestinationsConfiguredSkipsUpdate(t *testing.T) {
	called := false
	repo := &mocks.ImageRepository{
		UpdateBackupStatusFunc: func(storagePath string, backedUpTo []string) error {
			called = true
			return nil
		},
	}
	p := &ImageProcessor{repo: repo}

	p.backupFile("some/path.jpg")

	if called {
		t.Error("expected UpdateBackupStatus not to be called when no destinations are configured")
	}
}

// writeStoredFile creates a file under storageDir at the given relative
// path and returns its full path.
func writeStoredFile(t *testing.T, storageDir, relativePath string) string {
	t.Helper()
	path := filepath.Join(storageDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create storage subdirectory: %v", err)
	}
	if err := os.WriteFile(path, []byte("bytes"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	return path
}

func TestRetryMissingBackups_CopiesOnlyToMissingDestinations(t *testing.T) {
	storageDir := t.TempDir()
	path := writeStoredFile(t, storageDir, filepath.Join("2026", "08", "travel", "photo.jpg"))

	haveDest := &fakeDestination{name: "D:/backup1"}
	missingDest := &fakeDestination{name: "s3:bucket"}

	var added []string
	repo := &mocks.ImageRepository{
		FindImagesMissingBackupFunc: func(destination string, limit int64) ([]models.Image, error) {
			if destination == "s3:bucket" {
				return []models.Image{{StoragePath: path, BackedUpTo: []string{"D:/backup1"}}}, nil
			}
			return nil, nil
		},
		AddBackupDestinationFunc: func(storagePath, destination string) error {
			added = append(added, storagePath+"|"+destination)
			return nil
		},
	}
	p := &ImageProcessor{storageBaseDir: storageDir, repo: repo, destinations: []backup.Destination{haveDest, missingDest}}

	p.RetryMissingBackups()

	if len(haveDest.calls) != 0 {
		t.Errorf("haveDest.calls = %+v, want none", haveDest.calls)
	}
	wantRelative := filepath.Join("2026", "08", "travel", "photo.jpg")
	if len(missingDest.calls) != 1 || missingDest.calls[0].relativePath != wantRelative {
		t.Errorf("missingDest.calls = %+v, want one call with relativePath %q", missingDest.calls, wantRelative)
	}
	if !reflect.DeepEqual(added, []string{path + "|s3:bucket"}) {
		t.Errorf("AddBackupDestination calls = %v, want [%s|s3:bucket]", added, path)
	}
}

func TestRetryMissingBackups_StopsDestinationAfterRepeatedFailures(t *testing.T) {
	storageDir := t.TempDir()
	var images []models.Image
	for _, name := range []string{"a.jpg", "b.jpg", "c.jpg", "d.jpg", "e.jpg"} {
		images = append(images, models.Image{StoragePath: writeStoredFile(t, storageDir, name)})
	}

	downDest := &fakeDestination{name: "E:/offline", failErr: errors.New("drive not mounted")}
	addCalled := false
	repo := &mocks.ImageRepository{
		FindImagesMissingBackupFunc: func(string, int64) ([]models.Image, error) { return images, nil },
		AddBackupDestinationFunc: func(string, string) error {
			addCalled = true
			return nil
		},
	}
	p := &ImageProcessor{storageBaseDir: storageDir, repo: repo, destinations: []backup.Destination{downDest}}

	p.RetryMissingBackups()

	if len(downDest.calls) != maxConsecutiveBackupFailures {
		t.Errorf("downDest got %d Copy calls, want %d before giving up for this cycle", len(downDest.calls), maxConsecutiveBackupFailures)
	}
	if addCalled {
		t.Error("expected no AddBackupDestination call when every copy failed")
	}
}

func TestRetryMissingBackups_SkipsFilesMissingFromStorage(t *testing.T) {
	storageDir := t.TempDir()
	var images []models.Image
	for i := 0; i < maxConsecutiveBackupFailures; i++ {
		images = append(images, models.Image{StoragePath: filepath.Join(storageDir, "gone", string(rune('a'+i))+".jpg")})
	}
	present := writeStoredFile(t, storageDir, "present.jpg")
	images = append(images, models.Image{StoragePath: present})

	dest := &fakeDestination{name: "D:/backup1"}
	var added []string
	repo := &mocks.ImageRepository{
		FindImagesMissingBackupFunc: func(string, int64) ([]models.Image, error) { return images, nil },
		AddBackupDestinationFunc: func(storagePath, destination string) error {
			added = append(added, storagePath)
			return nil
		},
	}
	p := &ImageProcessor{storageBaseDir: storageDir, repo: repo, destinations: []backup.Destination{dest}}

	p.RetryMissingBackups()

	if len(dest.calls) != 1 {
		t.Errorf("dest got %d Copy calls, want 1 (missing files skipped, not counted as destination failures)", len(dest.calls))
	}
	if !reflect.DeepEqual(added, []string{present}) {
		t.Errorf("AddBackupDestination calls = %v, want [%s]", added, present)
	}
}
