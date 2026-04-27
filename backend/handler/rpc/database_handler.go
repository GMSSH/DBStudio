package rpc

import (
	"fmt"
	"sqlmanager/pkg/db"
	"sqlmanager/pkg/rpcutil"
	"sqlmanager/service"

	"github.com/DemonZack/simplejrpc-go/core/gi18n"
	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// Database connection handlers

// TestConnection tests a database connection
func (s *Server) TestConnection(req *gsock.Request) (any, error) {
	var args struct {
		DBType   string `json:"dbType"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		FilePath string `json:"filePath"` // For SQLite
		Lang     string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	config := db.ConnectionConfig{
		DBType:   db.DBType(args.DBType),
		Host:     args.Host,
		Port:     args.Port,
		Username: args.Username,
		Password: args.Password,
		Database: args.Database,
		FilePath: args.FilePath,
	}

	dbService := service.NewDatabaseService()
	if err := dbService.TestConnection(config); err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionFailed"), err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": gi18n.Instance().T("ConnectionSuccess"),
	}, nil
}

// Connect establishes a database connection and returns connection ID
func (s *Server) Connect(req *gsock.Request) (any, error) {
	var args struct {
		DBType   string `json:"dbType"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		FilePath string `json:"filePath"` // For SQLite
		Lang     string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	config := db.ConnectionConfig{
		DBType:   db.DBType(args.DBType),
		Host:     args.Host,
		Port:     args.Port,
		Username: args.Username,
		Password: args.Password,
		Database: args.Database,
		FilePath: args.FilePath,
	}

	manager := db.GetManager()
	connID, err := manager.Add(config)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionFailed"), err.Error())
	}

	return map[string]interface{}{
		"connId":  connID,
		"message": gi18n.Instance().T("ConnectionSuccess"),
	}, nil
}

// Disconnect closes a database connection
func (s *Server) Disconnect(req *gsock.Request) (any, error) {
	var args struct {
		ConnID string `json:"connId"`
		Lang   string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	manager := db.GetManager()
	if err := manager.Remove(args.ConnID); err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("DisconnectFailed"), err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": gi18n.Instance().T("DisconnectSuccess"),
	}, nil
}

// ListConnections lists all active connections
func (s *Server) ListConnections(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)

	manager := db.GetManager()
	connections := manager.List()

	result := make([]map[string]interface{}, 0, len(connections))
	for _, conn := range connections {
		result = append(result, map[string]interface{}{
			"id":               conn.ID,
			"dbType":           conn.DBType,
			"connectionString": conn.ConnectionString,
			"lastAccessed":     conn.LastAccessed,
		})
	}

	return result, nil
}

// Database operation handlers

// ListDatabases lists all databases
func (s *Server) ListDatabases(req *gsock.Request) (any, error) {
	var args struct {
		ConnID string `json:"connId"`
		Lang   string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	dbService := service.NewDatabaseService()
	databases, err := dbService.ListDatabases(managedConn.Conn)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ListDatabasesError"), err.Error())
	}

	return databases, nil
}

// GetDatabaseInfo gets information about a specific database
func (s *Server) GetDatabaseInfo(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		Lang     string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	dbService := service.NewDatabaseService()
	info, err := dbService.GetDatabaseInfo(managedConn.Conn, args.Database)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("GetDatabaseInfoError"), err.Error())
	}

	return info, nil
}

// CreateDatabase creates a new database
func (s *Server) CreateDatabase(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		Lang     string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	if args.Database == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidDatabaseName"))
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	dbService := service.NewDatabaseService()
	if err := dbService.CreateDatabase(managedConn.Conn, args.Database); err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("CreateDatabaseError"), err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": gi18n.Instance().T("CreateDatabaseSuccess"),
	}, nil
}

// DropDatabase deletes a database
func (s *Server) DropDatabase(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		Lang     string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	if args.Database == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidDatabaseName"))
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	dbService := service.NewDatabaseService()
	if err := dbService.DropDatabase(managedConn.Conn, args.Database); err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("DropDatabaseError"), err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": gi18n.Instance().T("DropDatabaseSuccess"),
	}, nil
}

// SwitchDatabase switches to a different database
func (s *Server) SwitchDatabase(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		Lang     string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	dbService := service.NewDatabaseService()
	newConn, err := dbService.SwitchDatabase(managedConn.Conn, args.Database)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("SwitchDatabaseError"), err.Error())
	}

	// Close old connection and replace with new one
	managedConn.Conn.Close()
	managedConn.Conn = newConn
	config := newConn.GetConfig()
	managedConn.ConnectionString = fmt.Sprintf("%s://%s@%s:%d/%s",
		config.DBType, config.Username, config.Host, config.Port, config.Database)

	return map[string]interface{}{
		"success": true,
		"message": gi18n.Instance().T("SwitchDatabaseSuccess"),
	}, nil
}
