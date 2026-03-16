package schema

import (
	"reflect"

	"go.bryk.io/pkg/errors"
)

func StructToDynamicEntity(src any) (DynamicEntity, error) {
	v := reflect.ValueOf(src)
	if v.Kind() != reflect.Ptr {
		return nil, errors.New("expected pointer to struct")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil, errors.New("expected struct")
	}
	t := v.Type()
	result := make(DynamicEntity, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get(SchemaFieldTag)
		key := parseTagKey(tag, field.Name)
		if key == "" {
			continue
		}
		result[key] = v.Field(i).Interface()
	}
	return result, nil
}

func DynamicEntityToStruct[T any](src DynamicEntity, dest *T) error {
	// Validate destination is a non-nil pointer to a struct
	if dest == nil {
		return errors.New("destination pointer cannot be nil")
	}
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("destination must be a non-nil pointer to a struct")
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("destination must point to a struct")
	}

	t := elem.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get(SchemaFieldTag)
		key := parseTagKey(tag, field.Name)
		if key == "" {
			continue
		}
		val, exists := src[key]
		if !exists {
			continue
		}

		fieldValue := elem.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		valReflect := reflect.ValueOf(val)
		// Attempt to set if assignable or convertible
		if valReflect.Type().AssignableTo(fieldValue.Type()) {
			fieldValue.Set(valReflect)
		} else if valReflect.Type().ConvertibleTo(fieldValue.Type()) {
			fieldValue.Set(valReflect.Convert(fieldValue.Type()))
		} else {
			// Skip type mismatch for now, ignore quietly
			continue
		}
	}
	return nil
}

func parseTagKey(tag string, fieldName string) string {
	if tag == "-" {
		return ""
	}
	if tag == "" {
		return fieldName
	}
	for i, c := range tag {
		if c == ',' {
			return tag[:i]
		}
	}
	return tag
}
