package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxProductClassificationSchemaName = "accounting_tax_product_classification"

	TaxProductClassificationFieldId             = basemodel.FieldId
	TaxProductClassificationFieldOrgId          = basemodel.FieldOrgId
	TaxProductClassificationFieldCode           = "code"
	TaxProductClassificationFieldName           = "name"
	TaxProductClassificationFieldJurisdictionId = "jurisdiction_id"
	TaxProductClassificationFieldExternalCode   = "external_code"
	TaxProductClassificationFieldDescription    = "description"
	TaxProductClassificationFieldIsArchived     = basemodel.FieldIsArchived
)

//go:embed tax_product_classification.json
var taxProductClassificationSchemaJson string

func TaxProductClassificationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxProductClassificationSchemaJson)
}

// TaxProductClassification is what a product is for tax purposes.
//
// Determination runs Product -> Classification -> Rule -> Tax. The indirection is the point: a
// statutory rate change becomes a new rule rather than an edit to every product row.
type TaxProductClassification struct {
	basemodel.DynamicModelBase
}

func NewTaxProductClassification() *TaxProductClassification {
	return &TaxProductClassification{basemodel.NewDynamicModel()}
}

func NewTaxProductClassificationFrom(src dmodel.DynamicFields) *TaxProductClassification {
	return &TaxProductClassification{basemodel.NewDynamicModel(src)}
}

func (this TaxProductClassification) GetCode() *string {
	return this.GetFieldData().GetString(TaxProductClassificationFieldCode)
}

func (this *TaxProductClassification) SetCode(v *string) {
	this.GetFieldData().SetString(TaxProductClassificationFieldCode, v)
}

func (this TaxProductClassification) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(TaxProductClassificationFieldName)
}

func (this *TaxProductClassification) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(TaxProductClassificationFieldName, v)
}

func (this TaxProductClassification) GetJurisdictionId() *model.Id {
	return this.GetFieldData().GetModelId(TaxProductClassificationFieldJurisdictionId)
}

func (this *TaxProductClassification) SetJurisdictionId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxProductClassificationFieldJurisdictionId, v)
}

func (this TaxProductClassification) GetExternalCode() *string {
	return this.GetFieldData().GetString(TaxProductClassificationFieldExternalCode)
}

func (this *TaxProductClassification) SetExternalCode(v *string) {
	this.GetFieldData().SetString(TaxProductClassificationFieldExternalCode, v)
}

func (this TaxProductClassification) GetDescription() *string {
	return this.GetFieldData().GetString(TaxProductClassificationFieldDescription)
}

func (this *TaxProductClassification) SetDescription(v *string) {
	this.GetFieldData().SetString(TaxProductClassificationFieldDescription, v)
}
