package computed

import (
	"go.bryk.io/pkg/errors"
)

// Null-handling functions — the only family that consumes nil instead of propagating it.

func nullFunctions() []*Function {
	return []*Function{
		coalesceFunction(),
		nullifFunction(),
	}
}

func coalesceFunction() *Function {
	return &Function{
		Name:    "coalesce",
		MinArgs: 2,
		MaxArgs: -1,
		ReturnType: func(args []Type) (Type, error) {
			return unifyAll("coalesce", args)
		},
		Eval: func(args []any) (any, error) {
			for _, arg := range args {
				if arg != nil {
					return arg, nil
				}
			}
			return nil, nil
		},
	}
}

func nullifFunction() *Function {
	return &Function{
		Name:    "nullif",
		MinArgs: 2,
		MaxArgs: 2,
		ReturnType: func(args []Type) (Type, error) {
			if !args[0].ComparableWith(args[1]) {
				return TypeUnknown, errors.Errorf(
					"function \"nullif\" expects comparable arguments but received %s and %s", args[0], args[1])
			}
			return args[0], nil
		},
		Eval: evalNullif,
	}
}

func evalNullif(args []any) (any, error) {
	if args[0] == nil {
		return nil, nil
	}
	if args[1] == nil {
		return args[0], nil
	}
	equal, err := compareEquality(args[0], args[1])
	if err != nil {
		return nil, err
	}
	if equal {
		return nil, nil
	}
	return args[0], nil
}

func unifyAll(fnName string, args []Type) (Type, error) {
	result := TypeNull
	for _, arg := range args {
		unified, ok := unify(result, arg)
		if !ok {
			return TypeUnknown, errors.Errorf(
				"function %q cannot unify argument types %s and %s", fnName, result, arg)
		}
		result = unified
	}
	if result == TypeNull {
		return TypeUnknown, errors.Errorf("function %q requires at least one non-null argument type", fnName)
	}
	return result, nil
}
