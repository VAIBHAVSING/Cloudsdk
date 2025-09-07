package compute

import (
	"context"
	"fmt"

	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2ClientInterface defines methods we need from EC2 client for testing
type EC2ClientInterface interface {
	RunInstances(ctx context.Context, input *ec2.RunInstancesInput, opts ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	DescribeInstances(ctx context.Context, input *ec2.DescribeInstancesInput, opts ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	StartInstances(ctx context.Context, input *ec2.StartInstancesInput, opts ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	StopInstances(ctx context.Context, input *ec2.StopInstancesInput, opts ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	TerminateInstances(ctx context.Context, input *ec2.TerminateInstancesInput, opts ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
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
	client EC2ClientInterface
	instanceTypesSvc    *InstanceTypesServiceImpl
	placementGroupsSvc  *PlacementGroupsServiceImpl
	spotInstancesSvc    *SpotInstancesServiceImpl
}

// New creates a new AWSCompute instance with real AWS client
func New(cfg aws.Config) services.Compute {
	client := ec2.NewFromConfig(cfg)
	return &AWSCompute{
		client: client,
		instanceTypesSvc:    &InstanceTypesServiceImpl{client: client},
		placementGroupsSvc:  &PlacementGroupsServiceImpl{client: client},
		spotInstancesSvc:    &SpotInstancesServiceImpl{client: client},
	}
}

// NewWithClient creates a new AWSCompute instance with custom client (for testing)
func NewWithClient(client EC2ClientInterface) services.Compute {
	return &AWSCompute{
		client: client,
		instanceTypesSvc:    &InstanceTypesServiceImpl{client: client},
		placementGroupsSvc:  &PlacementGroupsServiceImpl{client: client},
		spotInstancesSvc:    &SpotInstancesServiceImpl{client: client},
	}
}

// CreateVM creates a new virtual machine
func (c *AWSCompute) CreateVM(ctx context.Context, config *services.VMConfig) (*services.VM, error) {
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(config.ImageID),
		InstanceType: types.InstanceType(config.InstanceType),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		KeyName:      aws.String(config.KeyName),
		UserData:     aws.String(config.UserData),
	}

	if len(config.SecurityGroups) > 0 {
		input.SecurityGroupIds = make([]string, len(config.SecurityGroups))
		copy(input.SecurityGroupIds, config.SecurityGroups)
	}

	resp, err := c.client.RunInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 instance: %w", err)
	}

	if len(resp.Instances) == 0 {
		return nil, fmt.Errorf("no instances created")
	}

	inst := resp.Instances[0]
	return &services.VM{
		ID:         aws.ToString(inst.InstanceId),
		Name:       config.Name, // AWS doesn't set name here, but we can tag later
		State:      string(inst.State.Name),
		PublicIP:   aws.ToString(inst.PublicIpAddress),
		PrivateIP:  aws.ToString(inst.PrivateIpAddress),
		LaunchTime: inst.LaunchTime.String(),
	}, nil
}

// ListVMs lists all virtual machines
func (c *AWSCompute) ListVMs(ctx context.Context) ([]*services.VM, error) {
	resp, err := c.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	var vms []*services.VM
	for _, res := range resp.Reservations {
		for _, inst := range res.Instances {
			vm := &services.VM{
				ID:         aws.ToString(inst.InstanceId),
				State:      string(inst.State.Name),
				PublicIP:   aws.ToString(inst.PublicIpAddress),
				PrivateIP:  aws.ToString(inst.PrivateIpAddress),
				LaunchTime: inst.LaunchTime.String(),
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
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{id},
	}

	resp, err := c.client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("instance not found")
	}

	inst := resp.Reservations[0].Instances[0]
	vm := &services.VM{
		ID:         aws.ToString(inst.InstanceId),
		State:      string(inst.State.Name),
		PublicIP:   aws.ToString(inst.PublicIpAddress),
		PrivateIP:  aws.ToString(inst.PrivateIpAddress),
		LaunchTime: inst.LaunchTime.String(),
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
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{id},
	}

	_, err := c.client.StartInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	return nil
}

func (c *AWSCompute) StopVM(ctx context.Context, id string) error {
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{id},
	}

	_, err := c.client.StopInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	return nil
}

func (c *AWSCompute) DeleteVM(ctx context.Context, id string) error {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{id},
	}

	_, err := c.client.TerminateInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to terminate instance: %w", err)
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

	resp, err := s.client.DescribeInstanceTypes(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance types: %w", err)
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
}

// Create creates a new placement group
func (s *PlacementGroupsServiceImpl) Create(ctx context.Context, config *services.PlacementGroupConfig) (*services.PlacementGroup, error) {
	input := &ec2.CreatePlacementGroupInput{
		GroupName:         aws.String(config.GroupName),
		Strategy:          types.PlacementStrategy(config.Strategy),
	}

	_, err := s.client.CreatePlacementGroup(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create placement group: %w", err)
	}

	// Describe to get the created placement group details
	describeInput := &ec2.DescribePlacementGroupsInput{
		GroupNames: []string{config.GroupName},
	}

	resp, err := s.client.DescribePlacementGroups(ctx, describeInput)
	if err != nil {
		return nil, fmt.Errorf("failed to describe placement group: %w", err)
	}

	if len(resp.PlacementGroups) == 0 {
		return nil, fmt.Errorf("placement group not found after creation")
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
	input := &ec2.DeletePlacementGroupInput{
		GroupName: aws.String(groupName),
	}

	_, err := s.client.DeletePlacementGroup(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete placement group: %w", err)
	}

	return nil
}

// List returns a list of placement groups
func (s *PlacementGroupsServiceImpl) List(ctx context.Context) ([]*services.PlacementGroup, error) {
	input := &ec2.DescribePlacementGroupsInput{}

	resp, err := s.client.DescribePlacementGroups(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe placement groups: %w", err)
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
}

// Request requests spot instances
func (s *SpotInstancesServiceImpl) Request(ctx context.Context, config *services.SpotInstanceConfig) (*services.SpotInstanceRequest, error) {
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
		input.LaunchSpecification.KeyName = aws.String(spec.KeyName)
		if len(spec.SecurityGroups) > 0 {
			input.LaunchSpecification.SecurityGroupIds = spec.SecurityGroups
		}
		input.LaunchSpecification.UserData = aws.String(spec.UserData)
	}

	resp, err := s.client.RequestSpotInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to request spot instances: %w", err)
	}

	if len(resp.SpotInstanceRequests) == 0 {
		return nil, fmt.Errorf("no spot instance requests created")
	}

	req := resp.SpotInstanceRequests[0]
	return &services.SpotInstanceRequest{
		SpotInstanceRequestId: aws.ToString(req.SpotInstanceRequestId),
		State:                 string(req.State),
		Status:                aws.ToString(req.Status.Code),
		SpotPrice:             aws.ToString(req.SpotPrice),
		CreateTime:            aws.ToTime(req.CreateTime).String(),
	}, nil
}

// Describe describes spot instance requests
func (s *SpotInstancesServiceImpl) Describe(ctx context.Context, requestIds []string) ([]*services.SpotInstanceRequest, error) {
	input := &ec2.DescribeSpotInstanceRequestsInput{}

	if len(requestIds) > 0 {
		input.SpotInstanceRequestIds = requestIds
	}

	resp, err := s.client.DescribeSpotInstanceRequests(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe spot instance requests: %w", err)
	}

	var requests []*services.SpotInstanceRequest
	for _, req := range resp.SpotInstanceRequests {
		request := &services.SpotInstanceRequest{
			SpotInstanceRequestId: aws.ToString(req.SpotInstanceRequestId),
			State:                 string(req.State),
			Status:                aws.ToString(req.Status.Code),
			SpotPrice:             aws.ToString(req.SpotPrice),
			CreateTime:            aws.ToTime(req.CreateTime).String(),
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
	input := &ec2.CancelSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []string{requestId},
	}

	_, err := s.client.CancelSpotInstanceRequests(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to cancel spot instance request: %w", err)
	}

	return nil
}
