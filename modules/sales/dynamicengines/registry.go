// Package dynamicengines declares the resource engines the Sales module serves, and creates them
// during Init(). It may import only the domain, the module's own interfaces and dynamicresource —
// never app/, infra/ or transport/ — so both sales and sales/transport/restful can import it without
// a cycle. The rules its callbacks enforce live in domain/services.
package dynamicengines

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// junctionSchemas lists association schemas that need an engine built but not served: a schema here
// gets a repository and nothing else — no route, no IAM resource row, no CRUD. They need an engine
// only because a repository cannot be built without one (the query builder and database client are
// private to the registry). crud.ManageM2m is not usable here because it resolves the far side
// against a locally registered schema, and sales_channel_payment_rel points at a paymentinvoice
// payment method.
var junctionSchemas = []string{
	models.SalesChannelPaymentRelSchemaName,
	models.SalesChannelFulfillmentMethodSchemaName,
}

// EngineSchemaNames keeps route registration and engine creation from drifting apart.
// EngineSchemaNames answers every schema Sales serves. Kept as a name distinct from SchemaNames
// because the module's surface tests are written against it; both now answer the same list, the
// legacy generation having been removed.
func EngineSchemaNames() []string {
	return SchemaNames()
}

// InitDynamicEngines creates this module's engines and publishes them into the dependency container
// so other modules can inject them by name.
// InitDynamicEngines registers every resource this module serves, and the association
// repositories. Nothing is built here: an onion is a container constructor, and BuildAllEngines
// forces the construction once the application services it may inject are registered.
func InitDynamicEngines() error {
	return stdErr.Join(
		registerSalesFulfillmentMethodEngine(),
		registerSalesChannelEngine(),
		registerSalesPointEngine(),
		registerSalesPricelistEngine(),
		registerSalesPricelistItemEngine(),
		registerSalesComboEngine(),
		registerSalesComboComponentEngine(),
		registerSalesOrderEngine(),
		registerSalesOrderLineEngine(),
		registerSalesOrderLineComponentEngine(),
		registerSalesOrderAdjustmentEngine(),
		registerSalesOrderEventEngine(),
		registerSalesBillEngine(),
		registerSalesBillLineEngine(),
		registerSalesBillRelationEngine(),
		registerSalesPaymentEngine(),
		registerSalesFulfillmentRequestEngine(),
		registerSalesFulfillmentRequestLineEngine(),
		registerSalesOrderFulfillmentEngine(),
		registerSalesOrderFulfillmentItemEngine(),
		registerSalesFulfillmentAttemptEngine(),
		registerSalesFulfillmentAttemptItemEngine(),
		registerSalesFulfillmentTargetChangeEngine(),
		registerSalesReturnEngine(),
		registerSalesReturnLineEngine(),
		registerSalesRefundPaymentEngine(),
		registerSalesPromotionProgramEngine(),
		registerSalesPromotionConditionGroupEngine(),
		registerSalesPromotionConditionEngine(),
		registerSalesPromotionConditionTargetEngine(),
		registerSalesPromotionRewardEngine(),
		registerSalesPromotionCompatibilityEngine(),
		registerSalesVoucherCodeEngine(),
		registerSalesVoucherRedemptionEngine(),
		registerSalesQuotationEngine(),
		registerSalesQuotationLineEngine(),
		registerSalesFiscalRequestEngine(),
		registerSalesBillingInstructionEngine(),
		registerSalesBillingIssuanceAttemptEngine(),
		registerSalesManualDiscountEngine(),
		registerSalesIntegrationOutboxEngine(),
		InitJunctionRepositories(),
	)
}

// buildOnion builds one onion with the module-wide rules applied and installs it into the resource
// hub. Every Sales resource refuses is_archived on create: archiving goes through the dedicated
// actions, and no schema declares the field itself (domain/models/archivable_test.go holds that).
func buildOnion(
	impl *composable.DynamicResourceEngineOnionImpl, param composable.BuildParam,
) composable.DynamicResourceEngineOnion {
	impl.RejectArchivedOnCreate = true
	onion := composable.MustBuild(impl, param)
	services.InstallResource(impl.SchemaName, onion.Repository(), onion.DomainService())
	return onion
}

// SchemaNames lists the resources served by a composable onion, for the boot-time check that each
// was built. It excludes the junctions, which have no onion.
func SchemaNames() []string {
	return []string{
		models.SalesFulfillmentMethodSchemaName,
		models.SalesChannelSchemaName,
		models.SalesPointSchemaName,
		models.SalesPricelistSchemaName,
		models.SalesPricelistItemSchemaName,
		models.SalesComboSchemaName,
		models.SalesComboComponentSchemaName,
		models.SalesOrderSchemaName,
		models.SalesOrderLineSchemaName,
		models.SalesOrderLineComponentSchemaName,
		models.SalesOrderAdjustmentSchemaName,
		models.SalesOrderEventSchemaName,
		models.SalesBillSchemaName,
		models.SalesBillLineSchemaName,
		models.SalesBillRelationSchemaName,
		models.SalesPaymentSchemaName,
		models.SalesFulfillmentRequestSchemaName,
		models.SalesFulfillmentRequestLineSchemaName,
		models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentItemSchemaName,
		models.SalesFulfillmentAttemptSchemaName,
		models.SalesFulfillmentAttemptItemSchemaName,
		models.SalesFulfillmentTargetChangeSchemaName,
		models.SalesReturnSchemaName,
		models.SalesReturnLineSchemaName,
		models.SalesRefundPaymentSchemaName,
		models.SalesPromotionProgramSchemaName,
		models.SalesPromotionConditionGroupSchemaName,
		models.SalesPromotionConditionSchemaName,
		models.SalesPromotionConditionTargetSchemaName,
		models.SalesPromotionRewardSchemaName,
		models.SalesPromotionCompatibilitySchemaName,
		models.SalesVoucherCodeSchemaName,
		models.SalesVoucherRedemptionSchemaName,
		models.SalesQuotationSchemaName,
		models.SalesQuotationLineSchemaName,
		models.SalesFiscalRequestSchemaName,
		models.SalesBillingInstructionSchemaName,
		models.SalesBillingIssuanceAttemptSchemaName,
		models.SalesManualDiscountSchemaName,
		models.SalesIntegrationOutboxSchemaName,
	}
}
