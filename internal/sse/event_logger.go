package sse

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// EventLogger wraps a standard logger and broadcasts log events via SSE
type EventLogger struct {
	*log.Logger
	broadcaster EventBroadcaster
	jobName     string
	runID       string
	logWriter   io.Writer
	logLevel    string
}

type EventLoggerOption func(*EventLogger)

func WithRunContext(runID string, logWriter io.Writer) EventLoggerOption {
	return func(logger *EventLogger) {
		logger.runID = runID
		logger.logWriter = logWriter
	}
}

// NewEventLogger creates a new event-aware logger
func NewEventLogger(broadcaster EventBroadcaster, jobName string, prefix string, flag int, options ...EventLoggerOption) *EventLogger {
	standardLogger := log.New(os.Stdout, prefix, flag)

	eventLogger := &EventLogger{
		Logger:      standardLogger,
		broadcaster: broadcaster,
		jobName:     jobName,
		logLevel:    "info",
	}

	for _, option := range options {
		option(eventLogger)
	}

	return eventLogger
}

// Println logs to standard output and broadcasts as an event
func (el *EventLogger) Println(v ...interface{}) {
	// Call original logger method
	el.Logger.Println(v...)

	message := fmt.Sprint(v...)
	el.writeRunLog("info", message)

	// Broadcast event if broadcaster is available
	if el.broadcaster != nil {
		el.broadcastLogEvent("info", message)
	}
}

// Printf logs to standard output and broadcasts as an event
func (el *EventLogger) Printf(format string, v ...interface{}) {
	// Call original logger method
	el.Logger.Printf(format, v...)

	message := fmt.Sprintf(format, v...)
	el.writeRunLog("info", message)

	// Broadcast event if broadcaster is available
	if el.broadcaster != nil {
		el.broadcastLogEvent("info", message)
	}
}

// Error logs an error message and broadcasts as an error event
func (el *EventLogger) Error(v ...interface{}) {
	// Log to standard output with error prefix
	message := fmt.Sprint(v...)
	el.Logger.Printf("[ERROR] %s", message)
	el.writeRunLog("error", message)

	// Broadcast event if broadcaster is available
	if el.broadcaster != nil {
		el.broadcastLogEvent("error", message)
	}
}

// Errorf logs a formatted error message and broadcasts as an error event
func (el *EventLogger) Errorf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	el.Logger.Printf("[ERROR] %s", message)
	el.writeRunLog("error", message)

	// Broadcast event if broadcaster is available
	if el.broadcaster != nil {
		el.broadcastLogEvent("error", message)
	}
}

// Success logs a success message and broadcasts as a success event
func (el *EventLogger) Success(v ...interface{}) {
	message := fmt.Sprint(v...)
	el.Logger.Printf("[SUCCESS] %s", message)
	el.writeRunLog("success", message)

	// Broadcast event if broadcaster is available
	if el.broadcaster != nil {
		el.broadcastLogEvent("success", message)
	}
}

// Successf logs a formatted success message and broadcasts as a success event
func (el *EventLogger) Successf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	el.Logger.Printf("[SUCCESS] %s", message)
	el.writeRunLog("success", message)

	// Broadcast event if broadcaster is available
	if el.broadcaster != nil {
		el.broadcastLogEvent("success", message)
	}
}

// Warning logs a warning message and broadcasts as a warning event
func (el *EventLogger) Warning(v ...interface{}) {
	message := fmt.Sprint(v...)
	el.Logger.Printf("[WARNING] %s", message)
	el.writeRunLog("warning", message)

	// Broadcast event if broadcaster is available
	if el.broadcaster != nil {
		el.broadcastLogEvent("warning", message)
	}
}

// Warningf logs a formatted warning message and broadcasts as a warning event
func (el *EventLogger) Warningf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	el.Logger.Printf("[WARNING] %s", message)
	el.writeRunLog("warning", message)

	// Broadcast event if broadcaster is available
	if el.broadcaster != nil {
		el.broadcastLogEvent("warning", message)
	}
}

// broadcastLogEvent creates and broadcasts a log event
func (el *EventLogger) broadcastLogEvent(level, message string) {
	event := Event{
		Type: "log",
		Data: map[string]interface{}{
			"level":     level,
			"message":   message,
			"job":       el.jobName,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	}
	if el.runID != "" {
		event.Data["run_id"] = el.runID
	}

	el.broadcaster.Broadcast(event)
}

// BroadcastJobEvent broadcasts a job-specific event (start, complete, error, etc.)
func (el *EventLogger) BroadcastJobEvent(eventType string, data map[string]interface{}) {
	if el.broadcaster == nil {
		return
	}

	// Add job name and timestamp to the data
	if data == nil {
		data = make(map[string]interface{})
	}
	data["job"] = el.jobName
	data["timestamp"] = time.Now()
	if el.runID != "" {
		data["run_id"] = el.runID
	}

	event := Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	el.broadcaster.Broadcast(event)
}

func (el *EventLogger) writeRunLog(level, message string) {
	if el.logWriter == nil {
		return
	}

	_, _ = fmt.Fprintf(
		el.logWriter,
		"%s [%s] [%s] %s\n",
		time.Now().Format(time.RFC3339),
		level,
		el.jobName,
		message,
	)
}
