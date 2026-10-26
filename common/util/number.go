package util

import (
	"encoding/json"
	"strconv"
)

// AnyToInt64 reads an integer that may arrive as any numeric Go type: a database driver scans
// bigint as int64, while a JSON decode yields float64 or json.Number. A fractional or non-numeric
// value reports false rather than truncating.
func AnyToInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case float64:
		if typed != float64(int64(typed)) {
			return 0, false
		}
		return int64(typed), true
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseInt(typed, 10, 64)
		return parsed, err == nil
	}
	return 0, false
}
