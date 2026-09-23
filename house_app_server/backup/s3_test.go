package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// These tests cover S3Destination's non-network logic — construction and
// name formatting. Verifying an actual PutObject call needs a real (or
// mocked) S3-compatible endpoint, which isn't available in this
// environment; that's a live/manual check once real credentials exist.

func TestNewS3Destination(t *testing.T) {
	dest, err := NewS3Destination(context.Background(), S3Config{
		Endpoint:  "https://s3.us-west-002.backblazeb2.com",
		Bucket:    "house-app-backup",
		AccessKey: "test-key",
		SecretKey: "test-secret",
		Region:    "us-west-002",
	})
	if err != nil {
		t.Fatalf("NewS3Destination returned error: %v", err)
	}
	if got, want := dest.Name(), "s3:house-app-backup"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestS3Destination_Copy_RespectsCanceledContext(t *testing.T) {
	dest, err := NewS3Destination(context.Background(), S3Config{
		Bucket:    "house-app-backup",
		AccessKey: "test-key",
		SecretKey: "test-secret",
		Region:    "us-east-1",
	})
	if err != nil {
		t.Fatalf("NewS3Destination returned error: %v", err)
	}

	srcPath := filepath.Join(t.TempDir(), "photo.jpg")
	if err := os.WriteFile(srcPath, []byte("image bytes"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := dest.Copy(ctx, "photo.jpg", srcPath); err == nil {
		t.Error("expected Copy to return an error for a canceled context")
	}
}

func TestS3Destination_Copy_MissingSourceFile(t *testing.T) {
	dest, err := NewS3Destination(context.Background(), S3Config{
		Bucket:    "house-app-backup",
		AccessKey: "test-key",
		SecretKey: "test-secret",
		Region:    "us-east-1",
	})
	if err != nil {
		t.Fatalf("NewS3Destination returned error: %v", err)
	}

	err = dest.Copy(context.Background(), "photo.jpg", filepath.Join(t.TempDir(), "does-not-exist.jpg"))
	if err == nil {
		t.Error("expected an error when the source file doesn't exist")
	}
}
