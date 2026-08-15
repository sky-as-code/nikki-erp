package model

import (
	"regexp"

	"go.bryk.io/pkg/errors"

	cjson "github.com/sky-as-code/nikki-erp/common/json"
)

// decodeDataType turns the JSON "data_type" value into a FieldDataType.
//
// It always calls the real constructors rather than writing into an Options map, because
// the option getters type-assert concrete Go types (length as []int, ranges as []int32 /
// []int64 / []string). A raw JSON decode yields float64 and []any, which would silently
// disable range checks or panic during validation.
func decodeDataType(raw any, fieldName string) FieldDataType {
	dto := normalizeDataType(raw, fieldName)
	dataType := constructDataType(dto, fieldName)

	if dto.Array {
		return dataType.ArrayType()
	}
	return dataType
}

// normalizeDataType accepts the bare-string form ("boolean") and the object form,
// returning the object form in both cases.
func normalizeDataType(raw any, fieldName string) *dataTypeJsonDto {
	switch value := raw.(type) {
	case string:
		return &dataTypeJsonDto{Type: value}
	case map[string]any:
		// Round-trip through the codec so the DTO tags do the field mapping.
		encoded, err := cjson.Marshal(value)
		if err != nil {
			panic(errors.Wrapf(err, "field '%s': cannot re-encode data_type", fieldName))
		}
		var dto dataTypeJsonDto
		if err := cjson.Unmarshal(encoded, &dto); err != nil {
			panic(errors.Wrapf(err, "field '%s': cannot decode data_type", fieldName))
		}
		return &dto
	default:
		panic(errors.Errorf("field '%s': data_type must be a string or an object, got %T", fieldName, raw))
	}
}

func constructDataType(dto *dataTypeJsonDto, fieldName string) FieldDataType {
	switch dto.Type {
	case "string":
		return FieldDataTypeString(jsonToInt(dto.Min, fieldName), jsonToInt(dto.Max, fieldName), stringOpts(dto, fieldName)...)
	case "secret":
		return FieldDataTypeSecret(jsonToInt(dto.Min, fieldName), jsonToInt(dto.Max, fieldName))
	case "langjson":
		return FieldDataTypeLangJson(jsonToInt(dto.Min, fieldName), jsonToInt(dto.Max, fieldName))
	case "int32":
		return FieldDataTypeInt32(int32(jsonToInt64(dto.Min, fieldName)), int32(jsonToInt64(dto.Max, fieldName)))
	case "int64":
		return FieldDataTypeInt64(jsonToInt64(dto.Min, fieldName), jsonToInt64(dto.Max, fieldName))
	case "decimal":
		return FieldDataTypeDecimal(jsonToString(dto.Min, fieldName), jsonToString(dto.Max, fieldName), dto.Scale)
	case "enum_string":
		return FieldDataTypeEnumString(jsonToStringSlice(dto.Values, fieldName))
	case "enum_int32":
		return FieldDataTypeEnumInt32(jsonToInt32Slice(dto.Values, fieldName))
	case "email":
		return FieldDataTypeEmail()
	case "phone":
		return FieldDataTypePhone()
	case "url":
		return FieldDataTypeUrl()
	case "ulid":
		return FieldDataTypeUlid()
	case "uuid":
		return FieldDataTypeUuid()
	case "boolean":
		return FieldDataTypeBoolean()
	case "date":
		return FieldDataTypeDate()
	case "time":
		return FieldDataTypeTime()
	case "datetime":
		return FieldDataTypeDateTime()
	case "etag":
		return FieldDataTypeEtag()
	case "langcode":
		return FieldDataTypeLangCode()
	case "slug":
		return FieldDataTypeSlug()
	case "jsonmap":
		return FieldDataTypeJsonMap()
	case "model":
		return FieldDataTypeModel()
	default:
		panic(errors.Errorf("field '%s': unsupported data type '%s'", fieldName, dto.Type))
	}
}

func stringOpts(dto *dataTypeJsonDto, fieldName string) []FieldDataTypeStringOpts {
	if dto.Sanitize == "" && dto.Regex == "" {
		return nil
	}

	opts := FieldDataTypeStringOpts{}
	if dto.Sanitize != "" {
		opts.SanitizeType = SanitizeType(dto.Sanitize)
	}
	if dto.Regex != "" {
		compiled, err := regexp.Compile(dto.Regex)
		if err != nil {
			panic(errors.Wrapf(err, "field '%s': invalid regex", fieldName))
		}
		opts.Regex = compiled
	}

	return []FieldDataTypeStringOpts{opts}
}

// jsonToInt converts a decoded JSON number to int. json-iterator yields float64 by default,
// but may yield json.Number or an integer type depending on configuration, so all are handled.
func jsonToInt(raw any, fieldName string) int {
	return int(jsonToInt64(raw, fieldName))
}

func jsonToInt64(raw any, fieldName string) int64 {
	switch value := raw.(type) {
	case float64:
		return int64(value)
	case float32:
		return int64(value)
	case int:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	default:
		panic(errors.Errorf("field '%s': expected a number, got %T", fieldName, raw))
	}
}

func jsonToString(raw any, fieldName string) string {
	str, ok := raw.(string)
	if !ok {
		panic(errors.Errorf("field '%s': expected a string, got %T", fieldName, raw))
	}
	return str
}

func jsonToStringSlice(raw []any, fieldName string) []string {
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		values = append(values, jsonToString(item, fieldName))
	}
	return values
}

func jsonToInt32Slice(raw []any, fieldName string) []int32 {
	values := make([]int32, 0, len(raw))
	for _, item := range raw {
		values = append(values, int32(jsonToInt64(item, fieldName)))
	}
	return values
}
