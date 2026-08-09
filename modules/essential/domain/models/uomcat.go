package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	UomCatSchemaName = "essential_uomcat"

	UomCatFieldId             = basemodel.FieldId
	UomCatFieldName           = "name"
	UomCatFieldReferenceUomId = "reference_uom_id"
	UomCatFieldOrgId          = "org_id"

	UomCatEdgeUoms = "uoms"
)

//go:embed uomcat.json
var uomCatSchemaJson string

func UomCatSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(uomCatSchemaJson)
}

type UomCat struct {
	basemodel.DynamicModelBase
}

func NewUomCat() *UomCat {
	return &UomCat{basemodel.NewDynamicModel()}
}

func NewUomCatFrom(src dmodel.DynamicFields) *UomCat {
	return &UomCat{basemodel.NewDynamicModel(src)}
}

func (this UomCat) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(UomCatFieldName)
}

func (this *UomCat) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(UomCatFieldName, v)
}

// GetReferenceUomId returns the single Reference UoM of this category, or nil while the
// category has not been given one yet.
func (this UomCat) GetReferenceUomId() *model.Id {
	return this.GetFieldData().GetModelId(UomCatFieldReferenceUomId)
}

func (this *UomCat) SetReferenceUomId(v *model.Id) {
	this.GetFieldData().SetModelId(UomCatFieldReferenceUomId, v)
}

func (this UomCat) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(UomCatFieldOrgId)
}

func (this *UomCat) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(UomCatFieldOrgId, v)
}
