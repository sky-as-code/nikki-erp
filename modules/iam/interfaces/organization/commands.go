package organization

import (
	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
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
	Module:    "iam",
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
	Module:    "iam",
	Submodule: "organization",
	Action:    "deleteOrg",
}

type DeleteOrgCommand dyn.DeleteOneCommand

func (DeleteOrgCommand) CqrsRequestType() cqrs.RequestType {
	return deleteOrganizationCommandType
}

type DeleteOrgResult = dyn.OpResult[dyn.MutateResultData]

var getOrgQueryType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "organization",
	Action:    "getOrg",
}

type GetOrgQuery struct {
	Fields []string `json:"fields" query:"fields"`
	Id     *string  `json:"id" param:"id"`
	Slug   *string  `json:"slug"`
}

func (GetOrgQuery) CqrsRequestType() cqrs.RequestType {
	return getOrgQueryType
}

func (GetOrgQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.get_org_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(dmodel.DefineField().
					Name(basemodel.FieldFields).
					DataType(dmodel.FieldDataTypeString(model.MODEL_RULE_FIELDS_LENGTH_MIN, model.MODEL_RULE_FIELDS_LENGTH_MAX).ArrayType())).
				ExclusiveRequiredFields(basemodel.FieldId, domain.OrgFieldSlug).
				Field(dmodel.DefineField().
					Name(basemodel.FieldId).
					DataType(dmodel.FieldDataTypeUlid())).
				Field(dmodel.DefineField().
					Name(domain.OrgFieldSlug).
					DataType(dmodel.FieldDataTypeEmail()))
		},
	)
}

type GetOrgResult = dyn.OpResult[dyn.SingleResultData[domain.Organization]]

var orgExistsQueryType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "organization",
	Action:    "orgExists",
}

type OrgExistsQuery dyn.ExistsQuery

func (OrgExistsQuery) CqrsRequestType() cqrs.RequestType {
	return orgExistsQueryType
}

type OrgExistsResult = dyn.OpResult[dyn.ExistsResultData]

var manageOrgUsersCommandType = cqrs.RequestType{
	Module:    "iam",
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
	Module:    "iam",
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
	Module:    "iam",
	Submodule: "organization",
	Action:    "setOrgIsArchived",
}

type SetOrgIsArchivedCommand dyn.SetIsArchivedCommand

func (SetOrgIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setOrgIsArchivedCommandType
}

type SetOrgIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var updateOrgCommandType = cqrs.RequestType{
	Module:    "iam",
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
