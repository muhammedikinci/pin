package sse

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	RunStatusQueued  = "queued"
	RunStatusRunning = "running"
	RunStatusSuccess = "success"
	RunStatusFailed  = "failed"
)

type RunRecord struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Source      string    `json:"source"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
	LogPath     string    `json:"log_path,omitempty"`
}

type RunStore struct {
	mutex   sync.RWMutex
	runs    map[string]RunRecord
	order   []string
	limit   int
	runsDir string
}

func NewRunStore(limit int) *RunStore {
	return NewRunStoreWithDir(limit, "")
}

func NewRunStoreWithDir(limit int, runsDir string) *RunStore {
	if limit < 1 {
		limit = 50
	}

	rs := &RunStore{
		runs:    make(map[string]RunRecord),
		order:   make([]string, 0, limit),
		limit:   limit,
		runsDir: runsDir,
	}

	if runsDir != "" {
		rs.loadExisting()
	}

	return rs
}

func (rs *RunStore) StartRun(source string) RunRecord {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	record := RunRecord{
		ID:        uuid.NewString(),
		Status:    RunStatusQueued,
		Source:    source,
		StartedAt: time.Now(),
	}
	if rs.runsDir != "" {
		record.LogPath = rs.logPath(record.ID)
	}

	rs.runs[record.ID] = record
	rs.order = append([]string{record.ID}, rs.order...)
	rs.trimLocked()
	rs.persistLocked(record)

	return record
}

func (rs *RunStore) MarkRunning(id string) {
	rs.update(id, func(record RunRecord) RunRecord {
		record.Status = RunStatusRunning
		return record
	})
}

func (rs *RunStore) MarkSucceeded(id string) {
	rs.update(id, func(record RunRecord) RunRecord {
		record.Status = RunStatusSuccess
		record.CompletedAt = time.Now()
		record.Error = ""
		return record
	})
}

func (rs *RunStore) MarkFailed(id string, runError error) {
	rs.update(id, func(record RunRecord) RunRecord {
		record.Status = RunStatusFailed
		record.CompletedAt = time.Now()
		if runError != nil {
			record.Error = runError.Error()
		}
		return record
	})
}

func (rs *RunStore) List() []RunRecord {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	records := make([]RunRecord, 0, len(rs.order))
	for _, id := range rs.order {
		records = append(records, rs.runs[id])
	}

	return records
}

func (rs *RunStore) Get(id string) (RunRecord, bool) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	record, ok := rs.runs[id]
	return record, ok
}

func (rs *RunStore) NewLogWriter(id string) io.Writer {
	return runLogWriter{store: rs, runID: id}
}

func (rs *RunStore) AppendEvent(event Event) {
	if event.Data == nil {
		return
	}

	runID, ok := event.Data["run_id"].(string)
	if !ok || runID == "" {
		return
	}

	message, _ := event.Data["message"].(string)
	if message == "" {
		message = event.Type
	}
	if errorMessage, ok := event.Data["error"].(string); ok && errorMessage != "" {
		message += ": " + errorMessage
	}
	job, _ := event.Data["job"].(string)
	level, _ := event.Data["level"].(string)

	prefix := event.Timestamp.Format(time.RFC3339)
	if event.Timestamp.IsZero() {
		prefix = time.Now().Format(time.RFC3339)
	}
	if level != "" {
		prefix += " [" + level + "]"
	}
	if job != "" {
		prefix += " [" + job + "]"
	}

	_ = rs.AppendLog(runID, []byte(prefix+" "+message+"\n"))
}

func (rs *RunStore) AppendLog(id string, data []byte) error {
	if rs.runsDir == "" || len(data) == 0 {
		return nil
	}
	if !isSafeRunID(id) {
		return fmt.Errorf("invalid run id")
	}
	if err := os.MkdirAll(rs.runsDir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(rs.logPath(id), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func (rs *RunStore) ReadLog(id string) ([]byte, error) {
	if rs.runsDir == "" {
		return nil, os.ErrNotExist
	}
	if !isSafeRunID(id) {
		return nil, fmt.Errorf("invalid run id")
	}

	return os.ReadFile(rs.logPath(id))
}

func (rs *RunStore) update(id string, updateFn func(RunRecord) RunRecord) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	record, ok := rs.runs[id]
	if !ok {
		return
	}

	updatedRecord := updateFn(record)
	rs.runs[id] = updatedRecord
	rs.persistLocked(updatedRecord)
}

func (rs *RunStore) trimLocked() {
	for len(rs.order) > rs.limit {
		lastID := rs.order[len(rs.order)-1]
		delete(rs.runs, lastID)
		rs.order = rs.order[:len(rs.order)-1]
	}
}

func (rs *RunStore) loadExisting() {
	if err := os.MkdirAll(rs.runsDir, 0755); err != nil {
		return
	}

	matches, err := filepath.Glob(filepath.Join(rs.runsDir, "*.json"))
	if err != nil {
		return
	}

	records := make([]RunRecord, 0, len(matches))
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var record RunRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}
		if !isSafeRunID(record.ID) {
			continue
		}
		if record.LogPath == "" {
			record.LogPath = rs.logPath(record.ID)
		}
		records = append(records, record)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].StartedAt.After(records[j].StartedAt)
	})

	for _, record := range records {
		if len(rs.order) >= rs.limit {
			break
		}
		rs.runs[record.ID] = record
		rs.order = append(rs.order, record.ID)
	}
}

func (rs *RunStore) persistLocked(record RunRecord) {
	if rs.runsDir == "" || !isSafeRunID(record.ID) {
		return
	}
	if err := os.MkdirAll(rs.runsDir, 0755); err != nil {
		return
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(rs.recordPath(record.ID), append(data, '\n'), 0644)
}

func (rs *RunStore) recordPath(id string) string {
	return filepath.Join(rs.runsDir, id+".json")
}

func (rs *RunStore) logPath(id string) string {
	return filepath.Join(rs.runsDir, id+".log")
}

func isSafeRunID(id string) bool {
	return id != "" && !strings.Contains(id, "/") && !strings.Contains(id, "\\") && !strings.Contains(id, "..")
}

type runLogWriter struct {
	store *RunStore
	runID string
}

func (w runLogWriter) Write(data []byte) (int, error) {
	if err := w.store.AppendLog(w.runID, data); err != nil {
		return 0, err
	}
	return len(data), nil
}
