package model

import (
	goerrors "errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-sanitize/sanitize"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/go-model"
)

func AddConversion[TIn any, TOut any](converter model.Converter) {
	model.AddConversion[TIn, TOut](converter)
}

func Copy(src, dest any) error {
	errs := model.Copy(dest, src)
	return goerrors.Join(errs...)
}

func MustCopy(src, dest any) {
	errs := model.Copy(dest, src)
	err := goerrors.Join(errs...)
	if err != nil {
		panic(errors.Wrap(err, "modelmapper.MustCopy() failed"))
	}
}

// Clone returns a deep clone of given object
func Clone[T interface{}](src T) (T, error) {
	clone, err := model.Clone(src)
	return clone.(T), errors.Wrap(err, fmt.Sprintf("modelmapper.Clone[%T]() failed", src))
}

// ToMap deeply converts a struct into a map[string]any
func ToMap(src any) (map[string]any, error) {
	outputMap, err := model.Map(src)
	return outputMap, errors.Wrap(err, "modelmapper.ToMap() failed")
}

var sanitizer, _ = sanitize.New()

func Sanitize(target any) {
	sanitizer.Sanitize(target)
}

func init() {
	AddConversion[map[string]string, *map[string]string](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf((*map[string]string)(nil)), nil
		}

		result := in.Interface().(map[string]string)
		return reflect.ValueOf(&result), nil
	})

	AddConversion[*map[string]string, map[string]string](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf((map[string]string)(nil)), nil
		}

		result := *in.Interface().(*map[string]string)
		return reflect.ValueOf(result), nil
	})

	AddConversion[string, *string](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(string)
		return reflect.ValueOf(&result), nil
	})

	AddConversion[*string, string](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf(""), nil
		}

		result := *in.Interface().(*string)
		return reflect.ValueOf(result), nil
	})

	AddConversion[int, *int](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(int)
		return reflect.ValueOf(&result), nil
	})

	AddConversion[*int, int](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf(0), nil
		}

		result := *in.Interface().(*int)
		return reflect.ValueOf(result), nil
	})

	AddConversion[uint, *uint](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(uint)
		return reflect.ValueOf(&result), nil
	})

	AddConversion[*uint, uint](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf(uint(0)), nil
		}

		result := *in.Interface().(*int)
		return reflect.ValueOf(result), nil
	})

	AddConversion[int64, *int64](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(int64)
		return reflect.ValueOf(&result), nil
	})

	AddConversion[*int64, int64](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf(int64(0)), nil
		}

		result := *in.Interface().(*int64)
		return reflect.ValueOf(result), nil
	})

	AddConversion[bool, *bool](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(bool)
		return reflect.ValueOf(&result), nil
	})

	AddConversion[*bool, bool](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf(false), nil
		}

		result := *in.Interface().(*bool)
		return reflect.ValueOf(result), nil
	})

	AddConversion[time.Time, int64](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(time.Time)
		millis := result.UnixMilli()
		return reflect.ValueOf(millis), nil
	})

	AddConversion[time.Time, *int64](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(time.Time)
		millis := result.UnixMilli()
		return reflect.ValueOf(&millis), nil
	})

	AddConversion[*time.Time, int64](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf(int64(0)), nil
		}
		result := in.Interface().(*time.Time)
		millis := result.UnixMilli()
		return reflect.ValueOf(millis), nil
	})

	AddConversion[*time.Time, *int64](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf((*int64)(nil)), nil
		}
		result := in.Interface().(*time.Time)
		millis := result.UnixMilli()
		return reflect.ValueOf(&millis), nil
	})

	AddConversion[time.Time, *time.Time](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(time.Time)
		return reflect.ValueOf(&result), nil
	})

	AddConversion[*time.Time, time.Time](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf(&time.Time{}), nil
		}
		result := in.Interface().(*time.Time)
		return reflect.ValueOf(*result), nil
	})
}
