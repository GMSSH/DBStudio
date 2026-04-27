package service

import (
	"context"
	"fmt"
	"path/filepath"
	"sqlmanager/pkg/db"
	"strings"
	"time"
)

// DatabaseInfo represents database metadata
type DatabaseInfo struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	TableCount int    `json:"tableCount"`
	Charset    string `json:"charset,omitempty"`
	IsSystem   bool   `json:"isSystem"` // Mark system databases
}

// DatabaseService handles database operations
type DatabaseService struct{}

// NewDatabaseService creates a new database service
func NewDatabaseService() *DatabaseService {
	return &DatabaseService{}
}

// ListDatabases returns a list of all databases
func (ds *DatabaseService) ListDatabases(conn db.DBConnection) ([]DatabaseInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var query string
	dbType := conn.GetDBType()

	switch dbType {
	case db.MySQL:
		query = "SHOW DATABASES"
	case db.PostgreSQL:
		query = "SELECT datname FROM pg_database WHERE datistemplate = false"
	case db.SQLite:
		// SQLite doesn't have multiple databases in the same connection
		// Return the database file as a single database entry
		config := conn.GetConfig()
		dbName := config.Database
		if dbName == "" && config.FilePath != "" {
			// Extract filename from file path as database name
			dbName = filepath.Base(config.FilePath)
		}
		if dbName == "" {
			dbName = "main" // SQLite default database name
		}
		return []DatabaseInfo{{Name: dbName, TableCount: 0, IsSystem: false}}, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	defer rows.Close()

	var databases []DatabaseInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		// Mark system databases but don't filter them
		info := DatabaseInfo{
			Name:     name,
			IsSystem: ds.isSystemDatabase(dbType, name),
		}
		databases = append(databases, info)
	}

	return databases, rows.Err()
}

// ListSchemas lists all schemas in a PostgreSQL database
func (ds *DatabaseService) ListSchemas(conn db.DBConnection, dbName string) ([]string, error) {
	ctx := context.Background()
	
	// Only PostgreSQL has schemas
	if conn.GetDBType() != db.PostgreSQL {
		return []string{}, nil
	}
	
	// PostgreSQL: list all non-system schemas
	query := `SELECT schema_name FROM information_schema.schemata 
	          WHERE schema_name NOT IN ('pg_catalog', 'information_schema') 
	          AND schema_name NOT LIKE 'pg_%' 
	          ORDER BY schema_name`
	
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	defer rows.Close()
	
	var schemas []string
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, fmt.Errorf("failed to scan schema name: %w", err)
		}
		schemas = append(schemas, schemaName)
	}
	
	return schemas, rows.Err()
}

// GetDatabaseInfo returns detailed information about a database
func (ds *DatabaseService) GetDatabaseInfo(conn db.DBConnection, dbName string) (*DatabaseInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info := &DatabaseInfo{Name: dbName}
	dbType := conn.GetDBType()

	switch dbType {
	case db.MySQL:
		// Get table count
		row := conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ?", dbName)
		row.Scan(&info.TableCount)

		// Get charset
		row = conn.QueryRow(ctx, "SELECT DEFAULT_CHARACTER_SET_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", dbName)
		row.Scan(&info.Charset)

	case db.PostgreSQL:
		// Get table count
		row := conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_catalog = $1 AND table_schema = 'public'", dbName)
		row.Scan(&info.TableCount)

		// Get encoding
		row = conn.QueryRow(ctx, "SELECT pg_encoding_to_char(encoding) FROM pg_database WHERE datname = $1", dbName)
		row.Scan(&info.Charset)

	case db.SQLite:
		// Get table count
		row := conn.QueryRow(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
		row.Scan(&info.TableCount)
	}

	return info, nil
}

// CreateDatabase creates a new database
func (ds *DatabaseService) CreateDatabase(conn db.DBConnection, dbName string) error {
	if !ds.isValidDatabaseName(dbName) {
		return fmt.Errorf("invalid database name: %s", dbName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var query string
	dbType := conn.GetDBType()

	switch dbType {
	case db.MySQL:
		// Escape backticks to prevent SQL injection
		escapedName := strings.ReplaceAll(dbName, "`", "``")
		query = fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", escapedName)
	case db.PostgreSQL:
		// Escape double quotes to prevent SQL injection
		escapedName := strings.ReplaceAll(dbName, `"`, `""`)
		query = fmt.Sprintf("CREATE DATABASE \"%s\"", escapedName)
	case db.SQLite:
		return fmt.Errorf("SQLite does not support creating databases via SQL")
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	_, err := conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}

// DropDatabase deletes a database
func (ds *DatabaseService) DropDatabase(conn db.DBConnection, dbName string) error {
	if !ds.isValidDatabaseName(dbName) {
		return fmt.Errorf("invalid database name: %s", dbName)
	}

	// Prevent dropping system databases
	if ds.isSystemDatabase(conn.GetDBType(), dbName) {
		return fmt.Errorf("cannot drop system database: %s", dbName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var query string
	dbType := conn.GetDBType()

	switch dbType {
	case db.MySQL:
		// Escape backticks to prevent SQL injection
		escapedName := strings.ReplaceAll(dbName, "`", "``")
		query = fmt.Sprintf("DROP DATABASE `%s`", escapedName)
	case db.PostgreSQL:
		// Escape double quotes to prevent SQL injection
		escapedName := strings.ReplaceAll(dbName, `"`, `""`)
		query = fmt.Sprintf("DROP DATABASE \"%s\"", escapedName)
	case db.SQLite:
		return fmt.Errorf("SQLite does not support dropping databases via SQL")
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	_, err := conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	return nil
}

// isValidDatabaseName validates a database name
func (ds *DatabaseService) isValidDatabaseName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	// Add more validation as needed
	return true
}

// isSystemDatabase checks if a database is a system database
func (ds *DatabaseService) isSystemDatabase(dbType db.DBType, name string) bool {
	switch dbType {
	case db.MySQL:
		systemDbs := map[string]bool{
			"information_schema": true,
			"mysql":              true,
			"performance_schema": true,
			"sys":                true,
		}
		return systemDbs[name]
	case db.PostgreSQL:
		systemDbs := map[string]bool{
			"postgres":  true,
			"template0": true,
			"template1": true,
		}
		return systemDbs[name]
	}
	return false
}

// SwitchDatabase switches to a different database (creates new connection)
func (ds *DatabaseService) SwitchDatabase(currentConn db.DBConnection, dbName string) (db.DBConnection, error) {
	config := currentConn.GetConfig()
	config.Database = dbName

	// Create new connection with the new database
	newConn, err := db.NewConnection(config.DBType)
	if err != nil {
		return nil, err
	}

	if err := newConn.Connect(config); err != nil {
		return nil, fmt.Errorf("failed to switch to database %s: %w", dbName, err)
	}

	return newConn, nil
}

// TestConnection tests if a connection configuration is valid
func (ds *DatabaseService) TestConnection(config db.ConnectionConfig) error {
	conn, err := db.NewConnection(config.DBType)
	if err != nil {
		return err
	}

	if err := conn.Connect(config); err != nil {
		return err
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return err
	}

	conn.Close()
	return nil
}
