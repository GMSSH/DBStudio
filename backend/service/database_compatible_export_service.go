package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sqlmanager/pkg/config"
	"sqlmanager/pkg/db"
	"sqlmanager/pkg/files"
	"strings"
)

type DatabaseExportMode string

const (
	DatabaseExportModeAuto       DatabaseExportMode = "auto"
	DatabaseExportModeFull       DatabaseExportMode = "full"
	DatabaseExportModeCompatible DatabaseExportMode = "compatible"
)

type DatabaseExportOptions struct {
	Mode              DatabaseExportMode `json:"mode"`
	IncludeRoutines   bool               `json:"includeRoutines"`
	IncludeTriggers   bool               `json:"includeTriggers"`
	IncludeEvents     bool               `json:"includeEvents"`
	IncludeTablespace bool               `json:"includeTablespace"`
}

type DatabaseCompatibleExportService struct{}

func NewDatabaseCompatibleExportService() *DatabaseCompatibleExportService {
	return &DatabaseCompatibleExportService{}
}

func (s *DatabaseCompatibleExportService) ExportDatabase(
	conn db.DBConnection,
	dbName string,
	outputPath string,
	onPhase DumpPhaseCallback,
	onProgress DumpProgressCallback,
) (*DatabaseExportResult, error) {
	if dbName == "" {
		return nil, fmt.Errorf("database is required")
	}

	if onPhase != nil {
		onPhase("inspect", "Loading database objects")
	}

	tableSvc := NewTableService()
	objects, err := tableSvc.ListTables(conn, dbName)
	if err != nil {
		return nil, err
	}

	designerSvc := NewTableDesignerService()
	importExportSvc := NewImportExportService()

	if err := files.EnsureDir(filepath.Dir(outputPath)); err != nil {
		return nil, err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	totalObjects := int64(len(objects))
	if onProgress != nil && totalObjects > 0 {
		onProgress(0, totalObjects)
	}

	fileName := filepath.Base(outputPath)
	for i, object := range objects {
		if onPhase != nil {
			onPhase("export-object", fmt.Sprintf("Exporting %s %s", object.Type, object.Name))
		}

		ddl, err := designerSvc.GetTableDDL(conn, dbName, object.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to export DDL for %s: %w", object.Name, err)
		}
		if _, err := file.WriteString(ddl); err != nil {
			return nil, err
		}
		if !strings.HasSuffix(ddl, ";\n") {
			if _, err := file.WriteString(";\n"); err != nil {
				return nil, err
			}
		}
		if _, err := file.WriteString("\n"); err != nil {
			return nil, err
		}

		if object.Type == "table" {
			exported, err := importExportSvc.ExportTable(conn, dbName, object.Name, ExportSQLInsert, 100000)
			if err != nil {
				return nil, fmt.Errorf("failed to export data for %s: %w", object.Name, err)
			}
			if exported.Data != "" {
				if _, err := file.WriteString(exported.Data); err != nil {
					return nil, err
				}
				if !strings.HasSuffix(exported.Data, "\n") {
					if _, err := file.WriteString("\n"); err != nil {
						return nil, err
					}
				}
			}
		}

		if _, err := file.WriteString("\n"); err != nil {
			return nil, err
		}

		if onProgress != nil && totalObjects > 0 {
			onProgress(int64(i+1), totalObjects)
		}
	}

	size, err := files.GetFileSize(outputPath)
	if err != nil {
		return nil, err
	}

	return &DatabaseExportResult{
		FilePath: outputPath,
		FileName: fileName,
		Size:     size,
		Target:   string(ExportTargetApp),
	}, nil
}

func IsMySQLPermissionError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "access denied") ||
		strings.Contains(msg, "1044") ||
		strings.Contains(msg, "1045") ||
		strings.Contains(msg, "1227") ||
		strings.Contains(msg, "show routine status") ||
		strings.Contains(msg, "show events") ||
		strings.Contains(msg, "show triggers")
}

func normalizeExportOptions(opts DatabaseExportOptions) DatabaseExportOptions {
	if opts.Mode == "" {
		opts.Mode = DatabaseExportModeAuto
	}
	return opts
}

func buildMySQLDumpArgs(savedConn *config.SavedConnection, database string, options DatabaseExportOptions) []string {
	args := []string{
		"--protocol=tcp",
		"--host", savedConn.Host,
		"--port", fmt.Sprintf("%d", savedConn.Port),
		"--user", savedConn.Username,
		"--single-transaction",
		"--default-character-set=utf8mb4",
	}

	if options.IncludeRoutines {
		args = append(args, "--routines")
	}
	if options.IncludeTriggers {
		args = append(args, "--triggers")
	} else {
		args = append(args, "--skip-triggers")
	}
	if options.IncludeEvents {
		args = append(args, "--events")
	}
	if !options.IncludeTablespace {
		args = append(args, "--no-tablespaces")
	}

	args = append(args, "--databases", database)
	return args
}
