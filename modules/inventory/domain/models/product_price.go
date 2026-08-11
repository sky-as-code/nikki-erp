package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// ProductPriceStatus is the pricing workflow state. Only an approved row is ever applied to a
// product: draft lets a price be prepared without taking effect, and expired keeps a superseded
// price readable rather than deleting it. See BR §6.12.
type ProductPriceStatus string

const (
	ProductPriceStatusDraft    = ProductPriceStatus("draft")
	ProductPriceStatusApproved = ProductPriceStatus("approved")
	ProductPriceStatusExpired  = ProductPriceStatus("expired")
)

func (this ProductPriceStatus) String() string {
	return string(this)
}

func WrapProductPriceStatus(s string) *ProductPriceStatus {
	st := ProductPriceStatus(s)
	return &st
}

const (
	ProductPriceSchemaName = "inventory_product_price"

	ProductPriceFieldId                = basemodel.FieldId
	ProductPriceFieldProductTemplateId = "product_template_id"
	ProductPriceFieldProductVariantId  = "product_variant_id"
	ProductPriceFieldPrice             = "price"
	ProductPriceFieldEffectiveFrom     = "effective_from"
	ProductPriceFieldEffectiveTo       = "effective_to"
	ProductPriceFieldStatus            = "status"
	ProductPriceFieldOrgId             = "org_id"

	ProductPriceEdgeTemplate = "template"
	ProductPriceEdgeVariant  = "variant"
)

//go:embed product_price.json
var productPriceSchemaJson string

func ProductPriceSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(productPriceSchemaJson)
}

type ProductPrice struct {
	basemodel.DynamicModelBase
}

func NewProductPrice() *ProductPrice {
	return &ProductPrice{basemodel.NewDynamicModel()}
}

func NewProductPriceFrom(src dmodel.DynamicFields) *ProductPrice {
	return &ProductPrice{basemodel.NewDynamicModel(src)}
}

// GetProductTemplateId returns the template this price applies to, or nil when the price targets
// a variant instead. Exactly one of the two is set; the schema's exclusive_required_fields group
// enforces it. See BR §6.12 rule 1.
func (this ProductPrice) GetProductTemplateId() *model.Id {
	return this.GetFieldData().GetModelId(ProductPriceFieldProductTemplateId)
}

func (this *ProductPrice) SetProductTemplateId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductPriceFieldProductTemplateId, v)
}

// GetProductVariantId returns the variant this price applies to, or nil when the price targets
// the template instead.
func (this ProductPrice) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(ProductPriceFieldProductVariantId)
}

func (this *ProductPrice) SetProductVariantId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductPriceFieldProductVariantId, v)
}

func (this ProductPrice) GetPrice() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductPriceFieldPrice)
}

func (this *ProductPrice) SetPrice(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductPriceFieldPrice, v)
}

// GetEffectiveFrom returns the first day this price applies. Nil means it applies from the
// beginning of time, which is the normal case for a standing price.
func (this ProductPrice) GetEffectiveFrom() *model.ModelDate {
	return this.GetFieldData().GetModelDate(ProductPriceFieldEffectiveFrom)
}

func (this *ProductPrice) SetEffectiveFrom(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(ProductPriceFieldEffectiveFrom, v)
}

// GetEffectiveTo returns the last day this price applies. Nil means it does not expire.
func (this ProductPrice) GetEffectiveTo() *model.ModelDate {
	return this.GetFieldData().GetModelDate(ProductPriceFieldEffectiveTo)
}

func (this *ProductPrice) SetEffectiveTo(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(ProductPriceFieldEffectiveTo, v)
}

func (this ProductPrice) GetStatus() *ProductPriceStatus {
	s := this.GetFieldData().GetString(ProductPriceFieldStatus)
	if s == nil {
		return nil
	}
	return WrapProductPriceStatus(*s)
}

func (this *ProductPrice) SetStatus(v *ProductPriceStatus) {
	if v == nil {
		this.GetFieldData().SetString(ProductPriceFieldStatus, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(ProductPriceFieldStatus, &s)
}

func (this ProductPrice) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(ProductPriceFieldOrgId)
}

func (this *ProductPrice) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductPriceFieldOrgId, v)
}
