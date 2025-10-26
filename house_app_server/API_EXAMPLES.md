# House App API Examples

This document provides examples of how to use the House App API using curl commands.

## Prerequisites

- Server running on `http://localhost:8080`
- curl installed on your system

## Upload Single Image

Upload a single image file with optional tags:

```bash
curl -X POST http://localhost:8080/api/v1/images \
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
  "tags": ["vacation,beach,2025"],
  "createdAt": "2025-10-26T10:30:00Z"
}
```

## Get All Images

Retrieve all stored images:

```bash
curl -X GET http://localhost:8080/api/v1/images
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
curl -X GET "http://localhost:8080/api/v1/images/search?q=vacation"
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

### Supported Image Formats

The bulk upload endpoint supports the following image formats:
- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- BMP (.bmp)
- WebP (.webp)

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

4. **Tags**: Tags are stored as provided. For the single upload endpoint, the entire tags form field is stored as one tag. You may want to modify the API to split comma-separated tags into an array.

## Swagger UI

For interactive API documentation, visit:
```
http://localhost:8080/swagger/index.html
```
