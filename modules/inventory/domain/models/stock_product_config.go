package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// StockProductConfiguration holds the stock-owned settings of a product line, currently the unit
// its stock is counted in. It is a separate resource rather than a column on the product template
// so Product does not own a stock concept. One row per template; variants inherit it and cannot
// override it, because a variant with its own inventory unit makes a template's stock unsummable.
// There is no is_archived: it lives and dies with its template, and an archived template's
// configuration must stay readable so historical balances still resolve their recorded unit.
const (
	StockProductConfigSchemaName = "inventory_stock_product_config"

	StockProductConfigFieldId                = basemodel.FieldId
	StockProductConfigFieldProductTemplateId = "product_template_id"
	StockProductConfigFieldInventoryUomId    = "inventory_uom_id"
	StockProductConfigFieldOrgId             = "org_id"

	StockProductConfigEdgeProductTemplate = "product_template"
)

//go:embed stock_product_config.json
var stockProductConfigSchemaJson string

func StockProductConfigSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(stockProductConfigSchemaJson)
}

type StockProductConfig struct {
	basemodel.DynamicModelBase
}

func NewStockProductConfig() *StockProductConfig {
	return &StockProductConfig{basemodel.NewDynamicModel()}
}

func NewStockProductConfigFrom(src dmodel.DynamicFields) *StockProductConfig {
	return &StockProductConfig{basemodel.NewDynamicModel(src)}
}

func (this StockProductConfig) GetProductTemplateId() *model.Id {
	return this.GetFieldData().GetModelId(StockProductConfigFieldProductTemplateId)
}

func (this *StockProductConfig) SetProductTemplateId(v *model.Id) {
	this.GetFieldData().SetModelId(StockProductConfigFieldProductTemplateId, v)
}

func (this StockProductConfig) GetInventoryUomId() *model.Id {
	return this.GetFieldData().GetModelId(StockProductConfigFieldInventoryUomId)
}

func (this *StockProductConfig) SetInventoryUomId(v *model.Id) {
	this.GetFieldData().SetModelId(StockProductConfigFieldInventoryUomId, v)
}

func (this StockProductConfig) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(StockProductConfigFieldOrgId)
}

func (this *StockProductConfig) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(StockProductConfigFieldOrgId, v)
}
