package compute

import (
	"context"
	"testing"
	"time"

	"github.com/VAIBHAVSING/Cloudsdk/go/services"
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
}

func (m *mockEC2Client) RunInstances(ctx context.Context, input *ec2.RunInstancesInput, opts ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	return m.runInstancesResponse, m.runInstancesError
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, input *ec2.DescribeInstancesInput, opts ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.describeInstancesResponse, m.describeInstancesError
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
