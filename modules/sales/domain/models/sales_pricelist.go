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

	SalesPricelistItemEdgeSalesPricelist = "sales_pricelist"
)

// PricelistScope ranks how specifically a pricelist applies, which is what decides between two that
// both match.
//
// Specificity beats priority, always: a point-scoped list wins over a channel-scoped one whatever
// their priority numbers, because otherwise a high-priority global list would silently undo every
// local price an operator had set. Priority only breaks ties between lists of the SAME scope.
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

// SalesPricelist is a set of prices that applies to some scope for some window.
//
// It exists because BR §87.2 forbids hard-coding every price onto the product: the same variant
// legitimately costs different amounts in an airport kiosk and a high street store, and a price
// scheduled for next month must not change what today's sales are charged.
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

// Scope reports how specifically this pricelist applies.
//
// The point scope is checked first: a list naming both a point and a channel is point-scoped, since
// the point already implies its channel and the narrower answer is the one that should win.
func (this SalesPricelist) Scope() PricelistScope {
	if this.GetSalesPointId() != nil {
		return PricelistScopePoint
	}
	if this.GetSalesChannelId() != nil {
		return PricelistScopeChannel
	}
	return PricelistScopeGlobal
}

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

// AppliesToQuantity reports whether this item's quantity break covers the quantity being bought.
//
// Inclusive of the break itself: an item declaring min_quantity 10 applies to exactly 10.
func (this SalesPricelistItem) AppliesToQuantity(quantity decimal.Decimal) bool {
	return quantity.GreaterThanOrEqual(decimalOrZero(this.GetMinQuantity()))
}
