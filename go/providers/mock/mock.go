// Package mock provides a comprehensive mock provider implementation for testing.
//
// The mock provider allows developers to test their applications without making
// actual cloud API calls, providing configurable responses, error injection,
// and comprehensive testing utilities.
//
// FEATURES:
//   - Implements all cloudsdk.Provider interface methods
//   - Configurable responses for different operations
//   - Error injection for testing failure scenarios
//   - Builder pattern for easy test setup
//   - Realistic data generation for testing
//   - Request/response recording for verification
//   - Support for all service types (Compute, Storage, Database)
//
// QUICK START:
//
// Basic mock provider:
//
//	provider := mock.New("us-east-1")
//	client := cloudsdk.New(provider)
//
//	// Use normally - all operations return success by default
//	vm, err := client.Compute().CreateVM(ctx, &services.VMConfig{
//	    Name: "test-vm",
//	    ImageID: "ami-12345",
//	    InstanceType: "t2.micro",
//	})
//
// Configure specific responses:
//
//	provider := mock.New("us-east-1").
//	    WithVMResponse("test-vm", &services.VM{
//	        ID: "i-1234567890abcdef0",
//	        Name: "test-vm",
//	        State: "running",
//	    }).
//	    WithError("CreateBucket", errors.New("bucket already exists"))
//
// Error injection for testing:
//
//	provider := mock.New("us-east-1").
//	    WithError("CreateVM", cloudsdk.NewCloudError(
//	        cloudsdk.ErrResourceConflict,
//	        "VM name already exists",
//	        "mock", "compute", "CreateVM"))
//
// TESTING PATTERNS:
//
// The mock provider supports common testing patterns:
//   - Success scenarios with realistic data
//   - Error scenarios with specific error types
//   - Resource state management (create, read, update, delete)
//   - Request verification and assertion
//   - Concurrent operation testing
//
// BUILDER PATTERN:
//
// The mock provider uses a fluent builder pattern for configuration:
//
//	provider := mock.New("us-east-1").
//	    WithSupportedServices(cloudsdk.ServiceCompute, cloudsdk.ServiceStorage).
//	    WithVMResponse("web-server", mockVM).
//	    WithBucketResponse("my-bucket", mockBucket).
//	    WithError("DeleteVM", mockError).
//	    WithDelay("CreateDB", 2*time.Second)
//
// VERIFICATION:
//
// The mock provider records all operations for verification:
//
//	provider := mock.New("us-east-1")
//	client := cloudsdk.New(provider)
//
//	// Perform operations
//	vm, err := client.Compute().CreateVM(ctx, vmConfig)
//
//	// Verify operations were called
//	assert.True(t, provider.WasCalled("CreateVM"))
//	assert.Equal(t, 1, provider.CallCount("CreateVM"))
//	assert.Equal(t, vmConfig, provider.LastCallArgs("CreateVM")[0])
package mock

import (
	"fmt"
	"sync"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

// MockProvider implements the cloudsdk.Provider interface for testing.
// It provides configurable responses, error injection, and operation recording
// to enable comprehensive testing without real cloud dependencies.
type MockProvider struct {
	// Configuration
	region            string
	supportedServices []cloudsdk.ServiceType

	// Response configuration
	vmResponses     map[string]*services.VM
	bucketResponses map[string]bool // true if bucket exists
	dbResponses     map[string]*services.DBInstance
	objectResponses map[string]map[string][]byte // bucket -> key -> data

	// Error injection
	errors map[string]error
	delays map[string]time.Duration

	// Operation recording
	mu           sync.RWMutex
	operations   []Operation
	callCounts   map[string]int
	lastCallArgs map[string][]interface{}

	// State management
	vmState     map[string]*services.VM
	bucketState map[string]*BucketState
	dbState     map[string]*services.DBInstance
}

// Operation represents a recorded operation for verification
type Operation struct {
	Method    string
	Args      []interface{}
	Result    interface{}
	Error     error
	Timestamp time.Time
}

// BucketState represents the state of a mock bucket
type BucketState struct {
	Name    string
	Region  string
	Objects map[string][]byte
	Tags    map[string]string
}

// New creates a new mock provider with default configuration.
// By default, the mock provider supports all services and returns
// success responses for all operations.
//
// Parameters:
//   - region: The mock region (can be any string for testing)
//
// Example:
//
//	provider := mock.New("us-east-1")
//	client := cloudsdk.New(provider)
func New(region string) *MockProvider {
	return &MockProvider{
		region: region,
		supportedServices: []cloudsdk.ServiceType{
			cloudsdk.ServiceCompute,
			cloudsdk.ServiceStorage,
			cloudsdk.ServiceDatabase,
		},
		vmResponses:     make(map[string]*services.VM),
		bucketResponses: make(map[string]bool),
		dbResponses:     make(map[string]*services.DBInstance),
		objectResponses: make(map[string]map[string][]byte),
		errors:          make(map[string]error),
		delays:          make(map[string]time.Duration),
		operations:      make([]Operation, 0),
		callCounts:      make(map[string]int),
		lastCallArgs:    make(map[string][]interface{}),
		vmState:         make(map[string]*services.VM),
		bucketState:     make(map[string]*BucketState),
		dbState:         make(map[string]*services.DBInstance),
	}
}

// WithSupportedServices configures which services the mock provider supports.
// This allows testing scenarios where providers have limited service support.
//
// Example:
//
//	// Mock a provider that only supports storage
//	provider := mock.New("us-east-1").
//	    WithSupportedServices(cloudsdk.ServiceStorage)
func (m *MockProvider) WithSupportedServices(services ...cloudsdk.ServiceType) *MockProvider {
	m.supportedServices = services
	return m
}

// WithVMResponse configures a specific response for VM operations.
// When a VM with the specified name is created or retrieved,
// the mock provider will return the configured VM object.
//
// Example:
//
//	mockVM := &services.VM{
//	    ID: "i-1234567890abcdef0",
//	    Name: "test-vm",
//	    State: "running",
//	    PublicIP: "203.0.113.1",
//	    PrivateIP: "10.0.1.100",
//	}
//	provider := mock.New("us-east-1").WithVMResponse("test-vm", mockVM)
func (m *MockProvider) WithVMResponse(name string, vm *services.VM) *MockProvider {
	m.vmResponses[name] = vm
	return m
}

// WithBucketResponse configures whether a bucket exists in the mock provider.
// This affects bucket creation, deletion, and listing operations.
//
// Example:
//
//	// Configure that "existing-bucket" already exists
//	provider := mock.New("us-east-1").WithBucketResponse("existing-bucket", true)
func (m *MockProvider) WithBucketResponse(name string, exists bool) *MockProvider {
	m.bucketResponses[name] = exists
	return m
}

// WithDBResponse configures a specific response for database operations.
// When a database with the specified name is created or retrieved,
// the mock provider will return the configured database instance.
//
// Example:
//
//	mockDB := &services.DBInstance{
//	    ID: "myapp-prod-db",
//	    Name: "myapp-prod-db",
//	    Engine: "postgres",
//	    Status: "available",
//	    Endpoint: "myapp-prod-db.cluster-xyz.us-east-1.rds.amazonaws.com",
//	}
//	provider := mock.New("us-east-1").WithDBResponse("myapp-prod-db", mockDB)
func (m *MockProvider) WithDBResponse(name string, db *services.DBInstance) *MockProvider {
	m.dbResponses[name] = db
	return m
}

// WithError configures the mock provider to return a specific error
// for the specified operation. This enables testing error scenarios.
//
// Example:
//
//	// Test authentication error
//	provider := mock.New("us-east-1").
//	    WithError("CreateVM", cloudsdk.NewAuthenticationError("mock", nil))
//
//	// Test resource conflict
//	provider := mock.New("us-east-1").
//	    WithError("CreateBucket", cloudsdk.NewCloudError(
//	        cloudsdk.ErrResourceConflict,
//	        "Bucket already exists",
//	        "mock", "storage", "CreateBucket"))
func (m *MockProvider) WithError(operation string, err error) *MockProvider {
	m.errors[operation] = err
	return m
}

// WithDelay configures the mock provider to introduce a delay
// for the specified operation. This enables testing timeout scenarios
// and concurrent operations.
//
// Example:
//
//	// Simulate slow database creation
//	provider := mock.New("us-east-1").
//	    WithDelay("CreateDB", 5*time.Second)
func (m *MockProvider) WithDelay(operation string, delay time.Duration) *MockProvider {
	m.delays[operation] = delay
	return m
}

// recordOperation records an operation for later verification
func (m *MockProvider) recordOperation(method string, args []interface{}, result interface{}, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	operation := Operation{
		Method:    method,
		Args:      args,
		Result:    result,
		Error:     err,
		Timestamp: time.Now(),
	}

	m.operations = append(m.operations, operation)
	m.callCounts[method]++
	m.lastCallArgs[method] = args
}

// checkError returns any configured error for the operation
func (m *MockProvider) checkError(operation string) error {
	if err, exists := m.errors[operation]; exists {
		return err
	}
	return nil
}

// applyDelay applies any configured delay for the operation
func (m *MockProvider) applyDelay(operation string) {
	if delay, exists := m.delays[operation]; exists {
		time.Sleep(delay)
	}
}

// WasCalled returns true if the specified operation was called
func (m *MockProvider) WasCalled(operation string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCounts[operation] > 0
}

// CallCount returns the number of times the specified operation was called
func (m *MockProvider) CallCount(operation string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCounts[operation]
}

// LastCallArgs returns the arguments from the last call to the specified operation
func (m *MockProvider) LastCallArgs(operation string) []interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastCallArgs[operation]
}

// AllOperations returns all recorded operations for verification
func (m *MockProvider) AllOperations() []Operation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent race conditions
	operations := make([]Operation, len(m.operations))
	copy(operations, m.operations)
	return operations
}

// Reset clears all recorded operations and state
func (m *MockProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.operations = make([]Operation, 0)
	m.callCounts = make(map[string]int)
	m.lastCallArgs = make(map[string][]interface{})
	m.vmState = make(map[string]*services.VM)
	m.bucketState = make(map[string]*BucketState)
	m.dbState = make(map[string]*services.DBInstance)
}

// Provider interface implementation

// Name returns the provider name identifier
func (m *MockProvider) Name() string {
	return "mock"
}

// Region returns the configured region
func (m *MockProvider) Region() string {
	return m.region
}

// SupportedServices returns the list of services supported by this mock provider
func (m *MockProvider) SupportedServices() []cloudsdk.ServiceType {
	return m.supportedServices
}

// Compute returns the mock compute service
func (m *MockProvider) Compute() services.Compute {
	return &MockCompute{provider: m}
}

// Storage returns the mock storage service
func (m *MockProvider) Storage() services.Storage {
	return &MockStorage{provider: m}
}

// Database returns the mock database service
func (m *MockProvider) Database() services.Database {
	return &MockDatabase{provider: m}
}

// generateVMID generates a realistic VM ID for testing
func generateVMID() string {
	return fmt.Sprintf("i-%016x", time.Now().UnixNano())
}

// generateDBInstanceID generates a realistic database instance ID for testing
func generateDBInstanceID(name string) string {
	return name
}

// generateEndpoint generates a realistic database endpoint for testing
func generateEndpoint(name, region string) string {
	return fmt.Sprintf("%s.cluster-xyz.%s.rds.amazonaws.com", name, region)
}
