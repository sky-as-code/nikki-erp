package models

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

// The template-owned fields a variant exposes as its own.
//
// A variant's display name, category and lifecycle all live on its template: storing a second
// copy on every variant would go stale the moment the template is edited, and reading the
// template separately costs a query per variant. So these are virtual — no column, filled at
// read time by ProductVariantDomainService — and exposed here as ordinary getters.
//
// There are deliberately no setters. A virtual value is derived, never authored: the only writer
// is FillFromTemplate below.

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

// FillFromTemplate copies the template-owned display fields onto this variant. The values are
// virtual: they exist on the variant only for the duration of a read.
//
// A nil template leaves every field nil rather than writing zero values. A variant whose template
// is missing must read as "unknown", not as a product with an empty name.
func (this *ProductVariant) FillFromTemplate(template *ProductTemplate) {
	if template == nil {
		return
	}

	fields := this.GetFieldData()
	fields.SetLangJson(ProductVariantFieldTemplateName, template.GetName())
	fields.SetLangJson(ProductVariantFieldTemplateShortName, template.GetShortName())
	fields.SetLangJson(ProductVariantFieldTemplateDescription, template.GetDescription())
	fields.SetLangJson(ProductVariantFieldTemplateSalesDescription, template.GetSalesDescription())
	fields.SetLangJson(ProductVariantFieldTemplatePurchaseDescription, template.GetPurchaseDescription())
	fields.SetModelId(ProductVariantFieldTemplateCategoryId, template.GetCategoryId())
	fields.SetModelId(ProductVariantFieldTemplateBrandId, template.GetBrandId())
	fields.SetModelId(ProductVariantFieldTemplateProductTypeId, template.GetProductTypeId())
	fields.SetBool(ProductVariantFieldTemplateSaleOk, template.GetSaleOk())

	if status := template.GetStatus(); status != nil {
		value := string(*status)
		fields.SetString(ProductVariantFieldTemplateStatus, &value)
	}
}

// TemplateVirtualFields lists the virtual field names in one place, so a caller can tell whether
// a request needs the template read at all.
var TemplateVirtualFields = []string{
	ProductVariantFieldTemplateName,
	ProductVariantFieldTemplateShortName,
	ProductVariantFieldTemplateDescription,
	ProductVariantFieldTemplateSalesDescription,
	ProductVariantFieldTemplatePurchaseDescription,
	ProductVariantFieldTemplateCategoryId,
	ProductVariantFieldTemplateBrandId,
	ProductVariantFieldTemplateProductTypeId,
	ProductVariantFieldTemplateStatus,
	ProductVariantFieldTemplateSaleOk,
}

// TemplateSourceField maps a variant's virtual field to the template field it derives from.
// It drives both the batched fill and the search rewrite (`template_name` -> `template.name`),
// so the two can never disagree about where a value comes from.
var TemplateSourceField = map[string]string{
	ProductVariantFieldTemplateName:                ProductTemplateFieldName,
	ProductVariantFieldTemplateShortName:           ProductTemplateFieldShortName,
	ProductVariantFieldTemplateDescription:         ProductTemplateFieldDescription,
	ProductVariantFieldTemplateSalesDescription:    ProductTemplateFieldSalesDescription,
	ProductVariantFieldTemplatePurchaseDescription: ProductTemplateFieldPurchaseDescription,
	ProductVariantFieldTemplateCategoryId:          ProductTemplateFieldCategoryId,
	ProductVariantFieldTemplateBrandId:             ProductTemplateFieldBrandId,
	ProductVariantFieldTemplateProductTypeId:       ProductTemplateFieldProductTypeId,
	ProductVariantFieldTemplateStatus:              ProductTemplateFieldStatus,
	ProductVariantFieldTemplateSaleOk:              ProductTemplateFieldSaleOk,
}
