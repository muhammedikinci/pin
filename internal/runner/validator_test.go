package runner

import (
	"testing"
)

func TestValidateRetryConfig(t *testing.T) {
	validator := NewPipelineValidator()

	tests := []struct {
		name      string
		configMap map[string]interface{}
		expectErr bool
		errorMsg  string
	}{
		{
			name: "valid retry config",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 3,
					"delay":    5,
					"backoff":  2.0,
				},
			},
			expectErr: false,
		},
		{
			name: "no retry config should be valid",
			configMap: map[string]interface{}{
				"image": "alpine:latest",
			},
			expectErr: false,
		},
		{
			name: "retry is not an object",
			configMap: map[string]interface{}{
				"retry": "invalid",
			},
			expectErr: true,
			errorMsg:  "'retry' must be an object",
		},
		{
			name: "attempts is not an integer",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": "invalid",
				},
			},
			expectErr: true,
			errorMsg:  "retry.attempts must be an integer",
		},
		{
			name: "attempts is less than 1",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 0,
				},
			},
			expectErr: true,
			errorMsg:  "retry.attempts must be at least 1",
		},
		{
			name: "attempts exceeds maximum",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 15,
				},
			},
			expectErr: true,
			errorMsg:  "retry.attempts must not exceed 10 (to prevent infinite loops)",
		},
		{
			name: "delay is not an integer",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 3,
					"delay":    "invalid",
				},
			},
			expectErr: true,
			errorMsg:  "retry.delay must be an integer (seconds)",
		},
		{
			name: "delay is negative",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 3,
					"delay":    -5,
				},
			},
			expectErr: true,
			errorMsg:  "retry.delay must be non-negative",
		},
		{
			name: "delay exceeds maximum",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 3,
					"delay":    400,
				},
			},
			expectErr: true,
			errorMsg:  "retry.delay must not exceed 300 seconds",
		},
		{
			name: "backoff is not a number",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 3,
					"delay":    5,
					"backoff":  "invalid",
				},
			},
			expectErr: true,
			errorMsg:  "retry.backoff must be a number",
		},
		{
			name: "backoff is zero or negative",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 3,
					"delay":    5,
					"backoff":  0.0,
				},
			},
			expectErr: true,
			errorMsg:  "retry.backoff must be greater than 0",
		},
		{
			name: "backoff exceeds maximum",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 3,
					"delay":    5,
					"backoff":  15.0,
				},
			},
			expectErr: true,
			errorMsg:  "retry.backoff must not exceed 10.0",
		},
		{
			name: "valid edge case values",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 10,   // max allowed
					"delay":    300,  // max allowed
					"backoff":  10.0, // max allowed
				},
			},
			expectErr: false,
		},
		{
			name: "valid minimum values",
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 1,    // min allowed
					"delay":    0,    // min allowed
					"backoff":  0.1,  // just above min
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateRetryConfig(tt.configMap)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateJobWithRetry(t *testing.T) {
	validator := NewPipelineValidator()

	// Mock viper to avoid dependency on actual config
	// This test focuses on the retry validation integration
	configMap := map[string]interface{}{
		"image": "alpine:latest",
		"retry": map[string]interface{}{
			"attempts": 3,
			"delay":    5,
			"backoff":  2.0,
		},
	}

	// Since validateJob is not exported and depends on viper,
	// we test validateRetryConfig directly which is the core functionality
	err := validator.validateRetryConfig(configMap)
	if err != nil {
		t.Errorf("validateRetryConfig failed: %v", err)
	}
}

func TestRetryConfigBoundaries(t *testing.T) {
	validator := NewPipelineValidator()

	// Test boundary values that should be valid
	validBoundaries := []map[string]interface{}{
		{
			"retry": map[string]interface{}{
				"attempts": 1,
			},
		},
		{
			"retry": map[string]interface{}{
				"attempts": 10,
			},
		},
		{
			"retry": map[string]interface{}{
				"delay": 0,
			},
		},
		{
			"retry": map[string]interface{}{
				"delay": 300,
			},
		},
		{
			"retry": map[string]interface{}{
				"backoff": 0.1,
			},
		},
		{
			"retry": map[string]interface{}{
				"backoff": 10.0,
			},
		},
	}

	for i, configMap := range validBoundaries {
		t.Run("valid_boundary_"+string(rune(i+'0')), func(t *testing.T) {
			err := validator.validateRetryConfig(configMap)
			if err != nil {
				t.Errorf("boundary test %d should be valid but got error: %v", i, err)
			}
		})
	}

	// Test boundary values that should be invalid
	invalidBoundaries := []struct {
		configMap map[string]interface{}
		name      string
	}{
		{
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 0,
				},
			},
			name: "attempts_too_low",
		},
		{
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"attempts": 11,
				},
			},
			name: "attempts_too_high",
		},
		{
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"delay": -1,
				},
			},
			name: "delay_negative",
		},
		{
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"delay": 301,
				},
			},
			name: "delay_too_high",
		},
		{
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"backoff": 0.0,
				},
			},
			name: "backoff_zero",
		},
		{
			configMap: map[string]interface{}{
				"retry": map[string]interface{}{
					"backoff": 10.1,
				},
			},
			name: "backoff_too_high",
		},
	}

	for _, tt := range invalidBoundaries {
		t.Run("invalid_boundary_"+tt.name, func(t *testing.T) {
			err := validator.validateRetryConfig(tt.configMap)
			if err == nil {
				t.Errorf("boundary test %s should be invalid but got no error", tt.name)
			}
		})
	}
}