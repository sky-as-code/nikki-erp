package hierarchy

import (
	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateHierarchyLevelCommand)(nil)
	req = (*DeleteHierarchyLevelCommand)(nil)
	req = (*GetHierarchyLevelQuery)(nil)
	req = (*HierarchyLevelExistsQuery)(nil)
	req = (*ManageHierarchyLevelUsersCommand)(nil)
	req = (*SearchHierarchyLevelsQuery)(nil)
	req = (*UpdateHierarchyLevelCommand)(nil)
	util.Unused(req)
}

var createHierarchyLevelCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "createHierarchyLevel",
}

type CreateHierarchyLevelCommand struct {
	domain.HierarchyLevel
}

func (CreateHierarchyLevelCommand) CqrsRequestType() cqrs.RequestType {
	return createHierarchyLevelCommandType
}

func (CreateHierarchyLevelCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.HierarchyLevelSchemaName)
}

type CreateHierarchyLevelResult = dyn.OpResult[domain.HierarchyLevel]

var deleteHierarchyLevelCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "deleteHierarchyLevel",
}

type DeleteHierarchyLevelCommand dyn.DeleteOneQuery

func (DeleteHierarchyLevelCommand) CqrsRequestType() cqrs.RequestType {
	return deleteHierarchyLevelCommandType
}

type DeleteHierarchyLevelResult = dyn.OpResult[dyn.MutateResultData]

var getHierarchyLevelByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "getHierarchyLevel",
}

type GetHierarchyLevelQuery dyn.GetOneQuery

func (GetHierarchyLevelQuery) CqrsRequestType() cqrs.RequestType {
	return getHierarchyLevelByIdQueryType
}

type GetHierarchyLevelResult = dyn.OpResult[domain.HierarchyLevel]

var searchHierarchyLevelsQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "searchHierarchyLevels",
}

type SearchHierarchyLevelsQuery dyn.SearchQuery

func (SearchHierarchyLevelsQuery) CqrsRequestType() cqrs.RequestType {
	return searchHierarchyLevelsQueryType
}

type SearchHierarchyLevelsResultData = dyn.PagedResultData[domain.HierarchyLevel]
type SearchHierarchyLevelsResult = dyn.OpResult[SearchHierarchyLevelsResultData]

var manageHierUsersCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "addRemoveUsers",
}

type ManageHierarchyLevelUsersCommand struct {
	HierarchyId model.Id                    `param:"hierarchy_id" json:"hierarchy_id"`
	Add         datastructure.Set[model.Id] `json:"add"`
	Remove      datastructure.Set[model.Id] `json:"remove"`
}

func (ManageHierarchyLevelUsersCommand) CqrsRequestType() cqrs.RequestType {
	return manageHierUsersCommandType
}

type ManageHierarchyLevelUsersResult = dyn.OpResult[dyn.MutateResultData]

var hierarchyLevelExistsQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "hierarchyLevelExists",
}

type HierarchyLevelExistsQuery dyn.ExistsQuery

func (HierarchyLevelExistsQuery) CqrsRequestType() cqrs.RequestType {
	return hierarchyLevelExistsQueryType
}

type HierarchyLevelExistsResult = dyn.OpResult[dyn.ExistsResultData]

var updateHierarchyLevelCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "updateHierarchyLevel",
}

type UpdateHierarchyLevelCommand struct {
	domain.HierarchyLevel
}

func (UpdateHierarchyLevelCommand) CqrsRequestType() cqrs.RequestType {
	return updateHierarchyLevelCommandType
}

func (UpdateHierarchyLevelCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.HierarchyLevelSchemaName)
}

type UpdateHierarchyLevelResult = dyn.OpResult[dyn.MutateResultData]
