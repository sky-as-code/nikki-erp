// Package dynamicengines declares the resource engines Accounting serves through the dynamic
// resource engine, and creates them during Init().
//
// It must stay a leaf package importing nothing else from accounting, so that both accounting and
// accounting/transport/restful can import it without a cycle.
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
	// SchemaName must be an XSchemaName constant, never a string derived from the resource path.
	SchemaName string

	// DefineActions adds resource-specific actions and validation on top of the built-in CRUD
	// ones. Optional: nil for a resource without custom behavior.
	DefineActions func(drif.DynamicResourceEngine) error
}

// engineSpecs lists the resources Accounting serves, in the dependency order of the schemas.
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

// EngineSchemaNames keeps route registration and engine creation from drifting apart.
func EngineSchemaNames() []string {
	return array.Map(engineSpecs, func(spec engineSpec) string {
		return spec.SchemaName
	})
}

// InitDynamicEngines creates this module's resource engines and publishes them into the dependency
// container.
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
