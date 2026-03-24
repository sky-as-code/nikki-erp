package model

import (
	"reflect"

	"go.bryk.io/pkg/errors"
)

const SchemaFieldTag = "entity"
const JsonFieldTag = "json"

type DynamicModel interface {
	DynamicModelGetter
	DynamicModelSetter
}

type DynamicModelGetter interface {
	GetFieldData() DynamicFields
}

type DynamicModelSetter interface {
	SetFieldData(data DynamicFields)
}

type SchemaGetter interface {
	GetFieldData() DynamicFields
	GetSchema() *ModelSchema
}

func StructToDynamicEntity(src any) (DynamicFields, error) {
	v := reflect.ValueOf(src)
	if v.Kind() != reflect.Ptr {
		return nil, errors.New("expected pointer to struct")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil, errors.New("expected struct")
	}
	t := v.Type()
	result := make(DynamicFields, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		key := resolveFieldKey(field)
		if key == "" {
			continue
		}
		fv := v.Field(i)
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				result[key] = nil
				continue
			}
			fv = fv.Elem()
		}
		result[key] = fv.Interface()
	}
	return result, nil
}

func DynamicEntityToStruct[T any](src DynamicFields, dest *T) error {
	if dest == nil {
		return errors.New("destination pointer cannot be nil")
	}
	elem := reflect.ValueOf(dest).Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("destination must point to a struct")
	}
	t := elem.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		key := resolveFieldKey(field)
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
		assignToField(fieldValue, val)
	}
	return nil
}

func assignToField(fieldValue reflect.Value, val any) {
	if val == nil {
		if fieldValue.Kind() == reflect.Ptr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}
	rv := reflect.ValueOf(val)
	if fieldValue.Kind() == reflect.Ptr {
		elemType := fieldValue.Type().Elem()
		ptr := reflect.New(elemType)
		if rv.Type().AssignableTo(elemType) {
			ptr.Elem().Set(rv)
			fieldValue.Set(ptr)
		} else if rv.Type().ConvertibleTo(elemType) {
			ptr.Elem().Set(rv.Convert(elemType))
			fieldValue.Set(ptr)
		}
		return
	}
	if rv.Type().AssignableTo(fieldValue.Type()) {
		fieldValue.Set(rv)
	} else if rv.Type().ConvertibleTo(fieldValue.Type()) {
		fieldValue.Set(rv.Convert(fieldValue.Type()))
	}
}

func resolveFieldKey(field reflect.StructField) string {
	if tag, ok := field.Tag.Lookup(SchemaFieldTag); ok && tag != "" {
		return parseTagKey(tag, field.Name)
	}
	return parseTagKey(field.Tag.Get(JsonFieldTag), field.Name)
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
