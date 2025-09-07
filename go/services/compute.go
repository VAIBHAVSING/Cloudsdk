package services

import "context"

// VMConfig represents the configuration for creating a virtual machine
type VMConfig struct {
	Name           string
	ImageID        string
	InstanceType   string
	KeyName        string
	SecurityGroups []string
	UserData       string
}

// VM represents a virtual machine
type VM struct {
	ID         string
	Name       string
	State      string
	PublicIP   string
	PrivateIP  string
	LaunchTime string
}

// InstanceTypeFilter represents filters for instance type queries
type InstanceTypeFilter struct {
	VCpus         *int32
	MemoryGB      *float64
	StorageGB     *int32
	NetworkPerf   *string
	InstanceTypes []string
}

// InstanceType represents an EC2 instance type
type InstanceType struct {
	InstanceType       string
	VCpus              int32
	MemoryGB           float64
	StorageGB          int32
	NetworkPerformance string
	CurrentGeneration  bool
}

// InstanceTypesService defines operations for EC2 instance types
type InstanceTypesService interface {
	List(ctx context.Context, filter *InstanceTypeFilter) ([]*InstanceType, error)
}

// PlacementGroupConfig represents configuration for creating a placement group
type PlacementGroupConfig struct {
	GroupName string
	Strategy  string // cluster, partition, spread
}

// PlacementGroup represents an EC2 placement group
type PlacementGroup struct {
	GroupName    string
	GroupId      string
	Strategy     string
	State        string
	GroupArn     string
}

// PlacementGroupsService defines operations for EC2 placement groups
type PlacementGroupsService interface {
	Create(ctx context.Context, config *PlacementGroupConfig) (*PlacementGroup, error)
	Delete(ctx context.Context, groupName string) error
	List(ctx context.Context) ([]*PlacementGroup, error)
}

// SpotInstanceConfig represents configuration for requesting spot instances
type SpotInstanceConfig struct {
	InstanceType        string
	ImageID             string
	SpotPrice           *string
	AvailabilityZone    *string
	LaunchSpecification *SpotLaunchSpec
}

// SpotLaunchSpec represents launch specification for spot instances
type SpotLaunchSpec struct {
	ImageID        string
	InstanceType   string
	KeyName        string
	SecurityGroups []string
	UserData       string
}

// SpotInstanceRequest represents a spot instance request
type SpotInstanceRequest struct {
	SpotInstanceRequestId string
	InstanceId            string
	State                 string
	Status                string
	SpotPrice             string
	LaunchSpecification   *SpotLaunchSpec
	CreateTime            string
}

// SpotInstancesService defines operations for EC2 spot instances
type SpotInstancesService interface {
	Request(ctx context.Context, config *SpotInstanceConfig) (*SpotInstanceRequest, error)
	Describe(ctx context.Context, requestIds []string) ([]*SpotInstanceRequest, error)
	Cancel(ctx context.Context, requestId string) error
}

// Compute defines the interface for compute operations
type Compute interface {
	CreateVM(ctx context.Context, config *VMConfig) (*VM, error)
	ListVMs(ctx context.Context) ([]*VM, error)
	GetVM(ctx context.Context, id string) (*VM, error)
	StartVM(ctx context.Context, id string) error
	StopVM(ctx context.Context, id string) error
	DeleteVM(ctx context.Context, id string) error

	// New services
	InstanceTypes() InstanceTypesService
	PlacementGroups() PlacementGroupsService
	SpotInstances() SpotInstancesService
}
