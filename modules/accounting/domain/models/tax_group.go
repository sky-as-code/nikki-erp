package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxGroupSchemaName = "accounting_tax_group"

	TaxGroupFieldId              = basemodel.FieldId
	TaxGroupFieldOrgId           = basemodel.FieldOrgId
	TaxGroupFieldCode            = "code"
	TaxGroupFieldName            = "name"
	TaxGroupFieldDisplayName     = "display_name"
	TaxGroupFieldDescription     = "description"
	TaxGroupFieldDisplaySequence = "display_sequence"
	TaxGroupFieldIsArchived      = basemodel.FieldIsArchived
)

//go:embed tax_group.json
var taxGroupSchemaJson string

func TaxGroupSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxGroupSchemaJson)
}

// TaxGroup aggregates taxes for display and reporting. It never carries a calculation formula:
// grouping says how an invoice summarises taxes, not how any of them is computed.
type TaxGroup struct {
	basemodel.DynamicModelBase
}

func NewTaxGroup() *TaxGroup {
	return &TaxGroup{basemodel.NewDynamicModel()}
}

func NewTaxGroupFrom(src dmodel.DynamicFields) *TaxGroup {
	return &TaxGroup{basemodel.NewDynamicModel(src)}
}

func (this TaxGroup) GetCode() *string {
	return this.GetFieldData().GetString(TaxGroupFieldCode)
}

func (this *TaxGroup) SetCode(v *string) {
	this.GetFieldData().SetString(TaxGroupFieldCode, v)
}

func (this TaxGroup) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(TaxGroupFieldName)
}

func (this *TaxGroup) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(TaxGroupFieldName, v)
}

func (this TaxGroup) GetDisplayName() *model.LangJson {
	return this.GetFieldData().GetLangJson(TaxGroupFieldDisplayName)
}

func (this *TaxGroup) SetDisplayName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(TaxGroupFieldDisplayName, v)
}

func (this TaxGroup) GetDescription() *string {
	return this.GetFieldData().GetString(TaxGroupFieldDescription)
}

func (this *TaxGroup) SetDescription(v *string) {
	this.GetFieldData().SetString(TaxGroupFieldDescription, v)
}

func (this TaxGroup) GetDisplaySequence() *int32 {
	return this.GetFieldData().GetInt32(TaxGroupFieldDisplaySequence)
}

func (this *TaxGroup) SetDisplaySequence(v *int32) {
	this.GetFieldData().SetInt32(TaxGroupFieldDisplaySequence, v)
}
