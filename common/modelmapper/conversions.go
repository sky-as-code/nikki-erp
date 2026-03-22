package modelmapper

import (
	"reflect"
	"time"
)

func init() {
	AddConversion[time.Time, int64](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(time.Time)
		millis := result.UnixMilli()
		return reflect.ValueOf(millis), nil
	})
	AddConversion[time.Time, string](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(time.Time)
		return reflect.ValueOf(result.Format(time.RFC3339)), nil
	})
	AddConversion[string, time.Time](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(string)
		parsed, err := time.Parse(time.RFC3339, result)
		if err != nil {
			return reflect.ValueOf(time.Time{}), err
		}
		return reflect.ValueOf(parsed), nil
	})
}
