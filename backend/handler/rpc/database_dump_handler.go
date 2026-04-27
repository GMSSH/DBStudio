package rpc

import (
	"fmt"
	"sqlmanager/pkg/rpcutil"
	"sqlmanager/service"

	"github.com/DemonZack/simplejrpc-go/net/gsock"
)

func (s *Server) ExportDatabase(req *gsock.Request) (any, error) {
	var args struct {
		ConnID     string `json:"connId"`
		Database   string `json:"database"`
		Target     string `json:"target"`
		TargetPath string `json:"targetPath"`
		Options    struct {
			Mode              string `json:"mode"`
			IncludeRoutines   bool   `json:"includeRoutines"`
			IncludeTriggers   bool   `json:"includeTriggers"`
			IncludeEvents     bool   `json:"includeEvents"`
			IncludeTablespace bool   `json:"includeTablespace"`
		} `json:"options"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	taskSvc := service.NewDatabaseDumpTaskService()
	result, err := taskSvc.StartExportTask(args.ConnID, args.Database, service.DatabaseExportTarget(args.Target), args.TargetPath, service.DatabaseExportOptions{
		Mode:              service.DatabaseExportMode(args.Options.Mode),
		IncludeRoutines:   args.Options.IncludeRoutines,
		IncludeTriggers:   args.Options.IncludeTriggers,
		IncludeEvents:     args.Options.IncludeEvents,
		IncludeTablespace: args.Options.IncludeTablespace,
	})
	if err != nil {
		return nil, fmt.Errorf("database export failed: %v", err)
	}
	return result, nil
}

func (s *Server) ImportDatabase(req *gsock.Request) (any, error) {
	var args struct {
		ConnID   string `json:"connId"`
		Database string `json:"database"`
		FilePath string `json:"filePath"`
		Options  struct {
			CreateDatabase bool   `json:"createDatabase"`
			Strategy       string `json:"strategy"`
		} `json:"options"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	taskSvc := service.NewDatabaseDumpTaskService()
	result, err := taskSvc.StartImportTask(args.ConnID, args.Database, args.FilePath, service.DatabaseImportOptions{
		CreateDatabase: args.Options.CreateDatabase,
		Strategy:       service.DatabaseImportStrategy(args.Options.Strategy),
	})
	if err != nil {
		return nil, fmt.Errorf("database import failed: %v", err)
	}
	return result, nil
}

func (s *Server) GetDatabaseDumpTask(req *gsock.Request) (any, error) {
	var args struct {
		TaskID string `json:"taskId"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	taskSvc := service.NewDatabaseDumpTaskService()
	task, err := taskSvc.GetTask(args.TaskID)
	if err != nil {
		return nil, fmt.Errorf("get database dump task failed: %v", err)
	}
	return task, nil
}

func (s *Server) ListDatabaseDumpTasks(req *gsock.Request) (any, error) {
	var args struct{}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	taskSvc := service.NewDatabaseDumpTaskService()
	return taskSvc.ListTasks(), nil
}

func (s *Server) GetDatabaseDumpTaskLogs(req *gsock.Request) (any, error) {
	var args struct {
		TaskID string `json:"taskId"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	taskSvc := service.NewDatabaseDumpTaskService()
	logs, err := taskSvc.GetTaskLogs(args.TaskID)
	if err != nil {
		return nil, fmt.Errorf("get database dump task logs failed: %v", err)
	}
	return logs, nil
}

func (s *Server) CancelDatabaseDumpTask(req *gsock.Request) (any, error) {
	var args struct {
		TaskID string `json:"taskId"`
	}

	if err := rpcutil.ParseParams(req, &args); err != nil {
		return nil, err
	}

	taskSvc := service.NewDatabaseDumpTaskService()
	task, err := taskSvc.CancelTask(args.TaskID)
	if err != nil {
		return nil, fmt.Errorf("cancel database dump task failed: %v", err)
	}
	return task, nil
}
