package runner

import (
	"testing"
)

func TestGetRetryConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected RetryConfig
	}{
		{
			name:  "nil input returns default config",
			input: nil,
			expected: RetryConfig{
				MaxAttempts:       1,
				DelaySeconds:     1,
				BackoffMultiplier: 1.0,
			},
		},
		{
			name:  "invalid type returns default config",
			input: "invalid",
			expected: RetryConfig{
				MaxAttempts:       1,
				DelaySeconds:     1,
				BackoffMultiplier: 1.0,
			},
		},
		{
			name: "valid retry config with all fields",
			input: map[string]interface{}{
				"attempts": 5,
				"delay":    10,
				"backoff":  2.5,
			},
			expected: RetryConfig{
				MaxAttempts:       5,
				DelaySeconds:     10,
				BackoffMultiplier: 2.5,
			},
		},
		{
			name: "partial retry config uses defaults for missing fields",
			input: map[string]interface{}{
				"attempts": 3,
			},
			expected: RetryConfig{
				MaxAttempts:       3,
				DelaySeconds:     1,  // default
				BackoffMultiplier: 1.0, // default
			},
		},
		{
			name: "invalid attempts type uses default",
			input: map[string]interface{}{
				"attempts": "invalid",
				"delay":    5,
			},
			expected: RetryConfig{
				MaxAttempts:       1,   // default due to invalid type
				DelaySeconds:     5,
				BackoffMultiplier: 1.0, // default
			},
		},
		{
			name: "zero or negative attempts uses default",
			input: map[string]interface{}{
				"attempts": 0,
				"delay":    2,
			},
			expected: RetryConfig{
				MaxAttempts:       1,   // default due to invalid value
				DelaySeconds:     2,
				BackoffMultiplier: 1.0, // default
			},
		},
		{
			name: "invalid delay type uses default",
			input: map[string]interface{}{
				"attempts": 3,
				"delay":    "invalid",
				"backoff":  1.5,
			},
			expected: RetryConfig{
				MaxAttempts:       3,
				DelaySeconds:     1,   // default due to invalid type
				BackoffMultiplier: 1.5,
			},
		},
		{
			name: "negative delay uses default",
			input: map[string]interface{}{
				"attempts": 2,
				"delay":    -5,
			},
			expected: RetryConfig{
				MaxAttempts:       2,
				DelaySeconds:     1, // default due to invalid value
				BackoffMultiplier: 1.0, // default
			},
		},
		{
			name: "invalid backoff type uses default",
			input: map[string]interface{}{
				"attempts": 4,
				"delay":    3,
				"backoff":  "invalid",
			},
			expected: RetryConfig{
				MaxAttempts:       4,
				DelaySeconds:     3,
				BackoffMultiplier: 1.0, // default due to invalid type
			},
		},
		{
			name: "zero or negative backoff uses default",
			input: map[string]interface{}{
				"attempts": 2,
				"delay":    1,
				"backoff":  0.0,
			},
			expected: RetryConfig{
				MaxAttempts:       2,
				DelaySeconds:     1,
				BackoffMultiplier: 1.0, // default due to invalid value
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRetryConfig(tt.input)

			if result.MaxAttempts != tt.expected.MaxAttempts {
				t.Errorf("MaxAttempts = %d, expected %d", result.MaxAttempts, tt.expected.MaxAttempts)
			}
			if result.DelaySeconds != tt.expected.DelaySeconds {
				t.Errorf("DelaySeconds = %d, expected %d", result.DelaySeconds, tt.expected.DelaySeconds)
			}
			if result.BackoffMultiplier != tt.expected.BackoffMultiplier {
				t.Errorf("BackoffMultiplier = %f, expected %f", result.BackoffMultiplier, tt.expected.BackoffMultiplier)
			}
		})
	}
}

func TestGenerateJobWithRetryConfig(t *testing.T) {
	configMap := map[string]interface{}{
		"image": "alpine:latest",
		"retry": map[string]interface{}{
			"attempts": 5,
			"delay":    10,
			"backoff":  2.0,
		},
	}

	job, err := generateJob(configMap)
	if err != nil {
		t.Fatalf("generateJob failed: %v", err)
	}

	expectedRetry := RetryConfig{
		MaxAttempts:       5,
		DelaySeconds:     10,
		BackoffMultiplier: 2.0,
	}

	if job.RetryConfig.MaxAttempts != expectedRetry.MaxAttempts {
		t.Errorf("RetryConfig.MaxAttempts = %d, expected %d", job.RetryConfig.MaxAttempts, expectedRetry.MaxAttempts)
	}
	if job.RetryConfig.DelaySeconds != expectedRetry.DelaySeconds {
		t.Errorf("RetryConfig.DelaySeconds = %d, expected %d", job.RetryConfig.DelaySeconds, expectedRetry.DelaySeconds)
	}
	if job.RetryConfig.BackoffMultiplier != expectedRetry.BackoffMultiplier {
		t.Errorf("RetryConfig.BackoffMultiplier = %f, expected %f", job.RetryConfig.BackoffMultiplier, expectedRetry.BackoffMultiplier)
	}
}

func TestGenerateJobWithoutRetryConfig(t *testing.T) {
	configMap := map[string]interface{}{
		"image": "alpine:latest",
	}

	job, err := generateJob(configMap)
	if err != nil {
		t.Fatalf("generateJob failed: %v", err)
	}

	// Should use default retry config
	expectedRetry := RetryConfig{
		MaxAttempts:       1,
		DelaySeconds:     1,
		BackoffMultiplier: 1.0,
	}

	if job.RetryConfig.MaxAttempts != expectedRetry.MaxAttempts {
		t.Errorf("RetryConfig.MaxAttempts = %d, expected %d", job.RetryConfig.MaxAttempts, expectedRetry.MaxAttempts)
	}
	if job.RetryConfig.DelaySeconds != expectedRetry.DelaySeconds {
		t.Errorf("RetryConfig.DelaySeconds = %d, expected %d", job.RetryConfig.DelaySeconds, expectedRetry.DelaySeconds)
	}
	if job.RetryConfig.BackoffMultiplier != expectedRetry.BackoffMultiplier {
		t.Errorf("RetryConfig.BackoffMultiplier = %f, expected %f", job.RetryConfig.BackoffMultiplier, expectedRetry.BackoffMultiplier)
	}
}