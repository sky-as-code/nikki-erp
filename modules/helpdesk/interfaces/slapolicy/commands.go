package slapolicy

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
)

func init() {
	var req cqrs.Request
	req = (*CreateSlaPolicyCommand)(nil)
	req = (*DeleteSlaPolicyCommand)(nil)
	req = (*GetSlaPolicyQuery)(nil)
	req = (*SlaPolicyExistsQuery)(nil)
	req = (*SearchSlaPoliciesQuery)(nil)
	req = (*UpdateSlaPolicyCommand)(nil)
	req = (*SetSlaPolicyIsArchivedCommand)(nil)
	util.Unused(req)
}

var createSlaPolicyCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "slapolicy", Action: "createSlaPolicy"}

type CreateSlaPolicyCommand struct{ models.SlaPolicy }

func (CreateSlaPolicyCommand) CqrsRequestType() cqrs.RequestType { return createSlaPolicyCommandType }
func (CreateSlaPolicyCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.SlaPolicySchemaName)
}

type CreateSlaPolicyResult = dyn.OpResult[models.SlaPolicy]

var deleteSlaPolicyCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "slapolicy", Action: "deleteSlaPolicy"}

type DeleteSlaPolicyCommand dyn.DeleteOneCommand

func (DeleteSlaPolicyCommand) CqrsRequestType() cqrs.RequestType { return deleteSlaPolicyCommandType }

type DeleteSlaPolicyResult = dyn.OpResult[dyn.MutateResultData]

var getSlaPolicyQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "slapolicy", Action: "getSlaPolicy"}

type GetSlaPolicyQuery dyn.GetOneQuery

func (GetSlaPolicyQuery) CqrsRequestType() cqrs.RequestType { return getSlaPolicyQueryType }

type GetSlaPolicyResult = dyn.OpResult[models.SlaPolicy]

var slaPolicyExistsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "slapolicy", Action: "slaPolicyExists"}

type SlaPolicyExistsQuery dyn.ExistsQuery

func (SlaPolicyExistsQuery) CqrsRequestType() cqrs.RequestType { return slaPolicyExistsQueryType }

type SlaPolicyExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchSlaPoliciesQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "slapolicy", Action: "searchSlaPolicies"}

type SearchSlaPoliciesQuery dyn.SearchQuery

func (SearchSlaPoliciesQuery) CqrsRequestType() cqrs.RequestType { return searchSlaPoliciesQueryType }

type SearchSlaPoliciesResultData = dyn.PagedResultData[models.SlaPolicy]
type SearchSlaPoliciesResult = dyn.OpResult[SearchSlaPoliciesResultData]

var updateSlaPolicyCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "slapolicy", Action: "updateSlaPolicy"}

type UpdateSlaPolicyCommand struct{ models.SlaPolicy }

func (UpdateSlaPolicyCommand) CqrsRequestType() cqrs.RequestType { return updateSlaPolicyCommandType }
func (UpdateSlaPolicyCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.SlaPolicySchemaName)
}

type UpdateSlaPolicyResult = dyn.OpResult[dyn.MutateResultData]

var setSlaPolicyIsArchivedCommandType = cqrs.RequestType{
	Module:    "helpdesk",
	Submodule: "slapolicy",
	Action:    "setSlaPolicyIsArchived",
}

type SetSlaPolicyIsArchivedCommand dyn.SetIsArchivedCommand

func (SetSlaPolicyIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setSlaPolicyIsArchivedCommandType
}

type SetSlaPolicyIsArchivedResult = dyn.OpResult[dyn.MutateResultData]
