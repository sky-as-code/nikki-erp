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

	ProductTypeId string
	CategoryId    string
	BrandId       string

	SaleOk     bool
	PurchaseOk bool

	// Sku and Barcode are variant-owned: with several variants a template-level value would have
	// nothing to refer to.
	Sku     string
	Barcode string

	// ImageId, Weight and the dimensions fall back to the template when the variant does not
	// override them. Null on the variant means "inherit", not "none".
	ImageId string
	Weight  *decimal.Decimal
	Length  *decimal.Decimal
	Width   *decimal.Decimal
	Height  *decimal.Decimal

	TemplateStatus string
	VariantStatus  string

	IsTemplateArchived bool
	IsVariantArchived  bool
}

// IsSelectable reports whether this product may be chosen for a new transaction. Both levels
// must permit it: an archived template withdraws every variant under it, and a variant may be
// withdrawn on its own.
func (this EffectiveProduct) IsSelectable() bool {
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
		"image_id":            this.ImageId,
		"product_type_id":     this.ProductTypeId,
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

// AttributeOptions is one attribute of a template together with the values that template allows.
type AttributeOptions struct {
	AttributeId string
	Mode        models.VariantCreationMode

	// ValueIds are the template-scoped allowed values; a value the template does not allow can
	// never enter a combination.
	ValueIds []string
}

// VariantSyncPlan is the difference between the combinations a template should have and the
// variants it does have. A plan rather than an action: synchronization must never invalidate a
// variant a transaction already references, so the caller decides what to do with Obsolete —
// archive it when used, delete it only when provably unused.
type VariantSyncPlan struct {
	// ToCreate are combinations with no variant yet.
	ToCreate []string

	// Obsolete are existing non-archived variants whose combination is no longer valid.
	Obsolete []string

	// Unchanged are combinations that already have a variant.
	Unchanged []string
}
