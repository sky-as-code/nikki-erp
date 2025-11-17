package model

import (
	"reflect"
	"strings"
)

type EntityMap map[string]any

const AttrsFieldName = "Attrs_"

type DynamicBase struct {
	Attrs_ EntityMap `json:"attrs_,omitempty"`
}

// ToEntityMap converts a struct into an EntityMap by mapping
// the struct fields tag "json" to the EntityMap keys. Especially, all keys in struct field Attrs_
// is flattened to the EntityMap keys.
func StructToEntityMap(src any) EntityMap {
	result := make(EntityMap)

	if src == nil {
		return result
	}

	value := dereferenceValue(reflect.ValueOf(src))
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return result
	}

	var attrsField *reflect.Value
	processStructFields(value, result, &attrsField)
	flattenAttrsField(attrsField, result)

	return result
}

// dereferenceValue dereferences pointers until it reaches a non-pointer value.
// Returns an invalid value if a nil pointer is encountered.
func dereferenceValue(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func parseJSONTag(jsonTag string) (tagName string, hasOmitempty bool) {
	if jsonTag == "" || jsonTag == "-" {
		return "", false
	}

	tagParts := strings.Split(jsonTag, ",")
	tagName = tagParts[0]
	if tagName == "" {
		return "", false
	}

	for _, part := range tagParts[1:] {
		if strings.TrimSpace(part) == "omitempty" {
			hasOmitempty = true
			break
		}
	}

	return tagName, hasOmitempty
}

func processField(field reflect.StructField, fieldValue reflect.Value, result EntityMap, tagName string, hasOmitempty bool) (isAttrsField bool) {
	if field.Name == AttrsFieldName {
		isAttrsField = true
		return
	}

	isAttrsField = false

	if fieldValue.Kind() == reflect.Pointer {
		if fieldValue.IsNil() {
			if hasOmitempty {
				return
			}
			result[tagName] = nil
			return
		}
		fieldValue = fieldValue.Elem()
	}

	result[tagName] = fieldValue.Interface()
	return
}

func processStructFields(value reflect.Value, result EntityMap, attrsField **reflect.Value) {
	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		isUnexported := field.PkgPath != ""

		if isUnexported {
			continue
		}

		tagName, hasOmitempty := parseJSONTag(field.Tag.Get("json"))
		if tagName == "" {
			continue
		}

		fieldValue := value.Field(i)
		if processField(field, fieldValue, result, tagName, hasOmitempty) {
			*attrsField = &fieldValue
		}
	}
}

func flattenAttrsField(attrsField *reflect.Value, result EntityMap) {
	if attrsField == nil {
		return
	}

	attrsValue := *attrsField
	attrsValue = dereferenceValue(attrsValue)
	if !attrsValue.IsValid() {
		return
	}

	if attrsValue.Kind() == reflect.Map {
		for _, key := range attrsValue.MapKeys() {
			keyStr := key.String()
			result[keyStr] = attrsValue.MapIndex(key).Interface()
		}
	}
}

// EntityMapToStruct converts an EntityMap into a struct by mapping
// the EntityMap keys to the struct fields tag "json".
// Any keys without corresponding struct field tag is mapped to the struct field Attrs_.
func EntityMapToStruct[T any](src EntityMap) *T {
	if src == nil {
		return nil
	}

	var zero T
	result := reflect.New(reflect.TypeOf(zero)).Elem()
	resultType := result.Type()

	unmappedAttrs, attrsFieldRef := copyMapToStruct(src, result, resultType)
	populateAttrsField(result, attrsFieldRef, unmappedAttrs)

	return result.Addr().Interface().(*T)
}

// attrsFieldRef holds a reference to an Attrs_ field (either direct or from embedded DynamicBase).
type attrsFieldRef struct {
	fieldIndex int
	isEmbedded bool
}

// copyMapToStruct loops through type T's fields and populates them from the EntityMap.
// Returns unmapped attributes that don't match any field and a reference to the Attrs_ field.
func copyMapToStruct(src EntityMap, result reflect.Value, resultType reflect.Type) (EntityMap, *attrsFieldRef) {
	unmappedAttrs := make(EntityMap)

	// Copy all keys from src to unmappedAttrs initially
	for key, value := range src {
		unmappedAttrs[key] = value
	}

	// Get the struct name of DynamicBase using reflection
	dynamicBaseType := reflect.TypeOf(DynamicBase{})
	dynamicBaseStructName := dynamicBaseType.Name()

	var attrsField *attrsFieldRef

	// Loop through all fields of the struct
	for i := 0; i < resultType.NumField(); i++ {
		field := resultType.Field(i)
		isUnexported := field.PkgPath != ""
		if isUnexported {
			continue
		}

		if processAttrsField(field, i, dynamicBaseStructName, &attrsField) {
			continue
		}

		tagName, _ := parseJSONTag(field.Tag.Get("json"))
		if tagName == "" {
			continue
		}

		if tagName := populateFieldFromMap(src, result, field, i); tagName != "" {
			delete(unmappedAttrs, tagName)
		}
	}

	return unmappedAttrs, attrsField
}

// processAttrsField processes a field to find Attrs_ field (direct or embedded).
// Returns true if the field was an Attrs_ field (direct or embedded), false otherwise.
func processAttrsField(field reflect.StructField, fieldIndex int, dynamicBaseStructName string, attrsField **attrsFieldRef) bool {
	// Check for direct Attrs_ field (takes precedence)
	if field.Name == AttrsFieldName {
		*attrsField = &attrsFieldRef{
			fieldIndex: fieldIndex,
			isEmbedded: false,
		}
		return true
	}

	// Check for anonymous embedded DynamicBase (only if we haven't found direct Attrs_)
	if *attrsField == nil && isEmbeddedDynamicBase(field, dynamicBaseStructName) {
		*attrsField = &attrsFieldRef{
			fieldIndex: fieldIndex,
			isEmbedded: true,
		}
		return true
	}

	return false
}

// isEmbeddedDynamicBase checks if a field is an anonymous embedded DynamicBase.
func isEmbeddedDynamicBase(field reflect.StructField, dynamicBaseStructName string) bool {
	if !field.Anonymous {
		return false
	}

	fieldType := field.Type
	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}
	return fieldType.Kind() == reflect.Struct && fieldType.Name() == dynamicBaseStructName
}

// populateFieldFromMap populates a field if the corresponding key exists in src EntityMap.
// Returns the tagName if the field was populated, empty string otherwise.
func populateFieldFromMap(
	src EntityMap, result reflect.Value, field reflect.StructField, fieldIndex int,
) string {
	tagName, _ := parseJSONTag(field.Tag.Get("json"))
	if tagName == "" {
		return ""
	}

	value, exists := src[tagName]
	if !exists {
		return ""
	}

	fieldValue := result.Field(fieldIndex)
	if !fieldValue.CanSet() {
		return ""
	}

	setFieldValue(fieldValue, field.Type, value)
	return tagName
}

// populateAttrsField populates the Attrs_ field
func populateAttrsField(result reflect.Value, attrsFieldRef *attrsFieldRef, attrs EntityMap) {
	if attrsFieldRef == nil {
		return
	}

	var attrsField reflect.Value

	if attrsFieldRef.isEmbedded {
		// Get the embedded DynamicBase field value
		embeddedFieldValue := result.Field(attrsFieldRef.fieldIndex)
		if !embeddedFieldValue.IsValid() {
			return
		}

		// Find Attrs_ field within the embedded DynamicBase
		attrsField = embeddedFieldValue.FieldByName(AttrsFieldName)
	} else {
		// Direct Attrs_ field
		attrsField = result.Field(attrsFieldRef.fieldIndex)
	}

	if !attrsField.IsValid() || !attrsField.CanSet() {
		return
	}

	if len(attrs) > 0 {
		attrsField.Set(reflect.ValueOf(attrs))
	} else {
		attrsField.Set(reflect.Zero(attrsField.Type()))
	}
}

func setFieldValue(fieldValue reflect.Value, fieldType reflect.Type, value any) {
	if !fieldValue.CanSet() {
		return
	}

	if value == nil {
		setNilValue(fieldValue, fieldType)
		return
	}

	valueType := reflect.TypeOf(value)
	valueValue := reflect.ValueOf(value)

	if fieldType.Kind() == reflect.Pointer {
		setPointerField(fieldValue, fieldType, valueType, valueValue)
		return
	}

	setDirectField(fieldValue, fieldType, valueType, valueValue)
}

func setNilValue(fieldValue reflect.Value, fieldType reflect.Type) {
	if fieldType.Kind() == reflect.Pointer {
		fieldValue.Set(reflect.Zero(fieldType))
	}
}

func setPointerField(
	fieldValue reflect.Value, fieldType reflect.Type,
	valueType reflect.Type, valueValue reflect.Value,
) {
	elemType := fieldType.Elem()
	if valueType.AssignableTo(elemType) {
		ptr := reflect.New(elemType)
		ptr.Elem().Set(valueValue)
		fieldValue.Set(ptr)
	} else if valueType.ConvertibleTo(elemType) {
		ptr := reflect.New(elemType)
		ptr.Elem().Set(valueValue.Convert(elemType))
		fieldValue.Set(ptr)
	}
}

func setDirectField(
	fieldValue reflect.Value,
	fieldType reflect.Type,
	valueType reflect.Type,
	valueValue reflect.Value,
) {
	if valueType.AssignableTo(fieldType) {
		fieldValue.Set(valueValue)
	} else if valueType.ConvertibleTo(fieldType) {
		fieldValue.Set(valueValue.Convert(fieldType))
	}
}
