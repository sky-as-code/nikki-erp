// Package dynamicengines declares the resource engines the Inventory module serves through the
// dynamic resource engine, and creates them during the module's Init().
//
// Its imports point one way only: the domain (models and services), the module's own interfaces,
// and the dynamicresource module — never app/, infra/ or transport/. That keeps the package
// importable by both inventory (which creates the engines) and inventory/transport/restful
// (which registers their routes) without a cycle.
//
// The package declares engines and adapts their callbacks; the rules those callbacks enforce
// live in domain/services. See docs/wiki/07. ERP backend module.md §6.7.
package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// engineSpec declares one resource engine the Inventory module owns.
type engineSpec struct {
	// SchemaName is the dynamic-model schema the engine serves. It must be an XSchemaName
	// constant, never a string derived from the resource path.
	SchemaName string

	// DefaultFields is the field set a listing search returns. Primary key fields are always
	// included by the query builder, so listing them here is redundant.
	DefaultFields []string

	// DefineActions adds resource-specific actions and validation on top of the built-in CRUD
	// ones. It is optional: a resource without custom behavior leaves it nil.
	DefineActions func(drif.DynamicResourceEngine) error
}

// engineSpecs lists the resources Inventory serves through the dynamic resource engine, each
// with the field set its listing UI needs.
var engineSpecs = []engineSpec{
	productTypeEngineSpec(),
	productCategoryEngineSpec(),
	brandEngineSpec(),
	productAttributeEngineSpec(),
	productAttributeValueEngineSpec(),
	productTemplateEngineSpec(),
	productTemplateAttributeEngineSpec(),
	productTemplateAttributeValueEngineSpec(),
	productVariantEngineSpec(),
	productVariantAttributeValueEngineSpec(),
}

// EngineSchemaNames lists the schemas Inventory creates an engine for, so that route
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
	engine, err := dynamicresource.Registry().NewEngine(spec.SchemaName, drif.NewEngineOptions{
		DefaultSearchFields: spec.DefaultFields,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create the '%s' resource engine", spec.SchemaName)
	}

	if spec.DefineActions != nil {
		if err := spec.DefineActions(engine); err != nil {
			return errors.Wrapf(err, "failed to define actions of the '%s' resource engine", spec.SchemaName)
		}
	}

	// After DefineActions, because ModifyAction replaces ValidateExtra rather than chaining it:
	// attaching the guard first would let a spec's own create validation silently drop it.
	if err := rejectArchivedOnCreate(engine); err != nil {
		return errors.Wrapf(err, "failed to attach the create guard of the '%s' resource engine", spec.SchemaName)
	}

	err = deps.RegisterNamed(
		dynamicresource.EngineDependencyName(spec.SchemaName),
		func() drif.DynamicResourceEngine { return engine },
	)
	return errors.Wrapf(err, "failed to register the '%s' resource engine", spec.SchemaName)
}
