package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The Warehouse Management resources: the warehouse, the resupply topology between warehouses, the
// storage policy a location may carry, and the putaway rules.
//
// All four are configuration. None changes a quantity — not even the operations that look like
// they might, such as reconfiguring a flow or suggesting a putaway destination. The Stock movement
// engine is the only thing that moves goods.

func warehouseEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.WarehouseSchemaName,
		DefineActions: defineWarehouseActions,
	}
}

func storageCategoryEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.StorageCategorySchemaName,
	}
}

func warehouseSupplyRelationEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.WarehouseSupplyRelationSchemaName,
	}
}

func putawayRuleEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.PutawayRuleSchemaName,
		DefineActions: definePutawayRuleActions,
	}
}
