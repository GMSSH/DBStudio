package service

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"sqlmanager/pkg/config"
	"sqlmanager/pkg/db"
	"sqlmanager/pkg/files"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	DatabaseDumpTaskTypeExport = "export"
	DatabaseDumpTaskTypeImport = "import"

	DatabaseDumpTaskStatusPending  = "pending"
	DatabaseDumpTaskStatusRunning  = "running"
	DatabaseDumpTaskStatusSuccess  = "success"
	DatabaseDumpTaskStatusFailed   = "failed"
	DatabaseDumpTaskStatusCanceled = "canceled"
)

type DatabaseDumpTask struct {
	ID                 string    `json:"id"`
	Type               string    `json:"type"`
	Mode               string    `json:"mode,omitempty"`
	Strategy           string    `json:"strategy,omitempty"`
	Status             string    `json:"status"`
	Phase              string    `json:"phase"`
	Message            string    `json:"message"`
	ConnID             string    `json:"connId"`
	ConnectionConfigID string    `json:"connectionConfigId,omitempty"`
	ConnectionName     string    `json:"connectionName,omitempty"`
	Database           string    `json:"database"`
	Target             string    `json:"target,omitempty"`
	TargetPath         string    `json:"targetPath,omitempty"`
	FilePath           string    `json:"filePath,omitempty"`
	FileName           string    `json:"fileName,omitempty"`
	Size               int64     `json:"size,omitempty"`
	TotalBytes         int64     `json:"totalBytes,omitempty"`
	ProcessedBytes     int64     `json:"processedBytes,omitempty"`
	ProgressPercent    int       `json:"progressPercent,omitempty"`
	Error              string    `json:"error,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	StartedAt          time.Time `json:"startedAt,omitempty"`
	FinishedAt         time.Time `json:"finishedAt,omitempty"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type DatabaseDumpTaskLog struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"taskId"`
	Level     string    `json:"level"`
	Phase     string    `json:"phase"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type databaseDumpTaskEntry struct {
	mu     sync.RWMutex
	task   DatabaseDumpTask
	cancel context.CancelFunc
}

func (e *databaseDumpTaskEntry) snapshot() *DatabaseDumpTask {
	e.mu.RLock()
	defer e.mu.RUnlock()

	taskCopy := e.task
	return &taskCopy
}

func (e *databaseDumpTaskEntry) mutate(fn func(task *DatabaseDumpTask)) {
	e.mu.Lock()
	defer e.mu.Unlock()

	fn(&e.task)
	e.task.UpdatedAt = time.Now()
}

type DatabaseDumpTaskService struct {
	mu    sync.RWMutex
	tasks map[string]*databaseDumpTaskEntry
	logs  map[string][]DatabaseDumpTaskLog
}

type databaseDumpTaskState struct {
	Tasks []*DatabaseDumpTask              `json:"tasks"`
	Logs  map[string][]DatabaseDumpTaskLog `json:"logs"`
}

var (
	globalDatabaseDumpTaskService *DatabaseDumpTaskService
	databaseDumpTaskServiceOnce   sync.Once
)

func NewDatabaseDumpTaskService() *DatabaseDumpTaskService {
	databaseDumpTaskServiceOnce.Do(func() {
		globalDatabaseDumpTaskService = &DatabaseDumpTaskService{
			tasks: make(map[string]*databaseDumpTaskEntry),
			logs:  make(map[string][]DatabaseDumpTaskLog),
		}
		globalDatabaseDumpTaskService.loadPersistedState()
	})

	return globalDatabaseDumpTaskService
}

func (s *DatabaseDumpTaskService) StartExportTask(connID, database string, target DatabaseExportTarget, targetPath string, options DatabaseExportOptions) (*DatabaseDumpTask, error) {
	taskID := uuid.NewString()
	now := time.Now()
	options = normalizeExportOptions(options)
	configID, connectionName := s.resolveConnectionMeta(connID)
	entry := &databaseDumpTaskEntry{
		task: DatabaseDumpTask{
			ID:                 taskID,
			Type:               DatabaseDumpTaskTypeExport,
			Mode:               string(options.Mode),
			Status:             DatabaseDumpTaskStatusPending,
			Phase:              "queued",
			Message:            "Export task queued",
			ConnID:             connID,
			ConnectionConfigID: configID,
			ConnectionName:     connectionName,
			Database:           database,
			Target:             string(target),
			TargetPath:         targetPath,
			CreatedAt:          now,
			UpdatedAt:          now,
		},
	}

	s.mu.Lock()
	s.tasks[taskID] = entry
	s.cleanupLocked()
	s.mu.Unlock()
	s.persistToDisk()

	go s.runExportTask(entry, connID, database, target, targetPath, options)

	return entry.snapshot(), nil
}

func (s *DatabaseDumpTaskService) StartImportTask(connID, database, filePath string, options DatabaseImportOptions) (*DatabaseDumpTask, error) {
	taskID := uuid.NewString()
	now := time.Now()
	options = normalizeImportOptions(options)
	configID, connectionName := s.resolveConnectionMeta(connID)
	entry := &databaseDumpTaskEntry{
		task: DatabaseDumpTask{
			ID:                 taskID,
			Type:               DatabaseDumpTaskTypeImport,
			Strategy:           string(options.Strategy),
			Status:             DatabaseDumpTaskStatusPending,
			Phase:              "queued",
			Message:            "Import task queued",
			ConnID:             connID,
			ConnectionConfigID: configID,
			ConnectionName:     connectionName,
			Database:           database,
			FilePath:           filePath,
			CreatedAt:          now,
			UpdatedAt:          now,
		},
	}

	s.mu.Lock()
	s.tasks[taskID] = entry
	s.cleanupLocked()
	s.mu.Unlock()
	s.persistToDisk()

	go s.runImportTask(entry, connID, database, filePath, options)

	return entry.snapshot(), nil
}

func (s *DatabaseDumpTaskService) ListTasks() []*DatabaseDumpTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*DatabaseDumpTask, 0, len(s.tasks))
	for _, entry := range s.tasks {
		result = append(result, entry.snapshot())
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result
}

func (s *DatabaseDumpTaskService) GetTask(taskID string) (*DatabaseDumpTask, error) {
	entry := s.getEntry(taskID)
	if entry == nil {
		return nil, fmt.Errorf("task not found")
	}

	return entry.snapshot(), nil
}

func (s *DatabaseDumpTaskService) GetTaskLogs(taskID string) ([]DatabaseDumpTaskLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.tasks[taskID]; !ok {
		return nil, fmt.Errorf("task not found")
	}

	logs := s.logs[taskID]
	result := make([]DatabaseDumpTaskLog, len(logs))
	copy(result, logs)
	return result, nil
}

func (s *DatabaseDumpTaskService) CancelTask(taskID string) (*DatabaseDumpTask, error) {
	entry := s.getEntry(taskID)
	if entry == nil {
		return nil, fmt.Errorf("task not found")
	}

	entry.mutate(func(task *DatabaseDumpTask) {
		if task.Status == DatabaseDumpTaskStatusSuccess || task.Status == DatabaseDumpTaskStatusFailed || task.Status == DatabaseDumpTaskStatusCanceled {
			return
		}
		task.Phase = "canceling"
		task.Message = "Cancel requested"
	})
	s.appendLog(taskID, "warning", "canceling", "Cancel requested")

	entry.mu.RLock()
	cancel := entry.cancel
	entry.mu.RUnlock()
	if cancel != nil {
		cancel()
	}

	return entry.snapshot(), nil
}

func (s *DatabaseDumpTaskService) runExportTask(entry *databaseDumpTaskEntry, connID, database string, target DatabaseExportTarget, targetPath string, options DatabaseExportOptions) {
	ctx, cancel := context.WithCancel(context.Background())
	entry.mu.Lock()
	entry.cancel = cancel
	entry.mu.Unlock()

	entry.mutate(func(task *DatabaseDumpTask) {
		task.Status = DatabaseDumpTaskStatusRunning
		task.Phase = "starting"
		task.Message = "Preparing database export"
		task.StartedAt = time.Now()
	})
	s.appendLog(entry.task.ID, "info", "starting", "Preparing database export")

	dumpSvc := NewDatabaseDumpService()
	result, err := dumpSvc.ExportMySQLDatabaseWithContext(
		ctx,
		connID,
		database,
		target,
		targetPath,
		options,
		func(phase, message string) {
			entry.mutate(func(task *DatabaseDumpTask) {
				task.Phase = phase
				task.Message = message
			})
			s.appendLog(entry.task.ID, "info", phase, message)
		},
		func(processed, total int64) {
			entry.mutate(func(task *DatabaseDumpTask) {
				task.ProcessedBytes = processed
				task.TotalBytes = total
				if total > 0 {
					task.ProgressPercent = int((processed * 100) / total)
				}
			})
		},
	)

	finishedAt := time.Now()
	if err != nil {
		entry.mutate(func(task *DatabaseDumpTask) {
			task.FinishedAt = finishedAt
			if ctx.Err() == context.Canceled {
				task.Status = DatabaseDumpTaskStatusCanceled
				task.Phase = "canceled"
				task.Message = "Export canceled"
				task.Error = ""
				return
			}

			task.Status = DatabaseDumpTaskStatusFailed
			task.Phase = "failed"
			task.Message = "Export failed"
			task.Error = err.Error()
		})
		if ctx.Err() == context.Canceled {
			s.appendLog(entry.task.ID, "warning", "canceled", "Export canceled")
		} else {
			s.appendLog(entry.task.ID, "error", "failed", err.Error())
		}
		return
	}

	entry.mutate(func(task *DatabaseDumpTask) {
		task.Status = DatabaseDumpTaskStatusSuccess
		task.Phase = "completed"
		task.Message = "Export completed"
		task.Mode = string(options.Mode)
		task.FilePath = result.FilePath
		task.FileName = result.FileName
		task.Size = result.Size
		task.ProcessedBytes = result.Size
		task.FinishedAt = finishedAt
	})
	s.appendLog(entry.task.ID, "info", "completed", "Export completed")
}

func (s *DatabaseDumpTaskService) runImportTask(entry *databaseDumpTaskEntry, connID, database, filePath string, options DatabaseImportOptions) {
	ctx, cancel := context.WithCancel(context.Background())
	entry.mu.Lock()
	entry.cancel = cancel
	entry.mu.Unlock()

	entry.mutate(func(task *DatabaseDumpTask) {
		task.Status = DatabaseDumpTaskStatusRunning
		task.Phase = "starting"
		task.Message = "Preparing database import"
		task.StartedAt = time.Now()
	})
	s.appendLog(entry.task.ID, "info", "starting", "Preparing database import")

	dumpSvc := NewDatabaseDumpService()
	result, err := dumpSvc.ImportMySQLDatabaseWithContext(
		ctx,
		connID,
		database,
		filePath,
		options,
		func(phase, message string) {
			entry.mutate(func(task *DatabaseDumpTask) {
				task.Phase = phase
				task.Message = message
			})
			s.appendLog(entry.task.ID, "info", phase, message)
		},
		func(processed, total int64) {
			entry.mutate(func(task *DatabaseDumpTask) {
				task.ProcessedBytes = processed
				task.TotalBytes = total
				if total > 0 {
					task.ProgressPercent = int((processed * 100) / total)
				}
			})
		},
	)

	finishedAt := time.Now()
	if err != nil {
		entry.mutate(func(task *DatabaseDumpTask) {
			task.FinishedAt = finishedAt
			if ctx.Err() == context.Canceled {
				task.Status = DatabaseDumpTaskStatusCanceled
				task.Phase = "canceled"
				task.Message = "Import canceled"
				task.Error = ""
				return
			}

			task.Status = DatabaseDumpTaskStatusFailed
			task.Phase = "failed"
			task.Message = "Import failed"
			task.Error = err.Error()
		})
		if ctx.Err() == context.Canceled {
			s.appendLog(entry.task.ID, "warning", "canceled", "Import canceled")
		} else {
			s.appendLog(entry.task.ID, "error", "failed", err.Error())
		}
		return
	}

	entry.mutate(func(task *DatabaseDumpTask) {
		task.Status = DatabaseDumpTaskStatusSuccess
		task.Phase = "completed"
		task.Message = "Import completed"
		task.Database = result.Database
		task.FilePath = result.FilePath
		task.FinishedAt = finishedAt
		task.ProcessedBytes = task.TotalBytes
		task.ProgressPercent = 100
	})
	s.appendLog(entry.task.ID, "info", "completed", "Import completed")
}

func (s *DatabaseDumpTaskService) getEntry(taskID string) *databaseDumpTaskEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.tasks[taskID]
}

func (s *DatabaseDumpTaskService) cleanupLocked() {
	if len(s.tasks) <= 32 {
		return
	}

	threshold := time.Now().Add(-24 * time.Hour)
	for id, entry := range s.tasks {
		task := entry.snapshot()
		if task.Status == DatabaseDumpTaskStatusRunning || task.Status == DatabaseDumpTaskStatusPending {
			continue
		}
		if task.UpdatedAt.Before(threshold) {
			delete(s.tasks, id)
			delete(s.logs, id)
		}
	}
}

func (s *DatabaseDumpTaskService) appendLog(taskID, level, phase, message string) {
	if taskID == "" || message == "" {
		return
	}

	s.mu.Lock()
	s.logs[taskID] = append(s.logs[taskID], DatabaseDumpTaskLog{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		TaskID:    taskID,
		Level:     level,
		Phase:     phase,
		Message:   message,
		CreatedAt: time.Now(),
	})
	s.mu.Unlock()
	s.persistToDisk()
}

func (s *DatabaseDumpTaskService) loadPersistedState() {
	statePath := databaseDumpTaskStatePath()
	if !files.Exists(statePath) {
		return
	}

	var persisted databaseDumpTaskState
	if err := files.ReadJSON(statePath, &persisted); err != nil {
		log.Printf("[DatabaseDumpTaskService] failed to load persisted tasks: %v", err)
		return
	}
	if persisted.Logs == nil {
		persisted.Logs = make(map[string][]DatabaseDumpTaskLog)
	}

	now := time.Now()
	needsPersist := false
	for _, task := range persisted.Tasks {
		if task == nil || task.ID == "" {
			continue
		}

		taskCopy := *task
		if taskCopy.Status == DatabaseDumpTaskStatusPending || taskCopy.Status == DatabaseDumpTaskStatusRunning {
			taskCopy.Status = DatabaseDumpTaskStatusFailed
			taskCopy.Phase = "interrupted"
			taskCopy.Message = "Task interrupted after app restart"
			if taskCopy.Error == "" {
				taskCopy.Error = "Task interrupted after app restart"
			}
			taskCopy.FinishedAt = now
			taskCopy.UpdatedAt = now
			persisted.Logs[taskCopy.ID] = append(persisted.Logs[taskCopy.ID], DatabaseDumpTaskLog{
				ID:        fmt.Sprintf("%d", now.UnixNano()),
				TaskID:    taskCopy.ID,
				Level:     "warning",
				Phase:     "interrupted",
				Message:   "Task interrupted after app restart",
				CreatedAt: now,
			})
			needsPersist = true
		}

		s.tasks[taskCopy.ID] = &databaseDumpTaskEntry{
			task: taskCopy,
		}
	}

	if persisted.Logs != nil {
		s.logs = persisted.Logs
	}

	if needsPersist {
		s.persistToDisk()
	}
}

func (s *DatabaseDumpTaskService) persistToDisk() {
	s.mu.RLock()
	logsCopy := make(map[string][]DatabaseDumpTaskLog, len(s.logs))
	for taskID, taskLogs := range s.logs {
		cloned := make([]DatabaseDumpTaskLog, len(taskLogs))
		copy(cloned, taskLogs)
		logsCopy[taskID] = cloned
	}

	tasks := make([]*DatabaseDumpTask, 0, len(s.tasks))
	for _, entry := range s.tasks {
		tasks = append(tasks, entry.snapshot())
	}
	s.mu.RUnlock()

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	if err := files.WriteJSON(databaseDumpTaskStatePath(), databaseDumpTaskState{
		Tasks: tasks,
		Logs:  logsCopy,
	}, true); err != nil {
		log.Printf("[DatabaseDumpTaskService] failed to persist tasks: %v", err)
	}
}

func databaseDumpTaskStatePath() string {
	return filepath.Join(config.DataDir(), "db-transfer", "tasks.json")
}

func (s *DatabaseDumpTaskService) resolveConnectionMeta(connID string) (string, string) {
	if connID == "" {
		return "", ""
	}

	manager := db.GetManager()
	managedConn, err := manager.Get(connID)
	if err != nil || managedConn == nil || managedConn.Conn == nil {
		return "", ""
	}

	cfg := managedConn.Conn.GetConfig()
	store := config.GetConnectionStore()
	allConfigs, err := store.ListConnections()
	if err != nil {
		return "", ""
	}

	for _, item := range allConfigs {
		if item == nil || item.DBType != string(cfg.DBType) {
			continue
		}
		if item.Host == cfg.Host && item.Port == cfg.Port && item.Username == cfg.Username && item.Database == cfg.Database && item.FilePath == cfg.FilePath {
			return item.ID, item.Name
		}
	}

	return "", ""
}
