package mock

import (
	"context"
	"fmt"
	"time"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

// MockDatabase implements the services.Database interface for testing.
// It provides configurable responses and error injection for all database operations.
type MockDatabase struct {
	provider *MockProvider
}

// CreateDB creates a mock database instance with configurable responses.
// Returns a configured database response if available, otherwise generates a realistic mock database.
//
// Error injection:
//   - Configure errors using WithError("CreateDB", error)
//   - Common test scenarios: authentication, authorization, resource conflicts
//
// Example:
//
//	// Success scenario
//	db, err := mockDatabase.CreateDB(ctx, &services.DBConfig{
//	    Name: "test-db",
//	    Engine: "postgres",
//	    InstanceClass: "db.t3.micro",
//	})
//
//	// Error scenario (configured with WithError)
//	provider := mock.New("us-east-1").
//	    WithError("CreateDB", cloudsdk.NewResourceConflictError(...))
func (m *MockDatabase) CreateDB(ctx context.Context, config *services.DBConfig) (*services.DBInstance, error) {
	m.provider.applyDelay("CreateDB")

	if err := m.provider.checkError("CreateDB"); err != nil {
		m.provider.recordOperation("CreateDB", []interface{}{config}, nil, err)
		return nil, err
	}

	// Check if we have a configured response for this database name
	if db, exists := m.provider.dbResponses[config.Name]; exists {
		// Store in state for later retrieval
		m.provider.dbState[db.ID] = db
		m.provider.recordOperation("CreateDB", []interface{}{config}, db, nil)
		return db, nil
	}

	// Check if database already exists
	for _, existingDB := range m.provider.dbState {
		if existingDB.Name == config.Name {
			err := cloudsdk.NewCloudError(
				cloudsdk.ErrResourceConflict,
				"Database instance already exists",
				"mock", "database", "CreateDB",
			).WithSuggestions(
				"Choose a different database name",
				"Delete the existing database first",
			)
			m.provider.recordOperation("CreateDB", []interface{}{config}, nil, err)
			return nil, err
		}
	}

	// Generate a realistic mock database
	db := &services.DBInstance{
		ID:         generateDBInstanceID(config.Name),
		Name:       config.Name,
		Engine:     fmt.Sprintf("%s-%s", config.Engine, getDefaultEngineVersion(config.Engine)),
		Status:     "available",
		Endpoint:   fmt.Sprintf("%s:%d", generateEndpoint(config.Name, m.provider.region), getDefaultPort(config.Engine)),
		LaunchTime: time.Now().Format(time.RFC3339),
	}

	// Store in state
	m.provider.dbState[db.ID] = db

	m.provider.recordOperation("CreateDB", []interface{}{config}, db, nil)
	return db, nil
}

// GetDB retrieves a mock database instance by ID.
// Returns a database from the mock state if it exists, otherwise returns a not found error.
//
// Error injection:
//   - Configure errors using WithError("GetDB", error)
//   - Automatically returns ErrResourceNotFound for non-existent databases
//
// Example:
//
//	db, err := mockDatabase.GetDB(ctx, "myapp-prod-db")
//	if err != nil {
//	    // Handle not found or configured error
//	}
func (m *MockDatabase) GetDB(ctx context.Context, id string) (*services.DBInstance, error) {
	m.provider.applyDelay("GetDB")

	if err := m.provider.checkError("GetDB"); err != nil {
		m.provider.recordOperation("GetDB", []interface{}{id}, nil, err)
		return nil, err
	}

	// Check if database exists in state
	if db, exists := m.provider.dbState[id]; exists {
		m.provider.recordOperation("GetDB", []interface{}{id}, db, nil)
		return db, nil
	}

	// Database not found
	err := cloudsdk.NewResourceNotFoundError("mock", "database", "database instance", id)
	m.provider.recordOperation("GetDB", []interface{}{id}, nil, err)
	return nil, err
}

// ListDBs returns all mock database instances in the current state.
// Returns an empty slice if no databases exist.
//
// Error injection:
//   - Configure errors using WithError("ListDBs", error)
//
// Example:
//
//	dbs, err := mockDatabase.ListDBs(ctx)
//	for _, db := range dbs {
//	    fmt.Printf("Database: %s (%s)\n", db.Name, db.Status)
//	}
func (m *MockDatabase) ListDBs(ctx context.Context) ([]*services.DBInstance, error) {
	m.provider.applyDelay("ListDBs")

	if err := m.provider.checkError("ListDBs"); err != nil {
		m.provider.recordOperation("ListDBs", []interface{}{}, nil, err)
		return nil, err
	}

	// Collect all databases from state
	dbs := make([]*services.DBInstance, 0, len(m.provider.dbState))
	for _, db := range m.provider.dbState {
		dbs = append(dbs, db)
	}

	m.provider.recordOperation("ListDBs", []interface{}{}, dbs, nil)
	return dbs, nil
}

// DeleteDB removes a mock database instance from the state.
// Returns an error if the database doesn't exist.
//
// Error injection:
//   - Configure errors using WithError("DeleteDB", error)
//   - Automatically returns ErrResourceNotFound for non-existent databases
//
// Example:
//
//	err := mockDatabase.DeleteDB(ctx, "myapp-prod-db")
//	if err != nil {
//	    // Handle not found or configured error
//	}
func (m *MockDatabase) DeleteDB(ctx context.Context, id string) error {
	m.provider.applyDelay("DeleteDB")

	if err := m.provider.checkError("DeleteDB"); err != nil {
		m.provider.recordOperation("DeleteDB", []interface{}{id}, nil, err)
		return err
	}

	// Check if database exists
	_, exists := m.provider.dbState[id]
	if !exists {
		err := cloudsdk.NewResourceNotFoundError("mock", "database", "database instance", id)
		m.provider.recordOperation("DeleteDB", []interface{}{id}, nil, err)
		return err
	}

	// Remove from state
	delete(m.provider.dbState, id)

	m.provider.recordOperation("DeleteDB", []interface{}{id}, nil, nil)
	return nil
}

// Helper functions for generating realistic mock data

// getDefaultEngineVersion returns a default version for the specified engine
func getDefaultEngineVersion(engine string) string {
	switch engine {
	case "postgres":
		return "14.9"
	case "mysql":
		return "8.0.35"
	case "mariadb":
		return "10.6.15"
	case "oracle-ee", "oracle-se2":
		return "19.0.0.0.ru-2023-10.rur-2023-10.r1"
	case "sqlserver-ex", "sqlserver-web", "sqlserver-se", "sqlserver-ee":
		return "15.00.4335.1.v1"
	default:
		return "1.0.0"
	}
}

// getDefaultPort returns the default port for the specified engine
func getDefaultPort(engine string) int32 {
	switch engine {
	case "postgres":
		return 5432
	case "mysql", "mariadb":
		return 3306
	case "oracle-ee", "oracle-se2":
		return 1521
	case "sqlserver-ex", "sqlserver-web", "sqlserver-se", "sqlserver-ee":
		return 1433
	default:
		return 3306
	}
}
