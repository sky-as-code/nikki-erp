package domain

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	InventoryMediaSchemaName = "inventory_media"

	InventoryMediaFieldId            = basemodel.FieldId
	InventoryMediaFieldStorageKey    = "storage_key"
	InventoryMediaFieldMediaType     = "media_type"
	InventoryMediaFieldResource      = "resource"
	InventoryMediaFieldPendingDelete = "pending_delete"
)

func InventoryMediaSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(InventoryMediaSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Inventory Media"}).
		TableName("inventory_media").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(dmodel.DefineField().
			Name(InventoryMediaFieldStorageKey).
			Label(model.LangJson{model.LanguageCodeEnUs: "Storage key"}).
			DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_DESC_LENGTH)).
			RequiredForCreate()).
		Field(dmodel.DefineField().
			Name(InventoryMediaFieldMediaType).
			Label(model.LangJson{model.LanguageCodeEnUs: "Media type"}).
			DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_SHORT_NAME_LENGTH)).
			RequiredForCreate()).
		Field(dmodel.DefineField().
			Name(InventoryMediaFieldResource).
			Label(model.LangJson{model.LanguageCodeEnUs: "Resource"}).
			DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_SHORT_NAME_LENGTH)).
			RequiredForCreate()).
		Field(dmodel.DefineField().
			Name(InventoryMediaFieldPendingDelete).
			Label(model.LangJson{model.LanguageCodeEnUs: "Pending Delete"}).
			DataType(dmodel.FieldDataTypeBoolean()).
			Default(false).
			RequiredForCreate()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.TraceableModelSchemaBuilder()).
		CompositeUnique(InventoryMediaFieldStorageKey)
}

type InventoryMedia struct {
	basemodel.DynamicModelBase
}

func NewInventoryMedia() *InventoryMedia {
	return &InventoryMedia{basemodel.NewDynamicModel()}
}

func NewInventoryMediaFrom(src dmodel.DynamicFields) *InventoryMedia {
	return &InventoryMedia{basemodel.NewDynamicModel(src)}
}

func (this InventoryMedia) GetId() *model.Id {
	return this.GetFieldData().GetModelId(InventoryMediaFieldId)
}

func (this *InventoryMedia) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(InventoryMediaFieldId, v)
}

func (this InventoryMedia) IsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

func (this InventoryMedia) MustIsArchived() bool {
	b := this.GetFieldData().GetBool(basemodel.FieldIsArchived)
	if b == nil {
		panic(errors.New("kiosk media is_archived is nil"))
	}
	return *b
}

func (this *InventoryMedia) SetIsArchived(v *bool) {
	this.GetFieldData().SetBool(basemodel.FieldIsArchived, v)
}

func (this InventoryMedia) GetStorageKey() *string {
	return this.GetFieldData().GetString(InventoryMediaFieldStorageKey)
}

func (this *InventoryMedia) SetStorageKey(v *string) {
	this.GetFieldData().SetString(InventoryMediaFieldStorageKey, v)
}

func (this InventoryMedia) GetMediaType() *string {
	return this.GetFieldData().GetString(InventoryMediaFieldMediaType)
}

func (this *InventoryMedia) SetMediaType(v *string) {
	this.GetFieldData().SetString(InventoryMediaFieldMediaType, v)
}

func (this InventoryMedia) GetResource() *string {
	return this.GetFieldData().GetString(InventoryMediaFieldResource)
}

func (this *InventoryMedia) SetResource(v *string) {
	this.GetFieldData().SetString(InventoryMediaFieldResource, v)
}

func (this InventoryMedia) GetPendingDelete() *bool {
	return this.GetFieldData().GetBool(InventoryMediaFieldPendingDelete)
}

func (this *InventoryMedia) SetPendingDelete(v *bool) {
	this.GetFieldData().SetBool(InventoryMediaFieldPendingDelete, v)
}
