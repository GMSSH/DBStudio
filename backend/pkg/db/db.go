package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DBType represents the type of database
type DBType string

const (
	MySQL      DBType = "mysql"
	PostgreSQL DBType = "postgres"
	SQLite     DBType = "sqlite"
)

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	DBType   DBType `json:"dbType"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	FilePath string `json:"filePath"` // For SQLite
}

// DBConnection represents a database connection interface
type DBConnection interface {
	// Connect establishes a database connection
	Connect(config ConnectionConfig) error

	// Close closes the database connection
	Close() error

	// Ping checks if the connection is alive
	Ping() error

	// Query executes a query that returns rows
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// QueryRow executes a query that is expected to return at most one row
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Exec executes a query that doesn't return rows
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// GetDB returns the underlying *sql.DB
	GetDB() *sql.DB

	// GetDBType returns the database type
	GetDBType() DBType

	// GetConfig returns the connection configuration
	GetConfig() ConnectionConfig

	// SelectDatabase selects/switches to a specific database
	SelectDatabase(database string) error
}

// BaseConnection provides common functionality for all database connections
type BaseConnection struct {
	db     *sql.DB
	dbType DBType
	config ConnectionConfig
}

// GetDB returns the underlying *sql.DB
func (bc *BaseConnection) GetDB() *sql.DB {
	return bc.db
}

// GetDBType returns the database type
func (bc *BaseConnection) GetDBType() DBType {
	return bc.dbType
}

// GetConfig returns the connection configuration
func (bc *BaseConnection) GetConfig() ConnectionConfig {
	return bc.config
}

// Close closes the database connection
func (bc *BaseConnection) Close() error {
	if bc.db != nil {
		return bc.db.Close()
	}
	return nil
}

// Ping checks if the connection is alive
func (bc *BaseConnection) Ping() error {
	if bc.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return bc.db.PingContext(ctx)
}

// Query executes a query that returns rows
func (bc *BaseConnection) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if bc.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	return bc.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (bc *BaseConnection) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if bc.db == nil {
		return nil
	}
	return bc.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (bc *BaseConnection) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if bc.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	return bc.db.ExecContext(ctx, query, args...)
}

// SetConnectionPool sets connection pool settings
func (bc *BaseConnection) SetConnectionPool(maxOpen, maxIdle int, maxLifetime time.Duration) {
	if bc.db != nil {
		bc.db.SetMaxOpenConns(maxOpen)
		bc.db.SetMaxIdleConns(maxIdle)
		bc.db.SetConnMaxLifetime(maxLifetime)
	}
}

// NewConnection creates a new database connection based on the database type
func NewConnection(dbType DBType) (DBConnection, error) {
	switch dbType {
	case MySQL:
		return &MySQLConnection{}, nil
	case PostgreSQL:
		return &PostgreSQLConnection{}, nil
	case SQLite:
		return &SQLiteConnection{}, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
