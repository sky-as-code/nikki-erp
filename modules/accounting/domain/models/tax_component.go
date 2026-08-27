package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxComponentSchemaName = "accounting_tax_component"

	TaxComponentFieldId                           = basemodel.FieldId
	TaxComponentFieldOrgId                        = basemodel.FieldOrgId
	TaxComponentFieldParentTaxDefinitionVersionId = "parent_tax_definition_version_id"
	TaxComponentFieldComponentTaxId               = "component_tax_id"
	TaxComponentFieldSequence                     = "sequence"
	TaxComponentFieldAffectSubsequentBaseOverride = "affect_subsequent_base_override"
	TaxComponentFieldIsArchived                   = basemodel.FieldIsArchived
)

//go:embed tax_component.json
var taxComponentSchemaJson string

func TaxComponentSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxComponentSchemaJson)
}

// TaxComponent is one child tax inside a group tax.
//
// It hangs off a definition version rather than the tax itself, so that publishing freezes the
// composition along with everything else that decides an amount.
type TaxComponent struct {
	basemodel.DynamicModelBase
}

func NewTaxComponent() *TaxComponent {
	return &TaxComponent{basemodel.NewDynamicModel()}
}

func NewTaxComponentFrom(src dmodel.DynamicFields) *TaxComponent {
	return &TaxComponent{basemodel.NewDynamicModel(src)}
}

func (this TaxComponent) GetParentTaxDefinitionVersionId() *model.Id {
	return this.GetFieldData().GetModelId(TaxComponentFieldParentTaxDefinitionVersionId)
}

func (this *TaxComponent) SetParentTaxDefinitionVersionId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxComponentFieldParentTaxDefinitionVersionId, v)
}

func (this TaxComponent) GetComponentTaxId() *model.Id {
	return this.GetFieldData().GetModelId(TaxComponentFieldComponentTaxId)
}

func (this *TaxComponent) SetComponentTaxId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxComponentFieldComponentTaxId, v)
}

func (this TaxComponent) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(TaxComponentFieldSequence)
}

func (this *TaxComponent) SetSequence(v *int32) {
	this.GetFieldData().SetInt32(TaxComponentFieldSequence, v)
}

func (this TaxComponent) GetAffectSubsequentBaseOverride() *bool {
	return this.GetFieldData().GetBool(TaxComponentFieldAffectSubsequentBaseOverride)
}

func (this *TaxComponent) SetAffectSubsequentBaseOverride(v *bool) {
	this.GetFieldData().SetBool(TaxComponentFieldAffectSubsequentBaseOverride, v)
}
