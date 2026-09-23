# House App Server

This is the backend server for the House App. It is a Go application that provides a RESTful API for managing household digital assets including image storage, file management, and more.

## Features

*   Image and video upload and management (single and bulk)
*   Search functionality by filename and tags
*   MongoDB storage for metadata
*   RESTful API with Swagger documentation
*   Support for common image (JPEG, PNG, GIF, BMP, WebP, HEIC/HEIF) and
    video (MP4, MOV, M4V, 3GP, WebM, AVI) formats
*   Backup redundancy to any number of local/network drives and/or
    S3-compatible cloud storage (see [backup/](backup))

## Getting Started

### Prerequisites

*   Go 1.24+ (bumped from 1.21 by the AWS SDK dependency backup/ uses)
*   MongoDB
*   Make (optional)

### Installation

1.  Initialize the project and install dependencies:

    ```sh
    go mod download
    ```

2.  Create a `.env` file in this directory (copy `.env.example`). At
    minimum set `MONGO_URI`; also set `API_KEY` if you want the API to
    require authentication (recommended), and at least one `BACKUP_*`
    variable if you don't want uploaded files living only in `STORAGE_DIR`
    with no redundancy (also recommended — see [backup/](backup)):

    ```env
    MONGO_URI=mongodb://localhost:27017
    API_KEY=some-shared-secret
    BACKUP_LOCAL_DIRS=D:/house_app_backup
    ```

### Running the Server

To run the server directly:

```sh
go run main.go
```

Or if using Make:

```sh
make go-run
```

The server will start on `http://localhost:8080`.

### API Documentation

Once the server is running, you can access the Swagger API documentation at:

```
http://localhost:8080/swagger/index.html
```

### Testing

To run the unit tests, use the following command:

```sh
go test ./...
```

Or with Make:

```sh
make go-test
```

To generate an HTML test coverage report, use:

```sh
make go-test-html
```

## API Endpoints

### Images

- `POST /api/v1/images` - Upload a single image
- `POST /api/v1/images/multiple` - Upload multiple image files in one multipart request
- `GET /api/v1/images` - Get all images
- `GET /api/v1/images/search?q=query` - Search images
- `POST /api/v1/images/bulk-upload` - Bulk copy images from a folder path the server can read

For detailed API documentation, see the Swagger UI when the server is running.

All `/api/v1/*` requests above require an `X-API-Key` header matching
`API_KEY` if that variable is set — see
[../README.md#authentication](../README.md#authentication). `/swagger/*`
is not gated.

## Background Processing

A separate `processor/` package (started from `main.go`, not from the
`controller` package above) polls `UPLOAD_DIR` on a timer, moves files
into `STORAGE_DIR/YYYY/MM/category/`, mirrors each file to every
configured `backup.Destination` (see [backup/](backup) and
[../README.md#backup--redundancy](../README.md#backup--redundancy)), and
updates each image's database record with the new path and backup status.
See [../IMAGE_PROCESSING.md](../IMAGE_PROCESSING.md) for configuration and
[../README.md#future-enhancements](../README.md#future-enhancements) for
remaining known gaps (e.g. AI categorization is still a stub).

## Building

To build the application:

```sh
go build -o house_app main.go
```

Or with Make:

```sh
make build
```

This will create a `coverage.html` file in the root of the `house_app_server` directory.

## Development

### Code Formatting and Linting

The `go-clean` command is useful for maintaining code quality. It formats your code, runs `go vet`, tidies dependencies, and regenerates the API documentation. It's good practice to run this command before committing changes.

```sh
make go-clean
```

### API Documentation

This project uses Swagger for API documentation. The documentation is generated from code comments. To regenerate the documentation, run:

```sh
make docs
```

The generated Swagger JSON can be found at `docs/swagger.json`.
