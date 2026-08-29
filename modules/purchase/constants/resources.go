package constants

import "github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"

// Resource codes for authorization. Each must be byte-identical to its schema name, because the
// engine derives the code it asserts against from the schema name; a drifted code denies every
// request with nothing in the 403 to explain it. Aliasing the model constants keeps them identical.
const (
	PurchaseConfigurationResource = models.ConfigurationSchemaName
	PurchaseSourcingGroupResource = models.SourcingGroupSchemaName
	PurchaseAgreementResource     = models.AgreementSchemaName
	PurchaseAgreementLineResource = models.AgreementLineSchemaName
	PurchaseOrderResource         = models.PurchaseOrderSchemaName
	PurchaseOrderLineResource     = models.PurchaseOrderLineSchemaName
	PurchaseAuditEventResource    = models.AuditEventSchemaName
	PurchaseVendorPriceResource   = models.VendorProductPriceSchemaName
)
