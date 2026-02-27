package enum_util

import (
	"database/sql/driver"
	"fmt"
)

func ValueSQL[T ~uint8](e *T, nameMap map[T]string) (driver.Value, error) {
	v, ok := nameMap[*e]
	if !ok {
		return nil, fmt.Errorf("invalid ScopeType: %d", e)
	}

	return v, nil
}

func ScanSQL[T ~uint8](src any, valueMap map[string]T, nameMap map[T]string) (T, error) {
	s, ok := src.(string)
	if !ok {
		return 0, fmt.Errorf("ScopeType must be a string, got %T", src)
	}

	v, ok := valueMap[s]
	if !ok {
		return 0, fmt.Errorf(
			"enum '%d' is not registered, must be one of: %v",
			v,
			DescriptionFromMap(nameMap),
		)
	}

	return v, nil
}
