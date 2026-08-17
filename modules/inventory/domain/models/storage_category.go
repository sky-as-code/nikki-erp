package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// StorageCategoryAllowNewItemPolicy says whether goods arriving may be mixed into a location that
// already holds something.
const (
	StorageCategoryAllowNewItemPolicyAllow           = "allow"
	StorageCategoryAllowNewItemPolicySameProductOnly = "same_product_only"
	StorageCategoryAllowNewItemPolicyEmptyOnly       = "empty_only"
)

// Storage Category has no status field, only is_archived. Nothing about it has an operational
// state independent of archiving: a category is either part of the master data available for new
// assignments or it is not. Adding a status to mirror Warehouse would create two ways to express
// one fact.
const (
	StorageCategorySchemaName = "inventory_storage_category"

	StorageCategoryFieldId                 = basemodel.FieldId
	StorageCategoryFieldCode               = "code"
	StorageCategoryFieldName               = "name"
	StorageCategoryFieldMaxWeight          = "max_weight"
	StorageCategoryFieldAllowNewItemPolicy = "allow_new_item_policy"
	StorageCategoryFieldDescription        = "description"
	StorageCategoryFieldOrgId              = "org_id"
)

//go:embed storage_category.json
var storageCategorySchemaJson string

func StorageCategorySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(storageCategorySchemaJson)
}

type StorageCategory struct {
	basemodel.DynamicModelBase
}

func NewStorageCategory() *StorageCategory {
	return &StorageCategory{basemodel.NewDynamicModel()}
}

func NewStorageCategoryFrom(src dmodel.DynamicFields) *StorageCategory {
	return &StorageCategory{basemodel.NewDynamicModel(src)}
}

func (this StorageCategory) GetCode() *string {
	return this.GetFieldData().GetString(StorageCategoryFieldCode)
}

func (this *StorageCategory) SetCode(v *string) {
	this.GetFieldData().SetString(StorageCategoryFieldCode, v)
}

func (this StorageCategory) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(StorageCategoryFieldName)
}

func (this *StorageCategory) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(StorageCategoryFieldName, v)
}

func (this StorageCategory) GetAllowNewItemPolicy() *string {
	return this.GetFieldData().GetString(StorageCategoryFieldAllowNewItemPolicy)
}

func (this *StorageCategory) SetAllowNewItemPolicy(v *string) {
	this.GetFieldData().SetString(StorageCategoryFieldAllowNewItemPolicy, v)
}

func (this StorageCategory) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(StorageCategoryFieldOrgId)
}

func (this *StorageCategory) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(StorageCategoryFieldOrgId, v)
}
