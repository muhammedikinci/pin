package errors

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewErrorFormatter(t *testing.T) {
	formatter := NewErrorFormatter(FormatConsole)

	if formatter.format != FormatConsole {
		t.Errorf("Expected format %s, got %s", FormatConsole, formatter.format)
	}

	if !formatter.colorized {
		t.Error("Expected colorized to be true by default")
	}

	if !formatter.showCause {
		t.Error("Expected showCause to be true by default")
	}

	if !formatter.showContext {
		t.Error("Expected showContext to be true by default")
	}
}

func TestErrorFormatter_WithColor(t *testing.T) {
	formatter := NewErrorFormatter(FormatConsole).WithColor(false)

	if formatter.colorized {
		t.Error("Expected colorized to be false")
	}
}

func TestErrorFormatter_WithCause(t *testing.T) {
	formatter := NewErrorFormatter(FormatConsole).WithCause(false)

	if formatter.showCause {
		t.Error("Expected showCause to be false")
	}
}

func TestErrorFormatter_WithContext(t *testing.T) {
	formatter := NewErrorFormatter(FormatConsole).WithContext(false)

	if formatter.showContext {
		t.Error("Expected showContext to be false")
	}
}

func TestErrorFormatter_FormatJSON(t *testing.T) {
	err := NewPinError(ErrCodeInvalidConfig, "test message").
		WithJob("build").
		WithContext("field", "image").
		AddSuggestion("Check configuration")

	formatter := NewErrorFormatter(FormatJSON)
	result := formatter.Format(err)

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(result), &jsonData); jsonErr != nil {
		t.Errorf("Failed to parse JSON output: %v", jsonErr)
	}

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
}

func TestErrorFormatter_FormatPlain(t *testing.T) {
	err := NewPinError(ErrCodeInvalidConfig, "test message").
		WithJob("build").
		AddSuggestion("Check configuration")

	formatter := NewErrorFormatter(FormatPlain)
	result := formatter.Format(err)

	// Plain format should not contain ANSI color codes
	if strings.Contains(result, "\x1b[") {
		t.Error("Plain format should not contain ANSI color codes")
	}

	// Should contain basic elements
	expectedElements := []string{
		"❌ ERROR",
		"INVALID_CONFIG",
		"test message",
		"build",
		"💡 Suggestions:",
		"Check configuration",
	}

	for _, element := range expectedElements {
		if !strings.Contains(result, element) {
			t.Errorf("Expected plain output to contain '%s', got:\n%s", element, result)
		}
	}
}

func TestErrorFormatter_FormatConsole(t *testing.T) {
	err := NewPinError(ErrCodeInvalidConfig, "test message").
		WithJob("build").
		WithContext("field", "image").
		WithContext("value", "alpine:latest").
		AddSuggestion("Check your configuration")

	formatter := NewErrorFormatter(FormatConsole)
	result := formatter.Format(err)

	// Should contain basic elements (may include color codes)
	expectedElements := []string{
		"ERROR",
		"INVALID_CONFIG",
		"test message",
		"build",
		"Context:",
		"field:",
		"value:",
		"💡 Suggestions:",
		"Check your configuration",
	}

	for _, element := range expectedElements {
		if !strings.Contains(result, element) {
			t.Errorf("Expected console output to contain '%s', got:\n%s", element, result)
		}
	}
}

func TestErrorFormatter_FormatConsoleWithoutColor(t *testing.T) {
	err := NewPinError(ErrCodeInvalidConfig, "test message").
		WithJob("build")

	formatter := NewErrorFormatter(FormatConsole).WithColor(false)
	result := formatter.Format(err)

	// Should not contain ANSI color codes when colorized is false
	if strings.Contains(result, "\x1b[") {
		t.Error("Console format without color should not contain ANSI color codes")
	}
}

func TestErrorFormatter_FormatWithoutContext(t *testing.T) {
	err := NewPinError(ErrCodeInvalidConfig, "test message").
		WithContext("field", "image")

	formatter := NewErrorFormatter(FormatConsole).WithContext(false).WithColor(false)
	result := formatter.Format(err)

	// Should not contain context when showContext is false
	if strings.Contains(result, "Context:") {
		t.Error("Output should not contain context when showContext is false")
	}
}

func TestErrorFormatter_FormatWithoutCause(t *testing.T) {
	err := NewPinError(ErrCodeInvalidConfig, "test message").
		WithCause(NewPinError(ErrCodeDockerConnection, "docker error"))

	formatter := NewErrorFormatter(FormatConsole).WithCause(false).WithColor(false)
	result := formatter.Format(err)

	// Should not contain cause when showCause is false
	if strings.Contains(result, "Underlying cause:") {
		t.Error("Output should not contain cause when showCause is false")
	}
}

func TestErrorFormatter_FormatMultiple(t *testing.T) {
	errors := []*PinError{
		NewPinError(ErrCodeInvalidConfig, "first error"),
		NewPinError(ErrCodeDockerConnection, "second error"),
		NewPinError(ErrCodeFileNotFound, "third error"),
	}

	formatter := NewErrorFormatter(FormatConsole).WithColor(false)
	result := formatter.FormatMultiple(errors)

	// Should contain header with error count
	if !strings.Contains(result, "Found 3 errors:") {
		t.Error("Expected header with error count")
	}

	// Should contain all error messages
	for i, err := range errors {
		errorNumber := strings.Contains(result, "Error "+string(rune('1'+i))+":")
		if !errorNumber {
			t.Errorf("Expected to find 'Error %d:'", i+1)
		}

		if !strings.Contains(result, err.Message) {
			t.Errorf("Expected to find error message '%s'", err.Message)
		}
	}
}

func TestErrorFormatter_FormatMultipleSingle(t *testing.T) {
	errors := []*PinError{
		NewPinError(ErrCodeInvalidConfig, "single error"),
	}

	formatter := NewErrorFormatter(FormatConsole).WithColor(false)
	result := formatter.FormatMultiple(errors)

	// Should not contain "Found X errors" header for single error
	if strings.Contains(result, "Found 1 errors:") {
		t.Error("Should not show 'Found X errors' header for single error")
	}

	// Should just show the single error
	if !strings.Contains(result, "single error") {
		t.Error("Expected to find single error message")
	}
}

func TestErrorFormatter_FormatMultipleEmpty(t *testing.T) {
	errors := []*PinError{}

	formatter := NewErrorFormatter(FormatConsole)
	result := formatter.FormatMultiple(errors)

	if result != "" {
		t.Errorf("Expected empty result for empty error array, got: %s", result)
	}
}

func TestSeverityFormatting(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityError, "❌ ERROR"},
		{SeverityWarning, "⚠️  WARNING"},
		{SeverityInfo, "ℹ️  INFO"},
	}

	formatter := NewErrorFormatter(FormatConsole).WithColor(false)

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			err := NewPinError(ErrCodeInvalidConfig, "test").WithSeverity(tt.severity)
			result := formatter.Format(err)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain '%s' for severity %s, got:\n%s", tt.expected, tt.severity, result)
			}
		})
	}
}

func TestDefaultFormatters(t *testing.T) {
	// Test that default formatters are properly initialized
	if ConsoleFormatter == nil {
		t.Error("ConsoleFormatter should not be nil")
	}

	if JSONFormatter == nil {
		t.Error("JSONFormatter should not be nil")
	}

	if PlainFormatter == nil {
		t.Error("PlainFormatter should not be nil")
	}

	// Test PlainFormatter has color disabled
	if PlainFormatter.colorized {
		t.Error("PlainFormatter should have colorized disabled")
	}
}

func TestHighlightCodeBlocks(t *testing.T) {
	formatter := NewErrorFormatter(FormatConsole)

	text := "Example:\n  image: alpine:latest\n  script:\n    - echo hello"
	result := formatter.highlightCodeBlocks(text)

	// When colorized is true, should contain color codes
	if !strings.Contains(result, "Example:") {
		t.Error("Expected result to contain original text")
	}

	// Test without color
	formatter.colorized = false
	resultNoColor := formatter.highlightCodeBlocks(text)

	if resultNoColor != text {
		t.Error("Expected unchanged text when colorized is false")
	}
}