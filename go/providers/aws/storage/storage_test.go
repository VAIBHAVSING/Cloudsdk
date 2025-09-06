package storage

import (
	"context"
	"testing"

	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
)

// mockS3Client is a mock implementation of the S3 client
type mockS3Client struct {
	createBucketResponse *s3.CreateBucketOutput
	createBucketError    error
	listBucketsResponse  *s3.ListBucketsOutput
	listBucketsError     error
}

func (m *mockS3Client) CreateBucket(ctx context.Context, input *s3.CreateBucketInput, opts ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	return m.createBucketResponse, m.createBucketError
}

func (m *mockS3Client) ListBuckets(ctx context.Context, input *s3.ListBucketsInput, opts ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	return m.listBucketsResponse, m.listBucketsError
}

func (m *mockS3Client) PutBucketVersioning(ctx context.Context, input *s3.PutBucketVersioningInput, opts ...func(*s3.Options)) (*s3.PutBucketVersioningOutput, error) {
	return &s3.PutBucketVersioningOutput{}, nil
}

func TestAWSStorage_CreateBucket(t *testing.T) {
	mockClient := &mockS3Client{
		createBucketResponse: &s3.CreateBucketOutput{},
		createBucketError:    nil,
	}

	storage := NewWithClient(mockClient)

	config := &services.BucketConfig{
		Name:   "test-bucket",
		Region: "us-east-1",
	}

	err := storage.CreateBucket(context.Background(), config)
	assert.NoError(t, err)
}

func TestAWSStorage_ListBuckets(t *testing.T) {
	mockClient := &mockS3Client{
		listBucketsResponse: &s3.ListBucketsOutput{
			Buckets: []types.Bucket{
				{Name: stringPtr("bucket1")},
				{Name: stringPtr("bucket2")},
			},
		},
		listBucketsError: nil,
	}

	storage := NewWithClient(mockClient)

	buckets, err := storage.ListBuckets(context.Background())
	assert.NoError(t, err)
	assert.Len(t, buckets, 2)
	assert.Equal(t, "bucket1", buckets[0])
	assert.Equal(t, "bucket2", buckets[1])
}

func stringPtr(s string) *string {
	return &s
}
