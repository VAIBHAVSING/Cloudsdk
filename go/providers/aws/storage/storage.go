package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// RetryConfig defines retry behavior for AWS S3 operations
type RetryConfig struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultRetryConfig provides sensible defaults for retry behavior
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:   3,
	InitialDelay:  100 * time.Millisecond,
	MaxDelay:      5 * time.Second,
	BackoffFactor: 2.0,
}

// retryWithBackoff executes a function with exponential backoff retry logic
func retryWithBackoff(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		if attempt > 1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
		if !isRetryableS3Error(err) {
			return err
		}

		if attempt < config.MaxAttempts {
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
			log.Printf("AWS Storage: Retrying operation (attempt %d/%d) after %v due to: %v",
				attempt+1, config.MaxAttempts, delay, err)
		}
	}

	return lastErr
}

// isRetryableS3Error determines if an S3 error should be retried
func isRetryableS3Error(err error) bool {
	if err == nil {
		return false
	}

	// Check for context cancellation/timeout - don't retry
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for AWS-specific retryable errors
	var ae smithy.APIError
	if errors.As(err, &ae) {
		code := ae.ErrorCode()
		switch code {
		case "Throttling", "ThrottlingException", "RequestLimitExceeded", "SlowDown":
			return true
		case "InternalError", "InternalFailure", "ServiceUnavailable":
			return true
		case "RequestTimeout", "RequestTimeoutException":
			return true
		}
	}

	// Check error message for common transient issues
	errMsg := strings.ToLower(err.Error())
	transientMessages := []string{
		"connection reset",
		"connection refused",
		"timeout",
		"temporary failure",
		"service unavailable",
		"internal error",
	}

	for _, msg := range transientMessages {
		if strings.Contains(errMsg, msg) {
			return true
		}
	}

	return false
}

// wrapS3Error converts S3 errors to CloudError with helpful context
func wrapS3Error(err error, provider, service, operation string) error {
	if err == nil {
		return nil
	}

	// Handle context errors
	if errors.Is(err, context.Canceled) {
		return cloudsdk.NewCloudError(cloudsdk.ErrNetworkTimeout, "Operation was cancelled", provider, service, operation).
			WithCause(err).
			WithSuggestions("Check if the operation timeout is sufficient", "Verify network connectivity")
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return cloudsdk.NewCloudError(cloudsdk.ErrNetworkTimeout, "Operation timed out", provider, service, operation).
			WithCause(err).
			WithSuggestions("Increase the operation timeout", "Check network connectivity", "Verify AWS S3 service status")
	}

	// Handle AWS-specific errors
	var ae smithy.APIError
	if errors.As(err, &ae) {
		code := ae.ErrorCode()
		message := ae.ErrorMessage()

		switch code {
		case "AccessDenied":
			return cloudsdk.NewAuthorizationError(provider, service, operation, err).
				WithSuggestions(
					"Verify your IAM user/role has the required S3 permissions",
					"Check bucket policies and ACLs",
					"Ensure you're operating in the correct AWS region",
				)

		case "InvalidAccessKeyId", "SignatureDoesNotMatch", "TokenRefreshRequired":
			return cloudsdk.NewAuthenticationError(provider, err).
				WithSuggestions(
					"Verify your AWS access key and secret key are correct",
					"Check if your credentials have expired",
					"Ensure your system clock is synchronized",
				)

		case "NoSuchBucket":
			bucketName := extractBucketNameFromError(message)
			return cloudsdk.NewResourceNotFoundError(provider, service, "bucket", bucketName).
				WithSuggestions(
					"Verify the bucket name is correct",
					"Check that the bucket exists in the specified region",
					"Ensure you have permission to access this bucket",
				)

		case "NoSuchKey":
			return cloudsdk.NewResourceNotFoundError(provider, service, "object", "unknown").
				WithSuggestions(
					"Verify the object key is correct",
					"Check that the object exists in the bucket",
					"Ensure the object hasn't been deleted",
				)

		case "BucketAlreadyExists", "BucketAlreadyOwnedByYou":
			return cloudsdk.NewCloudError(cloudsdk.ErrResourceConflict, "Bucket already exists", provider, service, operation).
				WithCause(err).
				WithSuggestions(
					"Choose a different bucket name (bucket names must be globally unique)",
					"Check if you already own this bucket",
					"Use a more specific naming convention",
				)

		case "InvalidBucketName":
			return cloudsdk.NewInvalidConfigError(provider, service, "BucketName", "Invalid bucket name").
				WithSuggestions(
					"Bucket names must be 3-63 characters long",
					"Use only lowercase letters, numbers, and hyphens",
					"Don't use periods or underscores",
					"Start and end with a letter or number",
				)

		case "SlowDown", "RequestLimitExceeded":
			return cloudsdk.NewRateLimitError(provider, service, operation, 0).
				WithCause(err).
				WithSuggestions(
					"Reduce the frequency of API calls",
					"Implement exponential backoff (this is done automatically)",
					"Consider using S3 Transfer Acceleration",
				)

		case "EntityTooLarge":
			return cloudsdk.NewInvalidConfigError(provider, service, "ObjectSize", "Object too large").
				WithSuggestions(
					"Use multipart upload for objects larger than 5GB",
					"Check the maximum object size limits",
					"Consider breaking large objects into smaller parts",
				)

		case "InvalidStorageClass":
			return cloudsdk.NewInvalidConfigError(provider, service, "StorageClass", "Invalid storage class").
				WithSuggestions(
					"Use valid storage classes: STANDARD, REDUCED_REDUNDANCY, STANDARD_IA, ONEZONE_IA, INTELLIGENT_TIERING, GLACIER, DEEP_ARCHIVE",
					"Check if the storage class is available in your region",
				)

		default:
			// Generic AWS error
			return cloudsdk.NewCloudError(cloudsdk.ErrProviderError, fmt.Sprintf("S3 error: %s", message), provider, service, operation).
				WithCause(err).
				WithSuggestions(
					"Check AWS S3 service status for any ongoing issues",
					"Verify your request parameters are valid",
					"Contact AWS Support if the issue persists",
				)
		}
	}

	// Generic error fallback
	return cloudsdk.NewCloudError(cloudsdk.ErrProviderError, "Unexpected error occurred", provider, service, operation).
		WithCause(err).
		WithSuggestions(
			"Check the underlying error for more details",
			"Verify your AWS configuration is correct",
			"Try the operation again",
		)
}

// extractBucketNameFromError attempts to extract bucket name from error messages
func extractBucketNameFromError(message string) string {
	// Look for bucket name in common error message patterns
	if strings.Contains(message, "bucket") {
		parts := strings.Fields(message)
		for i, part := range parts {
			if part == "bucket" && i+1 < len(parts) {
				return strings.Trim(parts[i+1], "',.")
			}
		}
	}
	return "unknown"
}

// logRequest logs S3 API requests for debugging (when debug is enabled)
func logRequest(operation string, input interface{}, debug bool) {
	if debug {
		log.Printf("AWS Storage: %s request: %+v", operation, input)
	}
}

// logResponse logs S3 API responses for debugging (when debug is enabled)
func logResponse(operation string, output interface{}, err error, debug bool) {
	if debug {
		if err != nil {
			log.Printf("AWS Storage: %s error: %v", operation, err)
		} else {
			log.Printf("AWS Storage: %s response: %+v", operation, output)
		}
	}
}

// S3ClientInterface defines methods we need from S3 client for testing
type S3ClientInterface interface {
	CreateBucket(ctx context.Context, input *s3.CreateBucketInput, opts ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	ListBuckets(ctx context.Context, input *s3.ListBucketsInput, opts ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	DeleteBucket(ctx context.Context, input *s3.DeleteBucketInput, opts ...func(*s3.Options)) (*s3.DeleteBucketOutput, error)
	PutBucketVersioning(ctx context.Context, input *s3.PutBucketVersioningInput, opts ...func(*s3.Options)) (*s3.PutBucketVersioningOutput, error)
	PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, opts ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	ListObjectsV2(ctx context.Context, input *s3.ListObjectsV2Input, opts ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// AWSStorage implements the Storage interface for AWS
type AWSStorage struct {
	client      S3ClientInterface
	debug       bool
	retryConfig RetryConfig
}

// New creates a new AWSStorage instance with real AWS client
func New(cfg aws.Config) services.Storage {
	client := s3.NewFromConfig(cfg)
	return &AWSStorage{
		client:      client,
		debug:       false,
		retryConfig: DefaultRetryConfig,
	}
}

// NewWithClient creates a new AWSStorage instance with custom client (for testing)
func NewWithClient(client S3ClientInterface) services.Storage {
	return &AWSStorage{
		client:      client,
		debug:       false,
		retryConfig: DefaultRetryConfig,
	}
}

// NewWithOptions creates a new AWSStorage instance with custom options
func NewWithOptions(cfg aws.Config, debug bool, retryConfig *RetryConfig) services.Storage {
	client := s3.NewFromConfig(cfg)

	finalRetryConfig := DefaultRetryConfig
	if retryConfig != nil {
		finalRetryConfig = *retryConfig
	}

	return &AWSStorage{
		client:      client,
		debug:       debug,
		retryConfig: finalRetryConfig,
	}
}

// CreateBucket creates a new S3 bucket
func (s *AWSStorage) CreateBucket(ctx context.Context, config *services.BucketConfig) error {
	// Validate input configuration
	if config == nil {
		return cloudsdk.NewInvalidConfigError("aws", "storage", "config", "bucket configuration cannot be nil")
	}
	if config.Name == "" {
		return cloudsdk.NewInvalidConfigError("aws", "storage", "Name", "bucket name is required")
	}

	// Validate bucket name format
	if err := validateBucketName(config.Name); err != nil {
		return cloudsdk.NewInvalidConfigError("aws", "storage", "Name", err.Error()).
			WithSuggestions(
				"Bucket names must be 3-63 characters long",
				"Use only lowercase letters, numbers, and hyphens",
				"Start and end with a letter or number",
				"Don't use periods or underscores",
			)
	}

	input := &s3.CreateBucketInput{
		Bucket: aws.String(config.Name),
	}

	// Add region configuration if specified and not us-east-1 (default)
	if config.Region != "" && config.Region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(config.Region),
		}
	}

	logRequest("CreateBucket", input, s.debug)

	var resp *s3.CreateBucketOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, s.retryConfig, func() error {
		resp, err = s.client.CreateBucket(ctx, input)
		return err
	})

	logResponse("CreateBucket", resp, retryErr, s.debug)

	if retryErr != nil {
		return wrapS3Error(retryErr, "aws", "storage", "CreateBucket")
	}

	// Enable versioning if requested
	if config.Versioning != nil && *config.Versioning {
		versioningInput := &s3.PutBucketVersioningInput{
			Bucket: aws.String(config.Name),
			VersioningConfiguration: &types.VersioningConfiguration{
				Status: types.BucketVersioningStatusEnabled,
			},
		}

		logRequest("PutBucketVersioning", versioningInput, s.debug)

		var versioningResp *s3.PutBucketVersioningOutput
		versioningRetryErr := retryWithBackoff(ctx, s.retryConfig, func() error {
			versioningResp, err = s.client.PutBucketVersioning(ctx, versioningInput)
			return err
		})

		logResponse("PutBucketVersioning", versioningResp, versioningRetryErr, s.debug)

		if versioningRetryErr != nil {
			log.Printf("AWS Storage: Warning - failed to enable versioning for bucket %s: %v", config.Name, versioningRetryErr)
			// Don't fail the entire operation if versioning fails
		}
	}

	return nil
}

// validateBucketName validates S3 bucket naming rules
func validateBucketName(name string) error {
	if len(name) < 3 || len(name) > 63 {
		return fmt.Errorf("bucket name must be between 3 and 63 characters")
	}

	// Check for valid characters and format
	for i, r := range name {
		if i == 0 || i == len(name)-1 {
			// First and last character must be alphanumeric
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
				return fmt.Errorf("bucket name must start and end with a letter or number")
			}
		} else {
			// Middle characters can be alphanumeric or hyphen
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
				return fmt.Errorf("bucket name can only contain lowercase letters, numbers, and hyphens")
			}
		}
	}

	// Check for consecutive hyphens or periods
	if strings.Contains(name, "--") || strings.Contains(name, "..") {
		return fmt.Errorf("bucket name cannot contain consecutive hyphens or periods")
	}

	// Check for IP address format
	if strings.Count(name, ".") == 3 {
		parts := strings.Split(name, ".")
		if len(parts) == 4 {
			allNumeric := true
			for _, part := range parts {
				if len(part) == 0 || len(part) > 3 {
					allNumeric = false
					break
				}
				for _, r := range part {
					if r < '0' || r > '9' {
						allNumeric = false
						break
					}
				}
				if !allNumeric {
					break
				}
			}
			if allNumeric {
				return fmt.Errorf("bucket name cannot be formatted as an IP address")
			}
		}
	}

	return nil
}

// ListBuckets lists all S3 buckets
func (s *AWSStorage) ListBuckets(ctx context.Context) ([]string, error) {
	input := &s3.ListBucketsInput{}

	logRequest("ListBuckets", input, s.debug)

	var resp *s3.ListBucketsOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, s.retryConfig, func() error {
		resp, err = s.client.ListBuckets(ctx, input)
		return err
	})

	logResponse("ListBuckets", resp, retryErr, s.debug)

	if retryErr != nil {
		return nil, wrapS3Error(retryErr, "aws", "storage", "ListBuckets")
	}

	buckets := make([]string, len(resp.Buckets))
	for i, b := range resp.Buckets {
		buckets[i] = aws.ToString(b.Name)
	}

	return buckets, nil
}

// DeleteBucket deletes an S3 bucket
func (s *AWSStorage) DeleteBucket(ctx context.Context, name string) error {
	// Validate input
	if name == "" {
		return cloudsdk.NewInvalidConfigError("aws", "storage", "name", "bucket name cannot be empty")
	}

	input := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}

	logRequest("DeleteBucket", input, s.debug)

	var resp *s3.DeleteBucketOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, s.retryConfig, func() error {
		resp, err = s.client.DeleteBucket(ctx, input)
		return err
	})

	logResponse("DeleteBucket", resp, retryErr, s.debug)

	if retryErr != nil {
		return wrapS3Error(retryErr, "aws", "storage", "DeleteBucket")
	}

	return nil
}

func (s *AWSStorage) PutObject(ctx context.Context, bucket, key string, body io.Reader) error {
	// Validate input
	if bucket == "" {
		return cloudsdk.NewInvalidConfigError("aws", "storage", "bucket", "bucket name cannot be empty")
	}
	if key == "" {
		return cloudsdk.NewInvalidConfigError("aws", "storage", "key", "object key cannot be empty")
	}
	if body == nil {
		return cloudsdk.NewInvalidConfigError("aws", "storage", "body", "object body cannot be nil")
	}

	// Wrap the reader with progress tracking if it's a large upload
	var wrappedBody io.Reader = body
	if s.debug {
		wrappedBody = &progressReader{
			reader: body,
			onProgress: func(bytesRead int64) {
				log.Printf("AWS Storage: PutObject progress - uploaded %d bytes to %s/%s", bytesRead, bucket, key)
			},
		}
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   wrappedBody,
	}

	logRequest("PutObject", map[string]interface{}{
		"Bucket": bucket,
		"Key":    key,
		"Body":   "[BINARY DATA]",
	}, s.debug)

	var resp *s3.PutObjectOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, s.retryConfig, func() error {
		resp, err = s.client.PutObject(ctx, input)
		return err
	})

	logResponse("PutObject", resp, retryErr, s.debug)

	if retryErr != nil {
		return wrapS3Error(retryErr, "aws", "storage", "PutObject")
	}

	return nil
}

// progressReader wraps an io.Reader to track upload progress
type progressReader struct {
	reader     io.Reader
	bytesRead  int64
	onProgress func(int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.bytesRead += int64(n)
	if pr.onProgress != nil && n > 0 {
		pr.onProgress(pr.bytesRead)
	}
	return n, err
}

func (s *AWSStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	// Validate input
	if bucket == "" {
		return nil, cloudsdk.NewInvalidConfigError("aws", "storage", "bucket", "bucket name cannot be empty")
	}
	if key == "" {
		return nil, cloudsdk.NewInvalidConfigError("aws", "storage", "key", "object key cannot be empty")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	logRequest("GetObject", input, s.debug)

	var resp *s3.GetObjectOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, s.retryConfig, func() error {
		resp, err = s.client.GetObject(ctx, input)
		return err
	})

	logResponse("GetObject", map[string]interface{}{
		"ContentLength": resp.ContentLength,
		"ContentType":   resp.ContentType,
		"LastModified":  resp.LastModified,
	}, retryErr, s.debug)

	if retryErr != nil {
		return nil, wrapS3Error(retryErr, "aws", "storage", "GetObject")
	}

	return resp.Body, nil
}

func (s *AWSStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	// Validate input
	if bucket == "" {
		return cloudsdk.NewInvalidConfigError("aws", "storage", "bucket", "bucket name cannot be empty")
	}
	if key == "" {
		return cloudsdk.NewInvalidConfigError("aws", "storage", "key", "object key cannot be empty")
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	logRequest("DeleteObject", input, s.debug)

	var resp *s3.DeleteObjectOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, s.retryConfig, func() error {
		resp, err = s.client.DeleteObject(ctx, input)
		return err
	})

	logResponse("DeleteObject", resp, retryErr, s.debug)

	if retryErr != nil {
		return wrapS3Error(retryErr, "aws", "storage", "DeleteObject")
	}

	return nil
}

func (s *AWSStorage) ListObjects(ctx context.Context, bucket string) ([]*services.Object, error) {
	// Validate input
	if bucket == "" {
		return nil, cloudsdk.NewInvalidConfigError("aws", "storage", "bucket", "bucket name cannot be empty")
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	logRequest("ListObjectsV2", input, s.debug)

	var resp *s3.ListObjectsV2Output
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, s.retryConfig, func() error {
		resp, err = s.client.ListObjectsV2(ctx, input)
		return err
	})

	logResponse("ListObjectsV2", resp, retryErr, s.debug)

	if retryErr != nil {
		return nil, wrapS3Error(retryErr, "aws", "storage", "ListObjects")
	}

	objects := make([]*services.Object, len(resp.Contents))
	for i, obj := range resp.Contents {
		object := &services.Object{
			Key:  aws.ToString(obj.Key),
			Size: aws.ToInt64(obj.Size),
		}

		// Safely handle optional fields
		if obj.LastModified != nil {
			object.LastModified = obj.LastModified.String()
		}
		if obj.ETag != nil {
			object.ETag = aws.ToString(obj.ETag)
		}
		// Note: StorageClass field not available in current Object struct

		objects[i] = object
	}

	return objects, nil
}
