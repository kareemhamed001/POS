package file

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	Upload(ctx context.Context, file *multipart.FileHeader, path string) (string, error)
	Delete(ctx context.Context, path string) error
	GetURL(path string) string
	Exists(ctx context.Context, path string) (bool, error)
}

// StorageConfig holds configuration for file storage
type StorageConfig struct {
	Type          string // "local" or "s3"
	LocalBasePath string // For local storage
	LocalBaseURL  string // URL prefix for local files
	S3Bucket      string
	S3Region      string
	S3AccessKey   string
	S3SecretKey   string
	S3Endpoint    string // Optional for S3-compatible services
	S3BaseURL     string // Custom domain or CloudFront URL
	S3ACL         string // e.g., "public-read", "private"
}

// NewFileStorage creates a FileStorage instance based on config
func NewFileStorage(cfg StorageConfig) (FileStorage, error) {
	switch strings.ToLower(cfg.Type) {
	case "s3":
		return NewS3Storage(cfg)
	case "local", "":
		return NewLocalStorage(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
}

// LocalStorage implements FileStorage for local disk
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a new local storage handler
func NewLocalStorage(cfg StorageConfig) *LocalStorage {
	basePath := cfg.LocalBasePath
	if basePath == "" {
		basePath = "uploads"
	}
	baseURL := cfg.LocalBaseURL
	if baseURL == "" {
		baseURL = "/uploads"
	}
	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

// Upload saves a file to local disk
func (l *LocalStorage) Upload(ctx context.Context, file *multipart.FileHeader, path string) (string, error) {
	// Ensure directory exists
	dir := filepath.Join(l.basePath, filepath.Dir(path))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Full destination path
	dst := filepath.Join(l.basePath, path)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination file
	out, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	// Copy file content
	if _, err := io.Copy(out, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return l.GetURL(path), nil
}

// Delete removes a file from local disk
func (l *LocalStorage) Delete(ctx context.Context, path string) error {
	dst := filepath.Join(l.basePath, path)
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetURL returns the URL path for accessing the file
func (l *LocalStorage) GetURL(path string) string {
	// Normalize path separators for URLs
	urlPath := filepath.ToSlash(path)
	return l.baseURL + "/" + strings.TrimPrefix(urlPath, "/")
}

// Exists checks if a file exists on local disk
func (l *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	dst := filepath.Join(l.basePath, path)
	_, err := os.Stat(dst)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// S3Storage implements FileStorage for AWS S3
type S3Storage struct {
	bucket   string
	region   string
	baseURL  string
	acl      string
	uploader *manager.Uploader
	s3Client *s3.Client
}

// NewS3Storage creates a new S3 storage handler
func NewS3Storage(cfg StorageConfig) (*S3Storage, error) {
	// Load AWS config
	loadOptions := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.S3Region),
	}

	if cfg.S3AccessKey != "" && cfg.S3SecretKey != "" {
		staticCredentials := credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKey,
			cfg.S3SecretKey,
			"",
		)
		loadOptions = append(loadOptions, config.WithCredentialsProvider(staticCredentials))
	}

	if cfg.S3Endpoint != "" {
		customEndpoint := cfg.S3Endpoint
		endpointResolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{URL: customEndpoint, SigningRegion: cfg.S3Region}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			},
		)
		loadOptions = append(loadOptions, config.WithEndpointResolverWithOptions(endpointResolver))
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		if cfg.S3Endpoint != "" {
			options.UsePathStyle = true
		}
	})
	uploader := manager.NewUploader(s3Client)

	baseURL := cfg.S3BaseURL
	if baseURL == "" {
		if cfg.S3Endpoint != "" {
			baseURL = fmt.Sprintf("%s/%s", cfg.S3Endpoint, cfg.S3Bucket)
		} else {
			baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.S3Bucket, cfg.S3Region)
		}
	}

	acl := cfg.S3ACL
	if acl == "" {
		acl = "public-read"
	}

	return &S3Storage{
		bucket:   cfg.S3Bucket,
		region:   cfg.S3Region,
		baseURL:  baseURL,
		acl:      acl,
		uploader: uploader,
		s3Client: s3Client,
	}, nil
}

// Upload saves a file to S3
func (s *S3Storage) Upload(ctx context.Context, file *multipart.FileHeader, path string) (string, error) {
	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Prepare upload input
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(path),
		Body:        src,
		ACL:         types.ObjectCannedACL(s.acl),
		ContentType: aws.String(file.Header.Get("Content-Type")),
	}

	// Upload to S3
	result, err := s.uploader.Upload(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Return custom URL if configured, otherwise use S3 location
	if s.baseURL != "" {
		return s.GetURL(path), nil
	}
	return result.Location, nil
}

// Delete removes a file from S3
func (s *S3Storage) Delete(ctx context.Context, path string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	_, err := s.s3Client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	return nil
}

// GetURL returns the URL for accessing the file
func (s *S3Storage) GetURL(path string) string {
	// Normalize path separators for URLs
	urlPath := filepath.ToSlash(path)
	return s.baseURL + "/" + strings.TrimPrefix(urlPath, "/")
}

// Exists checks if a file exists in S3
func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	_, err := s.s3Client.HeadObject(ctx, input)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GenerateFileName creates a unique filename with timestamp
func GenerateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d%s", timestamp, ext)
}

// GenerateFilePath creates a dated path structure (e.g., "2026/01/filename.jpg")
func GenerateFilePath(filename string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	return filepath.Join(year, month, filename)
}
