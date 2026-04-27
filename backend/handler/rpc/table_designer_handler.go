package rpc

import (
	"fmt"
	"sqlmanager/pkg/db"
	"sqlmanager/pkg/rpcutil"
	"sqlmanager/service"

	"github.com/DemonZack/simplejrpc-go/core/gi18n"
	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// Table Designer Handlers

// CreateTable creates a new table
func (s *Server) CreateTable(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string           `json:"connId"`
		Database string           `json:"database"`
		TableDef service.TableDef `json:"tableDef"`
		Lang     string           `json:"lang"`
	}
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}
	svc := service.NewTableDesignerService()
	if err := svc.CreateTable(managedConn.Conn, args.Database, args.TableDef); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}
	return map[string]interface{}{"success": true}, nil
}

// AlterTable alters a table structure
func (s *Server) AlterTable(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string            `json:"connId"`
		Database string            `json:"database"`
		Table    string            `json:"table"`
		Ops      []service.AlterOp `json:"ops"`
		Lang     string            `json:"lang"`
	}
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}
	svc := service.NewTableDesignerService()
	if err := svc.AlterTable(managedConn.Conn, args.Database, args.Table, args.Ops); err != nil {
		return nil, fmt.Errorf("failed to alter table: %v", err)
	}
	return map[string]interface{}{"success": true}, nil
}

// DropTable drops a table
func (s *Server) DropTable(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		Table    string `json:"table"`
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
	svc := service.NewTableDesignerService()
	if err := svc.DropTable(managedConn.Conn, args.Database, args.Table); err != nil {
		return nil, fmt.Errorf("failed to drop table: %v", err)
	}
	return map[string]interface{}{"success": true}, nil
}

// RenameTable renames a table
func (s *Server) RenameTable(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		OldName  string `json:"oldName"`
		NewName  string `json:"newName"`
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
	svc := service.NewTableDesignerService()
	if err := svc.RenameTable(managedConn.Conn, args.Database, args.OldName, args.NewName); err != nil {
		return nil, fmt.Errorf("failed to rename table: %v", err)
	}
	return map[string]interface{}{"success": true}, nil
}

// GetTableDDL returns the CREATE TABLE DDL
func (s *Server) GetTableDDL(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		Table    string `json:"table"`
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
	svc := service.NewTableDesignerService()
	ddl, err := svc.GetTableDDL(managedConn.Conn, args.Database, args.Table)
	if err != nil {
		return nil, fmt.Errorf("failed to get DDL: %v", err)
	}
	return ddl, nil
}
