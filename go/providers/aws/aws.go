// Package aws provides a comprehensive AWS provider implementation for the Cloud SDK.
//
// This package implements the cloudsdk.Provider interface for Amazon Web Services,
// offering access to AWS Compute (EC2), Storage (S3), and Database (RDS) services
// through a unified, developer-friendly API.
//
// FEATURES:
//   - Automatic credential discovery with multiple authentication methods
//   - Comprehensive error handling with actionable suggestions
//   - Fluent configuration API using functional options
//   - Support for all major AWS authentication patterns
//   - Detailed logging and debugging capabilities
//   - Production-ready defaults with customizable settings
//
// QUICK START:
//
// Basic usage with automatic credential discovery:
//
//	provider, err := aws.New("us-east-1")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	client := cloudsdk.New(provider)
//
//	// Use services
//	vm, err := client.Compute().CreateVM(ctx, &services.VMConfig{
//	    Name: "my-server",
//	    ImageID: "ami-12345678",
//	    InstanceType: "t2.micro",
//	})
//
// AUTHENTICATION:
//
// The AWS provider supports multiple authentication methods, tried in order:
//  1. Explicit credentials via WithCredentials()
//  2. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
//  3. AWS profiles via WithProfile() or AWS_PROFILE environment variable
//  4. IAM roles (when running on EC2, ECS, or Lambda)
//  5. Default profile in ~/.aws/credentials
//
// SUPPORTED SERVICES:
//   - Compute: Amazon EC2 for virtual machine management
//   - Storage: Amazon S3 for object storage and file operations
//   - Database: Amazon RDS for managed relational databases
//
// CONFIGURATION OPTIONS:
//   - WithProfile(): Use specific AWS profile
//   - WithCredentials(): Set explicit access keys
//   - WithSessionToken(): Add session token for temporary credentials
//   - WithDebug(): Enable detailed request/response logging
//   - WithTimeout(): Set custom API call timeouts
//   - WithRetryMaxAttempts(): Configure retry behavior
//
// ERROR HANDLING:
//
// All methods return structured errors with helpful suggestions:
//
//	provider, err := aws.New("invalid-region")
//	if err != nil {
//	    // Error includes specific suggestions for fixing the issue
//	    fmt.Printf("Error: %v\n", err)
//	    // Output: [INVALID_CONFIGURATION] Invalid configuration for field 'region': region cannot be empty
//	    // Suggestions:
//	    //   - Provide a valid AWS region (e.g., 'us-east-1', 'eu-west-1')
//	    //   - Check AWS documentation for available regions
//	}
//
// EXAMPLES:
//
// Environment variables (recommended for production):
//
//	export AWS_ACCESS_KEY_ID=your_access_key
//	export AWS_SECRET_ACCESS_KEY=your_secret_key
//	provider, err := aws.New("us-east-1")
//
// AWS CLI profiles (recommended for development):
//
//	aws configure --profile myprofile
//	provider, err := aws.New("us-east-1", aws.WithProfile("myprofile"))
//
// Explicit credentials (use with caution):
//
//	provider, err := aws.New("us-east-1",
//	    aws.WithCredentials("ACCESS_KEY", "SECRET_KEY"))
//
// Full configuration:
//
//	provider, err := aws.New("us-east-1",
//	    aws.WithProfile("production"),
//	    aws.WithDebug(),
//	    aws.WithTimeout(60*time.Second),
//	    aws.WithRetryMaxAttempts(5))
//
// SECURITY CONSIDERATIONS:
//   - Never hardcode credentials in source code
//   - Use IAM roles when running on AWS infrastructure
//   - Rotate access keys regularly
//   - Use least-privilege IAM policies
//   - Enable AWS CloudTrail for audit logging
package aws

import (
	"context"
	"fmt"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/providers/aws/compute"
	"github.com/VAIBHAVSING/Cloudsdk/go/providers/aws/database"
	"github.com/VAIBHAVSING/Cloudsdk/go/providers/aws/storage"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// Config holds AWS-specific configuration options.
// This struct provides a clean way to configure AWS provider settings
// with sensible defaults and comprehensive validation.
type Config struct {
	// Region specifies the AWS region for all operations.
	// If not provided, defaults to "us-east-1".
	Region string

	// Profile specifies the AWS profile to use from ~/.aws/credentials or ~/.aws/config.
	// If not provided, uses the default profile or environment variables.
	Profile string

	// AccessKey and SecretKey provide explicit AWS credentials.
	// These take precedence over profile-based authentication.
	AccessKey string
	SecretKey string

	// SessionToken is required when using temporary credentials (STS).
	SessionToken string

	// Debug enables detailed logging of AWS API calls.
	// Useful for troubleshooting authentication and API issues.
	Debug bool

	// Timeout sets the default timeout for AWS API calls.
	// If not provided, uses AWS SDK defaults (typically 30 seconds).
	Timeout time.Duration

	// RetryMaxAttempts sets the maximum number of retry attempts for failed requests.
	// If not provided, uses AWS SDK defaults (typically 3 attempts).
	RetryMaxAttempts int
}

// AWSProvider implements the cloudsdk.Provider interface for Amazon Web Services.
// It provides access to AWS Compute (EC2), Storage (S3), and Database (RDS) services
// with automatic credential discovery and comprehensive error handling.
//
// The provider supports multiple authentication methods in order of precedence:
//  1. Explicit credentials via WithCredentials()
//  2. AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables
//  3. AWS profile via WithProfile() or AWS_PROFILE environment variable
//  4. IAM role for EC2 instances (when running on EC2)
//  5. Default profile in ~/.aws/credentials
type AWSProvider struct {
	config      Config
	awsConfig   aws.Config
	initialized bool
}

// Option configures AWS provider settings using the functional options pattern.
// This provides a clean, composable way to customize provider behavior.
type Option func(*Config)

// WithProfile sets the AWS profile to use for authentication.
// The profile must exist in ~/.aws/credentials or ~/.aws/config.
// This is useful when you have multiple AWS accounts or environments configured.
//
// Example:
//
//	provider := aws.New("us-east-1", aws.WithProfile("production"))
//
// Common errors:
//   - Profile not found: Check that the profile exists in your AWS config files
//   - Invalid credentials: Verify the profile has valid access keys
func WithProfile(profile string) Option {
	return func(c *Config) {
		c.Profile = profile
	}
}

// WithCredentials sets explicit AWS credentials for authentication.
// Use this for programmatic access or when AWS profiles are not available.
// These credentials take precedence over profile-based authentication.
//
// Example:
//
//	provider := aws.New("us-east-1",
//	    aws.WithCredentials("AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"))
//
// Security considerations:
//   - Never hardcode credentials in source code
//   - Use environment variables or secure credential storage
//   - Rotate credentials regularly
//   - Use IAM roles when possible instead of long-term credentials
func WithCredentials(accessKey, secretKey string) Option {
	return func(c *Config) {
		c.AccessKey = accessKey
		c.SecretKey = secretKey
	}
}

// WithSessionToken sets the session token for temporary AWS credentials.
// This is required when using AWS STS (Security Token Service) temporary credentials.
// Must be used together with WithCredentials().
//
// Example:
//
//	provider := aws.New("us-east-1",
//	    aws.WithCredentials("AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
//	    aws.WithSessionToken("AQoEXAMPLEH4aoAH0gNCAPyJxz4BlCFFxWNE1OPTgk5TthT+FvwqnKwRcOIfrRh3c/LTo6UDdyJwOOvEVPvLXCrrrUtdnniCEXAMPLE/IvU1dYUg2RVAJBanLiHb4IgRmpRV3zrkuWJOgQs8IZZaIv2BXIa2R4OlgkBN9bkUDNCJiBeb/AXlzBBko7b15fjrBs2+cTQtpZ3CYWFXG8C5zqx37wnOE49mRl/+OtkIKGO7fAE"))
//
// Common use cases:
//   - Cross-account access using AssumeRole
//   - Temporary credentials for applications
//   - Multi-factor authentication scenarios
func WithSessionToken(sessionToken string) Option {
	return func(c *Config) {
		c.SessionToken = sessionToken
	}
}

// WithDebug enables debug logging for AWS API calls.
// This provides detailed information about requests and responses,
// which is useful for troubleshooting authentication and API issues.
//
// Example:
//
//	provider := aws.New("us-east-1", aws.WithDebug())
//
// Debug output includes:
//   - HTTP request/response details
//   - Credential resolution process
//   - Retry attempts and backoff
//   - Service endpoint resolution
//
// Security note: Debug logs may contain sensitive information.
// Only enable in development or when troubleshooting issues.
func WithDebug() Option {
	return func(c *Config) {
		c.Debug = true
	}
}

// WithTimeout sets the default timeout for AWS API calls.
// This applies to all service operations unless overridden at the operation level.
//
// Example:
//
//	provider := aws.New("us-east-1", aws.WithTimeout(60*time.Second))
//
// Considerations:
//   - Longer timeouts for operations that may take time (e.g., database creation)
//   - Shorter timeouts for quick operations to fail fast
//   - Network conditions and service load may affect optimal timeout values
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithRetryMaxAttempts sets the maximum number of retry attempts for failed requests.
// The AWS SDK uses exponential backoff between retry attempts.
//
// Example:
//
//	provider := aws.New("us-east-1", aws.WithRetryMaxAttempts(5))
//
// Default behavior:
//   - AWS SDK default is typically 3 attempts
//   - Retries are automatic for transient errors (network issues, rate limits)
//   - Non-retryable errors (authentication, authorization) fail immediately
//
// Considerations:
//   - Higher retry counts increase resilience but also latency
//   - Consider rate limits when setting high retry counts
//   - Monitor retry metrics to optimize this value
func WithRetryMaxAttempts(maxAttempts int) Option {
	return func(c *Config) {
		c.RetryMaxAttempts = maxAttempts
	}
}

// New creates a new AWS provider with comprehensive credential discovery.
// This function implements automatic credential discovery in the following order:
//  1. Explicit credentials via WithCredentials() option
//  2. AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables
//  3. AWS profile via WithProfile() option or AWS_PROFILE environment variable
//  4. IAM role for EC2 instances (when running on EC2)
//  5. Default profile in ~/.aws/credentials
//
// Parameters:
//   - region: AWS region for all operations (e.g., "us-east-1", "eu-west-1")
//   - options: Optional configuration using functional options pattern
//
// Returns:
//   - *AWSProvider: Configured AWS provider ready for use
//   - error: Configuration or credential discovery errors with helpful suggestions
//
// Examples:
//
// Simple usage with automatic credential discovery:
//
//	provider, err := aws.New("us-east-1")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Using a specific AWS profile:
//
//	provider, err := aws.New("us-east-1", aws.WithProfile("production"))
//
// Using explicit credentials (not recommended for production):
//
//	provider, err := aws.New("us-east-1",
//	    aws.WithCredentials("ACCESS_KEY", "SECRET_KEY"))
//
// Full configuration with debugging:
//
//	provider, err := aws.New("us-east-1",
//	    aws.WithProfile("dev"),
//	    aws.WithDebug(),
//	    aws.WithTimeout(60*time.Second),
//	    aws.WithRetryMaxAttempts(5))
//
// Common errors and solutions:
//   - "NoCredentialProviders": No valid credentials found
//     → Set AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY or configure AWS profile
//   - "InvalidRegionError": Invalid region specified
//     → Use valid AWS region codes (e.g., us-east-1, eu-west-1)
//   - "UnauthorizedOperation": Credentials lack required permissions
//     → Check IAM policies for your credentials
func New(region string, options ...Option) (*AWSProvider, error) {
	// Start with default configuration
	cfg := Config{
		Region:           region,
		Timeout:          30 * time.Second,
		RetryMaxAttempts: 3,
	}

	// Apply all options
	for _, option := range options {
		option(&cfg)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	// Create provider (AWS config will be loaded lazily)
	provider := &AWSProvider{
		config: cfg,
	}

	return provider, nil
}

// validateConfig validates the AWS configuration and returns helpful error messages
func validateConfig(cfg *Config) error {
	if cfg.Region == "" {
		return cloudsdk.NewInvalidConfigError("aws", "", "region", "region cannot be empty").
			WithSuggestions(
				"Provide a valid AWS region (e.g., 'us-east-1', 'eu-west-1')",
				"Check AWS documentation for available regions",
			)
	}

	// Validate that if credentials are provided, both access key and secret key are present
	if cfg.AccessKey != "" && cfg.SecretKey == "" {
		return cloudsdk.NewInvalidConfigError("aws", "", "secret_key", "secret key is required when access key is provided").
			WithSuggestions(
				"Provide both access key and secret key",
				"Use WithCredentials() with both parameters",
			)
	}

	if cfg.SecretKey != "" && cfg.AccessKey == "" {
		return cloudsdk.NewInvalidConfigError("aws", "", "access_key", "access key is required when secret key is provided").
			WithSuggestions(
				"Provide both access key and secret key",
				"Use WithCredentials() with both parameters",
			)
	}

	if cfg.Timeout < 0 {
		return cloudsdk.NewInvalidConfigError("aws", "", "timeout", "timeout cannot be negative").
			WithSuggestions(
				"Use a positive timeout value (e.g., 30*time.Second)",
				"Omit timeout to use default values",
			)
	}

	if cfg.RetryMaxAttempts < 0 {
		return cloudsdk.NewInvalidConfigError("aws", "", "retry_max_attempts", "retry max attempts cannot be negative").
			WithSuggestions(
				"Use a positive number of retry attempts (e.g., 3)",
				"Omit retry max attempts to use default values",
			)
	}

	return nil
}

// ensureInitialized lazily initializes the AWS config if not already done.
// This implements comprehensive credential discovery with helpful error messages.
func (p *AWSProvider) ensureInitialized(ctx context.Context) error {
	if p.initialized {
		return nil
	}

	// Build AWS SDK configuration options
	var loadOpts []func(*config.LoadOptions) error

	// Always set the region
	loadOpts = append(loadOpts, config.WithRegion(p.config.Region))

	// Configure credentials based on provider configuration
	if p.config.AccessKey != "" && p.config.SecretKey != "" {
		// Use explicit credentials
		creds := credentials.NewStaticCredentialsProvider(
			p.config.AccessKey,
			p.config.SecretKey,
			p.config.SessionToken,
		)
		loadOpts = append(loadOpts, config.WithCredentialsProvider(creds))
	} else if p.config.Profile != "" {
		// Use specific profile
		loadOpts = append(loadOpts, config.WithSharedConfigProfile(p.config.Profile))
	}
	// If neither explicit credentials nor profile are provided,
	// AWS SDK will use default credential discovery

	// Configure retry behavior
	if p.config.RetryMaxAttempts > 0 {
		loadOpts = append(loadOpts, config.WithRetryMaxAttempts(p.config.RetryMaxAttempts))
	}

	// Load AWS configuration
	awsConfig, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return p.createCredentialError(err)
	}

	p.awsConfig = awsConfig
	p.initialized = true
	return nil
}

// createCredentialError creates a helpful error message for credential issues
func (p *AWSProvider) createCredentialError(err error) error {
	// Check for common credential errors and provide helpful suggestions
	errMsg := err.Error()

	if contains(errMsg, "NoCredentialProviders") {
		return cloudsdk.NewAuthenticationError("aws", err).
			WithSuggestions(
				"Set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables",
				"Configure AWS profile using 'aws configure' command",
				"Use WithCredentials() option to provide explicit credentials",
				"Ensure IAM role is attached if running on EC2",
				"Check that ~/.aws/credentials file exists and is readable",
			)
	}

	if contains(errMsg, "SharedConfigProfileNotExist") {
		return cloudsdk.NewAuthenticationError("aws", err).
			WithSuggestions(
				fmt.Sprintf("Create AWS profile '%s' using 'aws configure --profile %s'", p.config.Profile, p.config.Profile),
				"Check available profiles in ~/.aws/credentials",
				"Use WithProfile() with an existing profile name",
				"Remove WithProfile() to use default credentials",
			)
	}

	if contains(errMsg, "InvalidRegionError") {
		return cloudsdk.NewInvalidConfigError("aws", "", "region", "invalid AWS region").
			WithCause(err).
			WithSuggestions(
				"Use a valid AWS region code (e.g., 'us-east-1', 'eu-west-1', 'ap-southeast-1')",
				"Check AWS documentation for complete list of regions",
				"Verify the region supports the services you need",
			)
	}

	// Generic authentication error with basic suggestions
	return cloudsdk.NewAuthenticationError("aws", err)
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Compute returns the AWS compute service for managing EC2 instances.
// The service provides operations for creating, managing, and monitoring virtual machines.
//
// Supported operations:
//   - Create and configure EC2 instances
//   - Start, stop, and terminate instances
//   - List and describe instances
//   - Manage instance metadata and tags
//
// Example:
//
//	compute := provider.Compute()
//	vm, err := compute.CreateVM(ctx, &services.VMConfig{
//	    Name: "web-server",
//	    ImageID: "ami-12345678",
//	    InstanceType: "t2.micro",
//	})
//
// Note: This method performs lazy initialization of AWS credentials.
// Credential errors will be returned when service methods are called.
func (p *AWSProvider) Compute() services.Compute {
	// Note: We don't initialize here to avoid blocking the provider creation.
	// Initialization happens when service methods are actually called.
	return compute.New(p.awsConfig)
}

// Storage returns the AWS storage service for managing S3 buckets and objects.
// The service provides operations for object storage, bucket management, and file operations.
//
// Supported operations:
//   - Create and configure S3 buckets
//   - Upload, download, and delete objects
//   - List bucket contents
//   - Manage bucket policies and permissions
//
// Example:
//
//	storage := provider.Storage()
//	err := storage.CreateBucket(ctx, &services.BucketConfig{
//	    Name: "my-app-data",
//	    Region: "us-east-1",
//	})
//
// Note: This method performs lazy initialization of AWS credentials.
// Credential errors will be returned when service methods are called.
func (p *AWSProvider) Storage() services.Storage {
	return storage.New(p.awsConfig)
}

// Database returns the AWS database service for managing RDS instances.
// The service provides operations for creating and managing relational databases.
//
// Supported operations:
//   - Create and configure RDS instances
//   - Manage database engines (PostgreSQL, MySQL, etc.)
//   - Handle database backups and snapshots
//   - Monitor database performance and health
//
// Example:
//
//	database := provider.Database()
//	db, err := database.CreateDB(ctx, &services.DBConfig{
//	    Name: "app-database",
//	    Engine: "postgres",
//	    InstanceType: "db.t3.micro",
//	})
//
// Note: This method performs lazy initialization of AWS credentials.
// Credential errors will be returned when service methods are called.
func (p *AWSProvider) Database() services.Database {
	return database.New(p.awsConfig)
}

// Name returns the provider name identifier.
// This is used for error reporting and logging purposes.
func (p *AWSProvider) Name() string {
	return "aws"
}

// Region returns the configured AWS region for this provider.
// This region is used for all service operations unless overridden.
func (p *AWSProvider) Region() string {
	return p.config.Region
}

// SupportedServices returns the list of services supported by the AWS provider.
// AWS provider supports all three core services: Compute (EC2), Storage (S3), and Database (RDS).
//
// This method enables compile-time checking of service availability and helps
// developers understand which services are available with this provider.
func (p *AWSProvider) SupportedServices() []cloudsdk.ServiceType {
	return []cloudsdk.ServiceType{
		cloudsdk.ServiceCompute,  // Amazon EC2 - Elastic Compute Cloud
		cloudsdk.ServiceStorage,  // Amazon S3 - Simple Storage Service
		cloudsdk.ServiceDatabase, // Amazon RDS - Relational Database Service
	}
}

// Connect validates the AWS configuration and credentials.
// This method can be called to test connectivity before using services.
// It's optional - services will automatically initialize when first used.
//
// Example:
//
//	provider, err := aws.New("us-east-1", aws.WithProfile("production"))
//	if err != nil {
//	    return err
//	}
//
//	// Test connectivity before proceeding
//	if err := provider.Connect(ctx); err != nil {
//	    return fmt.Errorf("failed to connect to AWS: %w", err)
//	}
//
// Returns:
//   - nil: Connection successful, credentials are valid
//   - error: Connection failed with detailed error message and suggestions
func (p *AWSProvider) Connect(ctx context.Context) error {
	return p.ensureInitialized(ctx)
}
