package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalDestination mirrors files into another local or network-mounted
// directory — a second internal drive, an external USB drive, a NAS
// share, etc. — preserving the same relative path structure as
// STORAGE_DIR. This is what makes "save to 2 or 3 drives" work without
// any cloud account: add one LocalDestination per drive.
type LocalDestination struct {
	baseDir string
}

// NewLocalDestination creates a LocalDestination rooted at baseDir.
// baseDir itself is used as Name(), so log lines and BackedUpTo entries
// show exactly which drive/path a copy landed on.
func NewLocalDestination(baseDir string) *LocalDestination {
	return &LocalDestination{baseDir: baseDir}
}

func (d *LocalDestination) Name() string {
	return d.baseDir
}

func (d *LocalDestination) Copy(ctx context.Context, relativePath, localPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	destPath := filepath.Join(d.baseDir, relativePath)

	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	src, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy file to backup destination: %w", err)
	}

	return nil
}

var _ Destination = (*LocalDestination)(nil)
