package runner

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
)

// Integration tests for retry functionality with real YAML parsing

func TestRetryIntegration_ValidConfig(t *testing.T) {
	yamlConfig := `
workflow:
  - test-job

test-job:
  image: alpine:latest
  retry:
    attempts: 3
    delay: 2
    backoff: 1.5
  script:
    - echo "integration test"
`

	// Setup viper to read from the YAML string
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer([]byte(yamlConfig)))
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Test pipeline parsing
	pipeline, err := parse()
	if err != nil {
		t.Fatalf("Failed to parse pipeline: %v", err)
	}

	// Verify pipeline structure
	if len(pipeline.Workflow) != 1 {
		t.Errorf("Expected 1 job, got %d", len(pipeline.Workflow))
	}

	job := pipeline.Workflow[0]
	if job.Name != "test-job" {
		t.Errorf("Expected job name 'test-job', got '%s'", job.Name)
	}

	// Verify retry configuration
	expectedRetry := RetryConfig{
		MaxAttempts:       3,
		DelaySeconds:     2,
		BackoffMultiplier: 1.5,
	}

	if job.RetryConfig.MaxAttempts != expectedRetry.MaxAttempts {
		t.Errorf("Expected MaxAttempts=%d, got %d", expectedRetry.MaxAttempts, job.RetryConfig.MaxAttempts)
	}
	if job.RetryConfig.DelaySeconds != expectedRetry.DelaySeconds {
		t.Errorf("Expected DelaySeconds=%d, got %d", expectedRetry.DelaySeconds, job.RetryConfig.DelaySeconds)
	}
	if job.RetryConfig.BackoffMultiplier != expectedRetry.BackoffMultiplier {
		t.Errorf("Expected BackoffMultiplier=%f, got %f", expectedRetry.BackoffMultiplier, job.RetryConfig.BackoffMultiplier)
	}
}

func TestRetryIntegration_NoRetryConfig(t *testing.T) {
	yamlConfig := `
workflow:
  - simple-job

simple-job:
  image: alpine:latest
  script:
    - echo "no retry"
`

	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer([]byte(yamlConfig)))
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	pipeline, err := parse()
	if err != nil {
		t.Fatalf("Failed to parse pipeline: %v", err)
	}

	job := pipeline.Workflow[0]
	
	// Should use default retry config
	defaultRetry := RetryConfig{
		MaxAttempts:       1,
		DelaySeconds:     1,
		BackoffMultiplier: 1.0,
	}

	if job.RetryConfig.MaxAttempts != defaultRetry.MaxAttempts {
		t.Errorf("Expected default MaxAttempts=%d, got %d", defaultRetry.MaxAttempts, job.RetryConfig.MaxAttempts)
	}
	if job.RetryConfig.DelaySeconds != defaultRetry.DelaySeconds {
		t.Errorf("Expected default DelaySeconds=%d, got %d", defaultRetry.DelaySeconds, job.RetryConfig.DelaySeconds)
	}
	if job.RetryConfig.BackoffMultiplier != defaultRetry.BackoffMultiplier {
		t.Errorf("Expected default BackoffMultiplier=%f, got %f", defaultRetry.BackoffMultiplier, job.RetryConfig.BackoffMultiplier)
	}
}

func TestRetryIntegration_PartialConfig(t *testing.T) {
	yamlConfig := `
workflow:
  - partial-job

partial-job:
  image: alpine:latest
  retry:
    attempts: 5
    # delay and backoff should use defaults
  script:
    - echo "partial retry config"
`

	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer([]byte(yamlConfig)))
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	pipeline, err := parse()
	if err != nil {
		t.Fatalf("Failed to parse pipeline: %v", err)
	}

	job := pipeline.Workflow[0]

	// Check that specified value is used
	if job.RetryConfig.MaxAttempts != 5 {
		t.Errorf("Expected MaxAttempts=5, got %d", job.RetryConfig.MaxAttempts)
	}

	// Check that defaults are used for unspecified values
	if job.RetryConfig.DelaySeconds != 1 {
		t.Errorf("Expected default DelaySeconds=1, got %d", job.RetryConfig.DelaySeconds)
	}
	if job.RetryConfig.BackoffMultiplier != 1.0 {
		t.Errorf("Expected default BackoffMultiplier=1.0, got %f", job.RetryConfig.BackoffMultiplier)
	}
}

func TestRetryIntegration_ValidationWithPipeline(t *testing.T) {
	invalidConfigs := []struct {
		name   string
		config string
	}{
		{
			name: "invalid_attempts",
			config: `
workflow:
  - invalid-job

invalid-job:
  image: alpine:latest
  retry:
    attempts: 15
`,
		},
		{
			name: "invalid_delay",
			config: `
workflow:
  - invalid-job

invalid-job:
  image: alpine:latest
  retry:
    delay: 400
`,
		},
		{
			name: "invalid_backoff",
			config: `
workflow:
  - invalid-job

invalid-job:
  image: alpine:latest
  retry:
    backoff: -1.0
`,
		},
	}

	validator := NewPipelineValidator()

	for _, tt := range invalidConfigs {
		t.Run(tt.name, func(t *testing.T) {
			viper.SetConfigType("yaml")
			err := viper.ReadConfig(bytes.NewBuffer([]byte(tt.config)))
			if err != nil {
				t.Fatalf("Failed to read config: %v", err)
			}

			// Test that validation catches the error
			err = validator.ValidatePipeline()
			if err == nil {
				t.Errorf("Expected validation error for %s, but got none", tt.name)
			}
		})
	}
}

func TestRetryIntegration_MultipleJobsWithDifferentRetries(t *testing.T) {
	yamlConfig := `
workflow:
  - no-retry-job
  - simple-retry-job
  - advanced-retry-job

no-retry-job:
  image: alpine:latest
  script:
    - echo "no retry"

simple-retry-job:
  image: alpine:latest
  retry:
    attempts: 3
  script:
    - echo "simple retry"

advanced-retry-job:
  image: alpine:latest
  retry:
    attempts: 5
    delay: 10
    backoff: 2.0
  script:
    - echo "advanced retry"
`

	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer([]byte(yamlConfig)))
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	pipeline, err := parse()
	if err != nil {
		t.Fatalf("Failed to parse pipeline: %v", err)
	}

	if len(pipeline.Workflow) != 3 {
		t.Errorf("Expected 3 jobs, got %d", len(pipeline.Workflow))
	}

	// Verify no-retry-job (uses defaults)
	noRetryJob := pipeline.Workflow[0]
	if noRetryJob.RetryConfig.MaxAttempts != 1 {
		t.Errorf("No-retry job should have MaxAttempts=1, got %d", noRetryJob.RetryConfig.MaxAttempts)
	}

	// Verify simple-retry-job
	simpleRetryJob := pipeline.Workflow[1]
	if simpleRetryJob.RetryConfig.MaxAttempts != 3 {
		t.Errorf("Simple retry job should have MaxAttempts=3, got %d", simpleRetryJob.RetryConfig.MaxAttempts)
	}
	if simpleRetryJob.RetryConfig.DelaySeconds != 1 { // should use default
		t.Errorf("Simple retry job should have default DelaySeconds=1, got %d", simpleRetryJob.RetryConfig.DelaySeconds)
	}

	// Verify advanced-retry-job
	advancedRetryJob := pipeline.Workflow[2]
	if advancedRetryJob.RetryConfig.MaxAttempts != 5 {
		t.Errorf("Advanced retry job should have MaxAttempts=5, got %d", advancedRetryJob.RetryConfig.MaxAttempts)
	}
	if advancedRetryJob.RetryConfig.DelaySeconds != 10 {
		t.Errorf("Advanced retry job should have DelaySeconds=10, got %d", advancedRetryJob.RetryConfig.DelaySeconds)
	}
	if advancedRetryJob.RetryConfig.BackoffMultiplier != 2.0 {
		t.Errorf("Advanced retry job should have BackoffMultiplier=2.0, got %f", advancedRetryJob.RetryConfig.BackoffMultiplier)
	}
}

func TestRetryIntegration_EdgeCases(t *testing.T) {
	edgeCases := []struct {
		name         string
		config       string
		shouldError  bool
		expectValues RetryConfig
	}{
		{
			name: "minimum_valid_values",
			config: `
workflow:
  - min-job

min-job:
  image: alpine:latest
  retry:
    attempts: 1
    delay: 0
    backoff: 0.1
`,
			shouldError: false,
			expectValues: RetryConfig{
				MaxAttempts:       1,
				DelaySeconds:     0,
				BackoffMultiplier: 0.1,
			},
		},
		{
			name: "maximum_valid_values",
			config: `
workflow:
  - max-job

max-job:
  image: alpine:latest
  retry:
    attempts: 10
    delay: 300
    backoff: 10.0
`,
			shouldError: false,
			expectValues: RetryConfig{
				MaxAttempts:       10,
				DelaySeconds:     300,
				BackoffMultiplier: 10.0,
			},
		},
		{
			name: "empty_retry_object",
			config: `
workflow:
  - empty-job

empty-job:
  image: alpine:latest
  retry: {}
`,
			shouldError: false,
			expectValues: RetryConfig{
				MaxAttempts:       1,   // default
				DelaySeconds:     1,   // default
				BackoffMultiplier: 1.0, // default
			},
		},
	}

	validator := NewPipelineValidator()

	for _, tt := range edgeCases {
		t.Run(tt.name, func(t *testing.T) {
			viper.SetConfigType("yaml")
			err := viper.ReadConfig(bytes.NewBuffer([]byte(tt.config)))
			if err != nil {
				t.Fatalf("Failed to read config: %v", err)
			}

			// Test validation
			err = validator.ValidatePipeline()
			if tt.shouldError && err == nil {
				t.Errorf("Expected validation error but got none")
				return
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no validation error but got: %v", err)
				return
			}

			if !tt.shouldError {
				// Test parsing
				pipeline, err := parse()
				if err != nil {
					t.Fatalf("Failed to parse pipeline: %v", err)
				}

				job := pipeline.Workflow[0]
				if job.RetryConfig.MaxAttempts != tt.expectValues.MaxAttempts {
					t.Errorf("Expected MaxAttempts=%d, got %d", tt.expectValues.MaxAttempts, job.RetryConfig.MaxAttempts)
				}
				if job.RetryConfig.DelaySeconds != tt.expectValues.DelaySeconds {
					t.Errorf("Expected DelaySeconds=%d, got %d", tt.expectValues.DelaySeconds, job.RetryConfig.DelaySeconds)
				}
				if job.RetryConfig.BackoffMultiplier != tt.expectValues.BackoffMultiplier {
					t.Errorf("Expected BackoffMultiplier=%f, got %f", tt.expectValues.BackoffMultiplier, job.RetryConfig.BackoffMultiplier)
				}
			}
		})
	}
}