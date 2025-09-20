package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidationErrorBuilder helps build validation errors with context and suggestions
type ValidationErrorBuilder struct{}

// NewValidationErrorBuilder creates a new validation error builder
func NewValidationErrorBuilder() *ValidationErrorBuilder {
	return &ValidationErrorBuilder{}
}

// MissingImageOrDockerfile creates an error for missing image/dockerfile
func (b *ValidationErrorBuilder) MissingImageOrDockerfile(job string) *PinError {
	return NewPinError(ErrCodeMissingField, "either 'image' or 'dockerfile' must be specified").
		WithJob(job).
		AddSuggestions(
			"Add an 'image' field with a Docker image name (e.g., 'alpine:latest')",
			"Add a 'dockerfile' field with path to your Dockerfile (e.g., './Dockerfile')",
			"Example with image:\n  "+job+":\n    image: golang:1.21-alpine\n    script:\n      - go build .",
			"Example with dockerfile:\n  "+job+":\n    dockerfile: ./Dockerfile\n    script:\n      - echo 'Using custom image'",
		)
}

// InvalidPortFormat creates an error for invalid port configuration
func (b *ValidationErrorBuilder) InvalidPortFormat(job string, portValue string, index int) *PinError {
	err := NewPinError(ErrCodeInvalidPortFormat, "invalid port format").
		WithJob(job).
		WithContext("port_value", portValue).
		WithContext("index", index).
		AddSuggestions(
			"Use format 'hostPort:containerPort' (e.g., '8080:80')",
			"Use format 'hostIP:hostPort:containerPort' (e.g., '127.0.0.1:8080:80')",
			"For localhost only: '127.0.0.1:3000:3000'",
			"For all interfaces: '3000:3000'",
		)

	if index >= 0 {
		err.WithContext("error_location", fmt.Sprintf("port array index %d", index))
	}

	return err
}

// EmptyScript creates an error for empty script configuration
func (b *ValidationErrorBuilder) EmptyScript(job string) *PinError {
	return NewPinError(ErrCodeInvalidFieldValue, "script cannot be empty").
		WithJob(job).
		AddSuggestions(
			"Add commands to the script array",
			"Example:\n  script:\n    - echo 'Hello World'\n    - ls -la",
			"For multiple commands:\n  script:\n    - npm install\n    - npm test\n    - npm run build",
		)
}

// InvalidRetryConfig creates an error for invalid retry configuration
func (b *ValidationErrorBuilder) InvalidRetryConfig(job string, field string, value interface{}, reason string) *PinError {
	return NewPinError(ErrCodeInvalidRetryConfig, fmt.Sprintf("invalid retry.%s: %s", field, reason)).
		WithJob(job).
		WithContext("field", field).
		WithContext("value", value).
		AddSuggestions(
			"retry.attempts: 1-10 (default: 1)",
			"retry.delay: 0-300 seconds (default: 1)",
			"retry.backoff: 0.1-10.0 multiplier (default: 1.0)",
			"Example:\n  retry:\n    attempts: 3\n    delay: 2\n    backoff: 2.0",
		)
}

// DockerErrorBuilder helps build Docker-related errors
type DockerErrorBuilder struct{}

// NewDockerErrorBuilder creates a new Docker error builder
func NewDockerErrorBuilder() *DockerErrorBuilder {
	return &DockerErrorBuilder{}
}

// ConnectionFailed creates an error for Docker connection failures
func (b *DockerErrorBuilder) ConnectionFailed(err error) *PinError {
	suggestions := []string{
		"Check if Docker is running: 'docker ps'",
		"Start Docker service if stopped",
	}

	// Add platform-specific suggestions
	if isLinux() {
		suggestions = append(suggestions,
			"Linux: 'sudo systemctl start docker'",
			"Check Docker permissions: 'sudo usermod -aG docker $USER'",
			"Re-login after adding user to docker group",
		)
	} else if isMacOS() {
		suggestions = append(suggestions,
			"macOS: Open Docker Desktop application",
			"Check Docker Desktop status in menu bar",
		)
	} else if isWindows() {
		suggestions = append(suggestions,
			"Windows: Start Docker Desktop",
			"Check if WSL2 is properly configured",
		)
	}

	suggestions = append(suggestions,
		"Verify Docker socket permissions: 'ls -la /var/run/docker.sock'",
		"Try custom Docker host in config:\n  docker:\n    host: tcp://localhost:2375",
	)

	return NewPinError(ErrCodeDockerConnection, "failed to connect to Docker daemon").
		WithCause(err).
		AddSuggestions(suggestions...)
}

// ImageNotFound creates an error for missing Docker images
func (b *DockerErrorBuilder) ImageNotFound(job string, imageName string, err error) *PinError {
	return NewPinError(ErrCodeImageNotFound, fmt.Sprintf("Docker image '%s' not found", imageName)).
		WithJob(job).
		WithContext("image", imageName).
		WithCause(err).
		AddSuggestions(
			"Check if image name is correct",
			"Pull the image manually: 'docker pull "+imageName+"'",
			"Use a different image that exists locally",
			"Check Docker Hub for available tags: https://hub.docker.com",
			"For private images, ensure you're logged in: 'docker login'",
		)
}

// ContainerFailed creates an error for container execution failures
func (b *DockerErrorBuilder) ContainerFailed(job string, exitCode int, err error) *PinError {
	suggestions := []string{
		"Check the script commands for syntax errors",
		"Verify all required files are copied to container",
		"Use 'soloExecution: true' to run commands separately for easier debugging",
		"Add debug output to your script:\n  script:\n    - echo 'Debug: Starting step'\n    - your-command",
	}

	if exitCode == 127 {
		suggestions = append(suggestions,
			"Exit code 127: Command not found",
			"Install missing command in container or use different image",
			"Check if the command exists: 'which your-command'",
		)
	} else if exitCode == 126 {
		suggestions = append(suggestions,
			"Exit code 126: Permission denied",
			"Make script executable: 'chmod +x script.sh'",
			"Check file permissions in container",
		)
	}

	return NewPinError(ErrCodeContainerFailed, fmt.Sprintf("container execution failed with exit code %d", exitCode)).
		WithJob(job).
		WithContext("exit_code", exitCode).
		WithCause(err).
		AddSuggestions(suggestions...)
}

// FileErrorBuilder helps build file-related errors
type FileErrorBuilder struct{}

// NewFileErrorBuilder creates a new file error builder
func NewFileErrorBuilder() *FileErrorBuilder {
	return &FileErrorBuilder{}
}

// FileNotFound creates an error for missing files
func (b *FileErrorBuilder) FileNotFound(filePath string, err error) *PinError {
	suggestions := []string{
		"Check if the file path is correct",
		"Verify the file exists: 'ls -la " + filepath.Dir(filePath) + "'",
		"Use absolute path instead of relative path",
	}

	if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		suggestions = append(suggestions,
			"Create a pipeline configuration file:",
			"Example pipeline.yaml:\n  workflow:\n    - hello\n  hello:\n    image: alpine:latest\n    script:\n      - echo 'Hello World'",
		)
	}

	if strings.Contains(filePath, "Dockerfile") {
		suggestions = append(suggestions,
			"Create a Dockerfile:",
			"Example Dockerfile:\n  FROM alpine:latest\n  RUN apk add --no-cache curl\n  WORKDIR /app",
		)
	}

	return NewPinError(ErrCodeFileNotFound, fmt.Sprintf("file not found: %s", filePath)).
		WithContext("file_path", filePath).
		WithCause(err).
		AddSuggestions(suggestions...)
}

// PermissionDenied creates an error for file permission issues
func (b *FileErrorBuilder) PermissionDenied(filePath string, err error) *PinError {
	return NewPinError(ErrCodeFilePermission, fmt.Sprintf("permission denied: %s", filePath)).
		WithContext("file_path", filePath).
		WithCause(err).
		AddSuggestions(
			"Check file permissions: 'ls -la "+filePath+"'",
			"Make file readable: 'chmod 644 "+filePath+"'",
			"Check directory permissions: 'ls -la "+filepath.Dir(filePath)+"'",
			"Run with appropriate permissions or change file ownership",
		)
}

// NetworkErrorBuilder helps build network-related errors
type NetworkErrorBuilder struct{}

// NewNetworkErrorBuilder creates a new network error builder
func NewNetworkErrorBuilder() *NetworkErrorBuilder {
	return &NetworkErrorBuilder{}
}

// PortInUse creates an error for port conflicts
func (b *NetworkErrorBuilder) PortInUse(job string, port string, err error) *PinError {
	return NewPinError(ErrCodePortInUse, fmt.Sprintf("port %s is already in use", port)).
		WithJob(job).
		WithContext("port", port).
		WithCause(err).
		AddSuggestions(
			"Use a different port: change '"+port+":80' to '8081:80'",
			"Stop the process using the port: 'lsof -ti:"+strings.Split(port, ":")[0]+" | xargs kill'",
			"Find what's using the port: 'lsof -i :"+strings.Split(port, ":")[0]+"'",
			"Bind to localhost only: '127.0.0.1:"+port+"'",
		)
}

// Helper functions for OS detection
func isLinux() bool {
	return strings.Contains(strings.ToLower(os.Getenv("GOOS")), "linux") ||
		fileExists("/etc/os-release")
}

func isMacOS() bool {
	return strings.Contains(strings.ToLower(os.Getenv("GOOS")), "darwin") ||
		fileExists("/System/Library/CoreServices/SystemVersion.plist")
}

func isWindows() bool {
	return strings.Contains(strings.ToLower(os.Getenv("GOOS")), "windows") ||
		os.Getenv("OS") == "Windows_NT"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}