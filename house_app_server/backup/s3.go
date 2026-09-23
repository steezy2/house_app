package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config holds the settings needed to reach any S3-compatible bucket.
// Endpoint is what actually selects the provider — AWS S3, Backblaze B2,
// Cloudflare R2, and MinIO all speak the same API, just at different
// hosts — so this isn't locked to one of them.
type S3Config struct {
	// Endpoint is the provider's S3-compatible API URL, e.g.
	// "https://s3.us-west-002.backblazeb2.com" for Backblaze B2 or
	// "https://<account-id>.r2.cloudflarestorage.com" for Cloudflare R2.
	// Leave empty to use AWS S3's default endpoints.
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	// Region is required by the SDK even for providers that otherwise
	// ignore it (Backblaze B2 and R2 both accept "auto" or "us-east-1").
	Region string
}

// S3Destination uploads files to an S3-compatible object store — the only
// backup.Destination that's actually off-site, protecting against a local
// disaster (fire, theft, disk failure) that a second local drive wouldn't.
type S3Destination struct {
	name   string
	bucket string
	client *s3.Client
}

// NewS3Destination builds an S3Destination from cfg. It doesn't verify
// connectivity or credentials — a bad bucket name, endpoint, or key pair
// only surfaces on the first Copy call.
func NewS3Destination(ctx context.Context, cfg S3Config) (*S3Destination, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		// Most non-AWS S3-compatible providers require path-style
		// requests (bucket in the URL path, not a subdomain).
		o.UsePathStyle = true
	})

	return &S3Destination{
		name:   fmt.Sprintf("s3:%s", cfg.Bucket),
		bucket: cfg.Bucket,
		client: client,
	}, nil
}

func (d *S3Destination) Name() string {
	return d.name
}

func (d *S3Destination) Copy(ctx context.Context, relativePath, localPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer file.Close()

	// Object keys use forward slashes regardless of the host OS.
	key := filepath.ToSlash(relativePath)

	if _, err := d.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
		Body:   file,
	}); err != nil {
		return fmt.Errorf("failed to upload to s3 destination: %w", err)
	}

	return nil
}

var _ Destination = (*S3Destination)(nil)
