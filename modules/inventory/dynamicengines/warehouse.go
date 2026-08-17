package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The Warehouse Management resources: the warehouse itself, the resupply topology between
// warehouses, the storage policy a location may carry, and the rules deciding where arriving goods
// should be put.
//
// All four are configuration. None of them changes a quantity, and the operations that look like
// they might — reconfiguring a flow, declaring a supply relation, suggesting a putaway
// destination — deliberately do not: the Stock movement engine is the only thing that moves goods.

func warehouseEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.WarehouseSchemaName,
		DefaultFields: []string{
			models.WarehouseFieldCode,
			models.WarehouseFieldName,
			models.WarehouseFieldWarehouseRole,
			models.WarehouseFieldParentWarehouseId,
			models.WarehouseFieldIncomingFlow,
			models.WarehouseFieldOutgoingFlow,
			models.WarehouseFieldStatus,
		},
		DefineActions: defineWarehouseActions,
	}
}

func storageCategoryEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.StorageCategorySchemaName,
		DefaultFields: []string{
			models.StorageCategoryFieldCode,
			models.StorageCategoryFieldName,
			models.StorageCategoryFieldMaxWeight,
			models.StorageCategoryFieldAllowNewItemPolicy,
		},
	}
}

func warehouseSupplyRelationEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.WarehouseSupplyRelationSchemaName,
		DefaultFields: []string{
			models.WarehouseSupplyRelationFieldSourceWarehouseId,
			models.WarehouseSupplyRelationFieldDestinationWarehouseId,
			models.WarehouseSupplyRelationFieldPriority,
			models.WarehouseSupplyRelationFieldIsDefault,
		},
	}
}

func putawayRuleEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.PutawayRuleSchemaName,
		DefaultFields: []string{
			models.PutawayRuleFieldCode,
			models.PutawayRuleFieldWarehouseId,
			models.PutawayRuleFieldSourceLocationId,
			models.PutawayRuleFieldDestinationLocationId,
			models.PutawayRuleFieldPriority,
			models.PutawayRuleFieldSublocationStrategy,
		},
		DefineActions: definePutawayRuleActions,
	}
}
