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
	CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	DeleteTags(ctx context.Context, input *ec2.DeleteTagsInput, opts ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
	DescribeTags(ctx context.Context, input *ec2.DescribeTagsInput, opts ...func(*ec2.Options)) (*ec2.DescribeTagsOutput, error)
}

// AWSCompute implements the Compute interface for AWS
type AWSCompute struct {
	client EC2ClientInterface
}

// New creates a new AWSCompute instance with real AWS client
func New(cfg aws.Config) services.Compute {
	client := ec2.NewFromConfig(cfg)
	return &AWSCompute{client: client}
}

// NewWithClient creates a new AWSCompute instance with custom client (for testing)
func NewWithClient(client EC2ClientInterface) services.Compute {
	return &AWSCompute{client: client}
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

// Tags returns the tagging service for compute resources
func (c *AWSCompute) Tags() services.Tagging {
	return NewTagging(c.client)
}
