package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// StockProductConfiguration is the stock-owned settings of a product line.
//
// Today it holds one thing: which unit the product's stock is counted in. It exists as its own
// resource rather than as a column on the product template because the answer belongs to Stock —
// it decides what a balance means — while the template is Product master data. Putting it on the
// template would make Product the owner of a stock concept, which the requirement forbids
// (CR §11.3, §11.4, PROD-INT-INV-009).
//
// One row per template. Variants inherit it and cannot override it: a variant with its own
// inventory unit would make a template's stock unsummable, and the requirement defers that to a
// change request of its own (CR §11.5, PROD-INT-INV-011, PROD-INT-INV-012).
//
// There is no is_archived. The configuration is not a thing a user retires on its own — it lives
// and dies with the template it configures, and an archived template's configuration must stay
// readable so historical balances keep resolving the unit they were recorded in.
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
