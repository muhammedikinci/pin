package runner

import (
	"os"
	"testing"
)

func TestConditionEvaluator_EvaluateCondition(t *testing.T) {
	ce := NewConditionEvaluator()

	tests := []struct {
		name      string
		condition string
		envVars   map[string]string
		expected  bool
	}{
		{
			name:      "Empty condition should return true",
			condition: "",
			expected:  true,
		},
		{
			name:      "Simple equality - true",
			condition: `"main" == "main"`,
			expected:  true,
		},
		{
			name:      "Simple equality - false",
			condition: `"main" == "dev"`,
			expected:  false,
		},
		{
			name:      "Environment variable equality - true",
			condition: `$BRANCH == "main"`,
			envVars:   map[string]string{"BRANCH": "main"},
			expected:  true,
		},
		{
			name:      "Environment variable equality - false",
			condition: `$BRANCH == "main"`,
			envVars:   map[string]string{"BRANCH": "dev"},
			expected:  false,
		},
		{
			name:      "Environment variable inequality - true",
			condition: `$BRANCH != "main"`,
			envVars:   map[string]string{"BRANCH": "dev"},
			expected:  true,
		},
		{
			name:      "Environment variable inequality - false",
			condition: `$BRANCH != "main"`,
			envVars:   map[string]string{"BRANCH": "main"},
			expected:  false,
		},
		{
			name:      "AND condition - true",
			condition: `$BRANCH == "main" && $ENV == "prod"`,
			envVars:   map[string]string{"BRANCH": "main", "ENV": "prod"},
			expected:  true,
		},
		{
			name:      "AND condition - false",
			condition: `$BRANCH == "main" && $ENV == "prod"`,
			envVars:   map[string]string{"BRANCH": "main", "ENV": "dev"},
			expected:  false,
		},
		{
			name:      "OR condition - true",
			condition: `$BRANCH == "main" || $BRANCH == "dev"`,
			envVars:   map[string]string{"BRANCH": "main"},
			expected:  true,
		},
		{
			name:      "OR condition - false",
			condition: `$BRANCH == "main" || $BRANCH == "dev"`,
			envVars:   map[string]string{"BRANCH": "feature"},
			expected:  false,
		},
		{
			name:      "Variable existence - true",
			condition: `$DEPLOY`,
			envVars:   map[string]string{"DEPLOY": "true"},
			expected:  true,
		},
		{
			name:      "Variable existence - false",
			condition: `$DEPLOY`,
			envVars:   map[string]string{},
			expected:  false,
		},
		{
			name:      "Variable false value",
			condition: `$DEPLOY`,
			envVars:   map[string]string{"DEPLOY": "false"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			result := ce.EvaluateCondition(tt.condition)
			if result != tt.expected {
				t.Errorf("EvaluateCondition(%q) = %v, expected %v", tt.condition, result, tt.expected)
			}
		})
	}
}

func TestConditionEvaluator_ResolveValue(t *testing.T) {
	ce := NewConditionEvaluator()

	tests := []struct {
		name     string
		value    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "Environment variable",
			value:    "$TEST_VAR",
			envVars:  map[string]string{"TEST_VAR": "test_value"},
			expected: "test_value",
		},
		{
			name:     "Double quoted string",
			value:    `"hello world"`,
			expected: "hello world",
		},
		{
			name:     "Single quoted string",
			value:    `'hello world'`,
			expected: "hello world",
		},
		{
			name:     "Plain string",
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "Empty environment variable",
			value:    "$NON_EXISTENT",
			envVars:  map[string]string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			result := ce.resolveValue(tt.value)
			if result != tt.expected {
				t.Errorf("resolveValue(%q) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestConditionEvaluator_IsValidCondition(t *testing.T) {
	ce := NewConditionEvaluator()

	tests := []struct {
		name      string
		condition string
		expected  bool
	}{
		{
			name:      "Empty condition",
			condition: "",
			expected:  true,
		},
		{
			name:      "Simple equality",
			condition: `$BRANCH == "main"`,
			expected:  true,
		},
		{
			name:      "Complex condition",
			condition: `$BRANCH == "main" && $ENV != "test"`,
			expected:  true,
		},
		{
			name:      "Invalid characters",
			condition: `$BRANCH == "main"; rm -rf /`,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ce.IsValidCondition(tt.condition)
			if result != tt.expected {
				t.Errorf("IsValidCondition(%q) = %v, expected %v", tt.condition, result, tt.expected)
			}
		})
	}
}