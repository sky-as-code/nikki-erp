package product

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The Products data shapes live here, in the abstraction layer, rather than in domain/services.
// The implementation depends on the contract, never the reverse: domain/services imports this
// package, and this package must never import domain/services.

// EffectiveProduct is the flattened view of a variant that consumer modules read.
//
// It exists so that Sales, Stock, Purchase and reporting do not each re-derive which fields come
// from the template and which from the variant. Every module implementing that inheritance
// itself is exactly the drift AC-PROD-032 forbids. See BR §7.5.
type EffectiveProduct struct {
	VariantId  string
	TemplateId string

	// Name is the template's; DisplayName appends the variant's attribute values to it. The
	// variant owns neither: a stored variant name drifts from the combination it describes.
	// See BR §5.5 and AC-PROD-013.
	Name        *model.LangJson
	DisplayName string

	ProductTypeId string
	CategoryId    string
	BrandId       string

	SaleOk     bool
	PurchaseOk bool

	// Sku and Barcode are variant-owned: with several variants a template-level value would
	// have nothing to refer to. See BR §6.9.1.
	Sku     string
	Barcode string

	// ImageId, Weight and the dimensions fall back to the template when the variant does not
	// override them. Null on the variant means "inherit", not "none". See BR §5.4.
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
// withdrawn on its own. See AC-PROD-019 and AC-PROD-020.
func (this EffectiveProduct) IsSelectable() bool {
	if this.IsTemplateArchived || this.IsVariantArchived {
		return false
	}
	if this.TemplateStatus == models.ProductTemplateStatusDiscontinued.String() {
		return false
	}
	return this.VariantStatus != models.ProductVariantStatusDiscontinued.String()
}

// ToFieldMap renders the effective product as the flat map shape consumer modules already store
// alongside their transaction lines.
//
// The keys are deliberately the ones the previous flat product model used, so that a consumer
// snapshotting "the product" keeps producing the same payload after the Template/Variant split.
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

	// Mode decides whether this selection contributes to variant identity at all. A NEVER
	// attribute is configuration carried alongside the product, not part of what makes the
	// variant distinct.
	Mode models.VariantCreationMode
}

// AttributeOptions is one attribute of a template together with the values that template allows.
type AttributeOptions struct {
	AttributeId string
	Mode        models.VariantCreationMode

	// ValueIds are the template-scoped allowed values, which is what a variant may be built
	// from. A value the template does not allow can never enter a combination.
	ValueIds []string
}

// VariantSyncPlan is the difference between the combinations a template should have and the
// variants it does have.
//
// It is deliberately a plan rather than an action: BR §8.5 requires that synchronization never
// invalidates a variant a transaction already references, so the caller decides what to do with
// Obsolete — archive it when it has been used, delete it only when provably unused.
type VariantSyncPlan struct {
	// ToCreate are combinations with no variant yet.
	ToCreate []string

	// Obsolete are existing non-archived variants whose combination is no longer valid.
	Obsolete []string

	// Unchanged are combinations that already have a variant.
	Unchanged []string
}
