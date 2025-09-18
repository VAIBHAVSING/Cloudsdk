package compute

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
)

// RetryConfig defines retry behavior for AWS operations
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
		if !isRetryableAWSError(err) {
			return err
		}

		if attempt < config.MaxAttempts {
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
			log.Printf("AWS Compute: Retrying operation (attempt %d/%d) after %v due to: %v",
				attempt+1, config.MaxAttempts, delay, err)
		}
	}

	return lastErr
}

// isRetryableAWSError determines if an AWS error should be retried
func isRetryableAWSError(err error) bool {
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

// wrapAWSError converts AWS errors to CloudError with helpful context
func wrapAWSError(err error, provider, service, operation string) error {
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
			WithSuggestions("Increase the operation timeout", "Check network connectivity", "Verify AWS service status")
	}

	// Handle AWS-specific errors
	var ae smithy.APIError
	if errors.As(err, &ae) {
		code := ae.ErrorCode()
		message := ae.ErrorMessage()

		switch code {
		case "UnauthorizedOperation", "AccessDenied":
			return cloudsdk.NewAuthorizationError(provider, service, operation, err).
				WithSuggestions(
					"Verify your IAM user/role has the required EC2 permissions",
					"Check if your account has the necessary service limits",
					"Ensure you're operating in the correct AWS region",
				)

		case "AuthFailure", "InvalidUserID.NotFound", "SignatureDoesNotMatch":
			return cloudsdk.NewAuthenticationError(provider, err).
				WithSuggestions(
					"Verify your AWS access key and secret key are correct",
					"Check if your credentials have expired",
					"Ensure your system clock is synchronized",
				)

		case "InvalidInstanceID.NotFound":
			instanceID := extractInstanceIDFromError(message)
			return cloudsdk.NewResourceNotFoundError(provider, service, "instance", instanceID).
				WithSuggestions(
					"Verify the instance ID is correct",
					"Check that the instance exists in the current region",
					"Ensure the instance hasn't been terminated",
				)

		case "InvalidAMIID.NotFound":
			return cloudsdk.NewInvalidConfigError(provider, service, "ImageID", "AMI not found").
				WithSuggestions(
					"Verify the AMI ID is correct and exists in your region",
					"Check if the AMI is public or if you have access to it",
					"Ensure the AMI is compatible with the instance type",
				)

		case "InvalidInstanceType":
			return cloudsdk.NewInvalidConfigError(provider, service, "InstanceType", "Invalid instance type").
				WithSuggestions(
					"Check the instance type is available in your region",
					"Verify the instance type spelling is correct",
					"Ensure your account has access to this instance type",
				)

		case "InvalidKeyPair.NotFound":
			return cloudsdk.NewInvalidConfigError(provider, service, "KeyName", "Key pair not found").
				WithSuggestions(
					"Verify the key pair name is correct",
					"Check that the key pair exists in the current region",
					"Create the key pair if it doesn't exist",
				)

		case "Throttling", "RequestLimitExceeded":
			return cloudsdk.NewRateLimitError(provider, service, operation, 0).
				WithCause(err).
				WithSuggestions(
					"Reduce the frequency of API calls",
					"Implement exponential backoff (this is done automatically)",
					"Consider using batch operations where available",
				)

		case "InsufficientInstanceCapacity":
			return cloudsdk.NewCloudError(cloudsdk.ErrResourceConflict, "Insufficient capacity for instance type", provider, service, operation).
				WithCause(err).
				WithSuggestions(
					"Try a different instance type",
					"Try launching in a different availability zone",
					"Wait and retry later when capacity becomes available",
				)

		case "InstanceLimitExceeded":
			return cloudsdk.NewCloudError(cloudsdk.ErrResourceConflict, "Instance limit exceeded", provider, service, operation).
				WithCause(err).
				WithSuggestions(
					"Request a limit increase from AWS Support",
					"Terminate unused instances to free up capacity",
					"Use different instance types with available capacity",
				)

		default:
			// Generic AWS error
			return cloudsdk.NewCloudError(cloudsdk.ErrProviderError, fmt.Sprintf("AWS error: %s", message), provider, service, operation).
				WithCause(err).
				WithSuggestions(
					"Check AWS service status for any ongoing issues",
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

// extractInstanceIDFromError attempts to extract instance ID from error messages
func extractInstanceIDFromError(message string) string {
	// Look for patterns like "i-1234567890abcdef0"
	parts := strings.Fields(message)
	for _, part := range parts {
		if strings.HasPrefix(part, "i-") && len(part) == 19 {
			return strings.Trim(part, "',.")
		}
	}
	return "unknown"
}

// logRequest logs AWS API requests for debugging (when debug is enabled)
func logRequest(operation string, input interface{}, debug bool) {
	if debug {
		log.Printf("AWS Compute: %s request: %+v", operation, input)
	}
}

// logResponse logs AWS API responses for debugging (when debug is enabled)
func logResponse(operation string, output interface{}, err error, debug bool) {
	if debug {
		if err != nil {
			log.Printf("AWS Compute: %s error: %v", operation, err)
		} else {
			log.Printf("AWS Compute: %s response: %+v", operation, output)
		}
	}
}

// EC2ClientInterface defines methods we need from EC2 client for testing
type EC2ClientInterface interface {
	RunInstances(ctx context.Context, input *ec2.RunInstancesInput, opts ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	DescribeInstances(ctx context.Context, input *ec2.DescribeInstancesInput, opts ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	StartInstances(ctx context.Context, input *ec2.StartInstancesInput, opts ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	StopInstances(ctx context.Context, input *ec2.StopInstancesInput, opts ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	TerminateInstances(ctx context.Context, input *ec2.TerminateInstancesInput, opts ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	DescribeInstanceTypes(ctx context.Context, input *ec2.DescribeInstanceTypesInput, opts ...func(*ec2.Options)) (*ec2.DescribeInstanceTypesOutput, error)
	DescribePlacementGroups(ctx context.Context, input *ec2.DescribePlacementGroupsInput, opts ...func(*ec2.Options)) (*ec2.DescribePlacementGroupsOutput, error)
	CreatePlacementGroup(ctx context.Context, input *ec2.CreatePlacementGroupInput, opts ...func(*ec2.Options)) (*ec2.CreatePlacementGroupOutput, error)
	DeletePlacementGroup(ctx context.Context, input *ec2.DeletePlacementGroupInput, opts ...func(*ec2.Options)) (*ec2.DeletePlacementGroupOutput, error)
	RequestSpotInstances(ctx context.Context, input *ec2.RequestSpotInstancesInput, opts ...func(*ec2.Options)) (*ec2.RequestSpotInstancesOutput, error)
	DescribeSpotInstanceRequests(ctx context.Context, input *ec2.DescribeSpotInstanceRequestsInput, opts ...func(*ec2.Options)) (*ec2.DescribeSpotInstanceRequestsOutput, error)
	CancelSpotInstanceRequests(ctx context.Context, input *ec2.CancelSpotInstanceRequestsInput, opts ...func(*ec2.Options)) (*ec2.CancelSpotInstanceRequestsOutput, error)
}

// AWSCompute implements the Compute interface for AWS
type AWSCompute struct {
	client             EC2ClientInterface
	instanceTypesSvc   *InstanceTypesServiceImpl
	placementGroupsSvc *PlacementGroupsServiceImpl
	spotInstancesSvc   *SpotInstancesServiceImpl
	debug              bool
	retryConfig        RetryConfig
}

// New creates a new AWSCompute instance with real AWS client
func New(cfg aws.Config) services.Compute {
	client := ec2.NewFromConfig(cfg)
	return &AWSCompute{
		client:             client,
		instanceTypesSvc:   &InstanceTypesServiceImpl{client: client, debug: false},
		placementGroupsSvc: &PlacementGroupsServiceImpl{client: client, debug: false},
		spotInstancesSvc:   &SpotInstancesServiceImpl{client: client, debug: false},
		debug:              false,
		retryConfig:        DefaultRetryConfig,
	}
}

// NewWithClient creates a new AWSCompute instance with custom client (for testing)
func NewWithClient(client EC2ClientInterface) services.Compute {
	return &AWSCompute{
		client:             client,
		instanceTypesSvc:   &InstanceTypesServiceImpl{client: client, debug: false},
		placementGroupsSvc: &PlacementGroupsServiceImpl{client: client, debug: false},
		spotInstancesSvc:   &SpotInstancesServiceImpl{client: client, debug: false},
		debug:              false,
		retryConfig:        DefaultRetryConfig,
	}
}

// NewWithOptions creates a new AWSCompute instance with custom options
func NewWithOptions(cfg aws.Config, debug bool, retryConfig *RetryConfig) services.Compute {
	client := ec2.NewFromConfig(cfg)

	finalRetryConfig := DefaultRetryConfig
	if retryConfig != nil {
		finalRetryConfig = *retryConfig
	}

	return &AWSCompute{
		client:             client,
		instanceTypesSvc:   &InstanceTypesServiceImpl{client: client, debug: debug},
		placementGroupsSvc: &PlacementGroupsServiceImpl{client: client, debug: debug},
		spotInstancesSvc:   &SpotInstancesServiceImpl{client: client, debug: debug},
		debug:              debug,
		retryConfig:        finalRetryConfig,
	}
}

// CreateVM creates a new virtual machine
func (c *AWSCompute) CreateVM(ctx context.Context, config *services.VMConfig) (*services.VM, error) {
	// Validate input configuration
	if config == nil {
		return nil, cloudsdk.NewInvalidConfigError("aws", "compute", "config", "configuration cannot be nil")
	}
	if config.ImageID == "" {
		return nil, cloudsdk.NewInvalidConfigError("aws", "compute", "ImageID", "image ID is required")
	}
	if config.InstanceType == "" {
		return nil, cloudsdk.NewInvalidConfigError("aws", "compute", "InstanceType", "instance type is required")
	}

	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(config.ImageID),
		InstanceType: types.InstanceType(config.InstanceType),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	// Add optional parameters
	if config.KeyName != "" {
		input.KeyName = aws.String(config.KeyName)
	}
	if config.UserData != "" {
		input.UserData = aws.String(config.UserData)
	}
	if len(config.SecurityGroups) > 0 {
		input.SecurityGroupIds = make([]string, len(config.SecurityGroups))
		copy(input.SecurityGroupIds, config.SecurityGroups)
	}

	logRequest("RunInstances", input, c.debug)

	var resp *ec2.RunInstancesOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, c.retryConfig, func() error {
		resp, err = c.client.RunInstances(ctx, input)
		return err
	})

	logResponse("RunInstances", resp, retryErr, c.debug)

	if retryErr != nil {
		return nil, wrapAWSError(retryErr, "aws", "compute", "CreateVM")
	}

	if len(resp.Instances) == 0 {
		return nil, cloudsdk.NewCloudError(cloudsdk.ErrProviderError, "No instances were created", "aws", "compute", "CreateVM").
			WithSuggestions(
				"Check AWS service status",
				"Verify your account limits",
				"Try again with different parameters",
			)
	}

	inst := resp.Instances[0]

	// Create VM response with proper error handling for optional fields
	vm := &services.VM{
		ID:    aws.ToString(inst.InstanceId),
		Name:  config.Name, // AWS doesn't set name here, but we can tag later
		State: string(inst.State.Name),
	}

	// Safely handle optional fields
	if inst.PublicIpAddress != nil {
		vm.PublicIP = aws.ToString(inst.PublicIpAddress)
	}
	if inst.PrivateIpAddress != nil {
		vm.PrivateIP = aws.ToString(inst.PrivateIpAddress)
	}
	if inst.LaunchTime != nil {
		vm.LaunchTime = inst.LaunchTime.String()
	}

	// Add name tag if specified
	if config.Name != "" {
		tagErr := c.addNameTag(ctx, aws.ToString(inst.InstanceId), config.Name)
		if tagErr != nil {
			log.Printf("AWS Compute: Warning - failed to add name tag to instance %s: %v",
				aws.ToString(inst.InstanceId), tagErr)
		}
	}

	return vm, nil
}

// addNameTag adds a Name tag to an EC2 instance
func (c *AWSCompute) addNameTag(ctx context.Context, instanceID, name string) error {
	input := &ec2.CreateTagsInput{
		Resources: []string{instanceID},
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
		},
	}

	_, err := c.client.CreateTags(ctx, input)
	return err
}

// ListVMs lists all virtual machines
func (c *AWSCompute) ListVMs(ctx context.Context) ([]*services.VM, error) {
	input := &ec2.DescribeInstancesInput{}

	logRequest("DescribeInstances", input, c.debug)

	var resp *ec2.DescribeInstancesOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, c.retryConfig, func() error {
		resp, err = c.client.DescribeInstances(ctx, input)
		return err
	})

	logResponse("DescribeInstances", resp, retryErr, c.debug)

	if retryErr != nil {
		return nil, wrapAWSError(retryErr, "aws", "compute", "ListVMs")
	}

	var vms []*services.VM
	for _, res := range resp.Reservations {
		for _, inst := range res.Instances {
			vm := &services.VM{
				ID:    aws.ToString(inst.InstanceId),
				State: string(inst.State.Name),
			}

			// Safely handle optional fields
			if inst.PublicIpAddress != nil {
				vm.PublicIP = aws.ToString(inst.PublicIpAddress)
			}
			if inst.PrivateIpAddress != nil {
				vm.PrivateIP = aws.ToString(inst.PrivateIpAddress)
			}
			if inst.LaunchTime != nil {
				vm.LaunchTime = inst.LaunchTime.String()
			}

			// Get name from tags
			for _, tag := range inst.Tags {
				if aws.ToString(tag.Key) == "Name" {
					vm.Name = aws.ToString(tag.Value)
					break
				}
			}

			vms = append(vms, vm)
		}
	}

	return vms, nil
}

// GetVM gets a specific virtual machine by ID
func (c *AWSCompute) GetVM(ctx context.Context, id string) (*services.VM, error) {
	// Validate input
	if id == "" {
		return nil, cloudsdk.NewInvalidConfigError("aws", "compute", "id", "instance ID cannot be empty")
	}

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{id},
	}

	logRequest("DescribeInstances", input, c.debug)

	var resp *ec2.DescribeInstancesOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, c.retryConfig, func() error {
		resp, err = c.client.DescribeInstances(ctx, input)
		return err
	})

	logResponse("DescribeInstances", resp, retryErr, c.debug)

	if retryErr != nil {
		return nil, wrapAWSError(retryErr, "aws", "compute", "GetVM")
	}

	if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
		return nil, cloudsdk.NewResourceNotFoundError("aws", "compute", "instance", id)
	}

	inst := resp.Reservations[0].Instances[0]
	vm := &services.VM{
		ID:    aws.ToString(inst.InstanceId),
		State: string(inst.State.Name),
	}

	// Safely handle optional fields
	if inst.PublicIpAddress != nil {
		vm.PublicIP = aws.ToString(inst.PublicIpAddress)
	}
	if inst.PrivateIpAddress != nil {
		vm.PrivateIP = aws.ToString(inst.PrivateIpAddress)
	}
	if inst.LaunchTime != nil {
		vm.LaunchTime = inst.LaunchTime.String()
	}

	// Get name from tags
	for _, tag := range inst.Tags {
		if aws.ToString(tag.Key) == "Name" {
			vm.Name = aws.ToString(tag.Value)
			break
		}
	}

	return vm, nil
}

func (c *AWSCompute) StartVM(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		return cloudsdk.NewInvalidConfigError("aws", "compute", "id", "instance ID cannot be empty")
	}

	input := &ec2.StartInstancesInput{
		InstanceIds: []string{id},
	}

	logRequest("StartInstances", input, c.debug)

	var resp *ec2.StartInstancesOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, c.retryConfig, func() error {
		resp, err = c.client.StartInstances(ctx, input)
		return err
	})

	logResponse("StartInstances", resp, retryErr, c.debug)

	if retryErr != nil {
		return wrapAWSError(retryErr, "aws", "compute", "StartVM")
	}

	return nil
}

func (c *AWSCompute) StopVM(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		return cloudsdk.NewInvalidConfigError("aws", "compute", "id", "instance ID cannot be empty")
	}

	input := &ec2.StopInstancesInput{
		InstanceIds: []string{id},
	}

	logRequest("StopInstances", input, c.debug)

	var resp *ec2.StopInstancesOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, c.retryConfig, func() error {
		resp, err = c.client.StopInstances(ctx, input)
		return err
	})

	logResponse("StopInstances", resp, retryErr, c.debug)

	if retryErr != nil {
		return wrapAWSError(retryErr, "aws", "compute", "StopVM")
	}

	return nil
}

func (c *AWSCompute) DeleteVM(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		return cloudsdk.NewInvalidConfigError("aws", "compute", "id", "instance ID cannot be empty")
	}

	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{id},
	}

	logRequest("TerminateInstances", input, c.debug)

	var resp *ec2.TerminateInstancesOutput
	var err error

	// Execute with retry logic
	retryErr := retryWithBackoff(ctx, c.retryConfig, func() error {
		resp, err = c.client.TerminateInstances(ctx, input)
		return err
	})

	logResponse("TerminateInstances", resp, retryErr, c.debug)

	if retryErr != nil {
		return wrapAWSError(retryErr, "aws", "compute", "DeleteVM")
	}

	return nil
}

// InstanceTypes returns the instance types service
func (c *AWSCompute) InstanceTypes() services.InstanceTypesService {
	return c.instanceTypesSvc
}

// PlacementGroups returns the placement groups service
func (c *AWSCompute) PlacementGroups() services.PlacementGroupsService {
	return c.placementGroupsSvc
}

// SpotInstances returns the spot instances service
func (c *AWSCompute) SpotInstances() services.SpotInstancesService {
	return c.spotInstancesSvc
}

// InstanceTypesServiceImpl implements InstanceTypesService
type InstanceTypesServiceImpl struct {
	client EC2ClientInterface
	debug  bool
}

// List returns a list of instance types based on filters
func (s *InstanceTypesServiceImpl) List(ctx context.Context, filter *services.InstanceTypeFilter) ([]*services.InstanceType, error) {
	input := &ec2.DescribeInstanceTypesInput{}

	if filter != nil {
		if len(filter.InstanceTypes) > 0 {
			input.InstanceTypes = make([]types.InstanceType, len(filter.InstanceTypes))
			for i, it := range filter.InstanceTypes {
				input.InstanceTypes[i] = types.InstanceType(it)
			}
		}
		// Note: AWS SDK doesn't support direct filtering by vCPU, memory, etc.
		// We'll need to filter the results after fetching
	}

	logRequest("DescribeInstanceTypes", input, s.debug)

	resp, err := s.client.DescribeInstanceTypes(ctx, input)

	logResponse("DescribeInstanceTypes", resp, err, s.debug)

	if err != nil {
		return nil, wrapAWSError(err, "aws", "compute", "ListInstanceTypes")
	}

	var instanceTypes []*services.InstanceType
	for _, it := range resp.InstanceTypes {
		instanceType := &services.InstanceType{
			InstanceType:       string(it.InstanceType),
			VCpus:              aws.ToInt32(it.VCpuInfo.DefaultVCpus),
			MemoryGB:           float64(aws.ToInt64(it.MemoryInfo.SizeInMiB)) / 1024,
			NetworkPerformance: aws.ToString(it.NetworkInfo.NetworkPerformance),
			CurrentGeneration:  aws.ToBool(it.CurrentGeneration),
		}

		// Calculate storage
		if it.InstanceStorageInfo != nil && len(it.InstanceStorageInfo.Disks) > 0 {
			for _, disk := range it.InstanceStorageInfo.Disks {
				if disk.SizeInGB != nil {
					instanceType.StorageGB += int32(aws.ToInt64(disk.SizeInGB))
				}
			}
		}

		// Apply filters
		if filter != nil {
			if filter.VCpus != nil && instanceType.VCpus != *filter.VCpus {
				continue
			}
			if filter.MemoryGB != nil && instanceType.MemoryGB < *filter.MemoryGB {
				continue
			}
			if filter.StorageGB != nil && instanceType.StorageGB < *filter.StorageGB {
				continue
			}
			if filter.NetworkPerf != nil && instanceType.NetworkPerformance != *filter.NetworkPerf {
				continue
			}
		}

		instanceTypes = append(instanceTypes, instanceType)
	}

	return instanceTypes, nil
}

// PlacementGroupsServiceImpl implements PlacementGroupsService
type PlacementGroupsServiceImpl struct {
	client EC2ClientInterface
	debug  bool
}

// Create creates a new placement group
func (s *PlacementGroupsServiceImpl) Create(ctx context.Context, config *services.PlacementGroupConfig) (*services.PlacementGroup, error) {
	// Validate input
	if config == nil {
		return nil, cloudsdk.NewInvalidConfigError("aws", "compute", "config", "placement group configuration cannot be nil")
	}
	if config.GroupName == "" {
		return nil, cloudsdk.NewInvalidConfigError("aws", "compute", "GroupName", "group name is required")
	}
	if config.Strategy == "" {
		return nil, cloudsdk.NewInvalidConfigError("aws", "compute", "Strategy", "strategy is required")
	}

	input := &ec2.CreatePlacementGroupInput{
		GroupName: aws.String(config.GroupName),
		Strategy:  types.PlacementStrategy(config.Strategy),
	}

	logRequest("CreatePlacementGroup", input, s.debug)

	_, err := s.client.CreatePlacementGroup(ctx, input)

	logResponse("CreatePlacementGroup", nil, err, s.debug)

	if err != nil {
		return nil, wrapAWSError(err, "aws", "compute", "CreatePlacementGroup")
	}

	// Describe to get the created placement group details
	describeInput := &ec2.DescribePlacementGroupsInput{
		GroupNames: []string{config.GroupName},
	}

	logRequest("DescribePlacementGroups", describeInput, s.debug)

	resp, err := s.client.DescribePlacementGroups(ctx, describeInput)

	logResponse("DescribePlacementGroups", resp, err, s.debug)

	if err != nil {
		return nil, wrapAWSError(err, "aws", "compute", "CreatePlacementGroup")
	}

	if len(resp.PlacementGroups) == 0 {
		return nil, cloudsdk.NewResourceNotFoundError("aws", "compute", "placement group", config.GroupName).
			WithSuggestions("The placement group may not have been created successfully", "Try creating the placement group again")
	}

	pg := resp.PlacementGroups[0]
	return &services.PlacementGroup{
		GroupName: aws.ToString(pg.GroupName),
		GroupId:   aws.ToString(pg.GroupId),
		Strategy:  string(pg.Strategy),
		State:     string(pg.State),
		GroupArn:  aws.ToString(pg.GroupArn),
	}, nil
}

// Delete deletes a placement group
func (s *PlacementGroupsServiceImpl) Delete(ctx context.Context, groupName string) error {
	// Validate input
	if groupName == "" {
		return cloudsdk.NewInvalidConfigError("aws", "compute", "groupName", "group name cannot be empty")
	}

	input := &ec2.DeletePlacementGroupInput{
		GroupName: aws.String(groupName),
	}

	logRequest("DeletePlacementGroup", input, s.debug)

	_, err := s.client.DeletePlacementGroup(ctx, input)

	logResponse("DeletePlacementGroup", nil, err, s.debug)

	if err != nil {
		return wrapAWSError(err, "aws", "compute", "DeletePlacementGroup")
	}

	return nil
}

// List returns a list of placement groups
func (s *PlacementGroupsServiceImpl) List(ctx context.Context) ([]*services.PlacementGroup, error) {
	input := &ec2.DescribePlacementGroupsInput{}

	logRequest("DescribePlacementGroups", input, s.debug)

	resp, err := s.client.DescribePlacementGroups(ctx, input)

	logResponse("DescribePlacementGroups", resp, err, s.debug)

	if err != nil {
		return nil, wrapAWSError(err, "aws", "compute", "ListPlacementGroups")
	}

	var placementGroups []*services.PlacementGroup
	for _, pg := range resp.PlacementGroups {
		placementGroups = append(placementGroups, &services.PlacementGroup{
			GroupName: aws.ToString(pg.GroupName),
			GroupId:   aws.ToString(pg.GroupId),
			Strategy:  string(pg.Strategy),
			State:     string(pg.State),
			GroupArn:  aws.ToString(pg.GroupArn),
		})
	}

	return placementGroups, nil
}

// SpotInstancesServiceImpl implements SpotInstancesService
type SpotInstancesServiceImpl struct {
	client EC2ClientInterface
	debug  bool
}

// Request requests spot instances
func (s *SpotInstancesServiceImpl) Request(ctx context.Context, config *services.SpotInstanceConfig) (*services.SpotInstanceRequest, error) {
	// Validate input
	if config == nil {
		return nil, cloudsdk.NewInvalidConfigError("aws", "compute", "config", "spot instance configuration cannot be nil")
	}
	if config.ImageID == "" {
		return nil, cloudsdk.NewInvalidConfigError("aws", "compute", "ImageID", "image ID is required")
	}
	if config.InstanceType == "" {
		return nil, cloudsdk.NewInvalidConfigError("aws", "compute", "InstanceType", "instance type is required")
	}

	input := &ec2.RequestSpotInstancesInput{
		LaunchSpecification: &types.RequestSpotLaunchSpecification{
			ImageId:      aws.String(config.ImageID),
			InstanceType: types.InstanceType(config.InstanceType),
		},
	}

	if config.SpotPrice != nil {
		input.SpotPrice = config.SpotPrice
	}

	if config.AvailabilityZone != nil {
		input.AvailabilityZoneGroup = config.AvailabilityZone
	}

	if config.LaunchSpecification != nil {
		spec := config.LaunchSpecification
		if spec.KeyName != "" {
			input.LaunchSpecification.KeyName = aws.String(spec.KeyName)
		}
		if len(spec.SecurityGroups) > 0 {
			input.LaunchSpecification.SecurityGroupIds = spec.SecurityGroups
		}
		if spec.UserData != "" {
			input.LaunchSpecification.UserData = aws.String(spec.UserData)
		}
	}

	logRequest("RequestSpotInstances", input, s.debug)

	resp, err := s.client.RequestSpotInstances(ctx, input)

	logResponse("RequestSpotInstances", resp, err, s.debug)

	if err != nil {
		return nil, wrapAWSError(err, "aws", "compute", "RequestSpotInstances")
	}

	if len(resp.SpotInstanceRequests) == 0 {
		return nil, cloudsdk.NewCloudError(cloudsdk.ErrProviderError, "No spot instance requests were created", "aws", "compute", "RequestSpotInstances").
			WithSuggestions(
				"Check AWS service status",
				"Verify your spot price is competitive",
				"Try again with different parameters",
			)
	}

	req := resp.SpotInstanceRequests[0]
	result := &services.SpotInstanceRequest{
		SpotInstanceRequestId: aws.ToString(req.SpotInstanceRequestId),
		State:                 string(req.State),
		Status:                aws.ToString(req.Status.Code),
		SpotPrice:             aws.ToString(req.SpotPrice),
	}

	if req.CreateTime != nil {
		result.CreateTime = aws.ToTime(req.CreateTime).String()
	}

	return result, nil
}

// Describe describes spot instance requests
func (s *SpotInstancesServiceImpl) Describe(ctx context.Context, requestIds []string) ([]*services.SpotInstanceRequest, error) {
	input := &ec2.DescribeSpotInstanceRequestsInput{}

	if len(requestIds) > 0 {
		input.SpotInstanceRequestIds = requestIds
	}

	logRequest("DescribeSpotInstanceRequests", input, s.debug)

	resp, err := s.client.DescribeSpotInstanceRequests(ctx, input)

	logResponse("DescribeSpotInstanceRequests", resp, err, s.debug)

	if err != nil {
		return nil, wrapAWSError(err, "aws", "compute", "DescribeSpotInstanceRequests")
	}

	var requests []*services.SpotInstanceRequest
	for _, req := range resp.SpotInstanceRequests {
		request := &services.SpotInstanceRequest{
			SpotInstanceRequestId: aws.ToString(req.SpotInstanceRequestId),
			State:                 string(req.State),
			Status:                aws.ToString(req.Status.Code),
			SpotPrice:             aws.ToString(req.SpotPrice),
		}

		if req.CreateTime != nil {
			request.CreateTime = aws.ToTime(req.CreateTime).String()
		}

		if req.InstanceId != nil {
			request.InstanceId = aws.ToString(req.InstanceId)
		}

		requests = append(requests, request)
	}

	return requests, nil
}

// Cancel cancels spot instance requests
func (s *SpotInstancesServiceImpl) Cancel(ctx context.Context, requestId string) error {
	// Validate input
	if requestId == "" {
		return cloudsdk.NewInvalidConfigError("aws", "compute", "requestId", "request ID cannot be empty")
	}

	input := &ec2.CancelSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []string{requestId},
	}

	logRequest("CancelSpotInstanceRequests", input, s.debug)

	_, err := s.client.CancelSpotInstanceRequests(ctx, input)

	logResponse("CancelSpotInstanceRequests", nil, err, s.debug)

	if err != nil {
		return wrapAWSError(err, "aws", "compute", "CancelSpotInstanceRequests")
	}

	return nil
}
