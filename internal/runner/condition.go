package runner

import (
	"os"
	"regexp"
	"strings"
)

type ConditionEvaluator struct{}

func NewConditionEvaluator() *ConditionEvaluator {
	return &ConditionEvaluator{}
}

func (ce *ConditionEvaluator) EvaluateCondition(condition string) bool {
	if condition == "" {
		return true
	}

	condition = strings.TrimSpace(condition)
	
	if strings.Contains(condition, "&&") {
		return ce.evaluateAnd(condition)
	} else if strings.Contains(condition, "||") {
		return ce.evaluateOr(condition)
	} else if strings.Contains(condition, "==") {
		return ce.evaluateEquality(condition)
	} else if strings.Contains(condition, "!=") {
		return ce.evaluateInequality(condition)
	}
	
	return ce.evaluateVariable(condition)
}

func (ce *ConditionEvaluator) evaluateEquality(condition string) bool {
	parts := strings.Split(condition, "==")
	if len(parts) != 2 {
		return false
	}
	
	left := ce.resolveValue(strings.TrimSpace(parts[0]))
	right := ce.resolveValue(strings.TrimSpace(parts[1]))
	
	return left == right
}

func (ce *ConditionEvaluator) evaluateInequality(condition string) bool {
	parts := strings.Split(condition, "!=")
	if len(parts) != 2 {
		return false
	}
	
	left := ce.resolveValue(strings.TrimSpace(parts[0]))
	right := ce.resolveValue(strings.TrimSpace(parts[1]))
	
	return left != right
}

func (ce *ConditionEvaluator) evaluateAnd(condition string) bool {
	parts := strings.Split(condition, "&&")
	for _, part := range parts {
		partTrimmed := strings.TrimSpace(part)
		if strings.Contains(partTrimmed, "==") {
			if !ce.evaluateEquality(partTrimmed) {
				return false
			}
		} else if strings.Contains(partTrimmed, "!=") {
			if !ce.evaluateInequality(partTrimmed) {
				return false
			}
		} else {
			if !ce.evaluateVariable(partTrimmed) {
				return false
			}
		}
	}
	return true
}

func (ce *ConditionEvaluator) evaluateOr(condition string) bool {
	parts := strings.Split(condition, "||")
	for _, part := range parts {
		partTrimmed := strings.TrimSpace(part)
		if strings.Contains(partTrimmed, "==") {
			if ce.evaluateEquality(partTrimmed) {
				return true
			}
		} else if strings.Contains(partTrimmed, "!=") {
			if ce.evaluateInequality(partTrimmed) {
				return true
			}
		} else {
			if ce.evaluateVariable(partTrimmed) {
				return true
			}
		}
	}
	return false
}

func (ce *ConditionEvaluator) evaluateVariable(condition string) bool {
	value := ce.resolveValue(condition)
	return value != "" && value != "false" && value != "0"
}

func (ce *ConditionEvaluator) resolveValue(value string) string {
	value = strings.TrimSpace(value)
	
	if strings.HasPrefix(value, "$") {
		envVar := value[1:]
		return os.Getenv(envVar)
	}
	
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return value[1 : len(value)-1]
	}
	
	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		return value[1 : len(value)-1]
	}
	
	return value
}

func (ce *ConditionEvaluator) IsValidCondition(condition string) bool {
	if condition == "" {
		return true
	}
	
	validPattern := regexp.MustCompile(`^[\w\s\$"'=!&|]+$`)
	return validPattern.MatchString(condition)
}