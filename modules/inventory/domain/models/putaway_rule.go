package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// PutawaySublocationStrategy says how a sublocation under the rule's destination is chosen.
const (
	PutawaySublocationStrategyFixed    = "fixed"
	PutawaySublocationStrategyLastUsed = "last_used"
	PutawaySublocationStrategyCategory = "category"
)

// Putaway Rule decides where arriving goods should be put. It only ever answers the question — the
// suggestion carries a destination and the rule that matched, and changes no quantity. Moving the
// goods is the Stock movement engine's job.
//
// It has no status field: a rule is evaluated when it is not archived, so a separate active flag
// would be a second way to say the same thing.
const (
	PutawayRuleSchemaName = "inventory_putaway_rule"

	PutawayRuleFieldId                    = basemodel.FieldId
	PutawayRuleFieldCode                  = "code"
	PutawayRuleFieldWarehouseId           = "warehouse_id"
	PutawayRuleFieldSourceLocationId      = "source_location_id"
	PutawayRuleFieldDestinationLocationId = "destination_location_id"
	PutawayRuleFieldStorageCategoryId     = "storage_category_id"
	PutawayRuleFieldProductId             = "product_id"
	PutawayRuleFieldProductCategoryId     = "product_category_id"
	PutawayRuleFieldPackageTypeId         = "package_type_id"
	PutawayRuleFieldPriority              = "priority"
	PutawayRuleFieldSublocationStrategy   = "sublocation_strategy"
	PutawayRuleFieldOrgId                 = "org_id"

	PutawayRuleEdgeWarehouse           = "warehouse"
	PutawayRuleEdgeSourceLocation      = "source_location"
	PutawayRuleEdgeDestinationLocation = "destination_location"
	PutawayRuleEdgeStorageCategory     = "storage_category"
)

//go:embed putaway_rule.json
var putawayRuleSchemaJson string

func PutawayRuleSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(putawayRuleSchemaJson)
}

type PutawayRule struct {
	basemodel.DynamicModelBase
}

func NewPutawayRule() *PutawayRule {
	return &PutawayRule{basemodel.NewDynamicModel()}
}

func NewPutawayRuleFrom(src dmodel.DynamicFields) *PutawayRule {
	return &PutawayRule{basemodel.NewDynamicModel(src)}
}

func (this PutawayRule) GetCode() *string {
	return this.GetFieldData().GetString(PutawayRuleFieldCode)
}

func (this *PutawayRule) SetCode(v *string) {
	this.GetFieldData().SetString(PutawayRuleFieldCode, v)
}

func (this PutawayRule) GetWarehouseId() *model.Id {
	return this.GetFieldData().GetModelId(PutawayRuleFieldWarehouseId)
}

func (this *PutawayRule) SetWarehouseId(v *model.Id) {
	this.GetFieldData().SetModelId(PutawayRuleFieldWarehouseId, v)
}

func (this PutawayRule) GetSourceLocationId() *model.Id {
	return this.GetFieldData().GetModelId(PutawayRuleFieldSourceLocationId)
}

func (this *PutawayRule) SetSourceLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(PutawayRuleFieldSourceLocationId, v)
}

func (this PutawayRule) GetDestinationLocationId() *model.Id {
	return this.GetFieldData().GetModelId(PutawayRuleFieldDestinationLocationId)
}

func (this *PutawayRule) SetDestinationLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(PutawayRuleFieldDestinationLocationId, v)
}

func (this PutawayRule) GetStorageCategoryId() *model.Id {
	return this.GetFieldData().GetModelId(PutawayRuleFieldStorageCategoryId)
}

func (this *PutawayRule) SetStorageCategoryId(v *model.Id) {
	this.GetFieldData().SetModelId(PutawayRuleFieldStorageCategoryId, v)
}

// GetPriority returns the order this rule is considered in. Lower is considered first.
func (this PutawayRule) GetPriority() *int32 {
	return this.GetFieldData().GetInt32(PutawayRuleFieldPriority)
}

func (this *PutawayRule) SetPriority(v *int32) {
	this.GetFieldData()[PutawayRuleFieldPriority] = v
}

func (this PutawayRule) GetProductId() *model.Id {
	return this.GetFieldData().GetModelId(PutawayRuleFieldProductId)
}

func (this PutawayRule) GetProductCategoryId() *model.Id {
	return this.GetFieldData().GetModelId(PutawayRuleFieldProductCategoryId)
}

func (this PutawayRule) GetPackageTypeId() *model.Id {
	return this.GetFieldData().GetModelId(PutawayRuleFieldPackageTypeId)
}

func (this PutawayRule) GetIsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

func (this PutawayRule) GetSublocationStrategy() *string {
	return this.GetFieldData().GetString(PutawayRuleFieldSublocationStrategy)
}

func (this *PutawayRule) SetSublocationStrategy(v *string) {
	this.GetFieldData().SetString(PutawayRuleFieldSublocationStrategy, v)
}

func (this PutawayRule) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(PutawayRuleFieldOrgId)
}

func (this *PutawayRule) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(PutawayRuleFieldOrgId, v)
}
