package computed

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"
)

// Row values arrive from the repository as a mix of values, pointers and driver types. The
// helpers below normalize them for evaluation. They are conversion-only: type CORRECTNESS is the
// job of schema validation and inference, so a coercion failure here is reported as an error
// rather than silently producing a zero value.

// normalizeValue unwraps non-nil pointers and maps typed nils to untyped nil, so evaluation only
// ever sees plain values or nil.
func normalizeValue(value any) any {
	if value == nil {
		return nil
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		return normalizeValue(rv.Elem().Interface())
	}
	return value
}

func isNilValue(value any) bool {
	return normalizeValue(value) == nil
}

// isIntegerValue reports whether the normalized value is a Go integer kind (not decimal/float).
func isIntegerValue(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func coerceDecimal(value any) (decimal.Decimal, error) {
	switch v := value.(type) {
	case decimal.Decimal:
		return v, nil
	case int:
		return decimal.NewFromInt(int64(v)), nil
	case int32:
		return decimal.NewFromInt32(v), nil
	case int64:
		return decimal.NewFromInt(v), nil
	case float64:
		return decimal.NewFromFloat(v), nil
	case float32:
		return decimal.NewFromFloat32(v), nil
	case string:
		parsed, err := decimal.NewFromString(v)
		return parsed, errors.WithStack(err)
	case json.Number:
		parsed, err := decimal.NewFromString(v.String())
		return parsed, errors.WithStack(err)
	}
	return decimal.Zero, errors.Errorf("value %v (%T) is not numeric", value, value)
}

func coerceInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	}
	dec, err := coerceDecimal(value)
	if err != nil {
		return 0, err
	}
	if !dec.IsInteger() {
		return 0, errors.Errorf("value %v is not an integer", value)
	}
	return dec.IntPart(), nil
}

func coerceString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case fmt.Stringer:
		return v.String(), nil
	}
	return "", errors.Errorf("value %v (%T) is not a string", value, value)
}

func coerceBool(value any) (bool, error) {
	if v, ok := value.(bool); ok {
		return v, nil
	}
	return false, errors.Errorf("value %v (%T) is not a boolean", value, value)
}

func coerceTime(value any) (time.Time, error) {
	if v, ok := value.(time.Time); ok {
		return v, nil
	}
	return time.Time{}, errors.Errorf("value %v (%T) is not a date/time", value, value)
}
