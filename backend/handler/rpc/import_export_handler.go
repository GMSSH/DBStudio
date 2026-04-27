package rpc

import (
	"fmt"
	"sqlmanager/pkg/db"
	"sqlmanager/pkg/rpcutil"
	"sqlmanager/service"
	"strings"

	"github.com/DemonZack/simplejrpc-go/core/gi18n"
	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// Import/Export Handlers

// ExportTable exports table data in the specified format
func (s *Server) ExportTable(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		Table    string `json:"table"`
		Format   string `json:"format"` // "csv" | "json" | "sql"
		Limit    int    `json:"limit"`
		Lang     string `json:"lang"`
	}
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	if args.Table == "" {
		return nil, fmt.Errorf("table name is required")
	}
	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}
	svc := service.NewImportExportService()
	result, err := svc.ExportTable(managedConn.Conn, args.Database, args.Table, service.ExportFormat(args.Format), args.Limit)
	if err != nil {
		return nil, fmt.Errorf("export failed: %v", err)
	}
	return result, nil
}

// ExportTableDDL exports the CREATE TABLE DDL
func (s *Server) ExportTableDDL(req *gsock.Request) (any, error) {
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
	svc := service.NewImportExportService()
	result, err := svc.ExportTableDDL(managedConn.Conn, args.Database, args.Table)
	if err != nil {
		return nil, fmt.Errorf("DDL export failed: %v", err)
	}
	return result, nil
}

// ImportCSV imports CSV data into a table
func (s *Server) ImportCSV(req *gsock.Request) (any, error) {
	var args struct {
		ConnID    string            `json:"connId"`
		Database  string            `json:"database"`
		Table     string            `json:"table"`
		CSVData   string            `json:"csvData"`
		Mapping   map[string]string `json:"mapping"`
		HeaderRow bool              `json:"headerRow"`
		Lang      string            `json:"lang"`
	}
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	if args.Table == "" {
		return nil, fmt.Errorf("table name is required")
	}
	if args.CSVData == "" {
		return nil, fmt.Errorf("CSV data is required")
	}
	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}
	svc := service.NewImportExportService()
	result, err := svc.ImportCSV(managedConn.Conn, args.Database, args.Table, args.CSVData, args.Mapping, args.HeaderRow)
	if err != nil {
		return nil, fmt.Errorf("import failed: %v", err)
	}
	return result, nil
}

// ImportTableSQL imports SQL INSERT statements into the current table only.
func (s *Server) ImportTableSQL(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		Table    string `json:"table"`
		SQLData  string `json:"sqlData"`
		Lang     string `json:"lang"`
	}
	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}
	if args.Table == "" {
		return nil, fmt.Errorf("table name is required")
	}
	if strings.TrimSpace(args.SQLData) == "" {
		return nil, fmt.Errorf("SQL data is required")
	}
	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}
	svc := service.NewImportExportService()
	result, err := svc.ImportTableSQL(managedConn.Conn, args.Database, args.Table, args.SQLData)
	if err != nil {
		return nil, fmt.Errorf("import failed: %v", err)
	}
	return result, nil
}
