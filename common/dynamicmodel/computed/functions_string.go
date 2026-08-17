package computed

import (
	"strings"
)

// String functions. All propagate nil: a nil operand makes the result nil, matching SQL NULL
// semantics (spec §33). A caller wanting different behavior wraps operands in coalesce.

func stringFunctions() []*Function {
	return []*Function{
		concatFunction(),
		unaryStringFunction("lower", strings.ToLower),
		unaryStringFunction("upper", strings.ToUpper),
		unaryStringFunction("trim", strings.TrimSpace),
		lengthFunction(),
	}
}

func concatFunction() *Function {
	return &Function{
		Name:    "concat",
		MinArgs: 2,
		MaxArgs: -1,
		ReturnType: func(args []Type) (Type, error) {
			return TypeString, requireTexty("concat", args)
		},
		Eval: func(args []any) (any, error) {
			if anyNil(args) {
				return nil, nil
			}
			var sb strings.Builder
			for _, arg := range args {
				part, err := coerceString(arg)
				if err != nil {
					return nil, err
				}
				sb.WriteString(part)
			}
			return sb.String(), nil
		},
	}
}

func unaryStringFunction(name string, apply func(string) string) *Function {
	return &Function{
		Name:    name,
		MinArgs: 1,
		MaxArgs: 1,
		ReturnType: func(args []Type) (Type, error) {
			return TypeString, requireTexty(name, args)
		},
		Eval: func(args []any) (any, error) {
			if anyNil(args) {
				return nil, nil
			}
			text, err := coerceString(args[0])
			if err != nil {
				return nil, err
			}
			return apply(text), nil
		},
	}
}

func lengthFunction() *Function {
	return &Function{
		Name:    "length",
		MinArgs: 1,
		MaxArgs: 1,
		ReturnType: func(args []Type) (Type, error) {
			return TypeInt32, requireTexty("length", args)
		},
		Eval: func(args []any) (any, error) {
			if anyNil(args) {
				return nil, nil
			}
			text, err := coerceString(args[0])
			if err != nil {
				return nil, err
			}
			return int32(len([]rune(text))), nil
		},
	}
}
