package field_validators

import "fmt"

// универсальные валидаторы

type FormulaValidator struct {
	Formula     func(interface{}) string
	FormulaRaw  func(interface{}) string
	Name        func(interface{}) string
	Description func(interface{}) string
	IsNotified  func(interface{}) string
	IsActive    func(interface{}) string
	IsHistoryOn func(interface{}) string
}

func ValidateBool(value interface{}) string {
	_, ok := value.(bool)
	if !ok {
		return "value must be a boolean"
	}
	return ""
}

func ValidateText(minLength, maxLength int) func(value interface{}) string {
	return func(value interface{}) string {
		str, ok := value.(string)
		if !ok || len(str) < minLength || len(str) > maxLength {
			return fmt.Sprintf("field should be in the range from %d to %d characters", minLength, maxLength)
		}
		return ""
	}
}
