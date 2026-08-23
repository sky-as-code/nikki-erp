package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Inventory Location is the canonical location resource of the whole module, shared by Warehouse
// and Stock rather than owned by either. It lives in its own file rather than alongside the stock
// resources for that reason: Warehouse configures it, and Stock only references its id.
//
// Beyond CRUD it carries a lifecycle of its own: suspend takes a location out of use while leaving
// whatever it holds exactly where it is, and move re-parents it and rewrites the cached paths
// beneath. Archiving is the built-in action, guarded in the domain service.

func inventoryLocationEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.InventoryLocationSchemaName,
		DefineActions: defineInventoryLocationActions,
	}
}
