package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	StockMoveDependencySchemaName = "inventory_stock_move_dependency"

	StockMoveDependencyFieldId                = basemodel.FieldId
	StockMoveDependencyFieldPredecessorMoveId = "predecessor_move_id"
	StockMoveDependencyFieldSuccessorMoveId   = "successor_move_id"
	StockMoveDependencyFieldOrgId             = "org_id"

	StockMoveDependencyEdgePredecessorMove = "predecessor_move"
	StockMoveDependencyEdgeSuccessorMove   = "successor_move"
)

//go:embed stock_move_dependency.json
var stockMoveDependencySchemaJson string

func StockMoveDependencySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(stockMoveDependencySchemaJson)
}

// StockMoveDependency orders the steps of a multi-step flow: the successor stays 'waiting' until
// its predecessor is done. The pair must stay acyclic, which a domain service enforces because the
// schema cannot express it. See BR §7.7.
type StockMoveDependency struct {
	basemodel.DynamicModelBase
}

func NewStockMoveDependency() *StockMoveDependency {
	return &StockMoveDependency{basemodel.NewDynamicModel()}
}

func NewStockMoveDependencyFrom(src dmodel.DynamicFields) *StockMoveDependency {
	return &StockMoveDependency{basemodel.NewDynamicModel(src)}
}

func (this StockMoveDependency) GetPredecessorMoveId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveDependencyFieldPredecessorMoveId)
}

func (this *StockMoveDependency) SetPredecessorMoveId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveDependencyFieldPredecessorMoveId, v)
}

func (this StockMoveDependency) GetSuccessorMoveId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveDependencyFieldSuccessorMoveId)
}

func (this *StockMoveDependency) SetSuccessorMoveId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveDependencyFieldSuccessorMoveId, v)
}

func (this StockMoveDependency) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveDependencyFieldOrgId)
}

func (this *StockMoveDependency) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveDependencyFieldOrgId, v)
}
