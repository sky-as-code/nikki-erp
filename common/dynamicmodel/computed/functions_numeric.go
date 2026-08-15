package computed

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"
)

// Numeric functions. All propagate nil. Each returns the same numeric family it was given: an
// integer argument yields an integer, a decimal yields a decimal — so rounding an int is the
// identity, exactly as in SQL.

func numericFunctions() []*Function {
	return []*Function{
		roundFunction(),
		unaryNumericFunction("ceil", func(d decimal.Decimal) decimal.Decimal { return d.Ceil() }),
		unaryNumericFunction("floor", func(d decimal.Decimal) decimal.Decimal { return d.Floor() }),
		unaryNumericFunction("abs", func(d decimal.Decimal) decimal.Decimal { return d.Abs() }),
	}
}

func roundFunction() *Function {
	return &Function{
		Name:    "round",
		MinArgs: 1,
		MaxArgs: 2,
		ReturnType: func(args []Type) (Type, error) {
			if err := requireNumeric("round", args[0]); err != nil {
				return TypeUnknown, err
			}
			if len(args) == 2 && args[1] != TypeNull && args[1] != TypeInt32 && args[1] != TypeInt64 {
				return TypeUnknown, errors.Errorf(
					"function \"round\" expects an integer precision but received %s", args[1])
			}
			return numericPassthroughType(args[0]), nil
		},
		Eval: evalRound,
	}
}

func evalRound(args []any) (any, error) {
	if anyNil(args) {
		return nil, nil
	}
	precision := int64(0)
	if len(args) == 2 {
		parsed, err := coerceInt64(args[1])
		if err != nil {
			return nil, err
		}
		precision = parsed
	}
	value, err := coerceDecimal(args[0])
	if err != nil {
		return nil, err
	}
	rounded := value.Round(int32(precision))
	return numericResultLike(args[0], rounded)
}

func unaryNumericFunction(name string, apply func(decimal.Decimal) decimal.Decimal) *Function {
	return &Function{
		Name:    name,
		MinArgs: 1,
		MaxArgs: 1,
		ReturnType: func(args []Type) (Type, error) {
			return numericPassthroughType(args[0]), requireNumeric(name, args[0])
		},
		Eval: func(args []any) (any, error) {
			if anyNil(args) {
				return nil, nil
			}
			value, err := coerceDecimal(args[0])
			if err != nil {
				return nil, err
			}
			return numericResultLike(args[0], apply(value))
		},
	}
}

// numericPassthroughType keeps the argument's own numeric type; a null argument reads as decimal
// so the expression still has a definite type.
func numericPassthroughType(arg Type) Type {
	if arg == TypeNull {
		return TypeDecimal
	}
	return arg
}

// numericResultLike converts a decimal result back to the integer kind of the original operand,
// so integer-typed fields keep producing integer values through these functions.
func numericResultLike(original any, result decimal.Decimal) (any, error) {
	if isIntegerValue(original) && result.IsInteger() {
		return result.IntPart(), nil
	}
	return result, nil
}
