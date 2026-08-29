package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Inventory Location is the module's canonical location resource, shared by Warehouse and Stock
// rather than owned by either: Warehouse configures it, Stock only references its id.
//
// Beyond CRUD it has its own lifecycle: suspend takes a location out of use while leaving whatever
// it holds where it is, and move re-parents it and rewrites the cached paths beneath. Archiving is
// the built-in action, guarded in the domain service.

func inventoryLocationEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.InventoryLocationSchemaName,
		DefineActions: defineInventoryLocationActions,
	}
}
