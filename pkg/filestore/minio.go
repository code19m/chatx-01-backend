package filestore

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Store defines the interface for file storage operations.
type Store interface {
	// Exists checks if a file exists at the given path.
	Exists(ctx context.Context, path string) (bool, error)

	// GetContentType retrieves the MIME type of a file.
	GetContentType(ctx context.Context, path string) (string, error)

	// Upload uploads a file to the storage.
	// Returns the path where the file was stored.
	Upload(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error

	// Download downloads a file from the storage.
	// Returns a reader for the file content.
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file from the storage.
	Delete(ctx context.Context, path string) error
}

// minioStore implements Store interface using MinIO client SDK.
type minioStore struct {
	client *minio.Client
	bucket string
}

// Config holds configuration for MinIO/S3 storage.
type Config struct {
	Endpoint        string // MinIO endpoint (host:port)
	Bucket          string // Bucket name
	AccessKeyID     string // Access key
	SecretAccessKey string // Secret key
	UseSSL          bool   // Use HTTPS instead of HTTP
}

// NewMinioStore creates a new MinIO file store using the official MinIO client SDK.
func NewMinioStore(cfg Config) Store {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create minio client: %v", err))
	}

	return &minioStore{
		client: client,
		bucket: cfg.Bucket,
	}
}

// Exists checks if a file exists at the given path.
func (s *minioStore) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat object: %w", err)
	}
	return true, nil
}

// GetContentType retrieves the content type of a file.
func (s *minioStore) GetContentType(ctx context.Context, path string) (string, error) {
	objInfo, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return "", fmt.Errorf("file not found")
		}
		return "", fmt.Errorf("failed to stat object: %w", err)
	}

	contentType := objInfo.ContentType
	if contentType == "" {
		return "application/octet-stream", nil
	}

	return contentType, nil
}

// Upload uploads a file to MinIO storage.
func (s *minioStore) Upload(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	_, err := s.client.PutObject(ctx, s.bucket, path, reader, size, opts)
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

// Download downloads a file from MinIO storage.
func (s *minioStore) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	// Check if object exists by reading stat
	_, err = obj.Stat()
	if err != nil {
		obj.Close()
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	return obj, nil
}

// Delete removes a file from MinIO storage.
func (s *minioStore) Delete(ctx context.Context, path string) error {
	err := s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}
