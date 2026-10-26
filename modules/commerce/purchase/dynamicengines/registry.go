// Package dynamicengines registers the composable onions of the Purchase resources and publishes
// their typed layers into the dependency container. Each onion is a container constructor that
// receives the ports its domain service validates against; Init only registers, and
// BuildAllEngines forces the construction once every port is bound.
package dynamicengines

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// InitDynamicEngines registers every resource this module serves. It must run after
// infra/external has bound the ports the onions inject.
func InitDynamicEngines() error {
	return stdErr.Join(
		registerConfigurationEngine(),
		registerVendorProductPriceEngine(),
		registerSourcingGroupEngine(),
		registerAgreementEngine(),
		registerAgreementLineEngine(),
		registerPurchaseOrderEngine(),
		registerPurchaseOrderLineEngine(),
		registerAuditEventEngine(),
	)
}

// SchemaNames lists the resources served by a composable onion, in RegisterModels order, for route
// registration and the boot-time check that each was built.
func SchemaNames() []string {
	return []string{
		models.ConfigurationSchemaName,
		models.VendorProductPriceSchemaName,
		models.SourcingGroupSchemaName,
		models.AgreementSchemaName,
		models.AgreementLineSchemaName,
		models.PurchaseOrderSchemaName,
		models.PurchaseOrderLineSchemaName,
		models.AuditEventSchemaName,
	}
}

// buildOnion builds one onion and installs it into the resource hub, through which the lifecycle
// operations, the totals recomputation and the usage checker reach the repositories.
func buildOnion(
	impl *composable.DynamicResourceEngineOnionImpl, param composable.BuildParam,
) composable.DynamicResourceEngineOnion {
	onion := composable.MustBuild(impl, param)
	services.InstallResource(impl.SchemaName, onion.Repository(), onion.DomainService())
	return onion
}
