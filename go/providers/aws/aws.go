package aws

import (
	"context"

	"github.com/VAIBHAVSING/Cloudsdk/go/providers/aws/compute"
	"github.com/VAIBHAVSING/Cloudsdk/go/providers/aws/database"
	"github.com/VAIBHAVSING/Cloudsdk/go/providers/aws/storage"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// AWSProvider implements the Provider interface for AWS
type AWSProvider struct {
	cfg aws.Config
}

// NewAWSProvider creates a new AWS provider with default configuration
func NewAWSProvider(ctx context.Context, region string) (*AWSProvider, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &AWSProvider{cfg: cfg}, nil
}

// Create AWS provider with cleaner API similar to Vercel AI SDK
func Create(ctx context.Context, region string) (*AWSProvider, error) {
	return NewAWSProvider(ctx, region)
}

// Compute returns the AWS compute service
func (p *AWSProvider) Compute() services.Compute {
	return compute.New(p.cfg)
}

// Storage returns the AWS storage service
func (p *AWSProvider) Storage() services.Storage {
	return storage.New(p.cfg)
}

// Database returns the AWS database service
func (p *AWSProvider) Database() services.Database {
	return database.New(p.cfg)
}
