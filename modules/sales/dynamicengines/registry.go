// Package dynamicengines declares the resource engines the Sales module serves through the
// dynamic resource engine, and creates them during the module's Init().
//
// Its imports point one way only: the domain, the module's own interfaces, and the dynamicresource
// module — never app/, infra/ or transport/. That keeps the package importable by both sales
// (which creates the engines) and sales/transport/restful (which registers their routes) without
// a cycle.
//
// The package declares engines and adapts their callbacks; the rules those callbacks enforce live
// in domain/services. See docs/wiki/07. ERP backend module.md §6.7.
package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// engineSpec declares one resource engine the Sales module owns.
type engineSpec struct {
	// SchemaName is the dynamic-model schema the engine serves. It must be an XSchemaName
	// constant, never a string derived from the resource path.
	SchemaName string

	// DefineActions adds resource-specific actions and validation on top of the built-in CRUD
	// ones. It is optional: a resource without custom behavior leaves it nil.
	DefineActions func(drif.DynamicResourceEngine) error
}

// engineSpecs lists the resources this module serves through the dynamic resource engine.
//
// The order matches RegisterModels: referenced before referencing. It does not have to — engines
// are created after every schema is registered — but keeping the two lists in the same order makes
// a missing entry obvious when reading them side by side.
//
// Junction tables never get an entry here. A _rel row is configured through its owner's
// capabilities, so it has no engine, no route and no IAM resource row — see the 25-engines-for-27-
// schemas split in vending_machine_new.
var engineSpecs = []engineSpec{
	salesChannelEngineSpec(),
	salesPointEngineSpec(),
}

func salesChannelEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.SalesChannelSchemaName,
		DefineActions: defineSalesChannelActions,
	}
}

func salesPointEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.SalesPointSchemaName,
		DefineActions: defineSalesPointActions,
	}
}

// EngineSchemaNames lists the schemas this module creates an engine for, so that route
// registration and engine creation cannot drift apart.
func EngineSchemaNames() []string {
	return array.Map(engineSpecs, func(spec engineSpec) string {
		return spec.SchemaName
	})
}

// InitDynamicEngines creates the resource engines this module owns and publishes them into the
// dependency container, so that other modules can inject them by name.
func InitDynamicEngines() error {
	for _, spec := range engineSpecs {
		if err := initEngine(spec); err != nil {
			return err
		}
	}
	return nil
}

func initEngine(spec engineSpec) error {
	engine, err := dynamicresource.Registry().NewEngine(spec.SchemaName, drif.NewEngineOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to create the '%s' resource engine", spec.SchemaName)
	}

	if spec.DefineActions != nil {
		if err := spec.DefineActions(engine); err != nil {
			return errors.Wrapf(err, "failed to define actions of the '%s' resource engine", spec.SchemaName)
		}
	}

	err = deps.RegisterNamed(
		dynamicresource.EngineDependencyName(spec.SchemaName),
		func() drif.DynamicResourceEngine { return engine },
	)
	return errors.Wrapf(err, "failed to register the '%s' resource engine", spec.SchemaName)
}
