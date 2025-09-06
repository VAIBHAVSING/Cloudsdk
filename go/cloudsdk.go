package cloudsdk

import (
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

// Config holds the configuration for the Cloud SDK
type Config struct {
	Region string
	// Add other global configs like credentials, etc.
}

// Provider defines the interface for cloud providers
type Provider interface {
	Compute() services.Compute
	Storage() services.Storage
	Database() services.Database
}

// Client is the main entry point for the Cloud SDK
type Client struct {
	provider Provider
}

// New creates a new Cloud SDK client with the specified provider
func New(provider Provider, config *Config) *Client {
	return &Client{provider: provider}
}

// Compute returns the compute service
func (c *Client) Compute() services.Compute {
	return c.provider.Compute()
}

// Storage returns the storage service
func (c *Client) Storage() services.Storage {
	return c.provider.Storage()
}

// Database returns the database service
func (c *Client) Database() services.Database {
	return c.provider.Database()
}
