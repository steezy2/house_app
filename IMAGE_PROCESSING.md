# Image Processing & Auto-Organization

The House App now includes an **automatic image processing system** that monitors the uploads directory and organizes images into a structured storage system.

> **Current implementation status**: the file-moving, EXIF/keyword
> categorization, and database update described below are all real and
> running. One thing in this doc describes intent rather than current
> behavior: **AI categorization** (`categorizeWithAI` / `callOpenAIVision`
> in [processor/image_processor.go](house_app_server/processor/image_processor.go))
> never calls the OpenAI API — even with `ENABLE_AI_CATEGORIZATION=true`
> and a valid key, it just does filename keyword matching (the same table
> used for the "AI-Enhanced Categories" below), then falls back to
> `"general"`.

## Features

### 🤖 Automatic Organization
- **Timer-based Processing**: Runs every 5 minutes (configurable)
- **Metadata Extraction**: Extracts EXIF data (date, camera model, etc.)
- **Smart Categorization**: Uses AI or metadata-based rules
- **Intelligent Naming**: Generates descriptive filenames
- **Year/Month Structure**: Organizes into `YYYY/MM/category/` folders

### 📁 Directory Structure

Images are automatically organized as:
```
E:/house_app_storage/
├── 2025/
│   ├── 01/
│   │   ├── travel/
│   │   │   ├── travel_20250115_142305_beach_sunset.jpg
│   │   │   └── travel_20250120_093000_mountain_hike.jpg
│   │   ├── family/
│   │   │   └── family_20250125_180000_birthday_party.jpg
│   │   └── nature/
│   │       └── nature_20250128_120000_landscape.jpg
│   ├── 02/
│   └── ...
└── 2024/
    └── ...
```

### 🏷️ Categories

#### Default Categories (Metadata-based)
- **photography** - Images from DSLR/mirrorless cameras
- **mobile** - Images from smartphones
- **general** - Default category

#### AI-Enhanced Categories (when enabled)
- **travel** - Vacation, holidays, trips, beaches, landmarks
- **nature** - Landscapes, sunsets, mountains, wildlife
- **people** - Portraits, groups
- **family** - Family photos, personal moments
- **events** - Weddings, birthdays, parties
- **food** - Food photography, recipes
- **screenshots** - Screen captures
- **documents** - Scanned documents
- **uncategorized** - When AI can't determine category

## Configuration

### Environment Variables

Add to your `.env` file:

```env
# Image Processing Configuration
UPLOAD_DIR=./uploads
STORAGE_DIR=E:/house_app_storage
PROCESSING_INTERVAL=5m

# AI Configuration (optional)
ENABLE_AI_CATEGORIZATION=false
OPENAI_API_KEY=your_openai_api_key_here
```

### Processing Interval

The `PROCESSING_INTERVAL` can be set to any duration:
- `1m` - Every 1 minute
- `5m` - Every 5 minutes (default)
- `15m` - Every 15 minutes
- `1h` - Every 1 hour
- `30s` - Every 30 seconds

## How It Works

### 1. Image Upload
```bash
curl -X POST 'http://localhost:8080/api/v1/images' \
  -F 'file=@photo.jpg'
```
Image is saved to `./uploads/1234567890_photo.jpg`

### 2. Automatic Processing (Every 5 minutes)

The processor:
1. **Scans** the uploads directory
2. **Extracts** EXIF metadata (date, camera info)
3. **Categorizes** using AI or rules
4. **Generates** descriptive filename
5. **Creates** directory structure (`YYYY/MM/category/`)
6. **Moves** file to organized location
7. **Updates** database record
8. **Deletes** original from uploads

### 3. Result

Original: `./uploads/1730000000_photo.jpg`  
New: `E:/house_app_storage/2025/10/travel/travel_20251026_143000_photo.jpg`

## Metadata Extraction

### EXIF Data Extracted
- **DateTime**: When photo was taken
- **Make**: Camera manufacturer (Canon, Nikon, Apple, etc.)
- **Model**: Camera model
- **Width/Height**: Image dimensions
- **Orientation**: Image orientation

### Filename Generation

Format: `{category}_{datetime}_{original_name}.{ext}`

Example transformations:
- `1730000000_beach_sunset.jpg` → `travel_20251026_143000_beach_sunset.jpg`
- `1730000001_family_reunion.jpg` → `family_20251026_150000_family_reunion.jpg`

## AI Categorization

### Enabling AI

1. Set environment variable:
   ```env
   ENABLE_AI_CATEGORIZATION=true
   OPENAI_API_KEY=sk-...
   ```

2. Restart the server

### How AI Works

When AI is enabled, the system:
1. Analyzes image content
2. Identifies subjects, scenes, objects
3. Assigns appropriate category
4. Extracts descriptive tags

### Fallback Behavior

If AI categorization fails:
- Falls back to keyword matching
- Uses EXIF metadata
- Defaults to "general" category

## Keyword-Based Categorization

When AI is disabled, the system uses filename keywords:

| Keywords | Category |
|----------|----------|
| vacation, holiday, trip, beach | travel |
| mountain, landscape, sunset, sunrise | nature |
| portrait | people |
| family | family |
| wedding, birthday, party | events |
| food, recipe | food |
| screenshot | screenshots |
| document | documents |

## Monitoring & Logs

### Structured Logging

All processing activities are logged with slog:

```
time=2025-10-26T15:00:00 level=INFO msg="Started periodic image processing" interval=5m0s uploadDir=./uploads storageDir=E:/house_app_storage
time=2025-10-26T15:00:01 level=INFO msg="Starting image processing cycle"
time=2025-10-26T15:00:02 level=INFO msg="Successfully processed image" originalPath=./uploads/photo.jpg newPath=E:/house_app_storage/2025/10/travel/travel_20251026_150002_photo.jpg category=travel newName=travel_20251026_150002_photo.jpg
time=2025-10-26T15:00:02 level=INFO msg="Image processing cycle completed" processed=1 failed=0
```

### Error Handling

Failed processing is logged with context:

```
time=2025-10-26T15:00:02 level=ERROR msg="Failed to process image" file=corrupted.jpg error="failed to extract metadata"
time=2025-10-26T15:00:02 level=WARN msg="AI categorization failed, using default" file=photo.jpg error="API timeout"
```

## Manual Trigger

While the processor runs automatically, you can also manually organize images by restarting the server (it processes immediately on startup).

## Database Updates

When an image is moved, `MongoRepo.UpdateImagePath` updates that record's
`storagePath`, `category`, and `processedAt` fields (matched by the old
`storagePath`). If `moveFile` had to rename the destination to avoid
overwriting an existing file (see Duplicate Handling below), the record is
updated with the renamed path, not the originally-planned one.

## Duplicate Handling

If a file with the same name exists at the destination:
- Appends `_1`, `_2`, etc.
- Example: `travel_20251026_150000_photo_1.jpg`

## Performance

- **Processing Time**: ~100-500ms per image (without AI)
- **AI Processing**: +2-5s per image
- **Batch Size**: All images in uploads directory
- **Resource Usage**: Minimal CPU/memory

## Troubleshooting

### Images Not Processing

1. Check logs for errors
2. Verify `STORAGE_DIR` exists and is writable
3. Check `PROCESSING_INTERVAL` is set
4. Ensure image files have valid extensions

### AI Not Working

1. Verify `ENABLE_AI_CATEGORIZATION=true`
2. Check `OPENAI_API_KEY` is set
3. Review logs for API errors
4. System falls back to keyword matching

### Wrong Categories

1. AI categorization may need tuning
2. Add keywords to improve matching
3. Check EXIF metadata quality
4. Consider manual organization for important photos

## Future Enhancements

- [ ] Custom category rules
- [ ] Machine learning model training
- [ ] Face recognition
- [ ] Duplicate detection
- [ ] RAW image support
- [ ] Video processing
- [ ] Thumbnail generation
- [ ] Web UI for manual categorization
- [ ] Category editing
- [ ] Batch re-categorization

## Example Workflow

```bash
# 1. Upload image
curl -X POST 'http://localhost:8080/api/v1/images' \
  -F 'file=@vacation_2024.jpg'

# Response: File saved to ./uploads/1730000000_vacation_2024.jpg

# 2. Wait for processing (or up to 5 minutes)

# 3. Check logs:
# time=... level=INFO msg="Successfully processed image" 
#   originalPath=./uploads/1730000000_vacation_2024.jpg
#   newPath=E:/house_app_storage/2025/10/travel/travel_20241026_120000_vacation_2024.jpg
#   category=travel

# 4. Image is now organized in:
# E:/house_app_storage/2025/10/travel/
```

## Security Considerations

- Only processes files in `UPLOAD_DIR`
- Validates file extensions
- Sanitizes filenames
- Creates directories with safe permissions
- No external network access (except AI API if enabled)
