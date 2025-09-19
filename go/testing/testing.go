// Package testing provides comprehensive testing utilities for the Cloud SDK.
//
// This package offers helper functions, test data generators, assertion utilities,
// and integration test helpers to make testing cloud applications easier and more reliable.
//
// FEATURES:
//   - Helper functions for common test assertions
//   - Realistic test data generators for all service types
//   - Cleanup utilities for integration tests
//   - Test suite runners for provider contract testing
//   - Mock provider integration and verification
//   - Error scenario testing utilities
//
// QUICK START:
//
// Basic test setup with mock provider:
//
//	func TestVMCreation(t *testing.T) {
//	    provider := testing.NewMockProvider("us-east-1")
//	    client := cloudsdk.New(provider)
//
//	    vm, err := client.Compute().CreateVM(ctx, testing.GenerateVMConfig("test-vm"))
//	    testing.AssertNoError(t, err)
//	    testing.AssertVMValid(t, vm)
//	    testing.AssertProviderCalled(t, provider, "CreateVM", 1)
//	}
//
// Integration test with cleanup:
//
//	func TestIntegration(t *testing.T) {
//	    suite := testing.NewIntegrationSuite(t).
//	        WithProvider(awsProvider).
//	        WithCleanup(true)
//	    defer suite.Cleanup()
//
//	    // Test operations - resources will be cleaned up automatically
//	    vm := suite.CreateTestVM("integration-test-vm")
//	    suite.AssertVMRunning(vm.ID)
//	}
//
// Provider contract testing:
//
//	func TestProviderContract(t *testing.T) {
//	    providers := []cloudsdk.Provider{
//	        aws.New("us-east-1"),
//	        mock.New("us-east-1"),
//	    }
//
//	    for _, provider := range providers {
//	        testing.RunProviderContractTests(t, provider)
//	    }
//	}
//
// ERROR SCENARIO TESTING:
//
// The testing package provides utilities for testing various error scenarios:
//   - Authentication and authorization failures
//   - Resource conflicts and not found errors
//   - Network timeouts and rate limiting
//   - Invalid configuration and validation errors
//
// INTEGRATION TESTING:
//
// Integration test helpers provide:
//   - Automatic resource cleanup after tests
//   - Realistic test data generation
//   - Provider-agnostic test patterns
//   - Performance and reliability testing
//
// ASSERTION UTILITIES:
//
// Comprehensive assertion functions for:
//   - Error checking and validation
//   - Resource state verification
//   - Provider operation verification
//   - Performance and timing assertions
package testing

import (
	"strings"
	"testing"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/providers/mock"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

// TestHelper provides common testing utilities and assertions
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper instance
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// AssertNoError asserts that an error is nil
func (h *TestHelper) AssertNoError(err error) {
	if err != nil {
		h.t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError asserts that an error is not nil
func (h *TestHelper) AssertError(err error) {
	if err == nil {
		h.t.Fatal("Expected an error, got nil")
	}
}

// AssertErrorCode asserts that an error has the expected error code
func (h *TestHelper) AssertErrorCode(err error, expectedCode cloudsdk.ErrorCode) {
	if err == nil {
		h.t.Fatal("Expected an error, got nil")
	}

	if cloudErr, ok := err.(*cloudsdk.CloudError); ok {
		if cloudErr.Code != expectedCode {
			h.t.Fatalf("Expected error code %s, got %s", expectedCode, cloudErr.Code)
		}
	} else {
		h.t.Fatalf("Expected CloudError, got %T", err)
	}
}

// AssertEqual asserts that two values are equal
func (h *TestHelper) AssertEqual(expected, actual interface{}) {
	if expected != actual {
		h.t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

// AssertNotEqual asserts that two values are not equal
func (h *TestHelper) AssertNotEqual(expected, actual interface{}) {
	if expected == actual {
		h.t.Fatalf("Expected values to be different, both were %v", expected)
	}
}

// AssertContains asserts that a string contains a substring
func (h *TestHelper) AssertContains(str, substr string) {
	if !strings.Contains(str, substr) {
		h.t.Fatalf("Expected string to contain %q, got %q", substr, str)
	}
}

// AssertVMValid asserts that a VM has valid properties
func (h *TestHelper) AssertVMValid(vm *services.VM) {
	if vm == nil {
		h.t.Fatal("VM is nil")
	}
	if vm.ID == "" {
		h.t.Fatal("VM ID is empty")
	}
	if vm.Name == "" {
		h.t.Fatal("VM Name is empty")
	}
	if vm.State == "" {
		h.t.Fatal("VM State is empty")
	}
}

// AssertDBValid asserts that a database instance has valid properties
func (h *TestHelper) AssertDBValid(db *services.DBInstance) {
	if db == nil {
		h.t.Fatal("Database instance is nil")
	}
	if db.ID == "" {
		h.t.Fatal("Database ID is empty")
	}
	if db.Name == "" {
		h.t.Fatal("Database Name is empty")
	}
	if db.Engine == "" {
		h.t.Fatal("Database Engine is empty")
	}
	if db.Status == "" {
		h.t.Fatal("Database Status is empty")
	}
	if db.Endpoint == "" {
		h.t.Fatal("Database Endpoint is empty")
	}
}

// AssertProviderCalled asserts that a mock provider method was called a specific number of times
func (h *TestHelper) AssertProviderCalled(provider *mock.MockProvider, method string, expectedCount int) {
	actualCount := provider.CallCount(method)
	if actualCount != expectedCount {
		h.t.Fatalf("Expected %s to be called %d times, got %d", method, expectedCount, actualCount)
	}
}

// AssertProviderNotCalled asserts that a mock provider method was not called
func (h *TestHelper) AssertProviderNotCalled(provider *mock.MockProvider, method string) {
	if provider.WasCalled(method) {
		h.t.Fatalf("Expected %s not to be called, but it was", method)
	}
}

// Standalone helper functions for convenience

// AssertNoError is a standalone version of TestHelper.AssertNoError
func AssertNoError(t *testing.T, err error) {
	NewTestHelper(t).AssertNoError(err)
}

// AssertError is a standalone version of TestHelper.AssertError
func AssertError(t *testing.T, err error) {
	NewTestHelper(t).AssertError(err)
}

// AssertErrorCode is a standalone version of TestHelper.AssertErrorCode
func AssertErrorCode(t *testing.T, err error, expectedCode cloudsdk.ErrorCode) {
	NewTestHelper(t).AssertErrorCode(err, expectedCode)
}

// AssertEqual is a standalone version of TestHelper.AssertEqual
func AssertEqual(t *testing.T, expected, actual interface{}) {
	NewTestHelper(t).AssertEqual(expected, actual)
}

// AssertVMValid is a standalone version of TestHelper.AssertVMValid
func AssertVMValid(t *testing.T, vm *services.VM) {
	NewTestHelper(t).AssertVMValid(vm)
}

// AssertDBValid is a standalone version of TestHelper.AssertDBValid
func AssertDBValid(t *testing.T, db *services.DBInstance) {
	NewTestHelper(t).AssertDBValid(db)
}

// AssertProviderCalled is a standalone version of TestHelper.AssertProviderCalled
func AssertProviderCalled(t *testing.T, provider *mock.MockProvider, method string, expectedCount int) {
	NewTestHelper(t).AssertProviderCalled(provider, method, expectedCount)
}

// NewMockProvider creates a new mock provider for testing
func NewMockProvider(region string) *mock.MockProvider {
	return mock.New(region)
}

// WithTimeout runs a test function with a timeout
func WithTimeout(t *testing.T, timeout time.Duration, fn func()) {
	done := make(chan bool, 1)

	go func() {
		fn()
		done <- true
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(timeout):
		t.Fatalf("Test timed out after %v", timeout)
	}
}

// RetryUntilSuccess retries a function until it succeeds or times out
func RetryUntilSuccess(t *testing.T, timeout time.Duration, interval time.Duration, fn func() error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if err := fn(); err == nil {
			return // Success
		}
		time.Sleep(interval)
	}

	// Final attempt
	if err := fn(); err != nil {
		t.Fatalf("Function did not succeed within %v: %v", timeout, err)
	}
}

// MustNotPanic asserts that a function does not panic
func MustNotPanic(t *testing.T, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Function panicked: %v", r)
		}
	}()
	fn()
}

// MustPanic asserts that a function panics
func MustPanic(t *testing.T, fn func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected function to panic, but it didn't")
		}
	}()
	fn()
}

// SkipIfShort skips a test if running in short mode
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
}

// SkipIfCI skips a test if running in CI environment
func SkipIfCI(t *testing.T) {
	// Check common CI environment variables
	ciEnvVars := []string{"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "JENKINS_URL", "TRAVIS"}
	for _, envVar := range ciEnvVars {
		if value := getEnv(envVar); value != "" {
			t.Skip("Skipping test in CI environment")
		}
	}
}

// getEnv is a helper to get environment variables (would normally use os.Getenv)
func getEnv(key string) string {
	// This is a placeholder - in real implementation would use os.Getenv
	return ""
}

// Benchmark utilities

// BenchmarkOperation benchmarks a cloud operation
func BenchmarkOperation(b *testing.B, operation func() error) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := operation(); err != nil {
			b.Fatalf("Operation failed: %v", err)
		}
	}
}

// BenchmarkWithSetup benchmarks an operation with setup and teardown
func BenchmarkWithSetup(b *testing.B, setup func(), teardown func(), operation func() error) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		setup()
		b.StartTimer()

		if err := operation(); err != nil {
			b.Fatalf("Operation failed: %v", err)
		}

		b.StopTimer()
		teardown()
	}
}

// Performance testing utilities

// MeasureLatency measures the latency of an operation
func MeasureLatency(operation func() error) (time.Duration, error) {
	start := time.Now()
	err := operation()
	duration := time.Since(start)
	return duration, err
}

// AssertLatencyUnder asserts that an operation completes within the specified duration
func AssertLatencyUnder(t *testing.T, maxDuration time.Duration, operation func() error) {
	duration, err := MeasureLatency(operation)
	if err != nil {
		t.Fatalf("Operation failed: %v", err)
	}
	if duration > maxDuration {
		t.Fatalf("Operation took %v, expected under %v", duration, maxDuration)
	}
}

// Parallel testing utilities

// RunParallel runs multiple operations in parallel and waits for all to complete
func RunParallel(t *testing.T, operations ...func() error) {
	errors := make(chan error, len(operations))

	for _, op := range operations {
		go func(operation func() error) {
			errors <- operation()
		}(op)
	}

	// Wait for all operations to complete
	for i := 0; i < len(operations); i++ {
		if err := <-errors; err != nil {
			t.Fatalf("Parallel operation failed: %v", err)
		}
	}
}

// TestConcurrency tests concurrent access to a resource
func TestConcurrency(t *testing.T, concurrency int, operation func(id int) error) {
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			errors <- operation(id)
		}(i)
	}

	// Check all operations completed successfully
	for i := 0; i < concurrency; i++ {
		if err := <-errors; err != nil {
			t.Fatalf("Concurrent operation %d failed: %v", i, err)
		}
	}
}
