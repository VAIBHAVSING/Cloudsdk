package compute

import (
	"context"
	"testing"
	"time"

	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

// mockEC2Client is a mock implementation of the EC2 client
type mockEC2Client struct {
	runInstancesResponse       *ec2.RunInstancesOutput
	runInstancesError          error
	describeInstancesResponse  *ec2.DescribeInstancesOutput
	describeInstancesError     error
	startInstancesResponse     *ec2.StartInstancesOutput
	startInstancesError        error
	stopInstancesResponse      *ec2.StopInstancesOutput
	stopInstancesError         error
	terminateInstancesResponse *ec2.TerminateInstancesOutput
	terminateInstancesError    error
	describeInstanceTypesResponse *ec2.DescribeInstanceTypesOutput
	describeInstanceTypesError    error
	describePlacementGroupsResponse *ec2.DescribePlacementGroupsOutput
	describePlacementGroupsError    error
	createPlacementGroupResponse *ec2.CreatePlacementGroupOutput
	createPlacementGroupError    error
	deletePlacementGroupResponse *ec2.DeletePlacementGroupOutput
	deletePlacementGroupError    error
	requestSpotInstancesResponse *ec2.RequestSpotInstancesOutput
	requestSpotInstancesError    error
	describeSpotInstanceRequestsResponse *ec2.DescribeSpotInstanceRequestsOutput
	describeSpotInstanceRequestsError    error
	cancelSpotInstanceRequestsResponse *ec2.CancelSpotInstanceRequestsOutput
	cancelSpotInstanceRequestsError    error
}

func (m *mockEC2Client) RunInstances(ctx context.Context, input *ec2.RunInstancesInput, opts ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	return m.runInstancesResponse, m.runInstancesError
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, input *ec2.DescribeInstancesInput, opts ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.describeInstancesResponse, m.describeInstancesError
}

func (m *mockEC2Client) DescribeInstanceTypes(ctx context.Context, input *ec2.DescribeInstanceTypesInput, opts ...func(*ec2.Options)) (*ec2.DescribeInstanceTypesOutput, error) {
	return m.describeInstanceTypesResponse, m.describeInstanceTypesError
}

func (m *mockEC2Client) DescribePlacementGroups(ctx context.Context, input *ec2.DescribePlacementGroupsInput, opts ...func(*ec2.Options)) (*ec2.DescribePlacementGroupsOutput, error) {
	return m.describePlacementGroupsResponse, m.describePlacementGroupsError
}

func (m *mockEC2Client) CreatePlacementGroup(ctx context.Context, input *ec2.CreatePlacementGroupInput, opts ...func(*ec2.Options)) (*ec2.CreatePlacementGroupOutput, error) {
	return m.createPlacementGroupResponse, m.createPlacementGroupError
}

func (m *mockEC2Client) DeletePlacementGroup(ctx context.Context, input *ec2.DeletePlacementGroupInput, opts ...func(*ec2.Options)) (*ec2.DeletePlacementGroupOutput, error) {
	return m.deletePlacementGroupResponse, m.deletePlacementGroupError
}

func (m *mockEC2Client) RequestSpotInstances(ctx context.Context, input *ec2.RequestSpotInstancesInput, opts ...func(*ec2.Options)) (*ec2.RequestSpotInstancesOutput, error) {
	return m.requestSpotInstancesResponse, m.requestSpotInstancesError
}

func (m *mockEC2Client) DescribeSpotInstanceRequests(ctx context.Context, input *ec2.DescribeSpotInstanceRequestsInput, opts ...func(*ec2.Options)) (*ec2.DescribeSpotInstanceRequestsOutput, error) {
	return m.describeSpotInstanceRequestsResponse, m.describeSpotInstanceRequestsError
}

func (m *mockEC2Client) CancelSpotInstanceRequests(ctx context.Context, input *ec2.CancelSpotInstanceRequestsInput, opts ...func(*ec2.Options)) (*ec2.CancelSpotInstanceRequestsOutput, error) {
	return m.cancelSpotInstanceRequestsResponse, m.cancelSpotInstanceRequestsError
}

func (m *mockEC2Client) StartInstances(ctx context.Context, input *ec2.StartInstancesInput, opts ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	return m.startInstancesResponse, m.startInstancesError
}

func (m *mockEC2Client) StopInstances(ctx context.Context, input *ec2.StopInstancesInput, opts ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	return m.stopInstancesResponse, m.stopInstancesError
}

func (m *mockEC2Client) TerminateInstances(ctx context.Context, input *ec2.TerminateInstancesInput, opts ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	return m.terminateInstancesResponse, m.terminateInstancesError
}

func TestAWSCompute_CreateVM(t *testing.T) {
	mockClient := &mockEC2Client{
		runInstancesResponse: &ec2.RunInstancesOutput{
			Instances: []types.Instance{
				{
					InstanceId:       stringPtr("i-1234567890abcdef0"),
					State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
					PublicIpAddress:  stringPtr("1.2.3.4"),
					PrivateIpAddress: stringPtr("10.0.0.1"),
					LaunchTime:       &time.Time{},
				},
			},
		},
		runInstancesError: nil,
	}

	// Test CreateVM through the interface
	compute := NewWithClient(mockClient)

	config := &services.VMConfig{
		Name:         "test-vm",
		ImageID:      "ami-12345",
		InstanceType: "t2.micro",
		KeyName:      "test-key",
	}

	vm, err := compute.CreateVM(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, vm)
	assert.Equal(t, "i-1234567890abcdef0", vm.ID)
	assert.Equal(t, "running", vm.State)
	assert.Equal(t, "1.2.3.4", vm.PublicIP)
	assert.Equal(t, "10.0.0.1", vm.PrivateIP)
}

func TestAWSCompute_ListVMs(t *testing.T) {
	mockClient := &mockEC2Client{
		describeInstancesResponse: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId:       stringPtr("i-1234567890abcdef0"),
							State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
							PublicIpAddress:  stringPtr("1.2.3.4"),
							PrivateIpAddress: stringPtr("10.0.0.1"),
							LaunchTime:       &time.Time{},
							Tags: []types.Tag{
								{Key: stringPtr("Name"), Value: stringPtr("test-vm")},
							},
						},
					},
				},
			},
		},
		describeInstancesError: nil,
	}

	compute := NewWithClient(mockClient)

	vms, err := compute.ListVMs(context.Background())

	assert.NoError(t, err)
	assert.Len(t, vms, 1)
	assert.Equal(t, "i-1234567890abcdef0", vms[0].ID)
	assert.Equal(t, "running", vms[0].State)
	assert.Equal(t, "test-vm", vms[0].Name)
}

func TestAWSCompute_InstanceTypes_List(t *testing.T) {
	mockClient := &mockEC2Client{
		describeInstanceTypesResponse: &ec2.DescribeInstanceTypesOutput{
			InstanceTypes: []types.InstanceTypeInfo{
				{
					InstanceType: types.InstanceType("t2.micro"),
					VCpuInfo: &types.VCpuInfo{
						DefaultVCpus: aws.Int32(1),
					},
					MemoryInfo: &types.MemoryInfo{
						SizeInMiB: aws.Int64(1024),
					},
					NetworkInfo: &types.NetworkInfo{
						NetworkPerformance: aws.String("Low"),
					},
					CurrentGeneration: aws.Bool(true),
				},
			},
		},
		describeInstanceTypesError: nil,
	}

	compute := NewWithClient(mockClient)

	filter := &services.InstanceTypeFilter{
		VCpus: aws.Int32(1),
	}

	instanceTypes, err := compute.InstanceTypes().List(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, instanceTypes, 1)
	assert.Equal(t, "t2.micro", instanceTypes[0].InstanceType)
	assert.Equal(t, int32(1), instanceTypes[0].VCpus)
	assert.Equal(t, 1.0, instanceTypes[0].MemoryGB)
}

func TestAWSCompute_PlacementGroups_List(t *testing.T) {
	mockClient := &mockEC2Client{
		describePlacementGroupsResponse: &ec2.DescribePlacementGroupsOutput{
			PlacementGroups: []types.PlacementGroup{
				{
					GroupName: aws.String("test-pg"),
					GroupId:   aws.String("pg-12345"),
					Strategy:  types.PlacementStrategy("cluster"),
					State:     types.PlacementGroupState("available"),
					GroupArn:  aws.String("arn:aws:ec2:us-east-1:123456789012:placement-group/test-pg"),
				},
			},
		},
		describePlacementGroupsError: nil,
	}

	compute := NewWithClient(mockClient)

	placementGroups, err := compute.PlacementGroups().List(context.Background())

	assert.NoError(t, err)
	assert.Len(t, placementGroups, 1)
	assert.Equal(t, "test-pg", placementGroups[0].GroupName)
	assert.Equal(t, "cluster", placementGroups[0].Strategy)
}

func TestAWSCompute_PlacementGroups_Create(t *testing.T) {
	mockClient := &mockEC2Client{
		createPlacementGroupResponse: &ec2.CreatePlacementGroupOutput{},
		createPlacementGroupError:    nil,
		describePlacementGroupsResponse: &ec2.DescribePlacementGroupsOutput{
			PlacementGroups: []types.PlacementGroup{
				{
					GroupName: aws.String("test-pg"),
					GroupId:   aws.String("pg-12345"),
					Strategy:  types.PlacementStrategy("cluster"),
					State:     types.PlacementGroupState("available"),
					GroupArn:  aws.String("arn:aws:ec2:us-east-1:123456789012:placement-group/test-pg"),
				},
			},
		},
		describePlacementGroupsError: nil,
	}

	compute := NewWithClient(mockClient)

	config := &services.PlacementGroupConfig{
		GroupName: "test-pg",
		Strategy:  "cluster",
	}

	pg, err := compute.PlacementGroups().Create(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, pg)
	assert.Equal(t, "test-pg", pg.GroupName)
	assert.Equal(t, "cluster", pg.Strategy)
}

func TestAWSCompute_SpotInstances_Request(t *testing.T) {
	mockClient := &mockEC2Client{
		requestSpotInstancesResponse: &ec2.RequestSpotInstancesOutput{
			SpotInstanceRequests: []types.SpotInstanceRequest{
				{
					SpotInstanceRequestId: aws.String("sir-12345"),
					State:                 types.SpotInstanceState("open"),
					Status: &types.SpotInstanceStatus{
						Code: aws.String("fulfilled"),
					},
					SpotPrice: aws.String("0.01"),
					CreateTime: aws.Time(time.Now()),
				},
			},
		},
		requestSpotInstancesError: nil,
	}

	compute := NewWithClient(mockClient)

	config := &services.SpotInstanceConfig{
		InstanceType: "t2.micro",
		ImageID:      "ami-12345",
	}

	request, err := compute.SpotInstances().Request(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, request)
	assert.Equal(t, "sir-12345", request.SpotInstanceRequestId)
	assert.Equal(t, "open", request.State)
}

func TestAWSCompute_SpotInstances_Describe(t *testing.T) {
	mockClient := &mockEC2Client{
		describeSpotInstanceRequestsResponse: &ec2.DescribeSpotInstanceRequestsOutput{
			SpotInstanceRequests: []types.SpotInstanceRequest{
				{
					SpotInstanceRequestId: aws.String("sir-12345"),
					State:                 types.SpotInstanceState("active"),
					Status: &types.SpotInstanceStatus{
						Code: aws.String("fulfilled"),
					},
					SpotPrice: aws.String("0.01"),
					CreateTime: aws.Time(time.Now()),
					InstanceId: aws.String("i-12345"),
				},
			},
		},
		describeSpotInstanceRequestsError: nil,
	}

	compute := NewWithClient(mockClient)

	requests, err := compute.SpotInstances().Describe(context.Background(), []string{"sir-12345"})

	assert.NoError(t, err)
	assert.Len(t, requests, 1)
	assert.Equal(t, "sir-12345", requests[0].SpotInstanceRequestId)
	assert.Equal(t, "i-12345", requests[0].InstanceId)
}

func TestAWSCompute_GetVM(t *testing.T) {
	mockClient := &mockEC2Client{
		describeInstancesResponse: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId:       stringPtr("i-1234567890abcdef0"),
							State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
							PublicIpAddress:  stringPtr("1.2.3.4"),
							PrivateIpAddress: stringPtr("10.0.0.1"),
							LaunchTime:       &time.Time{},
							Tags: []types.Tag{
								{Key: stringPtr("Name"), Value: stringPtr("test-vm")},
							},
						},
					},
				},
			},
		},
		describeInstancesError: nil,
	}

	compute := NewWithClient(mockClient)

	vm, err := compute.GetVM(context.Background(), "i-1234567890abcdef0")

	assert.NoError(t, err)
	assert.NotNil(t, vm)
	assert.Equal(t, "i-1234567890abcdef0", vm.ID)
	assert.Equal(t, "running", vm.State)
	assert.Equal(t, "test-vm", vm.Name)
}

func TestAWSCompute_StartVM(t *testing.T) {
	mockClient := &mockEC2Client{
		startInstancesResponse: &ec2.StartInstancesOutput{},
		startInstancesError:    nil,
	}

	compute := NewWithClient(mockClient)

	err := compute.StartVM(context.Background(), "i-1234567890abcdef0")

	assert.NoError(t, err)
}

func TestAWSCompute_StopVM(t *testing.T) {
	mockClient := &mockEC2Client{
		stopInstancesResponse: &ec2.StopInstancesOutput{},
		stopInstancesError:    nil,
	}

	compute := NewWithClient(mockClient)

	err := compute.StopVM(context.Background(), "i-1234567890abcdef0")

	assert.NoError(t, err)
}

func TestAWSCompute_DeleteVM(t *testing.T) {
	mockClient := &mockEC2Client{
		terminateInstancesResponse: &ec2.TerminateInstancesOutput{},
		terminateInstancesError:    nil,
	}

	compute := NewWithClient(mockClient)

	err := compute.DeleteVM(context.Background(), "i-1234567890abcdef0")

	assert.NoError(t, err)
}

func stringPtr(s string) *string {
	return &s
}
