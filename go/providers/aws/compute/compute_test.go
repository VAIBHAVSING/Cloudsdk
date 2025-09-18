package compute

import (
	"context"
	"fmt"
	"testing"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	cloudsdktesting "github.com/VAIBHAVSING/Cloudsdk/go/testing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// mockEC2Client is a mock implementation of the EC2 client
type mockEC2Client struct {
	runInstancesResponse                 *ec2.RunInstancesOutput
	runInstancesError                    error
	describeInstancesResponse            *ec2.DescribeInstancesOutput
	describeInstancesError               error
	startInstancesResponse               *ec2.StartInstancesOutput
	startInstancesError                  error
	stopInstancesResponse                *ec2.StopInstancesOutput
	stopInstancesError                   error
	terminateInstancesResponse           *ec2.TerminateInstancesOutput
	terminateInstancesError              error
	describeInstanceTypesResponse        *ec2.DescribeInstanceTypesOutput
	describeInstanceTypesError           error
	describePlacementGroupsResponse      *ec2.DescribePlacementGroupsOutput
	describePlacementGroupsError         error
	createPlacementGroupResponse         *ec2.CreatePlacementGroupOutput
	createPlacementGroupError            error
	deletePlacementGroupResponse         *ec2.DeletePlacementGroupOutput
	deletePlacementGroupError            error
	requestSpotInstancesResponse         *ec2.RequestSpotInstancesOutput
	requestSpotInstancesError            error
	describeSpotInstanceRequestsResponse *ec2.DescribeSpotInstanceRequestsOutput
	describeSpotInstanceRequestsError    error
	cancelSpotInstanceRequestsResponse   *ec2.CancelSpotInstanceRequestsOutput
	cancelSpotInstanceRequestsError      error
}

// CreateTags implements EC2ClientInterface.
func (m *mockEC2Client) CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return &ec2.CreateTagsOutput{}, nil
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
	helper := cloudsdktesting.NewTestHelper(t)

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

	compute := NewWithClient(mockClient)
	config := cloudsdktesting.GenerateVMConfig("test-vm")

	vm, err := compute.CreateVM(context.Background(), config)

	helper.AssertNoError(err)
	cloudsdktesting.AssertVMValid(t, vm)
	helper.AssertEqual("i-1234567890abcdef0", vm.ID)
	helper.AssertEqual("running", vm.State)
	helper.AssertEqual("1.2.3.4", vm.PublicIP)
	helper.AssertEqual("10.0.0.1", vm.PrivateIP)
}

func TestAWSCompute_CreateVM_ErrorScenarios(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	testCases := []struct {
		name          string
		mockError     error
		expectedError cloudsdk.ErrorCode
	}{
		{
			name:          "authentication error",
			mockError:     fmt.Errorf("UnauthorizedOperation: You are not authorized to perform this operation"),
			expectedError: cloudsdk.ErrAuthentication,
		},
		{
			name:          "invalid image error",
			mockError:     fmt.Errorf("InvalidAMIID.NotFound: The image id '[ami-12345]' does not exist"),
			expectedError: cloudsdk.ErrResourceNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockEC2Client{
				runInstancesError: tc.mockError,
			}

			compute := NewWithClient(mockClient)
			config := cloudsdktesting.GenerateVMConfig("test-vm")

			_, err := compute.CreateVM(context.Background(), config)
			helper.AssertError(err)
		})
	}
}

func TestAWSCompute_ListVMs(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

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

	helper.AssertNoError(err)
	helper.AssertEqual(1, len(vms))
	cloudsdktesting.AssertVMValid(t, vms[0])
	helper.AssertEqual("i-1234567890abcdef0", vms[0].ID)
	helper.AssertEqual("running", vms[0].State)
	helper.AssertEqual("test-vm", vms[0].Name)
}

func TestAWSCompute_GetVM(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

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

	helper.AssertNoError(err)
	cloudsdktesting.AssertVMValid(t, vm)
	helper.AssertEqual("i-1234567890abcdef0", vm.ID)
	helper.AssertEqual("running", vm.State)
	helper.AssertEqual("test-vm", vm.Name)
}

func TestAWSCompute_VMLifecycle(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	mockClient := &mockEC2Client{
		startInstancesResponse:     &ec2.StartInstancesOutput{},
		startInstancesError:        nil,
		stopInstancesResponse:      &ec2.StopInstancesOutput{},
		stopInstancesError:         nil,
		terminateInstancesResponse: &ec2.TerminateInstancesOutput{},
		terminateInstancesError:    nil,
	}

	compute := NewWithClient(mockClient)
	vmID := "i-1234567890abcdef0"

	// Test start VM
	err := compute.StartVM(context.Background(), vmID)
	helper.AssertNoError(err)

	// Test stop VM
	err = compute.StopVM(context.Background(), vmID)
	helper.AssertNoError(err)

	// Test delete VM
	err = compute.DeleteVM(context.Background(), vmID)
	helper.AssertNoError(err)
}

func TestAWSCompute_InstanceTypes(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

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

	helper.AssertNoError(err)
	helper.AssertEqual(1, len(instanceTypes))
	helper.AssertEqual("t2.micro", instanceTypes[0].InstanceType)
	helper.AssertEqual(int32(1), instanceTypes[0].VCpus)
	helper.AssertEqual(1.0, instanceTypes[0].MemoryGB)
}

func TestAWSCompute_PlacementGroups(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

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

	// Test create placement group
	config := &services.PlacementGroupConfig{
		GroupName: "test-pg",
		Strategy:  "cluster",
	}

	pg, err := compute.PlacementGroups().Create(context.Background(), config)
	helper.AssertNoError(err)
	helper.AssertNotEqual(nil, pg)
	helper.AssertEqual("test-pg", pg.GroupName)
	helper.AssertEqual("cluster", pg.Strategy)

	// Test list placement groups
	placementGroups, err := compute.PlacementGroups().List(context.Background())
	helper.AssertNoError(err)
	helper.AssertEqual(1, len(placementGroups))
	helper.AssertEqual("test-pg", placementGroups[0].GroupName)
	helper.AssertEqual("cluster", placementGroups[0].Strategy)
}

func TestAWSCompute_SpotInstances(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	mockClient := &mockEC2Client{
		requestSpotInstancesResponse: &ec2.RequestSpotInstancesOutput{
			SpotInstanceRequests: []types.SpotInstanceRequest{
				{
					SpotInstanceRequestId: aws.String("sir-12345"),
					State:                 types.SpotInstanceState("open"),
					Status: &types.SpotInstanceStatus{
						Code: aws.String("fulfilled"),
					},
					SpotPrice:  aws.String("0.01"),
					CreateTime: aws.Time(time.Now()),
				},
			},
		},
		requestSpotInstancesError: nil,
		describeSpotInstanceRequestsResponse: &ec2.DescribeSpotInstanceRequestsOutput{
			SpotInstanceRequests: []types.SpotInstanceRequest{
				{
					SpotInstanceRequestId: aws.String("sir-12345"),
					State:                 types.SpotInstanceState("active"),
					Status: &types.SpotInstanceStatus{
						Code: aws.String("fulfilled"),
					},
					SpotPrice:  aws.String("0.01"),
					CreateTime: aws.Time(time.Now()),
					InstanceId: aws.String("i-12345"),
				},
			},
		},
		describeSpotInstanceRequestsError:  nil,
		cancelSpotInstanceRequestsResponse: &ec2.CancelSpotInstanceRequestsOutput{},
		cancelSpotInstanceRequestsError:    nil,
	}

	compute := NewWithClient(mockClient)

	// Test request spot instance
	config := &services.SpotInstanceConfig{
		InstanceType: "t2.micro",
		ImageID:      "ami-12345",
	}

	request, err := compute.SpotInstances().Request(context.Background(), config)
	helper.AssertNoError(err)
	helper.AssertNotEqual(nil, request)
	helper.AssertEqual("sir-12345", request.SpotInstanceRequestId)
	helper.AssertEqual("open", request.State)

	// Test describe spot instance requests
	requests, err := compute.SpotInstances().Describe(context.Background(), []string{"sir-12345"})
	helper.AssertNoError(err)
	helper.AssertEqual(1, len(requests))
	helper.AssertEqual("sir-12345", requests[0].SpotInstanceRequestId)
	helper.AssertEqual("i-12345", requests[0].InstanceId)

	// Test cancel spot instance request
	err = compute.SpotInstances().Cancel(context.Background(), "sir-12345")
	helper.AssertNoError(err)
}

func TestAWSCompute_ConcurrentOperations(t *testing.T) {
	mockClient := &mockEC2Client{
		describeInstancesResponse: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{},
		},
		describeInstancesError: nil,
	}

	compute := NewWithClient(mockClient)

	// Test concurrent ListVMs calls
	cloudsdktesting.TestConcurrency(t, 10, func(id int) error {
		_, err := compute.ListVMs(context.Background())
		return err
	})
}

func BenchmarkAWSCompute_ListVMs(b *testing.B) {
	mockClient := &mockEC2Client{
		describeInstancesResponse: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{},
		},
		describeInstancesError: nil,
	}

	compute := NewWithClient(mockClient)

	cloudsdktesting.BenchmarkOperation(b, func() error {
		_, err := compute.ListVMs(context.Background())
		return err
	})
}

func TestAWSCompute_PerformanceLatency(t *testing.T) {
	mockClient := &mockEC2Client{
		describeInstancesResponse: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{},
		},
		describeInstancesError: nil,
	}

	compute := NewWithClient(mockClient)

	// Assert that ListVMs completes within reasonable time
	cloudsdktesting.AssertLatencyUnder(t, 100*time.Millisecond, func() error {
		_, err := compute.ListVMs(context.Background())
		return err
	})
}

func stringPtr(s string) *string {
	return &s
}
