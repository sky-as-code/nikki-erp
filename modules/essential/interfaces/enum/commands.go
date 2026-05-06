package enum

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

func init() {
	var req cqrs.Request
	req = (*CreateEnumCommand)(nil)
	req = (*DeleteEnumCommand)(nil)
	req = (*GetEnumQuery)(nil)
	req = (*SearchEnumsQuery)(nil)
	req = (*UpdateEnumCommand)(nil)
	req = (*EnumExistsQuery)(nil)
	util.Unused(req)
}

var createEnumCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "enum",
	Action:    "create",
}

type CreateEnumCommand struct {
	models.Enum
}

func (CreateEnumCommand) CqrsRequestType() cqrs.RequestType {
	return createEnumCommandType
}

func (this CreateEnumCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.EnumSchemaName)
}

type CreateEnumResult = dyn.OpResult[models.Enum]

var updateEnumCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "enum",
	Action:    "update",
}

type UpdateEnumCommand struct {
	models.Enum
}

func (UpdateEnumCommand) CqrsRequestType() cqrs.RequestType {
	return updateEnumCommandType
}

func (this UpdateEnumCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.EnumSchemaName)
}

type UpdateEnumResult = dyn.OpResult[dyn.MutateResultData]

var deleteEnumCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "enum",
	Action:    "delete",
}

type DeleteEnumCommand dyn.DeleteOneCommand

func (DeleteEnumCommand) CqrsRequestType() cqrs.RequestType {
	return deleteEnumCommandType
}

type DeleteEnumResult = dyn.OpResult[dyn.MutateResultData]

var getEnumQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "enum",
	Action:    "getEnum",
}

type GetEnumQuery dyn.GetOneQuery

func (GetEnumQuery) CqrsRequestType() cqrs.RequestType {
	return getEnumQueryType
}

type GetEnumResult = dyn.OpResult[models.Enum]

var searchEnumsQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "enum",
	Action:    "search",
}

type SearchEnumsQuery dyn.SearchQuery

func (SearchEnumsQuery) CqrsRequestType() cqrs.RequestType {
	return searchEnumsQueryType
}

type SearchEnumsResultData = dyn.PagedResultData[models.Enum]
type SearchEnumsResult = dyn.OpResult[SearchEnumsResultData]

var enumExistsQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "enum",
	Action:    "exists",
}

type EnumExistsQuery dyn.ExistsQuery

func (EnumExistsQuery) CqrsRequestType() cqrs.RequestType {
	return enumExistsQueryType
}

type EnumExistsResult = dyn.OpResult[dyn.ExistsResultData]
