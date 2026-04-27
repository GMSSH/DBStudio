package rpc

import (
	"fmt"
	"sqlmanager/pkg/config"
	"sqlmanager/pkg/rpcutil"

	"github.com/DemonZack/simplejrpc-go/core/gi18n"
	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// SaveConnectionConfig saves a connection configuration
func (s *Server) SaveConnectionConfig(req *gsock.Request) (any, error) {
	var args struct {
		ID       string `json:"id"`       // Empty for new, existing ID for update
		Name     string `json:"name"`
		DBType   string `json:"dbType"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		FilePath string `json:"filePath"`
		Lang     string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	// Validate required fields
	if args.Name == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidConnectionName"))
	}

	conn := &config.SavedConnection{
		ID:       args.ID,
		Name:     args.Name,
		DBType:   args.DBType,
		Host:     args.Host,
		Port:     args.Port,
		Username: args.Username,
		Password: args.Password,
		Database: args.Database,
		FilePath: args.FilePath,
	}

	store := config.GetConnectionStore()
	if err := store.SaveConnection(conn); err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("SaveConnectionFailed"), err.Error())
	}

	return map[string]interface{}{
		"id":      conn.ID,
		"message": gi18n.Instance().T("SaveConnectionSuccess"),
	}, nil
}

// ListConnectionConfigs returns all saved connection configurations (without passwords)
func (s *Server) ListConnectionConfigs(req *gsock.Request) (any, error) {
	rpcutil.SetLanguage(req)

	store := config.GetConnectionStore()
	connections, err := store.ListConnections()
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ListConnectionsFailed"), err.Error())
	}

	return connections, nil
}

// GetConnectionConfig retrieves a single connection configuration (with password)
func (s *Server) GetConnectionConfig(req *gsock.Request) (any, error) {
	var args struct {
		ID   string `json:"id"`
		Lang string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	if args.ID == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidConnectionID"))
	}

	store := config.GetConnectionStore()
	conn, err := store.GetConnection(args.ID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("GetConnectionFailed"), err.Error())
	}

	return conn, nil
}

// DeleteConnectionConfig deletes a connection configuration
func (s *Server) DeleteConnectionConfig(req *gsock.Request) (any, error) {
	var args struct {
		ID   string `json:"id"`
		Lang string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	if args.ID == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidConnectionID"))
	}

	store := config.GetConnectionStore()
	if err := store.DeleteConnection(args.ID); err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("DeleteConnectionFailed"), err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": gi18n.Instance().T("DeleteConnectionSuccess"),
	}, nil
}
