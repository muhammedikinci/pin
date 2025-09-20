package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidationErrorBuilder_MissingImageOrDockerfile(t *testing.T) {
	builder := NewValidationErrorBuilder()
	err := builder.MissingImageOrDockerfile("test-job")

	if err.Code != ErrCodeMissingField {
		t.Errorf("Expected code %s, got %s", ErrCodeMissingField, err.Code)
	}

	if err.Job != "test-job" {
		t.Errorf("Expected job 'test-job', got '%s'", err.Job)
	}

	expectedMessage := "either 'image' or 'dockerfile' must be specified"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
	}

	// Check suggestions
	if len(err.Suggestions) == 0 {
		t.Error("Expected suggestions to be provided")
	}

	// Check that suggestions contain relevant information
	suggestions := strings.Join(err.Suggestions, " ")
	if !strings.Contains(suggestions, "image") || !strings.Contains(suggestions, "dockerfile") {
		t.Error("Expected suggestions to mention both 'image' and 'dockerfile'")
	}
}

func TestValidationErrorBuilder_InvalidPortFormat(t *testing.T) {
	builder := NewValidationErrorBuilder()
	err := builder.InvalidPortFormat("test-job", "invalid:port:format:here", 2)

	if err.Code != ErrCodeInvalidPortFormat {
		t.Errorf("Expected code %s, got %s", ErrCodeInvalidPortFormat, err.Code)
	}

	if err.Job != "test-job" {
		t.Errorf("Expected job 'test-job', got '%s'", err.Job)
	}

	// Check context
	if err.Context["port_value"] != "invalid:port:format:here" {
		t.Error("Expected port_value in context")
	}

	if err.Context["index"] != 2 {
		t.Error("Expected index in context")
	}

	// Check suggestions contain port format examples
	suggestions := strings.Join(err.Suggestions, " ")
	if !strings.Contains(suggestions, "8080:80") {
		t.Error("Expected suggestions to contain port format examples")
	}
}

func TestValidationErrorBuilder_EmptyScript(t *testing.T) {
	builder := NewValidationErrorBuilder()
	err := builder.EmptyScript("build-job")

	if err.Code != ErrCodeInvalidFieldValue {
		t.Errorf("Expected code %s, got %s", ErrCodeInvalidFieldValue, err.Code)
	}

	if err.Job != "build-job" {
		t.Errorf("Expected job 'build-job', got '%s'", err.Job)
	}

	// Check suggestions contain script examples
	suggestions := strings.Join(err.Suggestions, " ")
	if !strings.Contains(suggestions, "script:") {
		t.Error("Expected suggestions to contain script examples")
	}
}

func TestValidationErrorBuilder_InvalidRetryConfig(t *testing.T) {
	builder := NewValidationErrorBuilder()
	err := builder.InvalidRetryConfig("retry-job", "attempts", 15, "exceeds maximum value")

	if err.Code != ErrCodeInvalidRetryConfig {
		t.Errorf("Expected code %s, got %s", ErrCodeInvalidRetryConfig, err.Code)
	}

	if err.Job != "retry-job" {
		t.Errorf("Expected job 'retry-job', got '%s'", err.Job)
	}

	// Check context
	if err.Context["field"] != "attempts" {
		t.Error("Expected field in context")
	}

	if err.Context["value"] != 15 {
		t.Error("Expected value in context")
	}

	// Check message contains field name and reason
	if !strings.Contains(err.Message, "attempts") || !strings.Contains(err.Message, "exceeds maximum value") {
		t.Error("Expected message to contain field name and reason")
	}
}

func TestDockerErrorBuilder_ConnectionFailed(t *testing.T) {
	builder := NewDockerErrorBuilder()
	cause := fmt.Errorf("connection refused")
	err := builder.ConnectionFailed(cause)

	if err.Code != ErrCodeDockerConnection {
		t.Errorf("Expected code %s, got %s", ErrCodeDockerConnection, err.Code)
	}

	if err.Cause != cause {
		t.Error("Expected cause to be set")
	}

	// Check suggestions contain Docker troubleshooting
	suggestions := strings.Join(err.Suggestions, " ")
	if !strings.Contains(suggestions, "Docker") {
		t.Error("Expected suggestions to contain Docker troubleshooting")
	}

	if !strings.Contains(suggestions, "docker ps") {
		t.Error("Expected suggestions to contain 'docker ps' command")
	}
}

func TestDockerErrorBuilder_ImageNotFound(t *testing.T) {
	builder := NewDockerErrorBuilder()
	cause := fmt.Errorf("image not found")
	err := builder.ImageNotFound("test-job", "nonexistent:latest", cause)

	if err.Code != ErrCodeImageNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeImageNotFound, err.Code)
	}

	if err.Job != "test-job" {
		t.Errorf("Expected job 'test-job', got '%s'", err.Job)
	}

	if err.Context["image"] != "nonexistent:latest" {
		t.Error("Expected image in context")
	}

	if err.Cause != cause {
		t.Error("Expected cause to be set")
	}

	// Check suggestions contain image troubleshooting
	suggestions := strings.Join(err.Suggestions, " ")
	if !strings.Contains(suggestions, "docker pull") {
		t.Error("Expected suggestions to contain 'docker pull' command")
	}
}

func TestDockerErrorBuilder_ContainerFailed(t *testing.T) {
	builder := NewDockerErrorBuilder()
	cause := fmt.Errorf("container error")

	// Test exit code 127 (command not found)
	err := builder.ContainerFailed("test-job", 127, cause)

	if err.Code != ErrCodeContainerFailed {
		t.Errorf("Expected code %s, got %s", ErrCodeContainerFailed, err.Code)
	}

	if err.Job != "test-job" {
		t.Errorf("Expected job 'test-job', got '%s'", err.Job)
	}

	if err.Context["exit_code"] != 127 {
		t.Error("Expected exit_code in context")
	}

	// Check suggestions for exit code 127
	suggestions := strings.Join(err.Suggestions, " ")
	if !strings.Contains(suggestions, "Command not found") {
		t.Error("Expected suggestions for exit code 127 to mention 'Command not found'")
	}

	// Test exit code 126 (permission denied)
	err126 := builder.ContainerFailed("test-job", 126, cause)
	suggestions126 := strings.Join(err126.Suggestions, " ")
	if !strings.Contains(suggestions126, "Permission denied") {
		t.Error("Expected suggestions for exit code 126 to mention 'Permission denied'")
	}
}

func TestFileErrorBuilder_FileNotFound(t *testing.T) {
	builder := NewFileErrorBuilder()
	cause := fmt.Errorf("no such file")
	err := builder.FileNotFound("/path/to/pipeline.yaml", cause)

	if err.Code != ErrCodeFileNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeFileNotFound, err.Code)
	}

	if err.Context["file_path"] != "/path/to/pipeline.yaml" {
		t.Error("Expected file_path in context")
	}

	if err.Cause != cause {
		t.Error("Expected cause to be set")
	}

	// Check suggestions for YAML files
	suggestions := strings.Join(err.Suggestions, " ")
	if !strings.Contains(suggestions, "workflow:") {
		t.Error("Expected suggestions for YAML files to contain pipeline example")
	}

	// Test Dockerfile
	dockerErr := builder.FileNotFound("/path/to/Dockerfile", cause)
	dockerSuggestions := strings.Join(dockerErr.Suggestions, " ")
	if !strings.Contains(dockerSuggestions, "FROM") {
		t.Error("Expected suggestions for Dockerfile to contain Dockerfile example")
	}
}

func TestFileErrorBuilder_PermissionDenied(t *testing.T) {
	builder := NewFileErrorBuilder()
	cause := fmt.Errorf("permission denied")
	err := builder.PermissionDenied("/path/to/file", cause)

	if err.Code != ErrCodeFilePermission {
		t.Errorf("Expected code %s, got %s", ErrCodeFilePermission, err.Code)
	}

	if err.Context["file_path"] != "/path/to/file" {
		t.Error("Expected file_path in context")
	}

	if err.Cause != cause {
		t.Error("Expected cause to be set")
	}

	// Check suggestions contain chmod
	suggestions := strings.Join(err.Suggestions, " ")
	if !strings.Contains(suggestions, "chmod") {
		t.Error("Expected suggestions to contain chmod command")
	}
}

func TestNetworkErrorBuilder_PortInUse(t *testing.T) {
	builder := NewNetworkErrorBuilder()
	cause := fmt.Errorf("port already in use")
	err := builder.PortInUse("web-job", "8080:80", cause)

	if err.Code != ErrCodePortInUse {
		t.Errorf("Expected code %s, got %s", ErrCodePortInUse, err.Code)
	}

	if err.Job != "web-job" {
		t.Errorf("Expected job 'web-job', got '%s'", err.Job)
	}

	if err.Context["port"] != "8080:80" {
		t.Error("Expected port in context")
	}

	if err.Cause != cause {
		t.Error("Expected cause to be set")
	}

	// Check suggestions contain port troubleshooting
	suggestions := strings.Join(err.Suggestions, " ")
	if !strings.Contains(suggestions, "lsof") {
		t.Error("Expected suggestions to contain lsof command")
	}

	if !strings.Contains(suggestions, "8081") {
		t.Error("Expected suggestions to suggest alternative port")
	}
}

func TestErrorBuilders_ChainedCalls(t *testing.T) {
	// Test that builder methods return PinError allowing for chained calls
	builder := NewValidationErrorBuilder()
	err := builder.MissingImageOrDockerfile("test").
		WithContext("additional", "info").
		AddSuggestion("Additional suggestion")

	if err.Context["additional"] != "info" {
		t.Error("Expected chained context to be applied")
	}

	// Check that the additional suggestion was added
	found := false
	for _, suggestion := range err.Suggestions {
		if suggestion == "Additional suggestion" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected additional suggestion to be present")
	}
}

func TestOSDetectionHelpers(t *testing.T) {
	// These functions depend on environment, so we just test they don't panic
	// and return boolean values
	_ = isLinux()
	_ = isMacOS()
	_ = isWindows()
	_ = fileExists("/nonexistent/path")
}