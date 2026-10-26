package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// ProductVariantStatus lets one variant be discontinued while its siblings stay on sale.
// Independent of is_archived.
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

// ArchiveSource records why a variant was archived, so unarchiving a template restores only the
// variants it cascaded to and leaves deliberately archived ones alone.
type ArchiveSource string

const (
	// ArchiveSourceUser is a deliberate archive by a user. Never undone by a template cascade.
	ArchiveSourceUser = ArchiveSource("user")

	// ArchiveSourceTemplateCascade followed from archiving the template; unarchiving the template
	// unarchives exactly these.
	ArchiveSourceTemplateCascade = ArchiveSource("template_cascade")

	// ArchiveSourceSystemSync is set by variant synchronization when a combination stopped being
	// valid.
	ArchiveSourceSystemSync = ArchiveSource("system_sync")
)

func (this ArchiveSource) String() string {
	return string(this)
}

func WrapArchiveSource(s string) *ArchiveSource {
	src := ArchiveSource(s)
	return &src
}

// VariantKind separates the sellable complete variant from the base of a mixed template.
type VariantKind string

const (
	// VariantKindComplete carries a value for every variant-creating attribute; the only kind a
	// transaction may name.
	VariantKindComplete = VariantKind("complete")

	// VariantKindBase carries only the instant values of a mixed (instant + dynamic) template and
	// anchors the complete variants resolved under it.
	VariantKindBase = VariantKind("base")
)

func (this VariantKind) String() string {
	return string(this)
}

func WrapVariantKind(s string) *VariantKind {
	k := VariantKind(s)
	return &k
}

const (
	ProductVariantSchemaName = "inventory_product_variant"

	ProductVariantFieldId                   = basemodel.FieldId
	ProductVariantFieldProductTemplateId    = "product_template_id"
	ProductVariantFieldCombinationKey       = "combination_key"
	ProductVariantFieldVariantKind          = "variant_kind"
	ProductVariantFieldBaseVariantId        = "base_variant_id"
	ProductVariantFieldSku                  = "sku"
	ProductVariantFieldPrimaryBarcode       = "primary_barcode"
	ProductVariantFieldVariantImageFilePath = "variant_image_file_path"
	ProductVariantFieldCost                 = "cost"
	ProductVariantFieldWeight               = "weight"
	ProductVariantFieldLength               = "length"
	ProductVariantFieldWidth                = "width"
	ProductVariantFieldHeight               = "height"
	ProductVariantFieldStatus               = "status"
	ProductVariantFieldArchiveSource        = "archive_source"
	ProductVariantFieldOrgId                = "org_id"

	// Computed fields copied from the owning template on read. No database column: each is declared
	// in product_variant.json as a related computed field and filled by the engine.
	ProductVariantFieldTemplateName                = "product_template_name"
	ProductVariantFieldTemplateShortName           = "product_template_short_name"
	ProductVariantFieldTemplateDescription         = "product_template_description"
	ProductVariantFieldTemplateSalesDescription    = "product_template_sales_description"
	ProductVariantFieldTemplatePurchaseDescription = "product_template_purchase_description"
	ProductVariantFieldTemplateCategoryId          = "product_template_category_id"
	ProductVariantFieldTemplateBrandId             = "product_template_brand_id"
	ProductVariantFieldTemplateProductType         = "product_template_product_type"
	ProductVariantFieldTemplateIsCombo             = "product_template_is_combo"
	ProductVariantFieldTemplateStatus              = "product_template_status"
	ProductVariantFieldTemplateSaleOk              = "product_template_sale_ok"
	ProductVariantFieldTemplatePurchaseOk          = "product_template_purchase_ok"

	// Pricing read model. Constants rather than literals because Sales reads these across a module
	// boundary, where a literal would survive a rename here and silently resolve to nothing.
	ProductVariantFieldTemplateBaseSalesPrice  = "product_template_base_sales_price"
	ProductVariantFieldSalesPriceExtraTotal    = "sales_price_extra_total"
	ProductVariantFieldEffectiveBaseSalesPrice = "effective_base_sales_price"
	ProductVariantFieldTemplateIsArchived      = "product_template_is_archived"
	ProductVariantFieldIsTransactionEligible   = "is_transaction_eligible"

	ProductVariantEdgeTemplate    = "product_template"
	ProductVariantEdgeBaseVariant = "base_variant"

	// EmptyCombinationKey is the combination of a template with no variant-generating attributes.
	// Such a template still gets exactly one concrete variant, which transactions reference.
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

// GetCombinationKey is the normalized identity of the attribute-value combination. An empty string
// is a valid key, not a missing one.
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

// GetVariantImageFilePath is an override; nil means fall back to the template's default image, not
// "no image".
func (this ProductVariant) GetVariantImageFilePath() *string {
	return this.GetFieldData().GetString(ProductVariantFieldVariantImageFilePath)
}

func (this *ProductVariant) SetVariantImageFilePath(v *string) {
	this.GetFieldData().SetString(ProductVariantFieldVariantImageFilePath, v)
}

func (this ProductVariant) GetCost() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductVariantFieldCost)
}

func (this *ProductVariant) SetCost(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductVariantFieldCost, v)
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
// discontinuing are separate, and either alone withdraws a variant.
func (this ProductVariant) IsSelectable() bool {
	if archived := this.IsArchived(); archived != nil && *archived {
		return false
	}
	if kind := this.GetVariantKind(); kind != nil && *kind == VariantKindBase {
		return false
	}
	status := this.GetStatus()
	return status == nil || *status == ProductVariantStatusActive
}

func (this ProductVariant) GetVariantKind() *VariantKind {
	s := this.GetFieldData().GetString(ProductVariantFieldVariantKind)
	if s == nil {
		return nil
	}
	return WrapVariantKind(*s)
}

func (this *ProductVariant) SetVariantKind(v *VariantKind) {
	if v == nil {
		this.GetFieldData().SetString(ProductVariantFieldVariantKind, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(ProductVariantFieldVariantKind, &s)
}

func (this ProductVariant) GetBaseVariantId() *model.Id {
	return this.GetFieldData().GetModelId(ProductVariantFieldBaseVariantId)
}

func (this *ProductVariant) SetBaseVariantId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductVariantFieldBaseVariantId, v)
}

func (this ProductVariant) GetIsTransactionEligible() *bool {
	return this.GetFieldData().GetBool(ProductVariantFieldIsTransactionEligible)
}
