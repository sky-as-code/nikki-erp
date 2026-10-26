package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	BrandSchemaName = "inventory_brand"

	BrandFieldId          = basemodel.FieldId
	BrandFieldCode        = "code"
	BrandFieldName        = "name"
	BrandFieldLogoId      = "logo_id"
	BrandFieldCountryId   = "country_id"
	BrandFieldWebsite     = "website"
	BrandFieldDescription = "description"
	BrandFieldOrgId       = "org_id"
)

//go:embed brand.json
var brandSchemaJson string

func BrandSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(brandSchemaJson)
}

type Brand struct {
	basemodel.DynamicModelBase
}

func NewBrand() *Brand {
	return &Brand{basemodel.NewDynamicModel()}
}

func NewBrandFrom(src dmodel.DynamicFields) *Brand {
	return &Brand{basemodel.NewDynamicModel(src)}
}

func (this Brand) GetCode() *string {
	return this.GetFieldData().GetString(BrandFieldCode)
}

func (this *Brand) SetCode(v *string) {
	this.GetFieldData().SetString(BrandFieldCode, v)
}

func (this Brand) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(BrandFieldName)
}

func (this *Brand) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(BrandFieldName, v)
}

func (this Brand) GetLogoId() *model.Id {
	return this.GetFieldData().GetModelId(BrandFieldLogoId)
}

func (this *Brand) SetLogoId(v *model.Id) {
	this.GetFieldData().SetModelId(BrandFieldLogoId, v)
}

func (this Brand) GetCountryId() *model.Id {
	return this.GetFieldData().GetModelId(BrandFieldCountryId)
}

func (this *Brand) SetCountryId(v *model.Id) {
	this.GetFieldData().SetModelId(BrandFieldCountryId, v)
}

func (this Brand) GetWebsite() *string {
	return this.GetFieldData().GetString(BrandFieldWebsite)
}

func (this *Brand) SetWebsite(v *string) {
	this.GetFieldData().SetString(BrandFieldWebsite, v)
}

func (this Brand) GetDescription() *model.LangJson {
	return this.GetFieldData().GetLangJson(BrandFieldDescription)
}

func (this *Brand) SetDescription(v *model.LangJson) {
	this.GetFieldData().SetLangJson(BrandFieldDescription, v)
}

func (this Brand) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(BrandFieldOrgId)
}

func (this *Brand) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(BrandFieldOrgId, v)
}
