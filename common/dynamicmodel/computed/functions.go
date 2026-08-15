package computed

import (
	"go.bryk.io/pkg/errors"
)

// Function is one whitelisted operation callable from a computed expression. Every entry is Go
// code baked into the binary: there is no mechanism that maps a schema-supplied string to
// anything but one of these fixed entries, which is what makes the DSL safe by construction.
type Function struct {
	Name string
	// MinArgs / MaxArgs bound the argument count; MaxArgs -1 means unbounded (e.g. concat).
	MinArgs int
	MaxArgs int
	// ReturnType infers the result type from the argument types, and rejects invalid argument
	// types. Runs at schema validation time, never per row.
	ReturnType func(args []Type) (Type, error)
	// Eval computes the value from already-evaluated, normalized argument values. Each function
	// decides its own nil handling: most propagate nil, the null-handling family consumes it.
	Eval func(args []any) (any, error)
}

// registry is the fixed function whitelist. It is populated once by newBuiltinRegistry and never
// mutated afterwards; there is deliberately no public registration API (spec §51.7).
var registry = newBuiltinRegistry()

// LookupFunction finds a registered function by name.
func LookupFunction(name string) (*Function, error) {
	fn, ok := registry[name]
	if !ok {
		return nil, errors.Errorf("Function %q is not registered for computed expressions", name)
	}
	return fn, nil
}

// FunctionNames lists the registered function names, for error messages and docs.
func FunctionNames() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

func newBuiltinRegistry() map[string]*Function {
	all := make(map[string]*Function)
	for _, fn := range builtinFunctions() {
		all[fn.Name] = fn
	}
	return all
}

func builtinFunctions() []*Function {
	var all []*Function
	all = append(all, stringFunctions()...)
	all = append(all, numericFunctions()...)
	all = append(all, nullFunctions()...)
	all = append(all, dateTimeFunctions()...)
	return all
}

// checkArgCount validates arity for a function call; used by both validation and defensive
// evaluation paths.
func checkArgCount(fn *Function, count int) error {
	if count < fn.MinArgs || (fn.MaxArgs >= 0 && count > fn.MaxArgs) {
		return errors.Errorf(
			"function %q expects between %d and %d arguments but received %d",
			fn.Name, fn.MinArgs, fn.MaxArgs, count)
	}
	return nil
}

func anyNil(args []any) bool {
	for _, arg := range args {
		if arg == nil {
			return true
		}
	}
	return false
}

func requireTexty(fnName string, args []Type) error {
	for _, arg := range args {
		if arg != TypeNull && !arg.IsTexty() {
			return errors.Errorf("function %q expects string arguments but received %s", fnName, arg)
		}
	}
	return nil
}

func requireNumeric(fnName string, arg Type) error {
	if arg != TypeNull && !arg.IsNumeric() {
		return errors.Errorf("function %q expects a numeric argument but received %s", fnName, arg)
	}
	return nil
}
