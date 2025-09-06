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
	runInstancesResponse *ec2.RunInstancesOutput
	runInstancesError    error
	describeInstancesResponse *ec2.DescribeInstancesOutput
	describeInstancesError    error
}

func (m *mockEC2Client) RunInstances(ctx context.Context, input *ec2.RunInstancesInput, opts ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	return m.runInstancesResponse, m.runInstancesError
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, input *ec2.DescribeInstancesInput, opts ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.describeInstancesResponse, m.describeInstancesError
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

func stringPtr(s string) *string {
	return &s
}
