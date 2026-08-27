// Package dynamicengines declares the resource engines the IAM module serves through
// the dynamic resource engine, and creates them during the module's Init().
//
// It is deliberately a leaf package: it imports the domain models and the dynamicresource
// module, but nothing else from iam. That lets both iam (which creates the engines) and
// iam/transport/restful (which registers their routes) import it without a cycle.
package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	dynengine "github.com/sky-as-code/nikki-erp/modules/dynamicresource/engine"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// engineSpec declares one resource engine the IAM module owns.
type engineSpec struct {
	// SchemaName is the dynamic-model schema the engine serves. It must be an
	// XSchemaName constant, never a string derived from the resource path.
	SchemaName string

	// DefineActions adds resource-specific actions on top of the built-in CRUD ones.
	// It is optional: a resource without custom actions leaves it nil.
	DefineActions func(drif.DynamicResourceEngine) error

	// WithdrawOrgScoping opts the resource out of the engine's default org scoping.
	//
	// Set it only where org_id is *optional* and its absence carries meaning - a NULL org_id
	// marking a global or domain-scoped row. Requiring org_id on such a resource would hide
	// every one of those rows instead of reporting an error. The reason belongs in a comment
	// on the spec that sets it.
	WithdrawOrgScoping bool
}

// engineSpecs lists the resources IAM serves through the dynamic resource engine,
// each with the field set its listing UI needs.
var engineSpecs = []engineSpec{
	userEngineSpec(),
	orgEngineSpec(),
	orgUnitEngineSpec(),
	groupEngineSpec(),
	roleEngineSpec(),
	entitlementEngineSpec(),
	resourceEngineSpec(),
	actionEngineSpec(),
	grantRequestEngineSpec(),
}

// EngineSchemaNames lists the schemas IAM creates an engine for, so that route
// registration and engine creation cannot drift apart.
func EngineSchemaNames() []string {
	return array.Map(engineSpecs, func(spec engineSpec) string {
		return spec.SchemaName
	})
}

// InitDynamicEngines creates the resource engines this module owns and publishes them
// into the dependency container, so that other modules can inject them by name.
//
// The hand-written layers of these resources (application service, domain service,
// repository and REST handlers) are untouched and keep serving their own routes;
// an engine serves the same resource at /v1/iam/{schema_name}.
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

	// After DefineActions, so that the module's own actions are withdrawn too.
	if spec.WithdrawOrgScoping {
		if err := dynengine.WithdrawOrgScoping(engine); err != nil {
			return errors.Wrapf(err, "failed to withdraw org scoping of the '%s' resource engine", spec.SchemaName)
		}
	}

	err = deps.RegisterNamed(
		dynamicresource.EngineDependencyName(spec.SchemaName),
		func() drif.DynamicResourceEngine { return engine },
	)
	return errors.Wrapf(err, "failed to register the '%s' resource engine", spec.SchemaName)
}
