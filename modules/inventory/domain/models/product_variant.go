package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// ProductVariantStatus lets a single variant be discontinued while its siblings stay on sale.
// Independent of is_archived. See BR §6.2.2.
type ProductVariantStatus string

const (
	ProductVariantStatusActive       = ProductVariantStatus("active")
	ProductVariantStatusDiscontinued = ProductVariantStatus("discontinued")
)

func (this ProductVariantStatus) String() string {
	return string(this)
}

func WrapProductVariantStatus(s string) *ProductVariantStatus {
	st := ProductVariantStatus(s)
	return &st
}

// ArchiveSource records why a variant was archived, so that unarchiving a template restores only
// the variants it cascaded to and leaves deliberately archived ones alone. A plain boolean cannot
// express that difference. See BR §8.9 and BR-PROD-TPL-003.
type ArchiveSource string

const (
	// ArchiveSourceUser is a deliberate archive by a user. Never undone by a template cascade.
	ArchiveSourceUser = ArchiveSource("user")

	// ArchiveSourceTemplateCascade is an archive that followed from archiving the template.
	// Unarchiving the template unarchives exactly these.
	ArchiveSourceTemplateCascade = ArchiveSource("template_cascade")

	// ArchiveSourceSystemSync is an archive performed by variant synchronization when a
	// combination stopped being valid.
	ArchiveSourceSystemSync = ArchiveSource("system_sync")
)

func (this ArchiveSource) String() string {
	return string(this)
}

func WrapArchiveSource(s string) *ArchiveSource {
	src := ArchiveSource(s)
	return &src
}

const (
	ProductVariantSchemaName = "inventory_product_variant"

	ProductVariantFieldId                = basemodel.FieldId
	ProductVariantFieldProductTemplateId = "product_template_id"
	ProductVariantFieldCombinationKey    = "combination_key"
	ProductVariantFieldSku               = "sku"
	ProductVariantFieldPrimaryBarcode    = "primary_barcode"
	ProductVariantFieldIsMaterialized    = "is_materialized"
	ProductVariantFieldVariantImageId    = "variant_image_id"
	ProductVariantFieldWeight            = "weight"
	ProductVariantFieldLength            = "length"
	ProductVariantFieldWidth             = "width"
	ProductVariantFieldHeight            = "height"
	ProductVariantFieldStatus            = "status"
	ProductVariantFieldArchiveSource     = "archive_source"
	ProductVariantFieldOrgId             = "org_id"

	// Computed fields, copied from the owning template when a variant is read. They have no
	// database column: each is declared in product_variant.json as a related computed field
	// (template_name copies template.name, ...), filled by the engine's computed-field layer.
	ProductVariantFieldTemplateName                = "template_name"
	ProductVariantFieldTemplateShortName           = "template_short_name"
	ProductVariantFieldTemplateDescription         = "template_description"
	ProductVariantFieldTemplateSalesDescription    = "template_sales_description"
	ProductVariantFieldTemplatePurchaseDescription = "template_purchase_description"
	ProductVariantFieldTemplateCategoryId          = "template_category_id"
	ProductVariantFieldTemplateBrandId             = "template_brand_id"
	ProductVariantFieldTemplateProductTypeId       = "template_product_type_id"
	ProductVariantFieldTemplateStatus              = "template_status"
	ProductVariantFieldTemplateSaleOk              = "template_sale_ok"

	ProductVariantEdgeTemplate = "template"

	// EmptyCombinationKey is the combination of a template that has no variant-generating
	// attributes. Such a template still gets exactly one concrete variant, which is what
	// transactions reference. See BR §4.5 and AC-PROD-008.
	EmptyCombinationKey = ""
)

//go:embed product_variant.json
var productVariantSchemaJson string

func ProductVariantSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(productVariantSchemaJson)
}

type ProductVariant struct {
	basemodel.DynamicModelBase
}

func NewProductVariant() *ProductVariant {
	return &ProductVariant{basemodel.NewDynamicModel()}
}

func NewProductVariantFrom(src dmodel.DynamicFields) *ProductVariant {
	return &ProductVariant{basemodel.NewDynamicModel(src)}
}

func (this ProductVariant) GetProductTemplateId() *model.Id {
	return this.GetFieldData().GetModelId(ProductVariantFieldProductTemplateId)
}

func (this *ProductVariant) SetProductTemplateId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductVariantFieldProductTemplateId, v)
}

// GetCombinationKey returns the normalized identity of this variant's attribute-value
// combination. An empty string is a valid key, not a missing one.
func (this ProductVariant) GetCombinationKey() *string {
	return this.GetFieldData().GetString(ProductVariantFieldCombinationKey)
}

func (this *ProductVariant) SetCombinationKey(v *string) {
	this.GetFieldData().SetString(ProductVariantFieldCombinationKey, v)
}

func (this ProductVariant) GetSku() *string {
	return this.GetFieldData().GetString(ProductVariantFieldSku)
}

func (this *ProductVariant) SetSku(v *string) {
	this.GetFieldData().SetString(ProductVariantFieldSku, v)
}

func (this ProductVariant) GetPrimaryBarcode() *string {
	return this.GetFieldData().GetString(ProductVariantFieldPrimaryBarcode)
}

func (this *ProductVariant) SetPrimaryBarcode(v *string) {
	this.GetFieldData().SetString(ProductVariantFieldPrimaryBarcode, v)
}

func (this ProductVariant) GetIsMaterialized() *bool {
	return this.GetFieldData().GetBool(ProductVariantFieldIsMaterialized)
}

func (this *ProductVariant) SetIsMaterialized(v *bool) {
	this.GetFieldData().SetBool(ProductVariantFieldIsMaterialized, v)
}

// GetVariantImageId returns this variant's image override. Nil means fall back to the template's
// default image rather than "no image". See BR §8.4 and AC-PROD-014.
func (this ProductVariant) GetVariantImageId() *model.Id {
	return this.GetFieldData().GetModelId(ProductVariantFieldVariantImageId)
}

func (this *ProductVariant) SetVariantImageId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductVariantFieldVariantImageId, v)
}

func (this ProductVariant) GetWeight() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductVariantFieldWeight)
}

func (this *ProductVariant) SetWeight(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductVariantFieldWeight, v)
}

func (this ProductVariant) GetLength() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductVariantFieldLength)
}

func (this *ProductVariant) SetLength(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductVariantFieldLength, v)
}

func (this ProductVariant) GetWidth() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductVariantFieldWidth)
}

func (this *ProductVariant) SetWidth(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductVariantFieldWidth, v)
}

func (this ProductVariant) GetHeight() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductVariantFieldHeight)
}

func (this *ProductVariant) SetHeight(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductVariantFieldHeight, v)
}

func (this ProductVariant) GetStatus() *ProductVariantStatus {
	s := this.GetFieldData().GetString(ProductVariantFieldStatus)
	if s == nil {
		return nil
	}
	return WrapProductVariantStatus(*s)
}

func (this *ProductVariant) SetStatus(v *ProductVariantStatus) {
	if v == nil {
		this.GetFieldData().SetString(ProductVariantFieldStatus, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(ProductVariantFieldStatus, &s)
}

func (this ProductVariant) GetArchiveSource() *ArchiveSource {
	s := this.GetFieldData().GetString(ProductVariantFieldArchiveSource)
	if s == nil {
		return nil
	}
	return WrapArchiveSource(*s)
}

func (this *ProductVariant) SetArchiveSource(v *ArchiveSource) {
	if v == nil {
		this.GetFieldData().SetString(ProductVariantFieldArchiveSource, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(ProductVariantFieldArchiveSource, &s)
}

func (this ProductVariant) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(ProductVariantFieldOrgId)
}

func (this *ProductVariant) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductVariantFieldOrgId, v)
}

// IsSelectable reports whether this variant may be chosen for a new transaction. Archiving and
// discontinuing are separate concepts, and either one alone is enough to withdraw a variant from
// new business. See AC-PROD-019 and AC-PROD-020.
func (this ProductVariant) IsSelectable() bool {
	if archived := this.IsArchived(); archived != nil && *archived {
		return false
	}
	status := this.GetStatus()
	return status == nil || *status == ProductVariantStatusActive
}
