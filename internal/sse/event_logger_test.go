package sse

import (
	"bytes"
	"log"
	"testing"

	"github.com/muhammedikinci/pin/internal/interfaces"
	"github.com/muhammedikinci/pin/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestEventLogger_NewEventLogger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)

	// Expect broadcast call with log event
	mockBroadcaster.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(event interfaces.Event) {
			if event.Type != "log" {
				t.Errorf("Expected event type 'log', got '%s'", event.Type)
			}

			data := event.Data
			if data["level"] != "info" {
				t.Errorf("Expected level 'info', got '%v'", data["level"])
			}

			if data["job"] != "test-job" {
				t.Errorf("Expected job 'test-job', got '%v'", data["job"])
			}

			if data["message"] != "test message" {
				t.Errorf("Expected message 'test message', got '%v'", data["message"])
			}
		}).
		Times(1)

	// Capture log output
	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)

	eventLogger := &EventLogger{
		Logger:      logger,
		broadcaster: mockBroadcaster,
		jobName:     "test-job",
	}

	eventLogger.Println("test message")

	// Check that standard logging still works
	if !bytes.Contains(buf.Bytes(), []byte("test message")) {
		t.Error("Expected standard log output to contain message")
	}
}

func TestEventLogger_Printf(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)

	mockBroadcaster.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(event interfaces.Event) {
			data := event.Data
			if data["message"] != "formatted message: 42" {
				t.Errorf("Expected formatted message, got '%v'", data["message"])
			}
		}).
		Times(1)

	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)

	eventLogger := &EventLogger{
		Logger:      logger,
		broadcaster: mockBroadcaster,
		jobName:     "test-job",
	}

	eventLogger.Printf("formatted message: %d", 42)
}

func TestEventLogger_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)

	mockBroadcaster.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(event interfaces.Event) {
			data := event.Data
			if data["level"] != "error" {
				t.Errorf("Expected level 'error', got '%v'", data["level"])
			}
		}).
		Times(1)

	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)

	eventLogger := &EventLogger{
		Logger:      logger,
		broadcaster: mockBroadcaster,
		jobName:     "test-job",
	}

	eventLogger.Error("error message")

	// Check that error is prefixed in standard log
	if !bytes.Contains(buf.Bytes(), []byte("[ERROR]")) {
		t.Error("Expected standard log output to contain [ERROR] prefix")
	}
}

func TestEventLogger_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)

	mockBroadcaster.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(event interfaces.Event) {
			data := event.Data
			if data["level"] != "success" {
				t.Errorf("Expected level 'success', got '%v'", data["level"])
			}
		}).
		Times(1)

	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)

	eventLogger := &EventLogger{
		Logger:      logger,
		broadcaster: mockBroadcaster,
		jobName:     "test-job",
	}

	eventLogger.Success("success message")

	if !bytes.Contains(buf.Bytes(), []byte("[SUCCESS]")) {
		t.Error("Expected standard log output to contain [SUCCESS] prefix")
	}
}

func TestEventLogger_Warning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)

	mockBroadcaster.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(event interfaces.Event) {
			data := event.Data
			if data["level"] != "warning" {
				t.Errorf("Expected level 'warning', got '%v'", data["level"])
			}
		}).
		Times(1)

	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)

	eventLogger := &EventLogger{
		Logger:      logger,
		broadcaster: mockBroadcaster,
		jobName:     "test-job",
	}

	eventLogger.Warning("warning message")

	if !bytes.Contains(buf.Bytes(), []byte("[WARNING]")) {
		t.Error("Expected standard log output to contain [WARNING] prefix")
	}
}

func TestEventLogger_BroadcastJobEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)

	mockBroadcaster.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(event interfaces.Event) {
			if event.Type != "job_started" {
				t.Errorf("Expected event type 'job_started', got '%s'", event.Type)
			}

			data := event.Data
			if data["job"] != "test-job" {
				t.Errorf("Expected job 'test-job', got '%v'", data["job"])
			}

			if data["custom"] != "value" {
				t.Errorf("Expected custom data to be preserved, got '%v'", data["custom"])
			}

			if data["timestamp"] == nil {
				t.Error("Expected timestamp to be added")
			}
		}).
		Times(1)

	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)

	eventLogger := &EventLogger{
		Logger:      logger,
		broadcaster: mockBroadcaster,
		jobName:     "test-job",
	}

	eventLogger.BroadcastJobEvent("job_started", map[string]interface{}{
		"custom": "value",
	})
}

func TestEventLogger_BroadcastJobEvent_NilBroadcaster(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)

	eventLogger := &EventLogger{
		Logger:      logger,
		broadcaster: nil,
		jobName:     "test-job",
	}

	// Should not panic with nil broadcaster
	eventLogger.BroadcastJobEvent("job_started", map[string]interface{}{
		"custom": "value",
	})
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