package services

import "context"

// VMConfig represents the configuration for creating a virtual machine
type VMConfig struct {
	Name       string
	ImageID    string
	InstanceType string
	KeyName    string
	SecurityGroups []string
	UserData   string
}

// VM represents a virtual machine
type VM struct {
	ID          string
	Name        string
	State       string
	PublicIP    string
	PrivateIP   string
	LaunchTime  string
}

// Compute defines the interface for compute operations
type Compute interface {
	CreateVM(ctx context.Context, config *VMConfig) (*VM, error)
	ListVMs(ctx context.Context) ([]*VM, error)
	GetVM(ctx context.Context, id string) (*VM, error)
	StartVM(ctx context.Context, id string) error
	StopVM(ctx context.Context, id string) error
	DeleteVM(ctx context.Context, id string) error
}
