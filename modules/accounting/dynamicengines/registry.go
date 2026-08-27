// Package dynamicengines declares the resource engines the Accounting module serves through the
// dynamic resource engine, and creates them during the module's Init().
//
// It is deliberately a leaf package: it imports the domain models and the dynamicresource module,
// but nothing else from accounting. That lets both accounting (which creates the engines) and
// accounting/transport/restful (which registers their routes) import it without a cycle.
package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// engineSpec declares one resource engine the Accounting module owns.
type engineSpec struct {
	// SchemaName is the dynamic-model schema the engine serves. It must be an XSchemaName
	// constant, never a string derived from the resource path.
	SchemaName string

	// DefineActions adds resource-specific actions and validation on top of the built-in CRUD
	// ones. It is optional: a resource without custom behavior leaves it nil.
	DefineActions func(drif.DynamicResourceEngine) error
}

// engineSpecs lists the resources Accounting serves through the dynamic resource engine.
//
// The order mirrors the dependency order of the schemas themselves, so that a reader looking for
// what references what can follow the list top to bottom.
var engineSpecs = []engineSpec{
	taxJurisdictionEngineSpec(),
	taxGroupEngineSpec(),
	taxRoundingPolicyEngineSpec(),
	taxProductClassificationEngineSpec(),
	taxEngineSpec(),
	taxDefinitionVersionEngineSpec(),
	taxRateVersionEngineSpec(),
	taxComponentEngineSpec(),
	taxMappingEngineSpec(),
	taxMappingLineEngineSpec(),
	taxRuleEngineSpec(),
	taxRuleConditionEngineSpec(),
	taxRuleResultEngineSpec(),
}

// EngineSchemaNames lists the schemas Accounting creates an engine for, so that route
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
