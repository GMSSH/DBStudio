package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLConnection represents a MySQL database connection
type MySQLConnection struct {
	BaseConnection
}

// Connect establishes a MySQL database connection
func (mc *MySQLConnection) Connect(config ConnectionConfig) error {
	mc.config = config
	mc.dbType = MySQL

	// Build connection string with multi-statement support
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&multiStatements=true",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	// Open connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// Set connection pool settings (optimized for desktop single-user app)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	mc.db = db
	return nil
}

// SelectDatabase switches to the specified database
func (mc *MySQLConnection) SelectDatabase(database string) error {
	if mc.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// Escape backticks to prevent SQL injection
	// Double any backticks in the database name
	escapedDb := strings.ReplaceAll(database, "`", "``")
	
	_, err := mc.db.Exec(fmt.Sprintf("USE `%s`", escapedDb))
	return err
}
