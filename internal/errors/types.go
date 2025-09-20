package errors

import (
	"fmt"
	"strings"
	"time"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Configuration errors
	ErrCodeInvalidConfig      ErrorCode = "INVALID_CONFIG"
	ErrCodeMissingField       ErrorCode = "MISSING_FIELD"
	ErrCodeInvalidFieldType   ErrorCode = "INVALID_FIELD_TYPE"
	ErrCodeInvalidFieldValue  ErrorCode = "INVALID_FIELD_VALUE"

	// Docker errors
	ErrCodeDockerConnection   ErrorCode = "DOCKER_CONNECTION"
	ErrCodeImageNotFound      ErrorCode = "IMAGE_NOT_FOUND"
	ErrCodeContainerFailed    ErrorCode = "CONTAINER_FAILED"
	ErrCodeImageBuildFailed   ErrorCode = "IMAGE_BUILD_FAILED"

	// File system errors
	ErrCodeFileNotFound       ErrorCode = "FILE_NOT_FOUND"
	ErrCodeFilePermission     ErrorCode = "FILE_PERMISSION"
	ErrCodeInvalidFilePath    ErrorCode = "INVALID_FILE_PATH"

	// Network errors
	ErrCodePortInUse          ErrorCode = "PORT_IN_USE"
	ErrCodeNetworkConnection  ErrorCode = "NETWORK_CONNECTION"
	ErrCodeInvalidPortFormat  ErrorCode = "INVALID_PORT_FORMAT"

	// Pipeline execution errors
	ErrCodePipelineValidation ErrorCode = "PIPELINE_VALIDATION"
	ErrCodeJobExecution       ErrorCode = "JOB_EXECUTION"
	ErrCodeScriptFailed       ErrorCode = "SCRIPT_FAILED"
	ErrCodeConditionFailed    ErrorCode = "CONDITION_FAILED"

	// Retry errors
	ErrCodeRetryExhausted     ErrorCode = "RETRY_EXHAUSTED"
	ErrCodeInvalidRetryConfig ErrorCode = "INVALID_RETRY_CONFIG"

	// System errors
	ErrCodeSystemResource     ErrorCode = "SYSTEM_RESOURCE"
	ErrCodePermissionDenied   ErrorCode = "PERMISSION_DENIED"
	ErrCodeTimeout            ErrorCode = "TIMEOUT"
)

// Severity levels for errors
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// PinError represents a comprehensive error with context and suggestions
type PinError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Severity    Severity               `json:"severity"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Cause       error                  `json:"cause,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Job         string                 `json:"job,omitempty"`
	File        string                 `json:"file,omitempty"`
	Line        int                    `json:"line,omitempty"`
}

// Error implements the error interface
func (e *PinError) Error() string {
	var parts []string

	if e.Job != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Job))
	}

	if e.Code != "" {
		parts = append(parts, string(e.Code))
	}

	parts = append(parts, e.Message)

	return strings.Join(parts, " ")
}

// Unwrap returns the underlying cause error
func (e *PinError) Unwrap() error {
	return e.Cause
}

// NewPinError creates a new PinError with basic information
func NewPinError(code ErrorCode, message string) *PinError {
	return &PinError{
		Code:      code,
		Message:   message,
		Severity:  SeverityError,
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// WithSeverity sets the severity level
func (e *PinError) WithSeverity(severity Severity) *PinError {
	e.Severity = severity
	return e
}

// WithContext adds context information
func (e *PinError) WithContext(key string, value interface{}) *PinError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithJob sets the job name where error occurred
func (e *PinError) WithJob(job string) *PinError {
	e.Job = job
	return e
}

// WithFile sets the file path where error occurred
func (e *PinError) WithFile(file string) *PinError {
	e.File = file
	return e
}

// WithCause sets the underlying cause error
func (e *PinError) WithCause(err error) *PinError {
	e.Cause = err
	return e
}

// AddSuggestion adds a suggestion for resolving the error
func (e *PinError) AddSuggestion(suggestion string) *PinError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// AddSuggestions adds multiple suggestions
func (e *PinError) AddSuggestions(suggestions ...string) *PinError {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

// Format returns a formatted error message with suggestions
func (e *PinError) Format() string {
	var parts []string

	// Header with severity emoji
	severityEmoji := map[Severity]string{
		SeverityError:   "❌",
		SeverityWarning: "⚠️",
		SeverityInfo:    "ℹ️",
	}

	emoji, ok := severityEmoji[e.Severity]
	if !ok {
		emoji = "❌"
	}

	header := fmt.Sprintf("%s %s", emoji, e.Error())
	parts = append(parts, header)

	// Context information
	if len(e.Context) > 0 {
		parts = append(parts, "\nContext:")
		for key, value := range e.Context {
			parts = append(parts, fmt.Sprintf("  %s: %v", key, value))
		}
	}

	// Underlying cause
	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("\nCause: %s", e.Cause.Error()))
	}

	// Suggestions
	if len(e.Suggestions) > 0 {
		parts = append(parts, "\nSuggestions:")
		for i, suggestion := range e.Suggestions {
			parts = append(parts, fmt.Sprintf("  %d. %s", i+1, suggestion))
		}
	}

	return strings.Join(parts, "\n")
}

// ToJSON returns a JSON representation of the error
func (e *PinError) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"code":      string(e.Code),
		"message":   e.Message,
		"severity":  string(e.Severity),
		"timestamp": e.Timestamp.Format(time.RFC3339),
	}

	if e.Job != "" {
		result["job"] = e.Job
	}

	if e.File != "" {
		result["file"] = e.File
	}

	if e.Line > 0 {
		result["line"] = e.Line
	}

	if len(e.Context) > 0 {
		result["context"] = e.Context
	}

	if len(e.Suggestions) > 0 {
		result["suggestions"] = e.Suggestions
	}

	if e.Cause != nil {
		result["cause"] = e.Cause.Error()
	}

	return result
}