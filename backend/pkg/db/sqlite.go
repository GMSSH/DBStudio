package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// SQLiteConnection represents a SQLite database connection
type SQLiteConnection struct {
	BaseConnection
}

// Connect establishes a SQLite database connection
func (sc *SQLiteConnection) Connect(config ConnectionConfig) error {
	sc.config = config
	sc.dbType = SQLite

	// Ensure the directory exists for the SQLite file
	if config.FilePath != "" {
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for SQLite database: %w", err)
		}
	} else {
		return fmt.Errorf("SQLite file path is required")
	}

	// Build connection string with optimizations
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL", config.FilePath)

	// Open connection
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open SQLite connection: %w", err)
	}

	// Set connection pool settings (SQLite handles concurrency differently)
	db.SetMaxOpenConns(1) // SQLite works best with a single connection
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // No limit

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	sc.db = db
	return nil
}

// SelectDatabase for SQLite is a no-op since SQLite uses a single database file
func (sc *SQLiteConnection) SelectDatabase(database string) error {
	// SQLite doesn't support multiple databases in the same connection
	// The database is the file itself
	return nil
}
