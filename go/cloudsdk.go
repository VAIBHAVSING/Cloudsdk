package cloudsdk

import (
	"fmt"
	"time"

	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

// Config holds the configuration for the Cloud SDK
type Config struct {
	Region string
	// Add other global configs like credentials, etc.
}

// ServiceType represents the type of cloud service
type ServiceType string

const (
	ServiceCompute  ServiceType = "compute"
	ServiceStorage  ServiceType = "storage"
	ServiceDatabase ServiceType = "database"
)

// ErrorCode represents standardized error types across all providers
type ErrorCode string

const (
	// Authentication and authorization errors
	ErrAuthentication ErrorCode = "AUTHENTICATION_FAILED"
	ErrAuthorization  ErrorCode = "AUTHORIZATION_FAILED"

	// Service availability errors
	ErrServiceNotSupported   ErrorCode = "SERVICE_NOT_SUPPORTED"
	ErrOperationNotSupported ErrorCode = "OPERATION_NOT_SUPPORTED"

	// Resource errors
	ErrResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"
	ErrResourceConflict ErrorCode = "RESOURCE_CONFLICT"

	// Network and rate limiting
	ErrRateLimit      ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrNetworkTimeout ErrorCode = "NETWORK_TIMEOUT"

	// Configuration errors
	ErrInvalidConfig ErrorCode = "INVALID_CONFIGURATION"
	ErrProviderError ErrorCode = "PROVIDER_ERROR"
)

// ErrorContext provides debugging information for troubleshooting
type ErrorContext struct {
	RequestID string            `json:"request_id,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Retryable bool              `json:"retryable"`
}

// CloudError provides structured error information with helpful context and suggestions.
// All SDK errors implement this interface for consistent error handling.
type CloudError struct {
	Code        ErrorCode    `json:"code"`
	Message     string       `json:"message"`
	Provider    string       `json:"provider"`
	Service     string       `json:"service"`
	Operation   string       `json:"operation"`
	Suggestions []string     `json:"suggestions,omitempty"`
	Cause       error        `json:"-"`
	Context     ErrorContext `json:"context,omitempty"`
}

// Error implements the error interface with rich context
func (e *CloudError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Code, e.Message)
	if e.Provider != "" {
		msg += fmt.Sprintf(" (provider: %s)", e.Provider)
	}
	if len(e.Suggestions) > 0 {
		msg += fmt.Sprintf("\nSuggestions:\n")
		for _, suggestion := range e.Suggestions {
			msg += fmt.Sprintf("  - %s\n", suggestion)
		}
	}
	return msg
}

// Unwrap returns the underlying cause error for error wrapping
func (e *CloudError) Unwrap() error {
	return e.Cause
}

// ServiceNotSupportedError is returned when a provider doesn't support a requested service.
// This helps developers understand provider limitations at development time.
type ServiceNotSupportedError struct {
	Provider string
	Service  ServiceType
}

// Error implements the error interface with helpful messaging
func (e *ServiceNotSupportedError) Error() string {
	return fmt.Sprintf("provider '%s' does not support '%s' service. "+
		"Please check provider documentation for supported services or switch to a compatible provider",
		e.Provider, e.Service)
}

// NewServiceNotSupportedError creates a new service not supported error with suggestions
func NewServiceNotSupportedError(provider string, service ServiceType) *ServiceNotSupportedError {
	return &ServiceNotSupportedError{
		Provider: provider,
		Service:  service,
	}
}

// NewCloudError creates a new CloudError with the specified parameters
func NewCloudError(code ErrorCode, message string, provider string, service string, operation string) *CloudError {
	return &CloudError{
		Code:      code,
		Message:   message,
		Provider:  provider,
		Service:   service,
		Operation: operation,
		Context: ErrorContext{
			Timestamp: time.Now(),
			Retryable: isRetryableError(code),
		},
	}
}

// WithSuggestions adds helpful suggestions to a CloudError
func (e *CloudError) WithSuggestions(suggestions ...string) *CloudError {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

// WithCause adds the underlying cause error
func (e *CloudError) WithCause(cause error) *CloudError {
	e.Cause = cause
	return e
}

// WithContext adds debugging context
func (e *CloudError) WithContext(requestID string, metadata map[string]string) *CloudError {
	e.Context.RequestID = requestID
	if e.Context.Metadata == nil {
		e.Context.Metadata = make(map[string]string)
	}
	for k, v := range metadata {
		e.Context.Metadata[k] = v
	}
	return e
}

// isRetryableError determines if an error code represents a retryable condition
func isRetryableError(code ErrorCode) bool {
	switch code {
	case ErrRateLimit, ErrNetworkTimeout:
		return true
	default:
		return false
	}
}

// Helper functions for common error scenarios

// NewAuthenticationError creates a new authentication error with helpful suggestions
func NewAuthenticationError(provider string, cause error) *CloudError {
	return NewCloudError(ErrAuthentication, "Authentication failed", provider, "", "authenticate").
		WithCause(cause).
		WithSuggestions(
			"Check your credentials are correctly configured",
			"Verify your access keys are not expired",
			"Ensure your credentials have the necessary permissions",
			"Try refreshing your authentication tokens",
		)
}

// NewAuthorizationError creates a new authorization error with helpful suggestions
func NewAuthorizationError(provider string, service string, operation string, cause error) *CloudError {
	return NewCloudError(ErrAuthorization, "Authorization failed", provider, service, operation).
		WithCause(cause).
		WithSuggestions(
			"Check that your credentials have the required permissions",
			"Verify your IAM policies allow this operation",
			"Ensure you're operating in the correct region/account",
		)
}

// NewResourceNotFoundError creates a new resource not found error
func NewResourceNotFoundError(provider string, service string, resourceType string, resourceID string) *CloudError {
	message := fmt.Sprintf("%s '%s' not found", resourceType, resourceID)
	return NewCloudError(ErrResourceNotFound, message, provider, service, "get").
		WithSuggestions(
			"Verify the resource ID is correct",
			"Check that the resource exists in the specified region",
			"Ensure you have permission to access this resource",
		)
}

// NewInvalidConfigError creates a new invalid configuration error
func NewInvalidConfigError(provider string, service string, field string, reason string) *CloudError {
	message := fmt.Sprintf("Invalid configuration for field '%s': %s", field, reason)
	return NewCloudError(ErrInvalidConfig, message, provider, service, "validate").
		WithSuggestions(
			"Check the field value meets the required format",
			"Refer to the provider documentation for valid values",
			"Ensure all required fields are provided",
		)
}

// NewRateLimitError creates a new rate limit error with retry suggestions
func NewRateLimitError(provider string, service string, operation string, retryAfter time.Duration) *CloudError {
	message := "Rate limit exceeded"
	suggestions := []string{
		"Reduce the frequency of API calls",
		"Implement exponential backoff in your retry logic",
	}
	if retryAfter > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Retry after %v", retryAfter))
	}

	return NewCloudError(ErrRateLimit, message, provider, service, operation).
		WithSuggestions(suggestions...)
}

// Provider defines the interface for cloud providers.
// Each provider must implement all services, but can return errors for unsupported ones.
type Provider interface {
	// Compute returns the compute service for managing virtual machines.
	// Returns ErrServiceNotSupported if the provider doesn't support compute operations.
	Compute() services.Compute

	// Storage returns the storage service for managing buckets and objects.
	// Returns ErrServiceNotSupported if the provider doesn't support storage operations.
	Storage() services.Storage

	// Database returns the database service for managing database instances.
	// Returns ErrServiceNotSupported if the provider doesn't support database operations.
	Database() services.Database

	// Name returns the provider name (e.g., "aws", "gcp", "azure").
	Name() string

	// Region returns the configured region for this provider.
	Region() string

	// SupportedServices returns a list of services supported by this provider.
	// This allows compile-time checking of service availability.
	SupportedServices() []ServiceType
}

// Client provides a unified interface to cloud providers.
// It automatically validates service availability at runtime.
type Client struct {
	provider Provider
}

// New creates a new cloud SDK client with the specified provider.
// The client will validate service availability when methods are called.
//
// Example:
//
//	provider := aws.New("us-east-1")
//	client := cloudsdk.New(provider)
func New(provider Provider, config *Config) *Client {
	return &Client{provider: provider}
}

// NewFromProvider creates a new Cloud SDK client from a provider (auto-configures from provider)
func NewFromProvider(provider Provider) *Client {
	return &Client{provider: provider}
}

// Compute returns the compute service if supported by the provider.
// Panics with ErrServiceNotSupported if compute is not available.
//
// Example:
//
//	vm, err := client.Compute().CreateVM(ctx, &services.VMConfig{
//	    Name: "my-server",
//	    ImageID: "ami-12345",
//	    InstanceType: "t2.micro",
//	})
func (c *Client) Compute() services.Compute {
	if !c.supportsService(ServiceCompute) {
		panic(NewServiceNotSupportedError(c.provider.Name(), ServiceCompute))
	}
	return c.provider.Compute()
}

// Storage returns the storage service if supported by the provider.
// Panics with ErrServiceNotSupported if storage is not available.
//
// Example:
//
//	err := client.Storage().CreateBucket(ctx, &services.BucketConfig{
//	    Name: "my-bucket",
//	    Region: "us-east-1",
//	})
func (c *Client) Storage() services.Storage {
	if !c.supportsService(ServiceStorage) {
		panic(NewServiceNotSupportedError(c.provider.Name(), ServiceStorage))
	}
	return c.provider.Storage()
}

// Database returns the database service if supported by the provider.
// Panics with ErrServiceNotSupported if database is not available.
//
// Example:
//
//	db, err := client.Database().CreateDB(ctx, &services.DBConfig{
//	    Name: "my-database",
//	    Engine: "postgres",
//	    InstanceType: "db.t3.micro",
//	})
func (c *Client) Database() services.Database {
	if !c.supportsService(ServiceDatabase) {
		panic(NewServiceNotSupportedError(c.provider.Name(), ServiceDatabase))
	}
	return c.provider.Database()
}

// supportsService checks if the provider supports the given service type
func (c *Client) supportsService(service ServiceType) bool {
	supported := c.provider.SupportedServices()
	for _, s := range supported {
		if s == service {
			return true
		}
	}
	return false
}
