// Package dynamicengines declares the resource engines the Purchase module serves, and creates them
// during Init(). It may import only the domain, the module's own interfaces and dynamicresource —
// never app/, infra/ or transport/ — so that both purchase and purchase/transport/restful can
// import it without a cycle. The rules its callbacks enforce live in domain/services.
package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// engineSpec declares one resource engine the Purchase module owns.
type engineSpec struct {
	// SchemaName must be an XSchemaName constant, never a string derived from the resource path.
	SchemaName string


	// DefineActions is optional; a resource without custom behavior leaves it nil.
	DefineActions func(drif.DynamicResourceEngine) error
}

// engineSpecs lists the resources this module serves. The order matches RegisterModels for
// readability only; engines are created after every schema is registered.
var engineSpecs = []engineSpec{
	configurationEngineSpec(),
	vendorProductPriceEngineSpec(),
	sourcingGroupEngineSpec(),
	agreementEngineSpec(),
	agreementLineEngineSpec(),
	purchaseOrderEngineSpec(),
	purchaseOrderLineEngineSpec(),
	auditEventEngineSpec(),
}

func configurationEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ConfigurationSchemaName,
	}
}

func vendorProductPriceEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.VendorProductPriceSchemaName,
		DefineActions: defineVendorProductPriceGuards,
	}
}

func sourcingGroupEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.SourcingGroupSchemaName,
		DefineActions: defineSourcingGroupGuards,
	}
}

func agreementEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.AgreementSchemaName,
		DefineActions: defineAgreementActions,
	}
}

func agreementLineEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.AgreementLineSchemaName,
	}
}

func purchaseOrderEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.PurchaseOrderSchemaName,
		DefineActions: defineOrderActions,
	}
}

func purchaseOrderLineEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.PurchaseOrderLineSchemaName,
	}
}

func auditEventEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.AuditEventSchemaName,
		DefineActions: defineAuditEventGuards,
	}
}

// EngineSchemaNames lists the schemas this module creates an engine for, so that route
// registration and engine creation cannot drift apart.
func EngineSchemaNames() []string {
	return array.Map(engineSpecs, func(spec engineSpec) string {
		return spec.SchemaName
	})
}

// InitDynamicEngines creates this module's engines and publishes them into the dependency
// container so other modules can inject them by name.
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
