package compute

import (
	"context"
	"fmt"

	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// AWSTagging implements the Tagging interface for AWS
type AWSTagging struct {
	client EC2ClientInterface
}

// NewTagging creates a new AWSTagging instance
func NewTagging(client EC2ClientInterface) services.Tagging {
	return &AWSTagging{client: client}
}

// Create creates tags for the specified resources
// Supports bulk operations and handles AWS EC2 CreateTags API
func (t *AWSTagging) Create(ctx context.Context, req *services.CreateTagsRequest) error {
	if len(req.ResourceIDs) == 0 {
		return fmt.Errorf("tagging: at least one resource ID must be specified")
	}
	
	if len(req.Tags) == 0 {
		return fmt.Errorf("tagging: at least one tag must be specified")
	}

	// Convert tags to AWS format
	awsTags := make([]types.Tag, 0, len(req.Tags))
	for key, value := range req.Tags {
		awsTags = append(awsTags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	input := &ec2.CreateTagsInput{
		Resources: req.ResourceIDs,
		Tags:      awsTags,
	}

	_, err := t.client.CreateTags(ctx, input)
	if err != nil {
		return fmt.Errorf("tagging: failed to create tags: %w", err)
	}

	return nil
}

// Delete deletes tags from the specified resources
// Supports bulk operations and handles AWS EC2 DeleteTags API
func (t *AWSTagging) Delete(ctx context.Context, req *services.DeleteTagsRequest) error {
	if len(req.ResourceIDs) == 0 {
		return fmt.Errorf("tagging: at least one resource ID must be specified")
	}
	
	if len(req.TagKeys) == 0 {
		return fmt.Errorf("tagging: at least one tag key must be specified")
	}

	// Convert tag keys to AWS format
	awsTags := make([]types.Tag, 0, len(req.TagKeys))
	for _, key := range req.TagKeys {
		awsTags = append(awsTags, types.Tag{
			Key: aws.String(key),
		})
	}

	input := &ec2.DeleteTagsInput{
		Resources: req.ResourceIDs,
		Tags:      awsTags,
	}

	_, err := t.client.DeleteTags(ctx, input)
	if err != nil {
		return fmt.Errorf("tagging: failed to delete tags: %w", err)
	}

	return nil
}

// List describes tags based on the provided filters
// Returns all tags if no filters are specified
func (t *AWSTagging) List(ctx context.Context, filter *services.DescribeTagsFilter) ([]*services.ResourceTag, error) {
	input := &ec2.DescribeTagsInput{}

	// Apply filters if provided
	var filters []types.Filter
	
	if len(filter.ResourceIDs) > 0 {
		filters = append(filters, types.Filter{
			Name:   aws.String("resource-id"),
			Values: filter.ResourceIDs,
		})
	}
	
	if len(filter.ResourceTypes) > 0 {
		filters = append(filters, types.Filter{
			Name:   aws.String("resource-type"),
			Values: filter.ResourceTypes,
		})
	}
	
	if len(filter.Keys) > 0 {
		filters = append(filters, types.Filter{
			Name:   aws.String("key"),
			Values: filter.Keys,
		})
	}
	
	if len(filter.Values) > 0 {
		filters = append(filters, types.Filter{
			Name:   aws.String("value"),
			Values: filter.Values,
		})
	}

	if len(filters) > 0 {
		input.Filters = filters
	}

	resp, err := t.client.DescribeTags(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("tagging: failed to describe tags: %w", err)
	}

	// Convert AWS tags to service tags
	result := make([]*services.ResourceTag, 0, len(resp.Tags))
	for _, tag := range resp.Tags {
		result = append(result, &services.ResourceTag{
			ResourceID:   aws.ToString(tag.ResourceId),
			ResourceType: string(tag.ResourceType),
			Key:          aws.ToString(tag.Key),
			Value:        aws.ToString(tag.Value),
		})
	}

	return result, nil
}

// CreateTagsByARN creates tags for resources specified by ARNs
func (t *AWSTagging) CreateTagsByARN(ctx context.Context, resourceARNs []string, tags map[string]string) error {
	// For EC2, ARNs and IDs can be used interchangeably in most tagging operations
	// This helper provides a consistent API for users who have ARNs
	return t.Create(ctx, &services.CreateTagsRequest{
		ResourceIDs: resourceARNs,
		Tags:        tags,
	})
}

// DeleteTagsByARN deletes tags from resources specified by ARNs
func (t *AWSTagging) DeleteTagsByARN(ctx context.Context, resourceARNs []string, tagKeys []string) error {
	// For EC2, ARNs and IDs can be used interchangeably in most tagging operations
	// This helper provides a consistent API for users who have ARNs
	return t.Delete(ctx, &services.DeleteTagsRequest{
		ResourceIDs: resourceARNs,
		TagKeys:     tagKeys,
	})
}