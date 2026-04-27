package db

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ConnectionManager manages database connections
type ConnectionManager struct {
	connections map[string]*ManagedConnection
	mu          sync.RWMutex
	idCounter   int
}

// ManagedConnection wraps a database connection with metadata
type ManagedConnection struct {
	ID               string
	DBType           DBType
	Conn             DBConnection
	LastAccessed     time.Time
	ConnectionString string
}

var (
	globalManager *ConnectionManager
	once          sync.Once
)

// GetManager returns the global connection manager instance
func GetManager() *ConnectionManager {
	once.Do(func() {
		globalManager = &ConnectionManager{
			connections: make(map[string]*ManagedConnection),
			idCounter:   0,
		}
	})
	return globalManager
}

// Add creates a new connection and adds it to the manager
func (cm *ConnectionManager) Add(config ConnectionConfig) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Create new connection
	conn, err := NewConnection(config.DBType)
	if err != nil {
		return "", err
	}

	// Connect to database
	if err := conn.Connect(config); err != nil {
		return "", err
	}

	// Generate unique ID
	cm.idCounter++
	id := fmt.Sprintf("conn_%d_%d", time.Now().Unix(), cm.idCounter)

	// Create managed connection
	mc := &ManagedConnection{
		ID:           id,
		DBType:       config.DBType,
		Conn:         conn,
		LastAccessed: time.Now(),
		ConnectionString: fmt.Sprintf("%s://%s@%s:%d/%s",
			config.DBType, config.Username, config.Host, config.Port, config.Database),
	}

	// Handle SQLite connection string differently
	if config.DBType == SQLite {
		mc.ConnectionString = fmt.Sprintf("sqlite://%s", config.FilePath)
	}

	cm.connections[id] = mc
	return id, nil
}

// Get retrieves a connection by ID
func (cm *ConnectionManager) Get(id string) (*ManagedConnection, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[id]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", id)
	}

	// Update last accessed time
	conn.LastAccessed = time.Now()
	return conn, nil
}

// Remove closes and removes a connection from the manager
func (cm *ConnectionManager) Remove(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[id]
	if !exists {
		return fmt.Errorf("connection not found: %s", id)
	}

	// Close the connection
	if err := conn.Conn.Close(); err != nil {
		return err
	}

	delete(cm.connections, id)
	return nil
}

// List returns all managed connections
func (cm *ConnectionManager) List() []*ManagedConnection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	list := make([]*ManagedConnection, 0, len(cm.connections))
	for _, conn := range cm.connections {
		list = append(list, conn)
	}
	return list
}

// CleanupIdle removes connections that have been idle for longer than the specified duration
func (cm *ConnectionManager) CleanupIdle(maxIdleTime time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	for id, conn := range cm.connections {
		if now.Sub(conn.LastAccessed) > maxIdleTime {
			conn.Conn.Close()
			delete(cm.connections, id)
			log.Printf("[ConnectionManager] cleaned idle connection: %s (idle %v)", id, now.Sub(conn.LastAccessed).Round(time.Second))
		}
	}
}

// StartCleanupRoutine starts a background routine to clean up idle connections
func (cm *ConnectionManager) StartCleanupRoutine(interval, maxIdleTime time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			cm.CleanupIdle(maxIdleTime)
		}
	}()
}

// Shutdown closes all managed connections and clears the connection map.
// Should be called when the application is shutting down.
func (cm *ConnectionManager) Shutdown() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for id, conn := range cm.connections {
		if err := conn.Conn.Close(); err != nil {
			log.Printf("[ConnectionManager] failed to close connection %s: %v", id, err)
		}
		delete(cm.connections, id)
	}
}
