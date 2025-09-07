package runner

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// PipelineValidator handles validation of pipeline YAML configuration
type PipelineValidator struct{}

// NewPipelineValidator creates a new pipeline validator
func NewPipelineValidator() *PipelineValidator {
	return &PipelineValidator{}
}

// ValidatePipeline validates the entire pipeline configuration
func (v *PipelineValidator) ValidatePipeline() error {
	// Check if workflow is defined
	workflows := viper.GetStringSlice("workflow")
	if len(workflows) == 0 {
		return errors.New("workflow must be defined and cannot be empty")
	}

	// Validate docker host configuration if present
	if err := v.validateDockerHost(); err != nil {
		return err
	}

	// Validate each job in the workflow
	for _, jobName := range workflows {
		if err := v.validateJob(jobName); err != nil {
			return fmt.Errorf("validation error in job '%s': %w", jobName, err)
		}
	}

	return nil
}

// validateJob validates a single job configuration
func (v *PipelineValidator) validateJob(jobName string) error {
	configMap := viper.GetStringMap(jobName)
	if len(configMap) == 0 {
		return fmt.Errorf("job '%s' is not defined or is empty", jobName)
	}

	// Validate required fields
	if err := v.validateImageOrDockerfile(configMap); err != nil {
		return err
	}

	// Validate optional fields if present
	if err := v.validateScript(configMap); err != nil {
		return err
	}

	if err := v.validatePorts(configMap); err != nil {
		return err
	}

	if err := v.validateEnvironmentVariables(configMap); err != nil {
		return err
	}

	if err := v.validateCopyIgnore(configMap); err != nil {
		return err
	}

	if err := v.validateWorkDir(configMap); err != nil {
		return err
	}

	if err := v.validateArtifactPath(configMap); err != nil {
		return err
	}

	if err := v.validateCondition(configMap); err != nil {
		return err
	}

	// Validate boolean fields
	if err := v.validateBooleanFields(configMap); err != nil {
		return err
	}

	if err := v.validateRetryConfig(configMap); err != nil {
		return err
	}

	return nil
}

// validateImageOrDockerfile ensures either image or dockerfile is specified
func (v *PipelineValidator) validateImageOrDockerfile(configMap map[string]interface{}) error {
	image := configMap["image"]
	dockerfile := configMap["dockerfile"]

	if image == nil && dockerfile == nil {
		return errors.New("either 'image' or 'dockerfile' must be specified")
	}

	if image != nil && dockerfile != nil {
		return errors.New("cannot specify both 'image' and 'dockerfile' in the same job")
	}

	if image != nil {
		if _, ok := image.(string); !ok {
			return errors.New("'image' must be a string")
		}
		imageStr := image.(string)
		if strings.TrimSpace(imageStr) == "" {
			return errors.New("'image' cannot be empty")
		}
	}

	if dockerfile != nil {
		if _, ok := dockerfile.(string); !ok {
			return errors.New("'dockerfile' must be a string")
		}
		dockerfileStr := dockerfile.(string)
		if strings.TrimSpace(dockerfileStr) == "" {
			return errors.New("'dockerfile' cannot be empty")
		}
	}

	return nil
}

// validateScript validates the script field
func (v *PipelineValidator) validateScript(configMap map[string]interface{}) error {
	script := configMap["script"]
	if script == nil {
		return nil // script is optional
	}

	refVal := reflect.ValueOf(script)
	
	if refVal.Kind() == reflect.Slice {
		if refVal.Len() == 0 {
			return errors.New("'script' array cannot be empty")
		}
		
		for i := 0; i < refVal.Len(); i++ {
			item := refVal.Index(i).Interface()
			if _, ok := item.(string); !ok {
				return fmt.Errorf("all script items must be strings, found %T at index %d", item, i)
			}
			if strings.TrimSpace(item.(string)) == "" {
				return fmt.Errorf("script item at index %d cannot be empty", i)
			}
		}
	} else if refVal.Kind() == reflect.String {
		if strings.TrimSpace(script.(string)) == "" {
			return errors.New("'script' cannot be empty")
		}
	} else {
		return errors.New("'script' must be a string or array of strings")
	}

	return nil
}

// validatePorts validates the port field
func (v *PipelineValidator) validatePorts(configMap map[string]interface{}) error {
	port := configMap["port"]
	if port == nil {
		return nil // port is optional
	}

	refVal := reflect.ValueOf(port)
	
	if refVal.Kind() == reflect.Slice {
		for i := 0; i < refVal.Len(); i++ {
			item := refVal.Index(i).Interface()
			if _, ok := item.(string); !ok {
				return fmt.Errorf("all port items must be strings, found %T at index %d", item, i)
			}
			if err := v.validatePortFormat(item.(string)); err != nil {
				return fmt.Errorf("invalid port format at index %d: %w", i, err)
			}
		}
	} else if refVal.Kind() == reflect.String {
		if err := v.validatePortFormat(port.(string)); err != nil {
			return fmt.Errorf("invalid port format: %w", err)
		}
	} else {
		return errors.New("'port' must be a string or array of strings")
	}

	return nil
}

// validatePortFormat validates a single port format
// Supports formats: "8080:80" or "127.0.0.1:8080:80" or "localhost:8080:80"
func (v *PipelineValidator) validatePortFormat(portStr string) error {
	parts := strings.Split(portStr, ":")
	
	switch len(parts) {
	case 2:
		// Format: "hostPort:containerPort" (e.g., "8080:80")
		if strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return errors.New("both host and container ports must be specified")
		}
	case 3:
		// Format: "hostIP:hostPort:containerPort" (e.g., "127.0.0.1:8080:80")
		if strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" || strings.TrimSpace(parts[2]) == "" {
			return errors.New("host IP, host port, and container port must all be specified")
		}
		// Basic IP/hostname validation
		hostIP := strings.TrimSpace(parts[0])
		if hostIP == "" {
			return errors.New("host IP cannot be empty")
		}
	default:
		return errors.New("port must be in format 'hostPort:containerPort' (e.g., '8080:80') or 'hostIP:hostPort:containerPort' (e.g., '127.0.0.1:8080:80')")
	}

	return nil
}

// validateEnvironmentVariables validates the env field
func (v *PipelineValidator) validateEnvironmentVariables(configMap map[string]interface{}) error {
	env := configMap["env"]
	if env == nil {
		return nil // env is optional
	}

	refVal := reflect.ValueOf(env)
	
	if refVal.Kind() == reflect.Slice {
		for i := 0; i < refVal.Len(); i++ {
			item := refVal.Index(i).Interface()
			if _, ok := item.(string); !ok {
				return fmt.Errorf("all environment variables must be strings, found %T at index %d", item, i)
			}
			if strings.TrimSpace(item.(string)) == "" {
				return fmt.Errorf("environment variable at index %d cannot be empty", i)
			}
		}
	} else if refVal.Kind() == reflect.String {
		if strings.TrimSpace(env.(string)) == "" {
			return errors.New("environment variable cannot be empty")
		}
	} else {
		return errors.New("'env' must be a string or array of strings")
	}

	return nil
}

// validateCopyIgnore validates the copyignore field
func (v *PipelineValidator) validateCopyIgnore(configMap map[string]interface{}) error {
	copyIgnore := configMap["copyignore"]
	if copyIgnore == nil {
		return nil // copyignore is optional
	}

	refVal := reflect.ValueOf(copyIgnore)
	
	if refVal.Kind() == reflect.Slice {
		for i := 0; i < refVal.Len(); i++ {
			item := refVal.Index(i).Interface()
			if _, ok := item.(string); !ok {
				return fmt.Errorf("all copyignore items must be strings, found %T at index %d", item, i)
			}
			if strings.TrimSpace(item.(string)) == "" {
				return fmt.Errorf("copyignore item at index %d cannot be empty", i)
			}
		}
	} else if refVal.Kind() == reflect.String {
		if strings.TrimSpace(copyIgnore.(string)) == "" {
			return errors.New("copyignore cannot be empty")
		}
	} else {
		return errors.New("'copyignore' must be a string or array of strings")
	}

	return nil
}

// validateWorkDir validates the workdir field
func (v *PipelineValidator) validateWorkDir(configMap map[string]interface{}) error {
	workDir := configMap["workdir"]
	if workDir == nil {
		return nil // workdir is optional
	}

	if _, ok := workDir.(string); !ok {
		return errors.New("'workdir' must be a string")
	}

	workDirStr := workDir.(string)
	if strings.TrimSpace(workDirStr) == "" {
		return errors.New("'workdir' cannot be empty")
	}

	return nil
}

// validateArtifactPath validates the artifactpath field
func (v *PipelineValidator) validateArtifactPath(configMap map[string]interface{}) error {
	artifactPath := configMap["artifactpath"]
	if artifactPath == nil {
		return nil // artifactpath is optional
	}

	if _, ok := artifactPath.(string); !ok {
		return errors.New("'artifactpath' must be a string")
	}

	artifactPathStr := artifactPath.(string)
	if strings.TrimSpace(artifactPathStr) == "" {
		return errors.New("'artifactpath' cannot be empty")
	}

	return nil
}

// validateCondition validates the condition field
func (v *PipelineValidator) validateCondition(configMap map[string]interface{}) error {
	condition := configMap["condition"]
	if condition == nil {
		return nil // condition is optional
	}

	if _, ok := condition.(string); !ok {
		return errors.New("'condition' must be a string")
	}

	conditionStr := condition.(string)
	if strings.TrimSpace(conditionStr) == "" {
		return errors.New("'condition' cannot be empty")
	}

	return nil
}

// validateBooleanFields validates boolean fields
func (v *PipelineValidator) validateBooleanFields(configMap map[string]interface{}) error {
	boolFields := []string{"copyfiles", "soloexecution", "parallel"}
	
	for _, field := range boolFields {
		value := configMap[field]
		if value != nil {
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("'%s' must be a boolean value", field)
			}
		}
	}

	return nil
}

// validateDockerHost validates the docker host configuration
func (v *PipelineValidator) validateDockerHost() error {
	dockerHost := viper.GetString("docker.host")
	if dockerHost == "" {
		return nil // docker.host is optional
	}

	// Basic docker host format validation
	dockerHost = strings.TrimSpace(dockerHost)
	if dockerHost == "" {
		return errors.New("docker.host cannot be empty")
	}

	// Check for common docker host formats
	validPrefixes := []string{
		"tcp://",
		"unix://",
		"npipe://",
		"ssh://",
	}

	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(dockerHost, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return fmt.Errorf("docker.host must start with a valid protocol (tcp://, unix://, npipe://, ssh://), got: %s", dockerHost)
	}

	// Additional validation for tcp:// format
	if strings.HasPrefix(dockerHost, "tcp://") {
		// Remove tcp:// prefix for validation
		hostPart := strings.TrimPrefix(dockerHost, "tcp://")
		if hostPart == "" {
			return errors.New("docker.host tcp:// format requires host and port (e.g., tcp://localhost:2375)")
		}
		
		// Check if it contains port
		if !strings.Contains(hostPart, ":") {
			return errors.New("docker.host tcp:// format must include port (e.g., tcp://localhost:2375)")
		}
	}

	return nil
}

// validateRetryConfig validates the retry configuration
func (v *PipelineValidator) validateRetryConfig(configMap map[string]interface{}) error {
	retry := configMap["retry"]
	if retry == nil {
		return nil // retry is optional
	}

	retryMap, ok := retry.(map[string]interface{})
	if !ok {
		return errors.New("'retry' must be an object")
	}

	// Validate attempts
	if attempts := retryMap["attempts"]; attempts != nil {
		if attemptsInt, ok := attempts.(int); ok {
			if attemptsInt < 1 {
				return errors.New("retry.attempts must be at least 1")
			}
			if attemptsInt > 10 {
				return errors.New("retry.attempts must not exceed 10 (to prevent infinite loops)")
			}
		} else {
			return errors.New("retry.attempts must be an integer")
		}
	}

	// Validate delay
	if delay := retryMap["delay"]; delay != nil {
		if delayInt, ok := delay.(int); ok {
			if delayInt < 0 {
				return errors.New("retry.delay must be non-negative")
			}
			if delayInt > 300 {
				return errors.New("retry.delay must not exceed 300 seconds")
			}
		} else {
			return errors.New("retry.delay must be an integer (seconds)")
		}
	}

	// Validate backoff multiplier
	if backoff := retryMap["backoff"]; backoff != nil {
		if backoffFloat, ok := backoff.(float64); ok {
			if backoffFloat <= 0 {
				return errors.New("retry.backoff must be greater than 0")
			}
			if backoffFloat > 10.0 {
				return errors.New("retry.backoff must not exceed 10.0")
			}
		} else {
			return errors.New("retry.backoff must be a number")
		}
	}

	return nil
}