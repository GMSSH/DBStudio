package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sqlmanager/pkg/config"
	"sqlmanager/pkg/db"
	"sqlmanager/pkg/files"
	"strings"
	"time"
)

type DatabaseExportTarget string

const (
	ExportTargetApp DatabaseExportTarget = "app"
)

type DatabaseExportResult struct {
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
	Size     int64  `json:"size"`
	Target   string `json:"target"`
}

type DatabaseImportResult struct {
	Database string `json:"database"`
	FilePath string `json:"filePath"`
}

type DatabaseImportStrategy string

const (
	ImportStrategySource  DatabaseImportStrategy = "source"
	ImportStrategyTarget  DatabaseImportStrategy = "target"
	ImportStrategyReplace DatabaseImportStrategy = "replace"
)

type DatabaseImportOptions struct {
	CreateDatabase bool                   `json:"createDatabase"`
	Strategy       DatabaseImportStrategy `json:"strategy"`
}

type DatabaseDumpService struct{}

type DumpPhaseCallback func(phase, message string)
type DumpProgressCallback func(processed, total int64)

func NewDatabaseDumpService() *DatabaseDumpService {
	return &DatabaseDumpService{}
}

func (s *DatabaseDumpService) ExportMySQLDatabase(connID, database string, target DatabaseExportTarget, targetPath string) (*DatabaseExportResult, error) {
	return s.ExportMySQLDatabaseWithContext(context.Background(), connID, database, target, targetPath, DatabaseExportOptions{
		Mode:              DatabaseExportModeAuto,
		IncludeRoutines:   true,
		IncludeTriggers:   true,
		IncludeEvents:     true,
		IncludeTablespace: false,
	}, nil, nil)
}

func (s *DatabaseDumpService) ExportMySQLDatabaseWithContext(
	ctx context.Context,
	connID,
	database string,
	target DatabaseExportTarget,
	targetPath string,
	options DatabaseExportOptions,
	onPhase DumpPhaseCallback,
	onProgress DumpProgressCallback,
) (*DatabaseExportResult, error) {
	if strings.TrimSpace(database) == "" {
		return nil, fmt.Errorf("database is required")
	}

	savedConn, err := s.resolveSavedConnection(connID)
	if err != nil {
		return nil, err
	}
	if savedConn.DBType != string(db.MySQL) {
		return nil, fmt.Errorf("database export only supports mysql")
	}
	options = normalizeExportOptions(options)

	outputPath, fileName, err := s.resolveExportPath(database, target, targetPath)
	if err != nil {
		return nil, err
	}
	if onPhase != nil {
		onPhase("resolving", "Resolving export target")
	}

	if options.Mode != DatabaseExportModeCompatible {
		args := buildMySQLDumpArgs(savedConn, database, options)

		if onPhase != nil {
			onPhase("dumping", "Running mysqldump")
		}

		if err := s.runMySQLDump(ctx, args, savedConn.Password, outputPath, onProgress); err == nil {
			size, sizeErr := files.GetFileSize(outputPath)
			if sizeErr != nil {
				return nil, sizeErr
			}

			return &DatabaseExportResult{
				FilePath: outputPath,
				FileName: fileName,
				Size:     size,
				Target:   string(target),
			}, nil
		} else if options.Mode == DatabaseExportModeFull || !IsMySQLPermissionError(err) {
			return nil, err
		} else if onPhase != nil {
			onPhase("fallback", "Falling back to compatible export")
		}
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(connID)
	if err != nil {
		return nil, fmt.Errorf("connection not found")
	}

	compatibleSvc := NewDatabaseCompatibleExportService()
	result, err := compatibleSvc.ExportDatabase(managedConn.Conn, database, outputPath, onPhase, onProgress)
	if err != nil {
		return nil, err
	}
	result.Target = string(target)
	return result, nil
}

func (s *DatabaseDumpService) ImportMySQLDatabase(connID, database, filePath string, createDatabase bool) (*DatabaseImportResult, error) {
	return s.ImportMySQLDatabaseWithContext(context.Background(), connID, database, filePath, DatabaseImportOptions{
		CreateDatabase: createDatabase,
		Strategy:       ImportStrategySource,
	}, nil, nil)
}

func (s *DatabaseDumpService) ImportMySQLDatabaseWithContext(
	ctx context.Context,
	connID,
	database,
	filePath string,
	options DatabaseImportOptions,
	onPhase DumpPhaseCallback,
	onProgress DumpProgressCallback,
) (*DatabaseImportResult, error) {
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("file path is required")
	}
	if !files.IsFile(filePath) {
		return nil, fmt.Errorf("import file not found")
	}
	fileSize, _ := files.GetFileSize(filePath)
	if onProgress != nil && fileSize > 0 {
		onProgress(0, fileSize)
	}

	savedConn, err := s.resolveSavedConnection(connID)
	if err != nil {
		return nil, err
	}
	if savedConn.DBType != string(db.MySQL) {
		return nil, fmt.Errorf("database import only supports mysql")
	}
	options = normalizeImportOptions(options)

	targetDatabase := strings.TrimSpace(database)
	if targetDatabase == "" && options.Strategy == ImportStrategySource {
		targetDatabase = strings.TrimSpace(savedConn.Database)
	}
	if (options.Strategy == ImportStrategyTarget || options.Strategy == ImportStrategyReplace) && targetDatabase == "" {
		return nil, fmt.Errorf("target database is required for this import strategy")
	}

	if options.Strategy == ImportStrategyReplace && targetDatabase != "" {
		if onPhase != nil {
			onPhase("drop-db", "Dropping target database before import")
		}
		if err := s.dropDatabaseIfExists(savedConn, targetDatabase); err != nil {
			return nil, err
		}
	}

	preparedFilePath, cleanup, err := s.prepareImportFile(filePath, targetDatabase, options, onPhase)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	if options.CreateDatabase && targetDatabase != "" {
		if onPhase != nil {
			onPhase("prepare-db", "Ensuring target database exists")
		}
		if err := s.ensureDatabaseExists(savedConn, targetDatabase); err != nil {
			return nil, err
		}
	}

	args := []string{
		"--protocol=tcp",
		"--host", savedConn.Host,
		"--port", fmt.Sprintf("%d", savedConn.Port),
		"--user", savedConn.Username,
	}
	if targetDatabase != "" {
		args = append(args, targetDatabase)
	}

	if onPhase != nil {
		onPhase("importing", "Running mysql import")
	}

	if err := s.runMySQLImport(ctx, args, savedConn.Password, preparedFilePath, fileSize, onProgress); err != nil {
		return nil, err
	}

	return &DatabaseImportResult{
		Database: targetDatabase,
		FilePath: filePath,
	}, nil
}

func normalizeImportOptions(options DatabaseImportOptions) DatabaseImportOptions {
	if options.Strategy == "" {
		options.Strategy = ImportStrategySource
	}
	return options
}

func (s *DatabaseDumpService) resolveSavedConnection(connID string) (*config.SavedConnection, error) {
	manager := db.GetManager()
	managedConn, err := manager.Get(connID)
	if err != nil {
		return nil, fmt.Errorf("connection not found")
	}

	cfg := managedConn.Conn.GetConfig()
	store := config.GetConnectionStore()
	allConfigs, err := store.ListConnections()
	if err != nil {
		return nil, err
	}

	for _, item := range allConfigs {
		if item == nil || item.DBType != string(cfg.DBType) {
			continue
		}
		if item.Host == cfg.Host && item.Port == cfg.Port && item.Username == cfg.Username && item.Database == cfg.Database && item.FilePath == cfg.FilePath {
			return store.GetConnection(item.ID)
		}
	}

	return &config.SavedConnection{
		DBType:   string(cfg.DBType),
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		Password: cfg.Password,
		Database: cfg.Database,
		FilePath: cfg.FilePath,
	}, nil
}

func (s *DatabaseDumpService) resolveExportPath(database string, target DatabaseExportTarget, targetPath string) (string, string, error) {
	fileName := fmt.Sprintf("%s_%s.sql", database, time.Now().Format("20060102_150405"))

	if strings.TrimSpace(targetPath) != "" {
		if strings.HasSuffix(strings.ToLower(targetPath), ".sql") {
			if err := files.EnsureDir(filepath.Dir(targetPath)); err != nil {
				return "", "", err
			}
			return targetPath, filepath.Base(targetPath), nil
		}
		if err := files.EnsureDir(targetPath); err != nil {
			return "", "", err
		}
		return filepath.Join(targetPath, fileName), fileName, nil
	}

	appDir := filepath.Join(config.DataDir(), "db-exports")
	if err := files.EnsureDir(appDir); err != nil {
		return "", "", err
	}
	return filepath.Join(appDir, fileName), fileName, nil
}

func (s *DatabaseDumpService) runMySQLDump(ctx context.Context, args []string, password, outputPath string, onProgress DumpProgressCallback) error {
	cmd := exec.CommandContext(ctx, "mysqldump", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", password))

	if err := files.EnsureDir(filepath.Dir(outputPath)); err != nil {
		return err
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	var stderr bytes.Buffer
	cmd.Stdout = outFile
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		_ = files.Remove(outputPath)
		if errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("mysqldump command not found, please install MySQL client tools on the server")
		}
		return fmt.Errorf("mysqldump failed: %s", strings.TrimSpace(stderr.String()))
	}

	if onProgress != nil {
		size, sizeErr := files.GetFileSize(outputPath)
		if sizeErr == nil {
			onProgress(size, size)
		}
	}

	return nil
}

func (s *DatabaseDumpService) runMySQLImport(ctx context.Context, args []string, password, filePath string, fileSize int64, onProgress DumpProgressCallback) error {
	cmd := exec.CommandContext(ctx, "mysql", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", password))

	inFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	var stderr bytes.Buffer
	cmd.Stdin = inFile
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("mysql command not found, please install MySQL client tools on the server")
		}
		return fmt.Errorf("mysql import failed: %s", strings.TrimSpace(stderr.String()))
	}

	if onProgress != nil && fileSize > 0 {
		onProgress(fileSize, fileSize)
	}

	return nil
}

func (s *DatabaseDumpService) ensureDatabaseExists(savedConn *config.SavedConnection, databaseName string) error {
	conn, err := db.NewConnection(db.MySQL)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Connect(db.ConnectionConfig{
		DBType:   db.MySQL,
		Host:     savedConn.Host,
		Port:     savedConn.Port,
		Username: savedConn.Username,
		Password: savedConn.Password,
		Database: "",
	}); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	query := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		strings.ReplaceAll(databaseName, "`", "``"),
	)
	_, err = conn.Exec(ctx, query)
	return err
}

func (s *DatabaseDumpService) dropDatabaseIfExists(savedConn *config.SavedConnection, databaseName string) error {
	conn, err := db.NewConnection(db.MySQL)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Connect(db.ConnectionConfig{
		DBType:   db.MySQL,
		Host:     savedConn.Host,
		Port:     savedConn.Port,
		Username: savedConn.Username,
		Password: savedConn.Password,
		Database: "",
	}); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	query := fmt.Sprintf(
		"DROP DATABASE IF EXISTS `%s`",
		strings.ReplaceAll(databaseName, "`", "``"),
	)
	_, err = conn.Exec(ctx, query)
	return err
}

func (s *DatabaseDumpService) prepareImportFile(
	filePath string,
	targetDatabase string,
	options DatabaseImportOptions,
	onPhase DumpPhaseCallback,
) (string, func(), error) {
	if options.Strategy == ImportStrategySource {
		return filePath, func() {}, nil
	}

	if targetDatabase == "" {
		return "", func() {}, fmt.Errorf("target database is required")
	}

	tempDir := filepath.Join(config.DataDir(), "db-imports", "prepared")
	if err := files.EnsureDir(tempDir); err != nil {
		return "", func() {}, err
	}

	tempPath := filepath.Join(
		tempDir,
		fmt.Sprintf("prepared_%s_%d.sql", sanitizeImportFileName(targetDatabase), time.Now().UnixNano()),
	)

	if onPhase != nil {
		onPhase("rewrite-sql", "Preparing SQL file for target database")
	}

	shouldCreateDatabase := options.CreateDatabase || options.Strategy == ImportStrategyReplace
	if err := rewriteSQLDumpForTarget(filePath, tempPath, targetDatabase, shouldCreateDatabase); err != nil {
		_ = files.Remove(tempPath)
		return "", func() {}, err
	}

	return tempPath, func() {
		_ = files.Remove(tempPath)
	}, nil
}

var (
	mysqlCreateDatabaseLine = regexp.MustCompile(`(?i)^\s*CREATE\s+DATABASE\b`)
	mysqlUseDatabaseLine    = regexp.MustCompile(`(?i)^\s*USE\s+`)
)

func rewriteSQLDumpForTarget(sourcePath, targetPath, targetDatabase string, createDatabase bool) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	reader := bufio.NewReaderSize(sourceFile, 64*1024)
	writer := bufio.NewWriterSize(targetFile, 64*1024)
	defer writer.Flush()

	escapedDatabase := strings.ReplaceAll(targetDatabase, "`", "``")
	if createDatabase {
		if _, err := writer.WriteString(fmt.Sprintf(
			"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;\n",
			escapedDatabase,
		)); err != nil {
			return err
		}
	}
	if _, err := writer.WriteString(fmt.Sprintf("USE `%s`;\n\n", escapedDatabase)); err != nil {
		return err
	}

	for {
		chunk, readErr := reader.ReadSlice('\n')
		if len(chunk) > 0 {
			if readErr == bufio.ErrBufferFull {
				if _, err := writer.Write(chunk); err != nil {
					return err
				}
				continue
			}

			line := string(chunk)
			if !mysqlCreateDatabaseLine.MatchString(line) && !mysqlUseDatabaseLine.MatchString(line) {
				if _, err := writer.Write(chunk); err != nil {
					return err
				}
			}
		}

		if readErr == nil {
			continue
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		return readErr
	}
}

func sanitizeImportFileName(value string) string {
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_':
			builder.WriteRune(r)
		default:
			builder.WriteRune('_')
		}
	}

	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return "database"
	}
	return result
}
