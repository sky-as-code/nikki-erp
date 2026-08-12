package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// StockLocationType says what a location is for, which decides whether stock sitting in it is
// company-owned. Only Internal holds real inventory; the rest are the counterparties and virtual
// locations that give every movement an opposite side, so that a receipt, a delivery, an
// adjustment and a scrap are all the same operation with different endpoints. See BR §4.2.
const (
	StockLocationTypeInternal      = "internal"
	StockLocationTypeCustomer      = "customer"
	StockLocationTypeSupplier      = "supplier"
	StockLocationTypeInventoryLoss = "inventory_loss"
	StockLocationTypeScrap         = "scrap"
	StockLocationTypeTransit       = "transit"
)

const (
	StockLocationSchemaName = "inventory_stock_location"

	StockLocationFieldId               = basemodel.FieldId
	StockLocationFieldCode             = "code"
	StockLocationFieldName             = "name"
	StockLocationFieldLocationType     = "location_type"
	StockLocationFieldParentLocationId = "parent_location_id"
	StockLocationFieldDescription      = "description"
	StockLocationFieldOrgId            = "org_id"

	StockLocationEdgeParentLocation = "parent_location"
)

//go:embed stock_location.json
var stockLocationSchemaJson string

func StockLocationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(stockLocationSchemaJson)
}

type StockLocation struct {
	basemodel.DynamicModelBase
}

func NewStockLocation() *StockLocation {
	return &StockLocation{basemodel.NewDynamicModel()}
}

func NewStockLocationFrom(src dmodel.DynamicFields) *StockLocation {
	return &StockLocation{basemodel.NewDynamicModel(src)}
}

func (this StockLocation) GetCode() *string {
	return this.GetFieldData().GetString(StockLocationFieldCode)
}

func (this *StockLocation) SetCode(v *string) {
	this.GetFieldData().SetString(StockLocationFieldCode, v)
}

func (this StockLocation) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(StockLocationFieldName)
}

func (this *StockLocation) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(StockLocationFieldName, v)
}

func (this StockLocation) GetLocationType() *string {
	return this.GetFieldData().GetString(StockLocationFieldLocationType)
}

func (this *StockLocation) SetLocationType(v *string) {
	this.GetFieldData().SetString(StockLocationFieldLocationType, v)
}

// GetParentLocationId returns the parent in the location tree, or nil for a root location.
func (this StockLocation) GetParentLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockLocationFieldParentLocationId)
}

func (this *StockLocation) SetParentLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockLocationFieldParentLocationId, v)
}

func (this StockLocation) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(StockLocationFieldOrgId)
}

func (this *StockLocation) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(StockLocationFieldOrgId, v)
}
