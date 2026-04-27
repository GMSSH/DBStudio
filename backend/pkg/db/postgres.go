package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQLConnection represents a PostgreSQL database connection
type PostgreSQLConnection struct {
	BaseConnection
}

// Connect establishes a PostgreSQL database connection
func (pc *PostgreSQLConnection) Connect(config ConnectionConfig) error {
	pc.config = config
	pc.dbType = PostgreSQL

	// Use default 'postgres' database if not specified
	dbName := config.Database
	if dbName == "" {
		dbName = "postgres"
	}

	// Build connection string with SSL disabled for local development
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		dbName,
	)

	// Open connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Set connection pool settings (optimized for desktop single-user app)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	pc.db = db
	return nil
}

// SelectDatabase switches to a different database (requires reconnection for PostgreSQL)
func (pc *PostgreSQLConnection) SelectDatabase(database string) error {
	if pc.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// PostgreSQL: Database is set at connection time via DSN
	// To switch databases, we need to reconnect with a new DSN
	
	// Close existing connection
	if err := pc.db.Close(); err != nil {
		return fmt.Errorf("failed to close existing connection: %w", err)
	}
	
	// Update config with new database
	newConfig := pc.config
	newConfig.Database = database
	
	// Reconnect with new database
	return pc.Connect(newConfig)
}
