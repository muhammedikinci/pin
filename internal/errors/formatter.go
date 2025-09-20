package errors

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// OutputFormat represents different error output formats
type OutputFormat string

const (
	FormatConsole OutputFormat = "console"
	FormatJSON    OutputFormat = "json"
	FormatPlain   OutputFormat = "plain"
)

// ErrorFormatter handles formatting of errors for different outputs
type ErrorFormatter struct {
	format     OutputFormat
	colorized  bool
	showCause  bool
	showContext bool
}

// NewErrorFormatter creates a new error formatter
func NewErrorFormatter(format OutputFormat) *ErrorFormatter {
	return &ErrorFormatter{
		format:      format,
		colorized:   true,
		showCause:   true,
		showContext: true,
	}
}

// WithColor enables or disables colorized output
func (f *ErrorFormatter) WithColor(enabled bool) *ErrorFormatter {
	f.colorized = enabled
	return f
}

// WithCause enables or disables showing the underlying cause
func (f *ErrorFormatter) WithCause(enabled bool) *ErrorFormatter {
	f.showCause = enabled
	return f
}

// WithContext enables or disables showing context information
func (f *ErrorFormatter) WithContext(enabled bool) *ErrorFormatter {
	f.showContext = enabled
	return f
}

// Format formats a PinError according to the configured format
func (f *ErrorFormatter) Format(err *PinError) string {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(err)
	case FormatPlain:
		return f.formatPlain(err)
	case FormatConsole:
		fallthrough
	default:
		return f.formatConsole(err)
	}
}

// formatConsole formats error for console output with colors and formatting
func (f *ErrorFormatter) formatConsole(err *PinError) string {
	var parts []string

	// Header with severity and job information
	header := f.buildHeader(err)
	parts = append(parts, header)

	// Main error message
	message := f.formatMessage(err.Message)
	parts = append(parts, message)

	// Context information
	if f.showContext && len(err.Context) > 0 {
		contextStr := f.formatContext(err.Context)
		parts = append(parts, contextStr)
	}

	// Underlying cause
	if f.showCause && err.Cause != nil {
		causeStr := f.formatCause(err.Cause)
		parts = append(parts, causeStr)
	}

	// Suggestions
	if len(err.Suggestions) > 0 {
		suggestionsStr := f.formatSuggestions(err.Suggestions)
		parts = append(parts, suggestionsStr)
	}

	return strings.Join(parts, "\n")
}

// formatJSON formats error as JSON
func (f *ErrorFormatter) formatJSON(err *PinError) string {
	data := err.ToJSON()
	jsonBytes, jsonErr := json.MarshalIndent(data, "", "  ")
	if jsonErr != nil {
		return fmt.Sprintf(`{"error": "failed to marshal error to JSON: %s"}`, jsonErr.Error())
	}
	return string(jsonBytes)
}

// formatPlain formats error as plain text without colors
func (f *ErrorFormatter) formatPlain(err *PinError) string {
	oldColorized := f.colorized
	f.colorized = false
	result := f.formatConsole(err)
	f.colorized = oldColorized
	return result
}

// buildHeader creates the error header with severity and job info
func (f *ErrorFormatter) buildHeader(err *PinError) string {
	var parts []string

	// Severity emoji and level
	severityInfo := f.getSeverityInfo(err.Severity)
	parts = append(parts, severityInfo)

	// Error code
	if err.Code != "" {
		codeStr := string(err.Code)
		if f.colorized {
			codeStr = color.New(color.FgCyan).Sprint(codeStr)
		}
		parts = append(parts, fmt.Sprintf("[%s]", codeStr))
	}

	// Job name
	if err.Job != "" {
		jobStr := err.Job
		if f.colorized {
			jobStr = color.New(color.FgYellow).Sprint(jobStr)
		}
		parts = append(parts, fmt.Sprintf("in job '%s'", jobStr))
	}

	// File and line info
	if err.File != "" {
		fileStr := err.File
		if f.colorized {
			fileStr = color.New(color.FgBlue).Sprint(fileStr)
		}
		if err.Line > 0 {
			parts = append(parts, fmt.Sprintf("at %s:%d", fileStr, err.Line))
		} else {
			parts = append(parts, fmt.Sprintf("in %s", fileStr))
		}
	}

	return strings.Join(parts, " ")
}

// getSeverityInfo returns emoji and color for severity level
func (f *ErrorFormatter) getSeverityInfo(severity Severity) string {
	switch severity {
	case SeverityError:
		if f.colorized {
			return color.New(color.FgRed).Sprint("❌ ERROR")
		}
		return "❌ ERROR"
	case SeverityWarning:
		if f.colorized {
			return color.New(color.FgYellow).Sprint("⚠️  WARNING")
		}
		return "⚠️  WARNING"
	case SeverityInfo:
		if f.colorized {
			return color.New(color.FgBlue).Sprint("ℹ️  INFO")
		}
		return "ℹ️  INFO"
	default:
		if f.colorized {
			return color.New(color.FgRed).Sprint("❌ ERROR")
		}
		return "❌ ERROR"
	}
}

// formatMessage formats the main error message
func (f *ErrorFormatter) formatMessage(message string) string {
	if f.colorized {
		return color.New(color.FgWhite, color.Bold).Sprintf("Message: %s", message)
	}
	return fmt.Sprintf("Message: %s", message)
}

// formatContext formats context information
func (f *ErrorFormatter) formatContext(context map[string]interface{}) string {
	var parts []string

	header := "Context:"
	if f.colorized {
		header = color.New(color.FgCyan, color.Bold).Sprint(header)
	}
	parts = append(parts, header)

	for key, value := range context {
		keyStr := key
		valueStr := fmt.Sprintf("%v", value)

		if f.colorized {
			keyStr = color.New(color.FgCyan).Sprint(keyStr)
			valueStr = color.New(color.FgWhite).Sprint(valueStr)
		}

		parts = append(parts, fmt.Sprintf("  %s: %s", keyStr, valueStr))
	}

	return strings.Join(parts, "\n")
}

// formatCause formats the underlying cause error
func (f *ErrorFormatter) formatCause(cause error) string {
	header := "Underlying cause:"
	causeStr := cause.Error()

	if f.colorized {
		header = color.New(color.FgMagenta, color.Bold).Sprint(header)
		causeStr = color.New(color.FgMagenta).Sprint(causeStr)
	}

	return fmt.Sprintf("%s\n  %s", header, causeStr)
}

// formatSuggestions formats error suggestions
func (f *ErrorFormatter) formatSuggestions(suggestions []string) string {
	var parts []string

	header := "💡 Suggestions:"
	if f.colorized {
		header = color.New(color.FgGreen, color.Bold).Sprint(header)
	}
	parts = append(parts, header)

	for i, suggestion := range suggestions {
		suggestionStr := suggestion
		if f.colorized {
			// Highlight code examples in suggestions
			if strings.Contains(suggestion, ":\n") || strings.Contains(suggestion, "Example") {
				suggestionStr = f.highlightCodeBlocks(suggestion)
			} else {
				suggestionStr = color.New(color.FgGreen).Sprint(suggestion)
			}
		}

		parts = append(parts, fmt.Sprintf("  %d. %s", i+1, suggestionStr))
	}

	return strings.Join(parts, "\n")
}

// highlightCodeBlocks highlights code examples in suggestions
func (f *ErrorFormatter) highlightCodeBlocks(text string) string {
	if !f.colorized {
		return text
	}

	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "-") ||
			strings.Contains(line, ":") && (strings.Contains(line, "image:") ||
				strings.Contains(line, "script:") || strings.Contains(line, "port:")) {
			// This looks like YAML, highlight it
			result = append(result, color.New(color.FgHiBlack).Sprint(line))
		} else if strings.HasPrefix(strings.TrimSpace(line), "docker ") ||
			strings.HasPrefix(strings.TrimSpace(line), "pin ") ||
			strings.HasPrefix(strings.TrimSpace(line), "curl ") {
			// This looks like a command, highlight it
			result = append(result, color.New(color.FgHiBlue).Sprint(line))
		} else {
			result = append(result, color.New(color.FgGreen).Sprint(line))
		}
	}

	return strings.Join(result, "\n")
}

// FormatMultiple formats multiple errors
func (f *ErrorFormatter) FormatMultiple(errors []*PinError) string {
	if len(errors) == 0 {
		return ""
	}

	if len(errors) == 1 {
		return f.Format(errors[0])
	}

	var parts []string

	header := fmt.Sprintf("Found %d errors:", len(errors))
	if f.colorized {
		header = color.New(color.FgRed, color.Bold).Sprint(header)
	}
	parts = append(parts, header)

	for i, err := range errors {
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("Error %d:", i+1))
		parts = append(parts, f.Format(err))
	}

	return strings.Join(parts, "\n")
}

// Default formatters
var (
	ConsoleFormatter = NewErrorFormatter(FormatConsole)
	JSONFormatter    = NewErrorFormatter(FormatJSON)
	PlainFormatter   = NewErrorFormatter(FormatPlain).WithColor(false)
)