package models

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

// The template-owned fields a variant exposes as its own.
//
// A variant's display name, category and lifecycle all live on its template: storing a second
// copy on every variant would go stale the moment the template is edited. Each template_{x}
// field is declared in product_variant.json as a related computed field copying template.{x},
// and the engine's computed-field layer fills it on every read with one batched template query
// per page. These getters only read the filled values.
//
// There are deliberately no setters. A computed value is derived, never authored: the engine is
// its only writer.

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

// GetTemplateStatus is the template's lifecycle, which is not this variant's own: an active
// variant of a discontinued template is a real state, so the two must stay readable apart.
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
