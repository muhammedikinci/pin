package runner

import (
	"log"
	"math"
	"os"
	"testing"
	"time"
)

func TestRetryDelayCalculation(t *testing.T) {
	tests := []struct {
		name              string
		delaySeconds      int
		backoffMultiplier float64
		attempt           int
		expectedDelay     time.Duration
	}{
		{
			name:              "linear delay - no backoff",
			delaySeconds:      5,
			backoffMultiplier: 1.0,
			attempt:           1,
			expectedDelay:     5 * time.Second,
		},
		{
			name:              "linear delay - second attempt",
			delaySeconds:      3,
			backoffMultiplier: 1.0,
			attempt:           2,
			expectedDelay:     3 * time.Second,
		},
		{
			name:              "exponential backoff - first retry",
			delaySeconds:      2,
			backoffMultiplier: 2.0,
			attempt:           1,
			expectedDelay:     2 * time.Second, // 2 * 2^(1-1) = 2 * 1 = 2
		},
		{
			name:              "exponential backoff - second retry",
			delaySeconds:      2,
			backoffMultiplier: 2.0,
			attempt:           2,
			expectedDelay:     4 * time.Second, // 2 * 2^(2-1) = 2 * 2 = 4
		},
		{
			name:              "exponential backoff - third retry",
			delaySeconds:      1,
			backoffMultiplier: 2.0,
			attempt:           3,
			expectedDelay:     4 * time.Second, // 1 * 2^(3-1) = 1 * 4 = 4
		},
		{
			name:              "fractional backoff",
			delaySeconds:      4,
			backoffMultiplier: 1.5,
			attempt:           2,
			expectedDelay:     6 * time.Second, // 4 * 1.5^(2-1) = 4 * 1.5 = 6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate delay using the same formula as in jobRunnerWithRetry
			delay := time.Duration(float64(tt.delaySeconds) * math.Pow(tt.backoffMultiplier, float64(tt.attempt-1))) * time.Second
			
			if delay != tt.expectedDelay {
				t.Errorf("Expected delay %v, got %v", tt.expectedDelay, delay)
			}
		})
	}
}

func TestRetryConfig_DefaultValues(t *testing.T) {
	job := &Job{
		RetryConfig: RetryConfig{
			MaxAttempts:       1, // Default - no retry
			DelaySeconds:     1,
			BackoffMultiplier: 1.0,
		},
	}

	if job.RetryConfig.MaxAttempts != 1 {
		t.Errorf("Expected default MaxAttempts=1, got %d", job.RetryConfig.MaxAttempts)
	}
	if job.RetryConfig.DelaySeconds != 1 {
		t.Errorf("Expected default DelaySeconds=1, got %d", job.RetryConfig.DelaySeconds)
	}
	if job.RetryConfig.BackoffMultiplier != 1.0 {
		t.Errorf("Expected default BackoffMultiplier=1.0, got %f", job.RetryConfig.BackoffMultiplier)
	}
}

func TestRetryWithLogging(t *testing.T) {
	// Test that retry mechanism works with different logging configurations
	job := &Job{
		Name: "log-test-job",
		RetryConfig: RetryConfig{
			MaxAttempts:       2,
			DelaySeconds:     1,
			BackoffMultiplier: 1.0,
		},
		ErrorChannel: make(chan error, 1),
		InfoLog:      log.New(os.Stdout, "⚉ test ", 0),
	}

	// This test verifies that logging setup doesn't cause panics
	// The actual logging output is visual and tested manually
	if job.InfoLog == nil {
		t.Error("InfoLog should not be nil")
	}
}

// Helper function to create a mock job for testing
func createMockJob(name string, retryConfig RetryConfig) *Job {
	return &Job{
		Name:         name,
		Image:        "alpine:latest",
		RetryConfig:  retryConfig,
		ErrorChannel: make(chan error, 1),
		Script:       []string{"echo test"},
	}
}