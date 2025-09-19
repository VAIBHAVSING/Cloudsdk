package database

import (
	"context"
	"testing"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	cloudsdktesting "github.com/VAIBHAVSING/Cloudsdk/go/testing"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// mockRDSClient is a mock implementation of the RDS client
type mockRDSClient struct {
	createDBInstanceResponse    *rds.CreateDBInstanceOutput
	createDBInstanceError       error
	describeDBInstancesResponse *rds.DescribeDBInstancesOutput
	describeDBInstancesError    error
	deleteDBInstanceResponse    *rds.DeleteDBInstanceOutput
	deleteDBInstanceError       error
}

func (m *mockRDSClient) CreateDBInstance(ctx context.Context, input *rds.CreateDBInstanceInput, opts ...func(*rds.Options)) (*rds.CreateDBInstanceOutput, error) {
	return m.createDBInstanceResponse, m.createDBInstanceError
}

func (m *mockRDSClient) DescribeDBInstances(ctx context.Context, input *rds.DescribeDBInstancesInput, opts ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	return m.describeDBInstancesResponse, m.describeDBInstancesError
}

func (m *mockRDSClient) DeleteDBInstance(ctx context.Context, input *rds.DeleteDBInstanceInput, opts ...func(*rds.Options)) (*rds.DeleteDBInstanceOutput, error) {
	return m.deleteDBInstanceResponse, m.deleteDBInstanceError
}

func TestAWSDatabase_CreateDB(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	mockClient := &mockRDSClient{
		createDBInstanceResponse: &rds.CreateDBInstanceOutput{
			DBInstance: &types.DBInstance{
				DBInstanceIdentifier: stringPtr("test-db"),
				DBInstanceStatus:     stringPtr("creating"),
				Engine:               stringPtr("mysql"),
				EngineVersion:        stringPtr("8.0.35"),
				Endpoint: &types.Endpoint{
					Address: stringPtr("test-db.rds.amazonaws.com"),
					Port:    int32Ptr(3306),
				},
				InstanceCreateTime: &time.Time{},
			},
		},
		createDBInstanceError: nil,
	}

	database := NewWithClient(mockClient)
	config := cloudsdktesting.GenerateDBConfig("test-db")

	db, err := database.CreateDB(context.Background(), config)
	helper.AssertNoError(err)
	cloudsdktesting.AssertDBValid(t, db)
	helper.AssertEqual("test-db", db.ID)
	helper.AssertEqual("test-db", db.Name)
	helper.AssertContains(db.Engine, "mysql")
	helper.AssertEqual("creating", db.Status)
	helper.AssertContains(db.Endpoint, "test-db.rds.amazonaws.com")
}

func TestAWSDatabase_CreateDB_ErrorScenarios(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	testCases := []struct {
		name      string
		mockError error
	}{
		{
			name:      "db instance already exists",
			mockError: &types.DBInstanceAlreadyExistsFault{},
		},
		{
			name:      "insufficient db instance capacity",
			mockError: &types.InsufficientDBInstanceCapacityFault{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockRDSClient{
				createDBInstanceError: tc.mockError,
			}

			database := NewWithClient(mockClient)
			config := cloudsdktesting.GenerateDBConfig("test-db")

			_, err := database.CreateDB(context.Background(), config)
			helper.AssertError(err)
		})
	}
}

func TestAWSDatabase_ListDBs(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	mockClient := &mockRDSClient{
		describeDBInstancesResponse: &rds.DescribeDBInstancesOutput{
			DBInstances: []types.DBInstance{
				{
					DBInstanceIdentifier: stringPtr("db1"),
					DBInstanceStatus:     stringPtr("available"),
					Engine:               stringPtr("postgres"),
					EngineVersion:        stringPtr("14.9"),
					Endpoint: &types.Endpoint{
						Address: stringPtr("db1.rds.amazonaws.com"),
						Port:    int32Ptr(5432),
					},
					InstanceCreateTime: &time.Time{},
				},
				{
					DBInstanceIdentifier: stringPtr("db2"),
					DBInstanceStatus:     stringPtr("available"),
					Engine:               stringPtr("mysql"),
					EngineVersion:        stringPtr("8.0.35"),
					Endpoint: &types.Endpoint{
						Address: stringPtr("db2.rds.amazonaws.com"),
						Port:    int32Ptr(3306),
					},
					InstanceCreateTime: &time.Time{},
				},
			},
		},
		describeDBInstancesError: nil,
	}

	database := NewWithClient(mockClient)

	dbs, err := database.ListDBs(context.Background())
	helper.AssertNoError(err)
	helper.AssertEqual(2, len(dbs))

	// Validate first database
	cloudsdktesting.AssertDBValid(t, dbs[0])
	helper.AssertEqual("db1", dbs[0].ID)
	helper.AssertContains(dbs[0].Engine, "postgres")

	// Validate second database
	cloudsdktesting.AssertDBValid(t, dbs[1])
	helper.AssertEqual("db2", dbs[1].ID)
	helper.AssertContains(dbs[1].Engine, "mysql")
}

func TestAWSDatabase_GetDB(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	mockClient := &mockRDSClient{
		describeDBInstancesResponse: &rds.DescribeDBInstancesOutput{
			DBInstances: []types.DBInstance{
				{
					DBInstanceIdentifier: stringPtr("test-db"),
					DBInstanceStatus:     stringPtr("available"),
					Engine:               stringPtr("postgres"),
					EngineVersion:        stringPtr("14.9"),
					Endpoint: &types.Endpoint{
						Address: stringPtr("test-db.rds.amazonaws.com"),
						Port:    int32Ptr(5432),
					},
					InstanceCreateTime: &time.Time{},
				},
			},
		},
		describeDBInstancesError: nil,
	}

	database := NewWithClient(mockClient)

	db, err := database.GetDB(context.Background(), "test-db")
	helper.AssertNoError(err)
	cloudsdktesting.AssertDBValid(t, db)
	helper.AssertEqual("test-db", db.ID)
	helper.AssertEqual("available", db.Status)
	helper.AssertContains(db.Engine, "postgres")
}

func TestAWSDatabase_DeleteDB(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	mockClient := &mockRDSClient{
		deleteDBInstanceResponse: &rds.DeleteDBInstanceOutput{},
		deleteDBInstanceError:    nil,
	}

	database := NewWithClient(mockClient)

	err := database.DeleteDB(context.Background(), "test-db")
	helper.AssertNoError(err)
}

func TestAWSDatabase_DatabaseLifecycle(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	mockClient := &mockRDSClient{
		createDBInstanceResponse: &rds.CreateDBInstanceOutput{
			DBInstance: &types.DBInstance{
				DBInstanceIdentifier: stringPtr("lifecycle-test-db"),
				DBInstanceStatus:     stringPtr("creating"),
				Engine:               stringPtr("postgres"),
				EngineVersion:        stringPtr("14.9"),
				Endpoint: &types.Endpoint{
					Address: stringPtr("lifecycle-test-db.rds.amazonaws.com"),
					Port:    int32Ptr(5432),
				},
				InstanceCreateTime: &time.Time{},
			},
		},
		createDBInstanceError: nil,
		describeDBInstancesResponse: &rds.DescribeDBInstancesOutput{
			DBInstances: []types.DBInstance{
				{
					DBInstanceIdentifier: stringPtr("lifecycle-test-db"),
					DBInstanceStatus:     stringPtr("available"),
					Engine:               stringPtr("postgres"),
					EngineVersion:        stringPtr("14.9"),
					Endpoint: &types.Endpoint{
						Address: stringPtr("lifecycle-test-db.rds.amazonaws.com"),
						Port:    int32Ptr(5432),
					},
					InstanceCreateTime: &time.Time{},
				},
			},
		},
		describeDBInstancesError: nil,
		deleteDBInstanceResponse: &rds.DeleteDBInstanceOutput{},
		deleteDBInstanceError:    nil,
	}

	database := NewWithClient(mockClient)
	config := cloudsdktesting.GenerateDBConfig("lifecycle-test-db")

	// Create database
	db, err := database.CreateDB(context.Background(), config)
	helper.AssertNoError(err)
	cloudsdktesting.AssertDBValid(t, db)

	// Get database
	retrievedDB, err := database.GetDB(context.Background(), db.ID)
	helper.AssertNoError(err)
	cloudsdktesting.AssertDBValid(t, retrievedDB)
	helper.AssertEqual(db.ID, retrievedDB.ID)

	// List databases (should include our database)
	dbs, err := database.ListDBs(context.Background())
	helper.AssertNoError(err)
	helper.AssertEqual(1, len(dbs))

	// Delete database
	err = database.DeleteDB(context.Background(), db.ID)
	helper.AssertNoError(err)
}

func TestAWSDatabase_ConcurrentOperations(t *testing.T) {
	mockClient := &mockRDSClient{
		describeDBInstancesResponse: &rds.DescribeDBInstancesOutput{
			DBInstances: []types.DBInstance{},
		},
		describeDBInstancesError: nil,
	}

	database := NewWithClient(mockClient)

	// Test concurrent ListDBs calls
	cloudsdktesting.TestConcurrency(t, 10, func(id int) error {
		_, err := database.ListDBs(context.Background())
		return err
	})
}

func BenchmarkAWSDatabase_ListDBs(b *testing.B) {
	mockClient := &mockRDSClient{
		describeDBInstancesResponse: &rds.DescribeDBInstancesOutput{
			DBInstances: []types.DBInstance{},
		},
		describeDBInstancesError: nil,
	}

	database := NewWithClient(mockClient)

	cloudsdktesting.BenchmarkOperation(b, func() error {
		_, err := database.ListDBs(context.Background())
		return err
	})
}

func TestAWSDatabase_WithMockProvider(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	// Test using our mock provider for comparison
	mockProvider := cloudsdktesting.NewMockProvider("us-east-1")
	client := cloudsdk.NewFromProvider(mockProvider)

	config := cloudsdktesting.GenerateDBConfig("mock-test-db")

	// Create database
	db, err := client.Database().CreateDB(context.Background(), config)
	helper.AssertNoError(err)
	cloudsdktesting.AssertDBValid(t, db)

	// Verify mock provider recorded the operation
	cloudsdktesting.AssertProviderCalled(t, mockProvider, "CreateDB", 1)

	// List databases
	dbs, err := client.Database().ListDBs(context.Background())
	helper.AssertNoError(err)
	helper.AssertEqual(1, len(dbs))
	helper.AssertEqual(db.ID, dbs[0].ID)
}

func TestAWSDatabase_ErrorInjection(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	// Test error injection with mock provider
	mockProvider := cloudsdktesting.NewMockProvider("us-east-1").
		WithError("CreateDB", cloudsdk.NewCloudError(
			cloudsdk.ErrResourceConflict,
			"Database instance already exists",
			"mock", "database", "CreateDB"))

	client := cloudsdk.NewFromProvider(mockProvider)
	config := cloudsdktesting.GenerateDBConfig("error-test-db")

	_, err := client.Database().CreateDB(context.Background(), config)
	helper.AssertError(err)
	helper.AssertErrorCode(err, cloudsdk.ErrResourceConflict)
}

func TestAWSDatabase_ProductionConfig(t *testing.T) {
	helper := cloudsdktesting.NewTestHelper(t)

	mockClient := &mockRDSClient{
		createDBInstanceResponse: &rds.CreateDBInstanceOutput{
			DBInstance: &types.DBInstance{
				DBInstanceIdentifier: stringPtr("prod-db"),
				DBInstanceStatus:     stringPtr("creating"),
				Engine:               stringPtr("postgres"),
				EngineVersion:        stringPtr("14.9"),
				Endpoint: &types.Endpoint{
					Address: stringPtr("prod-db.rds.amazonaws.com"),
					Port:    int32Ptr(5432),
				},
				InstanceCreateTime: &time.Time{},
			},
		},
		createDBInstanceError: nil,
	}

	database := NewWithClient(mockClient)

	// Test with production-like configuration
	config := cloudsdktesting.GenerateProductionDBConfig("prod-db")

	db, err := database.CreateDB(context.Background(), config)
	helper.AssertNoError(err)
	cloudsdktesting.AssertDBValid(t, db)
	helper.AssertEqual("prod-db", db.ID)
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}
