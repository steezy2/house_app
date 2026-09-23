# Changelog

All notable changes to this project will be documented in this file.

## 2026-09-23

### Added
- **Video support** end to end: `models.SupportedVideoExtensions`
  (mp4/mov/m4v/3gp/webm/avi) alongside the existing image allowlist,
  consulted by upload validation (`service.checkSupportedExt`), MIME-type
  detection (`contentTypeForExt`), and the background processor
  (`processor.isMediaFile`). Videos get their own `"video"` category
  instead of falling into `"general"`, and EXIF extraction now skips them
  outright instead of always attempting and failing.
- **Backup & redundancy**: new `backup` package with a `Destination`
  interface, a `LocalDestination` implementation (mirror to any number of
  local/network drives), and an `S3Destination` implementation (any
  S3-compatible provider — AWS S3, Backblaze B2, Cloudflare R2, MinIO).
  Configured via `BACKUP_LOCAL_DIRS` (comma-separated, 0 to N paths) and
  optional `BACKUP_S3_*` env vars. The processor mirrors each file to
  every configured destination right after moving it and records which
  ones succeeded in the new `models.Image.BackedUpTo` field
  (`clients.MongoRepo.UpdateBackupStatus`); one destination failing never
  blocks the others or the primary copy.
- **Mobile "Upload All Photos & Videos"**: `lib/services/media_library.dart`
  wraps `photo_manager` for full camera-roll access (photos and video),
  distinct from the existing hand-pick flow. Uploads are split into
  batches of ≤20 files/≤150MB (`lib/utils/batching.dart`) with a
  confirmation dialog showing the real asset count first and a progress
  bar during upload. Explicitly detects and surfaces iOS's "Limited
  Photos" access mode rather than silently uploading a partial library.
  This does not delete anything from the device — see Known limitations.

### Fixed
- `service.UploadImage` was trusting the client's declared `Content-Type`
  verbatim, including the generic `application/octet-stream` fallback
  many upload libraries send for a type they don't recognize. It now
  falls back to extension-based detection in that case, matching what
  `BulkUploadImages` already did.

### Changed
- Go module bumped from 1.21 to 1.24 (required by the AWS SDK v2
  dependency `backup/` uses for S3-compatible storage).

### Known limitations
- No chunked/resumable upload — a large video or oversized batch that
  fails partway through must be retried from scratch.
- No backfill of pre-existing `STORAGE_DIR` content into newly-added
  backup destinations; only files processed going forward are mirrored.
- Deleting originals from the phone after upload is not built yet —
  intentionally deferred until `BackedUpTo` can be trusted to cover every
  configured backup destination for a given file, so "clean the phone"
  can't outrun backup verification and risk data loss.

## 2026-08-05

### Changed
- Deduplicated logic that had drifted across the server and mobile app:
  - Server: unified the separate "supported image extension" lists in
    `service` and `processor` into `models.SupportedImageExtensions`;
    extracted the shared open/copy/build-record flow in
    `service.UploadImage`/`BulkUploadImages` into one `storeImage` helper;
    factored the repeated `ID`/`CreatedAt` setup in `clients.MongoRepo`
    into `newImageRecord`; added a `respondError` helper for the repeated
    JSON error shape in `controller`; consolidated three hand-copied
    `MockImageRepository` test doubles into one nil-safe
    `internal/mocks` package.
  - Mobile: extracted the error-classification logic duplicated between
    the Upload and Settings screens into `lib/utils/error_message.dart`
    (mirroring the sibling Insurance Marketplace app's own convention);
    factored the repeated `FakeSettingsStorage` test setup into
    `test/test_helpers.dart`.

## 2026-08-04

### Added
- **Flutter mobile app** (`house_app_mobile`): Upload tab (pick photos,
  optional tags, upload to `POST /images/multiple`), Library tab
  (metadata-only list via `GET /images`, no thumbnails since the server
  has no endpoint to serve image bytes), and Settings tab (server
  address + API key, stored in the platform keychain). Layout mirrors the
  sibling Insurance Marketplace mobile app's conventions, minus
  `go_router` and a login flow (neither needed here). Includes the
  Android/iOS platform config required to reach a plain-HTTP LAN server
  (`usesCleartextTraffic` / `NSAllowsLocalNetworking`), since both
  platforms block that by default.
- **Background image processor**: polls `UPLOAD_DIR` on a timer, extracts
  EXIF metadata, categorizes by metadata/filename keywords, generates a
  descriptive filename, and moves each file into a dated
  `STORAGE_DIR/YYYY/MM/category/` tree, updating the database record to
  match (`clients.MongoRepo.UpdateImagePath`).
- **API key authentication**: `controller.APIKeyMiddleware` gates every
  `/api/v1/*` route behind a shared `X-API-Key` header when `API_KEY` is
  configured (`/swagger/*` stays open). Logs a startup warning if unset
  rather than failing closed, since local development without it is a
  supported case.

### Fixed
- Tag splitting: `service.ParseTags` now actually splits and trims the
  comma-separated `tags` form value on single/multiple upload — previously
  the whole string was stored as one tag.
- `storagePath` going stale once the processor moved a file — including
  the case where a filename collision forced the processor to rename the
  destination (`_1`, `_2`, ...), which previously wasn't reflected back
  into the database record at all.
- Client-supplied filenames weren't sanitized before being joined into the
  storage path (a path-traversal risk via a crafted filename).
- Single/multiple upload had no content-type/extension validation — only
  the folder-based bulk-upload path enforced the image-format allowlist.
- A non-deterministic categorization bug: `processor`'s keyword-to-category
  table was a Go `map`, and Go randomizes map iteration order, so a
  filename matching multiple keywords (e.g. `beach_sunset.jpg` matching
  both `beach` and `sunset`) could resolve to a different category on
  different runs. Switched to an ordered slice so the first match always
  wins deterministically.

## 2025-10-26

### Added
- Initial House App server (`house_app_server`), migrated from an earlier
  bookmark-manager project (see `MIGRATION_SUMMARY.md`): Go/Gin API with
  MongoDB-backed image storage, single/bulk upload, filename/tag search,
  and generated Swagger documentation.
