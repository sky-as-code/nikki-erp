package orgunit

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
	req = (*CreateOrgUnitCommand)(nil)
	req = (*DeleteOrgUnitCommand)(nil)
	req = (*GetOrgUnitQuery)(nil)
	req = (*OrgUnitExistsQuery)(nil)
	req = (*ManageOrgUnitUsersCommand)(nil)
	req = (*SearchOrgUnitsQuery)(nil)
	req = (*UpdateOrgUnitCommand)(nil)
	util.Unused(req)
}

var createOrgUnitCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "orgunit",
	Action:    "createOrgUnit",
}

type CreateOrgUnitCommand struct {
	domain.OrganizationalUnit
}

func (CreateOrgUnitCommand) CqrsRequestType() cqrs.RequestType {
	return createOrgUnitCommandType
}

func (CreateOrgUnitCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.OrganizationalUnitSchemaName)
}

type CreateOrgUnitResult = dyn.OpResult[domain.OrganizationalUnit]

var deleteOrgUnitCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "orgunit",
	Action:    "deleteOrgUnit",
}

type DeleteOrgUnitCommand dyn.DeleteOneCommand

func (DeleteOrgUnitCommand) CqrsRequestType() cqrs.RequestType {
	return deleteOrgUnitCommandType
}

type DeleteOrgUnitResult = dyn.OpResult[dyn.MutateResultData]

var getOrgUnitByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "orgunit",
	Action:    "getOrgUnit",
}

type GetOrgUnitQuery dyn.GetOneQuery

func (GetOrgUnitQuery) CqrsRequestType() cqrs.RequestType {
	return getOrgUnitByIdQueryType
}

type GetOrgUnitResult = dyn.OpResult[domain.OrganizationalUnit]

var searchOrgUnitsQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "orgunit",
	Action:    "searchOrgUnits",
}

type SearchOrgUnitsQuery dyn.SearchQuery

func (SearchOrgUnitsQuery) CqrsRequestType() cqrs.RequestType {
	return searchOrgUnitsQueryType
}

type SearchOrgUnitsResultData = dyn.PagedResultData[domain.OrganizationalUnit]
type SearchOrgUnitsResult = dyn.OpResult[SearchOrgUnitsResultData]

var manageOrgUnitUsersCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "orgunit",
	Action:    "manageOrgUnitUsers",
}

type ManageOrgUnitUsersCommand struct {
	OrgUnitId model.Id                    `param:"orgunit_id" json:"orgunit_id"`
	Add       datastructure.Set[model.Id] `json:"add"`
	Remove    datastructure.Set[model.Id] `json:"remove"`
}

func (ManageOrgUnitUsersCommand) CqrsRequestType() cqrs.RequestType {
	return manageOrgUnitUsersCommandType
}

type ManageOrgUnitUsersResult = dyn.OpResult[dyn.MutateResultData]

var orgUnitExistsQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "orgunit",
	Action:    "orgUnitExists",
}

type OrgUnitExistsQuery dyn.ExistsQuery

func (OrgUnitExistsQuery) CqrsRequestType() cqrs.RequestType {
	return orgUnitExistsQueryType
}

type OrgUnitExistsResult = dyn.OpResult[dyn.ExistsResultData]

var updateOrgUnitCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "orgunit",
	Action:    "updateOrgUnit",
}

type UpdateOrgUnitCommand struct {
	domain.OrganizationalUnit
}

func (UpdateOrgUnitCommand) CqrsRequestType() cqrs.RequestType {
	return updateOrgUnitCommandType
}

func (UpdateOrgUnitCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.OrganizationalUnitSchemaName)
}

type UpdateOrgUnitResult = dyn.OpResult[dyn.MutateResultData]
