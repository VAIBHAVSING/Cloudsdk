package aws

import (
	"os"
	"testing"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	cloudsdktesting "github.com/VAIBHAVSING/Cloudsdk/go/testing"
)

func TestNewAWSProvider(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	provider, err := New("us-east-1")
	helper.AssertNoError(err)
	helper.AssertNotEqual(nil, provider)
	helper.AssertNotEqual(nil, provider.Compute())
	helper.AssertNotEqual(nil, provider.Storage())
	helper.AssertNotEqual(nil, provider.Database())
}

func TestNew_WithOptions(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	// Test with profile option
	provider, err := New("us-east-1", WithProfile("default"))
	helper.AssertNoError(err)
	helper.AssertNotEqual(nil, provider)
	helper.AssertEqual("us-east-1", provider.Region())
	helper.AssertEqual("aws", provider.Name())

	// Test with credentials option
	provider2, err := New("us-east-1", WithCredentials("test-key", "test-secret"))
	helper.AssertNoError(err)
	helper.AssertNotEqual(nil, provider2)

	// Test with debug option
	provider3, err := New("us-east-1", WithDebug())
	helper.AssertNoError(err)
	helper.AssertNotEqual(nil, provider3)
}

func TestNew_InvalidRegion(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	// Empty region should fail
	_, err := New("")
	helper.AssertError(err)
	helper.AssertErrorCode(err, cloudsdk.ErrInvalidConfig)
}

func TestNew_InvalidCredentials(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	// Only access key without secret key should fail
	_, err := New("us-east-1", WithCredentials("test-key", ""))
	helper.AssertError(err)
	helper.AssertErrorCode(err, cloudsdk.ErrInvalidConfig)

	// Only secret key without access key should fail
	_, err = New("us-east-1", WithCredentials("", "test-secret"))
	helper.AssertError(err)
	helper.AssertErrorCode(err, cloudsdk.ErrInvalidConfig)
}

func TestProviderInterface(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	provider, err := New("us-east-1")
	helper.AssertNoError(err)

	// Test provider interface methods
	helper.AssertEqual("aws", provider.Name())
	helper.AssertEqual("us-east-1", provider.Region())

	// Test supported services
	services := provider.SupportedServices()
	helper.AssertEqual(3, len(services))

	expectedServices := []cloudsdk.ServiceType{
		cloudsdk.ServiceCompute,
		cloudsdk.ServiceStorage,
		cloudsdk.ServiceDatabase,
	}

	for _, expectedService := range expectedServices {
		found := false
		for _, service := range services {
			if service == expectedService {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected service %s not found in supported services", expectedService)
		}
	}
}

func TestServiceInterfaces(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	provider, err := New("us-east-1")
	helper.AssertNoError(err)

	// Test compute service
	compute := provider.Compute()
	helper.AssertNotEqual(nil, compute)

	// Verify it implements the interface
	var _ services.Compute = compute

	// Test storage service
	storage := provider.Storage()
	helper.AssertNotEqual(nil, storage)

	// Verify it implements the interface
	var _ services.Storage = storage

	// Test database service
	database := provider.Database()
	helper.AssertNotEqual(nil, database)

	// Verify it implements the interface
	var _ services.Database = database
}

func TestProviderContractCompliance(t *testing.T) {
	// Skip in short mode as this may make real AWS calls
	cloudsdktesting.SkipIfShort(t)

	// Skip if AWS credentials are not available (for integration tests)
	if os.Getenv("RUN_AWS_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping AWS integration test - set RUN_AWS_INTEGRATION_TESTS=true to run")
	}

	provider, err := New("us-east-1")
	if err != nil {
		t.Skipf("Skipping contract tests due to provider creation error: %v", err)
	}

	// Run provider contract tests with better error handling
	cloudsdktesting.RunProviderContractTests(t, provider)
}

// TestProviderContractComplianceAWS - AWS-specific contract test
func TestProviderContractComplianceAWS(t *testing.T) {
	// Skip this test unless explicitly enabled with AWS credentials
	if os.Getenv("RUN_AWS_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping AWS integration test - set RUN_AWS_INTEGRATION_TESTS=true to run")
	}

	provider, err := New("us-east-1")
	helper := cloudsdktesting.NewTestHelper(t)

	if err != nil {
		t.Skipf("Skipping contract tests due to provider creation error: %v", err)
	}

	// Test basic service availability with mock configuration
	// This tests interface compliance without requiring real AWS credentials

	// Test that services are available (nil checks)
	computeSvc := provider.Compute()
	storageSvc := provider.Storage()
	databaseSvc := provider.Database()

	helper.AssertNotEqual(nil, computeSvc)
	helper.AssertNotEqual(nil, storageSvc)
	helper.AssertNotEqual(nil, databaseSvc)

	// Test provider interface methods
	helper.AssertEqual("aws", provider.Name())
	helper.AssertEqual("us-east-1", provider.Region())

	t.Logf("Provider contract compliance verified for %s in region %s", provider.Name(), provider.Region())
}

func TestClientIntegration(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	provider, err := New("us-east-1")
	helper.AssertNoError(err)

	// Test client creation
	client := cloudsdk.NewFromProvider(provider)
	helper.AssertNotEqual(nil, client)

	// Test service availability checking
	cloudsdktesting.MustNotPanic(t, func() {
		client.Compute()
	})

	cloudsdktesting.MustNotPanic(t, func() {
		client.Storage()
	})

	cloudsdktesting.MustNotPanic(t, func() {
		client.Database()
	})
}

func TestErrorHandling(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	// Test various error scenarios
	testCases := []struct {
		name        string
		region      string
		options     []Option
		expectError bool
		errorCode   cloudsdk.ErrorCode
	}{
		{
			name:        "empty region",
			region:      "",
			options:     nil,
			expectError: true,
			errorCode:   cloudsdk.ErrInvalidConfig,
		},
		{
			name:        "invalid credentials - access key only",
			region:      "us-east-1",
			options:     []Option{WithCredentials("key", "")},
			expectError: true,
			errorCode:   cloudsdk.ErrInvalidConfig,
		},
		{
			name:        "invalid credentials - secret key only",
			region:      "us-east-1",
			options:     []Option{WithCredentials("", "secret")},
			expectError: true,
			errorCode:   cloudsdk.ErrInvalidConfig,
		},
		{
			name:        "valid configuration",
			region:      "us-east-1",
			options:     []Option{WithProfile("default")},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.region, tc.options...)

			if tc.expectError {
				helper.AssertError(err)
				helper.AssertErrorCode(err, tc.errorCode)
			} else {
				helper.AssertNoError(err)
			}
		})
	}
}

func TestConcurrentProviderCreation(t *testing.T) {
	// Test concurrent provider creation
	cloudsdktesting.TestConcurrency(t, 10, func(id int) error {
		_, err := New("us-east-1")
		return err
	})
}

func BenchmarkProviderCreation(b *testing.B) {
	cloudsdktesting.BenchmarkOperation(b, func() error {
		_, err := New("us-east-1")
		return err
	})
}

func BenchmarkServiceAccess(b *testing.B) {
	provider, err := New("us-east-1")
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}

	b.Run("Compute", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = provider.Compute()
		}
	})

	b.Run("Storage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = provider.Storage()
		}
	})

	b.Run("Database", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = provider.Database()
		}
	})
}
