package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewPinError(t *testing.T) {
	code := ErrCodeInvalidConfig
	message := "test error message"

	err := NewPinError(code, message)

	if err.Code != code {
		t.Errorf("Expected code %s, got %s", code, err.Code)
	}

	if err.Message != message {
		t.Errorf("Expected message %s, got %s", message, err.Message)
	}

	if err.Severity != SeverityError {
		t.Errorf("Expected severity %s, got %s", SeverityError, err.Severity)
	}

	if err.Context == nil {
		t.Error("Expected Context to be initialized")
	}

	if time.Since(err.Timestamp) > time.Second {
		t.Error("Expected Timestamp to be recent")
	}
}

func TestPinError_WithSeverity(t *testing.T) {
	err := NewPinError(ErrCodeInvalidConfig, "test").WithSeverity(SeverityWarning)

	if err.Severity != SeverityWarning {
		t.Errorf("Expected severity %s, got %s", SeverityWarning, err.Severity)
	}
}

func TestPinError_WithContext(t *testing.T) {
	err := NewPinError(ErrCodeInvalidConfig, "test").
		WithContext("field", "image").
		WithContext("value", "alpine:latest")

	if err.Context["field"] != "image" {
		t.Errorf("Expected context field to be 'image', got %v", err.Context["field"])
	}

	if err.Context["value"] != "alpine:latest" {
		t.Errorf("Expected context value to be 'alpine:latest', got %v", err.Context["value"])
	}
}

func TestPinError_WithJob(t *testing.T) {
	jobName := "build"
	err := NewPinError(ErrCodeInvalidConfig, "test").WithJob(jobName)

	if err.Job != jobName {
		t.Errorf("Expected job %s, got %s", jobName, err.Job)
	}
}

func TestPinError_WithFile(t *testing.T) {
	filePath := "./pipeline.yaml"
	err := NewPinError(ErrCodeInvalidConfig, "test").WithFile(filePath)

	if err.File != filePath {
		t.Errorf("Expected file %s, got %s", filePath, err.File)
	}
}

func TestPinError_WithCause(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := NewPinError(ErrCodeInvalidConfig, "test").WithCause(cause)

	if err.Cause != cause {
		t.Errorf("Expected cause %v, got %v", cause, err.Cause)
	}
}

func TestPinError_AddSuggestion(t *testing.T) {
	err := NewPinError(ErrCodeInvalidConfig, "test").
		AddSuggestion("First suggestion").
		AddSuggestion("Second suggestion")

	if len(err.Suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(err.Suggestions))
	}

	if err.Suggestions[0] != "First suggestion" {
		t.Errorf("Expected first suggestion to be 'First suggestion', got %s", err.Suggestions[0])
	}

	if err.Suggestions[1] != "Second suggestion" {
		t.Errorf("Expected second suggestion to be 'Second suggestion', got %s", err.Suggestions[1])
	}
}

func TestPinError_AddSuggestions(t *testing.T) {
	suggestions := []string{"Suggestion 1", "Suggestion 2", "Suggestion 3"}
	err := NewPinError(ErrCodeInvalidConfig, "test").AddSuggestions(suggestions...)

	if len(err.Suggestions) != 3 {
		t.Errorf("Expected 3 suggestions, got %d", len(err.Suggestions))
	}

	for i, suggestion := range suggestions {
		if err.Suggestions[i] != suggestion {
			t.Errorf("Expected suggestion %d to be '%s', got '%s'", i, suggestion, err.Suggestions[i])
		}
	}
}

func TestPinError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *PinError
		expected string
	}{
		{
			name:     "Basic error",
			err:      NewPinError(ErrCodeInvalidConfig, "test message"),
			expected: "INVALID_CONFIG test message",
		},
		{
			name:     "Error with job",
			err:      NewPinError(ErrCodeInvalidConfig, "test message").WithJob("build"),
			expected: "[build] INVALID_CONFIG test message",
		},
		{
			name:     "Error without code",
			err:      &PinError{Message: "test message"},
			expected: "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Expected error string '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPinError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := NewPinError(ErrCodeInvalidConfig, "test").WithCause(cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Expected unwrapped error %v, got %v", cause, unwrapped)
	}

	// Test without cause
	errNoCause := NewPinError(ErrCodeInvalidConfig, "test")
	if errNoCause.Unwrap() != nil {
		t.Error("Expected nil when unwrapping error without cause")
	}
}

func TestPinError_Format(t *testing.T) {
	err := NewPinError(ErrCodeInvalidConfig, "test message").
		WithJob("build").
		WithContext("field", "image").
		WithContext("value", "alpine:latest").
		AddSuggestion("Check your configuration").
		AddSuggestion("Validate YAML syntax")

	formatted := err.Format()

	// Check that formatted output contains expected elements
	expectedElements := []string{
		"❌",
		"INVALID_CONFIG",
		"test message",
		"Context:",
		"field:",
		"value:",
		"Suggestions:",
		"Check your configuration",
		"Validate YAML syntax",
	}

	for _, element := range expectedElements {
		if !strings.Contains(formatted, element) {
			t.Errorf("Expected formatted output to contain '%s', got:\n%s", element, formatted)
		}
	}
}

func TestPinError_ToJSON(t *testing.T) {
	err := NewPinError(ErrCodeInvalidConfig, "test message").
		WithJob("build").
		WithFile("pipeline.yaml").
		WithContext("field", "image").
		AddSuggestion("Check configuration")

	jsonData := err.ToJSON()

	// Check required fields
	if jsonData["code"] != string(ErrCodeInvalidConfig) {
		t.Errorf("Expected code %s in JSON, got %v", ErrCodeInvalidConfig, jsonData["code"])
	}

	if jsonData["message"] != "test message" {
		t.Errorf("Expected message 'test message' in JSON, got %v", jsonData["message"])
	}

	if jsonData["job"] != "build" {
		t.Errorf("Expected job 'build' in JSON, got %v", jsonData["job"])
	}

	if jsonData["file"] != "pipeline.yaml" {
		t.Errorf("Expected file 'pipeline.yaml' in JSON, got %v", jsonData["file"])
	}

	// Check context
	context, ok := jsonData["context"].(map[string]interface{})
	if !ok {
		t.Error("Expected context to be a map")
	}

	if context["field"] != "image" {
		t.Errorf("Expected context field to be 'image', got %v", context["field"])
	}

	// Check suggestions
	suggestions, ok := jsonData["suggestions"].([]string)
	if !ok {
		t.Error("Expected suggestions to be a string array")
	}

	if len(suggestions) != 1 || suggestions[0] != "Check configuration" {
		t.Errorf("Expected suggestions to contain 'Check configuration', got %v", suggestions)
	}

	// Verify JSON marshaling works
	_, err2 := json.Marshal(jsonData)
	if err2 != nil {
		t.Errorf("Failed to marshal JSON data: %v", err2)
	}
}

func TestErrorCodes(t *testing.T) {
	// Test that all error codes are non-empty strings
	codes := []ErrorCode{
		ErrCodeInvalidConfig,
		ErrCodeMissingField,
		ErrCodeInvalidFieldType,
		ErrCodeInvalidFieldValue,
		ErrCodeDockerConnection,
		ErrCodeImageNotFound,
		ErrCodeContainerFailed,
		ErrCodeImageBuildFailed,
		ErrCodeFileNotFound,
		ErrCodeFilePermission,
		ErrCodeInvalidFilePath,
		ErrCodePortInUse,
		ErrCodeNetworkConnection,
		ErrCodeInvalidPortFormat,
		ErrCodePipelineValidation,
		ErrCodeJobExecution,
		ErrCodeScriptFailed,
		ErrCodeConditionFailed,
		ErrCodeRetryExhausted,
		ErrCodeInvalidRetryConfig,
		ErrCodeSystemResource,
		ErrCodePermissionDenied,
		ErrCodeTimeout,
	}

	for _, code := range codes {
		if string(code) == "" {
			t.Errorf("Error code %v is empty", code)
		}
	}
}

func TestSeverity(t *testing.T) {
	// Test that all severity levels are non-empty strings
	severities := []Severity{
		SeverityError,
		SeverityWarning,
		SeverityInfo,
	}

	for _, severity := range severities {
		if string(severity) == "" {
			t.Errorf("Severity %v is empty", severity)
		}
	}
}