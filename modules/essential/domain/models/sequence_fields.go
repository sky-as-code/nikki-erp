package models

import (
	"encoding/json"
	"math"
	"strconv"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// FieldString reads a string or *string value (model.Id is a string alias), empty when absent.
func FieldString(fields dmodel.DynamicFields, key string) string {
	switch typed := fields[key].(type) {
	case string:
		return typed
	case *string:
		if typed != nil {
			return *typed
		}
	}
	return ""
}

// FieldStringOr is FieldString with a fallback for an absent or empty value.
func FieldStringOr(fields dmodel.DynamicFields, key string, fallback string) string {
	if value := FieldString(fields, key); value != "" {
		return value
	}
	return fallback
}

// FieldInt64 reads any integer-like value, including the float64 and json.Number a decoded JSON body
// carries and the decimal string the API answers with. The second value reports a usable value.
func FieldInt64(fields dmodel.DynamicFields, key string) (int64, bool) {
	return AnyInt64(fields[key])
}

// AnyInt64 is FieldInt64 over a bare value.
func AnyInt64(raw any) (int64, bool) {
	switch typed := raw.(type) {
	case int64:
		return typed, true
	case *int64:
		if typed != nil {
			return *typed, true
		}
	case int32:
		return int64(typed), true
	case *int32:
		if typed != nil {
			return int64(*typed), true
		}
	case int:
		return int64(typed), true
	case float64:
		if typed == math.Trunc(typed) && math.Abs(typed) < math.MaxInt64 {
			return int64(typed), true
		}
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseInt(typed, 10, 64)
		return parsed, err == nil
	}
	return 0, false
}

// JsonMapString reads one string key of a jsonmap column value, which arrives as a map or as raw JSON.
func JsonMapString(raw any, key string) string {
	switch typed := raw.(type) {
	case map[string]any:
		value, _ := typed[key].(string)
		return value
	case dmodel.DynamicFields:
		value, _ := typed[key].(string)
		return value
	case []byte:
		return jsonBytesString(typed, key)
	case string:
		return jsonBytesString([]byte(typed), key)
	}
	return ""
}

func jsonBytesString(raw []byte, key string) string {
	decoded := map[string]any{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return ""
	}
	value, _ := decoded[key].(string)
	return value
}
