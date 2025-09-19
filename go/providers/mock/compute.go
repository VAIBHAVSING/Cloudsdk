package mock

import (
	"context"
	"fmt"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

// MockCompute implements the services.Compute interface for testing.
// It provides configurable responses and error injection for all compute operations.
type MockCompute struct {
	provider *MockProvider
}

// CreateVM creates a mock virtual machine with configurable responses.
// Returns a configured VM response if available, otherwise generates a realistic mock VM.
//
// Error injection:
//   - Configure errors using WithError("CreateVM", error)
//   - Common test scenarios: authentication, authorization, resource conflicts
//
// Example:
//
//	// Success scenario
//	vm, err := mockCompute.CreateVM(ctx, &services.VMConfig{
//	    Name: "test-vm",
//	    ImageID: "ami-12345",
//	    InstanceType: "t2.micro",
//	})
//
//	// Error scenario (configured with WithError)
//	provider := mock.New("us-east-1").
//	    WithError("CreateVM", cloudsdk.NewResourceConflictError(...))
func (m *MockCompute) CreateVM(ctx context.Context, config *services.VMConfig) (*services.VM, error) {
	m.provider.applyDelay("CreateVM")

	if err := m.provider.checkError("CreateVM"); err != nil {
		m.provider.recordOperation("CreateVM", []interface{}{config}, nil, err)
		return nil, err
	}

	// Check if we have a configured response for this VM name
	if vm, exists := m.provider.vmResponses[config.Name]; exists {
		// Store in state for later retrieval
		m.provider.vmState[vm.ID] = vm
		m.provider.recordOperation("CreateVM", []interface{}{config}, vm, nil)
		return vm, nil
	}

	// Generate a realistic mock VM
	vm := &services.VM{
		ID:         generateVMID(),
		Name:       config.Name,
		State:      "running",
		PublicIP:   "203.0.113." + fmt.Sprintf("%d", time.Now().Unix()%254+1),
		PrivateIP:  "10.0.1." + fmt.Sprintf("%d", time.Now().Unix()%254+1),
		LaunchTime: time.Now().Format(time.RFC3339),
	}

	// Store in state
	m.provider.vmState[vm.ID] = vm

	m.provider.recordOperation("CreateVM", []interface{}{config}, vm, nil)
	return vm, nil
}

// GetVM retrieves a mock virtual machine by ID.
// Returns a VM from the mock state if it exists, otherwise returns a not found error.
//
// Error injection:
//   - Configure errors using WithError("GetVM", error)
//   - Automatically returns ErrResourceNotFound for non-existent VMs
//
// Example:
//
//	vm, err := mockCompute.GetVM(ctx, "i-1234567890abcdef0")
//	if err != nil {
//	    // Handle not found or configured error
//	}
func (m *MockCompute) GetVM(ctx context.Context, id string) (*services.VM, error) {
	m.provider.applyDelay("GetVM")

	if err := m.provider.checkError("GetVM"); err != nil {
		m.provider.recordOperation("GetVM", []interface{}{id}, nil, err)
		return nil, err
	}

	// Check if VM exists in state
	if vm, exists := m.provider.vmState[id]; exists {
		m.provider.recordOperation("GetVM", []interface{}{id}, vm, nil)
		return vm, nil
	}

	// VM not found
	err := cloudsdk.NewResourceNotFoundError("mock", "compute", "VM", id)
	m.provider.recordOperation("GetVM", []interface{}{id}, nil, err)
	return nil, err
}

// ListVMs returns all mock virtual machines in the current state.
// Returns an empty slice if no VMs exist.
//
// Error injection:
//   - Configure errors using WithError("ListVMs", error)
//
// Example:
//
//	vms, err := mockCompute.ListVMs(ctx)
//	for _, vm := range vms {
//	    fmt.Printf("VM: %s (%s)\n", vm.Name, vm.State)
//	}
func (m *MockCompute) ListVMs(ctx context.Context) ([]*services.VM, error) {
	m.provider.applyDelay("ListVMs")

	if err := m.provider.checkError("ListVMs"); err != nil {
		m.provider.recordOperation("ListVMs", []interface{}{}, nil, err)
		return nil, err
	}

	// Collect all VMs from state
	vms := make([]*services.VM, 0, len(m.provider.vmState))
	for _, vm := range m.provider.vmState {
		vms = append(vms, vm)
	}

	m.provider.recordOperation("ListVMs", []interface{}{}, vms, nil)
	return vms, nil
}

// DeleteVM removes a mock virtual machine from the state.
// Returns an error if the VM doesn't exist.
//
// Error injection:
//   - Configure errors using WithError("DeleteVM", error)
//   - Automatically returns ErrResourceNotFound for non-existent VMs
//
// Example:
//
//	err := mockCompute.DeleteVM(ctx, "i-1234567890abcdef0")
//	if err != nil {
//	    // Handle not found or configured error
//	}
func (m *MockCompute) DeleteVM(ctx context.Context, id string) error {
	m.provider.applyDelay("DeleteVM")

	if err := m.provider.checkError("DeleteVM"); err != nil {
		m.provider.recordOperation("DeleteVM", []interface{}{id}, nil, err)
		return err
	}

	// Check if VM exists
	if _, exists := m.provider.vmState[id]; !exists {
		err := cloudsdk.NewResourceNotFoundError("mock", "compute", "VM", id)
		m.provider.recordOperation("DeleteVM", []interface{}{id}, nil, err)
		return err
	}

	// Remove from state
	delete(m.provider.vmState, id)

	m.provider.recordOperation("DeleteVM", []interface{}{id}, nil, nil)
	return nil
}

// StartVM starts a mock virtual machine by updating its state.
// Returns an error if the VM doesn't exist or is already running.
//
// Error injection:
//   - Configure errors using WithError("StartVM", error)
//   - Automatically returns ErrResourceNotFound for non-existent VMs
//   - Returns ErrInvalidConfig if VM is already running
//
// Example:
//
//	err := mockCompute.StartVM(ctx, "i-1234567890abcdef0")
//	if err != nil {
//	    // Handle not found, already running, or configured error
//	}
func (m *MockCompute) StartVM(ctx context.Context, id string) error {
	m.provider.applyDelay("StartVM")

	if err := m.provider.checkError("StartVM"); err != nil {
		m.provider.recordOperation("StartVM", []interface{}{id}, nil, err)
		return err
	}

	// Check if VM exists
	vm, exists := m.provider.vmState[id]
	if !exists {
		err := cloudsdk.NewResourceNotFoundError("mock", "compute", "VM", id)
		m.provider.recordOperation("StartVM", []interface{}{id}, nil, err)
		return err
	}

	// Check if already running
	if vm.State == "running" {
		err := cloudsdk.NewInvalidConfigError("mock", "compute", "state", "VM is already running")
		m.provider.recordOperation("StartVM", []interface{}{id}, nil, err)
		return err
	}

	// Update state
	vm.State = "running"

	m.provider.recordOperation("StartVM", []interface{}{id}, nil, nil)
	return nil
}

// StopVM stops a mock virtual machine by updating its state.
// Returns an error if the VM doesn't exist or is already stopped.
//
// Error injection:
//   - Configure errors using WithError("StopVM", error)
//   - Automatically returns ErrResourceNotFound for non-existent VMs
//   - Returns ErrInvalidConfig if VM is already stopped
//
// Example:
//
//	err := mockCompute.StopVM(ctx, "i-1234567890abcdef0")
//	if err != nil {
//	    // Handle not found, already stopped, or configured error
//	}
func (m *MockCompute) StopVM(ctx context.Context, id string) error {
	m.provider.applyDelay("StopVM")

	if err := m.provider.checkError("StopVM"); err != nil {
		m.provider.recordOperation("StopVM", []interface{}{id}, nil, err)
		return err
	}

	// Check if VM exists
	vm, exists := m.provider.vmState[id]
	if !exists {
		err := cloudsdk.NewResourceNotFoundError("mock", "compute", "VM", id)
		m.provider.recordOperation("StopVM", []interface{}{id}, nil, err)
		return err
	}

	// Check if already stopped
	if vm.State == "stopped" {
		err := cloudsdk.NewInvalidConfigError("mock", "compute", "state", "VM is already stopped")
		m.provider.recordOperation("StopVM", []interface{}{id}, nil, err)
		return err
	}

	// Update state
	vm.State = "stopped"

	m.provider.recordOperation("StopVM", []interface{}{id}, nil, nil)
	return nil
}

// InstanceTypes returns the mock instance types service
func (m *MockCompute) InstanceTypes() services.InstanceTypesService {
	return &MockInstanceTypesService{provider: m.provider}
}

// PlacementGroups returns the mock placement groups service
func (m *MockCompute) PlacementGroups() services.PlacementGroupsService {
	return &MockPlacementGroupsService{provider: m.provider}
}

// SpotInstances returns the mock spot instances service
func (m *MockCompute) SpotInstances() services.SpotInstancesService {
	return &MockSpotInstancesService{provider: m.provider}
}

// MockInstanceTypesService implements the services.InstanceTypesService interface for testing
type MockInstanceTypesService struct {
	provider *MockProvider
}

// List returns mock instance types
func (s *MockInstanceTypesService) List(ctx context.Context, filter *services.InstanceTypeFilter) ([]*services.InstanceType, error) {
	s.provider.applyDelay("ListInstanceTypes")
	if err := s.provider.checkError("ListInstanceTypes"); err != nil {
		s.provider.recordOperation("ListInstanceTypes", []interface{}{filter}, nil, err)
		return nil, err
	}

	// Generate mock instance types
	instanceTypes := []*services.InstanceType{
		{
			InstanceType:       "t2.micro",
			VCpus:              1,
			MemoryGB:           1.0,
			StorageGB:          0,
			NetworkPerformance: "Low to Moderate",
			CurrentGeneration:  true,
		},
		{
			InstanceType:       "t2.small",
			VCpus:              1,
			MemoryGB:           2.0,
			StorageGB:          0,
			NetworkPerformance: "Low to Moderate",
			CurrentGeneration:  true,
		},
		{
			InstanceType:       "m5.large",
			VCpus:              2,
			MemoryGB:           8.0,
			StorageGB:          0,
			NetworkPerformance: "Up to 10 Gigabit",
			CurrentGeneration:  true,
		},
	}

	// Apply filters if provided
	if filter != nil {
		filtered := make([]*services.InstanceType, 0)
		for _, it := range instanceTypes {
			if filter.VCpus != nil && it.VCpus != *filter.VCpus {
				continue
			}
			if filter.MemoryGB != nil && it.MemoryGB < *filter.MemoryGB {
				continue
			}
			if filter.StorageGB != nil && it.StorageGB < *filter.StorageGB {
				continue
			}
			if filter.NetworkPerf != nil && it.NetworkPerformance != *filter.NetworkPerf {
				continue
			}
			filtered = append(filtered, it)
		}
		instanceTypes = filtered
	}

	s.provider.recordOperation("ListInstanceTypes", []interface{}{filter}, instanceTypes, nil)
	return instanceTypes, nil
}

// MockPlacementGroupsService implements the services.PlacementGroupsService interface for testing
type MockPlacementGroupsService struct {
	provider *MockProvider
}

// Create creates a mock placement group
func (s *MockPlacementGroupsService) Create(ctx context.Context, config *services.PlacementGroupConfig) (*services.PlacementGroup, error) {
	s.provider.applyDelay("CreatePlacementGroup")
	if err := s.provider.checkError("CreatePlacementGroup"); err != nil {
		s.provider.recordOperation("CreatePlacementGroup", []interface{}{config}, nil, err)
		return nil, err
	}

	group := &services.PlacementGroup{
		GroupName: config.GroupName,
		GroupId:   fmt.Sprintf("pg-%016x", time.Now().UnixNano()),
		Strategy:  config.Strategy,
		State:     "available",
		GroupArn:  fmt.Sprintf("arn:aws:ec2:%s:123456789012:placement-group/%s", s.provider.region, config.GroupName),
	}

	s.provider.recordOperation("CreatePlacementGroup", []interface{}{config}, group, nil)
	return group, nil
}

// Delete deletes a mock placement group
func (s *MockPlacementGroupsService) Delete(ctx context.Context, groupName string) error {
	s.provider.applyDelay("DeletePlacementGroup")
	if err := s.provider.checkError("DeletePlacementGroup"); err != nil {
		s.provider.recordOperation("DeletePlacementGroup", []interface{}{groupName}, nil, err)
		return err
	}

	s.provider.recordOperation("DeletePlacementGroup", []interface{}{groupName}, nil, nil)
	return nil
}

// List lists mock placement groups
func (s *MockPlacementGroupsService) List(ctx context.Context) ([]*services.PlacementGroup, error) {
	s.provider.applyDelay("ListPlacementGroups")
	if err := s.provider.checkError("ListPlacementGroups"); err != nil {
		s.provider.recordOperation("ListPlacementGroups", []interface{}{}, nil, err)
		return nil, err
	}

	groups := []*services.PlacementGroup{
		{
			GroupName: "default-cluster",
			GroupId:   "pg-1234567890abcdef0",
			Strategy:  "cluster",
			State:     "available",
			GroupArn:  fmt.Sprintf("arn:aws:ec2:%s:123456789012:placement-group/default-cluster", s.provider.region),
		},
	}

	s.provider.recordOperation("ListPlacementGroups", []interface{}{}, groups, nil)
	return groups, nil
}

// MockSpotInstancesService implements the services.SpotInstancesService interface for testing
type MockSpotInstancesService struct {
	provider *MockProvider
}

// Request requests a mock spot instance
func (s *MockSpotInstancesService) Request(ctx context.Context, config *services.SpotInstanceConfig) (*services.SpotInstanceRequest, error) {
	s.provider.applyDelay("RequestSpotInstances")
	if err := s.provider.checkError("RequestSpotInstances"); err != nil {
		s.provider.recordOperation("RequestSpotInstances", []interface{}{config}, nil, err)
		return nil, err
	}

	request := &services.SpotInstanceRequest{
		SpotInstanceRequestId: fmt.Sprintf("sir-%016x", time.Now().UnixNano()),
		State:                 "active",
		Status:                "fulfilled",
		SpotPrice:             "0.05",
		CreateTime:            time.Now().String(),
		InstanceId:            generateVMID(),
	}

	s.provider.recordOperation("RequestSpotInstances", []interface{}{config}, request, nil)
	return request, nil
}

// Describe describes mock spot instance requests
func (s *MockSpotInstancesService) Describe(ctx context.Context, requestIds []string) ([]*services.SpotInstanceRequest, error) {
	s.provider.applyDelay("DescribeSpotInstanceRequests")
	if err := s.provider.checkError("DescribeSpotInstanceRequests"); err != nil {
		s.provider.recordOperation("DescribeSpotInstanceRequests", []interface{}{requestIds}, nil, err)
		return nil, err
	}

	requests := []*services.SpotInstanceRequest{
		{
			SpotInstanceRequestId: "sir-1234567890abcdef0",
			State:                 "active",
			Status:                "fulfilled",
			SpotPrice:             "0.05",
			CreateTime:            time.Now().Add(-1 * time.Hour).String(),
			InstanceId:            "i-1234567890abcdef0",
		},
	}

	s.provider.recordOperation("DescribeSpotInstanceRequests", []interface{}{requestIds}, requests, nil)
	return requests, nil
}

// Cancel cancels a mock spot instance request
func (s *MockSpotInstancesService) Cancel(ctx context.Context, requestId string) error {
	s.provider.applyDelay("CancelSpotInstanceRequests")
	if err := s.provider.checkError("CancelSpotInstanceRequests"); err != nil {
		s.provider.recordOperation("CancelSpotInstanceRequests", []interface{}{requestId}, nil, err)
		return err
	}

	s.provider.recordOperation("CancelSpotInstanceRequests", []interface{}{requestId}, nil, nil)
	return nil
}
