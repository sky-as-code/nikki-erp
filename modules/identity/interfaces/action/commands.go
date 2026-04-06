package action

import (
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func init() {
	var req cqrs.Request
	req = (*CreateActionCommand)(nil)
	req = (*DeleteActionCommand)(nil)
	req = (*GetActionQuery)(nil)
	req = (*ActionExistsQuery)(nil)
	req = (*SearchActionsQuery)(nil)
	req = (*UpdateActionCommand)(nil)
	util.Unused(req)
}

var createActionCommandType = cqrs.RequestType{Module: "identity", Submodule: "action", Action: "createAction"}

type CreateActionCommand struct {
	domain.Action
}

func (CreateActionCommand) CqrsRequestType() cqrs.RequestType { return createActionCommandType }

func (CreateActionCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ActionSchemaName)
}

type CreateActionResult = dyn.OpResult[domain.Action]

var deleteActionCommandType = cqrs.RequestType{Module: "identity", Submodule: "action", Action: "deleteAction"}

type DeleteActionCommand dyn.DeleteOneCommand

func (DeleteActionCommand) CqrsRequestType() cqrs.RequestType { return deleteActionCommandType }

type DeleteActionResult = dyn.OpResult[dyn.MutateResultData]

var getActionQueryType = cqrs.RequestType{Module: "identity", Submodule: "action", Action: "getAction"}

type GetActionQuery dyn.GetOneQuery

func (GetActionQuery) CqrsRequestType() cqrs.RequestType { return getActionQueryType }

type GetActionResult = dyn.OpResult[domain.Action]

var actionExistsQueryType = cqrs.RequestType{Module: "identity", Submodule: "action", Action: "actionExists"}

type ActionExistsQuery dyn.ExistsQuery

func (ActionExistsQuery) CqrsRequestType() cqrs.RequestType { return actionExistsQueryType }

type ActionExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchActionsQueryType = cqrs.RequestType{Module: "identity", Submodule: "action", Action: "searchActions"}

type SearchActionsQuery dyn.SearchQuery

func (SearchActionsQuery) CqrsRequestType() cqrs.RequestType { return searchActionsQueryType }

type SearchActionsResultData = dyn.PagedResultData[domain.Action]
type SearchActionsResult = dyn.OpResult[SearchActionsResultData]

var updateActionCommandType = cqrs.RequestType{Module: "identity", Submodule: "action", Action: "updateAction"}

type UpdateActionCommand struct {
	domain.Action
}

func (UpdateActionCommand) CqrsRequestType() cqrs.RequestType { return updateActionCommandType }

func (UpdateActionCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ActionSchemaName)
}

type UpdateActionResult = dyn.OpResult[dyn.MutateResultData]
