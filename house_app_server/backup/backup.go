// Package backup mirrors files the processor has organized into
// STORAGE_DIR to one or more secondary locations — local/network drives
// and/or S3-compatible cloud storage — so a single disk failure doesn't
// mean permanent loss of the only copy. See models.Image.BackedUpTo for
// how per-file backup status is tracked.
package backup

import "context"

// Destination is a place processor.processImage copies a file to, in
// addition to its primary copy in STORAGE_DIR. The processor calls Copy
// for each configured destination one at a time, so implementations don't
// need to be concurrency-safe beyond ordinary sequential use.
type Destination interface {
	// Name identifies this destination for logging and for
	// models.Image.BackedUpTo. It must be stable across restarts, since
	// it's how a future "is this file safe to delete from the source
	// device" check would tell "already backed up here" apart from "not
	// yet" for a given configured destination.
	Name() string

	// Copy stores the file currently at localPath (its path inside
	// STORAGE_DIR) under relativePath (e.g.
	// "2026/08/travel/travel_..._beach.jpg" — the same relative layout
	// STORAGE_DIR uses) at this destination, creating whatever directory
	// structure or remote key prefix it needs along the way.
	Copy(ctx context.Context, relativePath, localPath string) error
}
