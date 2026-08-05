# House App

A comprehensive house management API built with Go, providing services for image and file storage, and more. This application serves as a centralized hub for managing various household digital assets.

## Features

- **Image Storage**: Upload, store, and manage images with metadata
- **Bulk Upload**: Upload multiple images from a source folder in one operation
- **Search Functionality**: Search through stored images by filename or tags
- **RESTful API**: Clean and well-documented API endpoints
- **MongoDB Integration**: Persistent storage with MongoDB
- **Swagger Documentation**: Interactive API documentation
- **CORS Enabled**: Cross-Origin Resource Sharing for web applications
- **Structured Logging**: Clear error tracking with Go's slog package

## Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: MongoDB
- **API Documentation**: Swagger/OpenAPI
- **Environment Management**: godotenv
- **CORS**: Built-in middleware for cross-origin requests
- **Logging**: Structured logging with slog

## Getting Started

### Prerequisites

- Go 1.21 or higher
- MongoDB instance (local or cloud)
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/steezy2/house_app.git
cd house_app/house_app_server
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file in the `house_app_server` directory:
```env
MONGO_URI=mongodb://localhost:27017
```

4. Run the application:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Base URL
```
http://localhost:8080/api/v1
```

### Images

#### Upload Single Image
```http
POST /api/v1/images
Content-Type: multipart/form-data

Parameters:
- file (file, required): Image file to upload — must have a supported
  extension (see below), or the request is rejected with 400
- tags (string, optional): Comma-separated tags, split into an array
  (`"a, b"` → `["a", "b"]`)

Requires the `X-API-Key` header if `API_KEY` is configured — see
[Authentication](#authentication) below.

#### Upload Multiple Images
```http
POST /api/v1/images/multiple
Content-Type: multipart/form-data

Parameters:
- files (file[], required): Multiple image files (repeat the "files" field)
- tags (string, optional): Comma-separated tags applied to all files
```

**Response:** `models.BulkUploadResponse` (same shape as bulk upload below),
with `failedFiles` containing `"<filename>: <error>"` entries for any file
that failed to save.

**Response:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "filename": "photo.jpg",
  "originalPath": "photo.jpg",
  "storagePath": "./uploads/1234567890_photo.jpg",
  "size": 102400,
  "contentType": "image/jpeg",
  "tags": ["vacation", "beach"],
  "createdAt": "2025-10-26T10:30:00Z"
}
```

#### Get All Images
```http
GET /api/v1/images
```

**Response:**
```json
[
  {
    "id": "507f1f77bcf86cd799439011",
    "filename": "photo.jpg",
    "storagePath": "./uploads/1234567890_photo.jpg",
    "size": 102400,
    "contentType": "image/jpeg",
    "tags": ["vacation"],
    "createdAt": "2025-10-26T10:30:00Z"
  }
]
```

#### Search Images
```http
GET /api/v1/images/search?q=vacation
```

**Response:**
```json
[
  {
    "id": "507f1f77bcf86cd799439011",
    "filename": "beach_photo.jpg",
    "tags": ["vacation", "beach"],
    "createdAt": "2025-10-26T10:30:00Z"
  }
]
```

#### Bulk Upload Images
```http
POST /api/v1/images/bulk-upload
Content-Type: application/json

{
  "sourceFolder": "/path/to/images",
  "tags": ["family", "2025"]
}
```

**Response:**
```json
{
  "totalImages": 100,
  "successfulCount": 98,
  "failedCount": 2,
  "failedFiles": [
    "/path/to/images/corrupted.jpg",
    "/path/to/images/invalid.txt"
  ]
}
```

**Supported Image Formats:**
- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- BMP (.bmp)
- WebP (.webp)

> **Security note**: `sourceFolder` is still not sandboxed to any
> allowlist — the server will walk and copy any folder it has filesystem
> access to. Set `API_KEY` (see [Authentication](#authentication)) so this
> endpoint isn't reachable by anything but your own client; it does not
> restrict *which* folders can be requested, only *who* can request them.

## Authentication

If `API_KEY` is set (see [Environment Variables](#environment-variables)),
every `/api/v1/*` request must include a matching `X-API-Key` header or it
gets a `401`. The Swagger UI (`/swagger/*`) is not gated, since it's just
documentation. If `API_KEY` is unset, all endpoints are open — the server
logs a startup warning when that's the case. There's still only one shared
secret, not per-user accounts; that's intentional for a single-household
tool where the main client is a phone on the home network, not a
multi-tenant deployment.

```bash
curl -X POST http://localhost:8080/api/v1/images \
  -H "X-API-Key: $API_KEY" \
  -F "file=@/path/to/image.jpg"
```

## Background Image Processing

Uploaded files don't stay in `./uploads` — a background worker periodically
scans that directory, extracts EXIF metadata, picks a category, and moves
each file into a `STORAGE_DIR/YYYY/MM/category/` tree with a descriptive
name, then updates the image's database record to point at the new path.
See [IMAGE_PROCESSING.md](IMAGE_PROCESSING.md) for the full configuration
and behavior, including the current limitation that "AI" categorization is
still a keyword-matching stub, not a real vision-model call (see
[Future Enhancements](#future-enhancements)).

## Swagger Documentation

Access the interactive API documentation at:
```
http://localhost:8080/swagger/index.html
```

### Generating Swagger Docs

After making changes to API annotations, regenerate the Swagger documentation:

```bash
# Install swag CLI if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
swag init
```

The Swagger documentation is automatically updated from the comments in the code. The main annotations are in:
- `main.go`: General API information
- `controller/api.go`: Endpoint documentation

## Project Structure

```
house_app_server/
├── main.go              # Application entry point
├── go.mod               # Go module dependencies
├── clients/             # Data access layer
│   ├── mongo_repo.go    # MongoDB repository implementation
│   └── mongo_repo_test.go
├── controller/          # HTTP handlers and routing
│   ├── api.go           # API endpoints
│   └── api_test.go
├── docs/                # Swagger documentation (auto-generated)
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── models/              # Data models and interfaces
│   ├── models.go        # Data structures
│   ├── interfaces.go    # Repository interfaces
│   └── models_test.go
├── processor/           # Background job: organizes uploads into STORAGE_DIR
│   └── image_processor.go   # (no test file yet)
├── service/             # Business logic layer
│   ├── service.go       # Core business logic
│   └── service_test.go
└── uploads/             # Landing zone for new uploads (auto-created, auto-emptied
                          # by the processor — not a permanent store)
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Building for Production

```bash
# Build the binary
go build -o house_app main.go

# Run the binary
./house_app
```

### Using the Makefile

Actual targets, from [house_app_server/Makefile](house_app_server/Makefile):

```bash
# Run tests
make go-test

# Run tests + generate coverage.html
make go-test-html

# Build the application
make build

# Run the application
make go-run

# Regenerate Swagger docs
make docs

# fmt + vet + mod tidy + regenerate docs
make go-clean
```

## Environment Variables

All read in [main.go](house_app_server/main.go) /
[clients/mongo_repo.go](house_app_server/clients/mongo_repo.go) /
[processor/image_processor.go](house_app_server/processor/image_processor.go).
See [.env.example](house_app_server/.env.example) for a ready-to-copy file.

| Variable | Description | Default |
|----------|-------------|---------|
| `MONGO_URI` | MongoDB connection string (required — process exits if unset) | *(none)* |
| `SWAGGER_HOST` | Host shown in the Swagger UI's "try it out" requests | `localhost:8080` |
| `UPLOAD_DIR` | Landing directory for new uploads | `./uploads` |
| `STORAGE_DIR` | Root of the organized `YYYY/MM/category/` tree | `E:/house_app_storage` (Windows-specific — override on other OSes) |
| `PROCESSING_INTERVAL` | How often the background organizer runs, as a Go duration (`5m`, `1h`, ...) | `5m` |
| `ENABLE_AI_CATEGORIZATION` | Attempt AI-based categorization instead of metadata/keyword rules | `false` |
| `OPENAI_API_KEY` | Used only if `ENABLE_AI_CATEGORIZATION=true` — note the actual OpenAI call is not implemented yet, see [IMAGE_PROCESSING.md](IMAGE_PROCESSING.md) | *(none)* |
| `API_KEY` | Shared secret required as the `X-API-Key` header on every `/api/v1/*` request. Leave unset only for local development — the server logs a startup warning and runs unauthenticated | *(none — unauthenticated)* |

## Database Schema

### Images Collection

```javascript
{
  _id: ObjectId,
  filename: String,        // Sanitized original filename (path components stripped)
  originalPath: String,    // Same as filename for single/multiple upload;
                            // full source path for folder-based bulk-upload
  storagePath: String,     // Current file location — kept in sync by the
                            // background processor when it moves the file
  size: Number,             // File size in bytes
  contentType: String,      // MIME type
  tags: [String],           // Array of tags
  category: String,         // Set once the background processor categorizes the file
  metadata: Object,         // Additional metadata (currently unused)
  createdAt: Date,          // Upload timestamp
  processedAt: Date         // Set once the background processor moves the file
}
```

## Future Enhancements

### Fixed

- [x] **Auth** — `API_KEY` + `X-API-Key` header now gates every
  `/api/v1/*` route (still a single shared secret, not per-user accounts —
  see [Authentication](#authentication)).
- [x] **`storagePath` goes stale** — the processor now calls
  `UpdateImagePath` after moving a file, including when a filename
  collision forces it to append `_1`/`_2`/etc.
- [x] **Tag splitting** — `service.ParseTags` now splits and trims the
  `tags` form value on both upload endpoints.
- [x] **Filename handling on upload** — the client-supplied filename is
  reduced with `filepath.Base` before being joined into the storage path.
- [x] **No content-type/extension check on single/multiple upload** — both
  now enforce the same JPEG/PNG/GIF/BMP/WebP allowlist as bulk-upload.
- [x] **No test coverage for `processor/`** — added
  [processor/image_processor_test.go](house_app_server/processor/image_processor_test.go).

### Still open

- [ ] **`bulk-upload` path handling** — `sourceFolder` is walked with no
  allowlist/sandboxing beyond the `API_KEY` gate (see the security note
  under Bulk Upload above). Anyone holding the key can still request any
  folder the server process can read.
- [ ] **AI categorization is a stub** — `callOpenAIVision` never calls
  OpenAI; it does filename keyword matching. `OPENAI_API_KEY` is read but
  unused. See [IMAGE_PROCESSING.md](IMAGE_PROCESSING.md).
- [ ] **`STORAGE_DIR` default is an absolute Windows path** — fine for the
  current single-machine use case, but breaks silently (or errors) if run
  on another OS or another machine without setting the env var.
- [ ] **Single shared API key, no per-device revocation** — fine for one
  household, but if the phone client is ever lost, rotating `API_KEY`
  breaks every other client too, not just that device.

### Feature ideas

- [ ] File storage support (documents, videos)
- [ ] Cloud storage integration (S3, Google Cloud Storage)
- [ ] Advanced search with filters
- [ ] Image galleries and albums
- [ ] Thumbnail generation
- [ ] Duplicate detection
- [ ] Pagination on `GET /images` (currently returns the entire collection)

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is open source and available under the MIT License.

## Support

For issues, questions, or contributions, please open an issue on GitHub.
