package services

import "context"

// Tag represents a key-value pair for tagging resources
type Tag struct {
	Key   string
	Value string
}

// CreateTagsRequest represents the request for creating tags on AWS resources
type CreateTagsRequest struct {
	// ResourceIDs is a list of resource IDs or ARNs to tag
	ResourceIDs []string
	// Tags is a map of key-value pairs to create as tags
	Tags map[string]string
}

// DeleteTagsRequest represents the request for deleting tags from AWS resources
type DeleteTagsRequest struct {
	// ResourceIDs is a list of resource IDs or ARNs to remove tags from
	ResourceIDs []string
	// TagKeys is a list of tag keys to delete
	TagKeys []string
}

// DescribeTagsFilter represents filters for describing tags on AWS resources
type DescribeTagsFilter struct {
	// ResourceIDs filters tags by specific resource IDs or ARNs
	ResourceIDs []string
	// ResourceTypes filters tags by resource types (e.g., "instance", "volume")
	ResourceTypes []string
	// Keys filters tags by specific tag keys
	Keys []string
	// Values filters tags by specific tag values
	Values []string
}

// ResourceTag represents a tag associated with a resource
type ResourceTag struct {
	ResourceID   string
	ResourceType string
	Key          string
	Value        string
}

// TaggingHelper provides helper methods for tagging operations
type TaggingHelper interface {
	CreateTagsByARN(ctx context.Context, resourceARNs []string, tags map[string]string) error
	DeleteTagsByARN(ctx context.Context, resourceARNs []string, tagKeys []string) error
}

// Tagging defines the interface for resource tagging operations
type Tagging interface {
	Create(ctx context.Context, req *CreateTagsRequest) error
	Delete(ctx context.Context, req *DeleteTagsRequest) error
	List(ctx context.Context, filter *DescribeTagsFilter) ([]*ResourceTag, error)
	TaggingHelper
}