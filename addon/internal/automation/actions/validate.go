package actions

import (
	"fmt"
	"strings"

	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/domain"
)

func ValidateActionParams(actionType domain.ActionType, params map[string]any) error {
	for _, field := range actionType.ParamSchema {
		visible := isFieldVisible(field, params)
		required := field.Required && visible
		value, hasValue := params[field.Key]

		if required && (!hasValue || isEmpty(value)) {
			return fmt.Errorf("missing required param %q", field.Key)
		}
		if !hasValue || isEmpty(value) {
			continue
		}

		switch field.Kind {
		case domain.ParamString:
			if _, ok := value.(string); !ok {
				return fmt.Errorf("param %q must be string", field.Key)
			}
		case domain.ParamEnum:
			actual, ok := value.(string)
			if !ok {
				return fmt.Errorf("param %q must be enum string", field.Key)
			}
			if !contains(field.Options, actual) {
				return fmt.Errorf("param %q has invalid value %q", field.Key, actual)
			}
		case domain.ParamBool:
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("param %q must be bool", field.Key)
			}
		default:
			return fmt.Errorf("param %q has unsupported kind %q", field.Key, field.Kind)
		}
	}

	return nil
}

func isFieldVisible(field domain.ActionParamField, params map[string]any) bool {
	if field.VisibleIf == nil {
		return true
	}
	actual, ok := params[field.VisibleIf.Key]
	if !ok {
		return false
	}
	actualString, ok := actual.(string)
	if !ok {
		return false
	}
	return actualString == field.VisibleIf.Equals
}

func isEmpty(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	default:
		return false
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
