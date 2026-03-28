package model

// CoerceFilterBool parses search-graph filter values into bool (bool, *bool, string literals, etc.).
func CoerceFilterBool(value any) (bool, error) {
	return toBool(value)
}
