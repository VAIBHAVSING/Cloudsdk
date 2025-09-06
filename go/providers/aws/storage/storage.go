package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3ClientInterface defines methods we need from S3 client for testing
type S3ClientInterface interface {
	CreateBucket(ctx context.Context, input *s3.CreateBucketInput, opts ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	ListBuckets(ctx context.Context, input *s3.ListBucketsInput, opts ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	PutBucketVersioning(ctx context.Context, input *s3.PutBucketVersioningInput, opts ...func(*s3.Options)) (*s3.PutBucketVersioningOutput, error)
}

// AWSStorage implements the Storage interface for AWS
type AWSStorage struct {
	client S3ClientInterface
}

// New creates a new AWSStorage instance with real AWS client
func New(cfg aws.Config) services.Storage {
	client := s3.NewFromConfig(cfg)
	return &AWSStorage{client: client}
}

// NewWithClient creates a new AWSStorage instance with custom client (for testing)
func NewWithClient(client S3ClientInterface) services.Storage {
	return &AWSStorage{client: client}
}

// CreateBucket creates a new S3 bucket
func (s *AWSStorage) CreateBucket(ctx context.Context, config *services.BucketConfig) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(config.Name),
	}

	if config.Region != "" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(config.Region),
		}
	}

	_, err := s.client.CreateBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	if config.Versioning {
		_, err = s.client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
			Bucket: aws.String(config.Name),
			VersioningConfiguration: &types.VersioningConfiguration{
				Status: types.BucketVersioningStatusEnabled,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to enable versioning: %w", err)
		}
	}

	return nil
}

// ListBuckets lists all S3 buckets
func (s *AWSStorage) ListBuckets(ctx context.Context) ([]string, error) {
	resp, err := s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	buckets := make([]string, len(resp.Buckets))
	for i, b := range resp.Buckets {
		buckets[i] = aws.ToString(b.Name)
	}
	return buckets, nil
}

// Other methods - stub for now
func (s *AWSStorage) DeleteBucket(ctx context.Context, name string) error {
	return fmt.Errorf("not implemented")
}

func (s *AWSStorage) PutObject(ctx context.Context, bucket, key string, body io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (s *AWSStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *AWSStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	return fmt.Errorf("not implemented")
}

func (s *AWSStorage) ListObjects(ctx context.Context, bucket string) ([]*services.Object, error) {
	return nil, fmt.Errorf("not implemented")
}
