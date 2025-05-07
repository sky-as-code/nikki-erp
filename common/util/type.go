package util

import (
	"reflect"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func IsArrayType(target reflect.Type) bool {
	kind := target.Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

func IsConvertible(sourceValue any, targetType reflect.Type) bool {
	val := reflect.ValueOf(sourceValue)
	return val.Type().ConvertibleTo(targetType)
}

func IsErrorObj(target interface{}) bool {
	_, isErr := target.(error)
	return isErr
}

func IsErrorType(target reflect.Type) bool {
	return ImplementsInterface(target, errorInterface)
}

func IsFuncType(target reflect.Type) bool {
	return target.Kind() == reflect.Func
}

func IsInterfaceType(target reflect.Type) bool {
	return target.Kind() == reflect.Interface
}

func IsPointerType(target reflect.Type) bool {
	return target.Kind() == reflect.Ptr
}

func IsStructType(target reflect.Type) bool {
	return (target.Kind() == reflect.Struct ||
		(IsPointerType(target) && target.Elem().Kind() == reflect.Struct))
}

func ImplementsInterface(target reflect.Type, theInterface reflect.Type) bool {
	return (target.Implements(theInterface) ||
		(IsPointerType(target) && target.Elem().Implements(theInterface)))
}

func ToPtr[T interface{}](source T) *T {
	return &source
}
