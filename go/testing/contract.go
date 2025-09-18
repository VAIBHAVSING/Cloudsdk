package testing

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

// ProviderContractSuite tests that providers correctly implement the expected interfaces
type ProviderContractSuite struct {
	t        *testing.T
	provider cloudsdk.Provider
	client   *cloudsdk.Client
	timeout  time.Duration
}

// NewProviderContractSuite creates a new provider contract test suite
func NewProviderContractSuite(t *testing.T, provider cloudsdk.Provider) *ProviderContractSuite {
	return &ProviderContractSuite{
		t:        t,
		provider: provider,
		client:   cloudsdk.New(provider, nil),
		timeout:  2 * time.Minute,
	}
}

// RunAllTests runs all contract tests for the provider
func (s *ProviderContractSuite) RunAllTests() {
	s.t.Run("ProviderInterface", func(t *testing.T) { s.TestProviderInterface() })
	s.t.Run("ServiceAvailability", func(t *testing.T) { s.TestServiceAvailability() })

	// Test each supported service
	supportedServices := s.provider.SupportedServices()
	for _, serviceType := range supportedServices {
		switch serviceType {
		case cloudsdk.ServiceCompute:
			s.t.Run("ComputeService", func(t *testing.T) { s.TestComputeService() })
		case cloudsdk.ServiceStorage:
			s.t.Run("StorageService", func(t *testing.T) { s.TestStorageService() })
		case cloudsdk.ServiceDatabase:
			s.t.Run("DatabaseService", func(t *testing.T) { s.TestDatabaseService() })
		}
	}
}

// TestProviderInterface tests the basic provider interface
func (s *ProviderContractSuite) TestProviderInterface() {
	// Test Name() method
	name := s.provider.Name()
	if name == "" {
		s.t.Error("Provider Name() returned empty string")
	}

	// Test Region() method
	region := s.provider.Region()
	if region == "" {
		s.t.Error("Provider Region() returned empty string")
	}

	// Test SupportedServices() method
	services := s.provider.SupportedServices()
	if len(services) == 0 {
		s.t.Error("Provider SupportedServices() returned empty slice")
	}

	// Validate service types
	for _, service := range services {
		switch service {
		case cloudsdk.ServiceCompute, cloudsdk.ServiceStorage, cloudsdk.ServiceDatabase:
			// Valid service type
		default:
			s.t.Errorf("Provider returned invalid service type: %s", service)
		}
	}
}

// TestServiceAvailability tests that the client correctly handles service availability
func (s *ProviderContractSuite) TestServiceAvailability() {
	supportedServices := s.provider.SupportedServices()

	// Test supported services don't panic
	for _, serviceType := range supportedServices {
		switch serviceType {
		case cloudsdk.ServiceCompute:
			MustNotPanic(s.t, func() {
				s.client.Compute()
			})
		case cloudsdk.ServiceStorage:
			MustNotPanic(s.t, func() {
				s.client.Storage()
			})
		case cloudsdk.ServiceDatabase:
			MustNotPanic(s.t, func() {
				s.client.Database()
			})
		}
	}

	// Test unsupported services panic appropriately
	allServices := []cloudsdk.ServiceType{
		cloudsdk.ServiceCompute,
		cloudsdk.ServiceStorage,
		cloudsdk.ServiceDatabase,
	}

	for _, serviceType := range allServices {
		if !s.isServiceSupported(serviceType) {
			switch serviceType {
			case cloudsdk.ServiceCompute:
				MustPanic(s.t, func() {
					s.client.Compute()
				})
			case cloudsdk.ServiceStorage:
				MustPanic(s.t, func() {
					s.client.Storage()
				})
			case cloudsdk.ServiceDatabase:
				MustPanic(s.t, func() {
					s.client.Database()
				})
			}
		}
	}
}

// TestComputeService tests the compute service contract
func (s *ProviderContractSuite) TestComputeService() {
	if !s.isServiceSupported(cloudsdk.ServiceCompute) {
		s.t.Skip("Compute service not supported by provider")
	}

	compute := s.client.Compute()
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	// Test ListVMs (should not error, even if empty)
	vms, err := compute.ListVMs(ctx)
	if err != nil {
		s.t.Errorf("ListVMs failed: %v", err)
		return
	}

	// VMs can be empty, but should be a valid slice
	if vms == nil {
		s.t.Error("ListVMs returned nil slice")
	}

	// Test VM creation and lifecycle
	s.testVMLifecycle(compute, ctx)
}

// TestStorageService tests the storage service contract
func (s *ProviderContractSuite) TestStorageService() {
	if !s.isServiceSupported(cloudsdk.ServiceStorage) {
		s.t.Skip("Storage service not supported by provider")
	}

	storage := s.client.Storage()
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	// Test ListBuckets (should not error, even if empty)
	buckets, err := storage.ListBuckets(ctx)
	if err != nil {
		s.t.Errorf("ListBuckets failed: %v", err)
		return
	}

	// Buckets can be empty, but should be a valid slice
	if buckets == nil {
		s.t.Error("ListBuckets returned nil slice")
	}

	// Test bucket and object lifecycle
	s.testStorageLifecycle(storage, ctx)
}

// TestDatabaseService tests the database service contract
func (s *ProviderContractSuite) TestDatabaseService() {
	if !s.isServiceSupported(cloudsdk.ServiceDatabase) {
		s.t.Skip("Database service not supported by provider")
	}

	database := s.client.Database()
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	// Test ListDBs (should not error, even if empty)
	dbs, err := database.ListDBs(ctx)
	if err != nil {
		s.t.Errorf("ListDBs failed: %v", err)
		return
	}

	// DBs can be empty, but should be a valid slice
	if dbs == nil {
		s.t.Error("ListDBs returned nil slice")
	}

	// Test database lifecycle
	s.testDatabaseLifecycle(database, ctx)
}

// testVMLifecycle tests the complete VM lifecycle
func (s *ProviderContractSuite) testVMLifecycle(compute services.Compute, ctx context.Context) {
	// Create VM
	config := GenerateVMConfig("contract-test-vm")
	vm, err := compute.CreateVM(ctx, config)
	if err != nil {
		s.t.Errorf("CreateVM failed: %v", err)
		return
	}

	// Validate VM structure
	AssertVMValid(s.t, vm)

	// Get VM
	retrievedVM, err := compute.GetVM(ctx, vm.ID)
	if err != nil {
		s.t.Errorf("GetVM failed: %v", err)
	} else {
		AssertVMValid(s.t, retrievedVM)
		AssertEqual(s.t, vm.ID, retrievedVM.ID)
		AssertEqual(s.t, vm.Name, retrievedVM.Name)
	}

	// List VMs (should include our VM)
	vms, err := compute.ListVMs(ctx)
	if err != nil {
		s.t.Errorf("ListVMs failed: %v", err)
	} else {
		found := false
		for _, listedVM := range vms {
			if listedVM.ID == vm.ID {
				found = true
				break
			}
		}
		if !found {
			s.t.Error("Created VM not found in ListVMs result")
		}
	}

	// Test VM state operations (if VM supports them)
	s.testVMStateOperations(compute, ctx, vm.ID)

	// Delete VM
	if err := compute.DeleteVM(ctx, vm.ID); err != nil {
		s.t.Errorf("DeleteVM failed: %v", err)
	}

	// Verify VM is deleted
	_, err = compute.GetVM(ctx, vm.ID)
	if err == nil {
		s.t.Error("GetVM should fail after deletion")
	} else {
		// Should be a resource not found error
		if cloudErr, ok := err.(*cloudsdk.CloudError); ok {
			if cloudErr.Code != cloudsdk.ErrResourceNotFound {
				s.t.Errorf("Expected ErrResourceNotFound after deletion, got: %s", cloudErr.Code)
			}
		}
	}
}

// testVMStateOperations tests VM start/stop operations
func (s *ProviderContractSuite) testVMStateOperations(compute services.Compute, ctx context.Context, vmID string) {
	// Try to stop VM (may not be supported by all providers)
	if err := compute.StopVM(ctx, vmID); err != nil {
		s.t.Logf("StopVM not supported or failed: %v", err)
		return
	}

	// Try to start VM
	if err := compute.StartVM(ctx, vmID); err != nil {
		s.t.Logf("StartVM failed: %v", err)
	}
}

// testStorageLifecycle tests the complete storage lifecycle
func (s *ProviderContractSuite) testStorageLifecycle(storage services.Storage, ctx context.Context) {
	bucketName := GenerateBucketName("contract-test")

	// Create bucket
	config := GenerateBucketConfig(bucketName)
	if err := storage.CreateBucket(ctx, config); err != nil {
		s.t.Errorf("CreateBucket failed: %v", err)
		return
	}

	// List buckets (should include our bucket)
	buckets, err := storage.ListBuckets(ctx)
	if err != nil {
		s.t.Errorf("ListBuckets failed: %v", err)
	} else {
		found := false
		for _, bucket := range buckets {
			if bucket == bucketName {
				found = true
				break
			}
		}
		if !found {
			s.t.Error("Created bucket not found in ListBuckets result")
		}
	}

	// Test object operations
	s.testObjectOperations(storage, ctx, bucketName)

	// Delete bucket
	if err := storage.DeleteBucket(ctx, bucketName); err != nil {
		s.t.Errorf("DeleteBucket failed: %v", err)
	}
}

// testObjectOperations tests object operations within a bucket
func (s *ProviderContractSuite) testObjectOperations(storage services.Storage, ctx context.Context, bucketName string) {
	objectKey := "test-object.txt"
	testData := "Hello, World!"

	// Put object
	if err := storage.PutObject(ctx, bucketName, objectKey, strings.NewReader(testData)); err != nil {
		s.t.Errorf("PutObject failed: %v", err)
		return
	}

	// Get object
	reader, err := storage.GetObject(ctx, bucketName, objectKey)
	if err != nil {
		s.t.Errorf("GetObject failed: %v", err)
	} else {
		defer reader.Close()
		data, err := io.ReadAll(reader)
		if err != nil {
			s.t.Errorf("Failed to read object data: %v", err)
		} else if string(data) != testData {
			s.t.Errorf("Object data mismatch: expected %q, got %q", testData, string(data))
		}
	}

	// List objects
	objects, err := storage.ListObjects(ctx, bucketName)
	if err != nil {
		s.t.Errorf("ListObjects failed: %v", err)
	} else {
		found := false
		for _, obj := range objects {
			if obj.Key == objectKey {
				found = true
				if obj.Size != int64(len(testData)) {
					s.t.Errorf("Object size mismatch: expected %d, got %d", len(testData), obj.Size)
				}
				break
			}
		}
		if !found {
			s.t.Error("Created object not found in ListObjects result")
		}
	}

	// Delete object
	if err := storage.DeleteObject(ctx, bucketName, objectKey); err != nil {
		s.t.Errorf("DeleteObject failed: %v", err)
	}
}

// testDatabaseLifecycle tests the complete database lifecycle
func (s *ProviderContractSuite) testDatabaseLifecycle(database services.Database, ctx context.Context) {
	dbName := "contract-test-db"

	// Create database
	config := GenerateDBConfig(dbName)
	db, err := database.CreateDB(ctx, config)
	if err != nil {
		s.t.Errorf("CreateDB failed: %v", err)
		return
	}

	// Validate database structure
	AssertDBValid(s.t, db)

	// Get database
	retrievedDB, err := database.GetDB(ctx, db.ID)
	if err != nil {
		s.t.Errorf("GetDB failed: %v", err)
	} else {
		AssertDBValid(s.t, retrievedDB)
		AssertEqual(s.t, db.ID, retrievedDB.ID)
		AssertEqual(s.t, db.Name, retrievedDB.Name)
	}

	// List databases (should include our database)
	dbs, err := database.ListDBs(ctx)
	if err != nil {
		s.t.Errorf("ListDBs failed: %v", err)
	} else {
		found := false
		for _, listedDB := range dbs {
			if listedDB.ID == db.ID {
				found = true
				break
			}
		}
		if !found {
			s.t.Error("Created database not found in ListDBs result")
		}
	}

	// Delete database
	if err := database.DeleteDB(ctx, db.ID); err != nil {
		s.t.Errorf("DeleteDB failed: %v", err)
	}

	// Verify database is deleted
	_, err = database.GetDB(ctx, db.ID)
	if err == nil {
		s.t.Error("GetDB should fail after deletion")
	} else {
		// Should be a resource not found error
		if cloudErr, ok := err.(*cloudsdk.CloudError); ok {
			if cloudErr.Code != cloudsdk.ErrResourceNotFound {
				s.t.Errorf("Expected ErrResourceNotFound after deletion, got: %s", cloudErr.Code)
			}
		}
	}
}

// isServiceSupported checks if a service is supported by the provider
func (s *ProviderContractSuite) isServiceSupported(serviceType cloudsdk.ServiceType) bool {
	supportedServices := s.provider.SupportedServices()
	for _, service := range supportedServices {
		if service == serviceType {
			return true
		}
	}
	return false
}

// RunProviderContractTests is a convenience function to run all contract tests
func RunProviderContractTests(t *testing.T, provider cloudsdk.Provider) {
	suite := NewProviderContractSuite(t, provider)
	suite.RunAllTests()
}
