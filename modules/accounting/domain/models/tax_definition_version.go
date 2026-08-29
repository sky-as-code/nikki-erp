package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxDefinitionVersionSchemaName = "accounting_tax_definition_version"

	TaxDefinitionVersionFieldId                     = basemodel.FieldId
	TaxDefinitionVersionFieldOrgId                  = basemodel.FieldOrgId
	TaxDefinitionVersionFieldTaxId                  = "tax_id"
	TaxDefinitionVersionFieldVersionNo              = "version_no"
	TaxDefinitionVersionFieldUsage                  = "usage"
	TaxDefinitionVersionFieldJurisdictionId         = "jurisdiction_id"
	TaxDefinitionVersionFieldTaxGroupId             = "tax_group_id"
	TaxDefinitionVersionFieldCalculationType        = "calculation_type"
	TaxDefinitionVersionFieldTaxTreatment           = "tax_treatment"
	TaxDefinitionVersionFieldPriceInclusionMode     = "price_inclusion_mode"
	TaxDefinitionVersionFieldSequence               = "sequence"
	TaxDefinitionVersionFieldAffectSubsequentBase   = "affect_subsequent_base"
	TaxDefinitionVersionFieldBaseAffectedByPrevious = "base_affected_by_previous"
	TaxDefinitionVersionFieldEffectiveFrom          = "effective_from"
	TaxDefinitionVersionFieldEffectiveTo            = "effective_to"
	TaxDefinitionVersionFieldLegalReference         = "legal_reference"
	TaxDefinitionVersionFieldSupersedesVersionId    = "supersedes_version_id"
	TaxDefinitionVersionFieldLifecycleStatus        = "lifecycle_status"
	TaxDefinitionVersionFieldIsArchived             = basemodel.FieldIsArchived
)

//go:embed tax_definition_version.json
var taxDefinitionVersionSchemaJson string

func TaxDefinitionVersionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxDefinitionVersionSchemaJson)
}

// TaxDefinitionVersion carries every attribute of a tax that affects determination or calculation,
// for one effective period. Once published its material fields are immutable, so a transaction
// calculated against them stays explainable for as long as it exists.
type TaxDefinitionVersion struct {
	basemodel.DynamicModelBase
}

func NewTaxDefinitionVersion() *TaxDefinitionVersion {
	return &TaxDefinitionVersion{basemodel.NewDynamicModel()}
}

func NewTaxDefinitionVersionFrom(src dmodel.DynamicFields) *TaxDefinitionVersion {
	return &TaxDefinitionVersion{basemodel.NewDynamicModel(src)}
}

func (this TaxDefinitionVersion) GetTaxId() *model.Id {
	return this.GetFieldData().GetModelId(TaxDefinitionVersionFieldTaxId)
}

func (this *TaxDefinitionVersion) SetTaxId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxDefinitionVersionFieldTaxId, v)
}

func (this TaxDefinitionVersion) GetVersionNo() *int32 {
	return this.GetFieldData().GetInt32(TaxDefinitionVersionFieldVersionNo)
}

func (this *TaxDefinitionVersion) SetVersionNo(v *int32) {
	this.GetFieldData().SetInt32(TaxDefinitionVersionFieldVersionNo, v)
}

func (this TaxDefinitionVersion) GetUsage() *string {
	return this.GetFieldData().GetString(TaxDefinitionVersionFieldUsage)
}

func (this *TaxDefinitionVersion) SetUsage(v *string) {
	this.GetFieldData().SetString(TaxDefinitionVersionFieldUsage, v)
}

func (this TaxDefinitionVersion) GetJurisdictionId() *model.Id {
	return this.GetFieldData().GetModelId(TaxDefinitionVersionFieldJurisdictionId)
}

func (this *TaxDefinitionVersion) SetJurisdictionId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxDefinitionVersionFieldJurisdictionId, v)
}

func (this TaxDefinitionVersion) GetTaxGroupId() *model.Id {
	return this.GetFieldData().GetModelId(TaxDefinitionVersionFieldTaxGroupId)
}

func (this *TaxDefinitionVersion) SetTaxGroupId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxDefinitionVersionFieldTaxGroupId, v)
}

func (this TaxDefinitionVersion) GetCalculationType() *string {
	return this.GetFieldData().GetString(TaxDefinitionVersionFieldCalculationType)
}

func (this *TaxDefinitionVersion) SetCalculationType(v *string) {
	this.GetFieldData().SetString(TaxDefinitionVersionFieldCalculationType, v)
}

func (this TaxDefinitionVersion) GetTaxTreatment() *string {
	return this.GetFieldData().GetString(TaxDefinitionVersionFieldTaxTreatment)
}

func (this *TaxDefinitionVersion) SetTaxTreatment(v *string) {
	this.GetFieldData().SetString(TaxDefinitionVersionFieldTaxTreatment, v)
}

func (this TaxDefinitionVersion) GetPriceInclusionMode() *string {
	return this.GetFieldData().GetString(TaxDefinitionVersionFieldPriceInclusionMode)
}

func (this *TaxDefinitionVersion) SetPriceInclusionMode(v *string) {
	this.GetFieldData().SetString(TaxDefinitionVersionFieldPriceInclusionMode, v)
}

func (this TaxDefinitionVersion) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(TaxDefinitionVersionFieldSequence)
}

func (this *TaxDefinitionVersion) SetSequence(v *int32) {
	this.GetFieldData().SetInt32(TaxDefinitionVersionFieldSequence, v)
}

func (this TaxDefinitionVersion) GetAffectSubsequentBase() *bool {
	return this.GetFieldData().GetBool(TaxDefinitionVersionFieldAffectSubsequentBase)
}

func (this *TaxDefinitionVersion) SetAffectSubsequentBase(v *bool) {
	this.GetFieldData().SetBool(TaxDefinitionVersionFieldAffectSubsequentBase, v)
}

func (this TaxDefinitionVersion) GetBaseAffectedByPrevious() *bool {
	return this.GetFieldData().GetBool(TaxDefinitionVersionFieldBaseAffectedByPrevious)
}

func (this *TaxDefinitionVersion) SetBaseAffectedByPrevious(v *bool) {
	this.GetFieldData().SetBool(TaxDefinitionVersionFieldBaseAffectedByPrevious, v)
}

func (this TaxDefinitionVersion) GetEffectiveFrom() *model.ModelDate {
	return this.GetFieldData().GetModelDate(TaxDefinitionVersionFieldEffectiveFrom)
}

func (this *TaxDefinitionVersion) SetEffectiveFrom(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(TaxDefinitionVersionFieldEffectiveFrom, v)
}

func (this TaxDefinitionVersion) GetEffectiveTo() *model.ModelDate {
	return this.GetFieldData().GetModelDate(TaxDefinitionVersionFieldEffectiveTo)
}

func (this *TaxDefinitionVersion) SetEffectiveTo(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(TaxDefinitionVersionFieldEffectiveTo, v)
}

func (this TaxDefinitionVersion) GetLegalReference() *string {
	return this.GetFieldData().GetString(TaxDefinitionVersionFieldLegalReference)
}

func (this *TaxDefinitionVersion) SetLegalReference(v *string) {
	this.GetFieldData().SetString(TaxDefinitionVersionFieldLegalReference, v)
}

func (this TaxDefinitionVersion) GetSupersedesVersionId() *model.Id {
	return this.GetFieldData().GetModelId(TaxDefinitionVersionFieldSupersedesVersionId)
}

func (this *TaxDefinitionVersion) SetSupersedesVersionId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxDefinitionVersionFieldSupersedesVersionId, v)
}

func (this TaxDefinitionVersion) GetLifecycleStatus() *string {
	return this.GetFieldData().GetString(TaxDefinitionVersionFieldLifecycleStatus)
}

func (this *TaxDefinitionVersion) SetLifecycleStatus(v *string) {
	this.GetFieldData().SetString(TaxDefinitionVersionFieldLifecycleStatus, v)
}
