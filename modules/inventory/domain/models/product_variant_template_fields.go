package models

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

// The template-owned fields a variant exposes as its own. Each template_{x} field is declared in
// product_variant.json as a related computed field copying template.{x}, filled by the engine on
// read with one batched template query per page; a stored copy would go stale when the template is
// edited. There are deliberately no setters: the engine is the only writer of a computed value.

func (this ProductVariant) GetTemplateName() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductVariantFieldTemplateName)
}

func (this ProductVariant) GetTemplateShortName() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductVariantFieldTemplateShortName)
}

func (this ProductVariant) GetTemplateDescription() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductVariantFieldTemplateDescription)
}

func (this ProductVariant) GetTemplateSalesDescription() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductVariantFieldTemplateSalesDescription)
}

func (this ProductVariant) GetTemplatePurchaseDescription() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductVariantFieldTemplatePurchaseDescription)
}

func (this ProductVariant) GetTemplateCategoryId() *model.Id {
	return this.GetFieldData().GetModelId(ProductVariantFieldTemplateCategoryId)
}

func (this ProductVariant) GetTemplateBrandId() *model.Id {
	return this.GetFieldData().GetModelId(ProductVariantFieldTemplateBrandId)
}

func (this ProductVariant) GetTemplateProductTypeId() *model.Id {
	return this.GetFieldData().GetModelId(ProductVariantFieldTemplateProductTypeId)
}

// GetTemplateStatus is the template's lifecycle, not the variant's: an active variant of a
// discontinued template is a real state, so the two must stay readable apart.
func (this ProductVariant) GetTemplateStatus() *ProductTemplateStatus {
	status := this.GetFieldData().GetString(ProductVariantFieldTemplateStatus)
	if status == nil {
		return nil
	}
	return WrapProductTemplateStatus(*status)
}

func (this ProductVariant) GetTemplateSaleOk() *bool {
	return this.GetFieldData().GetBool(ProductVariantFieldTemplateSaleOk)
}
