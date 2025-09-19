package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/smithy-go"
)

// RetryConfig defines retry behavior for AWS RDS operations
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
		if !isRetryableRDSError(err) {
			return err
		}

		if attempt < config.MaxAttempts {
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
			log.Printf("AWS Database: Retrying operation (attempt %d/%d) after %v due to: %v",
				attempt+1, config.MaxAttempts, delay, err)
		}
	}

	return lastErr
}

// isRetryableRDSError determines if an RDS error should be retried
func isRetryableRDSError(err error) bool {
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
		case "Throttling", "ThrottlingException", "RequestLimitExceeded":
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

// wrapRDSError converts RDS errors to CloudError with helpful context
func wrapRDSError(err error, provider, service, operation string) error {
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
			WithSuggestions("Increase the operation timeout", "Check network connectivity", "Verify AWS RDS service status")
	}

	// Handle AWS-specific errors
	var ae smithy.APIError
	if errors.As(err, &ae) {
		code := ae.ErrorCode()
		message := ae.ErrorMessage()

		switch code {
		case "AccessDenied", "UnauthorizedOperation":
			return cloudsdk.NewAuthorizationError(provider, service, operation, err).
				WithSuggestions(
					"Verify your IAM user/role has the required RDS permissions",
					"Check if your account has the necessary service limits",
					"Ensure you're operating in the correct AWS region",
				)

		case "InvalidUserID.NotFound", "SignatureDoesNotMatch", "TokenRefreshRequired":
			return cloudsdk.NewAuthenticationError(provider, err).
				WithSuggestions(
					"Verify your AWS access key and secret key are correct",
					"Check if your credentials have expired",
					"Ensure your system clock is synchronized",
				)

		case "DBInstanceNotFoundFault":
			instanceID := extractDBInstanceIDFromError(message)
			return cloudsdk.NewResourceNotFoundError(provider, service, "database instance", instanceID).
				WithSuggestions(
					"Verify the database instance ID is correct",
					"Check that the instance exists in the current region",
					"Ensure the instance hasn't been deleted",
				)

		case "DBInstanceAlreadyExistsFault":
			return cloudsdk.NewCloudError(cloudsdk.ErrResourceConflict, "Database instance already exists", provider, service, operation).
				WithCause(err).
				WithSuggestions(
					"Choose a different database instance identifier",
					"Check if you already have an instance with this name",
					"Use a more specific naming convention",
				)

		case "InvalidDBInstanceClass":
			return cloudsdk.NewInvalidConfigError(provider, service, "InstanceClass", "Invalid database instance class").
				WithSuggestions(
					"Check the instance class is available in your region",
					"Verify the instance class spelling is correct",
					"Ensure your account has access to this instance class",
					"Use valid instance classes like db.t3.micro, db.t3.small, etc.",
				)

		case "InvalidEngine":
			return cloudsdk.NewInvalidConfigError(provider, service, "Engine", "Invalid database engine").
				WithSuggestions(
					"Use supported engines: mysql, postgres, oracle-ee, sqlserver-ex, etc.",
					"Check the engine name spelling is correct",
					"Verify the engine is available in your region",
				)

		case "InvalidParameterValue":
			return cloudsdk.NewInvalidConfigError(provider, service, "Parameter", message).
				WithSuggestions(
					"Check all parameter values are within valid ranges",
					"Verify parameter formats are correct",
					"Refer to AWS RDS documentation for valid values",
				)

		case "DBSubnetGroupNotFoundFault":
			return cloudsdk.NewInvalidConfigError(provider, service, "DBSubnetGroup", "DB subnet group not found").
				WithSuggestions(
					"Create a DB subnet group first",
					"Verify the subnet group name is correct",
					"Ensure the subnet group exists in the current region",
				)

		case "InvalidVPCNetworkStateFault":
			return cloudsdk.NewInvalidConfigError(provider, service, "VPC", "Invalid VPC network state").
				WithSuggestions(
					"Ensure your VPC is properly configured",
					"Check that subnets are available in multiple AZs",
					"Verify security groups allow database access",
				)

		case "InsufficientDBInstanceCapacity":
			return cloudsdk.NewCloudError(cloudsdk.ErrResourceConflict, "Insufficient capacity for database instance", provider, service, operation).
				WithCause(err).
				WithSuggestions(
					"Try a different instance class",
					"Try launching in a different availability zone",
					"Wait and retry later when capacity becomes available",
				)

		case "DBInstanceLimitExceeded":
			return cloudsdk.NewCloudError(cloudsdk.ErrResourceConflict, "Database instance limit exceeded", provider, service, operation).
				WithCause(err).
				WithSuggestions(
					"Request a limit increase from AWS Support",
					"Delete unused database instances to free up capacity",
					"Use different instance classes with available capacity",
				)

		case "Throttling", "RequestLimitExceeded":
			return cloudsdk.NewRateLimitError(provider, service, operation, 0).
				WithCause(err).
				WithSuggestions(
					"Reduce the frequency of API calls",
					"Implement exponential backoff (this is done automatically)",
					"Consider using batch operations where available",
				)

		default:
			// Generic AWS error
			return cloudsdk.NewCloudError(cloudsdk.ErrProviderError, fmt.Sprintf("RDS error: %s", message), provider, service, operation).
				WithCause(err).
				WithSuggestions(
					"Check AWS RDS service status for any ongoing issues",
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

// extractDBInstanceIDFromError attempts to extract DB instance ID from error messages
func extractDBInstanceIDFromError(message string) string {
	// Look for patterns in RDS error messages
	parts := strings.Fields(message)
	for _, part := range parts {
		// RDS instance identifiers are typically alphanumeric with hyphens
		if len(part) > 3 && strings.Contains(part, "-") {
			return strings.Trim(part, "',.")
		}
	}
	return "unknown"
}

// validateDBConfig validates database configuration parameters
func validateDBConfig(config *services.DBConfig) error {
	if config == nil {
		return fmt.Errorf("database configuration cannot be nil")
	}

	// Validate required fields
	if config.Name == "" {
		return fmt.Errorf("database instance identifier is required")
	}
	if config.Engine == "" {
		return fmt.Errorf("database engine is required")
	}
	if config.InstanceClass == "" {
		return fmt.Errorf("database instance class is required")
	}
	if config.MasterUsername == "" {
		return fmt.Errorf("master username is required")
	}
	if config.MasterPassword == "" {
		return fmt.Errorf("master password is required")
	}

	// Validate DB instance identifier format
	if err := validateDBInstanceIdentifier(config.Name); err != nil {
		return fmt.Errorf("invalid database instance identifier: %w", err)
	}

	// Validate engine
	if err := validateEngine(config.Engine); err != nil {
		return fmt.Errorf("invalid database engine: %w", err)
	}

	// Validate master username
	if err := validateMasterUsername(config.MasterUsername, config.Engine); err != nil {
		return fmt.Errorf("invalid master username: %w", err)
	}

	// Validate master password
	if err := validateMasterPassword(config.MasterPassword, config.Engine); err != nil {
		return fmt.Errorf("invalid master password: %w", err)
	}

	// Validate allocated storage
	if config.AllocatedStorage < 20 {
		return fmt.Errorf("allocated storage must be at least 20 GB")
	}

	return nil
}

// validateDBInstanceIdentifier validates RDS instance identifier format
func validateDBInstanceIdentifier(identifier string) error {
	if len(identifier) < 1 || len(identifier) > 63 {
		return fmt.Errorf("must be 1-63 characters long")
	}

	// Must start with a letter
	if !((identifier[0] >= 'a' && identifier[0] <= 'z') || (identifier[0] >= 'A' && identifier[0] <= 'Z')) {
		return fmt.Errorf("must start with a letter")
	}

	// Can contain letters, numbers, and hyphens
	validChars := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`)
	if !validChars.MatchString(identifier) {
		return fmt.Errorf("can only contain letters, numbers, and hyphens")
	}

	// Cannot end with a hyphen
	if strings.HasSuffix(identifier, "-") {
		return fmt.Errorf("cannot end with a hyphen")
	}

	// Cannot contain consecutive hyphens
	if strings.Contains(identifier, "--") {
		return fmt.Errorf("cannot contain consecutive hyphens")
	}

	return nil
}

// validateEngine validates database engine
func validateEngine(engine string) error {
	validEngines := map[string]bool{
		"mysql":             true,
		"postgres":          true,
		"oracle-ee":         true,
		"oracle-se2":        true,
		"oracle-se1":        true,
		"oracle-se":         true,
		"sqlserver-ee":      true,
		"sqlserver-se":      true,
		"sqlserver-ex":      true,
		"sqlserver-web":     true,
		"mariadb":           true,
		"aurora":            true,
		"aurora-mysql":      true,
		"aurora-postgresql": true,
	}

	if !validEngines[strings.ToLower(engine)] {
		return fmt.Errorf("unsupported engine '%s'. Valid engines: mysql, postgres, oracle-ee, sqlserver-ex, mariadb, etc.", engine)
	}

	return nil
}

// validateMasterUsername validates master username based on engine
func validateMasterUsername(username, engine string) error {
	if len(username) < 1 || len(username) > 63 {
		return fmt.Errorf("must be 1-63 characters long")
	}

	// Engine-specific validations
	switch strings.ToLower(engine) {
	case "mysql", "mariadb":
		if username == "root" {
			return fmt.Errorf("'root' is not allowed for MySQL/MariaDB")
		}
	case "postgres":
		if username == "postgres" {
			return fmt.Errorf("'postgres' is not allowed for PostgreSQL")
		}
	case "oracle-ee", "oracle-se2", "oracle-se1", "oracle-se":
		if username == "sys" || username == "system" {
			return fmt.Errorf("'sys' and 'system' are not allowed for Oracle")
		}
	}

	// General validation - must start with letter
	if !((username[0] >= 'a' && username[0] <= 'z') || (username[0] >= 'A' && username[0] <= 'Z')) {
		return fmt.Errorf("must start with a letter")
	}

	return nil
}

// validateMasterPassword validates master password based on engine
func validateMasterPassword(password, engine string) error {
	if len(password) < 8 {
		return fmt.Errorf("must be at least 8 characters long")
	}

	// Engine-specific length limits
	switch strings.ToLower(engine) {
	case "mysql", "mariadb":
		if len(password) > 41 {
			return fmt.Errorf("must be no more than 41 characters for MySQL/MariaDB")
		}
	case "postgres":
		if len(password) > 128 {
			return fmt.Errorf("must be no more than 128 characters for PostgreSQL")
		}
	case "oracle-ee", "oracle-se2", "oracle-se1", "oracle-se":
		if len(password) > 30 {
			return fmt.Errorf("must be no more than 30 characters for Oracle")
		}
	}

	// Check for printable ASCII characters
	for _, r := range password {
		if r < 32 || r > 126 {
			return fmt.Errorf("must contain only printable ASCII characters")
		}
	}

	return nil
}

// logRequest logs RDS API requests for debugging (when debug is enabled)
func logRequest(operation string, input interface{}, debug bool) {
	if debug {
		log.Printf("AWS Database: %s request: %+v", operation, input)
	}
}

// logResponse logs RDS API responses for debugging (when debug is enabled)
func logResponse(operation string, output interface{}, err error, debug bool) {
	if debug {
		if err != nil {
			log.Printf("AWS Database: %s error: %v", operation, err)
		} else {
			log.Printf("AWS Database: %s response: %+v", operation, output)
		}
	}
}

// RDSClientInterface defines methods we need from RDS client for testing
type RDSClientInterface interface {
	CreateDBInstance(ctx context.Context, input *rds.CreateDBInstanceInput, opts ...func(*rds.Options)) (*rds.CreateDBInstanceOutput, error)
	DescribeDBInstances(ctx context.Context, input *rds.DescribeDBInstancesInput, opts ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
	DeleteDBInstance(ctx context.Context, input *rds.DeleteDBInstanceInput, opts ...func(*rds.Options)) (*rds.DeleteDBInstanceOutput, error)
}

// AWSDatabase implements the Database interface for AWS
type AWSDatabase struct {
	client      RDSClientInterface
	debug       bool
	retryConfig RetryConfig
}

// New creates a new AWSDatabase instance with real AWS client
func New(cfg aws.Config) services.Database {
	client := rds.NewFromConfig(cfg)
	return &AWSDatabase{
		client:      client,
		debug:       false,
		retryConfig: DefaultRetryConfig,
	}
}

// NewWithClient creates a new AWSDatabase instance with custom client (for testing)
func NewWithClient(client RDSClientInterface) services.Database {
	return &AWSDatabase{
		client:      client,
		debug:       false,
		retryConfig: DefaultRetryConfig,
	}
}

// NewWithOptions creates a new AWSDatabase instance with custom options
func NewWithOptions(cfg aws.Config, debug bool, retryConfig *RetryConfig) services.Database {
	client := rds.NewFromConfig(cfg)

	finalRetryConfig := DefaultRetryConfig
	if retryConfig != nil {
		finalRetryConfig = *retryConfig
	}

	return &AWSDatabase{
		client:      client,
		debug:       debug,
		retryConfig: finalRetryConfig,
	}
}

// CreateDB creates a new RDS instance
func (d *AWSDatabase) CreateDB(ctx context.Context, config *services.DBConfig) (*services.DBInstance, error) {
	// Validate input configuration
	if err := validateDBConfig(config); err != nil {
		return nil, cloudsdk.NewInvalidConfigError("aws", "database", "config", err.Error())
	}

	input := &rds.CreateDBInstanceInput{
		DBInstanceIdentifier: aws.String(config.Name),
		DBInstanceClass:      aws.String(config.InstanceClass),
		Engine:               aws.String(config.Engine),
		MasterUsername:       aws.String(config.MasterUsername),
		MasterUserPassword:   aws.String(config.MasterPassword),
		AllocatedStorage:     aws.Int32(config.AllocatedStorage),
	}

	// Add optional parameters
	if config.EngineVersion != "" {
		input.EngineVersion = aws.String(config.EngineVersion)
	}
	if config.DBName != "" {
		input.DBName = aws.String(config.DBName)
	}
	// Note: VpcSecurityGroupIds, DBSubnetGroupName, BackupRetentionPeriod fields not available in current DBConfig struct
	if config.StorageEncrypted != nil {
		input.StorageEncrypted = config.StorageEncrypted
	}

	// Mask sensitive data in logs
	logInput := map[string]interface{}{
		"DBInstanceIdentifier": config.Name,
		"DBInstanceClass":      config.InstanceClass,
		"Engine":               config.Engine,
		"MasterUsername":       config.MasterUsername,
		"MasterUserPassword":   "[REDACTED]",
		"AllocatedStorage":     config.AllocatedStorage,
	}
	logRequest("CreateDBInstance", logInput, d.debug)

	var resp *rds.CreateDBInstanceOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, d.retryConfig, func() error {
		resp, err = d.client.CreateDBInstance(ctx, input)
		return err
	})

	logResponse("CreateDBInstance", resp, retryErr, d.debug)

	if retryErr != nil {
		return nil, wrapRDSError(retryErr, "aws", "database", "CreateDB")
	}

	if resp.DBInstance == nil {
		return nil, cloudsdk.NewCloudError(cloudsdk.ErrProviderError, "No database instance returned from AWS", "aws", "database", "CreateDB").
			WithSuggestions(
				"Check AWS RDS service status",
				"Verify your account limits",
				"Try again with different parameters",
			)
	}

	// Create DB instance response with proper error handling for optional fields
	dbInstance := &services.DBInstance{
		ID:     aws.ToString(resp.DBInstance.DBInstanceIdentifier),
		Name:   aws.ToString(resp.DBInstance.DBInstanceIdentifier),
		Engine: aws.ToString(resp.DBInstance.Engine),
		Status: aws.ToString(resp.DBInstance.DBInstanceStatus),
	}

	// Safely handle optional fields
	if resp.DBInstance.Endpoint != nil && resp.DBInstance.Endpoint.Address != nil {
		dbInstance.Endpoint = aws.ToString(resp.DBInstance.Endpoint.Address)
		// Note: Port field not available in current DBInstance struct
	}
	if resp.DBInstance.InstanceCreateTime != nil {
		dbInstance.LaunchTime = resp.DBInstance.InstanceCreateTime.String()
	}
	// Note: EngineVersion field not available in current DBInstance struct

	return dbInstance, nil
}

// ListDBs lists all RDS instances
func (d *AWSDatabase) ListDBs(ctx context.Context) ([]*services.DBInstance, error) {
	input := &rds.DescribeDBInstancesInput{}

	logRequest("DescribeDBInstances", input, d.debug)

	var resp *rds.DescribeDBInstancesOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, d.retryConfig, func() error {
		resp, err = d.client.DescribeDBInstances(ctx, input)
		return err
	})

	logResponse("DescribeDBInstances", resp, retryErr, d.debug)

	if retryErr != nil {
		return nil, wrapRDSError(retryErr, "aws", "database", "ListDBs")
	}

	dbs := make([]*services.DBInstance, len(resp.DBInstances))
	for i, inst := range resp.DBInstances {
		dbInstance := &services.DBInstance{
			ID:     aws.ToString(inst.DBInstanceIdentifier),
			Name:   aws.ToString(inst.DBInstanceIdentifier),
			Engine: aws.ToString(inst.Engine),
			Status: aws.ToString(inst.DBInstanceStatus),
		}

		// Safely handle optional fields
		if inst.Endpoint != nil && inst.Endpoint.Address != nil {
			dbInstance.Endpoint = aws.ToString(inst.Endpoint.Address)
			// Note: Port field not available in current DBInstance struct
		}
		if inst.InstanceCreateTime != nil {
			dbInstance.LaunchTime = inst.InstanceCreateTime.String()
		}
		// Note: EngineVersion field not available in current DBInstance struct

		dbs[i] = dbInstance
	}

	return dbs, nil
}

// GetDB gets a specific RDS instance by ID
func (d *AWSDatabase) GetDB(ctx context.Context, id string) (*services.DBInstance, error) {
	// Validate input
	if id == "" {
		return nil, cloudsdk.NewInvalidConfigError("aws", "database", "id", "database instance ID cannot be empty")
	}

	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(id),
	}

	logRequest("DescribeDBInstances", input, d.debug)

	var resp *rds.DescribeDBInstancesOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, d.retryConfig, func() error {
		resp, err = d.client.DescribeDBInstances(ctx, input)
		return err
	})

	logResponse("DescribeDBInstances", resp, retryErr, d.debug)

	if retryErr != nil {
		return nil, wrapRDSError(retryErr, "aws", "database", "GetDB")
	}

	if len(resp.DBInstances) == 0 {
		return nil, cloudsdk.NewResourceNotFoundError("aws", "database", "database instance", id)
	}

	inst := resp.DBInstances[0]
	dbInstance := &services.DBInstance{
		ID:     aws.ToString(inst.DBInstanceIdentifier),
		Name:   aws.ToString(inst.DBInstanceIdentifier),
		Engine: aws.ToString(inst.Engine),
		Status: aws.ToString(inst.DBInstanceStatus),
	}

	// Safely handle optional fields
	if inst.Endpoint != nil && inst.Endpoint.Address != nil {
		dbInstance.Endpoint = aws.ToString(inst.Endpoint.Address)
		// Note: Port field not available in current DBInstance struct
	}
	if inst.InstanceCreateTime != nil {
		dbInstance.LaunchTime = inst.InstanceCreateTime.String()
	}
	// Note: EngineVersion field not available in current DBInstance struct

	return dbInstance, nil
}

func (d *AWSDatabase) DeleteDB(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		return cloudsdk.NewInvalidConfigError("aws", "database", "id", "database instance ID cannot be empty")
	}

	input := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier:   aws.String(id),
		SkipFinalSnapshot:      aws.Bool(true), // Skip final snapshot by default for simplicity
		DeleteAutomatedBackups: aws.Bool(true), // Delete automated backups
	}

	logRequest("DeleteDBInstance", input, d.debug)

	var resp *rds.DeleteDBInstanceOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, d.retryConfig, func() error {
		resp, err = d.client.DeleteDBInstance(ctx, input)
		return err
	})

	logResponse("DeleteDBInstance", resp, retryErr, d.debug)

	if retryErr != nil {
		return wrapRDSError(retryErr, "aws", "database", "DeleteDB")
	}

	return nil
}
