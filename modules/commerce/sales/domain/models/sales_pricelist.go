package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesPricelistSchemaName = "sales_pricelist"

	SalesPricelistFieldId             = "id"
	SalesPricelistFieldOrgId          = "org_id"
	SalesPricelistFieldCode           = "code"
	SalesPricelistFieldName           = "name"
	SalesPricelistFieldDescription    = "description"
	SalesPricelistFieldCurrencyId     = "currency_id"
	SalesPricelistFieldIsDefault      = "is_default"
	SalesPricelistFieldSalesChannelId = "sales_channel_id"
	SalesPricelistFieldSalesPointId   = "sales_point_id"
	SalesPricelistFieldValidFrom      = "valid_from"
	SalesPricelistFieldValidUntil     = "valid_until"
	SalesPricelistFieldPriority       = "priority"

	SalesPricelistEdgeSalesChannel = "sales_channel"
	SalesPricelistEdgeSalesPoint   = "sales_point"
)

const (
	SalesPricelistItemSchemaName = "sales_pricelist_item"

	SalesPricelistItemFieldId               = "id"
	SalesPricelistItemFieldOrgId            = "org_id"
	SalesPricelistItemFieldSalesPricelistId = "sales_pricelist_id"
	SalesPricelistItemFieldProductVariantId = "product_variant_id"
	SalesPricelistItemFieldUomId            = "uom_id"
	SalesPricelistItemFieldPrice            = "price"
	SalesPricelistItemFieldMinQuantity      = "min_quantity"

	// Exactly one of the three id columns is set, chosen by applies_to; ALL_PRODUCTS sets none. The
	// schema cannot express "exactly one of these", so the domain service does.
	SalesPricelistItemFieldAppliesTo         = "applies_to"
	SalesPricelistItemFieldProductTemplateId = "product_template_id"
	SalesPricelistItemFieldProductCategoryId = "product_category_id"

	// Rule-level validity, distinct from the pricelist's own window.
	SalesPricelistItemFieldValidFrom = "valid_from"
	SalesPricelistItemFieldValidTo   = "valid_to"

	// Tie-break of last resort, with id after it.
	SalesPricelistItemFieldSequence = "sequence"

	// How the price is computed, and the operands each method needs.
	SalesPricelistItemFieldCalculationMethod = "calculation_method"
	SalesPricelistItemFieldDiscountPercent   = "discount_percent"
	SalesPricelistItemFieldBasePriceSource   = "base_price_source"
	SalesPricelistItemFieldBasePricelistId   = "base_pricelist_id"
	SalesPricelistItemFieldSurchargeAmount   = "surcharge_amount"
	SalesPricelistItemFieldRoundingIncrement = "rounding_increment"
	SalesPricelistItemFieldMinimumMargin     = "minimum_margin"
	SalesPricelistItemFieldMaximumMargin     = "maximum_margin"

	SalesPricelistItemEdgeSalesPricelist = "sales_pricelist"
)

// PricelistScope ranks how specifically a pricelist applies, which decides between two that both
// match. Specificity beats priority always: a point-scoped list wins over a channel-scoped one
// whatever their priority numbers. Priority only breaks ties between lists of the same scope.
type PricelistScope int

const (
	// PricelistScopeGlobal applies everywhere: both scope columns NULL.
	PricelistScopeGlobal PricelistScope = iota
	// PricelistScopeChannel applies to one sales channel.
	PricelistScopeChannel
	// PricelistScopePoint applies to one selling place, and is the most specific.
	PricelistScopePoint
)

//go:embed sales_pricelist.json
var salesPricelistSchemaJson string

func SalesPricelistSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPricelistSchemaJson)
}

//go:embed sales_pricelist_item.json
var salesPricelistItemSchemaJson string

func SalesPricelistItemSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPricelistItemSchemaJson)
}

// SalesPricelist is a set of prices that applies to some scope for some window. The same variant
// legitimately costs different amounts in different places, and a price scheduled for next month
// must not change what today's sales are charged.
type SalesPricelist struct {
	basemodel.DynamicModelBase
}

func NewSalesPricelist() *SalesPricelist {
	return &SalesPricelist{basemodel.NewDynamicModel()}
}

func NewSalesPricelistFrom(src dmodel.DynamicFields) *SalesPricelist {
	return &SalesPricelist{basemodel.NewDynamicModel(src)}
}

func (this SalesPricelist) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this SalesPricelist) GetCode() *string {
	return this.GetFieldData().GetString(SalesPricelistFieldCode)
}

func (this SalesPricelist) GetName() *string {
	return this.GetFieldData().GetString(SalesPricelistFieldName)
}

func (this SalesPricelist) GetDescription() *string {
	return this.GetFieldData().GetString(SalesPricelistFieldDescription)
}

func (this SalesPricelist) GetCurrencyId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPricelistFieldCurrencyId)
}

func (this SalesPricelist) GetIsDefault() *bool {
	return this.GetFieldData().GetBool(SalesPricelistFieldIsDefault)
}

func (this SalesPricelist) GetSalesChannelId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPricelistFieldSalesChannelId)
}

func (this SalesPricelist) GetSalesPointId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPricelistFieldSalesPointId)
}

func (this SalesPricelist) GetValidFrom() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesPricelistFieldValidFrom)
}

func (this SalesPricelist) GetValidUntil() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesPricelistFieldValidUntil)
}

func (this SalesPricelist) GetPriority() *int32 {
	return this.GetFieldData().GetInt32(SalesPricelistFieldPriority)
}

func (this SalesPricelist) GetIsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

// Scope reports how specifically this pricelist applies. A list naming both a point and a channel is
// point-scoped: the point already implies its channel, and the narrower answer wins.
func (this SalesPricelist) Scope() PricelistScope {
	if this.GetSalesPointId() != nil {
		return PricelistScopePoint
	}
	if this.GetSalesChannelId() != nil {
		return PricelistScopeChannel
	}
	return PricelistScopeGlobal
}

// The targets a pricelist rule may name, most specific first. Resolution walks them in this order,
// so the sequence here is the precedence.
const (
	PricelistAppliesToVariant     = "PRODUCT_VARIANT"
	PricelistAppliesToTemplate    = "PRODUCT_TEMPLATE"
	PricelistAppliesToCategory    = "PRODUCT_CATEGORY"
	PricelistAppliesToAllProducts = "ALL_PRODUCTS"
)

// How a rule arrives at its price.
const (
	PricelistMethodFixedPrice = "FIXED_PRICE"
	PricelistMethodDiscount   = "DISCOUNT"
	PricelistMethodFormula    = "FORMULA"
)

// What a FORMULA rule starts from. COST is a read of Inventory's number; Sales never writes it back.
const (
	PricelistBaseSourceBaseSalesPrice = "BASE_SALES_PRICE"
	PricelistBaseSourceOtherPricelist = "OTHER_PRICELIST"
	PricelistBaseSourceCost           = "COST"
)

// SalesPricelistItem is one price, for one variant, in one unit, from one quantity upward.
type SalesPricelistItem struct {
	basemodel.DynamicModelBase
}

func NewSalesPricelistItem() *SalesPricelistItem {
	return &SalesPricelistItem{basemodel.NewDynamicModel()}
}

func NewSalesPricelistItemFrom(src dmodel.DynamicFields) *SalesPricelistItem {
	return &SalesPricelistItem{basemodel.NewDynamicModel(src)}
}

func (this SalesPricelistItem) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this SalesPricelistItem) GetSalesPricelistId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPricelistItemFieldSalesPricelistId)
}

func (this SalesPricelistItem) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPricelistItemFieldProductVariantId)
}

func (this SalesPricelistItem) GetUomId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPricelistItemFieldUomId)
}

func (this SalesPricelistItem) GetPrice() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesPricelistItemFieldPrice)
}

func (this SalesPricelistItem) GetMinQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesPricelistItemFieldMinQuantity)
}

// AppliesToQuantity reports whether this item's quantity break covers the quantity being bought,
// inclusive of the break itself: min_quantity 10 applies to exactly 10.
func (this SalesPricelistItem) AppliesToQuantity(quantity decimal.Decimal) bool {
	return quantity.GreaterThanOrEqual(decimalOrZero(this.GetMinQuantity()))
}
