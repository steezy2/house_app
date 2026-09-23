# House App API Examples

This document provides examples of how to use the House App API using curl commands.

## Prerequisites

- Server running on `http://localhost:8080`
- curl installed on your system
- If `API_KEY` is set on the server, every request below needs
  `-H "X-API-Key: $API_KEY"` or it gets a `401`. Examples below assume
  `API_KEY` is exported in your shell; drop the header if you're running
  without one (local dev only).

## Upload Single Image

Upload a single image file with optional tags:

```bash
curl -X POST http://localhost:8080/api/v1/images \
  -H "X-API-Key: $API_KEY" \
  -F "file=@/path/to/your/image.jpg" \
  -F "tags=vacation,beach,2025"
```

Response:
```json
{
  "id": "507f1f77bcf86cd799439011",
  "filename": "image.jpg",
  "originalPath": "image.jpg",
  "storagePath": "./uploads/1730000000_image.jpg",
  "size": 102400,
  "contentType": "image/jpeg",
  "tags": ["vacation", "beach", "2025"],
  "createdAt": "2025-10-26T10:30:00Z"
}
```

Only `.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, and `.webp` files are
accepted — anything else gets a `400`.

## Upload Multiple Images

Upload several image files in a single multipart request (repeat the
`files` field once per file):

```bash
curl -X POST http://localhost:8080/api/v1/images/multiple \
  -H "X-API-Key: $API_KEY" \
  -F "files=@/path/to/photo1.jpg" \
  -F "files=@/path/to/photo2.jpg" \
  -F "tags=family,2025"
```

This is the endpoint a "select photos and upload" button on a phone would
call — pass one `files` part per selected photo in a single request.

Response:
```json
{
  "totalImages": 2,
  "successfulCount": 2,
  "failedCount": 0,
  "failedFiles": []
}
```

`failedFiles` entries (when present) are formatted as
`"<filename>: <error message>"`, not bare paths like the folder-based bulk
upload below.

## Get All Images

Retrieve all stored images:

```bash
curl -X GET http://localhost:8080/api/v1/images \
  -H "X-API-Key: $API_KEY"
```

Response:
```json
[
  {
    "id": "507f1f77bcf86cd799439011",
    "filename": "photo1.jpg",
    "storagePath": "./uploads/1730000000_photo1.jpg",
    "size": 102400,
    "contentType": "image/jpeg",
    "tags": ["vacation"],
    "createdAt": "2025-10-26T10:30:00Z"
  },
  {
    "id": "507f1f77bcf86cd799439012",
    "filename": "photo2.png",
    "storagePath": "./uploads/1730000001_photo2.png",
    "size": 204800,
    "contentType": "image/png",
    "tags": ["family"],
    "createdAt": "2025-10-26T11:00:00Z"
  }
]
```

## Search Images

Search for images by filename or tags:

```bash
curl -X GET "http://localhost:8080/api/v1/images/search?q=vacation" \
  -H "X-API-Key: $API_KEY"
```

Response:
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

## Bulk Upload Images

Upload multiple images from a folder:

```bash
curl -X POST http://localhost:8080/api/v1/images/bulk-upload \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "sourceFolder": "/path/to/images/folder",
    "tags": ["family", "2025", "reunion"]
  }'
```

Response:
```json
{
  "totalImages": 100,
  "successfulCount": 98,
  "failedCount": 2,
  "failedFiles": [
    "/path/to/images/folder/corrupted.jpg",
    "/path/to/images/folder/invalid.txt"
  ]
}
```

### Supported Formats

Every upload endpoint (single, multiple, and bulk) supports:
- Images: JPEG (.jpg, .jpeg), PNG (.png), GIF (.gif), BMP (.bmp), WebP
  (.webp), HEIC/HEIF (.heic, .heif)
- Video: MP4 (.mp4), QuickTime (.mov), M4V (.m4v), 3GP (.3gp), WebM
  (.webm), AVI (.avi)

## Notes

1. **Text Index Required for Search**: To use the search functionality, you need to create a text index on your MongoDB collection. Run the following command in the MongoDB shell:

```javascript
db.images.createIndex({
  filename: "text",
  tags: "text"
})
```

2. **File Paths**: For the bulk upload endpoint, use absolute paths on Windows (e.g., `C:\\Users\\username\\Pictures`) or Unix-style paths on Linux/Mac (e.g., `/home/username/pictures`).

3. **Storage**: Uploaded files are stored in the `./uploads` directory relative to where the server is running.

4. **Tags**: For the single- and multiple-file upload endpoints, the `tags`
   form field is split on commas and trimmed (`"a, b"` → `["a", "b"]`).
   The folder-based bulk-upload endpoint takes `tags` as a JSON array
   directly.

5. **Authentication**: set `API_KEY` on the server and pass it as
   `X-API-Key` on every request (see Prerequisites above) — without it,
   every endpoint here is open to anyone on the network. Auth does not
   sandbox *which* folders `bulk-upload` can read, only who can call it —
   only point that endpoint at a server you trust with full filesystem
   access.

## Swagger UI

For interactive API documentation, visit:
```
http://localhost:8080/swagger/index.html
```
