package mock

import (
	"context"
	"io"
	"strings"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

// MockStorage implements the services.Storage interface for testing.
// It provides configurable responses and error injection for all storage operations.
type MockStorage struct {
	provider *MockProvider
}

// CreateBucket creates a mock storage bucket with configurable responses.
// Returns an error if the bucket already exists or if configured to return an error.
//
// Error injection:
//   - Configure errors using WithError("CreateBucket", error)
//   - Automatically returns ErrResourceConflict if bucket already exists
//   - Common test scenarios: authentication, authorization, invalid names
//
// Example:
//
//	// Success scenario
//	err := mockStorage.CreateBucket(ctx, &services.BucketConfig{
//	    Name: "test-bucket",
//	    Region: "us-east-1",
//	})
//
//	// Error scenario (configured with WithError)
//	provider := mock.New("us-east-1").
//	    WithError("CreateBucket", cloudsdk.NewResourceConflictError(...))
func (m *MockStorage) CreateBucket(ctx context.Context, config *services.BucketConfig) error {
	m.provider.applyDelay("CreateBucket")

	if err := m.provider.checkError("CreateBucket"); err != nil {
		m.provider.recordOperation("CreateBucket", []interface{}{config}, nil, err)
		return err
	}

	// Check if bucket already exists
	if _, exists := m.provider.bucketState[config.Name]; exists {
		err := cloudsdk.NewCloudError(
			cloudsdk.ErrResourceConflict,
			"Bucket already exists",
			"mock", "storage", "CreateBucket",
		).WithSuggestions(
			"Choose a different bucket name",
			"Delete the existing bucket first",
		)
		m.provider.recordOperation("CreateBucket", []interface{}{config}, nil, err)
		return err
	}

	// Create bucket in state
	m.provider.bucketState[config.Name] = &BucketState{
		Name:    config.Name,
		Region:  config.Region,
		Objects: make(map[string][]byte),
		Tags:    config.Tags,
	}

	m.provider.recordOperation("CreateBucket", []interface{}{config}, nil, nil)
	return nil
}

// ListBuckets returns all mock bucket names from the current state.
// Returns an empty slice if no buckets exist.
//
// Error injection:
//   - Configure errors using WithError("ListBuckets", error)
//
// Example:
//
//	buckets, err := mockStorage.ListBuckets(ctx)
//	for _, bucket := range buckets {
//	    fmt.Printf("Bucket: %s\n", bucket)
//	}
func (m *MockStorage) ListBuckets(ctx context.Context) ([]string, error) {
	m.provider.applyDelay("ListBuckets")

	if err := m.provider.checkError("ListBuckets"); err != nil {
		m.provider.recordOperation("ListBuckets", []interface{}{}, nil, err)
		return nil, err
	}

	// Collect all bucket names from state
	buckets := make([]string, 0, len(m.provider.bucketState))
	for name := range m.provider.bucketState {
		buckets = append(buckets, name)
	}

	m.provider.recordOperation("ListBuckets", []interface{}{}, buckets, nil)
	return buckets, nil
}

// DeleteBucket removes a mock storage bucket from the state.
// Returns an error if the bucket doesn't exist or contains objects.
//
// Error injection:
//   - Configure errors using WithError("DeleteBucket", error)
//   - Automatically returns ErrResourceNotFound for non-existent buckets
//   - Returns ErrResourceConflict if bucket contains objects
//
// Example:
//
//	err := mockStorage.DeleteBucket(ctx, "test-bucket")
//	if err != nil {
//	    // Handle not found, not empty, or configured error
//	}
func (m *MockStorage) DeleteBucket(ctx context.Context, name string) error {
	m.provider.applyDelay("DeleteBucket")

	if err := m.provider.checkError("DeleteBucket"); err != nil {
		m.provider.recordOperation("DeleteBucket", []interface{}{name}, nil, err)
		return err
	}

	// Check if bucket exists
	bucket, exists := m.provider.bucketState[name]
	if !exists {
		err := cloudsdk.NewResourceNotFoundError("mock", "storage", "bucket", name)
		m.provider.recordOperation("DeleteBucket", []interface{}{name}, nil, err)
		return err
	}

	// Check if bucket is empty
	if len(bucket.Objects) > 0 {
		err := cloudsdk.NewCloudError(
			cloudsdk.ErrResourceConflict,
			"Bucket is not empty",
			"mock", "storage", "DeleteBucket",
		).WithSuggestions(
			"Delete all objects in the bucket first",
			"Use force delete if supported",
		)
		m.provider.recordOperation("DeleteBucket", []interface{}{name}, nil, err)
		return err
	}

	// Remove from state
	delete(m.provider.bucketState, name)

	m.provider.recordOperation("DeleteBucket", []interface{}{name}, nil, nil)
	return nil
}

// PutObject uploads mock data to a bucket and key.
// Stores the data in memory for later retrieval.
//
// Error injection:
//   - Configure errors using WithError("PutObject", error)
//   - Automatically returns ErrResourceNotFound if bucket doesn't exist
//
// Example:
//
//	data := strings.NewReader("Hello, World!")
//	err := mockStorage.PutObject(ctx, "test-bucket", "hello.txt", data)
//	if err != nil {
//	    // Handle bucket not found or configured error
//	}
func (m *MockStorage) PutObject(ctx context.Context, bucket, key string, data io.Reader) error {
	m.provider.applyDelay("PutObject")

	if err := m.provider.checkError("PutObject"); err != nil {
		m.provider.recordOperation("PutObject", []interface{}{bucket, key, data}, nil, err)
		return err
	}

	// Check if bucket exists
	bucketState, exists := m.provider.bucketState[bucket]
	if !exists {
		err := cloudsdk.NewResourceNotFoundError("mock", "storage", "bucket", bucket)
		m.provider.recordOperation("PutObject", []interface{}{bucket, key, data}, nil, err)
		return err
	}

	// Read data into memory
	dataBytes, err := io.ReadAll(data)
	if err != nil {
		m.provider.recordOperation("PutObject", []interface{}{bucket, key, data}, nil, err)
		return err
	}

	// Store object
	bucketState.Objects[key] = dataBytes

	m.provider.recordOperation("PutObject", []interface{}{bucket, key, data}, nil, nil)
	return nil
}

// GetObject retrieves mock data from a bucket and key.
// Returns the data as a ReadCloser.
//
// Error injection:
//   - Configure errors using WithError("GetObject", error)
//   - Automatically returns ErrResourceNotFound if bucket or object doesn't exist
//
// Example:
//
//	reader, err := mockStorage.GetObject(ctx, "test-bucket", "hello.txt")
//	if err != nil {
//	    // Handle not found or configured error
//	}
//	defer reader.Close()
//	data, _ := io.ReadAll(reader)
func (m *MockStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	m.provider.applyDelay("GetObject")

	if err := m.provider.checkError("GetObject"); err != nil {
		m.provider.recordOperation("GetObject", []interface{}{bucket, key}, nil, err)
		return nil, err
	}

	// Check if bucket exists
	bucketState, exists := m.provider.bucketState[bucket]
	if !exists {
		err := cloudsdk.NewResourceNotFoundError("mock", "storage", "bucket", bucket)
		m.provider.recordOperation("GetObject", []interface{}{bucket, key}, nil, err)
		return nil, err
	}

	// Check if object exists
	data, exists := bucketState.Objects[key]
	if !exists {
		err := cloudsdk.NewResourceNotFoundError("mock", "storage", "object", key)
		m.provider.recordOperation("GetObject", []interface{}{bucket, key}, nil, err)
		return nil, err
	}

	// Return data as ReadCloser
	reader := io.NopCloser(strings.NewReader(string(data)))
	m.provider.recordOperation("GetObject", []interface{}{bucket, key}, reader, nil)
	return reader, nil
}

// DeleteObject removes a mock object from a bucket.
// Returns an error if the bucket or object doesn't exist.
//
// Error injection:
//   - Configure errors using WithError("DeleteObject", error)
//   - Automatically returns ErrResourceNotFound if bucket or object doesn't exist
//
// Example:
//
//	err := mockStorage.DeleteObject(ctx, "test-bucket", "hello.txt")
//	if err != nil {
//	    // Handle not found or configured error
//	}
func (m *MockStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	m.provider.applyDelay("DeleteObject")

	if err := m.provider.checkError("DeleteObject"); err != nil {
		m.provider.recordOperation("DeleteObject", []interface{}{bucket, key}, nil, err)
		return err
	}

	// Check if bucket exists
	bucketState, exists := m.provider.bucketState[bucket]
	if !exists {
		err := cloudsdk.NewResourceNotFoundError("mock", "storage", "bucket", bucket)
		m.provider.recordOperation("DeleteObject", []interface{}{bucket, key}, nil, err)
		return err
	}

	// Check if object exists
	if _, exists := bucketState.Objects[key]; !exists {
		err := cloudsdk.NewResourceNotFoundError("mock", "storage", "object", key)
		m.provider.recordOperation("DeleteObject", []interface{}{bucket, key}, nil, err)
		return err
	}

	// Remove object
	delete(bucketState.Objects, key)

	m.provider.recordOperation("DeleteObject", []interface{}{bucket, key}, nil, nil)
	return nil
}

// ListObjects returns all mock objects in a bucket.
// Returns an empty slice if the bucket is empty.
//
// Error injection:
//   - Configure errors using WithError("ListObjects", error)
//   - Automatically returns ErrResourceNotFound if bucket doesn't exist
//
// Example:
//
//	objects, err := mockStorage.ListObjects(ctx, "test-bucket")
//	for _, obj := range objects {
//	    fmt.Printf("Object: %s (%d bytes)\n", obj.Key, obj.Size)
//	}
func (m *MockStorage) ListObjects(ctx context.Context, bucket string) ([]*services.Object, error) {
	m.provider.applyDelay("ListObjects")

	if err := m.provider.checkError("ListObjects"); err != nil {
		m.provider.recordOperation("ListObjects", []interface{}{bucket}, nil, err)
		return nil, err
	}

	// Check if bucket exists
	bucketState, exists := m.provider.bucketState[bucket]
	if !exists {
		err := cloudsdk.NewResourceNotFoundError("mock", "storage", "bucket", bucket)
		m.provider.recordOperation("ListObjects", []interface{}{bucket}, nil, err)
		return nil, err
	}

	// Collect all objects
	objects := make([]*services.Object, 0, len(bucketState.Objects))
	for key, data := range bucketState.Objects {
		obj := &services.Object{
			Key:          key,
			Size:         int64(len(data)),
			LastModified: "2024-01-01T00:00:00Z",                 // Mock timestamp
			ETag:         "\"d41d8cd98f00b204e9800998ecf8427e\"", // Mock ETag
		}
		objects = append(objects, obj)
	}

	m.provider.recordOperation("ListObjects", []interface{}{bucket}, objects, nil)
	return objects, nil
}
