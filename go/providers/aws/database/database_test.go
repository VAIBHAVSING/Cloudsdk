package database

import (
	"context"
	"testing"
	"time"

	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/stretchr/testify/assert"
)

// mockRDSClient is a mock implementation of the RDS client
type mockRDSClient struct {
	createDBInstanceResponse    *rds.CreateDBInstanceOutput
	createDBInstanceError       error
	describeDBInstancesResponse *rds.DescribeDBInstancesOutput
	describeDBInstancesError    error
}

func (m *mockRDSClient) CreateDBInstance(ctx context.Context, input *rds.CreateDBInstanceInput, opts ...func(*rds.Options)) (*rds.CreateDBInstanceOutput, error) {
	return m.createDBInstanceResponse, m.createDBInstanceError
}

func (m *mockRDSClient) DescribeDBInstances(ctx context.Context, input *rds.DescribeDBInstancesInput, opts ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	return m.describeDBInstancesResponse, m.describeDBInstancesError
}

func TestAWSDatabase_CreateDB(t *testing.T) {
	mockClient := &mockRDSClient{
		createDBInstanceResponse: &rds.CreateDBInstanceOutput{
			DBInstance: &types.DBInstance{
				DBInstanceIdentifier: stringPtr("test-db"),
				DBInstanceStatus:     stringPtr("creating"),
				Engine:               stringPtr("mysql"),
				Endpoint: &types.Endpoint{
					Address: stringPtr("test-db.rds.amazonaws.com"),
				},
				InstanceCreateTime: &time.Time{},
			},
		},
		createDBInstanceError: nil,
	}

	database := NewWithClient(mockClient)

	config := &services.DBConfig{
		Name:             "test-db",
		Engine:           "mysql",
		InstanceClass:    "db.t2.micro",
		MasterUsername:   "admin",
		MasterPassword:   "password",
		AllocatedStorage: 20,
	}

	db, err := database.CreateDB(context.Background(), config)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, "test-db", db.ID)
	assert.Equal(t, "mysql", db.Engine)
}

func TestAWSDatabase_ListDBs(t *testing.T) {
	mockClient := &mockRDSClient{
		describeDBInstancesResponse: &rds.DescribeDBInstancesOutput{
			DBInstances: []types.DBInstance{
				{
					DBInstanceIdentifier: stringPtr("db1"),
					DBInstanceStatus:     stringPtr("available"),
					Engine:               stringPtr("postgres"),
					Endpoint: &types.Endpoint{
						Address: stringPtr("db1.rds.amazonaws.com"),
					},
					InstanceCreateTime: &time.Time{},
				},
				{
					DBInstanceIdentifier: stringPtr("db2"),
					DBInstanceStatus:     stringPtr("available"),
					Engine:               stringPtr("mysql"),
					Endpoint: &types.Endpoint{
						Address: stringPtr("db2.rds.amazonaws.com"),
					},
					InstanceCreateTime: &time.Time{},
				},
			},
		},
		describeDBInstancesError: nil,
	}

	database := NewWithClient(mockClient)

	dbs, err := database.ListDBs(context.Background())
	assert.NoError(t, err)
	assert.Len(t, dbs, 2)
	assert.Equal(t, "db1", dbs[0].ID)
	assert.Equal(t, "db2", dbs[1].ID)
}

func stringPtr(s string) *string {
	return &s
}
