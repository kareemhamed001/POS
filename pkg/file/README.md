# File Storage Handler

A flexible file storage abstraction that supports both local disk and AWS S3 storage backends.

## Features

- **Interface-based design** - Easy to switch between storage backends
- **Local Storage** - Save files to local disk
- **S3 Storage** - Upload files to AWS S3 or S3-compatible services
- **Automatic path generation** - Date-based directory structure (YYYY/MM/)
- **Unique filenames** - Timestamp-based naming to prevent collisions

## Configuration

### Environment Variables

```bash
# Storage Type ("local" or "s3")
STORAGE_TYPE=local

# Local Storage Settings
STORAGE_LOCAL_PATH=uploads
STORAGE_LOCAL_URL=/uploads

# S3 Storage Settings
S3_BUCKET=my-bucket
S3_REGION=us-east-1
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
S3_ENDPOINT=               # Optional: for S3-compatible services (e.g., MinIO)
S3_BASE_URL=               # Optional: custom domain or CloudFront URL
S3_ACL=public-read         # S3 ACL (public-read, private, etc.)
```

## Usage

### Initialize Storage

```go
import "github.com/kareemhamed001/POS/pkg/file"

// From config
fileStorage, err := file.NewFileStorage(file.StorageConfig{
    Type:            config.StorageType,
    LocalBasePath:   config.StorageLocalPath,
    LocalBaseURL:    config.StorageLocalURL,
    S3Bucket:        config.S3Bucket,
    S3Region:        config.S3Region,
    S3AccessKey:     config.S3AccessKey,
    S3SecretKey:     config.S3SecretKey,
    S3Endpoint:      config.S3Endpoint,
    S3BaseURL:       config.S3BaseURL,
    S3ACL:           config.S3ACL,
})
if err != nil {
    log.Fatal(err)
}
```

### Upload Files

```go
// Generate unique path
filename := file.GenerateFileName(fileHeader.Filename)  // e.g., "1642345678901234567.jpg"
filePath := file.GenerateFilePath(filename)             // e.g., "2026/01/1642345678901234567.jpg"

// Upload (works for both local and S3)
imageURL, err := fileStorage.Upload(ctx, fileHeader, filePath)
if err != nil {
    return err
}
// imageURL: "/uploads/2026/01/1642345678901234567.jpg" (local)
//           "https://bucket.s3.region.amazonaws.com/2026/01/1642345678901234567.jpg" (S3)
```

### Delete Files

```go
err := fileStorage.Delete(ctx, "2026/01/1642345678901234567.jpg")
```

### Check if File Exists

```go
exists, err := fileStorage.Exists(ctx, "2026/01/1642345678901234567.jpg")
```

### Get File URL

```go
url := fileStorage.GetURL("2026/01/1642345678901234567.jpg")
```

## Storage Backends

### Local Storage

- Files saved to disk at `STORAGE_LOCAL_PATH`
- Accessed via `STORAGE_LOCAL_URL` prefix
- Served by Gin's `router.Static()` middleware
- Example: `uploads/2026/01/file.jpg` → `/uploads/2026/01/file.jpg`

### S3 Storage

- Files uploaded to AWS S3 or compatible services
- Supports custom endpoints (MinIO, DigitalOcean Spaces, etc.)
- Configurable ACL (public-read, private)
- Optional CloudFront or custom domain URL

## Switching Between Storage Types

Simply change `STORAGE_TYPE` environment variable:

```bash
# Use local storage
STORAGE_TYPE=local

# Use S3 storage
STORAGE_TYPE=s3
S3_BUCKET=my-bucket
S3_REGION=us-east-1
S3_ACCESS_KEY=xxx
S3_SECRET_KEY=xxx
```

No code changes required!

## Example: Product Image Upload

```go
// In handler
fileHeader := productRequest.Image
filename := file.GenerateFileName(fileHeader.Filename)
filePath := file.GenerateFilePath(filename)

imageURL, err := h.fileStorage.Upload(ctx.Request.Context(), fileHeader, filePath)
if err != nil {
    return helper.WriteAPIResponse(ctx, nil, "failed to save image", http.StatusInternalServerError)
}

product.ImageUrl = imageURL
```

## Notes

- AWS SDK is required for S3 support: `go get github.com/aws/aws-sdk-go`
- For local storage, ensure the upload directory has write permissions
- For S3, IAM user needs `s3:PutObject`, `s3:GetObject`, `s3:DeleteObject` permissions
- CloudFront can be used with S3 for CDN: set `S3_BASE_URL=https://d123.cloudfront.net`
