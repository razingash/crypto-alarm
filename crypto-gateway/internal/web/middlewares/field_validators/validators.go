package field_validators

import "fmt"

// универсальные валидаторы

type StrategyValidator struct {
	Name        func(interface{}) string
	Description func(interface{}) string
	IsNotified  func(interface{}) string
	IsActive    func(interface{}) string
	IsHistoryOn func(interface{}) string
	Cooldown    func(interface{}) string
	Conditions  func(value interface{}) string
}

func ValidateBool(value interface{}) string {
	_, ok := value.(bool)
	if !ok {
		return "value must be a boolean"
	}
	return ""
}

func ValidateCooldown(value interface{}) string {
	num, ok := value.(float64)
	if !ok {
		return "cooldown must be a number"
	}

	cooldown := int(num)
	if cooldown < 1 {
		return "cooldown must be at least 1 second"
	}
	if cooldown > 604800 {
		return "cooldown must not exceed 604800 seconds (7 days)"
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
