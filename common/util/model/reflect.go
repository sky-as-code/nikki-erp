package model

import (
	"reflect"
	"strings"

	util "github.com/sky-as-code/nikki-erp/common/util"
)

// NonNilFields inspects given struct obj and invokes the action for each field with non-nil value.
func NonNilFields(obj any, action func(reflect.StructField, reflect.Value)) {
	rv := IndirectValue(obj)
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		value := reflect.Indirect(rv).FieldByName(field.Name)
		vtype := value.Type()
		isNil := isNil(value)
		isPtr := util.IsPointerType(vtype)
		isNotNilPtr := isPtr && !isNil
		isArray := util.IsArrayType(vtype)
		if value.IsValid() && !isArray && (!isPtr || isNotNilPtr) {
			action(field, value)
		}
	}
}

func isNil(value reflect.Value) (result bool) {
	defer func() {
		if r := recover(); r != nil {
			result = false
		}
	}()
	result = value.IsNil()
	return
}

// TagsOfNonNilKeys inspects given struct obj and returns JSON tag names of fields whose value is not nil.
// Fields with nil value are omitted.
func TagsOfNonNilKeys(obj any) []string {
	keys := make([]string, 0)

	NonNilFields(obj, func(field reflect.StructField, _ reflect.Value) {
		keys = append(keys, ExtractJsonName(field))
	})

	return keys
}

// TagsOfNonNilKeysMap inspects given struct obj and returns a map of JSON tag names and its not-nil values.
// Fields with nil value are omitted.
func TagsOfNonNilKeysMap(obj any) map[string]any {
	keyValue := make(map[string]any)

	NonNilFields(obj, func(field reflect.StructField, value reflect.Value) {
		keyValue[ExtractJsonName(field)] = value.Interface()
	})

	return keyValue
}

func ExtractJsonName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json") // Eg: "email,omitempty"
	values := strings.Split(jsonTag, ",")
	return values[0]
}

func IndirectValue(obj any) reflect.Value {
	rv := reflect.ValueOf(obj)
	for util.IsPointerType(rv.Type()) {
		rv = reflect.Indirect(rv)
	}
	return rv
}
