package computed

import (
	"time"

	"go.bryk.io/pkg/errors"
)

// Date/time functions. today/now take zero arguments, so no schema input reaches them at all;
// the rest propagate nil like the string/numeric families.

// extractParts is the fixed whitelist for extract()'s first argument. It is validated as a
// compile-time literal at schema validation, never resolved from row data.
var extractParts = map[string]func(time.Time) int32{
	"year":   func(t time.Time) int32 { return int32(t.Year()) },
	"month":  func(t time.Time) int32 { return int32(t.Month()) },
	"day":    func(t time.Time) int32 { return int32(t.Day()) },
	"hour":   func(t time.Time) int32 { return int32(t.Hour()) },
	"minute": func(t time.Time) int32 { return int32(t.Minute()) },
	"second": func(t time.Time) int32 { return int32(t.Second()) },
	"dow":    func(t time.Time) int32 { return int32(t.Weekday()) },
	"doy":    func(t time.Time) int32 { return int32(t.YearDay()) },
}

// IsValidExtractPart reports whether the literal is an allowed extract() part name.
func IsValidExtractPart(part string) bool {
	_, ok := extractParts[part]
	return ok
}

func dateTimeFunctions() []*Function {
	return []*Function{
		nowFunction("today", TypeDate, func() time.Time {
			return time.Now().UTC().Truncate(24 * time.Hour)
		}),
		nowFunction("now", TypeDateTime, func() time.Time {
			return time.Now().UTC()
		}),
		dateAddFunction(),
		dateDiffFunction(),
		extractFunction(),
	}
}

func nowFunction(name string, resultType Type, clock func() time.Time) *Function {
	return &Function{
		Name:    name,
		MinArgs: 0,
		MaxArgs: 0,
		ReturnType: func(args []Type) (Type, error) {
			return resultType, nil
		},
		Eval: func(args []any) (any, error) {
			return clock(), nil
		},
	}
}

// dateAddFunction shifts a date/datetime by a whole number of days. The interval unit is fixed
// to days in this phase.
func dateAddFunction() *Function {
	return &Function{
		Name:    "date_add",
		MinArgs: 2,
		MaxArgs: 2,
		ReturnType: func(args []Type) (Type, error) {
			if err := requireTemporal("date_add", args[0]); err != nil {
				return TypeUnknown, err
			}
			if args[1] != TypeNull && args[1] != TypeInt32 && args[1] != TypeInt64 {
				return TypeUnknown, errors.Errorf(
					"function \"date_add\" expects an integer day count but received %s", args[1])
			}
			return temporalPassthroughType(args[0]), nil
		},
		Eval: evalDateAdd,
	}
}

func evalDateAdd(args []any) (any, error) {
	if anyNil(args) {
		return nil, nil
	}
	base, err := coerceTime(args[0])
	if err != nil {
		return nil, err
	}
	days, err := coerceInt64(args[1])
	if err != nil {
		return nil, err
	}
	return base.AddDate(0, 0, int(days)), nil
}

// dateDiffFunction gives whole days between two dates: first minus second.
func dateDiffFunction() *Function {
	return &Function{
		Name:    "date_diff",
		MinArgs: 2,
		MaxArgs: 2,
		ReturnType: func(args []Type) (Type, error) {
			if err := requireTemporal("date_diff", args[0]); err != nil {
				return TypeUnknown, err
			}
			return TypeInt32, requireTemporal("date_diff", args[1])
		},
		Eval: evalDateDiff,
	}
}

func evalDateDiff(args []any) (any, error) {
	if anyNil(args) {
		return nil, nil
	}
	first, err := coerceTime(args[0])
	if err != nil {
		return nil, err
	}
	second, err := coerceTime(args[1])
	if err != nil {
		return nil, err
	}
	return int32(first.Sub(second).Hours() / 24), nil
}

func extractFunction() *Function {
	return &Function{
		Name:    "extract",
		MinArgs: 2,
		MaxArgs: 2,
		ReturnType: func(args []Type) (Type, error) {
			if args[0] != TypeString && args[0] != TypeNull {
				return TypeUnknown, errors.Errorf(
					"function \"extract\" expects a literal part name but received %s", args[0])
			}
			return TypeInt32, requireTemporal("extract", args[1])
		},
		Eval: evalExtract,
	}
}

func evalExtract(args []any) (any, error) {
	if anyNil(args) {
		return nil, nil
	}
	part, err := coerceString(args[0])
	if err != nil {
		return nil, err
	}
	extract, ok := extractParts[part]
	if !ok {
		return nil, errors.Errorf("function \"extract\" does not support part %q", part)
	}
	value, err := coerceTime(args[1])
	if err != nil {
		return nil, err
	}
	return extract(value), nil
}

func requireTemporal(fnName string, arg Type) error {
	if arg != TypeNull && !arg.IsTemporal() {
		return errors.Errorf("function %q expects a date/time argument but received %s", fnName, arg)
	}
	return nil
}

func temporalPassthroughType(arg Type) Type {
	if arg == TypeNull {
		return TypeDateTime
	}
	return arg
}
