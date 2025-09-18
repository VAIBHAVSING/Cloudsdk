package storage

import (
	"context"
	"io"
	"strings"
	"testing"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	cloudsdktesting "github.com/VAIBHAVSING/Cloudsdk/go/testing"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// mockS3Client is a mock implementation of the S3 client
type mockS3Client struct {
	createBucketResponse *s3.CreateBucketOutput
	createBucketError    error
	listBucketsResponse  *s3.ListBucketsOutput
	listBucketsError     error
	deleteBucketResponse *s3.DeleteBucketOutput
	deleteBucketError    error
	putObjectResponse    *s3.PutObjectOutput
	putObjectError       error
	getObjectResponse    *s3.GetObjectOutput
	getObjectError       error
	deleteObjectResponse *s3.DeleteObjectOutput
	deleteObjectError    error
	listObjectsResponse  *s3.ListObjectsV2Output
	listObjectsError     error
}

func (m *mockS3Client) CreateBucket(ctx context.Context, input *s3.CreateBucketInput, opts ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	return m.createBucketResponse, m.createBucketError
}

func (m *mockS3Client) ListBuckets(ctx context.Context, input *s3.ListBucketsInput, opts ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	return m.listBucketsResponse, m.listBucketsError
}

func (m *mockS3Client) DeleteBucket(ctx context.Context, input *s3.DeleteBucketInput, opts ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return m.deleteBucketResponse, m.deleteBucketError
}

func (m *mockS3Client) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.putObjectResponse, m.putObjectError
}

func (m *mockS3Client) GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.getObjectResponse, m.getObjectError
}

func (m *mockS3Client) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, opts ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return m.deleteObjectResponse, m.deleteObjectError
}

func (m *mockS3Client) ListObjectsV2(ctx context.Context, input *s3.ListObjectsV2Input, opts ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	return m.listObjectsResponse, m.listObjectsError
}

func (m *mockS3Client) PutBucketVersioning(ctx context.Context, input *s3.PutBucketVersioningInput, opts ...func(*s3.Options)) (*s3.PutBucketVersioningOutput, error) {
	return &s3.PutBucketVersioningOutput{}, nil
}

func TestAWSStorage_CreateBucket(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	mockClient := &mockS3Client{
		createBucketResponse: &s3.CreateBucketOutput{},
		createBucketError:    nil,
	}

	storage := NewWithClient(mockClient)
	config := cloudsdktesting.GenerateBucketConfig("test-bucket")

	err := storage.CreateBucket(context.Background(), config)
	helper.AssertNoError(err)
}

func TestAWSStorage_CreateBucket_ErrorScenarios(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	testCases := []struct {
		name      string
		mockError error
	}{
		{
			name:      "bucket already exists",
			mockError: &types.BucketAlreadyExists{},
		},
		{
			name:      "bucket already owned by you",
			mockError: &types.BucketAlreadyOwnedByYou{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockS3Client{
				createBucketError: tc.mockError,
			}

			storage := NewWithClient(mockClient)
			config := cloudsdktesting.GenerateBucketConfig("test-bucket")

			err := storage.CreateBucket(context.Background(), config)
			helper.AssertError(err)
		})
	}
}

func TestAWSStorage_ListBuckets(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

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
	helper.AssertNoError(err)
	helper.AssertEqual(2, len(buckets))
	helper.AssertEqual("bucket1", buckets[0])
	helper.AssertEqual("bucket2", buckets[1])
}

func TestAWSStorage_DeleteBucket(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	mockClient := &mockS3Client{
		deleteBucketResponse: &s3.DeleteBucketOutput{},
		deleteBucketError:    nil,
	}

	storage := NewWithClient(mockClient)

	err := storage.DeleteBucket(context.Background(), "test-bucket")
	helper.AssertNoError(err)
}

func TestAWSStorage_ObjectOperations(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	testData := "Hello, World!"
	mockClient := &mockS3Client{
		putObjectResponse: &s3.PutObjectOutput{},
		putObjectError:    nil,
		getObjectResponse: &s3.GetObjectOutput{
			Body: &mockReadCloser{data: testData},
		},
		getObjectError:       nil,
		deleteObjectResponse: &s3.DeleteObjectOutput{},
		deleteObjectError:    nil,
		listObjectsResponse: &s3.ListObjectsV2Output{
			Contents: []types.Object{
				{
					Key:  stringPtr("test-object.txt"),
					Size: int64Ptr(int64(len(testData))),
					ETag: stringPtr("\"d41d8cd98f00b204e9800998ecf8427e\""),
				},
			},
		},
		listObjectsError: nil,
	}

	storage := NewWithClient(mockClient)
	bucketName := "test-bucket"
	objectKey := "test-object.txt"

	// Test put object
	err := storage.PutObject(context.Background(), bucketName, objectKey, strings.NewReader(testData))
	helper.AssertNoError(err)

	// Test get object
	reader, err := storage.GetObject(context.Background(), bucketName, objectKey)
	helper.AssertNoError(err)
	helper.AssertNotEqual(nil, reader)
	reader.Close()

	// Test list objects
	objects, err := storage.ListObjects(context.Background(), bucketName)
	helper.AssertNoError(err)
	helper.AssertEqual(1, len(objects))
	helper.AssertEqual(objectKey, objects[0].Key)
	helper.AssertEqual(int64(len(testData)), objects[0].Size)

	// Test delete object
	err = storage.DeleteObject(context.Background(), bucketName, objectKey)
	helper.AssertNoError(err)
}

func TestAWSStorage_BucketLifecycle(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	mockClient := &mockS3Client{
		createBucketResponse: &s3.CreateBucketOutput{},
		createBucketError:    nil,
		listBucketsResponse: &s3.ListBucketsOutput{
			Buckets: []types.Bucket{
				{Name: stringPtr("test-bucket")},
			},
		},
		listBucketsError:     nil,
		deleteBucketResponse: &s3.DeleteBucketOutput{},
		deleteBucketError:    nil,
	}

	storage := NewWithClient(mockClient)
	bucketName := cloudsdktesting.GenerateBucketName("lifecycle-test")
	config := cloudsdktesting.GenerateBucketConfig(bucketName)

	// Create bucket
	err := storage.CreateBucket(context.Background(), config)
	helper.AssertNoError(err)

	// List buckets (should include our bucket)
	buckets, err := storage.ListBuckets(context.Background())
	helper.AssertNoError(err)
	helper.AssertEqual(1, len(buckets))

	// Delete bucket
	err = storage.DeleteBucket(context.Background(), bucketName)
	helper.AssertNoError(err)
}

func TestAWSStorage_ConcurrentOperations(t *testing.T) {
	mockClient := &mockS3Client{
		listBucketsResponse: &s3.ListBucketsOutput{
			Buckets: []types.Bucket{},
		},
		listBucketsError: nil,
	}

	storage := NewWithClient(mockClient)

	// Test concurrent ListBuckets calls
	cloudsdktesting.TestConcurrency(t, 10, func(id int) error {
		_, err := storage.ListBuckets(context.Background())
		return err
	})
}

func BenchmarkAWSStorage_ListBuckets(b *testing.B) {
	mockClient := &mockS3Client{
		listBucketsResponse: &s3.ListBucketsOutput{
			Buckets: []types.Bucket{},
		},
		listBucketsError: nil,
	}

	storage := NewWithClient(mockClient)

	cloudsdktesting.BenchmarkOperation(b, func() error {
		_, err := storage.ListBuckets(context.Background())
		return err
	})
}

func TestAWSStorage_WithMockProvider(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	// Test using our mock provider for comparison
	mockProvider := cloudsdktesting.NewMockProvider("us-east-1")
	client := cloudsdk.NewFromProvider(mockProvider)

	bucketName := "mock-test-bucket"
	config := cloudsdktesting.GenerateBucketConfig(bucketName)

	// Create bucket
	err := client.Storage().CreateBucket(context.Background(), config)
	helper.AssertNoError(err)

	// Verify mock provider recorded the operation
	cloudsdktesting.AssertProviderCalled(t, mockProvider, "CreateBucket", 1)

	// List buckets
	buckets, err := client.Storage().ListBuckets(context.Background())
	helper.AssertNoError(err)
	helper.AssertEqual(1, len(buckets))
	helper.AssertEqual(bucketName, buckets[0])
}

func TestAWSStorage_ErrorInjection(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	// Test error injection with mock provider
	mockProvider := cloudsdktesting.NewMockProvider("us-east-1").
		WithError("CreateBucket", cloudsdk.NewCloudError(
			cloudsdk.ErrResourceConflict,
			"Bucket already exists",
			"mock", "storage", "CreateBucket"))

	client := cloudsdk.NewFromProvider(mockProvider)
	config := cloudsdktesting.GenerateBucketConfig("error-test-bucket")

	err := client.Storage().CreateBucket(context.Background(), config)
	helper.AssertError(err)
	helper.AssertErrorCode(err, cloudsdk.ErrResourceConflict)
}

// mockReadCloser implements io.ReadCloser for testing
type mockReadCloser struct {
	data   string
	offset int
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.offset >= len(m.data) {
		return 0, io.EOF
	}

	remaining := len(m.data) - m.offset
	if len(p) > remaining {
		copy(p, m.data[m.offset:])
		m.offset = len(m.data)
		return remaining, nil
	}

	copy(p, m.data[m.offset:m.offset+len(p)])
	m.offset += len(p)
	return len(p), nil
}

func (m *mockReadCloser) Close() error {
	return nil
}

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
