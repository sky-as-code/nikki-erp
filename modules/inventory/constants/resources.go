package constants

const InventoryModuleName = "inventory"

// Resource codes for authorization. They must be byte-identical to the dynamic-model schema names,
// because the dynamic resource engine asserts permissions using the schema name as the resource
// code. A code that drifts from its schema name denies every request with no obvious cause.
const (
	ResourceProductTemplate       = "inventory_product_template"
	ResourceProductVariant        = "inventory_product_variant"
	ResourceProductType           = "inventory_product_type"
	ResourceProductCategory       = "inventory_product_category"
	ResourceProductAttribute      = "inventory_product_attribute"
	ResourceProductAttributeValue = "inventory_product_attribute_value"
	ResourceBrand                 = "inventory_brand"

	// The canonical location resource of the whole Inventory module, owned by neither Warehouse
	// nor Stock: Warehouse configures it and Stock only references its id.
	ResourceInventoryLocation = "inventory_location"

	ResourceWarehouse               = "inventory_warehouse"
	ResourceStorageCategory         = "inventory_storage_category"
	ResourceWarehouseSupplyRelation = "inventory_warehouse_supply_relation"
	ResourcePutawayRule             = "inventory_putaway_rule"

	ResourceStockOperationType = "inventory_stock_operation_type"
	ResourceStockQuant         = "inventory_stock_quant"

	ResourceStockTransfer       = "inventory_stock_transfer"
	ResourceStockMove           = "inventory_stock_move"
	ResourceStockMoveLine       = "inventory_stock_move_line"
	ResourceStockMoveDependency = "inventory_stock_move_dependency"

	ResourceStockScrap = "inventory_stock_scrap"
)
