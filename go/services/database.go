package services

import "context"

// DBConfig represents the configuration for creating a database instance
type DBConfig struct {
	Name               string
	Engine             string
	EngineVersion      string
	InstanceClass      string
	AllocatedStorage   int32
	MasterUsername     string
	MasterPassword     string
	DBName             string
	VpcSecurityGroups  []string
}

// DBInstance represents a database instance
type DBInstance struct {
	ID            string
	Name          string
	Engine        string
	Status        string
	Endpoint      string
	LaunchTime    string
}

// Database defines the interface for database operations
type Database interface {
	CreateDB(ctx context.Context, config *DBConfig) (*DBInstance, error)
	ListDBs(ctx context.Context) ([]*DBInstance, error)
	GetDB(ctx context.Context, id string) (*DBInstance, error)
	DeleteDB(ctx context.Context, id string) error
}
