package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxJurisdictionSchemaName = "accounting_tax_jurisdiction"

	TaxJurisdictionFieldId            = basemodel.FieldId
	TaxJurisdictionFieldOrgId         = basemodel.FieldOrgId
	TaxJurisdictionFieldCode          = "code"
	TaxJurisdictionFieldName          = "name"
	TaxJurisdictionFieldCountryCode   = "country_code"
	TaxJurisdictionFieldLevel         = "level"
	TaxJurisdictionFieldParentId      = "parent_id"
	TaxJurisdictionFieldAuthorityName = "authority_name"
	TaxJurisdictionFieldIsArchived    = basemodel.FieldIsArchived
)

//go:embed tax_jurisdiction.json
var taxJurisdictionSchemaJson string

func TaxJurisdictionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxJurisdictionSchemaJson)
}

// TaxJurisdiction is a territory in which a taxing authority levies tax.
//
// Jurisdictions form a tree (country -> state -> city). The tree must stay acyclic, which the
// dynamic engine enforces on write because a self-referencing foreign key cannot.
type TaxJurisdiction struct {
	basemodel.DynamicModelBase
}

func NewTaxJurisdiction() *TaxJurisdiction {
	return &TaxJurisdiction{basemodel.NewDynamicModel()}
}

func NewTaxJurisdictionFrom(src dmodel.DynamicFields) *TaxJurisdiction {
	return &TaxJurisdiction{basemodel.NewDynamicModel(src)}
}

func (this TaxJurisdiction) GetCode() *string {
	return this.GetFieldData().GetString(TaxJurisdictionFieldCode)
}

func (this *TaxJurisdiction) SetCode(v *string) {
	this.GetFieldData().SetString(TaxJurisdictionFieldCode, v)
}

func (this TaxJurisdiction) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(TaxJurisdictionFieldName)
}

func (this *TaxJurisdiction) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(TaxJurisdictionFieldName, v)
}

func (this TaxJurisdiction) GetCountryCode() *string {
	return this.GetFieldData().GetString(TaxJurisdictionFieldCountryCode)
}

func (this *TaxJurisdiction) SetCountryCode(v *string) {
	this.GetFieldData().SetString(TaxJurisdictionFieldCountryCode, v)
}

func (this TaxJurisdiction) GetLevel() *string {
	return this.GetFieldData().GetString(TaxJurisdictionFieldLevel)
}

func (this *TaxJurisdiction) SetLevel(v *string) {
	this.GetFieldData().SetString(TaxJurisdictionFieldLevel, v)
}

func (this TaxJurisdiction) GetParentId() *model.Id {
	return this.GetFieldData().GetModelId(TaxJurisdictionFieldParentId)
}

func (this *TaxJurisdiction) SetParentId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxJurisdictionFieldParentId, v)
}

func (this TaxJurisdiction) GetAuthorityName() *string {
	return this.GetFieldData().GetString(TaxJurisdictionFieldAuthorityName)
}

func (this *TaxJurisdiction) SetAuthorityName(v *string) {
	this.GetFieldData().SetString(TaxJurisdictionFieldAuthorityName, v)
}
