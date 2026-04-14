package unit

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateUnitCommand)(nil)
	req = (*DeleteUnitCommand)(nil)
	req = (*GetUnitQuery)(nil)
	req = (*SearchUnitsQuery)(nil)
	req = (*UpdateUnitCommand)(nil)
	req = (*UnitExistsQuery)(nil)
	util.Unused(req)
}

var createUnitCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit",
	Action:    "create",
}

type CreateUnitCommand struct {
	domain.Unit
}

func (CreateUnitCommand) CqrsRequestType() cqrs.RequestType {
	return createUnitCommandType
}

func (this CreateUnitCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.UnitSchemaName)
}

type CreateUnitResult = dyn.OpResult[domain.Unit]

var updateUnitCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit",
	Action:    "update",
}

type UpdateUnitCommand struct {
	domain.Unit
}

func (UpdateUnitCommand) CqrsRequestType() cqrs.RequestType {
	return updateUnitCommandType
}

func (this UpdateUnitCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.UnitSchemaName)
}

type UpdateUnitResult = dyn.OpResult[dyn.MutateResultData]

var deleteUnitCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit",
	Action:    "delete",
}

type DeleteUnitCommand dyn.DeleteOneCommand

func (DeleteUnitCommand) CqrsRequestType() cqrs.RequestType {
	return deleteUnitCommandType
}

type DeleteUnitResult = dyn.OpResult[dyn.MutateResultData]

var getUnitQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit",
	Action:    "getUnit",
}

type GetUnitQuery dyn.GetOneQuery

func (GetUnitQuery) CqrsRequestType() cqrs.RequestType {
	return getUnitQueryType
}

type GetUnitResult = dyn.OpResult[domain.Unit]

var searchUnitsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit",
	Action:    "search",
}

type SearchUnitsQuery dyn.SearchQuery

func (SearchUnitsQuery) CqrsRequestType() cqrs.RequestType {
	return searchUnitsQueryType
}

type SearchUnitsResultData = dyn.PagedResultData[domain.Unit]
type SearchUnitsResult = dyn.OpResult[SearchUnitsResultData]

var unitExistsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit",
	Action:    "exists",
}

type UnitExistsQuery dyn.ExistsQuery

func (UnitExistsQuery) CqrsRequestType() cqrs.RequestType {
	return unitExistsQueryType
}

type UnitExistsResult = dyn.OpResult[dyn.ExistsResultData]
