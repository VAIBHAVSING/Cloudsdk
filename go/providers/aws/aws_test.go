package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAWSProvider(t *testing.T) {
	ctx := context.Background()

	provider, err := NewAWSProvider(ctx, "us-east-1")
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.Compute())
	assert.NotNil(t, provider.Storage())
	assert.NotNil(t, provider.Database())
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	provider, err := Create(ctx, "us-west-2")
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.Compute())
	assert.NotNil(t, provider.Storage())
	assert.NotNil(t, provider.Database())
}
