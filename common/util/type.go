package util

import (
	"reflect"

	"go.bryk.io/pkg/errors"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func IsArrayType(target reflect.Type) bool {
	kind := target.Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

func IsConvertible(sourceValue any, targetType reflect.Type) bool {
	if sourceValue == nil {
		return false
	}

	srcVal := reflect.ValueOf(sourceValue)
	srcType := srcVal.Type()

	// If targetType is a pointer, get its element type
	for targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	// If source is a pointer, get its element too
	for srcType.Kind() == reflect.Ptr {
		srcType = srcType.Elem()
	}

	return srcType.ConvertibleTo(targetType)
}

func ConvertType(sourceValue any, targetType reflect.Type) (any, error) {
	if sourceValue == nil {
		// Return zero pointer if targetType is a pointer, otherwise nil
		if targetType.Kind() == reflect.Ptr {
			return reflect.Zero(targetType).Interface(), nil
		}
		return nil, errors.Errorf("cannot convert nil to non-pointer type %v", targetType)
	}

	srcVal := reflect.ValueOf(sourceValue)
	srcType := srcVal.Type()

	// Unwrap pointer from targetType if needed
	isTargetPtr := targetType.Kind() == reflect.Ptr
	underlyingTargetType := targetType
	if isTargetPtr {
		underlyingTargetType = targetType.Elem()
	}

	// Unwrap source if it's a pointer
	for srcVal.Kind() == reflect.Ptr {
		if srcVal.IsNil() {
			return nil, errors.Errorf("cannot convert nil pointer to %v", targetType)
		}
		srcVal = srcVal.Elem()
		srcType = srcVal.Type()
	}

	// Check convertibility to base type
	if !srcType.ConvertibleTo(underlyingTargetType) {
		return nil, errors.Errorf("cannot convert %v (%v) to %v", sourceValue, srcType, targetType)
	}

	converted := srcVal.Convert(underlyingTargetType)

	if isTargetPtr {
		ptr := reflect.New(underlyingTargetType)
		ptr.Elem().Set(converted)
		return ptr.Interface(), nil
	}

	return converted.Interface(), nil
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
