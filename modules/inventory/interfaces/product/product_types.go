package product

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The Products data shapes live in this abstraction layer: domain/services imports this package,
// and this package must never import domain/services.

// EffectiveProduct is the flattened view of a variant consumer modules read, so no module
// re-derives which fields come from the template and which from the variant.
type EffectiveProduct struct {
	VariantId  string
	TemplateId string

	// Name is the template's; DisplayName appends the variant's attribute values to it. The
	// variant owns neither: a stored variant name drifts from the combination it describes.
	Name        *model.LangJson
	DisplayName string

	ProductType string
	IsCombo     bool
	CategoryId  string
	BrandId     string

	SaleOk     bool
	PurchaseOk bool

	// Sku and Barcode are variant-owned: with several variants a template-level value would have
	// nothing to refer to.
	Sku     string
	Barcode string

	// ImageFileId, Weight and the dimensions fall back to the template when the variant does not
	// override them. Null on the variant means "inherit", not "none".
	ImageFileId string
	Weight      *decimal.Decimal
	Length      *decimal.Decimal
	Width       *decimal.Decimal
	Height      *decimal.Decimal

	TemplateStatus string
	VariantStatus  string

	// VariantKind is complete for a sellable variant and base for the instant-only parent of a
	// mixed template's completes. A base is never transacted.
	VariantKind string

	IsTemplateArchived bool
	IsVariantArchived  bool
}

// IsSelectable reports whether this product may be chosen for a new transaction. Both levels
// must permit it: an archived template withdraws every variant under it, and a variant may be
// withdrawn on its own. Only a complete variant names a product; a base lacks its dynamic values.
func (this EffectiveProduct) IsSelectable() bool {
	if this.VariantKind != models.VariantKindComplete.String() {
		return false
	}
	if this.IsTemplateArchived || this.IsVariantArchived {
		return false
	}
	if this.TemplateStatus == models.ProductTemplateStatusDiscontinued.String() {
		return false
	}
	return this.VariantStatus != models.ProductVariantStatusDiscontinued.String()
}

// ToFieldMap renders the effective product as the flat map consumer modules store alongside
// their transaction lines. The keys are those of the pre-split flat product model, so a consumer
// snapshotting "the product" keeps producing the same payload.
func (this EffectiveProduct) ToFieldMap() map[string]any {
	return map[string]any{
		"id":                  this.VariantId,
		"variant_id":          this.VariantId,
		"product_template_id": this.TemplateId,
		"name":                this.DisplayName,
		"template_name":       this.Name,
		"sku":                 this.Sku,
		"barcode":             this.Barcode,
		"image_file_id":       this.ImageFileId,
		"product_type":        this.ProductType,
		"is_combo":            this.IsCombo,
		"category_id":         this.CategoryId,
		"brand_id":            this.BrandId,
		"sale_ok":             this.SaleOk,
		"purchase_ok":         this.PurchaseOk,
		"status":              this.VariantStatus,
		"is_archived":         this.IsVariantArchived,
	}
}

// AttributeSelection is one attribute-value choice going into a variant's identity.
type AttributeSelection struct {
	AttributeId string
	ValueId     string

	// Mode decides whether this selection contributes to variant identity. A NEVER attribute is
	// configuration carried alongside the product, not part of the variant's identity.
	Mode models.VariantCreationMode
}
