package testing

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

// IntegrationSuite provides utilities for integration testing with automatic cleanup
type IntegrationSuite struct {
	t        *testing.T
	provider cloudsdk.Provider
	client   *cloudsdk.Client
	cleanup  bool

	// Resource tracking for cleanup
	mu             sync.RWMutex
	createdVMs     []string
	createdBuckets []string
	createdDBs     []string

	// Configuration
	timeout time.Duration
	region  string
}

// NewIntegrationSuite creates a new integration test suite
func NewIntegrationSuite(t *testing.T) *IntegrationSuite {
	return &IntegrationSuite{
		t:              t,
		cleanup:        true,
		timeout:        5 * time.Minute,
		region:         "us-east-1",
		createdVMs:     make([]string, 0),
		createdBuckets: make([]string, 0),
		createdDBs:     make([]string, 0),
	}
}

// WithProvider sets the cloud provider for the integration suite
func (s *IntegrationSuite) WithProvider(provider cloudsdk.Provider) *IntegrationSuite {
	s.provider = provider
	s.client = cloudsdk.New(provider, nil)
	s.region = provider.Region()
	return s
}

// WithCleanup enables or disables automatic resource cleanup
func (s *IntegrationSuite) WithCleanup(cleanup bool) *IntegrationSuite {
	s.cleanup = cleanup
	return s
}

// WithTimeout sets the timeout for operations
func (s *IntegrationSuite) WithTimeout(timeout time.Duration) *IntegrationSuite {
	s.timeout = timeout
	return s
}

// WithRegion sets the region for the integration suite
func (s *IntegrationSuite) WithRegion(region string) *IntegrationSuite {
	s.region = region
	return s
}

// CreateTestVM creates a VM for testing and tracks it for cleanup
func (s *IntegrationSuite) CreateTestVM(name string) *services.VM {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	config := GenerateVMConfig(name)
	vm, err := s.client.Compute().CreateVM(ctx, config)
	if err != nil {
		s.t.Fatalf("Failed to create test VM: %v", err)
	}

	// Track for cleanup
	s.mu.Lock()
	s.createdVMs = append(s.createdVMs, vm.ID)
	s.mu.Unlock()

	return vm
}

// CreateTestBucket creates a bucket for testing and tracks it for cleanup
func (s *IntegrationSuite) CreateTestBucket(name string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	config := GenerateBucketConfig(name)
	err := s.client.Storage().CreateBucket(ctx, config)
	if err != nil {
		s.t.Fatalf("Failed to create test bucket: %v", err)
	}

	// Track for cleanup
	s.mu.Lock()
	s.createdBuckets = append(s.createdBuckets, name)
	s.mu.Unlock()
}

// CreateTestDB creates a database for testing and tracks it for cleanup
func (s *IntegrationSuite) CreateTestDB(name string) *services.DBInstance {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	config := GenerateDBConfig(name)
	db, err := s.client.Database().CreateDB(ctx, config)
	if err != nil {
		s.t.Fatalf("Failed to create test database: %v", err)
	}

	// Track for cleanup
	s.mu.Lock()
	s.createdDBs = append(s.createdDBs, db.ID)
	s.mu.Unlock()

	return db
}

// AssertVMRunning asserts that a VM is in running state
func (s *IntegrationSuite) AssertVMRunning(vmID string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	vm, err := s.client.Compute().GetVM(ctx, vmID)
	if err != nil {
		s.t.Fatalf("Failed to get VM %s: %v", vmID, err)
	}

	if vm.State != "running" {
		s.t.Fatalf("Expected VM %s to be running, got state: %s", vmID, vm.State)
	}
}

// AssertVMStopped asserts that a VM is in stopped state
func (s *IntegrationSuite) AssertVMStopped(vmID string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	vm, err := s.client.Compute().GetVM(ctx, vmID)
	if err != nil {
		s.t.Fatalf("Failed to get VM %s: %v", vmID, err)
	}

	if vm.State != "stopped" {
		s.t.Fatalf("Expected VM %s to be stopped, got state: %s", vmID, vm.State)
	}
}

// AssertBucketExists asserts that a bucket exists
func (s *IntegrationSuite) AssertBucketExists(bucketName string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	buckets, err := s.client.Storage().ListBuckets(ctx)
	if err != nil {
		s.t.Fatalf("Failed to list buckets: %v", err)
	}

	for _, bucket := range buckets {
		if bucket == bucketName {
			return // Found
		}
	}

	s.t.Fatalf("Bucket %s not found", bucketName)
}

// AssertDBAvailable asserts that a database is in available state
func (s *IntegrationSuite) AssertDBAvailable(dbID string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	db, err := s.client.Database().GetDB(ctx, dbID)
	if err != nil {
		s.t.Fatalf("Failed to get database %s: %v", dbID, err)
	}

	if db.Status != "available" {
		s.t.Fatalf("Expected database %s to be available, got status: %s", dbID, db.Status)
	}
}

// WaitForVMState waits for a VM to reach the specified state
func (s *IntegrationSuite) WaitForVMState(vmID string, expectedState string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.t.Fatalf("Timeout waiting for VM %s to reach state %s", vmID, expectedState)
		case <-ticker.C:
			vm, err := s.client.Compute().GetVM(ctx, vmID)
			if err != nil {
				s.t.Logf("Error getting VM %s: %v", vmID, err)
				continue
			}

			if vm.State == expectedState {
				return // Success
			}

			s.t.Logf("VM %s current state: %s, waiting for: %s", vmID, vm.State, expectedState)
		}
	}
}

// WaitForDBState waits for a database to reach the specified state
func (s *IntegrationSuite) WaitForDBState(dbID string, expectedState string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.t.Fatalf("Timeout waiting for database %s to reach state %s", dbID, expectedState)
		case <-ticker.C:
			db, err := s.client.Database().GetDB(ctx, dbID)
			if err != nil {
				s.t.Logf("Error getting database %s: %v", dbID, err)
				continue
			}

			if db.Status == expectedState {
				return // Success
			}

			s.t.Logf("Database %s current state: %s, waiting for: %s", dbID, db.Status, expectedState)
		}
	}
}

// Cleanup removes all created resources
func (s *IntegrationSuite) Cleanup() {
	if !s.cleanup {
		s.t.Log("Cleanup disabled, skipping resource cleanup")
		return
	}

	s.t.Log("Starting integration test cleanup...")

	// Cleanup VMs
	s.cleanupVMs()

	// Cleanup buckets
	s.cleanupBuckets()

	// Cleanup databases
	s.cleanupDBs()

	s.t.Log("Integration test cleanup completed")
}

// cleanupVMs removes all created VMs
func (s *IntegrationSuite) cleanupVMs() {
	s.mu.RLock()
	vms := make([]string, len(s.createdVMs))
	copy(vms, s.createdVMs)
	s.mu.RUnlock()

	for _, vmID := range vms {
		s.t.Logf("Cleaning up VM: %s", vmID)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

		if err := s.client.Compute().DeleteVM(ctx, vmID); err != nil {
			s.t.Logf("Failed to delete VM %s: %v", vmID, err)
		}

		cancel()
	}
}

// cleanupBuckets removes all created buckets
func (s *IntegrationSuite) cleanupBuckets() {
	s.mu.RLock()
	buckets := make([]string, len(s.createdBuckets))
	copy(buckets, s.createdBuckets)
	s.mu.RUnlock()

	for _, bucketName := range buckets {
		s.t.Logf("Cleaning up bucket: %s", bucketName)
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

		// First, try to delete all objects in the bucket
		if objects, err := s.client.Storage().ListObjects(ctx, bucketName); err == nil {
			for _, obj := range objects {
				if err := s.client.Storage().DeleteObject(ctx, bucketName, obj.Key); err != nil {
					s.t.Logf("Failed to delete object %s from bucket %s: %v", obj.Key, bucketName, err)
				}
			}
		}

		// Then delete the bucket
		if err := s.client.Storage().DeleteBucket(ctx, bucketName); err != nil {
			s.t.Logf("Failed to delete bucket %s: %v", bucketName, err)
		}

		cancel()
	}
}

// cleanupDBs removes all created databases
func (s *IntegrationSuite) cleanupDBs() {
	s.mu.RLock()
	dbs := make([]string, len(s.createdDBs))
	copy(dbs, s.createdDBs)
	s.mu.RUnlock()

	for _, dbID := range dbs {
		s.t.Logf("Cleaning up database: %s", dbID)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

		if err := s.client.Database().DeleteDB(ctx, dbID); err != nil {
			s.t.Logf("Failed to delete database %s: %v", dbID, err)
		}

		cancel()
	}
}

// Performance testing utilities

// MeasureOperationLatency measures the latency of a cloud operation
func (s *IntegrationSuite) MeasureOperationLatency(name string, operation func() error) time.Duration {
	start := time.Now()
	err := operation()
	duration := time.Since(start)

	if err != nil {
		s.t.Fatalf("Operation %s failed: %v", name, err)
	}

	s.t.Logf("Operation %s completed in %v", name, duration)
	return duration
}

// BenchmarkVMCreation benchmarks VM creation performance
func (s *IntegrationSuite) BenchmarkVMCreation(count int) []time.Duration {
	durations := make([]time.Duration, count)

	for i := 0; i < count; i++ {
		name := fmt.Sprintf("benchmark-vm-%d", i)
		var vm *services.VM
		duration := s.MeasureOperationLatency("CreateVM", func() error {
			vm = s.CreateTestVM(name)
			return nil
		})
		s.t.Logf("Created VM %s (%s) in %v", vm.Name, vm.ID, duration)
		durations[i] = duration
	}

	return durations
}

// TestConcurrentOperations tests concurrent cloud operations
func (s *IntegrationSuite) TestConcurrentOperations(concurrency int, operation func(id int) error) {
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			errors <- operation(id)
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < concurrency; i++ {
		if err := <-errors; err != nil {
			s.t.Fatalf("Concurrent operation %d failed: %v", i, err)
		}
	}
}

// Reliability testing

// TestOperationReliability tests operation reliability with retries
func (s *IntegrationSuite) TestOperationReliability(operation func() error, maxRetries int) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := operation(); err == nil {
			if attempt > 0 {
				s.t.Logf("Operation succeeded after %d retries", attempt)
			}
			return // Success
		} else {
			lastErr = err
			if attempt < maxRetries {
				s.t.Logf("Operation failed (attempt %d/%d): %v", attempt+1, maxRetries+1, err)
				time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
			}
		}
	}

	s.t.Fatalf("Operation failed after %d retries: %v", maxRetries+1, lastErr)
}

// Resource validation

// ValidateResourceState validates that resources are in expected states
func (s *IntegrationSuite) ValidateResourceState() {
	s.t.Log("Validating resource states...")

	// Validate VMs
	s.mu.RLock()
	for _, vmID := range s.createdVMs {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if vm, err := s.client.Compute().GetVM(ctx, vmID); err != nil {
			s.t.Errorf("Failed to validate VM %s: %v", vmID, err)
		} else {
			s.t.Logf("VM %s is in state: %s", vmID, vm.State)
		}
		cancel()
	}

	// Validate databases
	for _, dbID := range s.createdDBs {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if db, err := s.client.Database().GetDB(ctx, dbID); err != nil {
			s.t.Errorf("Failed to validate database %s: %v", dbID, err)
		} else {
			s.t.Logf("Database %s is in status: %s", dbID, db.Status)
		}
		cancel()
	}
	s.mu.RUnlock()
}
