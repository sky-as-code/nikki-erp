package organization

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
	req = (*CreateOrgCommand)(nil)
	req = (*DeleteOrgCommand)(nil)
	req = (*GetOrgQuery)(nil)
	req = (*SearchOrgsQuery)(nil)
	req = (*ManageOrgUsersCommand)(nil)
	req = (*OrgExistsQuery)(nil)
	req = (*UpdateOrgCommand)(nil)
	util.Unused(req)
}

var createOrganizationCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "createOrg",
}

type CreateOrgCommand struct {
	domain.Organization
}

func (CreateOrgCommand) CqrsRequestType() cqrs.RequestType {
	return createOrganizationCommandType
}

func (CreateOrgCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.OrganizationSchemaName)
}

type CreateOrgResult = dyn.OpResult[domain.Organization]

var deleteOrganizationCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "deleteOrg",
}

type DeleteOrgCommand dyn.DeleteOneCommand

func (DeleteOrgCommand) CqrsRequestType() cqrs.RequestType {
	return deleteOrganizationCommandType
}

type DeleteOrgResult = dyn.OpResult[dyn.MutateResultData]

var getOrgQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "getOrg",
}

type GetOrgQuery struct {
	Columns []string `json:"columns" query:"columns"`
	Id      *string  `json:"id" param:"id"`
	Slug    *string  `json:"slug"`
}

func (GetOrgQuery) CqrsRequestType() cqrs.RequestType {
	return getOrgQueryType
}

type GetOrgResult = dyn.OpResult[domain.Organization]

var orgExistsQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "orgExists",
}

type OrgExistsQuery dyn.ExistsQuery

func (OrgExistsQuery) CqrsRequestType() cqrs.RequestType {
	return orgExistsQueryType
}

type OrgExistsResult = dyn.OpResult[dyn.ExistsResultData]

var manageOrgUsersCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "manageOrgUsers",
}

type ManageOrgUsersCommand struct {
	OrgId  model.Id                    `json:"org_id" param:"org_id"`
	Add    datastructure.Set[model.Id] `json:"add"`
	Remove datastructure.Set[model.Id] `json:"remove"`
}

func (ManageOrgUsersCommand) CqrsRequestType() cqrs.RequestType {
	return manageOrgUsersCommandType
}

type ManageOrgUsersResult = dyn.OpResult[dyn.MutateResultData]

var searchOrgsQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "searchOrgs",
}

type SearchOrgsQuery dyn.SearchQuery

func (SearchOrgsQuery) CqrsRequestType() cqrs.RequestType {
	return searchOrgsQueryType
}

type SearchOrgsResultData = dyn.PagedResultData[domain.Organization]
type SearchOrgsResult = dyn.OpResult[SearchOrgsResultData]

var setOrgIsArchivedCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "setOrgIsArchived",
}

type SetOrgIsArchivedCommand dyn.SetIsArchivedCommand

func (SetOrgIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setOrgIsArchivedCommandType
}

type SetOrgIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var updateOrgCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "updateOrg",
}

type UpdateOrgCommand struct {
	domain.Organization
}

func (UpdateOrgCommand) CqrsRequestType() cqrs.RequestType {
	return updateOrgCommandType
}

func (UpdateOrgCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.OrganizationSchemaName)
}

type UpdateOrgResult = dyn.OpResult[dyn.MutateResultData]
