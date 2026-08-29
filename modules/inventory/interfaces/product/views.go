package product

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The wire shapes of the Products capabilities. They live here, not transport/, because the
// engine actions that build them are in dynamicengines/, which may not import outward. The engine
// passes an action's Data to JSON untouched, so these tagged structs define the JSON contract.
//
// Weights and dimensions travel as strings: most clients parse a JSON number as float64, losing
// precision before the value reaches a caller.

// EffectiveProductView is the flattened product a consumer reads instead of joining the template
// and variant itself.
type EffectiveProductView struct {
	VariantId  string `json:"variant_id"`
	TemplateId string `json:"template_id"`

	Name        *model.LangJson `json:"name,omitempty"`
	DisplayName string          `json:"display_name"`

	ProductTypeId string `json:"product_type_id,omitempty"`
	CategoryId    string `json:"category_id,omitempty"`
	BrandId       string `json:"brand_id,omitempty"`

	SaleOk     bool `json:"sale_ok"`
	PurchaseOk bool `json:"purchase_ok"`

	Sku     string `json:"sku,omitempty"`
	Barcode string `json:"barcode,omitempty"`

	ImageId string `json:"image_id,omitempty"`
	Weight  string `json:"weight,omitempty"`
	Length  string `json:"length,omitempty"`
	Width   string `json:"width,omitempty"`
	Height  string `json:"height,omitempty"`

	TemplateStatus string `json:"template_status,omitempty"`
	VariantStatus  string `json:"variant_status,omitempty"`

	IsTemplateArchived bool `json:"is_template_archived"`
	IsVariantArchived  bool `json:"is_variant_archived"`

	// IsSelectable saves every consumer from re-deriving the archive and status rules.
	IsSelectable bool `json:"is_selectable"`
}

func NewEffectiveProductView(product EffectiveProduct) EffectiveProductView {
	return EffectiveProductView{
		VariantId:          product.VariantId,
		TemplateId:         product.TemplateId,
		Name:               product.Name,
		DisplayName:        product.DisplayName,
		ProductTypeId:      product.ProductTypeId,
		CategoryId:         product.CategoryId,
		BrandId:            product.BrandId,
		SaleOk:             product.SaleOk,
		PurchaseOk:         product.PurchaseOk,
		Sku:                product.Sku,
		Barcode:            product.Barcode,
		ImageId:            product.ImageId,
		Weight:             decimalToString(product.Weight),
		Length:             decimalToString(product.Length),
		Width:              decimalToString(product.Width),
		Height:             decimalToString(product.Height),
		TemplateStatus:     product.TemplateStatus,
		VariantStatus:      product.VariantStatus,
		IsTemplateArchived: product.IsTemplateArchived,
		IsVariantArchived:  product.IsVariantArchived,
		IsSelectable:       product.IsSelectable(),
	}
}

// AttributeSelectionInput is one attribute-value choice as it arrives in a request body.
type AttributeSelectionInput struct {
	AttributeId string `json:"attribute_id"`
	ValueId     string `json:"value_id"`
	Mode        string `json:"mode"`
}

// ToSelection maps the request shape onto the domain type. An unrecognized mode falls back to
// INSTANT, keeping the attribute in the combination; dropping it would resolve the caller to a
// different variant than they asked for.
func (this AttributeSelectionInput) ToSelection() AttributeSelection {
	return AttributeSelection{
		AttributeId: this.AttributeId,
		ValueId:     this.ValueId,
		Mode:        wrapVariantCreationMode(this.Mode),
	}
}

type ResolveProductSelectionView struct {
	VariantId      string `json:"variant_id,omitempty"`
	CombinationKey string `json:"combination_key"`
	Materialized   bool   `json:"materialized"`
}

func NewResolveProductSelectionView(data ResolveProductSelectionResultData) ResolveProductSelectionView {
	return ResolveProductSelectionView{
		VariantId:      data.VariantId,
		CombinationKey: data.CombinationKey,
		Materialized:   data.Materialized,
	}
}

type GenerateVariantsView struct {
	CreatedVariantIds []string `json:"created_variant_ids"`

	// ObsoleteVariantIds are reported rather than removed: one a transaction already references
	// must be archived, not deleted, and only the caller knows which.
	ObsoleteVariantIds []string `json:"obsolete_variant_ids"`

	UnchangedCount int `json:"unchanged_count"`
}

func NewGenerateVariantsView(data GenerateVariantsResultData) GenerateVariantsView {
	return GenerateVariantsView{
		CreatedVariantIds:  orEmpty(data.CreatedVariantIds),
		ObsoleteVariantIds: orEmpty(data.ObsoleteVariantIds),
		UnchangedCount:     data.UnchangedCount,
	}
}

// decimalToString renders an optional decimal for JSON. An absent value stays absent rather than
// becoming "0", because null means "inherited from the template", not zero.
func decimalToString(value *decimal.Decimal) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func wrapVariantCreationMode(mode string) models.VariantCreationMode {
	if mode == "" {
		return models.VariantCreationModeInstant
	}
	wrapped := models.WrapVariantCreationMode(mode)
	switch *wrapped {
	case models.VariantCreationModeInstant,
		models.VariantCreationModeDynamic,
		models.VariantCreationModeNever:
		return *wrapped
	default:
		return models.VariantCreationModeInstant
	}
}

// orEmpty renders a nil slice as [] rather than null, so a client can iterate the response
// without a null check.
func orEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
