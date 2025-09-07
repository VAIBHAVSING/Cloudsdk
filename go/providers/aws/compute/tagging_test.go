package compute

import (
	"context"
	"testing"

	"github.com/VAIBHAVSING/Cloudsdk/go/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

func TestAWSTagging_Create(t *testing.T) {
	tests := []struct {
		name          string
		request       *services.CreateTagsRequest
		mockResponse  *ec2.CreateTagsOutput
		mockError     error
		expectedError string
	}{
		{
			name: "successful create tags",
			request: &services.CreateTagsRequest{
				ResourceIDs: []string{"i-1234567890abcdef0"},
				Tags: map[string]string{
					"Environment": "test",
					"Project":     "cloudsdk",
				},
			},
			mockResponse:  &ec2.CreateTagsOutput{},
			mockError:     nil,
			expectedError: "",
		},
		{
			name: "empty resource IDs",
			request: &services.CreateTagsRequest{
				ResourceIDs: []string{},
				Tags: map[string]string{
					"Environment": "test",
				},
			},
			mockResponse:  nil,
			mockError:     nil,
			expectedError: "tagging: at least one resource ID must be specified",
		},
		{
			name: "empty tags",
			request: &services.CreateTagsRequest{
				ResourceIDs: []string{"i-1234567890abcdef0"},
				Tags:        map[string]string{},
			},
			mockResponse:  nil,
			mockError:     nil,
			expectedError: "tagging: at least one tag must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockEC2Client{
				createTagsResponse: tt.mockResponse,
				createTagsError:    tt.mockError,
			}

			tagging := NewTagging(mockClient)
			err := tagging.Create(context.Background(), tt.request)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAWSTagging_Delete(t *testing.T) {
	tests := []struct {
		name          string
		request       *services.DeleteTagsRequest
		mockResponse  *ec2.DeleteTagsOutput
		mockError     error
		expectedError string
	}{
		{
			name: "successful delete tags",
			request: &services.DeleteTagsRequest{
				ResourceIDs: []string{"i-1234567890abcdef0"},
				TagKeys:     []string{"Environment", "Project"},
			},
			mockResponse:  &ec2.DeleteTagsOutput{},
			mockError:     nil,
			expectedError: "",
		},
		{
			name: "empty resource IDs",
			request: &services.DeleteTagsRequest{
				ResourceIDs: []string{},
				TagKeys:     []string{"Environment"},
			},
			mockResponse:  nil,
			mockError:     nil,
			expectedError: "tagging: at least one resource ID must be specified",
		},
		{
			name: "empty tag keys",
			request: &services.DeleteTagsRequest{
				ResourceIDs: []string{"i-1234567890abcdef0"},
				TagKeys:     []string{},
			},
			mockResponse:  nil,
			mockError:     nil,
			expectedError: "tagging: at least one tag key must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockEC2Client{
				deleteTagsResponse: tt.mockResponse,
				deleteTagsError:    tt.mockError,
			}

			tagging := NewTagging(mockClient)
			err := tagging.Delete(context.Background(), tt.request)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAWSTagging_List(t *testing.T) {
	tests := []struct {
		name          string
		filter        *services.DescribeTagsFilter
		mockResponse  *ec2.DescribeTagsOutput
		mockError     error
		expected      []*services.ResourceTag
		expectedError string
	}{
		{
			name: "successful list tags",
			filter: &services.DescribeTagsFilter{
				ResourceIDs: []string{"i-1234567890abcdef0"},
			},
			mockResponse: &ec2.DescribeTagsOutput{
				Tags: []types.TagDescription{
					{
						ResourceId:   aws.String("i-1234567890abcdef0"),
						ResourceType: types.ResourceTypeInstance,
						Key:          aws.String("Environment"),
						Value:        aws.String("test"),
					},
					{
						ResourceId:   aws.String("i-1234567890abcdef0"),
						ResourceType: types.ResourceTypeInstance,
						Key:          aws.String("Project"),
						Value:        aws.String("cloudsdk"),
					},
				},
			},
			mockError: nil,
			expected: []*services.ResourceTag{
				{
					ResourceID:   "i-1234567890abcdef0",
					ResourceType: "instance",
					Key:          "Environment",
					Value:        "test",
				},
				{
					ResourceID:   "i-1234567890abcdef0",
					ResourceType: "instance",
					Key:          "Project",
					Value:        "cloudsdk",
				},
			},
			expectedError: "",
		},
		{
			name: "empty filter",
			filter: &services.DescribeTagsFilter{},
			mockResponse: &ec2.DescribeTagsOutput{
				Tags: []types.TagDescription{},
			},
			mockError:     nil,
			expected:      []*services.ResourceTag{},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockEC2Client{
				describeTagsResponse: tt.mockResponse,
				describeTagsError:    tt.mockError,
			}

			tagging := NewTagging(mockClient)
			result, err := tagging.List(context.Background(), tt.filter)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(result))
				for i, expected := range tt.expected {
					assert.Equal(t, expected.ResourceID, result[i].ResourceID)
					assert.Equal(t, expected.ResourceType, result[i].ResourceType)
					assert.Equal(t, expected.Key, result[i].Key)
					assert.Equal(t, expected.Value, result[i].Value)
				}
			}
		})
	}
}

func TestAWSCompute_Tags(t *testing.T) {
	mockClient := &mockEC2Client{}
	compute := NewWithClient(mockClient)

	tagging := compute.Tags()
	assert.NotNil(t, tagging)
	
	// Verify it's the correct type
	_, ok := tagging.(*AWSTagging)
	assert.True(t, ok, "Tags() should return an *AWSTagging instance")
}

func TestAWSTagging_CreateTagsByARN(t *testing.T) {
	mockClient := &mockEC2Client{
		createTagsResponse: &ec2.CreateTagsOutput{},
		createTagsError:    nil,
	}

	tagging := NewTagging(mockClient)
	
	resourceARNs := []string{"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0"}
	tags := map[string]string{
		"Environment": "test",
		"Project":     "cloudsdk",
	}

	err := tagging.CreateTagsByARN(context.Background(), resourceARNs, tags)
	assert.NoError(t, err)
}

func TestAWSTagging_DeleteTagsByARN(t *testing.T) {
	mockClient := &mockEC2Client{
		deleteTagsResponse: &ec2.DeleteTagsOutput{},
		deleteTagsError:    nil,
	}

	tagging := NewTagging(mockClient)
	
	resourceARNs := []string{"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0"}
	tagKeys := []string{"Environment", "Project"}

	err := tagging.DeleteTagsByARN(context.Background(), resourceARNs, tagKeys)
	assert.NoError(t, err)
}