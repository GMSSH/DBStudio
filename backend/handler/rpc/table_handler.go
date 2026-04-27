package rpc

import (
	"fmt"
	"sqlmanager/pkg/db"
	"sqlmanager/pkg/rpcutil"
	"sqlmanager/service"

	"github.com/DemonZack/simplejrpc-go/core/gi18n"
	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

// Table operation handlers

// ListTables lists all tables in a database
func (s *Server) ListTables(req *gsock.Request) (any, error) {
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

	tableService := service.NewTableService()
	tables, err := tableService.ListTables(managedConn.Conn, args.Database)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ListTablesError"), err.Error())
	}

	return tables, nil
}

// GetTableSchema gets the schema of a table
func (s *Server) GetTableSchema(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		Table    string `json:"table"`
		Lang     string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	if args.Table == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidTableName"))
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	tableService := service.NewTableService()
	schema, err := tableService.GetTableSchema(managedConn.Conn, args.Database, args.Table)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("GetTableSchemaError"), err.Error())
	}

	return schema, nil
}

// GetTableData gets paginated data from a table with optional sort and filter
func (s *Server) GetTableData(req *gsock.Request) (any, error) {
	var args struct {
		ConnID      string `json:"connId"`
		Database    string `json:"database"`
		Table       string `json:"table"`
		Page        int    `json:"page"`
		PageSize    int    `json:"pageSize"`
		SortCol     string `json:"sortCol"`
		SortDir     string `json:"sortDir"`     // "asc" | "desc"
		Lang        string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	if args.Table == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidTableName"))
	}

	// Enforce max 500 rows
	if args.PageSize > 500 {
		return nil, fmt.Errorf(gi18n.Instance().T("MaxRowsExceeded"))
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	tableService := service.NewTableService()
	data, err := tableService.GetTableData(
		managedConn.Conn, args.Database, args.Table,
		args.Page, args.PageSize,
		args.SortCol, args.SortDir,
	)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("GetTableDataError"), err.Error())
	}

	return data, nil
}

// ExecuteSQL executes a SQL query
func (s *Server) ExecuteSQL(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		SQL      string `json:"sql"`
		Lang     string `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	if args.SQL == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidSQL"))
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	// Select database before executing SQL
	if args.Database != "" {
		if err := managedConn.Conn.SelectDatabase(args.Database); err != nil {
			return nil, fmt.Errorf("Failed to select database: %v", err)
		}
	}

	tableService := service.NewTableService()
	result, err := tableService.ExecuteSQL(managedConn.Conn, args.SQL)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ExecuteSQLError"), err.Error())
	}

	return result, nil
}

// UpdateRow updates a row in a table
func (s *Server) UpdateRow(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string                 `json:"connId"`
		Database string                 `json:"database"`
		Table    string                 `json:"table"`
		PKValues map[string]interface{} `json:"pkValues"`
		Updates  map[string]interface{} `json:"updates"`
		Lang     string                 `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	if args.Table == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidTableName"))
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	tableService := service.NewTableService()
	if err := tableService.UpdateRow(managedConn.Conn, args.Database, args.Table, args.PKValues, args.Updates); err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("UpdateRowError"), err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": gi18n.Instance().T("UpdateRowSuccess"),
	}, nil
}

// InsertRow inserts a new row into a table
func (s *Server) InsertRow(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string                 `json:"connId"`
		Database string                 `json:"database"`
		Table    string                 `json:"table"`
		Values   map[string]interface{} `json:"values"`
		Lang     string                 `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	if args.Table == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidTableName"))
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	tableService := service.NewTableService()
	if err := tableService.InsertRow(managedConn.Conn, args.Database, args.Table, args.Values); err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("InsertRowError"), err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": gi18n.Instance().T("InsertRowSuccess"),
	}, nil
}

// DeleteRow deletes a row from a table
func (s *Server) DeleteRow(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string                 `json:"connId"`
		Database string                 `json:"database"`
		Table    string                 `json:"table"`
		PKValues map[string]interface{} `json:"pkValues"`
		Lang     string                 `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	if args.Table == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidTableName"))
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	tableService := service.NewTableService()
	if err := tableService.DeleteRow(managedConn.Conn, args.Database, args.Table, args.PKValues); err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("DeleteRowError"), err.Error())
	}

	return map[string]interface{}{
		"success": true,
		"message": gi18n.Instance().T("DeleteRowSuccess"),
	}, nil
}

// BatchModify executes multiple UPDATE/INSERT/DELETE operations atomically within a transaction
func (s *Server) BatchModify(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string           `json:"connId"`
		Database string           `json:"database"`
		Table    string           `json:"table"`
		Ops      service.BatchOps `json:"ops"`
		Lang     string           `json:"lang"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	if args.Table == "" {
		return nil, fmt.Errorf(gi18n.Instance().T("InvalidTableName"))
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(args.ConnID)
	if err != nil {
		return nil, fmt.Errorf(gi18n.Instance().T("ConnectionNotFound"))
	}

	tableService := service.NewTableService()
	result, err := tableService.BatchModify(managedConn.Conn, args.Database, args.Table, args.Ops)
	if err != nil {
		return nil, fmt.Errorf("batch modify failed: %w", err)
	}

	return result, nil
}
