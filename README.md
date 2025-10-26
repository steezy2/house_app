# House App

A comprehensive house management API built with Go, providing services for image and file storage, and more. This application serves as a centralized hub for managing various household digital assets.

## Features

- **Image Storage**: Upload, store, and manage images with metadata
- **Bulk Upload**: Upload multiple images from a source folder in one operation
- **Search Functionality**: Search through stored images by filename or tags
- **RESTful API**: Clean and well-documented API endpoints
- **MongoDB Integration**: Persistent storage with MongoDB
- **Swagger Documentation**: Interactive API documentation

## Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: MongoDB
- **API Documentation**: Swagger/OpenAPI
- **Environment Management**: godotenv

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
- file (file, required): Image file to upload
- tags (string, optional): Comma-separated tags
```

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
├── service/             # Business logic layer
│   ├── service.go       # Core business logic
│   └── service_test.go
└── uploads/             # Uploaded files storage (created automatically)
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

If available, you can use the Makefile for common tasks:

```bash
# Run tests
make test

# Build the application
make build

# Run the application
make run
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MONGO_URI` | MongoDB connection string | `mongodb://localhost:27017` |

## Database Schema

### Images Collection

```javascript
{
  _id: ObjectId,
  filename: String,        // Original filename
  originalPath: String,    // Original file path
  storagePath: String,     // Path where file is stored
  size: Number,           // File size in bytes
  contentType: String,    // MIME type
  tags: [String],         // Array of tags
  metadata: Object,       // Additional metadata
  createdAt: Date        // Creation timestamp
}
```

## Future Enhancements

- [ ] File storage support (documents, videos)
- [ ] Image resizing and optimization
- [ ] User authentication and authorization
- [ ] Cloud storage integration (S3, Google Cloud Storage)
- [ ] Advanced search with filters
- [ ] Image galleries and albums
- [ ] Thumbnail generation
- [ ] Duplicate detection

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
