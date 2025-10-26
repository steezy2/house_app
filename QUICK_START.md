# Quick Start Guide - House App

## 🚀 Getting Started in 5 Minutes

### Step 1: Prerequisites Check
Ensure you have the following installed:
- ✅ Go 1.21 or higher (`go version`)
- ✅ MongoDB (local or cloud instance)

### Step 2: Clone & Navigate
```bash
cd house_app/house_app_server
```

### Step 3: Configure Environment
Create a `.env` file:
```bash
cp .env.example .env
```

Edit `.env` with your MongoDB connection string:
```env
MONGO_URI=mongodb://localhost:27017
```

### Step 4: Install Dependencies
```bash
go mod download
```

### Step 5: Run the Server
```bash
go run main.go
```

You should see:
```
Connected to MongoDB and using database 'house_app'
Starting server on :8080
```

### Step 6: Test the API

#### Open Swagger UI
Visit: http://localhost:8080/swagger/index.html

#### Or test with curl:
```bash
# Get all images (will be empty initially)
curl http://localhost:8080/api/v1/images

# Upload an image
curl -X POST http://localhost:8080/api/v1/images \
  -F "file=@/path/to/image.jpg" \
  -F "tags=test"
```

## 📊 Setting Up MongoDB Text Index (Optional)

To enable search functionality, create a text index:

```javascript
// Connect to MongoDB shell
mongosh

// Use the database
use house_app

// Create text index
db.images.createIndex({ filename: "text", tags: "text" })
```

## 🔧 Development Commands

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
make go-test-html

# Build binary
make build

# Format code
go fmt ./...

# Regenerate Swagger docs
swag init
```

## 📁 Testing Bulk Upload

1. Create a test folder with images:
```bash
mkdir ~/test-images
# Add some .jpg, .png files to this folder
```

2. Use the API to bulk upload:
```bash
curl -X POST http://localhost:8080/api/v1/images/bulk-upload \
  -H "Content-Type: application/json" \
  -d '{
    "sourceFolder": "/home/user/test-images",
    "tags": ["test", "bulk"]
  }'
```

## 🐛 Troubleshooting

### "MONGO_URI environment variable not set"
- Ensure `.env` file exists in `house_app_server/` directory
- Verify `MONGO_URI` is set in the `.env` file

### "Failed to connect to MongoDB"
- Check if MongoDB is running: `mongosh`
- Verify the connection string in `.env`
- For MongoDB Atlas, ensure IP whitelist is configured

### "Port already in use"
- Check if another process is using port 8080
- Kill the process or change the port in `main.go`

### Uploads not working
- Ensure the application has write permissions in the current directory
- The `./uploads` folder will be created automatically

## 📚 Next Steps

1. **Explore the API**: Check out `API_EXAMPLES.md` for detailed curl examples
2. **Read Documentation**: See `README.md` for comprehensive documentation
3. **View Migration Notes**: Check `MIGRATION_SUMMARY.md` for details on changes from bookmark server

## 🎯 Quick API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/images` | Upload single image |
| GET | `/api/v1/images` | Get all images |
| GET | `/api/v1/images/search?q=query` | Search images |
| POST | `/api/v1/images/bulk-upload` | Bulk upload from folder |

## 💡 Tips

- Use Swagger UI for interactive testing
- Check server logs for detailed error messages
- Run tests before deploying changes
- Keep MongoDB connection string secure
- Use environment variables for configuration

## 🔐 Security Notes

- Never commit `.env` file to version control
- Use strong authentication for production MongoDB
- Consider implementing API authentication
- Validate file types and sizes on upload
- Sanitize file names to prevent directory traversal

---

**Need help?** Open an issue on GitHub or check the documentation files.
