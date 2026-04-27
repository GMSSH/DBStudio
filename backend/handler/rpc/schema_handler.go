package rpc

import (
	"fmt"
	"sqlmanager/pkg/db"
	"sqlmanager/pkg/rpcutil"
	"sqlmanager/service"

	"github.com/DemonZack/simplejrpc-go/core/gi18n"
	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// Schema operation handlers for PostgreSQL

// ListSchemas lists all schemas in a PostgreSQL database
func (s *Server) ListSchemas(req *gsock.Request) (any, error) {
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

	// Only PostgreSQL supports schemas
	if managedConn.Conn.GetDBType() != db.PostgreSQL {
		return []string{}, nil
	}

	dbService := service.NewDatabaseService()
	schemas, err := dbService.ListSchemas(managedConn.Conn, args.Database)
	if err != nil {
		return nil, fmt.Errorf("Failed to list schemas: %v", err)
	}

	return schemas, nil
}
