package services

import (
	"context"
	"io"
)

// BucketConfig represents the configuration for creating a bucket
type BucketConfig struct {
	Name       string
	Region     string
	Versioning bool
	ACL        string
}

// Object represents a storage object
type Object struct {
	Key          string
	Size         int64
	LastModified string
	ETag         string
}

// Storage defines the interface for storage operations
type Storage interface {
	CreateBucket(ctx context.Context, config *BucketConfig) error
	ListBuckets(ctx context.Context) ([]string, error)
	DeleteBucket(ctx context.Context, name string) error
	PutObject(ctx context.Context, bucket, key string, body io.Reader) error
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	ListObjects(ctx context.Context, bucket string) ([]*Object, error)
}
