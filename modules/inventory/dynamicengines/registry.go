// Package dynamicengines declares the resource onions the Inventory module serves through the
// composable resource engine, and registers them into the dependency container during the
// module's Init().
//
// Each resource file wires the module's own repository, domain service and application service
// onto the composable defaults, publishes those typed layers, and installs the built onion into
// the domain services' resource hub so peer services can reach it at call time.
package dynamicengines

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
)

// InitDynamicEngines registers every resource onion this module owns. Nothing is built here: an
// onion is a container constructor, resolved the first time a consumer asks for it.
func InitDynamicEngines() error {
	return stdErr.Join(
		registerProductTypeEngine(),
		registerProductCategoryEngine(),
		registerBrandEngine(),
		registerProductAttributeEngine(),
		registerProductAttributeValueEngine(),
		registerProductTemplateEngine(),
		registerProductTemplateAttributeEngine(),
		registerProductTemplateAttributeValueEngine(),
		registerProductVariantEngine(),
		registerProductVariantAttributeValueEngine(),
		registerWarehouseEngine(),
		registerStorageCategoryEngine(),
		registerInventoryLocationEngine(),
		registerWarehouseSupplyRelationEngine(),
		registerPutawayRuleEngine(),
		registerStockOperationTypeEngine(),
		registerStockQuantEngine(),
		registerStockTransferEngine(),
		registerStockMoveEngine(),
		registerStockMoveLineEngine(),
		registerStockMoveDependencyEngine(),
		registerStockScrapEngine(),
		registerStockProductConfigEngine(),
		registerWarehouseProductGuardEngine(),
		registerStockReservationEngine(),
		registerInventoryIntegrationOutboxEngine(),
	)
}

// readOnlyCrudActions withholds the built-in writes from a resource whose rows are derived or
// engine-written. Bulk import is withheld too: importing a reservation or an outbox row would be
// a write by another name.
func readOnlyCrudActions() []composable.CrudAction {
	return []composable.CrudAction{
		composable.CrudActionGetById,
		composable.CrudActionGetByUnique,
		composable.CrudActionSearch,
		composable.CrudActionExists,
		composable.CrudActionGetSchema,
		composable.CrudActionComputeField,
	}
}

// buildOnion builds one onion with the module-wide rules applied and installs it into the
// resource hub. Every Inventory resource refuses is_archived on create: archiving goes through
// the /archived action, and the guard runs ahead of any rule a derived service adds.
func buildOnion(impl *composable.DynamicResourceEngineOnionImpl, param composable.BuildParam) composable.DynamicResourceEngineOnion {
	impl.RejectArchivedOnCreate = true
	onion := composable.MustBuild(impl, param)
	services.InstallResource(impl.SchemaName, onion.Repository(), onion.DomainService())
	return onion
}

// SchemaNames lists every resource this module declares, for the boot-time check that each was
// built.
func SchemaNames() []string {
	return []string{
		models.ProductTypeSchemaName, models.ProductCategorySchemaName, models.BrandSchemaName,
		models.ProductAttributeSchemaName, models.ProductAttributeValueSchemaName,
		models.ProductTemplateSchemaName, models.ProductTemplateAttributeSchemaName,
		models.ProductTemplateAttributeValueSchemaName, models.ProductVariantSchemaName,
		models.ProductVariantAttributeValueSchemaName, models.WarehouseSchemaName,
		models.StorageCategorySchemaName, models.InventoryLocationSchemaName,
		models.WarehouseSupplyRelationSchemaName, models.PutawayRuleSchemaName,
		models.StockOperationTypeSchemaName, models.StockQuantSchemaName, models.StockTransferSchemaName,
		models.StockMoveSchemaName, models.StockMoveLineSchemaName, models.StockMoveDependencySchemaName,
		models.StockScrapSchemaName, models.StockProductConfigSchemaName,
		models.WarehouseProductGuardSchemaName, models.StockReservationSchemaName,
		models.InventoryIntegrationOutboxSchemaName,
	}
}
