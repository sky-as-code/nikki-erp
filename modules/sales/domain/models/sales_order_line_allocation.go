package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesOrderLineAllocationSchemaName = "sales_order_line_allocation"

	SalesOrderLineAllocationFieldId               = "id"
	SalesOrderLineAllocationFieldOrgId            = "org_id"
	SalesOrderLineAllocationFieldSalesOrderLineId = "sales_order_line_id"
	SalesOrderLineAllocationFieldSourceLocationId = "source_location_id"
	SalesOrderLineAllocationFieldQuantity         = "quantity"

	SalesOrderLineAllocationEdgeSalesOrderLine = "sales_order_line"
)

//go:embed sales_order_line_allocation.json
var salesOrderLineAllocationSchemaJson string

func SalesOrderLineAllocationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesOrderLineAllocationSchemaJson)
}

// SalesOrderLineAllocation assigns part of a commercial order line to the inventory location
// that will fulfil it. A line may have many allocations, but each location occurs once on that
// line; fulfillment items are later created from these persisted allocations.
type SalesOrderLineAllocation struct {
	basemodel.DynamicModelBase
}

func NewSalesOrderLineAllocation() *SalesOrderLineAllocation {
	return &SalesOrderLineAllocation{basemodel.NewDynamicModel()}
}

func NewSalesOrderLineAllocationFrom(src dmodel.DynamicFields) *SalesOrderLineAllocation {
	return &SalesOrderLineAllocation{basemodel.NewDynamicModel(src)}
}

func (this SalesOrderLineAllocation) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesOrderLineAllocation) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesOrderLineAllocation) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderLineAllocationFieldOrgId)
}

func (this *SalesOrderLineAllocation) SetOrgId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderLineAllocationFieldOrgId, id)
}

func (this SalesOrderLineAllocation) GetSalesOrderLineId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderLineAllocationFieldSalesOrderLineId)
}

func (this *SalesOrderLineAllocation) SetSalesOrderLineId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderLineAllocationFieldSalesOrderLineId, id)
}

func (this SalesOrderLineAllocation) GetSourceLocationId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderLineAllocationFieldSourceLocationId)
}

func (this *SalesOrderLineAllocation) SetSourceLocationId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderLineAllocationFieldSourceLocationId, id)
}

func (this SalesOrderLineAllocation) GetQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineAllocationFieldQuantity)
}

func (this *SalesOrderLineAllocation) SetQuantity(quantity *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineAllocationFieldQuantity, quantity)
}
