package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalDestination_Name(t *testing.T) {
	d := NewLocalDestination(`D:\backup1`)
	if got := d.Name(); got != `D:\backup1` {
		t.Errorf("Name() = %q, want %q", got, `D:\backup1`)
	}
}

func TestLocalDestination_Copy(t *testing.T) {
	srcDir := t.TempDir()
	destBase := t.TempDir()

	srcPath := filepath.Join(srcDir, "photo.jpg")
	if err := os.WriteFile(srcPath, []byte("image bytes"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	d := NewLocalDestination(destBase)
	relativePath := filepath.Join("2026", "08", "travel", "photo.jpg")

	if err := d.Copy(context.Background(), relativePath, srcPath); err != nil {
		t.Fatalf("Copy returned error: %v", err)
	}

	destPath := filepath.Join(destBase, relativePath)
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("expected copied file at %q: %v", destPath, err)
	}
	if string(data) != "image bytes" {
		t.Errorf("copied file contents = %q, want %q", data, "image bytes")
	}

	// The source file must be untouched — Copy backs up, it doesn't move.
	if _, err := os.Stat(srcPath); err != nil {
		t.Errorf("expected source file to still exist: %v", err)
	}
}

func TestLocalDestination_Copy_CreatesNestedDirectories(t *testing.T) {
	srcDir := t.TempDir()
	destBase := t.TempDir()

	srcPath := filepath.Join(srcDir, "clip.mp4")
	if err := os.WriteFile(srcPath, []byte("video bytes"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	d := NewLocalDestination(destBase)
	// destBase doesn't have a "2026/08/video" subtree yet — Copy must
	// create it.
	relativePath := filepath.Join("2026", "08", "video", "clip.mp4")

	if err := d.Copy(context.Background(), relativePath, srcPath); err != nil {
		t.Fatalf("Copy returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destBase, relativePath)); err != nil {
		t.Errorf("expected nested destination file to exist: %v", err)
	}
}

func TestLocalDestination_Copy_RespectsCanceledContext(t *testing.T) {
	srcDir := t.TempDir()
	destBase := t.TempDir()

	srcPath := filepath.Join(srcDir, "photo.jpg")
	if err := os.WriteFile(srcPath, []byte("image bytes"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	d := NewLocalDestination(destBase)
	if err := d.Copy(ctx, "photo.jpg", srcPath); err == nil {
		t.Error("expected Copy to return an error for a canceled context")
	}
}

func TestLocalDestination_Copy_MissingSourceFile(t *testing.T) {
	destBase := t.TempDir()
	d := NewLocalDestination(destBase)

	err := d.Copy(context.Background(), "photo.jpg", filepath.Join(t.TempDir(), "does-not-exist.jpg"))
	if err == nil {
		t.Error("expected an error when the source file doesn't exist")
	}
}

func TestLocalDestination_Copy_FailedCopyLeavesNoPartialFile(t *testing.T) {
	destBase := t.TempDir()
	d := NewLocalDestination(destBase)

	// A directory opens fine but fails on read, standing in for a source
	// or drive that errors partway through the copy.
	err := d.Copy(context.Background(), "photo.jpg", t.TempDir())
	if err == nil {
		t.Fatal("expected an error when the source can't be read")
	}

	entries, err := os.ReadDir(destBase)
	if err != nil {
		t.Fatalf("failed to read destination dir: %v", err)
	}
	for _, e := range entries {
		t.Errorf("failed copy left %q behind in the destination", e.Name())
	}
}
