package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// A combination exclusion forbids two template values from appearing together in one variant.
// The values are NO ACTION rather than RESTRICT: deleting the template cascades to both the
// exclusion and, through the line, to the values, and RESTRICT would refuse that one statement
// depending on trigger order. NO ACTION still refuses deleting a referenced value on its own.

const (
	ProductTemplateCombinationExclusionSchemaName = "inventory_product_template_combination_exclusion"

	ProductTemplateCombinationExclusionFieldId                = basemodel.FieldId
	ProductTemplateCombinationExclusionFieldProductTemplateId = "product_template_id"
	ProductTemplateCombinationExclusionFieldLeftValueId       = "left_template_attribute_value_id"
	ProductTemplateCombinationExclusionFieldRightValueId      = "right_template_attribute_value_id"
	ProductTemplateCombinationExclusionFieldOrgId             = "org_id"

	ProductTemplateCombinationExclusionEdgeTemplate   = "template"
	ProductTemplateCombinationExclusionEdgeLeftValue  = "left_value"
	ProductTemplateCombinationExclusionEdgeRightValue = "right_value"
)

//go:embed product_template_combination_exclusion.json
var productTemplateCombinationExclusionSchemaJson string

func ProductTemplateCombinationExclusionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(productTemplateCombinationExclusionSchemaJson)
}

type ProductTemplateCombinationExclusion struct {
	basemodel.DynamicModelBase
}

func NewProductTemplateCombinationExclusion() *ProductTemplateCombinationExclusion {
	return &ProductTemplateCombinationExclusion{basemodel.NewDynamicModel()}
}

func NewProductTemplateCombinationExclusionFrom(src dmodel.DynamicFields) *ProductTemplateCombinationExclusion {
	return &ProductTemplateCombinationExclusion{basemodel.NewDynamicModel(src)}
}

func (this ProductTemplateCombinationExclusion) GetProductTemplateId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateCombinationExclusionFieldProductTemplateId)
}

func (this *ProductTemplateCombinationExclusion) SetProductTemplateId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateCombinationExclusionFieldProductTemplateId, v)
}

func (this ProductTemplateCombinationExclusion) GetLeftValueId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateCombinationExclusionFieldLeftValueId)
}

func (this *ProductTemplateCombinationExclusion) SetLeftValueId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateCombinationExclusionFieldLeftValueId, v)
}

func (this ProductTemplateCombinationExclusion) GetRightValueId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateCombinationExclusionFieldRightValueId)
}

func (this *ProductTemplateCombinationExclusion) SetRightValueId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateCombinationExclusionFieldRightValueId, v)
}

func (this ProductTemplateCombinationExclusion) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateCombinationExclusionFieldOrgId)
}

func (this *ProductTemplateCombinationExclusion) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateCombinationExclusionFieldOrgId, v)
}
