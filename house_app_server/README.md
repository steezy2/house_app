# House App Server

This is the backend server for the House App. It is a Go application that provides a RESTful API for managing household digital assets including image storage, file management, and more.

## Features

*   Image upload and management (single and bulk)
*   Search functionality for images by filename and tags
*   MongoDB storage for metadata
*   RESTful API with Swagger documentation
*   Support for multiple image formats (JPEG, PNG, GIF, BMP, WebP)

## Getting Started

### Prerequisites

*   Go 1.21+
*   MongoDB
*   Make (optional)

### Installation

1.  Initialize the project and install dependencies:

    ```sh
    go mod download
    ```

2.  Create a `.env` file in the root directory with the following content:

    ```env
    MONGO_URI=mongodb://localhost:27017
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
- `GET /api/v1/images` - Get all images
- `GET /api/v1/images/search?q=query` - Search images
- `POST /api/v1/images/bulk-upload` - Bulk upload images from a folder

For detailed API documentation, see the Swagger UI when the server is running.

## Building

To build the application:

```sh
go build -o house_app main.go
```

Or with Make:

```sh
make build
```

This will create a `coverage.html` file in the root of the `bookmark_server` directory.

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
