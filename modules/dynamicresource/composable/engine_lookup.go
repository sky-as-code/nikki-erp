package composable

import (
	"fmt"
	"reflect"

	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// ResolveEngines returns the onions registered under dynengine_{schema} for each schema name, in
// order. It is for code that works over a list of resources rather than naming each typed layer,
// such as an org cleanup; a resource that needs its own rules is injected by type instead.
//
// dig resolves a named value only through a struct tag, so the parameter struct is built at run
// time with one tagged field per schema.
func ResolveEngines(schemaNames ...string) ([]DynamicResourceEngineOnion, error) {
	engineType := reflect.TypeOf((*DynamicResourceEngineOnion)(nil)).Elem()
	fields := []reflect.StructField{{Name: "In", Type: reflect.TypeOf(dig.In{}), Anonymous: true}}
	for i, name := range schemaNames {
		fields = append(fields, reflect.StructField{
			Name: fmt.Sprintf("Engine%d", i),
			Type: engineType,
			Tag:  reflect.StructTag(fmt.Sprintf(`name:"%s"`, EngineDependencyName(name))),
		})
	}
	paramType := reflect.StructOf(fields)

	engines := make([]DynamicResourceEngineOnion, len(schemaNames))
	invoker := reflect.MakeFunc(
		reflect.FuncOf([]reflect.Type{paramType}, nil, false),
		func(args []reflect.Value) []reflect.Value {
			for i := range schemaNames {
				engines[i] = args[0].Field(i + 1).Interface().(DynamicResourceEngineOnion)
			}
			return nil
		},
	)
	if err := deps.Invoke(invoker.Interface()); err != nil {
		return nil, errors.Wrap(err, "resolve resource engines")
	}
	return engines, nil
}

// ResolveDomainServices is ResolveEngines narrowed to each onion's domain service.
func ResolveDomainServices(schemaNames ...string) ([]CrudDomainService, error) {
	engines, err := ResolveEngines(schemaNames...)
	if err != nil {
		return nil, err
	}
	svcs := make([]CrudDomainService, len(engines))
	for i, engine := range engines {
		svcs[i] = engine.DomainService()
	}
	return svcs, nil
}
