package utils

import "fmt"

func AnyToString(value any, defaultString string) string {
	if value == nil {
		return defaultString
	}
	switch v := value.(type) {
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
		return defaultString
	case int:
		return fmt.Sprintf("%d", v)
	case *int:
		if v != nil {
			return fmt.Sprintf("%d", *v)
		}
		return defaultString
	case float64:
		return fmt.Sprintf("%f", v)
	case *float64:
		if v != nil {
			return fmt.Sprintf("%f", *v)
		}
		return defaultString
	case bool:
		return returnTrueOrFalse(v)
	case *bool:
		return returnTrueOrFalse(v)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func returnTrueOrFalse(value any) string {
	if value == nil {
		return "false"
	}
	switch v := value.(type) {
	case bool:
		return fmt.Sprintf("%t", v)
	case *bool:
		if v != nil {
			return fmt.Sprintf("%t", *v)
		}
		return "false"
	default:
		return "false"
	}
}
