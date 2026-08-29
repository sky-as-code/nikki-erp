package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxSchemaName = "accounting_tax"

	TaxFieldId           = basemodel.FieldId
	TaxFieldOrgId        = basemodel.FieldOrgId
	TaxFieldCode         = "code"
	TaxFieldName         = "name"
	TaxFieldTaxKind      = "tax_kind"
	TaxFieldInvoiceLabel = "invoice_label"
	TaxFieldDescription  = "description"
	TaxFieldIsArchived   = basemodel.FieldIsArchived
)

//go:embed tax.json
var taxSchemaJson string

func TaxSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxSchemaJson)
}

// Tax is the stable business identity of a tax and holds nothing that affects calculation.
// Everything deciding an amount (rate, treatment, calculation type, effective period) lives on
// TaxDefinitionVersion, so a change is a new version rather than an edit that would reinterpret
// past transactions.
type Tax struct {
	basemodel.DynamicModelBase
}

func NewTax() *Tax {
	return &Tax{basemodel.NewDynamicModel()}
}

func NewTaxFrom(src dmodel.DynamicFields) *Tax {
	return &Tax{basemodel.NewDynamicModel(src)}
}

func (this Tax) GetCode() *string {
	return this.GetFieldData().GetString(TaxFieldCode)
}

func (this *Tax) SetCode(v *string) {
	this.GetFieldData().SetString(TaxFieldCode, v)
}

func (this Tax) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(TaxFieldName)
}

func (this *Tax) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(TaxFieldName, v)
}

func (this Tax) GetTaxKind() *string {
	return this.GetFieldData().GetString(TaxFieldTaxKind)
}

func (this *Tax) SetTaxKind(v *string) {
	this.GetFieldData().SetString(TaxFieldTaxKind, v)
}

func (this Tax) GetInvoiceLabel() *model.LangJson {
	return this.GetFieldData().GetLangJson(TaxFieldInvoiceLabel)
}

func (this *Tax) SetInvoiceLabel(v *model.LangJson) {
	this.GetFieldData().SetLangJson(TaxFieldInvoiceLabel, v)
}

func (this Tax) GetDescription() *string {
	return this.GetFieldData().GetString(TaxFieldDescription)
}

func (this *Tax) SetDescription(v *string) {
	this.GetFieldData().SetString(TaxFieldDescription, v)
}
