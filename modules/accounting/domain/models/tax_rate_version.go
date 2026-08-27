package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxRateVersionSchemaName = "accounting_tax_rate_version"

	TaxRateVersionFieldId              = basemodel.FieldId
	TaxRateVersionFieldOrgId           = basemodel.FieldOrgId
	TaxRateVersionFieldTaxId           = "tax_id"
	TaxRateVersionFieldVersionNo       = "version_no"
	TaxRateVersionFieldRate            = "rate"
	TaxRateVersionFieldFixedAmount     = "fixed_amount"
	TaxRateVersionFieldCurrencyCode    = "currency_code"
	TaxRateVersionFieldRateUomId       = "rate_uom_id"
	TaxRateVersionFieldEffectiveFrom   = "effective_from"
	TaxRateVersionFieldEffectiveTo     = "effective_to"
	TaxRateVersionFieldLegalReference  = "legal_reference"
	TaxRateVersionFieldDescription     = "description"
	TaxRateVersionFieldLifecycleStatus = "lifecycle_status"
	TaxRateVersionFieldIsArchived      = basemodel.FieldIsArchived
)

//go:embed tax_rate_version.json
var taxRateVersionSchemaJson string

func TaxRateVersionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxRateVersionSchemaJson)
}

// TaxRateVersion is the rate or fixed amount of a tax over one effective period.
//
// Two published versions of the same tax may never overlap, so that a tax_date resolves to
// exactly one rate. Zero matches and two matches are both errors that surface as unresolved;
// neither may be papered over by taking the newest row.
type TaxRateVersion struct {
	basemodel.DynamicModelBase
}

func NewTaxRateVersion() *TaxRateVersion {
	return &TaxRateVersion{basemodel.NewDynamicModel()}
}

func NewTaxRateVersionFrom(src dmodel.DynamicFields) *TaxRateVersion {
	return &TaxRateVersion{basemodel.NewDynamicModel(src)}
}

func (this TaxRateVersion) GetTaxId() *model.Id {
	return this.GetFieldData().GetModelId(TaxRateVersionFieldTaxId)
}

func (this *TaxRateVersion) SetTaxId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxRateVersionFieldTaxId, v)
}

func (this TaxRateVersion) GetVersionNo() *int32 {
	return this.GetFieldData().GetInt32(TaxRateVersionFieldVersionNo)
}

func (this *TaxRateVersion) SetVersionNo(v *int32) {
	this.GetFieldData().SetInt32(TaxRateVersionFieldVersionNo, v)
}

func (this TaxRateVersion) GetRate() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(TaxRateVersionFieldRate)
}

func (this *TaxRateVersion) SetRate(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(TaxRateVersionFieldRate, v)
}

func (this TaxRateVersion) GetFixedAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(TaxRateVersionFieldFixedAmount)
}

func (this *TaxRateVersion) SetFixedAmount(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(TaxRateVersionFieldFixedAmount, v)
}

func (this TaxRateVersion) GetCurrencyCode() *string {
	return this.GetFieldData().GetString(TaxRateVersionFieldCurrencyCode)
}

func (this *TaxRateVersion) SetCurrencyCode(v *string) {
	this.GetFieldData().SetString(TaxRateVersionFieldCurrencyCode, v)
}

func (this TaxRateVersion) GetRateUomId() *model.Id {
	return this.GetFieldData().GetModelId(TaxRateVersionFieldRateUomId)
}

func (this *TaxRateVersion) SetRateUomId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxRateVersionFieldRateUomId, v)
}

func (this TaxRateVersion) GetEffectiveFrom() *model.ModelDate {
	return this.GetFieldData().GetModelDate(TaxRateVersionFieldEffectiveFrom)
}

func (this *TaxRateVersion) SetEffectiveFrom(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(TaxRateVersionFieldEffectiveFrom, v)
}

func (this TaxRateVersion) GetEffectiveTo() *model.ModelDate {
	return this.GetFieldData().GetModelDate(TaxRateVersionFieldEffectiveTo)
}

func (this *TaxRateVersion) SetEffectiveTo(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(TaxRateVersionFieldEffectiveTo, v)
}

func (this TaxRateVersion) GetLegalReference() *string {
	return this.GetFieldData().GetString(TaxRateVersionFieldLegalReference)
}

func (this *TaxRateVersion) SetLegalReference(v *string) {
	this.GetFieldData().SetString(TaxRateVersionFieldLegalReference, v)
}

func (this TaxRateVersion) GetDescription() *string {
	return this.GetFieldData().GetString(TaxRateVersionFieldDescription)
}

func (this *TaxRateVersion) SetDescription(v *string) {
	this.GetFieldData().SetString(TaxRateVersionFieldDescription, v)
}

func (this TaxRateVersion) GetLifecycleStatus() *string {
	return this.GetFieldData().GetString(TaxRateVersionFieldLifecycleStatus)
}

func (this *TaxRateVersion) SetLifecycleStatus(v *string) {
	this.GetFieldData().SetString(TaxRateVersionFieldLifecycleStatus, v)
}
