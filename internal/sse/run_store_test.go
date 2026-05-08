package sse

import (
	"bytes"
	"errors"
	"testing"
)

func TestRunStore_PersistsRunsAndLogs(t *testing.T) {
	runsDir := t.TempDir()
	store := NewRunStoreWithDir(50, runsDir)

	run := store.StartRun("test")
	store.MarkRunning(run.ID)
	if _, err := store.NewLogWriter(run.ID).Write([]byte("hello log\n")); err != nil {
		t.Fatalf("expected log write to succeed: %v", err)
	}
	store.MarkFailed(run.ID, errors.New("boom"))

	reloadedStore := NewRunStoreWithDir(50, runsDir)
	record, ok := reloadedStore.Get(run.ID)
	if !ok {
		t.Fatalf("expected run %s to be loaded from disk", run.ID)
	}
	if record.Status != RunStatusFailed {
		t.Fatalf("expected failed status, got %s", record.Status)
	}

	logContent, err := reloadedStore.ReadLog(run.ID)
	if err != nil {
		t.Fatalf("expected log to be readable: %v", err)
	}
	if !bytes.Contains(logContent, []byte("hello log")) {
		t.Fatalf("expected log content to be persisted, got %s", string(logContent))
	}
}
