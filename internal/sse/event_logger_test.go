package sse

import (
	"bytes"
	"log"
	"testing"
)

func TestEventLogger_NewEventLogger(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()

	logger := NewEventLogger(mockBroadcaster, "test-job", "test: ", 0)

	if logger == nil {
		t.Fatal("Expected logger to be created, got nil")
	}

	if logger.jobName != "test-job" {
		t.Errorf("Expected job name 'test-job', got '%s'", logger.jobName)
	}

	if logger.broadcaster != mockBroadcaster {
		t.Error("Expected broadcaster to be set correctly")
	}

	if logger.Logger == nil {
		t.Error("Expected embedded Logger to be initialized")
	}
}

func TestEventLogger_Println(t *testing.T) {
	broadcaster := NewEventBroadcaster()

	// Capture log output
	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)

	eventLogger := &EventLogger{
		Logger:      logger,
		broadcaster: broadcaster,
		jobName:     "test-job",
	}

	eventLogger.Println("test message")

	// Check that standard logging still works
	if !bytes.Contains(buf.Bytes(), []byte("test message")) {
		t.Error("Expected standard log output to contain message")
	}
}

func TestEventLogger_WithNilBroadcaster(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)

	eventLogger := &EventLogger{
		Logger:      logger,
		broadcaster: nil,
		jobName:     "test-job",
	}

	// Should still log to standard output without panicking
	eventLogger.Println("test message")
	eventLogger.Error("error message")
	eventLogger.Success("success message")

	// Check standard logging still works
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("test message")) {
		t.Error("Expected standard log output to contain test message")
	}
	if !bytes.Contains([]byte(output), []byte("[ERROR]")) {
		t.Error("Expected standard log output to contain error")
	}
	if !bytes.Contains([]byte(output), []byte("[SUCCESS]")) {
		t.Error("Expected standard log output to contain success")
	}
}