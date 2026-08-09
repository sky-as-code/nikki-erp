package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// UomType classifies a UoM by how its conversion factor relates to the category's
// Reference UoM. See BR-UOM-ESS-009.
type UomType string

const (
	// UomTypeReference is the category's base unit: factor is exactly 1, and there is
	// exactly one such UoM per category.
	UomTypeReference = UomType("reference")

	// UomTypeBiggerEqual is a UoM larger than or equal to the reference: factor >= 1.
	UomTypeBiggerEqual = UomType("bigger_equal")

	// UomTypeSmaller is a UoM smaller than the reference: 0 < factor < 1.
	UomTypeSmaller = UomType("smaller")
)

func (this UomType) String() string {
	return string(this)
}

func WrapUomType(s string) *UomType {
	t := UomType(s)
	return &t
}

const (
	UomSchemaName = "essential_uom"

	UomFieldId         = basemodel.FieldId
	UomFieldName       = "name"
	UomFieldSymbol     = "symbol"
	UomFieldCategoryId = "category_id"
	UomFieldUomType    = "uom_type"
	UomFieldFactor     = "factor"
	UomFieldRounding   = "rounding"
	UomFieldOrgId      = "org_id"

	UomEdgeCategory = "category"
)

//go:embed uom.json
var uomSchemaJson string

func UomSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(uomSchemaJson)
}

type Uom struct {
	basemodel.DynamicModelBase
}

func NewUom() *Uom {
	return &Uom{basemodel.NewDynamicModel()}
}

func NewUomFrom(src dmodel.DynamicFields) *Uom {
	return &Uom{basemodel.NewDynamicModel(src)}
}

func (this Uom) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(UomFieldName)
}

func (this *Uom) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(UomFieldName, v)
}

func (this Uom) GetSymbol() *string {
	return this.GetFieldData().GetString(UomFieldSymbol)
}

func (this *Uom) SetSymbol(v *string) {
	this.GetFieldData().SetString(UomFieldSymbol, v)
}

func (this Uom) GetCategoryId() *model.Id {
	return this.GetFieldData().GetModelId(UomFieldCategoryId)
}

func (this *Uom) SetCategoryId(v *model.Id) {
	this.GetFieldData().SetModelId(UomFieldCategoryId, v)
}

func (this Uom) GetUomType() *UomType {
	s := this.GetFieldData().GetString(UomFieldUomType)
	if s == nil {
		return nil
	}
	return WrapUomType(*s)
}

func (this *Uom) SetUomType(v *UomType) {
	if v == nil {
		this.GetFieldData().SetString(UomFieldUomType, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(UomFieldUomType, &s)
}

// GetFactor returns the conversion factor relative to the category's Reference UoM.
// Quantity in Reference UoM = Quantity in this UoM x factor. See BR-UOM-ESS-008.
func (this Uom) GetFactor() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(UomFieldFactor)
}

func (this *Uom) SetFactor(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(UomFieldFactor, v)
}

// GetRounding returns the rounding precision of this UoM, in the range 0 <= rounding <= 1.
// It is independent of the conversion factor. See BR-UOM-ESS-015 and BR-UOM-ESS-016.
func (this Uom) GetRounding() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(UomFieldRounding)
}

func (this *Uom) SetRounding(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(UomFieldRounding, v)
}

func (this Uom) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(UomFieldOrgId)
}

func (this *Uom) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(UomFieldOrgId, v)
}
