package database

import (
	"context"
	"fmt"

	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// RDSClientInterface defines methods we need from RDS client for testing
type RDSClientInterface interface {
	CreateDBInstance(ctx context.Context, input *rds.CreateDBInstanceInput, opts ...func(*rds.Options)) (*rds.CreateDBInstanceOutput, error)
	DescribeDBInstances(ctx context.Context, input *rds.DescribeDBInstancesInput, opts ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
}

// AWSDatabase implements the Database interface for AWS
type AWSDatabase struct {
	client RDSClientInterface
}

// New creates a new AWSDatabase instance with real AWS client
func New(cfg aws.Config) services.Database {
	client := rds.NewFromConfig(cfg)
	return &AWSDatabase{client: client}
}

// NewWithClient creates a new AWSDatabase instance with custom client (for testing)
func NewWithClient(client RDSClientInterface) services.Database {
	return &AWSDatabase{client: client}
}

// CreateDB creates a new RDS instance
func (d *AWSDatabase) CreateDB(ctx context.Context, config *services.DBConfig) (*services.DBInstance, error) {
	input := &rds.CreateDBInstanceInput{
		DBInstanceIdentifier: aws.String(config.Name),
		DBInstanceClass:      aws.String(config.InstanceClass),
		Engine:               aws.String(config.Engine),
		MasterUsername:       aws.String(config.MasterUsername),
		MasterUserPassword:   aws.String(config.MasterPassword),
		AllocatedStorage:     aws.Int32(config.AllocatedStorage),
	}

	if config.EngineVersion != "" {
		input.EngineVersion = aws.String(config.EngineVersion)
	}
	if config.DBName != "" {
		input.DBName = aws.String(config.DBName)
	}

	resp, err := d.client.CreateDBInstance(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create DB instance: %w", err)
	}

	if resp.DBInstance == nil {
		return nil, fmt.Errorf("no DB instance returned from AWS")
	}

	return &services.DBInstance{
		ID:         aws.ToString(resp.DBInstance.DBInstanceIdentifier),
		Name:       aws.ToString(resp.DBInstance.DBInstanceIdentifier),
		Engine:     aws.ToString(resp.DBInstance.Engine),
		Status:     aws.ToString(resp.DBInstance.DBInstanceStatus),
		Endpoint:   aws.ToString(resp.DBInstance.Endpoint.Address),
		LaunchTime: resp.DBInstance.InstanceCreateTime.String(),
	}, nil
}

// ListDBs lists all RDS instances
func (d *AWSDatabase) ListDBs(ctx context.Context) ([]*services.DBInstance, error) {
	resp, err := d.client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe DB instances: %w", err)
	}

	dbs := make([]*services.DBInstance, len(resp.DBInstances))
	for i, inst := range resp.DBInstances {
		dbs[i] = &services.DBInstance{
			ID:         aws.ToString(inst.DBInstanceIdentifier),
			Name:       aws.ToString(inst.DBInstanceIdentifier),
			Engine:     aws.ToString(inst.Engine),
			Status:     aws.ToString(inst.DBInstanceStatus),
			Endpoint:   aws.ToString(inst.Endpoint.Address),
			LaunchTime: inst.InstanceCreateTime.String(),
		}
	}
	return dbs, nil
}

// Other methods - stub for now
func (d *AWSDatabase) GetDB(ctx context.Context, id string) (*services.DBInstance, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *AWSDatabase) DeleteDB(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}
