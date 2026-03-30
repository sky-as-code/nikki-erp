package group

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
	req = (*CreateGroupCommand)(nil)
	req = (*DeleteGroupCommand)(nil)
	req = (*GetGroupQuery)(nil)
	req = (*GroupExistsQuery)(nil)
	req = (*ManageGroupUsersCommand)(nil)
	req = (*UpdateGroupCommand)(nil)
	util.Unused(req)
}

var createGroupCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "createGroup",
}

type CreateGroupCommand struct {
	domain.Group
}

func (CreateGroupCommand) CqrsRequestType() cqrs.RequestType {
	return createGroupCommandType
}

func (CreateGroupCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.GroupSchemaName)
}

type CreateGroupResult = dyn.OpResult[domain.Group]

var deleteGroupCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "deleteGroup",
}

type DeleteGroupCommand dyn.DeleteOneQuery

func (DeleteGroupCommand) CqrsRequestType() cqrs.RequestType {
	return deleteGroupCommandType
}

type DeleteGroupResult = dyn.OpResult[dyn.MutateResultData]

var getGroupByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "getGroup",
}

type GetGroupQuery dyn.GetOneQuery

func (GetGroupQuery) CqrsRequestType() cqrs.RequestType {
	return getGroupByIdQueryType
}

type GetGroupResult = dyn.OpResult[domain.Group]

var existsCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "groupExists",
}

type GroupExistsQuery dyn.ExistsQuery

func (GroupExistsQuery) CqrsRequestType() cqrs.RequestType {
	return existsCommandType
}

type GroupExistsResult = dyn.OpResult[dyn.ExistsResultData]

var manageGroupUsersCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "manageGroupUsers",
}

type ManageGroupUsersCommand struct {
	GroupId model.Id                    `json:"group_id" param:"group_id"`
	Add     datastructure.Set[model.Id] `json:"add"`
	Remove  datastructure.Set[model.Id] `json:"remove"`
}

func (ManageGroupUsersCommand) CqrsRequestType() cqrs.RequestType {
	return manageGroupUsersCommandType
}

type ManageGroupUsersResult = dyn.OpResult[dyn.MutateResultData]

var searchGroupsQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "searchGroups",
}

type SearchGroupsQuery dyn.SearchQuery

func (SearchGroupsQuery) CqrsRequestType() cqrs.RequestType {
	return searchGroupsQueryType
}

type SearchGroupsResultData = dyn.PagedResultData[domain.Group]
type SearchGroupsResult = dyn.OpResult[SearchGroupsResultData]

var updateGroupCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "updateGroup",
}

type UpdateGroupCommand struct {
	domain.Group
}

func (UpdateGroupCommand) CqrsRequestType() cqrs.RequestType {
	return updateGroupCommandType
}

func (UpdateGroupCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.GroupSchemaName)
}

type UpdateGroupResult = dyn.OpResult[dyn.MutateResultData]
