package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	WarehouseProductGuardSchemaName = "inventory_warehouse_product_guard"

	WarehouseProductGuardFieldId               = basemodel.FieldId
	WarehouseProductGuardFieldWarehouseId      = "warehouse_id"
	WarehouseProductGuardFieldProductVariantId = "product_variant_id"
	WarehouseProductGuardFieldOrgId            = "org_id"
)

//go:embed inventory_warehouse_product_guard.json
var warehouseProductGuardSchemaJson string

func WarehouseProductGuardSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(warehouseProductGuardSchemaJson)
}

// WarehouseProductGuard is a lock target and nothing more: one row per (org, warehouse, variant),
// taken FOR UPDATE by every operation that reads or changes that scope's availability. Locking
// quants alone is not enough, because two requests can pick two different quants and still both
// spend the same warehouse-level commitment. It is never exposed over REST.
type WarehouseProductGuard struct {
	basemodel.DynamicModelBase
}

func NewWarehouseProductGuardFrom(src dmodel.DynamicFields) *WarehouseProductGuard {
	return &WarehouseProductGuard{basemodel.NewDynamicModel(src)}
}

func (this WarehouseProductGuard) GetId() *model.Id {
	return this.GetFieldData().GetModelId(WarehouseProductGuardFieldId)
}

func (this WarehouseProductGuard) GetWarehouseId() *model.Id {
	return this.GetFieldData().GetModelId(WarehouseProductGuardFieldWarehouseId)
}

func (this WarehouseProductGuard) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(WarehouseProductGuardFieldProductVariantId)
}

func (this WarehouseProductGuard) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(WarehouseProductGuardFieldOrgId)
}
